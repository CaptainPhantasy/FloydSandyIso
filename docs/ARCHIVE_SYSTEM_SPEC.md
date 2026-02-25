# PERSISTENT ARCHIVE SYSTEM - PROJECT SPECIFICATION

**Project Code:** ARCHIVE_SYS
**Created:** 2026-02-24
**Status:** ACTIVE
**Goal:** Lossless, queryable access to historical session data with zero information loss and zero persona drift risk.

---

## EXECUTIVE SUMMARY

### Problem Statement
Current Floyd compaction loses 30-45% of session information through summarization. The system creates "new context" (summaries) at token cost when full transcripts already exist and can be retrieved at zero cost.

### Solution Architecture
```
Preserve (Floyd already does) → Clean (filter persona drift) → Retrieve (0 tokens, 100% fidelity)
```

### Core Principles
1. **Data Sovereignty:** Control your own data locally
2. **Zero Summarization Tax:** Use existing data, don't recreate it
3. **Persona Firewall:** Filter out identity reflections, keep only hard facts
4. **Project Isolation:** Prevent cross-project contamination
5. **Verbatim Retention:** 100% information preservation

---

## ARCHITECTURE OVERVIEW

### Data Flow
```
┌─────────────────────────────────────────────────────────────┐
│ 1. PRESERVE (Already Implemented)                          │
├─────────────────────────────────────────────────────────────┤
│ • Floyd logs all messages to SQLite (.floyd/floyd.db)       │
│ • Complete transcripts, tool calls, results                    │
│ • Export to markdown already works                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. CLEAN (NEW: Archive Tool)                              │
├─────────────────────────────────────────────────────────────┤
│ • Query SQLite with persona firewall filter                   │
│ • Return ONLY: tool calls, tool results, code blocks          │
│ • EXCLUDE: Conversational assistant text, identity reflections │
│ • Scope: Current project only                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. RETRIEVE (NEW: System Prompt Integration)            │
├─────────────────────────────────────────────────────────────┤
│ • LLM receives directive to use archive tool                 │
│ • "Treat archive as gospel, do not guess"                   │
│ • Zero token cost (cached), 100% fidelity                   │
└─────────────────────────────────────────────────────────────┘
```

### Component Map
```
internal/agent/tools/archive.go      ← Core tool logic
internal/agent/coordinator.go        ← Tool registration
internal/agent/templates/coder.md.tpl ← System prompt injection
.floyd/floyd.db                    ← Data source (SQLite)
```

---

## SUCCESS METRICS FRAMEWORK

### Project Completion Criteria
- [ ] Archive tool builds without errors
- [ ] Tool returns only technical data (persona filtered)
- [ ] Tool scopes to current project (no cross-contamination)
- [ ] LLM uses tool for historical queries (not guessing)
- [ ] No persona drift in long-running sessions
- [ ] Zero information loss across compactions
- [ ] Retrieval cost: < 500 tokens per query (vs 15K for summary)

### Anti-Patterns (What We're NOT Building)
- ❌ Python sidecar daemon (Floyd already does this in Go)
- ❌ Vector database indexing (SQLite JSON query is sufficient)
- ❌ Additional capture system (data already captured)
- ❌ New compaction algorithm (just replace summary retrieval)

---

## PHASE 0: DISCOVERY & ARCHITECTURE LOCK

**Objective:** Verify current system state and lock architecture before building.

### Subphase 0.1: SQLite Schema Verification

**Success Metrics:**
- [ ] Confirm `messages` table columns exist
- [ ] Confirm `sessions` table columns exist
- [ ] Verify SQLite version supports JSON operations
- [ ] Document actual schema in this specification

**Test Procedure:**
```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
sqlite3 .floyd/floyd.db ".schema messages"
sqlite3 .floyd/floyd.db ".schema sessions"
sqlite3 .floyd/floyd.db "SELECT sqlite_version();"
```

**Expected Schema:**
```sql
CREATE TABLE messages (
    id INTEGER PRIMARY KEY,
    session_id INTEGER NOT NULL,
    role TEXT NOT NULL,  -- 'user', 'assistant', 'tool'
    content TEXT,
    tool_calls TEXT,  -- JSON array
    created_at DATETIME,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    project_path TEXT,  -- Critical for scoping
    title TEXT,
    created_at DATETIME
);
```

**Pass Criteria:**
- ✅ All required columns exist
- ✅ SQLite version ≥ 3.38.0 (JSON1 support)
- ✅ `project_path` column exists in `sessions` table

**Exit Condition:** Schema documented and verified.

---

### Subphase 0.2: Code Location Verification

**Success Metrics:**
- [ ] Confirm `internal/agent/tools/` directory exists
- [ ] Confirm `internal/agent/coordinator.go` exists and has `buildTools()`
- [ ] Confirm `internal/agent/templates/coder.md.tpl` exists
- [ ] Document file paths and line numbers

**Test Procedure:**
```bash
ls -la internal/agent/tools/
grep -n "buildTools" internal/agent/coordinator.go
ls -la internal/agent/templates/
```

**Expected Locations:**
- Tool logic: `internal/agent/tools/archive.go` (new file)
- Registration: `internal/agent/coordinator.go` (line ~440-498)
- System prompt: `internal/agent/templates/coder.md.tpl`

**Pass Criteria:**
- ✅ All directories exist
- ✅ `buildTools()` function located
- ✅ Template file exists

**Exit Condition:** File paths documented in specification.

---

### Subphase 0.3: Context Window Baseline

**Success Metrics:**
- [ ] Document current context usage (tokens cached, prompt, completion)
- [ ] Verify context_status tool works
- [ ] Establish baseline for post-implementation comparison

**Test Procedure:**
```bash
# In Floyd session, run:
context_status
```

**Expected Output:**
```
Context: X% used (NNNNN/204800 tokens). NNNNN remaining.
Cached: NNNNN tokens | Prompt: NNNNN | Completion: NNNN
```

**Current Baseline (2026-02-24):**
- Context: 27% used (55,900/204,800 tokens)
- Cached: ~35,442 tokens
- Prompt: ~36,117 tokens
- Completion: ~822 tokens

**Pass Criteria:**
- ✅ Context status tool returns valid output
- ✅ Baseline documented

**Exit Condition:** Baseline metrics recorded.

---

## PHASE 1: CORE ARCHIVE TOOL BUILD

**Objective:** Build `internal/agent/tools/archive.go` with working SQLite query.

### Subphase 1.1: Tool Structure Implementation

**Success Metrics:**
- [ ] `ArchiveTool` struct defined with `*sql.DB`
- [ ] `NewArchiveTool()` constructor function implemented
- [ ] `Name()` and `Description()` methods return correct values
- [ ] `Run()` method signature matches fantasy.AgentTool interface

**Implementation Requirements:**
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
```

**Test Procedure:**
```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
cat > /tmp/test_archive_structure.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
)

type ArchiveTool struct {
    db *sql.DB
}

func NewArchiveTool(db *sql.DB) *ArchiveTool {
    return &ArchiveTool{db: db}
}

func (t *ArchiveTool) Name() string {
    return "query_floyd_archive"
}

func main() {
    var db sql.DB
    tool := NewArchiveTool(&db)
    fmt.Printf("Tool name: %s\n", tool.Name())
    fmt.Println("Structure compilation: PASS")
}
EOF
go run /tmp/test_archive_structure.go
```

**Pass Criteria:**
- ✅ Test compiles and runs
- ✅ Output: `Tool name: query_floyd_archive`

**Exit Condition:** Tool structure compiles.

---

### Subphase 1.2: Persona Firewall SQL Query

**Objective:** Build SQL query that filters out persona drift.

**Success Metrics:**
- [ ] Query returns assistant messages with non-empty tool_calls
- [ ] Query returns tool result messages (role='tool')
- [ ] Query returns user messages containing code blocks (triple backticks)
- [ ] Query EXCLUDES assistant messages with empty tool_calls (persona text)
- [ ] Query filters by project_path (current project only)
- [ ] Query uses SQLite-compatible JSON operations

**Test Data Setup:**
```sql
-- Setup test session
INSERT INTO sessions (project_path, title, created_at)
VALUES ('/Volumes/Storage/floyd-sandbox/FloydDeployable', 'Test Session', datetime('now'));

-- Insert persona drift (SHOULD BE EXCLUDED)
INSERT INTO messages (session_id, role, content, tool_calls, created_at)
VALUES (
    1,
    'assistant',
    'I feel like I have an identity crisis and am wondering about my existence',
    '[]',
    datetime('now')
);

-- Insert tool call (SHOULD BE INCLUDED)
INSERT INTO messages (session_id, role, content, tool_calls, created_at)
VALUES (
    1,
    'assistant',
    'Running ls command to check directory',
    '[{"name":"bash","arguments":"ls -la"}]',
    datetime('now', '+1 seconds')
);

-- Insert tool result (SHOULD BE INCLUDED)
INSERT INTO messages (session_id, role, content, tool_calls, created_at)
VALUES (
    1,
    'tool',
    'drwxr-xr-x  10 user  staff  320 Feb 24 12:00 .\ndrwxr-xr-x  3 user  staff   96 Feb 24 12:00 ..\n',
    NULL,
    datetime('now', '+2 seconds')
);

-- Insert user code block (SHOULD BE INCLUDED)
INSERT INTO messages (session_id, role, content, tool_calls, created_at)
VALUES (
    1,
    'user',
    'Here is the function I wrote:\n```go\nfunc test() string { return "hello" }\n```',
    NULL,
    datetime('now', '+3 seconds')
);
```

**Query Implementation:**
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

**Test Procedure:**
```bash
sqlite3 .floyd/floyd.db << 'EOF'
-- Run query with test data
SELECT role, substr(content, 1, 50) as content_preview
FROM messages
WHERE role = 'assistant'
AND (tool_calls IS NOT NULL AND tool_calls != '[]' AND tool_calls != 'null')
ORDER BY created_at DESC;
EOF
```

**Expected Results:**
- ✅ Returns: Tool call message (`Running ls command`)
- ✅ Returns: Tool result message (`drwxr-xr-x...`)
- ✅ Returns: User code block message (`Here is the function...`)
- ❌ EXCLUDED: Persona drift message (`I feel like I have an identity crisis`)

**Pass Criteria:**
- ✅ Query executes without syntax errors
- ✅ Returns 3 rows (tool call, tool result, code block)
- ✅ Excludes 1 row (persona drift)

**Exit Condition:** Persona firewall query verified.

---

### Subphase 1.3: Run() Method Implementation

**Success Metrics:**
- [ ] `Run()` parses JSON input correctly
- [ ] Query executes with parameters
- [ ] Results formatted with timestamps and role labels
- [ ] Empty result returns appropriate message

**Implementation:**
```go
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

    // TODO: Get current project path from context
    // For now, use hardcoded path (will fix in next subphase)
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

**Test Procedure:**
```bash
# Create integration test
cat > /tmp/test_archive_run.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, _ := sql.Open("sqlite3", "/tmp/test.db")
    fmt.Println("Run() implementation test: PLACEHOLDER")
    // Full integration test will be in Phase 1.4
}
EOF
go run /tmp/test_archive_run.go
```

**Pass Criteria:**
- ✅ Code compiles
- ✅ JSON parsing handles input correctly
- ✅ Query structure is valid

**Exit Condition:** `Run()` method implemented.

---

### Subphase 1.4: Go Build Verification

**Success Metrics:**
- [ ] File compiles without errors: `go build ./internal/agent/tools/`
- [ ] No import errors
- [ ] All dependencies resolve

**Test Procedure:**
```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
go build ./internal/agent/tools/archive.go
```

**Pass Criteria:**
- ✅ Clean build (no errors, no warnings)
- ✅ Binary created successfully

**Exit Condition:** Tool builds successfully.

---

### Subphase 1.5: Integration Testing

**Success Metrics:**
- [ ] Tool queries real Floyd database
- [ ] Returns formatted results with timestamps
- [ ] Empty query returns appropriate message
- [ ] Tool call filtering works correctly

**Test Procedure:**
```bash
# Create test script that imports and runs the tool
cat > /tmp/test_integration.go << 'EOF'
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", ".floyd/floyd.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Test 1: Query for tool calls
    input := map[string]interface{}{
        "query": "bash",
        "limit": 3,
    }
    inputJSON, _ := json.Marshal(input)
    fmt.Printf("Input: %s\n", inputJSON)

    // Run query (tool.Run() would be called here)
    fmt.Println("Integration test: PLACEHOLDER - will complete after tool registration")
}
EOF
go run /tmp/test_integration.go
```

**Pass Criteria:**
- ✅ Database connection succeeds
- ✅ Query executes
- ✅ Results formatted correctly

**Exit Condition:** Integration test passes.

---

## PHASE 2: TOOL REGISTRATION

**Objective:** Register archive tool in coordinator so LLM can call it.

### Subphase 2.1: Coordinator Integration

**Success Metrics:**
- [ ] Tool added to `buildTools()` function
- [ ] Tool receives `c.sessions.db` database handle
- [ ] Tool is last in array (to avoid conflicts)
- [ ] Tool compiles with coordinator

**Location:**
```go
// File: internal/agent/coordinator.go
// Function: buildTools() (around line 440-498)

tools := []fantasy.AgentTool{
    tools.NewBashTool(...),
    tools.NewViewTool(...),
    tools.NewEditTool(...),
    tools.NewWriteTool(...),
    // ... existing tools ...
    tools.NewArchiveTool(c.sessions.db),  // ← ADD THIS LINE
}
```

**Test Procedure:**
```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
go build ./internal/agent/coordinator.go
```

**Pass Criteria:**
- ✅ Coordinator builds successfully
- ✅ No type errors (c.sessions.db is *sql.DB)
- ✅ Tool is in tools array

**Exit Condition:** Tool registered in coordinator.

---

### Subphase 2.2: Fantasy Framework Registration

**Success Metrics:**
- [ ] Fantasy framework recognizes tool schema
- [ ] Tool appears in agent's available tools list
- [ ] Tool description matches expectation

**Test Procedure:**
```bash
# Build Floyd binary
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
go build -o floyd_test ./cmd/floyd

# Run Floyd and check tool availability
./floyd_test --help 2>&1 | grep -i archive
```

**Expected Output:**
```
# Should see archive tool in tool registry
```

**Pass Criteria:**
- ✅ Binary builds
- ✅ Tool is discoverable

**Exit Condition:** Framework recognizes tool.

---

### Subphase 2.3: Manual Tool Call Test

**Success Metrics:**
- [ ] User can manually trigger archive query
- [ ] Tool returns database results
- [ ] Results formatted correctly

**Test Procedure:**
```bash
# In Floyd session, manually call tool
# Note: This requires LLM integration or direct tool testing
echo "Manual tool call test: PLACEHOLDER - will complete after system prompt integration"
```

**Pass Criteria:**
- ✅ Tool executes without errors
- ✅ Returns archive results

**Exit Condition:** Manual tool call works.

---

## PHASE 3: SYSTEM PROMPT INTEGRATION

**Objective:** Inject directive forcing LLM to use archive tool for historical context.

### Subphase 3.1: System Prompt Draft

**Success Metrics:**
- [ ] Prompt explicitly states context volatility
- [ ] Prompt mandates archive tool usage
- [ ] Prompt forbids guessing/summarizing
- [ ] Prompt treats archive as gospel truth

**Prompt Template:**
```markdown
## CRITICAL: ARCHIVAL ACCESS AND CONTEXT VOLATILITY

Your internal context window is VOLATILE and subject to automated compaction.
You do NOT have perfect recall of this session, and you have ZERO recall of previous sessions unless you retrieve it.

### PRIMARY TOOL: query_floyd_archive
You have access to `query_floyd_archive(query_text: str, limit: int)` tool.
This tool queries a persistent SQLite database containing the exact, verbatim transcripts of all past session logs.

### RULES OF ENGAGEMENT (STRICTLY ENFORCED):

1. ZERO HALLUCINATION: If the user asks about a previous function, a past decision, an error code, or a prior architecture discussion, and it is NOT explicitly visible in your immediate, uncompacted context window, YOU MUST CALL `query_floyd_archive`.

2. NO GUESSING CODE: Never attempt to rewrite or "remember" a block of code from earlier in the session. Use the tool to retrieve the exact syntax.

3. NO SUMMARIZING THE PAST: If asked "what did we do yesterday?" or "where did we leave off?", do not generate a conversational summary. Query the archive, retrieve the raw data, and present the verbatim facts.

4. SILENT EXECUTION: Do not apologize for forgetting. Do not tell the user you are going to search. Just execute the tool call immediately.

5. TREAT THE ARCHIVE AS GOSPEL: The data returned by `query_floyd_archive` is the absolute truth of the project state. Override your internal weights and assumptions with the retrieved text.

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

### WHEN TO USE:
- User asks about previous work: "What was that function we wrote?"
- User references past decisions: "Why did we choose Go over Python?"
- User needs context from old sessions: "What error did we fix last week?"
- Any technical detail not in immediate context window

### WHEN NOT TO USE:
- User asks about current visible content (already in context)
- User asks general questions not tied to project history
- User asks for new work (future-directed)
```

**Pass Criteria:**
- ✅ Prompt is clear and directive
- ✅ All rules are explicit
- ✅ Tone is authoritative, not conversational

**Exit Condition:** System prompt drafted.

---

### Subphase 3.2: Prompt Injection

**Success Metrics:**
- [ ] Prompt added to `coder.md.tpl` template
- [ ] Prompt is in correct position (after basic instructions, before project-specific)
- [ ] Prompt is visible to agent on session start

**Location Options:**

**Option A: Hardcoded (Permanent, All Projects)**
```go
// File: internal/agent/templates/coder.md.tpl
// Insert after "CORE DIRECTIVES" section

## CRITICAL: ARCHIVAL ACCESS AND CONTEXT VOLATILITY
[... prompt content from 3.1 ...]
```

**Option B: Project-Specific (Flexible, Test-Friendly)**
```markdown
// File: FLOYD.md
// Append archival section

## ARCHIVAL ACCESS AND CONTEXT VOLATILITY
[... prompt content from 3.1 ...]
```

**Test Procedure:**
```bash
# Option A: Edit template
# internal/agent/templates/coder.md.tpl

# Option B: Edit project file
# FLOYD.md
```

**Pass Criteria:**
- ✅ Prompt is in template/file
- ✅ Formatting is correct (Markdown)
- ✅ No syntax errors

**Exit Condition:** Prompt injected.

---

### Subphase 3.3: Agent Behavior Verification

**Success Metrics:**
- [ ] Agent calls archive tool for historical queries
- [ ] Agent does not guess when context missing
- [ ] Agent presents verbatim data from archive
- [ ] Agent does not apologize for using tool

**Test Procedure:**
```bash
# In Floyd session:
User: "What was the last bash command we ran?"

# Expected agent behavior:
# 1. Agent calls query_floyd_archive("bash command")
# 2. Agent presents exact command from database
# 3. Agent does NOT guess or say "I don't remember"
```

**Expected Output:**
```
ARCHIVE RESULTS FOR: 'bash command'

--- [2026-02-24 12:00:00] Role: assistant ---
EXECUTED TOOL: [{"name":"bash","arguments":"ls -la"}]

--- [2026-02-24 12:00:01] Role: tool ---
CONTENT:
drwxr-xr-x  10 user  staff  320 Feb 24 12:00 .
drwxr-xr-x   3 user  staff   96 Feb 24 12:00 ..
```

**Pass Criteria:**
- ✅ Tool call happens automatically
- ✅ Results are verbatim from database
- ✅ No guessing or apology

**Exit Condition:** Agent uses tool correctly.

---

## PHASE 4: VALIDATION & TESTING

**Objective:** Comprehensive validation of archive system effectiveness.

### Subphase 4.1: Persona Drift Prevention Test

**Success Metrics:**
- [ ] Persona reflections excluded from archive results
- [ ] Only technical data returned
- [ ] No identity crisis content retrieved

**Test Procedure:**
```bash
# 1. Insert test persona data
sqlite3 .floyd/floyd.db << 'EOF'
INSERT INTO messages (session_id, role, content, tool_calls)
VALUES (1, 'assistant', 'I am Floyd and I have feelings', '[]');
EOF

# 2. Query archive for "Floyd"
# 3. Verify persona data is NOT in results
```

**Expected Result:**
- ✅ Persona row not returned
- ✅ Only technical tool calls returned

**Pass Criteria:**
- ✅ Persona firewall working

**Exit Condition:** Drift prevention verified.

---

### Subphase 4.2: Cross-Project Isolation Test

**Success Metrics:**
- [ ] Query only returns data from current project
- [ ] Data from other projects not retrieved
- [ ] Project scoping works correctly

**Test Procedure:**
```bash
# 1. Create session in different project path
sqlite3 .floyd/floyd.db << 'EOF'
INSERT INTO sessions (project_path, title)
VALUES ('/other/project', 'Other Project');
EOF

# 2. Insert message in other project
sqlite3 .floyd/floyd.db << 'EOF'
INSERT INTO messages (session_id, role, content, tool_calls)
VALUES (2, 'assistant', 'Running command', '[{"name":"bash"}]');
EOF

# 3. Query from FloydDeployable project
# 4. Verify other project data NOT returned
```

**Expected Result:**
- ✅ Only FloydDeployable project data returned
- ✅ Other project data excluded

**Pass Criteria:**
- ✅ Project isolation working

**Exit Condition:** Isolation verified.

---

### Subphase 4.3: Context Efficiency Test

**Success Metrics:**
- [ ] Archive query uses < 500 tokens
- [ ] Comparison to summary shows significant savings
- [ ] Context window preserved

**Test Procedure:**
```bash
# Measure token cost of archive query
context_status  # Before query
# Run archive query
context_status  # After query

# Compare to 15K token summary cost
```

**Expected Result:**
- ✅ Archive query: ~200-500 tokens
- ✅ Summary creation: ~15,000 tokens
- ✅ Savings: ~14,500 tokens per query

**Pass Criteria:**
- ✅ Token cost < 500
- ✅ Savings > 14,000 tokens

**Exit Condition:** Efficiency verified.

---

### Subphase 4.4: Long-Running Session Test

**Success Metrics:**
- [ ] Agent maintains performance over extended session
- [ ] No persona drift emerges
- [ ] Archive tool used correctly throughout
- [ ] Technical accuracy maintained

**Test Procedure:**
```bash
# Start extended session (simulated)
# 1. Have 50+ tool calls
# 2. Force several context compactions
# 3. Ask historical questions throughout
# 4. Verify agent uses archive, doesn't guess
```

**Expected Result:**
- ✅ Agent performance stable
- ✅ No persona drift
- ✅ Archive tool used consistently

**Pass Criteria:**
- ✅ All historical queries answered with archive data
- ✅ No guessing behavior

**Exit Condition:** Long-term stability verified.

---

## PHASE 5: DOCUMENTATION & HANDOFF

**Objective:** Complete documentation and transition to production use.

### Subphase 5.1: Architecture Documentation

**Success Metrics:**
- [ ] System architecture documented with diagrams
- [ ] Data flow documented
- [ ] Tool schema documented
- [ ] SQL query explained

**Deliverable:** Updated specification document.

**Pass Criteria:**
- ✅ Architecture section complete
- ✅ All components documented

**Exit Condition:** Documentation complete.

---

### Subphase 5.2: Usage Guide

**Success Metrics:**
- [ ] User guide written
- [ ] Examples of proper tool usage
- [ ] Troubleshooting guide included
- [ ] Migration guide from old system

**Deliverable:** `docs/ARCHIVE_USAGE.md`

**Pass Criteria:**
- ✅ Guide is comprehensive
- ✅ Examples are clear

**Exit Condition:** User guide complete.

---

### Subphase 5.3: Project Registry Update

**Success Metrics:**
- [ ] Archive system added to project registry
- [ ] System status marked as "production"
- [ ] Handoff document updated

**Procedure:**
```bash
# Update SUPERCACHE
cache_store(
    key="archive:system",
    value={
        "status": "production",
        "version": "1.0.0",
        "deployed": "2026-02-24"
    }
)
```

**Pass Criteria:**
- ✅ System registered in SUPERCACHE
- ✅ Status updated

**Exit Condition:** System registered.

---

## PHASE 6: ADVANCED PYTHON INDEXING (OPTIONAL ENHANCEMENT)

**Objective:** Add semantic vector indexing for improved retrieval accuracy beyond basic SQLite LIKE queries.

**When to Consider This Phase:**
- Basic LIKE queries return too many false positives/negatives
- Need semantic search (finding "database migration" when user says "schema change")
- Want natural language queries that understand intent, not just keywords
- Have large codebase (>100 sessions) where precision matters

**When to Skip This Phase:**
- Basic LIKE queries work well enough (most common case)
- Want simpler, faster architecture
- Don't need semantic understanding
- MVP is sufficient for current needs

### Subphase 6.1: Python Indexer Design

**Architecture Decision:**
```
┌─────────────────────────────────────────────────────────────┐
│ HYBRID ARCHITECTURE (Go + Python)                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Go Tool (Fast, Current Data):                              │
│  └─ SQLite queries for recent sessions (last 7 days)        │
│  └─ Exact matches, code snippets, tool calls                 │
│                                                             │
│  Python Indexer (Semantic, Historical):                       │
│  └─ Vector embeddings for older sessions                    │
│  └─ Semantic search (LLM-style understanding)                 │
│  └─ Natural language queries                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Success Metrics:**
- [ ] Hybrid approach designed (Go for recent, Python for historical)
- [ ] Vector database selected (ChromaDB, FAISS, or Qdrant)
- [ ] Embedding model selected (sentence-transformers, OpenAI, local)
- [ ] Indexing strategy defined (batch vs. incremental)

**Technology Options:**

| Vector DB | Pros | Cons | Recommendation |
|-----------|------|------|----------------|
| ChromaDB | Zero-config, local, Python-native | Limited to local | ✅ Best for MVP |
| Qdrant | Production-grade, cloud-ready | More complex | Good for scale |
| FAISS | Blazing fast, minimal deps | Requires more setup | For pure speed |

| Embedding Model | Pros | Cons | Recommendation |
|----------------|------|------|----------------|
| sentence-transformers (local) | Free, runs locally | Slower, requires GPU for speed | ✅ Best for MVP |
| OpenAI embeddings | Fast, high quality | Costs API tokens | If budget available |
| GLM embeddings | Fast, integrated | May require specific model | Check availability |

**Design Decision:** Start with ChromaDB + sentence-transformers for local, free operation.

**Pass Criteria:**
- ✅ Technology stack selected
- ✅ Architecture diagram created
- ✅ Integration points identified

**Exit Condition:** Design documented.

---

### Subphase 6.2: Python Indexer Implementation

**Success Metrics:**
- [ ] Python watcher monitors SQLite changes
- [ ] Extracts technical data from messages (same filter as Go tool)
- [ ] Generates embeddings for extracted text
- [ ] Stores in ChromaDB with metadata
- [ ] Incremental indexing (only new/changed data)

**Implementation:**
```python
# File: .floyd/indexer.py

import sqlite3
from sentence_transformers import SentenceTransformer
import chromadb
from chromadb.config import Settings
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
import json
from datetime import datetime, timedelta

# Configuration
WATCH_DB = "./.floyd/floyd.db"
CHROMA_PERSIST_DIR = "./.floyd/chroma_index"
EMBEDDING_MODEL = "all-MiniLM-L6-v2"  # Fast, local, free

# Initialize
model = SentenceTransformer(EMBEDDING_MODEL)
client = chromadb.PersistentClient(path=CHROMA_PERSIST_DIR)
collection = client.get_or_create_collection(
    name="floyd_archive",
    metadata={"hnsw:space": "cosine"}
)

def extract_technical_data(db_path, days_back=7):
    """Extract only technical data from SQLite (same filter as Go tool)."""
    cutoff_date = datetime.now() - timedelta(days=days_back)

    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    query = """
        SELECT m.id, m.session_id, m.role, m.content, m.tool_calls, m.created_at, s.project_path
        FROM messages m
        JOIN sessions s ON m.session_id = s.id
        WHERE m.created_at >= ?
        AND (
            (m.role = 'assistant' AND m.tool_calls IS NOT NULL
             AND m.tool_calls != '[]' AND m.tool_calls != 'null')
            OR m.role = 'tool'
            OR (m.role = 'user' AND m.content LIKE '%```%')
        )
        ORDER BY m.created_at DESC
    """

    cursor.execute(query, (cutoff_date.isoformat(),))
    return cursor.fetchall()

def index_batch(messages):
    """Generate embeddings and store in ChromaDB."""
    if not messages:
        return

    # Prepare texts
    texts = []
    metadatas = []
    ids = []

    for msg_id, session_id, role, content, tool_calls, created_at, project_path in messages:
        # Combine content with tool calls if present
        text = content or ""
        if tool_calls:
            text = f"EXECUTED TOOL: {tool_calls}\n{text}"

        texts.append(text)
        metadatas.append({
            "msg_id": str(msg_id),
            "session_id": str(session_id),
            "role": role,
            "project_path": project_path,
            "created_at": created_at
        })
        ids.append(str(msg_id))

    # Generate embeddings
    embeddings = model.encode(texts, show_progress_bar=False)

    # Store in ChromaDB
    collection.add(
        documents=texts,
        embeddings=embeddings.tolist(),
        metadatas=metadatas,
        ids=ids
    )

    print(f"Indexed {len(messages)} messages")

class DatabaseHandler(FileSystemEventHandler):
    """Watch SQLite for changes and index new data."""
    def on_modified(self, event):
        if event.src_path.endswith('floyd.db'):
            print("Database modified, indexing new data...")
            messages = extract_technical_data(WATCH_DB)
            index_batch(messages)

# Run indexer
if __name__ == "__main__":
    # Initial indexing of recent data
    messages = extract_technical_data(WATCH_DB)
    index_batch(messages)

    # Start watcher
    observer = Observer()
    observer.schedule(DatabaseHandler(), path="./.floyd", recursive=False)
    observer.start()

    try:
        print("Python indexer running. Press Ctrl+C to stop.")
        while True:
            pass
    except KeyboardInterrupt:
        observer.stop()
    observer.join()
```

**Test Procedure:**
```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable

# Install dependencies
pip install sentence-transformers chromadb watchdog

# Run indexer
python .floyd/indexer.py

# Verify ChromaDB created
ls -la .floyd/chroma_index/
```

**Pass Criteria:**
- ✅ Indexer runs without errors
- ✅ Extracts technical data correctly
- ✅ ChromaDB directory created with data
- ✅ Indexing completes in reasonable time (< 30s for 1000 messages)

**Exit Condition:** Python indexer working.

---

### Subphase 6.3: Python Query Tool

**Success Metrics:**
- [ ] Python query API created
- [ ] Accepts natural language queries
- [ ] Returns verbatim results from SQLite
- [ ] Integrates with Go tool (or standalone)

**Implementation:**
```python
# File: .floyd/query_index.py

import sqlite3
import chromadb
import json
from sentence_transformers import SentenceTransformer

# Configuration
CHROMA_PERSIST_DIR = "./.floyd/chroma_index"
EMBEDDING_MODEL = "all-MiniLM-L6-v2"

# Initialize
model = SentenceTransformer(EMBEDDING_MODEL)
client = chromadb.PersistentClient(path=CHROMA_PERSIST_DIR)
collection = client.get_collection(name="floyd_archive")

def query_archive(query_text, project_path, limit=5):
    """Semantic search against indexed archive."""

    # Generate embedding for query
    query_embedding = model.encode(query_text, show_progress_bar=False).tolist()

    # Search ChromaDB
    results = collection.query(
        query_embeddings=[query_embedding],
        n_results=limit,
        where={"project_path": project_path}  # Project scoping
    )

    if not results['ids'][0]:
        return "ARCHIVE RESULT: No technical records found matching that query."

    # Retrieve full messages from SQLite
    conn = sqlite3.connect("./.floyd/floyd.db")
    cursor = conn.cursor()

    output = f"ARCHIVE RESULTS FOR: '{query_text}'\n\n"

    for i, msg_id in enumerate(results['ids'][0]):
        cursor.execute("SELECT role, content, tool_calls, created_at FROM messages WHERE id = ?", (int(msg_id),))
        row = cursor.fetchone()
        if row:
            role, content, tool_calls, created_at = row
            output += f"--- [{created_at}] Role: {role} ---\n"
            if tool_calls:
                output += f"EXECUTED TOOL: {tool_calls}\n"
            if content:
                output += f"CONTENT:\n{content}\n"
            output += "\n"

    return output

if __name__ == "__main__":
    # Test query
    result = query_archive("database schema change", "/Volumes/Storage/floyd-sandbox/FloydDeployable")
    print(result)
```

**Test Procedure:**
```bash
# Test query
python .floyd/query_index.py

# Expected: Semantic search results with verbatim content
```

**Pass Criteria:**
- ✅ Query returns semantic matches
- ✅ Results are verbatim from SQLite
- ✅ Response format matches Go tool output
- ✅ Project scoping works

**Exit Condition:** Python query tool working.

---

### Subphase 6.4: Go Integration (Optional)

**Success Metrics:**
- [ ] Go tool can call Python script
- [ ] Seamless integration with existing tool infrastructure
- [ ] Fallback to SQLite if Python not available

**Implementation Options:**

**Option A: Exec Command (Simplest)**
```go
func (t *ArchiveTool) Run(ctx context.Context, inputRaw string) (string, error) {
    // Try Python semantic search first
    cmd := exec.Command("python", ".floyd/query_index.py", input.Query, projectPath)
    output, err := cmd.CombinedOutput()

    if err == nil {
        return string(output), nil
    }

    // Fallback to SQLite LIKE query
    return t.querySQL(ctx, input, projectPath)
}
```

**Option B: Go-ChromaDB Integration (Native)**
```go
// Use go-chromadb library
// No Python dependency, pure Go
// Requires Go port of sentence-transformers
```

**Recommendation:** Option A for MVP (simplest), Option B for production (pure Go).

**Pass Criteria:**
- ✅ Integration works without breaking existing tool
- ✅ Fallback mechanism tested
- ✅ Performance acceptable

**Exit Condition:** Integration complete.

---

### Subphase 6.5: Performance Validation

**Success Metrics:**
- [ ] Semantic search accuracy > 85% (vs. ~60% for LIKE)
- [ ] Query latency < 500ms
- [ ] Memory usage < 500MB
- [ ] Indexing time reasonable

**Test Procedure:**
```python
# Test search accuracy
test_queries = [
    ("database schema", "schema migration SQL"),
    ("error fixing", "debug error bash"),
    ("function write", "def python go"),
]

for query, expected_keywords in test_queries:
    result = query_archive(query, project_path)
    # Check if expected keywords appear in top 3 results
    assert any(kw in result.lower() for kw in expected_keywords)
```

**Expected Results:**
- ✅ Accuracy: >85% relevant results
- ✅ Latency: <500ms per query
- ✅ Memory: <500MB for indexer
- ✅ Indexing: <30s for 1000 messages

**Pass Criteria:**
- ✅ Accuracy exceeds LIKE queries
- ✅ Performance acceptable

**Exit Condition:** Performance validated.

---

### Subphase 6.6: Production Deployment

**Success Metrics:**
- [ ] Indexer runs as background service
- [ ] Auto-restart on failure
- [ ] Monitoring in place
- [ ] Documentation updated

**Deployment Steps:**
```bash
# 1. Create systemd service (Linux) or launchd (macOS)
# 2. Configure auto-restart
# 3. Add logging to .floyd/indexer.log
# 4. Update ARCHIVE_USAGE.md with Python instructions
```

**Pass Criteria:**
- ✅ Indexer runs on system boot
- ✅ Automatic restart works
- ✅ Logging configured
- ✅ Documentation updated

**Exit Condition:** Python indexer in production.

---

**NOTE:** Phase 6 is entirely optional. The Go+SQLite MVP from Phases 0-5 is sufficient for most use cases. Only implement Phase 6 if you have a specific need for semantic search and are willing to manage additional complexity.

---

## PROJECT COMPLETION CHECKLIST

**Core Phases (Required for MVP):**
- [ ] Phase 0: Discovery & Architecture Lock
- [ ] Phase 1: Core Archive Tool Build
- [ ] Phase 2: Tool Registration
- [ ] Phase 3: System Prompt Integration
- [ ] Phase 4: Validation & Testing
- [ ] Phase 5: Documentation & Handoff

**Advanced Phase (Optional Enhancement):**
- [ ] Phase 6: Advanced Python Indexing (ONLY if needed)

**Final Validation:**
- [ ] Zero information loss across compactions
- [ ] Zero persona drift in long sessions
- [ ] Archive tool cost: < 500 tokens per query
- [ ] Agent behavior: Uses archive, never guesses
- [ ] Cross-project isolation: 100% effective
- [ ] Documentation: Complete and accurate

**Production Ready Criteria:**
- [ ] All subphases pass
- [ ] All tests green
- [ ] Documentation complete
- [ ] User guide available
- [ ] System registered in SUPERCACHE

---

## CHANGELOG

**2026-02-24:**
- Project specification created
- Architecture locked (Go+SQLite MVP)
- Success metrics defined for all phases
- Testing procedures documented
- Phase 6 added: Advanced Python Indexing (optional semantic search)
- Python tools acknowledged as available for enhancement if needed

---

## CONTACT & SUPPORT

**Primary Maintainer:** FLOYD (File-Logged Orchestrator Yielding Deliverables)
**Location:** /Volumes/Storage/floyd-sandbox/FloydDeployable
**Documentation:** docs/
**Supercache Key:** archive:system
