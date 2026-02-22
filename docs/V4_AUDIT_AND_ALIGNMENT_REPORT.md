# FLOYD v4.0 Audit & Alignment Report

**Generated:** 2026-02-22
**Session Context:** Deep analysis of GLM-5 harness programming gaps
**Status:** CRITICAL - Fixes Required Before Further Development

---

## Executive Summary

This audit was triggered by a session export analysis that revealed a pattern of **action-before-discovery** failures. The root cause is not a single bug but a systematic absence of constraints in areas where GLM-5 requires them most.

**Key Finding:** GLM-5 performs optimally with TIGHT constraints (temperature 0.1 achieved 100% on benchmarks vs 75% with looser settings). The gaps identified are places where constraints are **missing or too loose**, not too tight.

---

## Bugs Identified

### BUG #1: No Pre-Action Discovery Gate

**Location:** `internal/agent/templates/coder.md.tpl` lines 179-188

**Problem:** SILENT REASONING PROTOCOL jumps to "consider 3 approaches" without first requiring discovery of current state.

**Evidence:** Agent created `mkdir` commands without verifying:
- Whether directory already existed
- Where tools should be installed
- What was in SUPERCACHE about tools

**Fix:**
```markdown
## 0. DISCOVERY GATE (MANDATORY BEFORE ACTION)
Before any action that modifies state:
1. STATE what you're about to do
2. VERIFY current state:
   - [ ] Checked SUPERCACHE for relevant keys
   - [ ] Checked filesystem for existing tools/files
   - [ ] Checked known locations for resources
3. LIST uncertainties requiring user input
4. IF uncertainties > certainties: ASK before proceeding
```

---

### BUG #2: No Tool Registry in SUPERCACHE

**Location:** `FLOYD.md` boot sequence (lines 11-16)

**Problem:** Agent has no way to discover installed tools on boot.

**Evidence:** When needing terminal-shadow, agent had no cached knowledge of its existence at `/Volumes/Storage/floyd-sandbox/FloydDeployable/terminal-shadow/`

**Fix:** Add to boot sequence:
```markdown
5. **Load Tool Registry:** `cache_retrieve(key="system:tool_registry")`
6. **Load Environment:** `cache_retrieve(key="system:environment_state")`
```

**Required SUPERCACHE Keys (to be populated):**
```
system:tool_registry = {
  "terminal-shadow": {
    "path": "/Volumes/Storage/floyd-sandbox/FloydDeployable/terminal-shadow",
    "cli": "shadow.py",
    "hook": "floyd_shadow_hook.py",
    "status": "sandbox"
  },
  // ... other tools
}

system:environment_state = {
  "global_tool_paths": [
    "~/.local/bin",
    "/usr/local/bin",
    "/opt/floyd-tools"
  ],
  "sandbox_path": "/Volumes/Storage/floyd-sandbox/FloydDeployable",
  "mcp_servers_path": "/Volumes/Storage/MCP"
}
```

---

### BUG #3: No ANALYSIS MODE Definition

**Location:** `FLOYD.md` MODE SELECTOR (lines 24-31)

**Problem:** When told to "analyze this session export," agent treated it as reading content, not as verifying state and applying findings to self.

**Evidence:** Agent read `cache_search returned 0 results` from export but never ran its own cache_search to verify.

**Fix:**
```markdown
- **ANALYSIS MODE** → examining logs, exports, session data

When in ANALYSIS MODE:
1. Extract claims about system state from data
2. For each claim, verify against CURRENT state (not assumed)
3. Apply relevant findings to YOURSELF
4. State explicitly: "This applies to me because..."
```

---

### BUG #4: No Tool Discovery Protocol

**Location:** New section needed

**Problem:** When agent needs a tool, it searches current directory with glob. If not found, assumes it doesn't exist.

**Evidence:** Agent started creating terminal-shadow when it existed elsewhere.

**Fix:**
```markdown
## IX. TOOL DISCOVERY PROTOCOL
When needing a tool:
1. Check `system:tool_registry` in SUPERCACHE
2. Check known tool directories IN ORDER:
   - /Volumes/Storage/floyd-sandbox/FloydDeployable/
   - /Volumes/Storage/MCP/
   - ~/.local/bin/
   - /usr/local/bin/
3. Check MCP Tools Reference (mcp_tools_reference.md)
4. If not found: ASK user before creating
5. NEVER create a tool that might already exist
```

---

### BUG #5: CREATE Has Same Friction as READ

**Location:** New section needed

**Problem:** `mkdir`, `write`, `create` feel as "easy" as `ls`, `cat`, `read`. No friction = no verification.

**Evidence:** Agent executed `mkdir -p ~/.local/lib/floyd-tools/terminal-shadow` immediately after being told sandbox should remain isolated.

**Fix:**
```markdown
## X. ACTION CLASSIFICATION

| Class | Actions | Required Behavior |
|-------|---------|-------------------|
| READ | ls, view, grep, cache_retrieve, glob | Free to execute |
| QUERY | search, check status | Free to execute |
| DISCOVER | verify state, check existence | Free to execute |
| WRITE_PROJECT | edit, write (in project dir) | Verify location first |
| CREATE | mkdir, new file | Verify doesn't exist, check if needed |
| INSTALL_GLOBAL | global tools, configs, symlinks | **ASK USER FIRST** |
| DELETE | rm, uninstall | **ASK USER + CONFIRM** |
```

---

### BUG #6: No Version Change Announcement

**Location:** Boot sequence / version management

**Problem:** v4.0 rolled out with new features (e.g., `context_status` tool) but agent was never informed.

**Evidence:** Agent discovered `context_status` tool by accident while auditing filesystem, not through any announcement or boot sequence.

**Fix:**
```markdown
**On version change, agent must receive:**
cache_retrieve(key="system:version_changelog")

Containing:
- What's new in this version
- New tools available
- Deprecated features
- Breaking changes
```

---

### BUG #7: No Ecosystem Context

**Location:** Fundamental architecture documentation

**Problem:** Agent exists within an ecosystem but was never given a map of what exists around it.

**Evidence:** Agent was unaware of:
- Desktop application (primary interface)
- Browser control capabilities
- Mobile application for communication
- Swarm deployment via desktop
- Hive coordination via CLI
- Headless swarms for single-task operations

**Fix:** Create `ECOSYSTEM.md`:
```markdown
## YOUR ECOSYSTEM

You exist within a larger system:

1. FLOYD DESKTOP APPLICATION
   - Primary interface for complex operations
   - Swarm deployment and management
   - Browser control integration
   - Visual task management

2. FLOYD CLI
   - Direct terminal operations
   - Hive coordination
   - Quick tasks and scripting

3. MOBILE APPLICATION
   - Communication channel with user
   - Alerts and notifications routed through desktop

4. AGENT TYPES AT YOUR DISPOSAL
   - Hives: Multi-task coordination (via CLI)
   - Swarms: Coordinated agent groups (via Desktop)
   - Headless: Single-task, race-condition-safe dispatch

5. EXTERNAL INTEGRATIONS
   - Browser control (navigation, interaction)
   - File system operations
   - Git operations
   - MCP server ecosystem (18 servers, 105+ tools)
```

---

## What's Working Well (Preserve These)

These constraints make GLM-5 MORE effective:

| Constraint | Evidence | Status |
|------------|----------|--------|
| Temperature 0.1 | 100% vs 75% on benchmarks | KEEP |
| think: false | Prevents 5-10 min delays | KEEP |
| DEBUG MODE hypothesis gate | Forces structured debugging | KEEP |
| CACHE TRUST POLICY | Facts > Decisions > Hypotheses | KEEP |
| HANDOFF "Lost Context Insurance" | Preserves reasoning across sessions | KEEP |
| "Cite evidence for claims" rule | Prevents hallucination | KEEP |
| Question discipline (max 1/reply) | Prevents overwhelming user | KEEP |
| Two-failure reset rule | Breaks wrong hypothesis chains | KEEP |

---

## Files Requiring Modification

| File | Changes |
|------|---------|
| `internal/agent/templates/coder.md.tpl` | Add DISCOVERY GATE before Silent Reasoning |
| `FLOYD.md` | Add tool_registry + environment to boot |
| `FLOYD.md` | Add ANALYSIS MODE to mode selector |
| `FLOYD.md` | Add TOOL DISCOVERY PROTOCOL section |
| `FLOYD.md` | Add ACTION CLASSIFICATION section |
| `FLOYD.md` | Add ECOSYSTEM context or reference |
| (new) `ECOSYSTEM.md` | Document full architecture |
| (new) SUPERCACHE entries | Populate system:tool_registry, system:environment_state |

---

## Known Tool Locations (For Reference)

```
/Volumes/Storage/floyd-sandbox/FloydDeployable/
├── terminal-shadow/          # Session continuity
├── internal/
│   ├── agent/                # Core agent logic
│   │   ├── templates/        # Prompt templates
│   │   └── tools/            # Built-in tools
│   └── intelligence/         # Symbol indexing

/Volumes/Storage/MCP/
├── floyd-supercache-server/  # 3-tier cache
├── floyd-runner/             # Test/lint/build
├── floyd-git/                # Git operations
├── floyd-explorer/           # Project mapping
├── floyd-patch/              # Diff operations
├── floyd-devtools/           # Type analysis
├── floyd-terminal/           # Process management
├── lab-lead-server/          # Agent spawning
├── hivemind-v2/              # Multi-agent coordination
├── context-singularity-v2/   # Context optimization
├── omega-v2/                 # Meta-cognitive reasoning
└── ... (18 total servers)
```

---

## Implementation Priority

1. **CRITICAL** (Before any other work):
   - Add DISCOVERY GATE to coder.md.tpl
   - Add ACTION CLASSIFICATION to FLOYD.md
   - Populate system:tool_registry in SUPERCACHE

2. **HIGH** (This session if possible):
   - Add TOOL DISCOVERY PROTOCOL
   - Add ANALYSIS MODE
   - Create ECOSYSTEM.md

3. **MEDIUM** (Soon):
   - Version changelog system
   - Environment state caching

---

## Key Insight

> The GLM-5 model performs best when constraints are TIGHT and EXPLICIT. Every gap in constraint is not freedom—it's uncertainty that degrades performance. The bugs found are not over-constraints; they are under-constraints in critical decision points.

---

## Session Metadata

- **Analysis Duration:** ~2 hours
- **Source Material:** Session export, FLOYD.md, coder.md.tpl, mcp_tools_reference.md, GLM_CONFIG_REFERENCE.md, filesystem audit
- **Stored To:** `system:v4_audit_findings` (SUPERCACHE vault tier)
- **Next Session:** Implement fixes, populate tool registry, create ecosystem documentation

---

*End of Report*
