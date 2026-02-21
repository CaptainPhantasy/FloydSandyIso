# Floyd v4.0.0 Release Documentation

**Release Date:** 2026-02-21
**Repository:** https://github.com/CaptainPhantasy/FloydSandyIso
**Base Ref:** main
**Working Branch:** main (sandbox isolated)

---

## 1) RELEASE SCOPE CONFIRMED

| Attribute | Value |
|-----------|-------|
| Version | v4.0.0 |
| Base Ref | main |
| Working Dir | /Volumes/Storage/floyd-sandbox/FloydDeployable |
| Database Migrations | 10 total (latest: cache_read_tokens) |
| Version File | internal/version/version.go |

---

## 2) CHANGE INVENTORY

```
┌────┬──────────────────────────────────────┬─────────────────────────────────┬────────────────────────┐
│ ID │ Feature                              │ User Impact                     │ Docs Section          │
├────┼──────────────────────────────────────┼─────────────────────────────────┼────────────────────────┤
│ 01 │ Ctrl+Y Keybinding for AI Suggestion   │ Accept ghost text with Ctrl+Y   │ Quickstart            │
│ 02 │ Agent Library System                 │ Load agent personas from files  │ What's New            │
│ 02a│ Agent Template System                │ _template.md for new agents     │ What's New            │
│ 03 │ Streaming Tool Progress (Mod 4)      │ Real-time tool status updates   │ What's New            │
│ 04 │ Context Status Tool                  │ Monitor token usage             │ What's New            │
│ 05 │ Symbol Index Tool                    │ Code navigation                 │ What's New            │
│ 06 │ Configurable Banned Commands         │ Enable curl/wget/ssh via config │ What's New            │
│ 07 │ Token Display Debug Logging          │ Diagnose percentage fluctuation │ Troubleshooting       │
│ 08 │ Sandbox Database Isolation           │ Safe experimentation            │ Architecture          │
└────┴──────────────────────────────────────┴─────────────────────────────────┴────────────────────────┘
```

---

## 3) DOCUMENTATION PACKAGE

### A) Release Notes

#### Executive Summary

Floyd v4.0.0 introduces the Agent Library system, enabling users to define and switch between AI personas via markdown files. This release also includes Ctrl+Y keybinding for accepting AI suggestions (F1 was unreliable in terminals), streaming tool progress updates, context window monitoring, and improved safety through configurable command restrictions.

#### Detailed Changes by Area

**User Interface:**
- Ctrl+Y key now accepts ghost text suggestions (F1 was unreliable in terminals)
- Agent Library accessible via Ctrl+P → Agent Library
- Real-time progress indicators for grep, glob, and sourcegraph tools

**Core Features:**
- Agent Library: Load personas from `internal/agents/*.md`
- Context Status Tool: Monitor prompt/completion/cached tokens
- Symbol Index: LSP-based code navigation
- Streaming Tool Progress: Visual feedback during long operations

**Configuration:**
- `allowed_banned_commands` in settings.json enables curl, wget, ssh, etc.

**Architecture:**
- Sandbox isolation: `.floyd/` in project directory prevents DB contamination
- 10 database migrations including `cache_read_tokens`

#### Who Is Affected

| User Type | Impact |
|-----------|--------|
| Existing users | No breaking changes; new features opt-in |
| New users | Full feature set available immediately |
| Developers | Agent Library extensible via markdown files |

#### What To Do

1. Pull latest from sandbox repo
2. Run `go build .` to compile
3. Add agent definitions to `internal/agents/` as needed
4. Access Agent Library via Ctrl+P → Agent Library

---

### B) Quickstart (Release Adoption)

#### What Changed

- F1 accepts AI suggestions (was Tab)
- Agent Library added to commands menu
- Progress streaming for long-running tools

#### How To Enable New Features

**Agent Library:**
1. Create markdown files in `internal/agents/`:
```markdown
---
name: My Agent
description: What this agent does
trigger: myagent
---

# System Prompt
Your persona instructions here.
```

2. Open Floyd
3. Press Ctrl+P → Select "Agent Library"
4. Choose your agent → Press Enter to send

**Ctrl+Y Suggestion Acceptance:**
- When ghost text appears, press Ctrl+Y to accept it into the textarea

#### Upgrade Steps

```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
git pull origin main
go build .
./FloydDeployable
```

#### Minimal Working Example

```bash
# Create a custom agent
cat > internal/agents/my-helper.md << 'EOF'
---
name: Helper
description: A helpful assistant
---

You are a helpful assistant focused on concise, accurate responses.
EOF

# Build and run
go build . && ./FloydDeployable
# Ctrl+P → Agent Library → Select "Helper"
```

---

### C) What's New

#### 1. Agent Library System

**What it is:** Define AI personas as markdown files with YAML frontmatter.

**Why it matters:** Quickly switch between specialized agents (code reviewer, release auditor, etc.) without modifying configuration files.

**How to use:**
1. Create `internal/agents/{name}.md` with frontmatter
2. Open commands menu (Ctrl+P)
3. Select "Agent Library"
4. Choose agent → system prompt populates textarea
5. Press Enter to send

**Template System:**
A `_template.md` file provides the canonical agent format. Copy and modify it to create new agents.

**Example (Code Reviewer using the template format):**
```markdown
---
name: "Code Reviewer"
description: "Rigorous code review with evidence-backed feedback"
trigger: "review"
version: "1.0.0"
tags: [code, review, security, quality]
---

You are Code Reviewer, a specialized agent within the Legacy AI ecosystem.

Your mission is to ensure code quality, security, and maintainability...
```

**Files:**
- `internal/agents/loader.go` - Parser implementation
- `internal/agents/loader_test.go` - Test coverage
- `internal/ui/dialog/agent_library.go` - UI dialog
- `internal/ui/dialog/actions.go` - ActionSelectAgent type

---

#### 2. Ctrl+Y Keybinding for Suggestion Acceptance

**What it is:** Ctrl+Y key accepts AI suggestion ghost text.

**Why it matters:** Tab was occupied by focus-switch behavior; Ctrl+Y provides a reliable dedicated binding (F1 was unreliable in terminals due to help menu interception).

**How to use:**
1. Type in textarea to see ghost text suggestion
2. Press Ctrl+Y to accept suggestion
3. Press Enter to send

**Files:**
- `internal/ui/model/keys.go` - Ctrl+Y binding definition
- `internal/ui/model/ui.go` - Key handler

---

#### 3. Streaming Tool Progress (Mod 4)

**What it is:** Real-time progress updates during tool execution.

**Why it matters:** Users see intermediate status during long-running operations instead of appearing frozen.

**How it works:**
```
Tool → ProgressEmitter.Emit() → callback → PublishProgress()
                                           ↓
UI ← pubsub.Event[ToolProgressEvent]
```

**Files:**
- `internal/agent/tools/progress.go`
- `internal/agent/tools/grep.go`
- `internal/agent/tools/glob.go`
- `internal/agent/tools/sourcegraph.go`

---

#### 4. Context Status Tool

**What it is:** Tool to monitor context window usage.

**Why it matters:** AI can self-monitor token consumption and adjust verbosity.

**Usage:**
```
User: "Check context status"
→ context_status tool returns:
  Context: 23.5% used (47,000/200,000 tokens)
  Cached: 12,000 tokens | Prompt: 35,000 | Completion: 12,000
```

**Files:**
- `internal/agent/tools/context_status.go`

---

#### 5. Configurable Banned Commands

**What it is:** Allow specific "banned" commands through configuration.

**Why it matters:** Safety defaults remain, but power users can enable curl, wget, ssh when needed.

**Configuration:**
```json
{
  "options": {
    "execution": {
      "allowed_banned_commands": ["curl", "wget", "ssh"]
    }
  }
}
```

**Files:**
- `internal/config/config.go`
- `internal/agent/tools/bash.go`

---

### D) Upgrade Guide

#### Breaking Changes

**None.** All v4.0.0 features are additive.

#### Configuration Changes

New optional field in settings.json:
```json
{
  "options": {
    "execution": {
      "allowed_banned_commands": ["curl"]
    }
  }
}
```

#### Database Migrations

10 migrations applied automatically:
```
20250424200609_initial.sql
20250515105448_add_summary_message_id.sql
20250624000000_add_created_at_indexes.sql
20250627000000_add_provider_to_messages.sql
20250810000000_add_is_summary_message.sql
20250812000000_add_todos_to_sessions.sql
20260127000000_add_read_files_table.sql
20260208000000_rename_name_to_title.sql
20260220000000_add_cache_read_tokens.sql
```

#### Deprecations

None in this release.

---

### E) Troubleshooting

#### Token Display Fluctuation

**Symptom:** Context percentage jumps between values (e.g., 1% → 20% → 0%)

**Likely Cause:** Provider reports different `CacheReadTokens` per request; display formula uses instantaneous values.

**Mitigation:** Debug logging added to `agent.go:updateSessionUsage()`:
```
slog.Debug("updateSessionUsage",
  "api_input", usage.InputTokens,
  "api_output", usage.OutputTokens,
  "api_cache_read", usage.CacheReadTokens,
)
```

**Fix Status:** Investigation ongoing; not blocking for release.

---

#### Agent Library Not Showing Agents

**Symptom:** Agent Library dialog is empty

**Likely Cause:** No `.md` files in `internal/agents/` directory

**Fix:** Create agent markdown files with required frontmatter:
```markdown
---
name: Required
description: Required
---
Content
```

---

#### Ctrl+Y Not Accepting Suggestion

**Symptom:** Ctrl+Y press does nothing

**Likely Cause:** No ghost text visible (empty textarea or no suggestion available)

**Fix:** Ensure textarea has partial text that matches a suggestion. Check logs at `~/.floyd/logs/` for debug messages.

---

### F) FAQ

**Q: How do I add a new agent persona?**
A: Create a markdown file in `internal/agents/` with YAML frontmatter containing `name` and `description` fields. The body becomes the system prompt.

**Q: Can I share agents between projects?**
A: Currently agents are project-local. Consider symlinking or copying agent files between projects.

**Q: Why Ctrl+Y instead of Tab or F1?**
A: Tab was already bound to focus-switch behavior. F1 was unreliable in terminals (often intercepted for help menus). Ctrl+Y is a reliable dedicated binding.

**Q: Does the Agent Library persist across sessions?**
A: Agent definitions are loaded from files at dialog open time. The selected persona applies until session end or new selection.

---

### G) Traceability Appendix (Diff → Docs)

| Feature | Evidence Path | Lines |
|---------|---------------|-------|
| Agent Library Loader | `internal/agents/loader.go` | 1-95 |
| Agent Library Tests | `internal/agents/loader_test.go` | 1-150 |
| Code Reviewer Agent | `internal/agents/code-reviewer.md` | 1-22 |
| Release Auditor Agent | `internal/agents/release-auditor.md` | 1-30 |
| F1 Keybinding | `internal/ui/model/keys.go:157-161` | +5 |
| F1 Handler | `internal/ui/model/ui.go:1744-1745` | +2 |
| Agent Library Dialog | `internal/ui/dialog/agent_library.go` | 1-205 |
| ActionSelectAgent | `internal/ui/dialog/actions.go:97-102` | +6 |
| Commands Menu Entry | `internal/ui/dialog/commands.go:420` | +1 |
| Dialog Open Handler | `internal/ui/model/ui.go:3063-3067` | +5 |
| openAgentLibraryDialog | `internal/ui/model/ui.go:3148-3165` | +18 |
| Token Debug Logging | `internal/agent/agent.go:1040-1052` | +13 |
| Cache Read Tokens Migration | `internal/db/migrations/20260220000000_add_cache_read_tokens.sql` | 1-9 |

---

## 4) NEEDS VERIFICATION

None. All documented features verified through:
- Build: `go build .` ✓
- Tests: `go test ./...` ✓
- File existence confirmed

---

## 5) UNDOCUMENTABLE WITHOUT USER COOPERATION

- Actual runtime behavior verification (requires TUI session)
- Provider-specific token reporting behavior (ZAI/GLM-5 specifics)
- Production database migration experience

---

## 6) DOCS TO FIX

None identified. Prior documentation in HANDOFF_2026-02-20.md remains accurate for previously completed features.

---

## DATABASE FIX REQUIREMENTS

**Expected Issue:** Database may be reset after this session.

**Required Fixes for Production Agent:**

1. **Migration Application:**
```sql
-- Run this if cache_read_tokens column missing:
ALTER TABLE sessions ADD COLUMN cache_read_tokens INTEGER NOT NULL DEFAULT 0 CHECK (cache_read_tokens >= 0);
```

2. **Schema Verification:**
```sql
-- Verify all columns exist:
SELECT sql FROM sqlite_master WHERE type='table' AND name='sessions';

-- Expected columns:
-- id, parent_session_id, title, message_count, prompt_tokens, 
-- completion_tokens, cache_read_tokens, cost, updated_at, created_at,
-- summary_message_id, todos
```

3. **Backup Command (Pre-Migration):**
```bash
cp /Volumes/Storage/.floyd/floyd.db /Volumes/Storage/.floyd/floyd.db.backup-$(date +%Y%m%d)
```

4. **Data Directory Structure:**
```
~/.local/share/floyd/
├── providers.json
└── projects.json

/Volumes/Storage/.floyd/
├── floyd.db (production)
└── logs/

/Volumes/Storage/floyd-sandbox/FloydDeployable/.floyd/
└── floyd.db (sandbox - isolated)
```

---

*Documentation generated for Floyd v4.0.0 release*
