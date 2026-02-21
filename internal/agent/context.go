package agent

import (
	"strings"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/message"
)

// MessageTier represents the importance level of a message for summarization
type MessageTier int

const (
	// Tier1Preserve - Never summarize, always keep full content
	// System prompt, first user request, explicit requirements
	Tier1Preserve MessageTier = 1

	// Tier2Compress - Intelligently compress while preserving key info
	// Tool calls/results, exploration, decisions
	Tier2Compress MessageTier = 2

	// Tier3Discard - Can be discarded during summarization
	// Duplicates, verbose output, failed branches
	Tier3Discard MessageTier = 3
)

// TieredContext holds messages organized by tier
type TieredContext struct {
	Preserve []message.Message
	Compress []message.Message
	Discard  []message.Message

	// Metadata for summarization
	ToolSummaries  []ToolSummary
	FileOperations []FileOperation
	KeyDecisions   []string
}

// ToolSummary captures essential info from tool interactions
type ToolSummary struct {
	Name    string
	Action  string
	Outcome string
}

// FileOperation tracks file read/write operations
type FileOperation struct {
	Path    string
	Action  string
	Summary string
}

// preserveKeywords indicate user messages with explicit requirements
var preserveKeywords = []string{
	"requirement", "must", "need to", "important", "critical",
	"don't", "do not", "never", "always", "essential",
	"constraint", "requirement", "spec", "specification",
}

// ClassifyMessage determines the tier for a message
func ClassifyMessage(msg message.Message, isFirstUser bool) MessageTier {
	// TIER 1: First user message (original request)
	if isFirstUser && msg.Role == message.User {
		return Tier1Preserve
	}

	// TIER 1: User messages with explicit requirements
	if msg.Role == message.User {
		text := strings.ToLower(msg.Content().Text)
		for _, kw := range preserveKeywords {
			if strings.Contains(text, kw) {
				return Tier1Preserve
			}
		}
	}

	// TIER 3: Error tool results (keep resolution, discard details)
	if msg.Role == message.Tool {
		results := msg.ToolResults()
		for _, result := range results {
			if result.IsError {
				return Tier3Discard
			}
		}

		// TIER 3: Very long tool results (>3000 chars = verbose output)
		for _, result := range results {
			if len(result.Content) > 3000 {
				return Tier3Discard
			}
		}
	}

	// TIER 2: Everything else (tool calls, normal messages, etc.)
	return Tier2Compress
}

// PrepareTieredContext organizes messages by tier
func PrepareTieredContext(msgs []message.Message) *TieredContext {
	tc := &TieredContext{
		Preserve: make([]message.Message, 0),
		Compress: make([]message.Message, 0),
		Discard:  make([]message.Message, 0),
	}

	for i, msg := range msgs {
		tier := ClassifyMessage(msg, i == 0)

		switch tier {
		case Tier1Preserve:
			tc.Preserve = append(tc.Preserve, msg)
		case Tier2Compress:
			tc.Compress = append(tc.Compress, msg)
		case Tier3Discard:
			tc.Discard = append(tc.Discard, msg)
		}

		// Extract tool summaries from TIER 2 messages
		if tier == Tier2Compress {
			tc.extractToolInfo(msg)
		}
	}

	return tc
}

// extractToolInfo extracts tool operation summaries from a message
func (tc *TieredContext) extractToolInfo(msg message.Message) {
	// Extract from assistant messages with tool calls
	if msg.Role == message.Assistant {
		calls := msg.ToolCalls()
		for _, call := range calls {
			tc.ToolSummaries = append(tc.ToolSummaries, ToolSummary{
				Name: call.Name,
			})
		}
	}

	// Extract file operations from tool results
	if msg.Role == message.Tool {
		results := msg.ToolResults()
		for _, result := range results {
			if isFileTool(result.Name) {
				tc.FileOperations = append(tc.FileOperations, FileOperation{
					Path:   extractFilePath(result.Content),
					Action: inferFileAction(result.Name),
				})
			}
		}
	}
}

// isFileTool checks if a tool name relates to file operations
func isFileTool(name string) bool {
	fileTools := []string{"view", "write", "edit", "multiedit", "glob", "grep", "ls"}
	nameLower := strings.ToLower(name)
	for _, t := range fileTools {
		if strings.Contains(nameLower, t) {
			return true
		}
	}
	return false
}

// extractFilePath attempts to extract a file path from tool output
func extractFilePath(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for common path patterns
		if strings.HasPrefix(line, "/") || strings.HasPrefix(line, "./") || strings.HasPrefix(line, "~/") {
			// Truncate to reasonable length
			if len(line) > 100 {
				return line[:100]
			}
			return line
		}
	}
	if len(lines) > 0 && len(lines[0]) < 100 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

// inferFileAction infers the action type from tool name
func inferFileAction(name string) string {
	nameLower := strings.ToLower(name)
	switch {
	case strings.Contains(nameLower, "write"):
		return "created"
	case strings.Contains(nameLower, "edit"):
		return "modified"
	case strings.Contains(nameLower, "view"):
		return "read"
	case strings.Contains(nameLower, "glob") || strings.Contains(nameLower, "grep"):
		return "searched"
	case strings.Contains(nameLower, "ls"):
		return "listed"
	default:
		return "accessed"
	}
}

// GetStats returns statistics about the tiered context
func (tc *TieredContext) GetStats() (preserve, compress, discard int) {
	return len(tc.Preserve), len(tc.Compress), len(tc.Discard)
}
