# FLOYD Ecosystem Map

**Generated:** 2026-02-22
**Version:** 4.0.0
**Purpose:** Complete directory map of the FLOYD agent ecosystem

---

## Core Components

### 1. FLOYD CLI (Core Agent) - v4.0.0
```
/Volumes/Storage/floyd-sandbox/FloydDeployable/
├── floyd4                    # Main executable
├── src/
│   ├── agent/                # Core agent logic
│   │   ├── manager.ts        # Manager agent with ecosystem awareness
│   │   ├── workers/          # Worker swarm implementations
│   │   └── profiles.ts       # Agent profiles
│   ├── mcp/                  # MCP client integration
│   │   ├── patch-server.ts   # Floyd-Patch (5 tools)
│   │   ├── runner-server.ts  # Floyd-Runner (6 tools)
│   │   ├── cache-server.ts   # Floyd-Supercache (12 tools)
│   │   ├── git-server.ts     # Floyd-Git (7 tools)
│   │   └── explorer-server.ts # Floyd-Explorer (5 tools)
│   └── store/                # Session & state management
├── .floyd/
│   ├── floyd.db              # Session database
│   ├── mcp.json              # MCP server configuration
│   └── supercache/           # 3-tier cache storage
└── docs/
    ├── FLOYD_ECOSYSTEM_MAP.md
    └── FLOYD_ECOSYSTEM_AWARENESS.md
```

**MCP Servers (CLI):** 18 servers, 98+ tools
**Key Capability:** Full toolset, hives, swarms, batch operations

### 2. FLOYD Desktop Application - v4.0.0
```
/Volumes/Storage/FloydDesktopWeb-v2/
├── electron/                 # Electron main process
├── src/                      # React frontend
│   ├── components/           # UI components
│   ├── hooks/                # React hooks
│   └── lib/                  # Utilities
├── server/                   # Backend server (port 3001)
│   ├── mcp-client.ts         # MCP integration
│   ├── ws-mcp-server.ts      # WebSocket MCP server (port 3005)
│   ├── tool-executor.ts      # Tool execution
│   ├── skills-manager.ts     # Skills system
│   ├── projects-manager.ts   # Project management
│   └── browork-manager.ts    # Browser automation
├── .floyd-data/              # Sessions and settings
└── docs/                     # Documentation
```

**API:** http://localhost:3001
**Frontend:** http://localhost:5173
**WebSocket MCP:** ws://localhost:3005
**MCP Servers:** 12 servers, 45+ tools
**Key Capability:** Visual interface, browser automation, streaming chat

### 3. FLOYD Mobile PWA
```
/Volumes/Storage/FLOYD MOBILE  PWA w: NGROK TUNNEL/
```
- Progressive Web App for mobile communication
- NGROK tunnel: https://floyd-mobile.ngrok-free.app (port 8765)
- Connects to Desktop instance
- **Key Capability:** Remote access from mobile devices

### 4. FLOYD Chrome Extension
```
/Volumes/Storage/FLOYD Extension for Chrome/
```
- Browser integration
- Browser control capabilities
- Connects to Desktop via WebSocket MCP (port 3005)
- **Key Capability:** In-browser AI assistance

### 5. FLOYD IDE (FLOYD CURSE'M)
```
/Applications/FLOYD CURSE'M.app/
```
- VS Code extension integration
- Multi-chat interface
- Project-aware coding
- Connects to Floyd API (port 3001)
- **Key Capability:** IDE-integrated AI assistance

---

## MCP Server Ecosystem

```
/Volumes/Storage/MCP/
├── floyd-http-server/        # NEW: HTTP client (7 tools)
├── floyd-supercache-server/  # 3-tier cache (12 tools)
├── floyd-runner/             # Test/lint/build (6 tools)
├── floyd-git/                # Git operations (7 tools)
├── floyd-explorer/           # Project mapping (5 tools)
├── floyd-patch/              # Diff operations (5 tools)
├── floyd-devtools/           # Type analysis (6 tools)
├── floyd-terminal/           # Process management (9 tools)
├── floyd-safe-ops/           # Impact simulation (3 tools)
├── lab-lead-server/          # Lab coordination (6 tools)
├── hivemind-v2/              # Multi-agent coordination (11 tools)
├── context-singularity-v2/   # Context optimization (9 tools)
├── omega-v2/                 # Meta-cognitive reasoning (6 tools)
├── pattern-crystallizer-v2/  # Pattern extraction (5 tools)
├── novel-concepts-server/    # AI concept generation (10 tools)
├── gemini-tools-server/      # Dependency viz, bug freezing (3 tools)
└── (External ZAI servers)    # Vision, web search, GitHub (13 tools)
```

**Total:** 19 MCP servers, 105+ tools

---

## Inter-Instance Communication

### HTTP API Endpoints

| Instance | Local URL | Remote URL | Purpose |
|----------|-----------|------------|---------|
| Desktop API | http://localhost:3001 | https://crm-ai-pro-test.ngrok-free.app | Chat, settings, tools |
| Desktop Frontend | http://localhost:5173 | (via ngrok) | Web UI |
| Mobile PWA | (dev only) | https://floyd-mobile.ngrok-free.app | Mobile UI |
| WebSocket MCP | ws://localhost:3005 | (local only) | Chrome extension |

### SUPERCACHE Bridge

All instances share SUPERCACHE at `~/.floyd/supercache/`:
- **project tier:** Session data (1-hour TTL)
- **reasoning tier:** Decision chains (no TTL)
- **vault tier:** Permanent patterns

Communication pattern:
1. Instance A: `cache_store(key="cross:msg", value="...")`
2. Instance B: `cache_retrieve(key="cross:msg")`

---

## Knowledge Systems

```
/Volumes/Storage/Knowledge Bases/
├── 01_TECHNICAL/             # AI models, platforms, databases
├── 02_LEGAL/                 # Tax, workplace, privacy law
├── 03_ADVOCACY/              # Elder care, disaster response
├── 04_BUSINESS/              # QuickBooks, Google, Stripe
├── _SYSTEM/                  # System configs
│   ├── RAGBOT3000/           # Frontend
│   ├── personas/             # Domain personas
│   └── services/             # Query service
├── _Tooling/                 # Tools (UniversalScraper, etc.)
├── knowledge-library/        # Legacy knowledge bases
└── ingest_knowledge.py       # Vector ingestion script
```

---

## Communication Channels

| Channel | Location | Purpose |
|---------|----------|---------|
| Desktop App | FloydDesktopWeb-v2 | Primary interface, swarm deployment |
| Mobile PWA | floyd-mobile.ngrok-free.app | Remote communication |
| Chrome Extension | FLOYD Extension for Chrome | Browser control |
| CLI | floyd-sandbox/FloydDeployable | Direct terminal, hive coordination |
| IDE | FLOYD CURSE'M | VS Code integration |

---

## Agent Types Available

| Type | Deployed Via | Use Case |
|------|--------------|----------|
| Hives | CLI | Multi-task coordination |
| Swarms | Desktop | Coordinated agent groups |
| Headless | Both | Single-task, race-condition-safe |

---

## Key Files for Agent Awareness

| File | Purpose |
|------|---------|
| `~/.floyd/floyd.db` | Session database |
| `~/.floyd/supercache/` | 3-tier cache storage |
| `system:project_registry` | SUPERCACHE project list |
| `system:tool_registry` | Tool locations |
| `system:ecosystem_map` | This map in cache |
| `system:ecosystem_awareness` | Full awareness document |

---

## External Integrations

- **ZAI API** - External MCP servers (web-search-prime, zai-mcp-server, etc.)
- **Qdrant** - Vector database for knowledge search
- **DigitalOcean** - Production deployment (159.65.221.69)
- **Ngrok** - Tunnel service for remote access

---

## Version 4.0.0 Updates (2026-02-22)

### New Features:
1. **floyd-http-server** - HTTP client for inter-instance communication (7 tools)
2. **Ecosystem awareness** - All instances know about each other
3. **Tool registry** - Central tool documentation
4. **SUPERCACHE bridge** - Message passing between instances
5. **Updated system prompts** - Multi-context awareness

### Instance Capabilities Matrix:

| Capability | CLI | Desktop | Mobile | IDE |
|------------|-----|---------|--------|-----|
| Full MCP tools | 98+ | 45+ | - | - |
| HTTP client | Yes | Yes | - | - |
| SUPERCACHE | Yes | Yes | - | - |
| Browser automation | Yes | Yes | - | - |
| Visual interface | - | Yes | Yes | - |
| IDE integration | - | - | - | Yes |
| Hives/swarms | Yes | Yes | - | - |

---

*This map should be loaded into SUPERCACHE as `system:ecosystem_map`*
*Related: `FLOYD_ECOSYSTEM_AWARENESS.md` for detailed system prompts*
