# Enterprise Semantic Filter Architecture

## Problem Statement

The current `query_floyd_archive` tool uses literal SQL `LIKE` matching. A query for "China school thing" will NOT find "Chinese education system" because the words don't match exactly.

What's needed is **vague recollection → precise retrieval** with minimal token burn.

---

## Current State vs. Needed State

| Capability | Current | Needed |
|------------|---------|--------|
| String matching | `LIKE '%word%'` (exact) | Fuzzy, stemmed, phrase-aware |
| Time awareness | None | "last week", "yesterday", "3 days ago" |
| Semantic similarity | None | "China school" → "Chinese education" |
| Result size | Full messages | Summaries + drill-down |
| Token efficiency | Returns everything | Returns card catalog, not library |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        QUERY LAYER                               │
│  "that thing about Chinese schools from last week"              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    QUERY PROCESSOR                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Time Parser  │  │ Entity       │  │ Query        │          │
│  │ "last week"  │  │ Extractor    │  │ Expansion    │          │
│  │ → date range │  │ "China"→CN   │  │ school→edu   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      INDEX LAYER                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    FTS5 Virtual Table                    │   │
│  │  sessions_fts(session_id, title, summary, tools, code)   │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Metadata Store                        │   │
│  │  sessions_meta(session_id, created_at, tokens, files[]) │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    RESULT LAYER                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Ranked Results (BM25)                                    │  │
│  │  1. Session ABC - "Chinese education reform" (0.87)      │  │
│  │  2. Session XYZ - "China school system discussion" (0.72)│  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Summarized Output (~500 tokens max)                      │  │
│  │  "In session ABC, you were analyzing China's 9-year      │  │
│  │   compulsory education system. Key files: education.go.  │  │
│  │   Tools used: web_search, view. Decision: Use cached data│  │
│  │   from UNESCO."                                          │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Layer 1: Query Processor

### Time Parser

Converts natural language time references to date ranges:

```go
type TimeRange struct {
    Start time.Time
    End   time.Time
}

func ParseTimeQuery(query string) (TimeRange, string) {
    // "last week" → (7 days ago, now)
    // "yesterday" → (1 day ago, now)
    // "3 days ago" → (3 days ago, 3 days ago + 1 day)
    // "before christmas" → (beginning, Dec 25)
}
```

### Entity Extractor

Recognizes and normalizes entities:

```go
type Entity struct {
    Text     string
    Type     string // "location", "person", "org", "tech"
    Normalized string
}

// "China" → Entity{Text: "China", Type: "location", Normalized: "CN|China|Chinese"}
// "React" → Entity{Text: "React", Type: "tech", Normalized: "React|ReactJS|react.js"}
```

### Query Expansion

Stemming and synonyms:

```
school → school, schools, schooling, education, educational, edu
fix → fix, fixed, fixing, bug, issue, problem, repair
```

Uses Porter stemmer + domain-specific synonym maps.

---

## Layer 2: Index Layer

### FTS5 Virtual Table

SQLite FTS5 provides full-text search with:

- **BM25 ranking** — most relevant results first
- **Phrase matching** — `"exact phrase"` works
- **Prefix matching** — `edu*` matches education, educational
- **Stemming** — automatically handled via tokenizer

```sql
CREATE VIRTUAL TABLE sessions_fts USING fts5(
    session_id,
    title,
    summary,        -- Auto-generated session summary
    tools_used,     -- Comma-separated tool names
    files_touched,  -- Comma-separated file paths
    code_snippets,  -- Extracted code blocks
    entities,       -- Extracted entities
    tokenize='porter unicode61'
);
```

### Metadata Store

```sql
CREATE TABLE sessions_meta (
    session_id TEXT PRIMARY KEY,
    created_at INTEGER,
    token_count INTEGER,
    message_count INTEGER,
    has_handoff BOOLEAN,
    project_path TEXT
);

CREATE INDEX idx_sessions_created ON sessions_meta(created_at DESC);
CREATE INDEX idx_sessions_project ON sessions_meta(project_path);
```

---

## Layer 3: Result Layer

### Two-Phase Retrieval

**Phase 1: Card Catalog** (~50 tokens)

```
Found 3 sessions matching "Chinese education":

1. [Feb 24] Education system analysis - 0.87 relevance
   Tools: web_search, view | Files: education.go

2. [Feb 21] School data discussion - 0.72 relevance
   Tools: bash, grep | Files: data/schools.csv

3. [Feb 15] China research notes - 0.54 relevance
   Tools: fetch | Files: notes.md

Use 'drill <number>' for details on a specific session.
```

**Phase 2: Drill-Down** (~200 tokens per session)

```
Session: Education system analysis (Feb 24, 2026)

SUMMARY:
You were analyzing China's 9-year compulsory education system
for a comparison with the US model. Key decision was to use
UNESCO cached data instead of live API calls.

FILES TOUCHED:
- internal/education/comparison.go (created)
- data/unesco_cache.json (read)

KEY DECISIONS:
- Use UNESCO data (cached) over live API
- Implement comparison as separate module
- Defer visualization to Phase 2

TOOLS USED:
- web_search: "China education statistics 2024"
- view: unesco_cache.json
- write: comparison.go

CODE SNIPPET:
func compareSystems(cn, us EducationData) ComparisonResult {...}
```

### Token Budget Enforcement

```go
const (
    MaxCatalogTokens = 100   // Phase 1 output
    MaxDrillTokens   = 500   // Phase 2 per session
    MaxCodeSnippet   = 200   // Per code block
)

func truncateWithIndicator(text string, maxTokens int) string {
    // Truncate to token limit, add "..." indicator
    // Preserve structure (don't cut mid-code-block)
}
```

---

## Indexing Strategy

### When to Index

| Event | Action |
|-------|--------|
| Session created | Create placeholder in FTS |
| Message added | No immediate action (batch) |
| Tool executed | Queue for batch index |
| Session ends | Full index pass |
| Handoff created | Index immediately |

### What to Index

```
✅ INCLUDE:
- Session title (auto-generated)
- Tool calls and their purposes (inferred)
- Tool results (summarized, not full output)
- Code blocks (with language tag)
- File paths touched
- Decisions made (extracted from assistant messages)
- Entities mentioned

❌ EXCLUDE:
- Conversational filler ("let me check", "here you go")
- Full tool outputs (summarize to key points)
- Reasoning chains (too long, low value)
- Raw data (CSV contents, JSON blobs)
```

### Summary Generation

Each session gets an auto-generated summary at close:

```go
type SessionSummary struct {
    OneLiner     string   // "Analyzed Chinese education system"
    KeyDecisions []string // ["Use UNESCO cached data", ...]
    FilesTouched []string // ["internal/education/comparison.go", ...]
    ToolsUsed    []string // ["web_search", "view", ...]
    Entities     []string // ["China", "education", "UNESCO", ...]
}
```

Generated by small model at session end, stored in FTS.

---

## Query API

```go
type ArchiveQueryParams struct {
    Query       string `json:"query"`                 // Natural language query
    TimeRange   string `json:"time_range,omitempty"`  // "last week", "yesterday"
    ProjectPath string `json:"project_path,omitempty"`
    Tools       []string `json:"tools,omitempty"`     // Filter by tools used
    Files       []string `json:"files,omitempty"`     // Filter by files touched
    Limit       int     `json:"limit,omitempty"`       // Max results (default: 5)
    Drills      []int   `json:"drills,omitempty"`      // Session numbers to drill
}
```

### Example Queries

```
"Chinese education last week"
→ TimeRange: (7 days ago, now), Query: "Chinese education"

"that bug fix with the authentication"
→ Query: "bug fix authentication", Tools: maybe ["edit", "test"]

"when did I work on the API"
→ Query: "API", Return: chronological list

"session where I used web_search for China data"
→ Query: "China data", Tools: ["web_search"]
```

---

## Implementation Phases

### Phase 1: FTS5 Foundation (1-2 days)

1. Create `sessions_fts` virtual table
2. Create `sessions_meta` table
3. Implement basic indexing on session close
4. Implement FTS5 search query
5. Update `query_floyd_archive` to use FTS5

### Phase 2: Query Processor (2-3 days)

1. Implement time parser ("last week", "yesterday")
2. Add entity extraction (basic NER)
3. Implement query expansion with stemming
4. Add filtering by tools/files

### Phase 3: Result Layer (1-2 days)

1. Implement two-phase retrieval
2. Add token budget enforcement
3. Implement drill-down command
4. Add result summarization

### Phase 4: Auto-Summarization (2-3 days)

1. Generate session summaries at close
2. Extract key decisions from messages
3. Index summaries in FTS
4. Update handoff to reference searchable archive

### Phase 5: Integration (1 day)

1. Add to allowed tools by default
2. Auto-query on session start (optional)
3. Update handoff instructions
4. Add to boot sequence (optional)

---

## Token Efficiency Analysis

| Operation | Current | With FTS5 | Savings |
|-----------|---------|-----------|---------|
| Search query | ~500 tokens | ~50 tokens | 90% |
| View full session | ~10K tokens | ~500 tokens | 95% |
| Find relevant session | Scan all | Ranked results | Time + tokens |

**Key insight:** The index should answer "where is it?" not "what is it?". The drill-down fetches details only when needed.

---

## Future Enhancements

### Vector Embeddings (Phase 6+)

For truly semantic search ("that thing about schools in Asia"):

1. Embed session summaries with local model (all-MiniLM-L6-v2)
2. Store embeddings in SQLite with sqlite-vss
3. Hybrid search: FTS5 for keywords + vectors for semantics

### Cross-Session Learning

1. Track which sessions you return to
2. Boost ranking for frequently-accessed sessions
3. Learn your entity preferences

### Proactive Suggestions

1. On session start, suggest related past sessions
2. "Last time you worked on this file, you..."
3. Context-aware retrieval based on current task

---

## Conclusion

The current implementation filters out chitchat but can't find what you meant. FTS5 + structured indexing + two-phase retrieval gives you:

1. **Vague → Precise:** "China school thing" finds "Chinese education system"
2. **Time-aware:** "last week" actually works
3. **Token-efficient:** Card catalog first, drill-down only if needed
4. **Ranked:** Most relevant results first, not random order

Total implementation: ~7-10 days for full system, ~2 days for minimum viable FTS5.

---

*Spec version: 1.0*
*Created: 2026-03-01*
