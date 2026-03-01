# Session Continuity Architecture: Analysis & Revised Proposal

**Date:** 2026-02-28
**Status:** Architectural Review Required
**Project:** FloydDeployable v4.7

---

## Executive Summary

This report documents a critical architectural insight: **the original design for Phase 2 (Auto-Export) is unnecessary** because the data already exists. The session data is persisted in real-time to SQLite. What's actually needed is a mechanism for the agent to **query the existing database** with a semantic filter and **derive its own summary** during session initialization.

---

## Part 1: Current State Analysis

### 1.1 Data Already Exists

The Floyd system already maintains a complete, real-time log of all session activity:

```
.floyd/floyd.db (SQLite database)
├── sessions table
│   ├── id (TEXT PRIMARY KEY)
│   ├── parent_session_id (TEXT)
│   ├── title (TEXT)
│   ├── message_count (INTEGER)
│   ├── prompt_tokens (INTEGER)
│   ├── completion_tokens (INTEGER)
│   ├── cache_read_tokens (INTEGER)
│   ├── cost (REAL)
│   ├── summary_message_id (TEXT)
│   ├── todos (TEXT JSON)
│   ├── created_at (INTEGER Unix timestamp)
│   └── updated_at (INTEGER Unix timestamp)
│
└── messages table
    ├── id (TEXT PRIMARY KEY)
    ├── session_id (TEXT FK → sessions.id)
    ├── role (TEXT: 'user', 'assistant', 'tool')
    ├── parts (TEXT JSON - full message content)
    ├── model (TEXT)
    ├── provider (TEXT)
    ├── is_summary_message (INTEGER)
    ├── created_at (INTEGER)
    ├── updated_at (INTEGER)
    └── finished_at (INTEGER)
```

**Key Insight:** Every message, tool call, tool result, and code block is already persisted. No "export" operation is required for data preservation.

### 1.2 The Export Fallacy

The original `contextsidebarfinal.md` design called for:

```
Phase 2: Auto-Export System
- Trigger at 85% context usage
- Export to `.floyd/transcripts/{session_id}_{timestamp}.md`
- Include: session metadata, conversation summary, message log, tool executions index
```

**Why This Is Redundant:**
1. Data is already in the database - no export needed
2. Creating markdown files duplicates data that already exists
3. The transcript format is human-readable, but the agent can query structured data directly
4. Export at 85% is arbitrary - the trigger point should be higher (95%+) or session-end

### 1.3 What Actually Needs to Happen

When a new session starts and needs context from a previous session:

```
┌─────────────────────────────────────────────────────────────────┐
│  NEW SESSION BOOT SEQUENCE (Revised)                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   1. Agent initializes (standard boot)                          │
│          │                                                      │
│          ▼                                                      │
│   2. Check for HANDOFF.md in project root                      │
│          │                                                      │
│          ├── [EXISTS]                                           │
│          │       │                                              │
│          │       ▼                                              │
│          │   3a. Read HANDOFF.md (small pointer file)           │
│          │       - Contains: session_id, last task, todos       │
│          │       - Does NOT contain: full transcript            │
│          │       │                                              │
│          │       ▼                                              │
│          │   3b. Agent queries .floyd/floyd.db directly        │
│          │       using query_floyd_archive tool                 │
│          │       │                                              │
│          │       ▼                                              │
│          │   3c. Semantic filter applied:                       │
│          │       - INCLUDE: tool_calls, tool_results, code      │
│          │       - EXCLUDE: conversational text, persona drift  │
│          │       │                                              │
│          │       ▼                                              │
│          │   3d. Agent DERIVES ITS OWN SUMMARY                  │
│          │       from the queried data                          │
│          │       │                                              │
│          │       ▼                                              │
│          │   3e. Report: "Continuing from session {id}"         │
│          │       "Queried N tool executions from archive."      │
│          │                                                      │
│          └── [NOT EXISTS]                                       │
│                  │                                              │
│                  ▼                                              │
│              Standard fresh session                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Part 2: Key Architectural Changes

### 2.1 Original vs. Revised Design

| Aspect | Original Design | Revised Design |
|--------|-----------------|----------------|
| **Data Storage** | Export to markdown files | Query existing SQLite database |
| **Export Trigger** | 85% context (too early) | Not needed; optionally 95%+ for HANDOFF pointer |
| **Summary Generation** | Pre-generated at export | Agent derives its own at session start |
| **Data Access** | Read markdown transcript | Query database with semantic filter |
| **Token Cost** | Full transcript loaded | Only relevant data queried |

### 2.2 Why the Agent Should Derive Its Own Summary

1. **Context Efficiency:** The agent queries only what it needs, not the entire transcript
2. **Relevance:** The agent determines relevance based on current task, not pre-generated summary
3. **Freshness:** Summary is generated with current model's understanding, not a previous session's
4. **Flexibility:** Different continuation scenarios need different summaries

### 2.3 The Semantic Firewall

The critical innovation is the **semantic filter** that queries only technical data:

**What Gets Queried (INCLUDED):**
- Assistant messages with `tool_calls` (actual actions taken)
- Tool result messages (hard system output)
- User messages containing code blocks (``` delimiters)

**What Gets Filtered (EXCLUDED):**
- Assistant plain text without tool calls (conversational filler)
- User plain text without code blocks (conversational context)
- Any identity/persona reflections (prevents persona drift)

---

## Part 3: Semantic Filter Implementation (Proposal Code)

The following is the complete semantic filter implementation from `docs/ARCHIVE_SYSTEM_SPEC.md`:

### 3.1 SQL Query with Persona Firewall

```sql
SELECT
    m.role,
    m.content,
    m.tool_calls,
    m.created_at
FROM messages m
JOIN sessions s ON m.session_id = s.id
WHERE
    -- Project scoping: Only current project
    s.project_path = :project_path

    -- Text matching: Search for query term in content OR tool_calls
    AND (
        m.content LIKE '%' || :query || '%'
        OR (m.tool_calls IS NOT NULL AND m.tool_calls LIKE '%' || :query || '%')
    )

    -- PERSONA FIREWALL: Only return technical data
    AND (
        -- 1. Assistant messages that actually executed tools
        (
            m.role = 'assistant'
            AND m.tool_calls IS NOT NULL
            AND m.tool_calls != '[]'
            AND m.tool_calls != 'null'
        )
        OR
        -- 2. Tool results (hard system output)
        m.role = 'tool'
        OR
        -- 3. User messages containing code blocks
        (
            m.role = 'user'
            AND m.content LIKE '%```%'
        )
    )
ORDER BY m.created_at DESC
LIMIT :limit
```

### 3.2 Go Tool Implementation

```go
package tools

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "strings"
)

// ArchiveInput defines schema for LLM tool call
type ArchiveInput struct {
    Query string `json:"query" jsonschema:"description=The specific technical detail, tool execution, or code snippet to search for in past sessions."`
    Limit int    `json:"limit,omitempty" jsonschema:"description=Max results to return. Default 5."`
}

type ArchiveTool struct {
    db *sql.DB
}

func NewArchiveTool(db *sql.DB) *ArchiveTool {
    return &ArchiveTool{db: db}
}

func (t *ArchiveTool) Name() string {
    return "query_floyd_archive"
}

func (t *ArchiveTool) Description() string {
    return "Query the persistent database for past code, tool results, and technical decisions. Use this whenever you lack historical context."
}

func (t *ArchiveTool) Run(ctx context.Context, inputRaw string) (string, error) {
    var input ArchiveInput
    if err := json.Unmarshal([]byte(inputRaw), &input); err != nil {
        return "", fmt.Errorf("failed to parse input: %w", err)
    }

    if input.Query == "" {
        return "ARCHIVE ERROR: Query cannot be empty", nil
    }

    limit := input.Limit
    if limit == 0 {
        limit = 5
    }

    // Get current project path from context
    projectPath := "/Volumes/Storage/floyd-sandbox/FloydDeployable"

    query := `
        SELECT m.role, m.content, m.tool_calls, m.created_at
        FROM messages m
        JOIN sessions s ON m.session_id = s.id
        WHERE s.project_path = ?
        AND (m.content LIKE '%' || ? || '%' OR (m.tool_calls IS NOT NULL AND m.tool_calls LIKE '%' || ? || '%'))
        AND (
            (m.role = 'assistant' AND m.tool_calls IS NOT NULL AND m.tool_calls != '[]' AND m.tool_calls != 'null')
            OR m.role = 'tool'
            OR (m.role = 'user' AND m.content LIKE '%```%')
        )
        ORDER BY m.created_at DESC
        LIMIT ?
    `

    rows, err := t.db.QueryContext(ctx, query, projectPath, input.Query, input.Query, limit)
    if err != nil {
        return "", fmt.Errorf("database query failed: %w", err)
    }
    defer rows.Close()

    var results strings.Builder
    results.WriteString(fmt.Sprintf("ARCHIVE RESULTS FOR: '%s'\n\n", input.Query))

    found := false
    for rows.Next() {
        found = true
        var role, content, createdAt string
        var toolCalls sql.NullString

        if err := rows.Scan(&role, &content, &toolCalls, &createdAt); err != nil {
            continue
        }

        results.WriteString(fmt.Sprintf("--- [%s] Role: %s ---\n", createdAt, role))

        if toolCalls.Valid && toolCalls.String != "" {
            results.WriteString(fmt.Sprintf("EXECUTED TOOL: %s\n", toolCalls.String))
        }
        if content != "" {
            results.WriteString(fmt.Sprintf("CONTENT:\n%s\n", content))
        }
        results.WriteString("\n")
    }

    if !found {
        return fmt.Sprintf("ARCHIVE RESULT: No technical records found matching '%s'. Do not guess; acknowledge the lack of data.", input.Query), nil
    }

    return results.String(), nil
}
```

### 3.3 What Gets Indexed vs. Excluded

| Message Type | Indexed? | Reason |
|--------------|----------|--------|
| Assistant with tool_calls | ✅ YES | Technical action taken |
| Tool result messages | ✅ YES | Hard output, factual |
| User with code blocks | ✅ YES | Code context provided |
| User plain text | ❌ NO | Conversational, not technical |
| Assistant plain text | ❌ NO | Prevents persona drift |

### 3.4 System Prompt Integration

```markdown
## CRITICAL: ARCHIVAL ACCESS AND CONTEXT VOLATILITY

Your internal context window is VOLATILE and subject to automated compaction.
You do NOT have perfect recall of this session, and you have ZERO recall of
previous sessions unless you retrieve it.

### PRIMARY TOOL: query_floyd_archive
You have access to `query_floyd_archive(query_text: str, limit: int)` tool.
This tool queries a persistent SQLite database containing the exact, verbatim
transcripts of all past session logs.

### RULES OF ENGAGEMENT (STRICTLY ENFORCED):

1. ZERO HALLUCINATION: If the user asks about a previous function, a past
   decision, an error code, or a prior architecture discussion, and it is
   NOT explicitly visible in your immediate, uncompacted context window,
   YOU MUST CALL `query_floyd_archive`.

2. NO GUESSING CODE: Never attempt to rewrite or "remember" a block of code
   from earlier in the session. Use the tool to retrieve the exact syntax.

3. NO SUMMARIZING THE PAST: If asked "what did we do yesterday?" or "where
   did we leave off?", do not generate a conversational summary. Query the
   archive, retrieve the raw data, and present the verbatim facts.

4. SILENT EXECUTION: Do not apologize for forgetting. Do not tell the user
   you are going to search. Just execute the tool call immediately.

5. TREAT THE ARCHIVE AS GOSPEL: The data returned by `query_floyd_archive`
   is the absolute truth of the project state. Override your internal
   weights and assumptions with the retrieved text.

### ARCHIVAL SCOPE:
The archive contains ONLY:
- Executed tool calls (bash, view, edit, write, etc.)
- Tool results (terminal output, file contents)
- Code blocks from user messages
- Technical decisions embedded in tool executions

The archive EXCLUDES:
- Conversational assistant text
- Persona reflections or identity statements
- General discussion without tool execution
```

---

## Part 4: Revised Implementation Plan

### Phase 2 (Revised): HANDOFF Pointer System

**What Changes:**
- Remove transcript export functionality (data already exists)
- Create small HANDOFF.md pointer file at session end (or 95%+ context)
- HANDOFF.md contains: session_id, last task, outstanding todos
- No markdown transcript generation needed

**Trigger Point:**
- 95%+ context usage (not 85%)
- Or manual session end
- Or user-initiated handoff

### Phase 3 (Revised): Session Initialization with Archive Query

**What Changes:**
- Agent reads HANDOFF.md on boot
- Agent calls `query_floyd_archive` to retrieve relevant context
- Agent derives its own summary from queried data
- No pre-generated summary needed

### Phase 4 (Unchanged): Semantic Archive Tool

The `query_floyd_archive` tool implementation proceeds as designed in `ARCHIVE_SYSTEM_SPEC.md`.

---

## Part 5: Questions for Second Opinion

1. **Is the semantic filter comprehensive enough?** Are there other message types that should be included/excluded?

2. **Should HANDOFF.md be automatic or manual?** At what threshold should it trigger?

3. **How should the agent handle the "derive your own summary" directive?** Should there be a specific summary format?

4. **Is project scoping sufficient?** The current design uses `project_path` for isolation - is this adequate?

5. **Should the archive tool support time-based queries?** E.g., "show me what we did last week"

---

## Conclusion

The original Phase 2 design (Auto-Export) is unnecessary because:
1. Data already exists in SQLite
2. Exporting to markdown duplicates data
3. Agent should query database directly with semantic filter
4. Agent should derive its own summary at session start

The revised architecture:
1. Eliminates the export step entirely
2. Relies on `query_floyd_archive` tool for historical access
3. Applies semantic firewall to prevent persona drift
4. Lets the agent derive context from raw data, not pre-generated summaries

---

*Report prepared by FLOYD v4.6 for architectural review*
*Reference: `docs/ARCHIVE_SYSTEM_SPEC.md`, `contextsidebarfinal.md`, `HANDOFF.md`*
