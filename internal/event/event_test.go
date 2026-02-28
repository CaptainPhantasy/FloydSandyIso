package event

// These tests verify that the stubbed telemetry functions work correctly
// as no-ops without panicking.

import (
	"testing"
)

func TestError(t *testing.T) {
	t.Run("handles nil error without panicking", func(t *testing.T) {
		// Verify Error handles nil gracefully
		Error(nil)
	})

	t.Run("handles string error without panicking", func(t *testing.T) {
		Error("some error")
	})

	t.Run("handles error type without panicking", func(t *testing.T) {
		Error(newDefaultTestError("runtime error"), "key", "value")
	})

	t.Run("handles error with properties without panicking", func(t *testing.T) {
		Error("test error",
			"type", "test",
			"severity", "high",
			"source", "unit-test",
		)
	})
}

func TestNoOpFunctions(t *testing.T) {
	t.Run("all no-op functions can be called without panicking", func(t *testing.T) {
		// These should all be no-ops and not panic
		Init()
		Flush()
		Alias("test-id")
		SetNonInteractive(true)
		SessionCreated()
		SessionDeleted()
		SessionSwitched()
		FilePickerOpened()
		LSPStarted("gopls")
		LSPRestarted("gopls")
		MCPStarted("test-mcp")
		PromptSent("key", "value")
		PromptResponded("key", "value")
		TokensUsed("count", 100)
	})
}

func TestGetID(t *testing.T) {
	t.Run("returns local identifier", func(t *testing.T) {
		id := GetID()
		if id != "local" {
			t.Errorf("Expected GetID() = 'local', got '%s'", id)
		}
	})
}

func TestAppInitialized(t *testing.T) {
	t.Run("sets app start time", func(t *testing.T) {
		AppInitialized()
		// Just verify it doesn't panic - the appStartTime is set internally
	})
}

func TestAppExited(t *testing.T) {
	t.Run("handles exit without panicking", func(t *testing.T) {
		AppInitialized() // Ensure start time is set
		AppExited()      // Should not panic
	})
}

// newDefaultTestError creates a test error that mimics runtime panic
// errors. This helps us testing that the Error function can handle various
// error types, including those that might be passed from a panic recovery
// scenario.
func newDefaultTestError(s string) error {
	return testError(s)
}

type testError string

func (e testError) Error() string {
	return string(e)
}
