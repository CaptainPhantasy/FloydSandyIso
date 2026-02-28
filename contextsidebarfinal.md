# Context Sidebar & Session Continuity System

> Implementation Design Document  
> Date: 2026-02-28  
> Project: FloydDeployable
> Current Build: v4.6
> Target Build: v4.7

---

## Build Status Summary

### Current State (v4.6 - Built and Tested)

The following changes have been implemented and locally tested in v4.6:

| Feature | Status | Files Modified |
|---------|--------|----------------|
| Dual Token Display (total/effective) | ✅ Complete | `internal/ui/common/elements.go` |
| Two-line Layout (usage + cache) | ✅ Complete | `internal/ui/common/elements.go` |
| Conservative Warning Thresholds | ✅ Complete | `internal/ui/model/header.go`, `elements.go` |
| MCP Section Removed from Sidebar | ✅ Complete | `internal/ui/model/sidebar.go` |
| Telemetry Stubbed (no-ops) | ✅ Complete | `internal/event/event.go` |
| Prompt Quota Timestamp Fix | ✅ Complete | `internal/ui/model/prompt_usage.go` |
| DB Migration Order Fix | ✅ Complete | `internal/db/connect.go` |
| Context Status Tool Enhanced | ✅ Complete | `internal/agent/tools/context_status.go` |
| Unit Tests Added | ✅ Complete | `internal/ui/common/elements_test.go` |

**Uncommitted Changes:**
```
modified:   internal/agent/tools/context_status.go
modified:   internal/db/connect.go
deleted:    internal/event/all.go
deleted:    internal/event/identifier.go
deleted:    internal/event/logger.go
modified:   internal/event/event.go
modified:   internal/event/event_test.go
modified:   internal/ui/common/elements.go
new file:   internal/ui/common/elements_test.go
modified:   internal/ui/model/header.go
modified:   internal/ui/model/prompt_usage.go
modified:   internal/ui/model/sidebar.go
```

### Target State (v4.7 - To Be Implemented)

When the features in this document are fully implemented and tested, the project should be built as **v4.7**.

**v4.7 will include:**
1. Sidebar Stoplight Indicator (🟢/🟡/🔴)
2. Auto-Export at 85% Context Threshold
3. Session Handoff System (HANDOFF.md)
4. Semantic Archive Tool (query_floyd_archive)

---

## Version History: v4.0 → v4.6 → v4.7

### v4.0.0 (Base Release)

**Agent Capabilities:**
- Tool Registry Boot Discovery (105+ MCP tools at startup)
- Environment State Caching (ecosystem component discovery)
- Parallel Bash Execution (4 concurrent jobs, 87% speedup)
- Workflow Engine with Checkpoints
- Symbol Index for Code Search (tree-sitter based)
- Multi-Agent Headless Workspace
- Vision/Multi-modal Input Pipeline

**UI/Performance:**
- Streaming Tool Progress
- Smart Context Compression
- MCP Health Monitoring + Auto-Restart
- SUPERCACHE Namespace Support

**Safety:**
- SafeOps Semantic Impact Analysis

### v4.5 (Pre-Sidebar Work)

- Context Caching Display (CacheReadTokens tracking)
- Error Preservation in Prompt Usage loader

### v4.6 (Current Build)

**UI Improvements:**
- Dual Token Display: Shows both `total` and `effective` tokens
- Two-line format: Usage on line 1, cache % on line 2
- Conservative warnings based on TOTAL tokens (not effective)
- MCP section removed from sidebar (redundant with auto-discovery)

**Bug Fixes:**
- Prompt quota timestamp comparison (seconds vs milliseconds)
- DB migration order (ensureColumns now runs AFTER goose.Up)

**Independence:**
- Telemetry fully stubbed (no PostHog calls)

### v4.7 (Target Build - This Document)

**Sidebar Stoplight:**
- GREEN (0-70%): Normal operation
- YELLOW (71-84%): Caution indicator
- RED (85%+): Critical + auto-export trigger

**Session Continuity:**
- Auto-export transcript to `.floyd/transcripts/`
- HANDOFF.md generation with session metadata
- New session boot detection of prior session
- Archive tool for querying past sessions

---

## Overview

A comprehensive system for:
1. Visual context monitoring via sidebar stoplight indicator
2. Automatic session export before context exhaustion
3. Seamless handoff to new sessions with persistent memory
4. Semantic archive tool for querying past sessions

---

## Part 1: Sidebar Stoplight Indicator

### Current Format (v4.6)

```
22% (45K total / 2K effective) $0.24
   95% cached
```

### Target Format (v4.7)

**Compact format (fits ~30 char sidebar width):**

```
NORMAL STATE (< 70%):
┌──────────────────────────────┐
│ 🟢 22% • 45K/2K • $0.24      │  Line 1: indicator • pct • total/effective • cost
│ 95% cached                  │  Line 2: cache hit rate (if > 0%)
└──────────────────────────────┘

CAUTION STATE (71-84%):
┌──────────────────────────────┐
│ 🟡 75% • 150K/45K • $0.80    │  Yellow indicator
│ 70% cached                  │
└──────────────────────────────┘

CRITICAL STATE (85%+):
┌──────────────────────────────┐
│ 🔴 88% • 176K/25K • $1.20    │  Red indicator
│ 86% cached                  │
│ ⚠ Near limit • Ctrl+N       │  Warning line with action hint
└──────────────────────────────┘
```

### Color Thresholds

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│   0% ════════════════ 70% ══════════ 85% ═══════════════════════ 100%  │
│   │        GREEN       │   YELLOW   │           RED                 │  │
│   │                    │            │            │                  │  │
│   │   Safe zone        │  Caution   │   CRITICAL - triggers         │  │
│   │   No warning       │  zone      │   auto-export & handoff       │  │
│   │                    │            │                              │  │
│   ▼                    ▼            ▼                              ▼  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

| Threshold | Color | State | Behavior |
|-----------|-------|-------|----------|
| 0% - 70% | 🟢 GREEN | Normal | No warning |
| 71% - 84% | 🟡 YELLOW | Caution | Visual indicator only |
| 85%+ | 🔴 RED | Critical | Auto-export + handoff prompt |

### Implementation Checklist - Phase 1

- [ ] Modify `internal/ui/common/elements.go` - `formatTokensAndCost()`
- [ ] Add stoplight color logic (GREEN/YELLOW/RED) based on percentage
- [ ] Change separators from `/` to `•` for compactness
- [ ] Change token format from `X total / Y effective` to `X/Y`
- [ ] Add warning line "⚠ Near limit • Ctrl+N" when ≥85%
- [ ] Update tests in `internal/ui/common/elements_test.go`

---

## Part 2: Auto-Export System

### Trigger
- **Primary:** 85% context usage (RED threshold)
- **Secondary:** Manual trigger via command

### Export Format

**File location:** `{project}/.floyd/transcripts/{session_id}_{timestamp}.md`

**File structure:**

```markdown
# Session Transcript: {session_title}
> Session ID: {session_id}
> Exported: {timestamp}
> Project: {working_dir}
> Model: {model_name}
> Final Context: {pct}% ({total} / {context_window} tokens)

---

## Conversation Summary

[Brief auto-generated summary if available, or "No summary generated"]

---

## Message Log

### User
{first user message preview...}

### Assistant
{first assistant response preview...}

### Tool: bash
```
{tool execution details}
```

### Tool: edit
```go
// {file_path}
{code changes}
```

[... continues for all messages ...]

---

## Tool Executions Index

| Timestamp | Tool | Key Parameters |
|-----------|------|----------------|
| 14:23:01 | bash | `git status` |
| 14:24:15 | edit | `internal/ui/model/sidebar.go` |
| 14:25:30 | grep | pattern: "context" |

---

## Files Modified

- internal/ui/model/sidebar.go
- internal/ui/common/elements.go
- internal/agent/tools/archive.go

---

## Key Decisions Made

1. Sidebar stoplight uses 70/85 thresholds
2. Auto-export triggers at 85%
3. Archive tool queries tool executions, not conversational text
```

### Implementation Checklist - Phase 2

- [ ] Create `internal/agent/tools/transcript_export.go`
- [ ] Define export markdown format
- [ ] Hook into context percentage check (trigger at 85%)
- [ ] Create `.floyd/transcripts/` directory structure
- [ ] Implement export at 85% threshold

---

## Part 3: Session Handoff

### HANDOFF.md Structure

**Location:** `{project}/HANDOFF.md`

```markdown
# FLOYD Session Handoff

> Last Updated: {timestamp}
> Previous Session: {session_id}

## Active Context

- **Working Directory:** {working_dir}
- **Branch:** {git_branch}
- **Last Task:** {brief description from last user message}

## Transcript Location

`.floyd/transcripts/{session_id}_{timestamp}.md`

## Quick Index

The transcript contains:
- **Tool Executions:** {count} (bash, edit, grep, etc.)
- **Files Modified:** {count}
- **Key Decisions:** {count}

## How to Continue

1. Use `query_floyd_archive` tool to search past tool executions
2. Reference the transcript file for full conversation history
3. Do NOT read the entire transcript verbatim (will consume context)

## Outstanding Items

- [ ] {any todos or pending tasks}

---

*This file is auto-generated by FLOYD when context reaches 85%*
```

### New Session Initialization Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     NEW SESSION BOOT SEQUENCE                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   1. Agent initializes (standard boot)                                  │
│          │                                                              │
│          ▼                                                              │
│   2. Check for HANDOFF.md in project root                              │
│          │                                                              │
│          ├── [EXISTS] ─────────────────────────────────┐                │
│          │                                              │                │
│          │   3a. Read HANDOFF.md (small file)          │                │
│          │          │                                  │                │
│          │          ▼                                  │                │
│          │   3b. Note transcript path                  │                │
│          │          │                                  │                │
│          │          ▼                                  │                │
│          │   3c. DO NOT read transcript fully          │                │
│          │       (it's indexed, not loaded)            │                │
│          │          │                                  │                │
│          │          ▼                                  │                │
│          │   3d. Report to user:                       │                │
│          │       "Continuing from previous session."   │                │
│          │       "Use archive tool to query history."  │                │
│          │                                              │                │
│          ├── [NOT EXISTS] ─────────────────────────────┤                │
│          │                                              │                │
│          │   3f. Standard fresh session                 │                │
│          │       No handoff message                     │                │
│          │                                              │                │
│          ▼                                              ▼                │
│                                                                         │
│   4. Agent ready for input                                              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Implementation Checklist - Phase 3

- [ ] Create/update HANDOFF.md on export
- [ ] Modify boot sequence to check HANDOFF.md
- [ ] Display "Continuing from previous session" message
- [ ] Store transcript path for reference

---

## Part 4: Semantic Archive Tool

### Purpose
Query past sessions for technical content WITHOUT loading entire transcripts into context.

### The Semantic Firewall

**Key insight:** We only query tool executions and code blocks, ignoring conversational filler.

This prevents "persona drift" where the LLM might pick up old personality quirks or irrelevant conversational context.

### What Gets Indexed

| Message Type | Indexed? | Reason |
|--------------|----------|--------|
| Assistant with tool_calls | ✅ YES | Technical action taken |
| Tool result messages | ✅ YES | Hard output, factual |
| User with code blocks | ✅ YES | Code context provided |
| User plain text | ❌ NO | Conversational, not technical |
| Assistant plain text | ❌ NO | Prevents persona drift |

### Implementation Checklist - Phase 4

- [ ] Create `internal/agent/tools/archive.go`
- [ ] Implement semantic firewall query
- [ ] Register tool in agent tool list
- [ ] Add tool description for LLM guidance

---

## Part 5: Dialog Integration

### Implementation Checklist - Phase 5

- [ ] Create context warning dialog component
- [ ] Implement [Continue] [New Session] [View Transcript] actions
- [ ] Test dialog appearance at 85%

---

## Deployment Checklist

### Before Building v4.7

1. **Complete all Phase 1-5 implementation items**
2. **Run full test suite:**
   ```bash
   cd /Volumes/Storage/floyd-sandbox/FloydDeployable
   go test ./...
   ```
3. **Update version:**
   ```bash
   # Edit internal/version/version.go
   # Change: var Version = "v4.6"
   # To:     var Version = "v4.7"
   ```
4. **Update changelog.go** with v4.7 entries
5. **Build and smoke test:**
   ```bash
   go build -o floyd .
   ./floyd --version  # Should show v4.7
   ```

### After Successful v4.7 Build

1. **Commit changes:**
   ```bash
   git add -A
   git commit -m "feat: v4.7 - Context Sidebar & Session Continuity System

   - Sidebar stoplight indicator (GREEN/YELLOW/RED)
   - Auto-export at 85% context threshold
   - Session handoff via HANDOFF.md
   - Semantic archive tool (query_floyd_archive)
   "
   ```
2. **Create release tag:**
   ```bash
   git tag v4.7
   ```

---

## File Structure

```
{project}/
├── .floyd/
│   ├── transcripts/
│   │   ├── abc123_20260228_142000.md
│   │   └── def456_20260228_160000.md
│   └── floyd.db
├── HANDOFF.md
└── [project files...]

internal/agent/tools/
├── archive.go           # Semantic archive tool (NEW in v4.7)
├── transcript_export.go # Export functionality (NEW in v4.7)
└── [existing tools...]

internal/ui/
├── common/
│   └── elements.go      # Sidebar formatting (stoplight) - MODIFIED in v4.7
├── model/
│   └── sidebar.go       # Sidebar rendering - MODIFIED in v4.6
└── dialog/
    └── context_warning.go # Handoff dialog (NEW in v4.7)
```

---

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Sidebar indicator, not status bar | User preference - sidebar is the context home |
| 70/85 thresholds | 70% yellow = early warning, 85% red = action needed |
| Auto-export at 85% | Leaves 15% buffer for final exchanges before handoff |
| Semantic firewall | Prevents persona drift, focuses on technical content |
| HANDOFF.md | Small pointer file, transcript stored separately |
| Don't auto-load transcript | Would fill context window immediately |
| Archive tool on-demand | User/agent queries only what's needed |
| Version bump to v4.7 | Significant new features warrant minor version increment |

---

## End of Document

*This document serves as the implementation guide for v4.7. The current build is v4.6 with dual token display and related fixes. When the features in this document are implemented, build as v4.7.*
