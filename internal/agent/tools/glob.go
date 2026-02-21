package tools

import (
	"bytes"
	"context"
	"errors"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/fantasy"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/fsext"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/charlievieth/fastwalk"
)

const GlobToolName = "glob"

//go:embed glob.md
var globDescription []byte

type GlobParams struct {
	Pattern string `json:"pattern" description:"The glob pattern to match files against"`
	Path    string `json:"path,omitempty" description:"The directory to search in. Defaults to the current working directory."`
}

type GlobResponseMetadata struct {
	NumberOfFiles int  `json:"number_of_files"`
	Truncated     bool `json:"truncated"`
}

func NewGlobTool(workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		GlobToolName,
		string(globDescription),
		func(ctx context.Context, params GlobParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Create progress emitter at the top of the handler
			emitter := NewProgressEmitter(call.ID, func(e ToolProgressEvent) {
				if progressCallback != nil {
					progressCallback(call.ID, e)
				}
			})

			if params.Pattern == "" {
				return fantasy.NewTextErrorResponse("pattern is required"), nil
			}

			// Emit start after input validation
			emitter.EmitStart("Starting file search...")

			searchPath := params.Path
			if searchPath == "" {
				searchPath = workingDir
			}

			files, truncated, err := globFilesWithProgress(ctx, params.Pattern, searchPath, 100, emitter)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("error finding files: %w", err)
			}

			// Emit completion
			emitter.EmitComplete(fmt.Sprintf("Found %d matching files", len(files)))

			var output string
			if len(files) == 0 {
				output = "No files found"
			} else {
				normalizeFilePaths(files)
				output = strings.Join(files, "\n")
				if truncated {
					output += "\n\n(Results are truncated. Consider using a more specific path or pattern.)"
				}
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(output),
				GlobResponseMetadata{
					NumberOfFiles: len(files),
					Truncated:     truncated,
				},
			), nil
		})
}

func globFilesWithProgress(ctx context.Context, pattern, searchPath string, limit int, emitter *ProgressEmitter) ([]string, bool, error) {
	// Try ripgrep first
	cmdRg := getRgCmd(ctx, pattern)
	if cmdRg != nil {
		cmdRg.Dir = searchPath
		matches, err := runRipgrep(cmdRg, searchPath, limit)
		if err == nil {
			return matches, len(matches) >= limit && limit > 0, nil
		}
		slog.Warn("Ripgrep execution failed, falling back to doublestar", "error", err)
	}

	// Use custom walker with progress tracking
	return globWithDoubleStarAndProgress(pattern, searchPath, limit, emitter)
}

func globWithDoubleStarAndProgress(pattern, searchPath string, limit int, emitter *ProgressEmitter) ([]string, bool, error) {
	// Normalize pattern to forward slashes
	pattern = filepath.ToSlash(pattern)

	walker := fsext.NewFastGlobWalker(searchPath)

	var dirsScanned int
	var filesFound int
	maxDepth := 50 // Default max depth for percent calculation

	type fileResult struct {
		path    string
		modTime int64
	}
	var results []fileResult
	var truncated bool

	conf := fastwalk.Config{
		Follow:  true,
		ToSlash: fastwalk.DefaultToSlash(),
		Sort:    fastwalk.SortFilesFirst,
	}

	err := fastwalk.Walk(&conf, searchPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if d.IsDir() {
			dirsScanned++
			if walker.ShouldSkip(path) {
				return filepath.SkipDir
			}

			// Emit progress every 50 directories
			if dirsScanned%50 == 0 {
				percent := min(90, int(float64(dirsScanned)/float64(maxDepth*10)*100))
				emitter.Emit(fmt.Sprintf("Scanned %d directories, found %d files...", dirsScanned, filesFound), percent)
			}
		} else {
			// Track files for progress reporting
			filesFound++
		}

		if walker.ShouldSkip(path) {
			return nil
		}

		relPath, err := filepath.Rel(searchPath, path)
		if err != nil {
			relPath = path
		}

		// Normalize separators to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Check if path matches the pattern
		matched, err := doublestar.Match(pattern, relPath)
		if err != nil || !matched {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		results = append(results, fileResult{path: path, modTime: info.ModTime().UnixNano()})
		if limit > 0 && len(results) >= limit*2 {
			truncated = true
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil && !errors.Is(err, filepath.SkipAll) {
		return nil, false, fmt.Errorf("fastwalk error: %w", err)
	}

	// Sort by modification time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].modTime > results[j].modTime
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
		truncated = true
	}

	matches := make([]string, len(results))
	for i, r := range results {
		matches[i] = r.path
	}

	return matches, truncated || errors.Is(err, filepath.SkipAll), nil
}

func globFiles(ctx context.Context, pattern, searchPath string, limit int) ([]string, bool, error) {
	cmdRg := getRgCmd(ctx, pattern)
	if cmdRg != nil {
		cmdRg.Dir = searchPath
		matches, err := runRipgrep(cmdRg, searchPath, limit)
		if err == nil {
			return matches, len(matches) >= limit && limit > 0, nil
		}
		slog.Warn("Ripgrep execution failed, falling back to doublestar", "error", err)
	}

	return fsext.GlobWithDoubleStar(pattern, searchPath, limit)
}

func runRipgrep(cmd *exec.Cmd, searchRoot string, limit int) ([]string, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("ripgrep: %w\n%s", err, out)
	}

	var matches []string
	for p := range bytes.SplitSeq(out, []byte{0}) {
		if len(p) == 0 {
			continue
		}
		absPath := string(p)
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(searchRoot, absPath)
		}
		if fsext.SkipHidden(absPath) {
			continue
		}
		matches = append(matches, absPath)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return len(matches[i]) < len(matches[j])
	})

	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

func normalizeFilePaths(paths []string) {
	for i, p := range paths {
		paths[i] = filepath.ToSlash(p)
	}
}
