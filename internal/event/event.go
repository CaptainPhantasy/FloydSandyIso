package event

// Telemetry is disabled for this isolated instance.
// All functions are no-ops to eliminate telemetry overhead while maintaining API compatibility.

import (
	"time"
)

var (
	distinctId  = "local"
	appStartTime time.Time
)

// GetID returns a local identifier (no telemetry).
func GetID() string { return distinctId }

// Init is a no-op.
func Init() {}

// SetNonInteractive is a no-op.
func SetNonInteractive(_ bool) {}

// Flush is a no-op.
func Flush() {}

// Alias is a no-op.
func Alias(_ string) {}

// Error is a no-op.
func Error(_ any, _ ...any) {}

// send is a no-op (internal).
func send(_ string, _ ...any) {}

// AppInitialized records app start time locally (no telemetry).
func AppInitialized() {
	appStartTime = time.Now()
}

// AppExited is a no-op.
func AppExited() {}

// SessionCreated is a no-op.
func SessionCreated() {}

// SessionDeleted is a no-op.
func SessionDeleted() {}

// SessionSwitched is a no-op.
func SessionSwitched() {}

// FilePickerOpened is a no-op.
func FilePickerOpened() {}

// LSPStarted is a no-op.
func LSPStarted(_ string) {}

// LSPRestarted is a no-op.
func LSPRestarted(_ string) {}

// MCPStarted is a no-op.
func MCPStarted(_ string) {}

// PromptSent is a no-op.
func PromptSent(_ ...any) {}

// PromptResponded is a no-op.
func PromptResponded(_ ...any) {}

// TokensUsed is a no-op.
func TokensUsed(_ ...any) {}
