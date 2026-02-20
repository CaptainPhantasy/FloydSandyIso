# FLOYD v4.0 EVOLUTION ROADMAP

**Created:** 2026-02-20
**Current Version:** v3.7.01
**Target Version:** v4.0
**Philosophy:** Stability first. Competitive through intelligence, not complexity.

---

## VERSION PROGRESSION

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  v3.7 ──► v3.8 ──► v3.9 ──► v3.95 ──► v4.0                                  │
│    │        │        │         │         │                                   │
│    │        │        │         │         └── Competitive Edge               │
│    │        │        │         └── Intelligence Layer                        │
│    │        │        └── Performance Layer                                   │
│    │        └── Reliability Layer                                            │
│    └── Current (Sandbox Isolated)                                            │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## LEVEL 1: RELIABILITY FOUNDATION (v3.7 → v3.8)

### Modification 1: MCP Health/Restart System

**Scope:** 1 week
**Files:** `internal/mcp/manager.go` (NEW), `internal/mcp/health.go` (NEW)

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  MCP Health/Restart Architecture                                             │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                     │
│   │ MCP Server  │    │ Health      │    │ Restart     │                     │
│   │ Manager     │◄──►│ Monitor     │◄──►│ Queue       │                     │
│   └─────────────┘    └─────────────┘    └─────────────┘                     │
│         │                  │                   │                             │
│         ▼                  ▼                   ▼                             │
│   ┌─────────────────────────────────────────────────────┐                   │
│   │  Per-Server State:                                   │                   │
│   │  - last_heartbeat: timestamp                        │                   │
│   │  - restart_count: int                               │                   │
│   │  - status: healthy|degraded|failed                  │                   │
│   │  - backoff: exponential (1s, 2s, 4s, 8s, max 60s)  │                   │
│   └─────────────────────────────────────────────────────┘                   │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Implementation:**
1. Health check goroutine (30s interval ping)
2. Auto-restart with exponential backoff
3. Max 5 restart attempts before marking failed
4. Structured error propagation to caller
5. Status reporting in UI (optional)

**Why First:** Highest reliability gain, smallest scope. MCP servers crash silently now.

---

### Modification 2: SUPERCACHE Namespaces

**Scope:** 1 week
**Files:** `MCP/floyd-supercache-server/src/index.ts`

**What:**
```typescript
// Current
cache_store(key: string, value: any) → void
cache_retrieve(key: string) → any

// After
cache_store(key: string, value: any, namespace?: string) → void
cache_retrieve(key: string, namespace?: string) → any
```

**Default Behavior:**
- `namespace` defaults to `"global"` (backwards compatible)
- Project-specific: namespace = project name or hash
- System keys: namespace = `"system"`

**Why Second:** Foundation for multi-project support. Prevents key collisions before we add more features.

---

## LEVEL 2: PERFORMANCE LAYER (v3.8 → v3.9)

### Modification 3: Parallel Bash Execution

**Scope:** 1-2 weeks
**Files:** `internal/agent/tools/bash.go`, `internal/execution/parallel.go` (NEW)

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Parallel Bash Execution Flow                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Floyd receives multiple bash commands:                                    │
│                                                                              │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  1. ls -la /src                                                      │   │
│   │  2. npm test                                                         │   │
│   │  3. git status                                                       │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                               │
│                              ▼                                               │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Dependency Analysis:                                                │   │
│   │  - Independent? → Execute in parallel                               │   │
│   │  - Dependent? → Execute sequentially                                │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                              │                                               │
│              ┌───────────────┼───────────────┐                              │
│              ▼               ▼               ▼                              │
│         ┌────────┐      ┌────────┐      ┌────────┐                         │
│         │ Job 1  │      │ Job 2  │      │ Job 3  │                         │
│         │ ls -la │      │ npm    │      │ git    │                         │
│         └────────┘      └────────┘      └────────┘                         │
│              │               │               │                              │
│              └───────────────┴───────────────┘                              │
│                              │                                               │
│                              ▼                                               │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  Result Synthesis:                                                   │   │
│   │  - Collate outputs                                                  │   │
│   │  - Preserve order metadata                                          │   │
│   │  - Single coherent response to LLM                                  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Safety Rules:**
- Max 4 concurrent jobs
- Read-only ops can always parallelize
- Write ops require explicit dependency declaration
- 30s timeout per job (configurable)

**Why Third:** 87% speedup on typical exploration tasks. No multi-agent token multiplication.

---

### Modification 4: Streaming Tool Progress

**Scope:** 1 week
**Files:** `internal/ui/model/messages.go`, `internal/agent/tools/*.go`

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Current vs After                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  CURRENT:                                                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  [Tool: bash] Running...                                             │    │
│  │  (user waits, nothing happens, then full output appears)            │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  AFTER:                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  [Tool: bash] Running npm install...                                 │    │
│  │  ├─ Installing dependencies...                                       │    │
│  │  ├─ added 847 packages in 12s                                        │    │
│  │  └─ Done                                                             │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Add `OnProgress` callback to tool interface
- Stream stdout/stderr in real-time
- Truncate in UI (show last N lines)
- Progress indicator for long ops

**Why Fourth:** Critical for user trust. Long operations feel broken without feedback.

---

## LEVEL 3: INTELLIGENCE LAYER (v3.9 → v3.95)

### Modification 5: Codebase Symbol Index

**Scope:** 2-3 weeks
**Files:** `internal/intelligence/symbols.go` (NEW), `internal/intelligence/parser_*.go`

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Symbol Index Architecture                                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   Source Files ──► Tree-sitter Parser ──► Symbol Extractor                  │
│                                              │                               │
│                                              ▼                               │
│                                    ┌─────────────────┐                       │
│                                    │  Symbol Index   │                       │
│                                    │  ─────────────  │                       │
│                                    │  name: string   │                       │
│                                    │  kind: func|var │                       │
│                                    │  file: path     │                       │
│                                    │  line: int      │                       │
│                                    │  signature: str │                       │
│                                    │  docstring: str │                       │
│                                    └─────────────────┘                       │
│                                              │                               │
│                                              ▼                               │
│                                    ┌─────────────────┐                       │
│                                    │  Query API      │                       │
│                                    │  ─────────────  │                       │
│                                    │  fuzzy_search() │                       │
│                                    │  by_kind()      │                       │
│                                    │  by_file()      │                       │
│                                    │  dependencies() │                       │
│                                    └─────────────────┘                       │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Supported Languages (Phase 1):**
- Go (tree-sitter-go)
- TypeScript (tree-sitter-typescript)
- Python (tree-sitter-python)

**Why Fifth:** Semantic search without embeddings dependency. Fast, deterministic, no ML infrastructure needed.

---

### Modification 6: Smart Context Compression

**Scope:** 2 weeks
**Files:** `internal/agent/summarizer.go`, `internal/agent/context.go`

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Context Compression Strategy                                                │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Current Problem:                                                            │
│  ─────────────────                                                           │
│  Summarization throws away too much context. Important details get lost.    │
│                                                                              │
│  Solution: Tiered Compression                                                │
│  ─────────────────────────────────                                           │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  TIER 1: Preserve Forever (Never Summarize)                         │    │
│  │  - System prompt                                                     │    │
│  │  - Project configuration                                             │    │
│  │  - User's explicit requirements                                     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  TIER 2: Compress Intelligently                                     │    │
│  │  - Tool calls/results → structured summary                          │    │
│  │  - Exploration paths → "Searched X, found Y"                        │    │
│  │  - Error resolution → "Fixed bug in Z by doing W"                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  TIER 3: Discard                                                    │    │
│  │  - Duplicate information                                            │    │
│  │  - Failed exploration branches                                      │    │
│  │  - Verbose output (already processed)                               │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Implementation:**
- Tag messages with importance tier
- Smarter summarization prompts
- Preserve tool call signatures, compress results
- Track what was summarized for potential recall

**Why Sixth:** Context window is our scarcest resource. Better compression = longer effective sessions.

---

## LEVEL 4: COMPETITIVE EDGE (v3.95 → v4.0)

### Modification 7: Multi-Modal Input Pipeline

**Scope:** 2-3 weeks
**Files:** `internal/agent/tools/vision.go` (NEW), `internal/agent/message.go`

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Multi-Modal Input Pipeline                                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Input Sources:                                                              │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                        │
│  │ Paste   │  │ Drag &  │  │ Screen  │  │ URL     │                        │
│  │ Image   │  │ Drop    │  │ Capture │  │ Fetch   │                        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘                        │
│       │            │            │            │                              │
│       └────────────┴────────────┴────────────┘                              │
│                          │                                                   │
│                          ▼                                                   │
│              ┌─────────────────────┐                                        │
│              │  Image Processor    │                                        │
│              │  ─────────────────  │                                        │
│              │  - Resize if needed │                                        │
│              │  - Convert format   │                                        │
│              │  - Extract metadata │                                        │
│              └─────────────────────┘                                        │
│                          │                                                   │
│                          ▼                                                   │
│              ┌─────────────────────┐                                        │
│              │  Vision MCP Call    │                                        │
│              │  (ZAI or local)     │                                        │
│              └─────────────────────┘                                        │
│                          │                                                   │
│                          ▼                                                   │
│              ┌─────────────────────┐                                        │
│              │  Description        │                                        │
│              │  injected into      │                                        │
│              │  conversation       │                                        │
│              └─────────────────────┘                                        │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Use Cases:**
- Screenshot analysis ("What's wrong with this UI?")
- Diagram interpretation
- Error message screenshots
- Design mockup to code

**Provider Support:**
- ZAI MCP (cloud)
- Local models (future)
- Claude native (if available)

**Why Seventh:** Big labs push multimodal hard. Users expect it. Critical for debugging visual issues.

---

### Modification 8: Agentic Workflow Engine

**Scope:** 3-4 weeks
**Files:** `internal/workflow/engine.go` (NEW), `internal/workflow/steps.go`

**What:**
```
┌──────────────────────────────────────────────────────────────────────────────┐
│  Agentic Workflow Engine                                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Goal: Chain complex operations with persistent state and recovery          │
│                                                                              │
│  Example Workflow: "Refactor Authentication"                                │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Step 1: ANALYZE                                                     │    │
│  │  - Find all auth-related files                                       │    │
│  │  - Extract current patterns                                          │    │
│  │  - Identify dependencies                                             │    │
│  │  Status: ✓ Complete                                                  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                          │                                                   │
│                          ▼                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Step 2: PLAN                                                        │    │
│  │  - Generate refactoring plan                                         │    │
│  │  - User approval gate                                                │    │
│  │  Status: ⏸ Awaiting approval                                        │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                          │                                                   │
│                          ▼                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Step 3: EXECUTE                                                     │    │
│  │  - Apply changes file by file                                        │    │
│  │  - Run tests after each file                                         │    │
│  │  Status: ▶ In Progress (3/12 files)                                 │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                          │                                                   │
│                          ▼                                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Step 4: VERIFY                                                      │    │
│  │  - Full test suite                                                   │    │
│  │  - Integration tests                                                 │    │
│  │  Status: ○ Pending                                                   │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  Recovery: If interrupted, resume from last checkpoint                      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

**Core Features:**
- Workflow definition (YAML or code)
- Checkpoint persistence
- User approval gates
- Automatic rollback on failure
- Progress visualization

**Pre-defined Workflows:**
- Feature implementation
- Bug fix with verification
- Code review and cleanup
- Migration scripts

**Why Eighth:** This is the "agentic" differentiator. Not just chat—structured goal pursuit.

---

## IMPLEMENTATION PRIORITY ORDER

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  SEQUENCE RATIONALE                                                          │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1-2 (Reliability)     Must be rock-solid before adding features           │
│  3-4 (Performance)     User experience foundation                           │
│  5-6 (Intelligence)    Build on stable, fast base                          │
│  7-8 (Competitive)     Final polish and differentiation                    │
│                                                                              │
│  Each pair is a stable checkpoint. Ship v3.8, verify, then v3.9, etc.      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## COMPETITIVE POSITIONING

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  HOW WE STAY COMPETITIVE                                                     │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Big Labs Push:              Our Response:                                  │
│  ─────────────────           ─────────────                                  │
│  Reasoning models (o1/o3)    GLM-5 with internal thinking                   │
│  Tool calling                MCP ecosystem (extensible)                     │
│  Prompt caching              ✓ Already have (fixed)                         │
│  Multimodal                  Mod 7 (image understanding)                    │
│  Long context                Mod 6 (smart compression)                      │
│  Computer use                Mod 8 (workflow engine)                        │
│                                                                              │
│  Our Advantages:                                                             │
│  ────────────────                                                            │
│  - Multi-provider (not locked to one lab)                                   │
│  - Local-first (privacy, no cloud required)                                 │
│  - Extensible via MCP (community tools)                                     │
│  - Transparent (all operations visible)                                     │
│  - Self-hostable (full control)                                             │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## SELF-MODIFICATION PROTOCOL

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  DEVELOPMENT WORKFLOW                                                        │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. Work in /Volumes/Storage/floyd-sandbox/floyd-next                       │
│  2. Create feature branch: feature/mod-N-description                        │
│  3. Implement + test locally                                                │
│  4. Present diff to user for approval                                       │
│  5. User tests: old+new together                                            │
│  6. Merge to main on approval                                               │
│  7. Tag release: v3.8, v3.9, v3.95, v4.0                                    │
│                                                                              │
│  RULE: Never modify running instance. Always sandbox first.                 │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## TIMELINE ESTIMATE

| Level | Modifications | Duration | Target Version |
|-------|---------------|----------|----------------|
| 1 | MCP Health + SUPERCACHE Namespaces | 2 weeks | v3.8 |
| 2 | Parallel Bash + Streaming Progress | 2-3 weeks | v3.9 |
| 3 | Symbol Index + Context Compression | 4-5 weeks | v3.95 |
| 4 | Multi-Modal + Workflow Engine | 5-7 weeks | v4.0 |

**Total:** 13-17 weeks to v4.0

---

## SUCCESS METRICS

| Level | Metric | Target |
|-------|--------|--------|
| 1 | MCP uptime | 99%+ (auto-recovery) |
| 2 | Operation speed | 50%+ faster on multi-op tasks |
| 3 | Context efficiency | 2x effective context window |
| 4 | User capability | Handle visual inputs, complex workflows |

---

## EXPLICITLY NOT DOING (YET)

| Item | Reason |
|------|--------|
| Multi-Agent TMUX | Token multiplication, complexity |
| Full embeddings search | Start with symbols (Mod 5) |
| Session branching | Needs more thought |
| Docker sandbox | Optional, many users lack Docker |
| Voice input | Nice-to-have, not core |

---

*Awaiting license restoration and go/no-go on Level 1.*
