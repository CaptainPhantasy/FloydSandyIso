# FLOYD ECOSYSTEM: Optimal Usage & Latent Capabilities

**Generated:** 2026-02-22
**Analysis:** Complete ecosystem review covering CLI, Desktop, Chrome Extension, Mobile, MCP servers, and coordination patterns

---

## EXECUTIVE SUMMARY

The FLOYD ecosystem is a **homegrown Claude-equivalent** with remarkable parallel capabilities:

| Aspect | FLOYD | Claude |
|--------|-------|--------|
| **CLI Agent** | floyd-sandbox/FloydDeployable | claude-code CLI |
| **Desktop App** | FloydDesktopWeb-v2 | Claude Desktop App |
| **Coworkers/Swarming** | Browork Manager + Hivemind-v2 | Claude Coworkers |
| **Browser Control** | FloydChrome Extension | Claude for Chrome |
| **Mobile** | FloydMobile PWA + NGROK | Claude Mobile App |
| **MCP Ecosystem** | 18 servers, 105+ tools | MCP community |

**The Core Insight:** FLOYD was built to mirror Claude's multi-surface architecture but with **deeper integration points** and **more aggressive automation** (swarms, hivemind, computer-use).

---

## PART I: INTENDED USE (As Designed)

### The Vision: Ubiquitous AI Across All Surfaces

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        FLOYD ECOSYSTEM - INTENDED DESIGN                      │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────┐          │
│  │   DESKTOP    │         │     CLI      │         │    MOBILE    │          │
│  │   (PRIMARY)  │◄────────►│   (POWER)    │◄────────►│  (REMOTE)    │          │
│  │              │  WS/MCP  │              │  MCP/DB  │              │          │
│  │ FloydDesktop │─────────►│ FloydDeploy  │─────────►│ FloydMobile  │          │
│  │    Web-v2    │         │   (floyd4)    │         │    PWA       │          │
│  └──────┬───────┘         └──────┬───────┘         └──────┬───────┘          │
│         │                        │                        │                   │
│         │                        ▼                        │                   │
│         │                 ┌─────────────┐                │                   │
│         │                 │   MCP LAB   │                │                   │
│         │                 │ 18 Servers  │◄───────────────┘                   │
│         │                 │ 105+ Tools  │                                    │
│         │                 └──────┬───────┘                                    │
│         │                        │                                          │
│         ▼                        ▼                                          │
│  ┌──────────────────────────────────────────────────────────────┐       │
│  │                    CHROME EXTENSION                          │       │
│  │                Browser Control Layer                          │       │
│  └──────────────────────────────────────────────────────────────┘       │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 1. FloydDesktopWeb-v2 (Primary Interface)

**Location:** `/Volumes/Storage/FloydDesktopWeb-v2/`

**Intended Role:** Central hub for agent coordination, swarm management, and visual task tracking

**Key Capabilities:**
- **WebSocket MCP Server** (port 3005) - Routes tool calls between CLI and Chrome extension
- **Browork Manager** - Sub-agent delegation (like Claude Coworkers)
- **Skills Manager** - Extensible skill system
- **Projects Manager** - Project-based organization
- **Tool Executor** - Full filesystem, terminal, code execution
- **Multi-Provider Support** - Anthropic, OpenAI, GLM-4.x/5.x

**Optimal Usage:**
```
1. Start FloydDesktop first (it runs the MCP server)
2. Connect Chrome extension to it (ws://localhost:3005)
3. Use CLI for power operations, Desktop for swarm orchestration
4. Mobile PWA connects via NGROK tunnel for remote access
```

### 2. FloydDeployable CLI (Power User Interface)

**Location:** `/Volumes/Storage/floyd-sandbox/FloydDeployable/`

**Intended Role:** Direct terminal interaction, hive coordination, development operations

**Key Capabilities:**
- **TUI (Terminal UI)** - Bubbletea-based interactive interface
- **Non-interactive mode** - Scriptable execution
- **Agent Library** - Markdown-based persona system
- **MCP Client** - Connects to all 18 MCP servers
- **Session Management** - SQLite-based conversation history
- **Tool Integration** - 20+ built-in tools + MCP tools

**Optimal Usage:**
```bash
# Direct interaction
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
./floyd4

# Non-interactive (scripting)
./floyd4 "Run tests and report failures" > output.md

# With specific agent
./floyd4 --agent code-reviewer "Review this PR"
```

### 3. FloydChrome Extension (Browser Control)

**Location:** `/Volumes/Storage/FLOYD Extension for Chrome/FloydChromeBuild/floydchrome/`

**Intended Role:** Enable Computer Use workflows - agent sees and controls browser

**Key Capabilities:**
- **Screenshot capture** - Vision model integration
- **Page reading** - Accessibility tree extraction
- **Element interaction** - Click, type, find
- **Tab management** - Create, list, switch tabs
- **MCP Server implementation** - WebSocket client to Desktop
- **Safety layer** - Permission checks, sanitization

**Tools (6):**
| Tool | Description |
|------|-------------|
| `browser_navigate` | Navigate to URL |
| `browser_read_page` | Get accessibility tree |
| `browser_screenshot` | Capture screenshot |
| `browser_click` | Click element |
| `browser_type` | Type text |
| `browser_get_tabs` | List open tabs |

**Optimal Usage:**
```
1. Load extension in Chrome (chrome://extensions → Load unpacked → dist/)
2. Start FloydDesktop (WebSocket server on port 3005)
3. Extension auto-connects
4. CLI/Desktop can now: "Navigate to X, take screenshot, find button Y, click it"
```

### 4. FloydMobile PWA (Remote Access)

**Location:** `/Volumes/Storage/FLOYD MOBILE  PWA w: NGROK TUNNEL/`

**Intended Role:** Remote communication channel when away from desktop

**Key Capabilities:**
- **Progressive Web App** - Installable on mobile
- **NGROK Tunnel** - Remote access to Desktop API
- **Agent communication** - Send queries, receive responses

**Optimal Usage:**
```
1. Start FloydDesktop with NGROK tunnel active
2. Open FloydMobile PWA on phone
3. Configure tunnel URL
4. Send queries remotely, receive notifications
```

---

## PART II: OPTIMAL USAGE (How It SHOULD Be Used)

### The Golden Path: Full Ecosystem Engagement

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         OPTIMAL FLOYD WORKFLOW                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  1. BOOT SEQUENCE                                                              │
│     ┌─────────────────────────────────────────────────────────────────────┐    │
│     │ Start FloydDesktopWeb-v2 → MCP servers auto-initialize               │    │
│     │ Load Chrome Extension → Connects to Desktop (ws://localhost:3005)   │    │
│     │ Open Mobile PWA → Connects via NGROK                                │    │
│     └─────────────────────────────────────────────────────────────────────┘    │
│                                                                                 │
│  2. DAILY WORKFLOW                                                             │
│     ┌─────────────────────────────────────────────────────────────────────┐    │
│     │ DEVELOPMENT WORK:                                                     │    │
│     │ • Use CLI (floyd4) for code tasks, testing, git ops                 │    │
│     │ • Use Desktop for swarm deployment on multi-file projects            │    │
│     │ • Chrome Extension enables web testing, scraping                    │    │
│     │                                                                        │    │
│     │ RESEARCH WORK:                                                        │    │
│     │ • Use Desktop Browork Manager to spawn researcher agents            │    │
│     │ • Leverage web-search-prime, web-reader, zread MCP servers          │    │
│     │                                                                        │    │
│     │ COORDINATION WORK:                                                    │    │
│     │ • Use Hivemind-v2 for multi-agent parallel development              │    │
│     │ • Use lab-lead to discover optimal tools for any task               │    │
│     │                                                                        │    │
│     │ REMOTE WORK:                                                           │    │
│     │ • Use Mobile PWA to send queries from phone                          │    │
│     │ • Receive notifications on completion                                │    │
│     └─────────────────────────────────────────────────────────────────────┘    │
│                                                                                 │
│  3. AGENT TYPE SELECTION                                                        │
│     ┌─────────────────────────────────────────────────────────────────────┐    │
│     │ Task Type              │ Use Interface  │ Agent Type                 │    │
│     ├─────────────────────────────────────────────────────────────────────┤    │
│     │ Quick question        │ CLI            │ Default coder             │    │
│     │ Code review           │ CLI            │ Agent Library persona     │    │
│     │ Multi-file refactor   │ Desktop Swarm │ Browork Manager            │    │
│     │ Parallel development  │ Hivemind-v2    │ Coordinated agents         │    │
│     │ Web automation        │ CLI + Chrome  │ Computer Use workflow      │    │
│     │ Research              │ Desktop        │ Researcher agent cluster   │    │
│     │ Remote query          │ Mobile PWA     │ Message to Desktop         │    │
│     └─────────────────────────────────────────────────────────────────────┘    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Example: Full Stack Development Task

**User Request:** "Build a complete authentication system"

**Optimal FLOYD Response:**

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ STEP 1: Task Decomposition (Desktop → Hivemind-v2)                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│ User in FloydDesktop: "Build complete auth system"                            │
│                                                                                 │
│ FloydDesktop → Hivemind-v2 → distributed_task_board.create_tasks([             │
│   { id: 'auth_1', description: 'JWT generation', domain: 'backend' },         │
│   { id: 'auth_2', description: 'Login UI', domain: 'frontend' },             │
│   { id: 'auth_3', description: 'Tests', domain: 'testing' },                 │
│   { id: 'auth_4', description: 'Email service', domain: 'backend' }           │
│ ])                                                                            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│ STEP 2: Intelligent Routing (Level 2 - Specified)                             │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│ Hivemind-v2 → lab_spawn_agent({ type: 'coder' }) × 3                         │
│                                                                                 │
│ Agent Alpha (backend specialist) → auth_1, auth_4                             │
│ Agent Beta (frontend specialist) → auth_2                                     │
│ Agent Gamma (testing specialist) → auth_3                                     │
│                                                                                 │
│ Each agent gets:                                                               │
│ - MCP tools for their domain (runner, git, patch, etc.)                      │
│ - SUPERCACHE access for coordination                                          │
│ - File locking to prevent conflicts                                            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│ STEP 3: Parallel Execution with Computer Use                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│ Agent Alpha needs to test email signup →                                      │
│   Floyd CLI → Chrome Extension → browser_navigate('localhost:8080/signup')   │
│   → browser_screenshot → vision model analysis → browser_type(...)            │
│   → browser_click('#submit') → browser_screenshot → verification             │
│                                                                                 │
│ Simultaneously:                                                               │
│ Agent Beta building React login UI                                            │
│ Agent Gamma writing integration tests                                         │
│                                                                                 │
│ SUPERCACHE maintains:                                                         │
│ - project:task_locks (file ownership)                                         │
│ - project:progress (task completion)                                          │
│ - reasoning:chain (shared reasoning)                                          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────────┐
│ STEP 4: Knowledge Capture (Level 4 - Specified)                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│ On completion → pattern_crystallizer → crystallize_success_pattern({          │
│   task: 'JWT auth implementation',                                             │
│   solution: code_structure,                                                   │
│   tags: ['backend', 'security', 'jwt']                                       │
│ })                                                                            │
│                                                                                 │
│ Pattern stored in SUPERCACHE vault tier                                      │
│ Future agents retrieve via cache_search({ tier: 'vault', query: 'auth' })    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## PART III: LATENT/UNDERUTILIZED CAPABILITIES

### BAKED-IN BUT UNUSED FEATURES

#### 1. LAB-LEAD: Central Tool Discovery (Severely Underutilized)

**Location:** `/Volumes/Storage/MCP/lab-lead-server/`

**What It Does:**
- `lab_inventory()` - Complete inventory of all 18 servers + 105 tools
- `lab_find_tool({ task: "..." })` - AI-powered tool discovery
- `lab_spawn_agent({ agent_type: "..." })` - Generate specialist agent configs
- `lab_get_tool_registry()` - Get MCP config for all servers

**Current Usage:** Probably rarely called directly
**Optimal Usage:** Should be called at boot to populate tool registry

**Latent Capability:**
```javascript
// DESIRED: Auto-populate tool registry on boot
lab_get_tool_registry({ format: 'mcp_config' })
// Returns complete MCP config for Claude/Floyd config.json
// Should be stored in SUPERCACHE as system:tool_registry
```

#### 2. HIVEMIND-v2: 6-Level Evolution (Level 1 Only Used)

**Location:** `/Volumes/Storage/MCP/hivemind-v2/`

**What Exists:**
- ✅ Level 1: Basic coordination (file locking, task board)
- 📋 Level 2: Intelligent routing (specialist agents)
- 📋 Level 3: Dynamic scaling (auto-spawn agents)
- 📋 Level 4: Cross-agent learning (pattern sharing)
- 📋 Level 5: Meta-optimization (self-improving algorithms)
- 📋 Level 6: Permanent evolution (SEAL weight updates)

**Latent Capability:**
The Hivemind could:
- Auto-scale agent count based on workload
- Route tasks to specialist agents (frontend vs backend)
- Learn from successes and share patterns
- Optimize its own coordination algorithms
- Permanently improve via weight updates

**Why Underutilized:** Requires implementation work beyond Level 1

#### 3. SUPERCACHE: 3-Tier System (Probably Underutilized)

**Location:** `/Volumes/Storage/MCP/floyd-supercache-server/`

**Tiers:**
- **Project Tier:** Session-specific data
- **Reasoning Tier:** Chain-of-thought, temporary working memory
- **Vault Tier:** Permanent knowledge, success patterns, learnings

**Latent Capability:**
```javascript
// Most users probably only use basic cache_store/retrieve
// But they could be using:

// 1. EPISODIC MEMORY BANK
episodic_memory_bank({
  action: 'store',
  episode: {
    trigger: 'User asked for auth system',
    reasoning: 'Decomposed into 4 subtasks',
    solution: 'Used Hivemind to parallelize',
    outcome: 'success'
  }
})

// 2. PATTERN CRYSTALLIZER
crystallize_pattern({
  name: 'parallel_auth_development',
  pattern: solution_structure,
  tags: ['auth', 'parallel', 'hivemind']
})

// 3. CONTEXT SINGULARITY
compress_context({
  method: 'semantic',
  target_ratio: 0.5,
  preserve: ['auth', 'security', 'jwt']
})
```

#### 4. BROWORK MANAGER: Coworker Alternative (Barely Used)

**Location:** `FloydDesktopWeb-v2/server/browork-manager.ts`

**What It Does:**
- Spawns autonomous agents to work on tasks in parallel
- Supports Anthropic, OpenAI, GLM providers
- Task status tracking, logs, progress callbacks

**Current Usage:** Probably only manual spawning via Desktop UI
**Optimal Usage:** Auto-spawn for any multi-part task

**Latent Capability:**
```typescript
// Should be integrated with Hivemind for automatic spawning
// When task decomposed → Browork spawns workers automatically
```

#### 5. AGENT LIBRARY SYSTEM (New in v4.0, Underutilized)

**Location:** `floyd-sandbox/FloydDeployable/internal/agents/`

**What It Does:**
- Markdown-based persona definitions
- Ctrl+P → Agent Library dialog
- Dynamic system prompt switching

**Current Usage:** Only manual selection via TUI
**Optimal Usage:** Auto-select agent based on task type

**Latent Capability:**
```go
// DESIRED: Auto-agent selection
// User: "Review this PR" → Auto-loads code-reviewer.md
// User: "Audit release" → Auto-loads release-auditor.md
// User: "Write docs" → Auto-loads technical-writer.md (create it!)
```

#### 6. CHROME EXTENSION: Computer Use (Not Fully Realized)

**Location:** FloydChrome Extension

**What Exists:**
- Screenshot capture
- Element finding
- Click/type interaction
- Vision model integration ready

**Latent Capability:**
The extension enables full Computer Use workflows:
```
Screenshot → Vision model analyzes → Identifies actionable elements →
Agent decides action → Click/type → Screenshot → Loop
```

This is Claude's "Computer Use" paradigm and it's BUILT into FLOYD but not actively used!

#### 7. OMEGA-v2: Meta-Cognitive Reasoning (Probably Unused)

**Location:** `/Volumes/Storage/MCP/omega-v2/`

**What It Does:**
- High-level reasoning and strategy
- Consensus protocol (multiple AI viewpoints)
- Meta-cognitive analysis

**Latent Capability:**
```javascript
// For complex decisions, run consensus protocol
omega_consensus({
  question: 'Should we use Redux or Context for state?',
  agents: ['optimistic', 'pessimistic', 'pragmatic', 'security'],
  timeout: 60000
})
// Returns 4 perspectives + consensus recommendation
```

#### 8. SAFE-OPS: Impact Simulation (Before Running)

**Location:** `/Volumes/Storage/MCP/floyd-safe-ops-server/`

**What It Does:**
- `impact_simulate({ changes: [...] })` - Simulate before applying
- `safe_refactor` - Refactor with verification
- `verify` - Post-change verification

**Latent Capability:**
```javascript
// BEFORE running changes, simulate:
impact_simulate({
  changes: [
    { file: 'auth.ts', change: 'add JWT validation' },
    { file: 'api.ts', change: 'add auth middleware' }
  ]
})
// Returns: affected files, risk level, test impact
```

---

## PART IV: RECOMMENDED WORKFLOW CHANGES

### What You Should Be Doing Differently

#### 1. BOOT ROUTINE CHANGE

**Current:** Start CLI directly
```bash
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
./floyd4
```

**Optimal:**
```bash
# 1. Start Desktop first (runs MCP servers)
cd /Volumes/Storage/FloydDesktopWeb-v2
npm run dev &

# 2. Start CLI with full MCP awareness
cd /Volumes/Storage/floyd-sandbox/FloydDeployable
./floyd4

# 3. Check tool registry automatically populated
# Boot summary should show "Tools available: 105+ from 18 servers"
```

#### 2. TASK ROUTING CHANGE

**Current:** Single agent does everything

**Optimal:** Use lab-lead to find optimal tools
```
Your query → lab_find_tool({ task: "I need to analyze dependencies" })
→ Returns: floyd-devtools (dependency_graph), gemini-tools (dep_viz)
→ Agent uses optimal tools instead of guessing
```

#### 3. MULTI-AGENT CHANGE

**Current:** Sequential processing

**Optimal:** Use Hivemind for parallel work
```
"Implement auth system with tests"
→ Hivemind decomposes into 4 tasks
→ Spawns 3 specialist agents
→ All work in parallel with file locking
→ Complete in 1/3 the time
```

#### 4. KNOWGE CAPTURE CHANGE

**Current:** Each session starts fresh

**Optimal:** Use SUPERCACHE vault tier
```javascript
// After successful task
episodic_memory_bank.store({
  trigger: user_request,
  solution: implementation,
  outcome: 'success',
  tags: ['auth', 'jwt', 'parallel']
})
// Future sessions can retrieve this knowledge
```

---

## PART V: THE IDE YOU MENTIONED

I don't see a dedicated IDE in the explored directories. Based on the ecosystem, the IDE capabilities might be:

1. **FloydDesktopWeb-v2** - Acts as IDE-like interface with:
   - Project management
   - File operations
   - Tool execution
   - Swarm coordination

2. **FloydChrome Extension** - Browser-based IDE for web development

3. **Or a separate project** in another location

If you have a dedicated IDE project, please point me to it and I'll include it in the ecosystem map.

---

## PART VI: SUMMARY TABLE

| Component | Status | Optimal Usage | Latent Capability |
|-----------|--------|---------------|-------------------|
| **FloydDeployable CLI** | ✅ ACTIVE | Terminal power user | Agent Library auto-selection |
| **FloydDesktopWeb-v2** | ✅ ACTIVE | Swarm orchestration hub | Auto-spawn via Browork |
| **FloydChrome Extension** | ✅ ACTIVE | Browser automation | Full Computer Use workflows |
| **FloydMobile PWA** | ⚠️ UNKNOWN | Remote queries | Bi-directional sync |
| **lab-lead-server** | ✅ ACTIVE | Tool discovery | Auto-populate tool registry |
| **hivemind-v2** | ✅ LEVEL 1 ONLY | Multi-agent coord | Levels 2-6 unimplemented |
| **SUPERCACHE** | ✅ ACTIVE | Basic caching | 3-tier system underutilized |
| **Browork Manager** | ✅ ACTIVE | Manual spawn | Auto-spawn integration |
| **omega-v2** | ✅ ACTIVE | Meta-reasoning | Consensus protocol |
| **safe-ops-server** | ✅ ACTIVE | Impact simulation | Pre-commit checks |
| **pattern-crystallizer-v2** | ✅ ACTIVE | Pattern extraction | Auto-learning loops |

---

## CONCLUSION

**FLOYD is more powerful than its current usage suggests.** The ecosystem has:

1. **Full Claude-equivalent architecture** - Desktop, CLI, Browser, Mobile
2. **More aggressive automation** - Hivemind swarming, computer use
3. **Deeper tool integration** - 18 MCP servers vs Claude's handful
4. **Latent meta-cognition** - Omega-v2 consensus, pattern learning
5. **Persistent memory** - SUPERCACHE vault tier for永久 learning

**The key bottleneck:** Implementation gaps (Hivemind Levels 2-6) and lack of integration between components (auto-spawning, auto-tool-selection).

**Highest ROI improvements:**
1. Populate tool registry via lab-lead at boot
2. Implement Hivemind Level 2 (intelligent routing)
3. Auto-select Agent Library personas based on task
4. Integrate SUPERCACHE vault tier for learning
5. Full Computer Use workflows via Chrome extension

---

*End of Analysis*
