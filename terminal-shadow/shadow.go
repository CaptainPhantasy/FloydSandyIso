// Package shadow provides Go integration for the Terminal Shadow Python harness.
// This allows Floyd to automatically log significant command executions.
package shadow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CommandLog represents a command execution to be logged.
type CommandLog struct {
	Command    string `json:"command"`
	ExitCode   int    `json:"exit_code"`
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	WorkingDir string `json:"working_dir"`
	DurationMs int    `json:"duration_ms"`
}

// ShadowConfig holds configuration for the shadow integration.
type ShadowConfig struct {
	Enabled     bool
	ProjectPath string
	HookPath    string // Path to floyd_shadow_hook.py
}

// DefaultShadowConfig returns default configuration.
func DefaultShadowConfig(projectPath string) *ShadowConfig {
	return &ShadowConfig{
		Enabled:     false,
		ProjectPath: projectPath,
		HookPath:    filepath.Join(projectPath, "terminal-shadow", "floyd_shadow_hook.py"),
	}
}

// Logger handles shadow logging operations.
type Logger struct {
	config *ShadowConfig
}

// NewLogger creates a new shadow logger.
func NewLogger(config *ShadowConfig) *Logger {
	return &Logger{config: config}
}

// LogCommand logs a command execution to the shadow system.
// This is a non-blocking operation that spawns a goroutine.
func (l *Logger) LogCommand(log CommandLog) {
	if !l.config.Enabled {
		return
	}

	// Run in goroutine to not block command execution
	go l.logCommandSync(log)
}

// LogCommandSync logs a command execution synchronously.
func (l *Logger) logCommandSync(log CommandLog) error {
	if !l.config.Enabled {
		return nil
	}

	// Check if hook exists
	if _, err := os.Stat(l.config.HookPath); os.IsNotExist(err) {
		return fmt.Errorf("shadow hook not found: %s", l.config.HookPath)
	}

	// Prepare command
	cmd := exec.Command("python3", l.config.HookPath, "log",
		"--project", l.config.ProjectPath,
		"--command", log.Command,
		"--exit-code", fmt.Sprintf("%d", log.ExitCode),
		"--duration-ms", fmt.Sprintf("%d", log.DurationMs),
	)

	// Pass stdout/stderr via env to avoid command line length limits
	// For large outputs, we truncate to prevent issues
	if len(log.Stdout) > 8000 {
		log.Stdout = log.Stdout[:8000] + "\n... (truncated)"
	}
	if len(log.Stderr) > 8000 {
		log.Stderr = log.Stderr[:8000] + "\n... (truncated)"
	}

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SHADOW_STDOUT=%s", log.Stdout),
		fmt.Sprintf("SHADOW_STDERR=%s", log.Stderr),
	)

	// Use environment variables for large outputs
	if log.Stdout != "" {
		cmd.Args = append(cmd.Args, "--stdout", truncate(log.Stdout, 2000))
	}
	if log.Stderr != "" {
		cmd.Args = append(cmd.Args, "--stderr", truncate(log.Stderr, 2000))
	}
	if log.WorkingDir != "" {
		cmd.Args = append(cmd.Args, "--working-dir", log.WorkingDir)
	}

	// Run with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	return cmd.Run()
}

// ShouldLog determines if a command result should be logged.
// This is a quick check that doesn't require Python.
func ShouldLog(command string, exitCode int) bool {
	// Always log errors
	if exitCode != 0 {
		return true
	}

	// Check for significant commands
	significantCmds := []string{
		"git commit", "git push", "git merge",
		"go build", "go test", "go run",
		"npm test", "npm run build",
		"pytest", "cargo build", "cargo test",
		"make", "docker build", "docker push",
	}

	cmdLower := strings.ToLower(command)
	for _, sig := range significantCmds {
		if strings.Contains(cmdLower, strings.ToLower(sig)) {
			return true
		}
	}

	// Skip trivial commands
	trivialCmds := []string{
		"ls", "cd", "pwd", "clear", "cat",
		"head", "tail", "less", "echo", "which",
	}

	cmdBase := strings.Fields(command)
	if len(cmdBase) > 0 {
		for _, trivial := range trivialCmds {
			if strings.ToLower(cmdBase[0]) == trivial {
				return false
			}
		}
	}

	return false
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// IsEnabled checks if shadow is enabled for the project.
func IsEnabled(projectPath string) bool {
	// Check for shadow config file
	configPath := filepath.Join(projectPath, "shadow_config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	// Check for .shadow_enabled marker
	markerPath := filepath.Join(projectPath, ".shadow_enabled")
	if _, err := os.Stat(markerPath); err == nil {
		return true
	}

	return false
}

// Summary represents a summary of shadow activity.
type Summary struct {
	TotalErrors   int      `json:"total_errors"`
	TotalSuccess  int      `json:"total_success"`
	RecentErrors  []string `json:"recent_errors"`
	RecentSuccess []string `json:"recent_success"`
}

// GetSummary extracts a summary from HANDOFF.md.
func GetSummary(projectPath string) (*Summary, error) {
	handoffPath := filepath.Join(projectPath, "HANDOFF.md")
	content, err := os.ReadFile(handoffPath)
	if err != nil {
		return nil, err
	}

	summary := &Summary{}

	// Count errors and successes (simplified parsing)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "### ⚠ Error:") {
			summary.TotalErrors++
			if len(summary.RecentErrors) < 5 {
				// Extract error title
				start := strings.Index(line, "Error:")
				if start != -1 {
					summary.RecentErrors = append(summary.RecentErrors,
						strings.TrimSpace(line[start+6:]))
				}
			}
		}
		if strings.Contains(line, "### ✓") {
			summary.TotalSuccess++
			if len(summary.RecentSuccess) < 5 {
				// Extract success title
				start := strings.Index(line, "✓")
				if start != -1 {
					summary.RecentSuccess = append(summary.RecentSuccess,
						strings.TrimSpace(line[start+1:]))
				}
			}
		}
	}

	return summary, nil
}

// ToJSON returns the summary as JSON.
func (s *Summary) ToJSON() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}
