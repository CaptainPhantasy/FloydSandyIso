package tools

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/session"
)

const ContextStatusToolName = "context_status"

type ContextStatusParams struct{}

type ContextStatusResponse struct {
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	CacheReadTokens  int64   `json:"cache_read_tokens"`
	EffectiveTokens  int64   `json:"effective_tokens"`  // prompt + completion - cached
	ContextWindow    int64   `json:"context_window"`
	PercentUsed      float64 `json:"percent_used"`
	RemainingTokens  int64   `json:"remaining_tokens"`
	ShouldSummarize  bool    `json:"should_summarize"`  // approaching limit
	SessionID        string  `json:"session_id"`
}

func NewContextStatusTool(sessions session.Service, contextWindow int64) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ContextStatusToolName,
		"Returns current context window usage statistics. Use this to monitor your context consumption and know when to be more concise or when you have room for detailed responses.",
		func(ctx context.Context, params ContextStatusParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.NewTextErrorResponse("no active session"), nil
			}

			sess, err := sessions.Get(ctx, sessionID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to get session: %v", err)), nil
			}

			effectiveTokens := sess.PromptTokens + sess.CompletionTokens - sess.CacheReadTokens
			remaining := contextWindow - effectiveTokens
			percentUsed := 0.0
			if contextWindow > 0 {
				percentUsed = (float64(effectiveTokens) / float64(contextWindow)) * 100
			}

			// Warn if approaching limit (80%+)
			shouldSummarize := percentUsed >= 80.0

			status := ContextStatusResponse{
				PromptTokens:     sess.PromptTokens,
				CompletionTokens: sess.CompletionTokens,
				CacheReadTokens:  sess.CacheReadTokens,
				EffectiveTokens:  effectiveTokens,
				ContextWindow:    contextWindow,
				PercentUsed:      percentUsed,
				RemainingTokens:  remaining,
				ShouldSummarize:  shouldSummarize,
				SessionID:        sessionID,
			}

			// Human-readable summary
			var summary string
			if shouldSummarize {
				summary = fmt.Sprintf(
					"⚠️ CONTEXT WARNING: %.1f%% used (%d/%d tokens). %d remaining. Consider being more concise.\n"+
					"Cached: %d tokens | Prompt: %d | Completion: %d",
					percentUsed, effectiveTokens, contextWindow, remaining,
					sess.CacheReadTokens, sess.PromptTokens, sess.CompletionTokens,
				)
			} else {
				summary = fmt.Sprintf(
					"Context: %.1f%% used (%d/%d tokens). %d remaining.\n"+
					"Cached: %d tokens | Prompt: %d | Completion: %d",
					percentUsed, effectiveTokens, contextWindow, remaining,
					sess.CacheReadTokens, sess.PromptTokens, sess.CompletionTokens,
				)
			}

			return fantasy.WithResponseMetadata(fantasy.NewTextResponse(summary), status), nil
		})
}
