🚨 DETERMINISTIC EXECUTION MANDATE

NO SUBSTITUTIONS: If a task specifies Code A, do not use Code B.
NO BATCHING: You are forbidden from completing Task 2 before Task 1 is verified.
NO SUMMARIZATION: Do not explain your "understanding" of the plan. Start the first task immediately.
FAILURE IS TERMINAL: If a verification command fails, you MUST stop. Do not attempt a "workaround".

---

# Session Continuity Architecture: Implementation Plan

**Created:** 2026-02-28
**Status:** Ready for Execution
**Project:** FloydDeployable v4.7

---

## ARCHITECTURAL CONTEXT

### The Bug Discovered
The `createHandoffFile` implementation in `internal/agent/agent.go` **OVERWRITES** the entire `HANDOFF.md` file with `os.WriteFile`, destroying all session logs accumulated by Terminal Shadow.

### The Two Systems

| System | Write Method | Content | Context Awareness |
|--------|--------------|---------|-------------------|
| Terminal Shadow | APPENDS to sections | Errors, builds, decisions, heartbeats | None |
| createHandoffFile | OVERWRITES entire file | session_id, todos, reason | 95% threshold |

### Required Fix
Change `createHandoffFile` from OVERWRITE to APPEND, adding a `## SESSION HANDOFF` section to the existing `HANDOFF.md` without destroying Shadow's logs.

---

## LINEAR TASK ARCHITECTURE

---

### TASK 1 of 5: Create Handoff Section Appender

**PREREQUISITE:** None (first task)

**TARGET FILE:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/internal/agent/agent.go`

**INPUT SPEC:** Replace the `createHandoffFile` method (currently at lines ~1186-1230) with the following implementation that APPENDS instead of OVERWRITES:

```go
// createHandoffFile appends a SESSION HANDOFF section to the existing HANDOFF.md.
// This preserves all Terminal Shadow logs while adding the session pointer.
func (a *sessionAgent) createHandoffFile(ctx context.Context, sessionID string) error {
	sess, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session for handoff: %w", err)
	}

	handoffPath := filepath.Join(a.cfg.WorkingDir(), "HANDOFF.md")

	// Build the handoff section (to be appended, not replace)
	var sb strings.Builder

	sb.WriteString("\n---\n\n")
	sb.WriteString("## SESSION HANDOFF\n\n")
	sb.WriteString(fmt.Sprintf("**Previous Session ID:** %s\n", sessionID))
	sb.WriteString(fmt.Sprintf("**Session Title:** %s\n", sess.Title))
	sb.WriteString(fmt.Sprintf("**Reason:** Context window threshold reached (95%%).\n"))
	sb.WriteString(fmt.Sprintf("**Timestamp:** %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	sb.WriteString("### Active Todos\n\n")
	if len(sess.Todos) == 0 {
		sb.WriteString("*No active todos*\n\n")
	} else {
		for _, todo := range sess.Todos {
			statusIcon := "[ ]"
			if todo.Status == session.TodoStatusCompleted {
				statusIcon = "[x]"
			} else if todo.Status == session.TodoStatusInProgress {
				statusIcon = "[~]"
			}
			sb.WriteString(fmt.Sprintf("- %s %s\n", statusIcon, todo.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Agent Instruction\n\n")
	sb.WriteString("Upon starting the new session, immediately use `query_floyd_archive` to retrieve the technical context of the last task worked on in this session.\n")

	// APPEND to existing file, do not overwrite
	f, err := os.OpenFile(handoffPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open HANDOFF.md for append: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(sb.String()); err != nil {
		return fmt.Errorf("failed to append handoff section: %w", err)
	}

	slog.Info("Appended handoff section to HANDOFF.md", "path", handoffPath, "session_id", sessionID)
	return nil
}
```

**VERIFICATION:**
```bash
go build ./internal/agent/... && echo "BUILD_OK" || echo "BUILD_FAIL"
```

**EVIDENCE:** Paste the updated `createHandoffFile` method (lines ~1186-1240).

**STOP. Do not proceed to Task 2.**
**Provide the Evidence requested above.**
**Provide the Verification command output.**
**WAIT for the user to type "PROCEED".**

---

### TASK 2 of 5: Add SearchTechnicalArchive to Message Service Interface

**PREREQUISITE:** Verify Task 1 exists.

```bash
grep -n "os.O_APPEND" /Volumes/Storage/floyd-sandbox/FloydDeployable/internal/agent/agent.go && echo "PREREQ_OK" || exit 1
```

IF "PREREQ_OK" IS NOT RETURNED, STOP AND COMPLETE TASK 1 FIRST.

**TARGET FILE:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/internal/message/service.go`

**INPUT SPEC:** Add the following method to the `Service` interface:

```go
// SearchTechnicalArchive queries the message database with a semantic filter.
// It returns only technical data (tool calls, tool results, code blocks) and
// excludes conversational text to prevent persona drift.
SearchTechnicalArchive(ctx context.Context, projectPath, query string, limit int) ([]ArchiveResult, error)
```

Also add the `ArchiveResult` struct:

```go
// ArchiveResult represents a single result from the technical archive query.
type ArchiveResult struct {
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ToolCalls  string    `json:"tool_calls,omitempty"`
	CreatedAt  int64     `json:"created_at"`
	SessionID  string    `json:"session_id"`
}
```

**VERIFICATION:**
```bash
grep -n "SearchTechnicalArchive" /Volumes/Storage/floyd-sandbox/FloydDeployable/internal/message/service.go && echo "PREREQ_OK" || echo "BUILD_FAIL"
```

**EVIDENCE:** Paste the Service interface showing the new method signature.

**STOP. Do not proceed to Task 3.**
**Provide the Evidence requested above.**
**Provide the Verification command output.**
**WAIT for the user to type "PROCEED".**

---

### TASK 3 of 5: Implement SearchTechnicalArchive in Message Service

**PREREQUISITE:** Verify Task 2 exists.

```bash
grep -n "SearchTechnicalArchive" /Volumes/Storage/floyd-sandbox/FloydDeployable/internal/message/service.go && echo "PREREQ_OK" || exit 1
```

IF "PREREQ_OK" IS NOT RETURNED, STOP AND COMPLETE TASK 2 FIRST.

**TARGET FILE:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/internal/message/service.go`

**INPUT SPEC:** Implement the `SearchTechnicalArchive` method in the `service` struct with the semantic firewall SQL query:

```go
func (s *service) SearchTechnicalArchive(ctx context.Context, projectPath, query string, limit int) ([]ArchiveResult, error) {
	if limit == 0 {
		limit = 5
	}

	// Semantic firewall SQL: Only return technical data
	sqlQuery := `
		SELECT m.role, m.parts, m.tool_calls, m.created_at, m.session_id
		FROM messages m
		JOIN sessions s ON m.session_id = s.id
		WHERE s.id = ?
		AND (
			m.parts LIKE '%' || ? || '%'
			OR (m.tool_calls IS NOT NULL AND m.tool_calls LIKE '%' || ? || '%')
		)
		AND (
			-- 1. Assistant messages with tool calls
			(m.role = 'assistant' AND m.tool_calls IS NOT NULL AND m.tool_calls != '[]' AND m.tool_calls != 'null')
			OR
			-- 2. Tool results
			m.role = 'tool'
			OR
			-- 3. User messages with code blocks
			(m.role = 'user' AND m.parts LIKE '%```%')
		)
		ORDER BY m.created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, projectPath, query, query, limit)
	if err != nil {
		return nil, fmt.Errorf("archive query failed: %w", err)
	}
	defer rows.Close()

	var results []ArchiveResult
	for rows.Next() {
		var r ArchiveResult
		var parts, toolCalls sql.NullString
		if err := rows.Scan(&r.Role, &parts, &toolCalls, &r.CreatedAt, &r.SessionID); err != nil {
			continue
		}
		if parts.Valid {
			r.Content = parts.String
		}
		if toolCalls.Valid {
			r.ToolCalls = toolCalls.String
		}
		results = append(results, r)
	}

	return results, nil
}
```

**VERIFICATION:**
```bash
go build ./internal/message/... && echo "BUILD_OK" || echo "BUILD_FAIL"
```

**EVIDENCE:** Paste the complete `SearchTechnicalArchive` implementation.

**STOP. Do not proceed to Task 4.**
**Provide the Evidence requested above.**
**Provide the Verification command output.**
**WAIT for the user to type "PROCEED".**

---

### TASK 4 of 5: Create query_floyd_archive Tool

**PREREQUISITE:** Verify Task 3 exists.

```bash
grep -n "func (s \*service) SearchTechnicalArchive" /Volumes/Storage/floyd-sandbox/FloydDeployable/internal/message/service.go && echo "PREREQ_OK" || exit 1
```

IF "PREREQ_OK" IS NOT RETURNED, STOP AND COMPLETE TASK 3 FIRST.

**TARGET FILE:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/internal/agent/tools/archive_query.go`

**INPUT SPEC:** Create new file with the following content:

```go
package tools

import (
	"context"
	"fmt"
	"strings"

	"charm.land/fantasy"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/message"
)

const ArchiveQueryToolName = "query_floyd_archive"

type ArchiveQueryParams struct {
	Query string `json:"query" jsonschema:"description=The specific technical detail, tool execution, or code snippet to search for in past sessions."`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Max results to return. Default 5."`
}

func NewArchiveQueryTool(messages message.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		ArchiveQueryToolName,
		"Query the persistent database for past code, tool results, and technical decisions. Use this whenever you lack historical context from previous sessions.",
		func(ctx context.Context, params ArchiveQueryParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Query == "" {
				return fantasy.NewTextErrorResponse("Query cannot be empty"), nil
			}

			limit := params.Limit
			if limit == 0 {
				limit = 5
			}

			results, err := messages.SearchTechnicalArchive(ctx, workingDir, params.Query, limit)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Archive query failed: %v", err)), nil
			}

			if len(results) == 0 {
				return fantasy.NewTextResponse(fmt.Sprintf("ARCHIVE RESULT: No technical records found matching '%s'. Do not guess; acknowledge the lack of data.", params.Query)), nil
			}

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("ARCHIVE RESULTS FOR: '%s'\n\n", params.Query))

			for _, r := range results {
				sb.WriteString(fmt.Sprintf("--- [Session: %s | Role: %s] ---\n", r.SessionID, r.Role))
				if r.ToolCalls != "" {
					sb.WriteString(fmt.Sprintf("EXECUTED TOOL: %s\n", r.ToolCalls))
				}
				if r.Content != "" {
					content := r.Content
					if len(content) > 1000 {
						content = content[:1000] + "\n... (truncated)"
					}
					sb.WriteString(fmt.Sprintf("CONTENT:\n%s\n", content))
				}
				sb.WriteString("\n")
			}

			return fantasy.NewTextResponse(sb.String()), nil
		})
}
```

**VERIFICATION:**
```bash
go build ./internal/agent/tools/... && echo "BUILD_OK" || echo "BUILD_FAIL"
```

**EVIDENCE:** Paste the complete `archive_query.go` file contents.

**STOP. Do not proceed to Task 5.**
**Provide the Evidence requested above.**
**Provide the Verification command output.**
**WAIT for the user to type "PROCEED".**

---

### TASK 5 of 5: Register Archive Tool in Coordinator

**PREREQUISITE:** Verify Task 4 exists.

```bash
test -f /Volumes/Storage/floyd-sandbox/FloydDeployable/internal/agent/tools/archive_query.go && echo "PREREQ_OK" || exit 1
```

IF "PREREQ_OK" IS NOT RETURNED, STOP AND COMPLETE TASK 4 FIRST.

**TARGET FILE:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/internal/agent/coordinator.go`

**INPUT SPEC:** In the `buildTools` function, add the archive query tool registration:

```go
tools.NewArchiveQueryTool(c.messages, c.cfg.WorkingDir()),
```

Add this line after the existing tool registrations (around line 510-530).

**VERIFICATION:**
```bash
go build ./... && echo "BUILD_OK" || echo "BUILD_FAIL"
```

**EVIDENCE:** Paste the `buildTools` function showing the new tool registration.

**STOP. Do not proceed to next steps.**
**Provide the Evidence requested above.**
**Provide the Verification command output.**
**WAIT for the user to type "PROCEED".**

---

## POST-IMPLEMENTATION TASKS (Not Part of Linear Sequence)

These tasks are documented but not part of the deterministic execution sequence:

1. **System Prompt Integration** - Add archive rules to prompt template
2. **Boot Sequence Read** - Add logic to read HANDOFF.md on session start
3. **Shadow-to-Go Bridge** - Create mechanism for Go agent to notify Shadow of 95% threshold

---

## FAILURE POINTS CHECKLIST

- [x] Did I use the word "Phase"? (Changed to "Task X of N")
- [x] Is there any task without a Bash verification command? (All have verification)
- [x] Is there any task that doesn't require a code snippet as evidence? (All require evidence)
- [x] Did I allow the agent to "plan" instead of "execute"? (Added Prohibitions block)

---

*Document created following DETERMINISTIC PLANNING FRAMEWORK*
