# FLOYD ECOSYSTEM - v4.0 Current Architecture

**Generated:** 2026-02-22
**Purpose:** Complete map of the FLOYD v4.0 ecosystem with clear distinction between CURRENT (v4.0) and LEGACY components

---

## QUICK REFERENCE: What's ACTIVE vs LEGACY

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    FLOYD v4.0 - ACTIVE COMPONENTS                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ✓ FloydDeployable/     - CURRENT v4.0 CLI (main development)                │
│  ✓ floyd4/floyd         - v4.0.0 executable (2026-02-21 build)               │
│  ✓ FloydDesktopWeb-v2/  - CURRENT desktop app (2026-02-19 updated)          │
│  ✓ MCP/ (18 servers)     - Current MCP ecosystem                            │
│  └─────────────────────────────────────────────────────────────────────────────│
│                                                                                 │
│                    LEGACY / HISTORICAL - DO NOT USE                           │
│  ✗ floyd-main/          - Old main branch (superseded by FloydDeployable)  │
│  ✗ floyd-harness/       - Old harness configs (2025-02-18)                  │
│  ✗ floyd-benchmark/     - Old benchmarking (2025-02-11)                    │
│  ✗ FloydDesktopWeb-Standalone/ - Old desktop build (2024-02-19)           │
│  ✗ CLI AGENT BUILDS/     - Old CLI builds (2025-11-09)                      │
│  ✗ Legacy Agents/       - Pre-v4 agent versions                             │
│  ✗ TUI-Rebuild-v2-MCP/   - Old TUI experiment (2025-02-03)                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## I. CURRENT: FLOYD CLI v4.0 (Core Agent)

**Location:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/`

**Status:** ✅ **ACTIVE - Primary Development**

**Last Updated:** 2026-02-21 (v4.0.0 release)

**Key Characteristics:**
- Built from `FloydSandyIso` repository (main branch)
- Sandbox database isolation (`.floyd/` in project directory)
- 10 database migrations applied
- Agent Library system with markdown-based personas
- Streaming tool progress
- Context status monitoring
- GLM-4.6/GLM-5 optimized (temperature 0.1, tool_stream enabled)

**Directory Structure:**
```
floyd-sandbox/FloydDeployable/
├── floyd4                   # v4.0.0 executable (83MB, 2026-02-21)
├── floyd                    # Symlink/alternative build (83MB, 2025-02-19)
├── cmd/                     # CLI entry points
│   ├── root.go             # Main command handler
│   ├── run.go              # Non-interactive mode
│   └── ai.go               # Interactive mode
├── internal/
│   ├── agent/              # Core agent logic
│   │   ├── coordinator.go  # Agent orchestration
│   │   ├── agent.go        # SessionAgent implementation
│   │   ├── templates/      # Prompt templates
│   │   │   ├── coder.md.tpl          # Main system prompt
│   │   │   ├── task.md.tpl           # Task agent prompt
│   │   │   └── mcp_tools_reference.md # MCP tool catalog
│   │   └── tools/          # Built-in tools
│   ├── app/                # Application wiring
│   ├── config/             # Configuration management
│   ├── db/                 # SQLite migrations (10 total)
│   ├── mcp/                # MCP client integration
│   └── ui/                 # Bubbletea TUI
├── terminal-shadow/        # Session continuity tool (NEW)
├── internal/agents/        # Agent Library markdown files
│   ├── _template.md        # Canonical agent format
│   ├── code-reviewer.md    # Code review persona
│   └── release-auditor.md  # Release audit persona
├── .floyd/                 # Sandbox database (project-isolated)
├── FLOYD.md                # Agent protocol (286 lines)
├── floyd.json              # Project config
├── floyd-schema.json       # JSON schema for config
├── docs/                   # Documentation
└── templates/              # Go templates
```

**Configuration Files:**
```
.floyd/floyd.json          # Project-specific config (highest priority)
~/.floyd/floyd.json        # Global user config
floyd.json                 # Local project config (can be git-tracked)
```

**v4.0 Features (from RELEASE_v4.0.0.md):**
1. Agent Library System - Load personas from markdown files
2. Ctrl+Y keybinding - Accept AI suggestions
3. Streaming Tool Progress - Real-time status updates
4. Context Status Tool - Monitor token usage
5. Symbol Index Tool - LSP-based navigation
6. Configurable Banned Commands - Allow curl/wget/ssh
7. Token Display Debug Logging
8. Sandbox Database Isolation

---

## II. CURRENT: FLOYD Desktop Application v2

**Location:** `/Volumes/Storage/FloydDesktopWeb-v2/`

**Status:** ✅ **ACTIVE - Primary Desktop Interface**

**Last Updated:** 2026-02-19 (Feb 19, 2025 timestamps indicate this is the current active desktop app)

**Architecture:**
```
FloydDesktopWeb-v2/
├── electron/              # Electron main process
├── src/                   # React frontend
│   ├── components/        # UI components
│   ├── hooks/             # React hooks
│   └── lib/               # Utilities
├── server/                # Backend server
│   ├── mcp/               # MCP integration
│   ├── ws-mcp-server.ts   # WebSocket MCP server (KEY integration point)
│   └── tool-executor.ts   # Tool execution handler
├── release/               # Built releases (Mac, Windows)
└── docs/                  # Documentation
```

**Integration with CLI:**
- WebSocket MCP server connects to CLI's MCP client
- Can spawn and coordinate swarms of agents
- Browser control capabilities via Chrome extension

---

## III. LEGACY: Old/Superseded Components

### A. Old CLI Builds

| Path | Status | Reason |
|------|--------|--------|
| `/Volumes/Storage/floyd-main/` | ✗ LEGACY | Old main branch, superseded by FloydDeployable |
| `/Volumes/Storage/floyd-harness/` | ✗ LEGACY | Old harness configuration (Feb 2025) |
| `/Volumes/Storage/floyd-benchmark/` | ✗ LEGACY | Old benchmarking setup (Feb 2025) |

### B. Old Desktop Builds

| Path | Status | Reason |
|------|--------|--------|
| `/Volumes/Storage/FloydDesktopWeb-Standalone/` | ✗ LEGACY | Old standalone build (Feb 2024) |

### C. Old Agent Implementations

| Path | Status | Reason |
|------|--------|--------|
| `/Volumes/Storage/CLI AGENT BUILDS/` | ✗ LEGACY | Pre-v4 CLI builds (Nov 2025) |
| `/Volumes/Storage/Legacy Agents/` | ✗ LEGACY | Pre-v4 agent versions |
| `/Volumes/Storage/TUI-Rebuild-v2-MCP/` | ✗ LEGACY | Old TUI experiment (Feb 2025) |

### D. Correction: floyd-harness

| Path | Status | Purpose |
|------|--------|---------|
| `/Volumes/Storage/floyd-harness/` | ✅ ACTIVE | API harness for using Floyd in other applications (NOT legacy) |

**Note:** Timestamp shows Feb 2025 but this is a typo - it's an active API harness.

---

## IV. CURRENT: Chrome Extension (Browser Control)

**Location:** `/Volumes/Storage/FLOYD Extension for Chrome/FloydChromeBuild/floydchrome/`

**Status:** ✅ **ACTIVE - Browser Automation Layer**

**Last Updated:** 2025-02-15 (handoff confirms operational)

**Purpose:** Enable Computer Use workflows - agent sees and controls browser

**Architecture:**
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    FloydChrome Extension Architecture                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  FLOYD CLI ──MCP──▶ FloydDesktop WebSocket Server (port 3005)                   │
│                           │                                                  │
│                           ▼                                                  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                    FloydChrome Extension                               │  │
│  │  ┌────────────────────────────────────────────────────────────────┐    │  │
│  │  │  Background Service Worker                                        │    │  │
│  │  │  - WebSocket MCP client (connects to port 3005)                  │    │  │
│  │  │  - Tool executor and router                                      │    │  │
│  │  │  - Safety layer (permissions, sanitization)                     │    │  │
│  │  └────────────────────────────────────────────────────────────────┘    │  │
│  │                           │                                          │           │  │
│  │  ┌────────────────────────────────────────────────────────────────┐    │  │
│  │  │  TOOLS (6 total)                                                  │    │  │
│  │  │  • browser_navigate    - Navigate to URLs                       │    │  │
│  │  │  • browser_read_page   - Get accessibility tree              │    │  │
│  │  │  • browser_screenshot  - Capture screenshots (Computer Use)    │    │  │
│  │  │  • browser_click       - Click elements                      │    │  │
│  │  │  • browser_type        - Type text                           │    │  │
│  │  │  • browser_get_tabs    - List all tabs                       │    │  │
│  │  └────────────────────────────────────────────────────────────────┘    │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                           │                                                  │
│                           ▼                                                  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                    Google Chrome Browser APIs                            │  │
│  │                    (Debugger, Tabs, Scripting)                           │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Tools Reference:**

| Tool | Description | Input | Returns |
|------|-------------|-------|--------|
| `browser_navigate` | Navigate to URL | `url`, `tabId?` | Navigation confirmation |
| `browser_read_page` | Get accessibility tree | `tabId?` | DOM structure |
| `browser_screenshot` | Capture screenshot (NEW) | `fullPage?`, `selector?`, `tabId?` | Base64 PNG |
| `browser_click` | Click element | `x, y` or `selector`, `tabId?` | Click confirmation |
| `browser_type` | Type text | `text`, `tabId?` | Type confirmation |
| `browser_get_tabs` | List open tabs | - | Array of tabs |
| `browser_find` | Find element by query | `query`, `tabId?` | Matching elements |
| `browser_create_tab` | Create new tab | `url?` | New tab ID |

**Connection Details:**
- **WebSocket URL:** `ws://localhost:3005`
- **MCP Protocol:** Full Model Context Protocol implementation
- **Native Messaging:** Fallback via `com.floyd.chrome`

**Installation:**
```bash
cd "/Volumes/Storage/FLOYD Extension for Chrome/FloydChromeBuild/floydchrome"
npm install
npm run build

# Load in Chrome:
# chrome://extensions → Developer mode → Load unpacked → dist/
```

**Usage Example (Computer Use Workflow):**
```
User: "Navigate to example.com, find the search box, type 'hello', and screenshot"

Agent:
1. browser_navigate({ url: "https://example.com" })
2. browser_screenshot() → Vision model analyzes
3. browser_find({ query: "search input" })
4. browser_type({ text: "hello" })
5. browser_click({ selector: "[type='submit']" })
6. browser_screenshot() → Verify result
```

**Computer Use Capability:**
The extension enables Claude-style "Computer Use" workflows:
1. **Capture** - Screenshot for vision model
2. **Analyze** - Vision model identifies actionable elements
3. **Act** - Agent performs actions
4. **Repeat** - Loop until task complete

---

## V. CURRENT: Mobile PWA (Remote Access)

**Location:** `/Volumes/Storage/FLOYD MOBILE  PWA w: NGROK TUNNEL/`

**Status:** ⚠️ **PRESUMED ACTIVE - Last updated Feb 2025**

**Purpose:** Remote communication channel when away from desktop

**Architecture:**
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         FloydMobile PWA Architecture                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────┐          │
│  │   MOBILE     │         │   NGROK      │         │   DESKTOP    │          │
│  │  DEVICE      │─────────▶│   TUNNEL     │─────────▶│    APP       │          │
│  │              │  HTTPS   │              │  WS/MCP  │              │          │
│  │  FloydMobile │─────────▶│   (Public)  │─────────▶│ FloydDesktop │          │
│  │     PWA      │         │              │         │              │          │
│  └──────────────┘         └──────────────┘         └──────────────┘          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Capabilities:**
- **Progressive Web App** - Installable on mobile devices
- **NGROK Tunnel** - Secure remote access to Desktop API
- **Query Sending** - Send prompts from phone
- **Response Receiving** - Get agent responses remotely
- **Notifications** - Alerts on task completion

**Integration Points:**
- Connects to FloydDesktop WebSocket server (via NGROK)
- Shares session database with Desktop
- Accesses same MCP tool ecosystem

**Usage:**
```bash
# On Desktop:
cd /Volumes/Storage/FloydDesktopWeb-v2
npm run start
# Ensure NGROK tunnel is active

# On Mobile:
# 1. Open FloydMobile PWA in browser
# 2. Configure NGROK tunnel URL
# 3. Send queries remotely
```

---

## VI. LEGACY: Old/Superseded Components

### A. Old CLI Builds

| Path | Status | Reason |
|------|--------|--------|
| `/Volumes/Storage/floyd-main/` | ✗ LEGACY | Old main branch, superseded by FloydDeployable |
| `/Volumes/Storage/floyd-benchmark/` | ✗ LEGACY | Old benchmarking setup (Feb 2025) |

### B. Old Desktop Builds

| Path | Status | Reason |
|------|--------|--------|
| `/Volumes/Storage/FloydDesktopWeb-Standalone/` | ✗ LEGACY | Old standalone build (Feb 2024) |

### C. Old Agent Implementations

| Path | Status | Reason |
|------|--------|--------|
| `/Volumes/Storage/CLI AGENT BUILDS/` | ✗ LEGACY | Pre-v4 CLI builds (Nov 2025) |
| `/Volumes/Storage/Legacy Agents/` | ✗ LEGACY | Pre-v4 agent versions |
| `/Volumes/Storage/TUI-Rebuild-v2-MCP/` | ✗ LEGACY | Old TUI experiment (Feb 2025) |

---

## VII. MCP SERVER ECOSYSTEM (18 Servers - All Active)

**Location:** `/Volumes/Storage/MCP/`

**Status:** ✅ **ALL ACTIVE - Core to v4.0 functionality**

```
MCP/
├── floyd-supercache-server/     # 3-tier cache (project/reasoning/vault)
├── floyd-runner/                # Test/lint/build execution
├── floyd-git/                   # Git operations
├── floyd-explorer/              # Project mapping & file reading
├── floyd-patch/                 # Diff operations
├── floyd-devtools/              # Type analysis, git bisect
├── floyd-terminal/              # Process management
├── floyd-safe-ops/              # Impact simulation
├── lab-lead-server/             # Agent spawning & tool discovery
├── hivemind-v2/                 # Multi-agent coordination
├── context-singularity-v2/      # Context optimization
├── omega-v2/                    # Meta-cognitive reasoning
├── pattern-crystallizer-v2/     # Pattern extraction
├── novel-concepts-server/       # AI concept generation
├── gemini-tools-server/         # Dependency viz, bug freezing
├── context-singularity-v2/      # Context packing
├── legacy-bridge/               # Legacy system integration
└── [5 more specialized servers]
```

**Total Tool Count:** 105+ tools across 18 servers

---

## V. AGENT TYPES AND DEPLOYMENT

| Type | Deployed Via | Use Case | Status |
|------|--------------|----------|--------|
| **Hives** | CLI (floyd-sandbox) | Multi-task coordination | ✅ ACTIVE |
| **Swarms** | Desktop (FloydDesktopWeb-v2) | Coordinated agent groups | ✅ ACTIVE |
| **Headless** | Both | Single-task, race-condition-safe | ✅ ACTIVE |

---

## VI. EXTERNAL INTEGRATIONS

### Active Integrations

| Service | Purpose | Endpoint/Config |
|---------|---------|----------------|
| **ZAI API** | GLM-4.6/GLM-5 LLM endpoint | `https://rube.app` |
| **Qdrant** | Vector database for knowledge search | Configured in MCP |
| **DigitalOcean** | Production deployment | 159.65.221.69 |
| **GitHub** | Repository hosting | github.com/CaptainPhantasy/FloydSandyIso |

---

## VII. DATABASE STRUCTURE

### v4.0 Database Schema

**Location:** Project-specific `.floyd/floyd.db`

**Migrations Applied (10 total):**
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
[v4.0.0 additional migrations]
```

**Key Tables:**
- `sessions` - Session management with token tracking
- `messages` - Message history with tool calls
- `read_files` - File tracking
- `todos` - Task management

---

## VIII. CONFIGURATION HIERARCHY

### Priority Order (Highest to Lowest)

1. **`.floyd/floyd.json`** - Project-specific (gitignored, sandbox isolated)
2. **`~/.floyd/floyd.json`** - Global user config
3. **`floyd.json`** - Local project (can be git-tracked)

### Example v4.0 Config

```json
{
  "$schema": "./floyd-schema.json",
  "models": {
    "large": {
      "model": "glm-4.6-chat-v2",
      "provider": "z-ai",
      "temperature": 0.1,
      "max_tokens": 8192
    }
  },
  "providers": {
    "z-ai": {
      "base_url": "https://rube.app",
      "api_key": "$ZAI_API_KEY",
      "type": "openai-compat"
    }
  },
  "mcp": {
    "floyd-supercache": { "type": "stdio", "command": "node", "args": ["src/index.js"] }
  },
  "options": {
    "execution": {
      "allowed_banned_commands": ["curl", "wget"]
    }
  }
}
```

---

## IX. VERSION COMPARISON

### Floyd CLI Version History

| Version | Location | Date | Key Features |
|---------|----------|------|--------------|
| **v4.0.0** | floyd-sandbox/FloydDeployable | 2026-02-21 | Agent Library, Streaming Progress, Context Status |
| v3.7.01 | floyd-main | ~2025-02-18 | Pre-sandbox isolation |
| < v3.7 | Legacy Agents/ | Pre-2025 | Early versions |

### Desktop App Version History

| Version | Location | Date | Status |
|---------|----------|------|--------|
| **v2** | FloydDesktopWeb-v2 | 2026-02-19 | ✅ CURRENT |
| v1 | FloydDesktopWeb-Standalone | 2024-02-19 | ✗ LEGACY |

---

## X. KEY FILES FOR AGENT AWARENESS

### Files to Cache in SUPERCACHE

| File | Purpose | Cache Key |
|------|---------|-----------|
| `~/.floyd/floyd.db` | Session database | N/A (SQLite) |
| `FLOYD.md` | Agent protocol | `system:floyd_protocol` |
| This map | Ecosystem reference | `system:ecosystem_map` |
| Tool locations | Tool discovery | `system:tool_registry` (TO BE POPULATED) |

---

## XI. ACTIVE DEVELOPMENT PATHS

### Current Focus Areas (Based on V4_AUDIT_AND_ALIGNMENT_REPORT.md)

1. **GLM-5 Optimization** - Tight constraints for better performance
2. **Discovery Gates** - Pre-action verification for agents
3. **Tool Registry** - Centralized tool discovery
4. **Ecosystem Awareness** - Agent knowledge of full system

### Next Planned Features (FLOYD_EVOLUTION_ROADMAP.md)

- v4.1: Enhanced MCP health monitoring
- v4.2: SUPERCACHE namespaces
- v4.3: Parallel bash execution
- v4.5: Intelligence layer
- v4.9: Competitive features

---

## XII. COMMUNICATION FLOW

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           USER INTERACTION FLOW                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐                │
│  │   USER   │───▶│  ENTRY   │───▶│  FLOYD   │───▶│   GLM    │                │
│  │          │    │  POINT   │    │   v4.0   │    │  -4.6/5  │                │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘                │
│       │              │                 │                 │                      │
│       │              ▼                 ▼                 ▼                      │
│       │         ┌─────────┐       ┌─────────┐       ┌─────────┐              │
│       │         │   CLI   │       │ Desktop │       │  Mobile │              │
│       │         │(floyd4) │       │   v2    │       │   PWA   │              │
│       │         └─────────┘       └─────────┘       └─────────┘              │
│       │                                                                  │
│       ▼                                                                  │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐                              │
│  │ RESPONSE │◀───│   TUI    │◀───│  AGENT   │◀───── ZAI GLM Response      │
│  │ DISPLAY  │    │ RENDER   │    │ COORDIN  │                              │
│  └──────────┘    └──────────┘    └──────────┘                              │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## XIII. SUMMARY: WHAT TO USE

### For Development

```bash
# Primary CLI
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
./floyd4

# Or build fresh
go build . && ./FloydDeployable
```

### For Desktop

```bash
# Desktop app v2
cd /Volumes/Storage/FloydDesktopWeb-v2
npm install && npm run start
```

### For Browser Automation

```bash
# Chrome Extension
cd "/Volumes/Storage/FLOYD Extension for Chrome/FloydChromeBuild/floydchrome"
npm install && npm run build
# Load in Chrome: chrome://extensions → Load unpacked → dist/
```

### For Mobile Access

```bash
# Mobile PWA
# Navigate to FloydMobile PWA URL with NGROK tunnel configured
# Location: /Volumes/Storage/FLOYD MOBILE  PWA w: NGROK TUNNEL/
```

### For MCP Tools

```bash
# MCP servers are at
cd /Volumes/Storage/MCP/<server-name>
npm install && node src/index.js
```

---

## XIV. ECOSYSTEM INTEGRATION STATUS

| Component | Integration Status | Notes |
|-----------|-------------------|-------|
| **CLI → Desktop** | ✅ WebSocket MCP | Full bidirectional |
| **CLI → Chrome Extension** | ✅ Via Desktop WebSocket | Tool routing works |
| **Desktop → Mobile** | ⚠️ Via NGROK | Needs tunnel setup |
| **All Components → MCP** | ✅ Full access | 18 servers, 105+ tools |
| **SUPERCACHE Population** | ❌ NOT IMPLEMENTED | Tool registry empty |
| **Hivemind Levels 2-6** | ❌ NOT IMPLEMENTED | Only Level 1 active |

---

*This document distinguishes ACTIVE v4.0 components from LEGACY iterations. Always use FloydDeployable for CLI work and FloydDesktopWeb-v2 for desktop functionality.*
