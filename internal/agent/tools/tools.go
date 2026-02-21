package tools

import (
	"context"
)

type (
	sessionIDContextKey      string
	messageIDContextKey      string
	supportsImagesKey        string
	modelNameKey             string
	progressCallbackCtxKey   string
)

const (
	// SessionIDContextKey is the key for the session ID in the context.
	SessionIDContextKey sessionIDContextKey = "session_id"
	// MessageIDContextKey is the key for the message ID in the context.
	MessageIDContextKey messageIDContextKey = "message_id"
	// SupportsImagesContextKey is the key for the model's image support capability.
	SupportsImagesContextKey supportsImagesKey = "supports_images"
	// ModelNameContextKey is the key for the model name in the context.
	ModelNameContextKey modelNameKey = "model_name"
	// ProgressCallbackContextKey is the key for the tool progress callback in the context.
	ProgressCallbackContextKey progressCallbackCtxKey = "progress_callback"
)

// GetSessionFromContext retrieves the session ID from the context.
func GetSessionFromContext(ctx context.Context) string {
	sessionID := ctx.Value(SessionIDContextKey)
	if sessionID == nil {
		return ""
	}
	s, ok := sessionID.(string)
	if !ok {
		return ""
	}
	return s
}

// GetMessageFromContext retrieves the message ID from the context.
func GetMessageFromContext(ctx context.Context) string {
	messageID := ctx.Value(MessageIDContextKey)
	if messageID == nil {
		return ""
	}
	s, ok := messageID.(string)
	if !ok {
		return ""
	}
	return s
}

// GetSupportsImagesFromContext retrieves whether the model supports images from the context.
func GetSupportsImagesFromContext(ctx context.Context) bool {
	supportsImages := ctx.Value(SupportsImagesContextKey)
	if supportsImages == nil {
		return false
	}
	if supports, ok := supportsImages.(bool); ok {
		return supports
	}
	return false
}

// GetModelNameFromContext retrieves the model name from the context.
func GetModelNameFromContext(ctx context.Context) string {
	modelName := ctx.Value(ModelNameContextKey)
	if modelName == nil {
		return ""
	}
	s, ok := modelName.(string)
	if !ok {
		return ""
	}
	return s
}

// GetProgressCallbackFromContext retrieves a progress callback bound to a specific tool call ID.
// Returns nil if no callback is set in the context.
func GetProgressCallbackFromContext(ctx context.Context) func(ToolProgressEvent) {
	cb := ctx.Value(ProgressCallbackContextKey)
	if cb == nil {
		return nil
	}
	callback, ok := cb.(func(ToolProgressEvent))
	if !ok {
		return nil
	}
	return callback
}
