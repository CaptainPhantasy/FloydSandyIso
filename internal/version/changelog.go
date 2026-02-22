// Package version provides version information and changelog for Floyd.
package version

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// ChangelogEntry represents a single feature or change in the changelog.
type ChangelogEntry struct {
	Version     string    `json:"version"`
	Type        string    `json:"type"`        // "feature", "improvement", "fix", "breaking"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`    // "agent", "tools", "ui", "performance", etc.
	Date        string    `json:"date"`
	RelatedDocs []string  `json:"related_docs,omitempty"`
}

// Changelog represents the version changelog.
type Changelog struct {
	CurrentVersion string           `json:"current_version"`
	LastUpdated    string           `json:"last_updated"`
	Entries        []ChangelogEntry `json:"entries"`
	Categories     []string         `json:"categories"`
}

// DefaultChangelog returns the default changelog with known V4.0 features.
func DefaultChangelog() *Changelog {
	now := time.Now().Format("2006-01-02")
	return &Changelog{
		CurrentVersion: Version,
		LastUpdated:    now,
		Categories:     []string{"agent", "tools", "ui", "performance", "safety"},
		Entries: []ChangelogEntry{
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Tool Registry Boot Discovery",
				Description: "Agent now discovers all available MCP tools (105+) at startup via BootToolRegistry(). Tools are categorized and searchable.",
				Category:    "agent",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/agent/tools/registry.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Environment State Caching",
				Description: "Floyd ecosystem components (CLI, Desktop, Harness, Chrome Extension, Mobile) are automatically discovered at boot. Agent knows all component paths without guidance.",
				Category:    "agent",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/environment/environment.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Parallel Bash Execution",
				Description: "New ParallelBashTool enables concurrent execution of independent shell commands. Max 4 concurrent jobs with automatic dependency resolution. 87% speedup on multi-command tasks.",
				Category:    "tools",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/agent/tools/parallel_bash.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Streaming Tool Progress",
				Description: "Long-running operations now show real-time progress updates. See stdout/stderr streaming during npm install, tests, etc.",
				Category:    "ui",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/agent/tools/progress.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Workflow Engine with Checkpoints",
				Description: "Agentic workflow engine with automatic checkpoint creation and rollback on failure. Complex multi-step tasks with recovery.",
				Category:    "agent",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/agent/tools/workflow.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "SafeOps Semantic Impact Analysis",
				Description: "Impact analysis now uses Context-Singularity for semantic dependency tracing. Risk scores (0-100) based on actual dependents, not just file counts.",
				Category:    "safety",
				Date:        "2026-02-22",
				RelatedDocs: []string{"SafeOps integration"},
			},
			{
				Version:     "v4.0.0",
				Type:        "improvement",
				Title:       "Smart Context Compression",
				Description: "Context is now intelligently compressed: preserve (system prompt, requirements), compress (tool results, exploration), discard (duplicates, failed branches). 2x effective context window.",
				Category:    "performance",
				Date:        "2026-02-22",
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Symbol Index for Code Search",
				Description: "Fast semantic code search without embeddings using tree-sitter. Search by symbol name across Go, TypeScript, Python codebases.",
				Category:    "tools",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/agent/tools/symbol_index.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Multi-Agent Headless Workspace",
				Description: "Workspace abstraction for multi-agent execution. Chalkboard for shared state (SUPERCACHE-backed). Multiple workers with goroutine-based isolation.",
				Category:    "agent",
				Date:        "2026-02-22",
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "Vision/Multi-modal Input Pipeline",
				Description: "Support for analyzing screenshots, diagrams, and design mockups. Vision model integration for 'What's wrong with this UI?' type queries.",
				Category:    "agent",
				Date:        "2026-02-22",
				RelatedDocs: []string{"internal/agent/tools/vision.go"},
			},
			{
				Version:     "v4.0.0",
				Type:        "improvement",
				Title:       "MCP Health Monitoring + Auto-Restart",
				Description: "MCP servers now have health monitoring with 30s interval checks. Auto-restart with exponential backoff (1s, 2s, 4s, 8s, max 60s). Max 5 attempts before marking failed.",
				Category:    "performance",
				Date:        "2026-02-22",
			},
			{
				Version:     "v4.0.0",
				Type:        "feature",
				Title:       "SUPERCACHE Namespace Support",
				Description: "SUPERCACHE now supports namespaces for project isolation. Default namespace='global' (backwards compatible). Prevents key collisions between projects.",
				Category:    "performance",
				Date:        "2026-02-22",
			},
		},
	}
}

// GetChangelog returns the changelog, loading from config if available.
func GetChangelog() *Changelog {
	changelog := DefaultChangelog()

	// Check if changelog is defined in config
	// TODO: Load from floyd.json or CLAUDE.md if defined

	return changelog
}

// GetNewFeaturesSince returns features added since the specified version.
func GetNewFeaturesSince(sinceVersion string) []ChangelogEntry {
	changelog := GetChangelog()
	var newFeatures []ChangelogEntry

	for _, entry := range changelog.Entries {
		if isNewerVersion(entry.Version, sinceVersion) {
			newFeatures = append(newFeatures, entry)
		}
	}

	return newFeatures
}

// isNewerVersion compares two version strings (simplified).
func isNewerVersion(current, previous string) bool {
	// Simplified version comparison
	// In production, use a proper semver library
	return strings.HasPrefix(current, "v") && current > previous
}

// FormatForBoot formats the changelog for display during agent boot.
func FormatForBoot() string {
	changelog := GetChangelog()

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Floyd %s - New Features:\n", changelog.CurrentVersion))

	// Group by category
	categories := make(map[string][]ChangelogEntry)
	for _, entry := range changelog.Entries {
		categories[entry.Category] = append(categories[entry.Category], entry)
	}

	// Display in priority order
	priority := []string{"agent", "tools", "performance", "safety", "ui"}
	for _, cat := range priority {
		if entries, ok := categories[cat]; ok {
			summary.WriteString(fmt.Sprintf("\n  [%s]\n", strings.Title(cat)))
			for _, entry := range entries {
				indicator := "✓"
				if entry.Type == "feature" {
					indicator = "★"
				}
				summary.WriteString(fmt.Sprintf("    %s %s: %s\n", indicator, entry.Title, entry.Description))
			}
		}
	}

	return summary.String()
}

// FormatCompact returns a compact one-line summary of new features.
func FormatCompact() string {
	changelog := GetChangelog()
	featureCount := 0
	improvementCount := 0

	for _, entry := range changelog.Entries {
		if entry.Type == "feature" {
			featureCount++
		} else if entry.Type == "improvement" {
			improvementCount++
		}
	}

	return fmt.Sprintf("%d features, %d improvements", featureCount, improvementCount)
}

// GetBootSummary returns the boot summary for changelog.
func GetBootSummary() string {
	changelog := GetChangelog()
	return fmt.Sprintf("Changelog: %s - %s", changelog.CurrentVersion, FormatCompact())
}

// GetJSON returns the changelog as JSON.
func GetJSON() (string, error) {
	changelog := GetChangelog()
	data, err := json.MarshalIndent(changelog, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal changelog: %w", err)
	}
	return string(data), nil
}

// LogBoot logs the changelog at boot time.
func LogBoot() {
	slog.Info("Floyd changelog", "version", Version, "features", FormatCompact())
}
