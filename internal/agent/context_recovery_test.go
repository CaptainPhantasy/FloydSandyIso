package agent

import (
	"testing"

	"charm.land/fantasy"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/message"
	"github.com/stretchr/testify/require"
)

func TestCalculateSemanticScoreFlagsMatchedCategory(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		check func(t *testing.T, score SemanticScore)
	}{
		{
			name: "code location",
			text: "the failure is in internal/agent/context.go:112",
			check: func(t *testing.T, score SemanticScore) {
				require.True(t, score.HasCodeLocation)
			},
		},
		{
			name: "decision",
			text: "we decided to preserve the summary because it contains the rationale",
			check: func(t *testing.T, score SemanticScore) {
				require.True(t, score.HasDecision)
			},
		},
		{
			name: "fix outcome",
			text: "implemented the fix successfully",
			check: func(t *testing.T, score SemanticScore) {
				require.True(t, score.HasFixOutcome)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := message.Message{
				Role:  message.User,
				Parts: []message.ContentPart{message.TextContent{Text: tt.text}},
			}
			tt.check(t, CalculateSemanticScore(msg, 0, 1))
		})
	}
}

func TestSubAgentBypassesRateLimitConfiguration(t *testing.T) {
	agent := &sessionAgent{isSubAgent: true}
	called := false

	result, err := agent.runWithRateLimit("test-provider", func() (*fantasy.AgentResult, error) {
		called = true
		return &fantasy.AgentResult{}, nil
	})

	require.NoError(t, err)
	require.True(t, called)
	require.NotNil(t, result)
}
