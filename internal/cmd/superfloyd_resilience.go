package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/db"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/model"
)

type runQualityGateStatus struct {
	Enabled    bool
	Applied    bool
	Reason     string
	Checks     []string
	Violations []string
}

type runtimeFailure struct {
	Hash      string `json:"hash"`
	Message   string `json:"message"`
	Occurred  int64  `json:"occurred"`
	ExitClass string `json:"exit_class"`
}

type runtimeHealthState struct {
	Failures []runtimeFailure `json:"failures"`
}

func qualityGatesEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SUPERFLOYD_QUALITY_GATES")))
	if v == "" {
		return true
	}
	return v != "0" && v != "false" && v != "off"
}

func degradationControlsEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SUPERFLOYD_DEGRADATION_CONTROLS")))
	if v == "" {
		return true
	}
	return v != "0" && v != "false" && v != "off"
}

func consistencyLockEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SUPERFLOYD_CONSISTENCY_LOCK")))
	if v == "" {
		return true
	}
	return v != "0" && v != "false" && v != "off"
}

func autoStabilizeEnabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("SUPERFLOYD_AUTOSTABILIZE")))
	if v == "" {
		return true
	}
	return v != "0" && v != "false" && v != "off"
}

func isSuperFloydBinary() bool {
	name := strings.ToLower(strings.TrimSpace(filepathBase(os.Args[0])))
	return strings.Contains(name, "superfloyd")
}

func applyRunQualityGates(prompt string) runQualityGateStatus {
	status := runQualityGateStatus{Enabled: qualityGatesEnabled()}
	if !status.Enabled {
		status.Reason = "disabled via SUPERFLOYD_QUALITY_GATES"
		return status
	}

	p := strings.TrimSpace(prompt)
	status.Checks = append(status.Checks,
		"prompt_non_empty",
		"prompt_has_min_signal",
		"prompt_not_oversized_without_autostabilize",
	)

	if p == "" {
		status.Violations = append(status.Violations, "prompt is empty")
	}
	if len([]rune(p)) < 5 {
		status.Violations = append(status.Violations, "prompt too short for meaningful execution")
	}
	if len([]rune(p)) > 200000 {
		status.Violations = append(status.Violations, "prompt exceeds hard safety size")
	}

	status.Applied = true
	if len(status.Violations) > 0 {
		status.Reason = "quality gate violation"
	} else {
		status.Reason = "passed"
	}
	return status
}

func enforceConsistencyLock(cfg *config.Config) error {
	if !consistencyLockEnabled() || !isSuperFloydBinary() {
		return nil
	}

	if model.AcceptSuggestionPrimaryBinding != "`" {
		return fmt.Errorf("consistency lock failed: accept-suggestion binding drifted from `")
	}

	if cfg == nil {
		return fmt.Errorf("consistency lock failed: nil config")
	}
	if _, ok := cfg.MCP["floyd-supercache-server"]; !ok {
		if _, fallback := cfg.MCP["floyd-supercache"]; !fallback {
			return fmt.Errorf("consistency lock failed: required MCP config missing (floyd-supercache-server)")
		}
	}

	bootPath := filepath.Join(cfg.WorkingDir(), "FLOYD.md")
	if b, err := os.ReadFile(bootPath); err == nil {
		if !strings.Contains(string(b), "I am FLOYD v4.6.1") {
			return fmt.Errorf("consistency lock failed: boot contract drift in %s", bootPath)
		}
	}

	return nil
}

func applyAutoStabilizeIfNeeded(ctx context.Context, cfg *config.Config, prompt string) string {
	if !isSuperFloydBinary() || !autoStabilizeEnabled() || !degradationControlsEnabled() {
		return prompt
	}

	trimmed := strings.TrimSpace(prompt)
	if len(trimmed) == 0 {
		return prompt
	}

	maxRunes := 12000
	runes := []rune(trimmed)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "\n\n[superfloyd-auto-stabilize] Prompt was truncated to maintain reliability under high-load context conditions."
	}

	if cfg != nil {
		if shouldStabilizeFromBenchmarks(ctx, cfg) {
			return trimmed + "\n\n[superfloyd-auto-stabilize] Use concise responses, deterministic steps, and minimal speculative branching."
		}
	}

	return prompt
}

func shouldStabilizeFromBenchmarks(ctx context.Context, cfg *config.Config) bool {
	if cfg == nil {
		return false
	}
	conn, err := db.Connect(ctx, cfg.Options.DataDirectory)
	if err != nil {
		return false
	}
	defer conn.Close()

	stats, err := gatherStats(ctx, conn)
	if err != nil {
		return false
	}
	if stats.Total.TotalSessions < 10 {
		return false
	}
	// Trigger stabilize mode on pressure signals.
	if stats.AvgResponseTimeMs >= 22000 || stats.Total.AvgTokensPerSession >= 90000 {
		return true
	}
	return false
}

func enforceRetryBudget(dataDir string) error {
	if !isSuperFloydBinary() || !degradationControlsEnabled() {
		return nil
	}
	state, _ := loadRuntimeHealth(dataDir)
	now := time.Now().Unix()
	state.Failures = keepRecentFailures(state.Failures, now-3600)
	if len(state.Failures) >= 6 {
		return fmt.Errorf("retry budget exceeded: %d failures in last hour; stabilize before retrying", len(state.Failures))
	}
	return nil
}

func maybeTripCircuitBreaker(dataDir string, runErr error) error {
	if runErr == nil || !isSuperFloydBinary() || !degradationControlsEnabled() {
		return runErr
	}

	state, _ := loadRuntimeHealth(dataDir)
	now := time.Now().Unix()
	state.Failures = keepRecentFailures(state.Failures, now-600)

	msg := strings.TrimSpace(runErr.Error())
	hash := failureHash(msg)
	state.Failures = append(state.Failures, runtimeFailure{
		Hash:      hash,
		Message:   msg,
		Occurred:  now,
		ExitClass: "run",
	})
	_ = saveRuntimeHealth(dataDir, state)

	hits := 0
	for _, f := range state.Failures {
		if f.Hash == hash {
			hits++
		}
	}
	if hits >= 2 {
		return fmt.Errorf("circuit breaker engaged for repeated run failure (%s). gather new observation before retry", hash)
	}
	return runErr
}

func recordRunSuccess(dataDir string) {
	state, _ := loadRuntimeHealth(dataDir)
	state.Failures = nil
	_ = saveRuntimeHealth(dataDir, state)
}

func loadRuntimeHealth(dataDir string) (runtimeHealthState, error) {
	path := filepath.Join(dataDir, "runtime_health.json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return runtimeHealthState{}, nil
		}
		return runtimeHealthState{}, err
	}
	var st runtimeHealthState
	if err := json.Unmarshal(b, &st); err != nil {
		return runtimeHealthState{}, nil
	}
	return st, nil
}

func saveRuntimeHealth(dataDir string, st runtimeHealthState) error {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dataDir, "runtime_health.json")
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func keepRecentFailures(items []runtimeFailure, cutoff int64) []runtimeFailure {
	out := make([]runtimeFailure, 0, len(items))
	for _, it := range items {
		if it.Occurred >= cutoff {
			out = append(out, it)
		}
	}
	return out
}

func failureHash(msg string) string {
	h := crc32.ChecksumIEEE([]byte(strings.ToLower(strings.TrimSpace(msg))))
	return fmt.Sprintf("%08x", h)
}

func filepathBase(path string) string {
	if path == "" {
		return ""
	}
	i := strings.LastIndex(path, "/")
	if i >= 0 && i+1 < len(path) {
		return path[i+1:]
	}
	return path
}
