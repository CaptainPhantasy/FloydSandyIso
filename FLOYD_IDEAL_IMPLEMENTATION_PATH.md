# FLOYD ECOSYSTEM: Ideal Implementation Path

**Generated:** 2026-02-22
**Status:** Consolidated Plan from All Roadmaps
**Current Version:** v3.7.01 → Target v4.0

---

## EXECUTIVE SUMMARY

After analyzing **7 planning documents** and the current ecosystem state, this document provides the **ideal implementation path** that consolidates all plans into a coherent, executable roadmap.

### Documents Analyzed
1. `FULL_UTILIZATION_ROADMAP.md` - Latent feature utilization
2. `safeops-complete-summary.md` - Phase 1 complete, Phase 2 60% done
3. `phase2-progress-report.md` - SafeOps enhancements
4. `refactoring-templates.md` - 15 refactoring templates defined
5. `phase2-context-singularity-integration.md` - Semantic impact analysis
6. `FLOYD_EVOLUTION_ROADMAP.md` - v4.0 progression plan
7. `V4_TECHNICAL_SPECS.md` - Concrete implementation specs
8. `PUSH_VERIFICATION_HOOK.md` - Git safety mechanism

### Current Ecosystem State

| Component | Status | Notes |
|-----------|--------|-------|
| **floyd-main** | v3.7.01 | Core CLI, operational |
| **FloydDesktopWeb-v2** | Active | 12 MCP servers connected, browork-manager exists |
| **floyd-harness** | Active | GLM-5 + DeepSeek proxy |
| **floyd-safe-ops** | Phase 2 (60%) | Config V2, Context-Singularity integration designed |
| **MCP Servers** | 12 operational | All built, need health monitoring |
| **SUPERCACHE** | Operational | Namespace support designed but not implemented |
| **Chrome Extension** | Needs Build | Exists but not packaged |
| **Mobile PWA** | Active | ngrok tunnel required |

---

## THE IDEAL PATH: CONVERGED ROADMAP

The ideal path combines the **stability-first approach** from FLOYD_EVOLUTION_ROADMAP with the **latent feature utilization** from FULL_UTILIZATION_ROADMAP and the **technical specs** from V4_TECHNICAL_SPECS.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         CONVERGED IMPLEMENTATION PATH                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  STAGE 0: FOUNDATION (Week 1) - BOOT OPTIMIZATION                       │   │
│  │  ─────────────────────────────────────────────────────────────────────   │   │
│  │  • Tool registry at boot (lab-lead integration)                         │   │
│  │  • Environment state caching                                            │   │
│  │  • Version changelog system                                             │   │
│  │  • Push verification hook deployment                                    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  STAGE 1: RELIABILITY (Weeks 2-3) - v3.8 TARGET                         │   │
│  │  ─────────────────────────────────────────────────────────────────────   │   │
│  │  • MCP health monitoring + auto-restart                                 │   │
│  │  • SUPERCACHE namespace support                                         │   │
│  │  • SafeOps Phase 2 completion (dashboard + benchmarking)               │   │
│  │  • Context-Singularity integration (semantic impact)                    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  STAGE 2: PERFORMANCE (Weeks 4-5) - v3.9 TARGET                         │   │
│  │  ─────────────────────────────────────────────────────────────────────   │   │
│  │  • Parallel bash execution                                              │   │
│  │  • Streaming tool progress                                             │   │
│  │  • Browork ↔ Hivemind integration                                      │   │
│  │  • Multi-agent headless workspace                                      │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  STAGE 3: INTELLIGENCE (Weeks 6-8) - v3.95 TARGET                        │   │
│  │  ─────────────────────────────────────────────────────────────────────   │   │
│  │  • Codebase symbol index (tree-sitter)                                  │   │
│  │  • Smart context compression                                            │   │
│  │  • Agent auto-selection (task classification)                          │   │
│  │  • Knowledge capture (pattern crystallization + episodic memory)        │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  STAGE 4: COMPETITIVE EDGE (Weeks 9-12) - v4.0 TARGET                    │   │
│  │  ─────────────────────────────────────────────────────────────────────   │   │
│  │  • Multi-modal input pipeline (vision integration)                      │   │
│  │  • Chrome extension Computer Use workflows                             │   │
│  │  • Agentic workflow engine (checkpoint + recovery)                     │   │
│  │  • Session branching                                                   │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                    │                                          │
│                                    ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │  STAGE 5: POLISH (Weeks 13-14) - STABLE RELEASE                          │   │
│  │  ─────────────────────────────────────────────────────────────────────   │   │
│  │  • Documentation complete                                              │   │
│  │  • All tests passing                                                    │   │
│  │  • Performance benchmarks met                                           │   │
│  │  • Security audit complete                                              │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## STAGE 0: FOUNDATION (Week 1) - BOOT OPTIMIZATION

**Priority:** CRITICAL - Blocks all other optimizations
**Status:** NOT STARTED
**Owner:** TBD

### Overview
The agent currently has no mechanism to discover available tools or environment state at startup. This causes "action before discovery" failures.

### Tasks

#### 0.1 Tool Registry at Boot
**File:** `internal/agent/coordinator.go` (boot sequence)

```go
// Add to boot sequence
func (c *Coordinator) bootToolRegistry(ctx context.Context) error {
    // Call lab-lead to get tool registry
    registry, err := c.mcpClient.CallTool(ctx, "lab-lead-server", "lab_get_tool_registry", map[string]any{})
    if err != nil {
        return err
    }

    // Store in SUPERCACHE
    return c.cache.Store(ctx, "system:tool_registry", registry, "vault")
}
```

**Deliverables:**
- Tool registry populated at startup
- Boot summary shows "Tools available: 105+ from 18 servers"

#### 0.2 Environment State Caching
**File:** `internal/cmd/root.go` (init)

```go
var environmentState = map[string]any{
    "global_tool_paths": []string{"~/.local/bin", "/usr/local/bin"},
    "sandbox_path": "/Volumes/Storage/floyd-sandbox/FloydDeployable",
    "mcp_servers_path": "/Volumes/Storage/MCP",
    "desktop_path": "/Volumes/Storage/FloydDesktopWeb-v2",
    "chrome_extension_path": "/Volumes/Storage/FLOYD Extension for Chrome",
    "mobile_path": "/Volumes/Storage/FLOYD MOBILE  PWA w: NGROK TUNNEL/",
}
```

**Deliverables:**
- Environment map cached as `system:environment_state`
- Agent can locate components without guidance

#### 0.3 Version Changelog System
**File:** Create `internal/version/changelog.go`

```go
type Changelog struct {
    CurrentVersion string
    NewFeatures    []string
    Deprecated     []string
    BreakingChanges []string
}

func LoadChangelog() (*Changelog, error) {
    // Load from SUPERCACHE or file
}
```

**Deliverables:**
- Version changelog stored as `system:version_changelog`
- Agent discovers new features automatically
- Changelog reviewed during boot sequence

#### 0.4 Push Verification Hook
**File:** `.git/hooks/pre-push` (in all repos)

**Deliverables:**
- Hook installed in floyd-main
- Hook installed in FloydDesktopWeb-v2
- Hook installed in MCP
- Hook installed in FloydDeployable

**Verification:**
```bash
# Test hook
git push origin test --dry-run
# Should prompt for verification code
```

### Success Criteria
- [ ] Agent displays tool count at startup
- [ ] Agent knows all component paths
- [ ] Version changelog checked on boot
- [ ] All repos have push verification hook

**Estimated Time:** 6-8 hours

---

## STAGE 1: RELIABILITY (Weeks 2-3) - v3.8 TARGET

**Priority:** HIGH - Must be rock-solid before adding features
**Status:** PARTIAL (MCP servers exist, need health monitoring)
**Owner:** TBD

### Overview
Establish reliability foundation before adding intelligence or performance features.

### Tasks

#### 1.1 MCP Health Monitoring + Auto-Restart
**Files:**
- `internal/mcp/manager.go` (NEW)
- `internal/mcp/health.go` (NEW)

```go
type Server struct {
    Name         string
    Status       ServerStatus // healthy, degraded, failed
    LastPing     time.Time
    RestartCount int
    Backoff      time.Duration // exponential
}

func (m *Manager) healthCheck(serverName string) {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        if err := m.ping(serverName); err != nil {
            m.attemptRestart(serverName)
        }
    }
}
```

**Deliverables:**
- Health check goroutine (30s interval)
- Auto-restart with exponential backoff (1s, 2s, 4s, 8s, max 60s)
- Max 5 restart attempts before marking failed
- Status reporting in UI

#### 1.2 SUPERCACHE Namespace Support
**File:** `MCP/floyd-supercache-server/src/index.ts`

```typescript
interface CacheStoreInput {
  key: string;
  value: any;
  namespace?: string;  // NEW
  tier?: 'project' | 'reasoning' | 'vault';
  tags?: string[];
}

function namespacedKey(namespace: string, key: string): string {
  if (!namespace || namespace === 'global') return key;
  return `${namespace}:${key}`;
}
```

**Deliverables:**
- Namespace parameter added to cache_store
- Namespace parameter added to cache_retrieve
- Default namespace = "global" (backwards compatible)
- Project isolation via namespaces

#### 1.3 Complete SafeOps Phase 2
**Files:**
- `docs/phase2-progress-report.md` (update)
- Impact dashboard (NEW)
- Performance benchmarking (NEW)

**Remaining Tasks:**
- [ ] PH2-2: Create impact visualization dashboard
- [ ] PH2-4: Add performance benchmarking

**Impact Dashboard Features:**
- Impact History Chart (risk scores over time)
- Risk Distribution (pie chart)
- Top Affected Files (heat map)
- Dependency Graph Visualization
- Metrics Dashboard (KPIs)

**Performance Benchmarking:**
- Tool execution times
- Operation performance
- System metrics (memory, CPU)
- Daily reports and trend charts

#### 1.4 Context-Singularity Integration
**File:** `MCP/floyd-safe-ops-server/src/index.ts` (enhance)

```typescript
async function calculateSemanticImpact(
  changedFiles: string[],
  useContextSingularity: boolean
): Promise<SemanticImpact> {
  if (!useContextSingularity) {
    return calculateSimpleImpact(changedFiles);
  }

  const traceResult = await callContextSingularity({
    tool: 'trace_origin',
    args: { filePath: file, options: { includeTransitive: true } }
  });

  return {
    riskScore: calculateRiskScore(traceResult),
    directDependents: traceResult.directDependents,
    transitiveDependents: traceResult.transitiveDependents,
    publicApiChanges: traceResult.hasPublicApiChanges,
    testingRecommendations: suggestTests(traceResult)
  };
}
```

**Deliverables:**
- Semantic impact analysis integrated
- Risk scoring (0-100) based on dependents, usage, API changes
- Smart testing recommendations
- Suggested reviewers based on domain

### Success Criteria
- [ ] MCP servers auto-recover from crashes
- [ ] SUPERCACHE namespaces prevent key collisions
- [ ] SafeOps Phase 2 100% complete
- [ ] Semantic impact analysis operational

**Estimated Time:** 12-16 hours

---

## STAGE 2: PERFORMANCE (Weeks 4-5) - v3.9 TARGET

**Priority:** HIGH - User experience foundation
**Status:** NOT STARTED
**Owner:** TBD

### Overview
Add performance improvements for faster operations and better UX.

### Tasks

#### 2.1 Parallel Bash Execution
**Files:**
- `internal/agent/tools/bash.go` (modify)
- `internal/execution/parallel.go` (NEW)

```go
type Job struct {
    ID       string
    Command  string
    Depends  []string // Job IDs that must complete first
}

type ParallelExecutor struct {
    maxConcurrency int
    jobs           map[string]*Job
}

func (e *ParallelExecutor) Execute(jobs []Job) error {
    // Dependency analysis
    // Independent jobs → parallel
    // Dependent jobs → sequential
    // Max 4 concurrent
}
```

**Safety Rules:**
- Max 4 concurrent jobs
- Read-only ops can parallelize
- Write ops require dependency declaration
- 30s timeout per job

**Expected Result:** 87% speedup on multi-command tasks

#### 2.2 Streaming Tool Progress
**Files:**
- `internal/ui/model/messages.go` (modify)
- `internal/agent/tools/*.go` (add callbacks)

```go
type ProgressCallback func(progress string)

type Tool interface {
    Execute(ctx context.Context, args any, onProgress ProgressCallback) (*Result, error)
}

// During execution
onProgress("Installing dependencies...")
onProgress("added 847 packages in 12s")
onProgress("Done")
```

**Deliverables:**
- Real-time stdout/stderr streaming
- Progress indicators for long ops
- Truncated display in UI (last N lines)

#### 2.3 Browork ↔ Hivemind Integration
**Files:**
- `FloydDesktopWeb-v2/server/browork-manager.ts` (modify)
- `MCP/hivemind-v2/src/layer1-task-integration.ts` (create)

```typescript
// Browork spawns workers via Hivemind distributed task board
class BroworkHivemindBridge {
    async spawnWorker(task: Task): Promise<Worker> {
        const taskId = await hivemind.submit_task({
            type: 'distributed_task_board',
            task: task
        });
        return new Worker(taskId, task);
    }
}
```

**Deliverables:**
- Browork uses Hivemind for task distribution
- File-level conflict prevention via SUPERCACHE locks
- Desktop can spawn 3+ parallel workers
- CLI can spawn hives for multi-task coordination

#### 2.4 Multi-Agent Headless Workspace
**Files:**
- `internal/workspace/workspace.go` (NEW)
- `internal/workspace/headless.go` (NEW)
- `internal/workspace/chalkboard.go` (NEW)

```go
type Workspace interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    SpawnWorker(ctx context.Context, spec WorkerSpec) (*Worker, error)
    GetChalkboard() Chalkboard
}

type HeadlessWorkspace struct {
    id         string
    workers    map[string]*Worker
    chalkboard *Chalkboard
}
```

**Deliverables:**
- Workspace abstraction for multi-agent execution
- Chalkboard for shared state (SUPERCACHE-backed)
- Headless workspace (goroutine-based workers)
- Worker lifecycle management

### Success Criteria
- [ ] Multi-command tasks 50%+ faster
- [ ] Long operations show real-time progress
- [ ] Browork + Hivemind coordinated
- [ ] Headless workspace operational

**Estimated Time:** 16-20 hours

---

## STAGE 3: INTELLIGENCE (Weeks 6-8) - v3.95 TARGET

**Priority:** HIGH - Build on stable, fast base
**Status:** NOT STARTED
**Owner:** TBD

### Overview
Add intelligence features for smarter agent behavior.

### Tasks

#### 3.1 Codebase Symbol Index
**Files:**
- `internal/intelligence/symbols.go` (NEW)
- `internal/intelligence/parser_*.go` (language-specific)

```go
type SymbolIndex struct {
    symbols []Symbol
    index   map[string][]Symbol // name -> symbols
}

type Symbol struct {
    Name      string
    Kind      SymbolKind // func, var, const, type, etc.
    File      string
    Line      int
    Signature string
    Docstring string
}

func (si *SymbolIndex) FuzzySearch(query string) []Symbol {
    // Fast, deterministic search
}
```

**Supported Languages (Phase 1):**
- Go (tree-sitter-go)
- TypeScript (tree-sitter-typescript)
- Python (tree-sitter-python)

**Why:** Semantic search without embeddings. Fast, deterministic.

#### 3.2 Smart Context Compression
**Files:**
- `internal/agent/summarizer.go` (modify)
- `internal/agent/context.go` (modify)

```go
type CompressionTier int

const (
    TierPreserve CompressionTier = iota // Never summarize
    TierCompress                        // Compress intelligently
    TierDiscard                         // Drop entirely
)

type Message struct {
    Content string
    Tier    CompressionTier
}
```

**Tier Rules:**
- **Preserve:** System prompt, project config, user requirements
- **Compress:** Tool results → structured summary, exploration → "Searched X, found Y"
- **Discard:** Duplicates, failed branches, verbose output

**Expected Result:** 2x effective context window

#### 3.3 Agent Auto-Selection
**Files:**
- `internal/agent/coordinator.go` (add classifier)
- `internal/agents/*.md` (persona definitions)

```go
func classifyTask(prompt string) AgentType {
    keywords := map[AgentType][]string{
        CodeReviewer:    {"review", "audit", "check"},
        ReleaseAuditor:  {"release", "deploy", "publish"},
        TestingSpecialist: {"test", "spec"},
        TechnicalWriter: {"docs", "document"},
        Coder:           {"default"},
    }

    // Return best match
}
```

**New Personas to Create:**
- `testing-specialist.md` - TDD, test coverage, CI/CD
- `technical-writer.md` - Documentation, README, guides
- `security-auditor.md` - Security review, vulnerability scan
- `performance-optimizer.md` - Profiling, optimization
- `database-specialist.md` - Schema, migrations, queries
- `frontend-specialist.md` - React, Vue, CSS, UI/UX
- `backend-specialist.md` - APIs, services, integration

#### 3.4 Knowledge Capture
**Files:**
- `internal/learning/patterns.go` (NEW)
- `internal/learning/episodic.go` (NEW)

```go
// After successful task completion
func capturePattern(task Task, solution Solution) error {
    pattern := Pattern{
        Name:     task.Type + "_" + solution.Type,
        Pattern:  solution.Structure,
        Tags:     []string{task.Domain, task.Language},
        Success:  true,
    }

    return crystallizer.Store(pattern)
}

func storeEpisode(episode Episode) error {
    return episodicMemoryBank.Store(episode)
}
```

**Deliverables:**
- Pattern crystallization after successful tasks
- Episodic memory storage (reasoning chains)
- Knowledge retrieval on task start
- Agent becomes more capable over time

### Success Criteria
- [ ] Symbol index enables fast code search
- [ ] Context compression doubles effective window
- [ ] Agent auto-selects optimal persona
- [ ] Past solutions retrievable for new tasks

**Estimated Time:** 20-24 hours

---

## STAGE 4: COMPETITIVE EDGE (Weeks 9-12) - v4.0 TARGET

**Priority:** MEDIUM - Final polish and differentiation
**Status:** NOT STARTED
**Owner:** TBD

### Overview
Add features that differentiate Floyd from competitors.

### Tasks

#### 4.1 Multi-Modal Input Pipeline
**Files:**
- `internal/agent/tools/vision.go` (NEW)
- `internal/agent/message.go` (modify)

```go
type ImageInput struct {
    Source   ImageSource // paste, drag-drop, screenshot, url
    Data     []byte
    Metadata ImageMetadata
}

type VisionTool struct {
    provider VisionProvider // ZAI MCP, local, Claude native
}

func (v *VisionTool) Analyze(ctx context.Context, img ImageInput) (string, error) {
    // Call vision model
    // Return description
}
```

**Use Cases:**
- Screenshot analysis ("What's wrong with this UI?")
- Diagram interpretation
- Error message screenshots
- Design mockup to code

#### 4.2 Chrome Extension Computer Use
**Files:**
- `internal/agent/tools/browser_automation.go` (NEW)
- Build and package extension

```go
type BrowserAutomation struct {
    extension ChromeExtension
    vision    VisionTool
}

func (b *BrowserAutomation) ComputerUse(ctx context.Context, task string) error {
    for {
        screenshot := b.extension.Screenshot()
        analysis := b.vision.Analyze(screenshot)
        action := b.decideAction(analysis)
        b.extension.Execute(action)

        if task.Complete() { break }
    }
}
```

**Workflow:**
1. Agent calls browser_screenshot()
2. Screenshot sent to vision model
3. Vision model identifies actionable elements
4. Agent decides action (click, type, navigate)
5. Action executed via browser_* tools
6. Loop until task complete

#### 4.3 Agentic Workflow Engine
**Files:**
- `internal/workflow/engine.go` (NEW)
- `internal/workflow/steps.go` (NEW)

```go
type Workflow struct {
    ID          string
    Description string
    Phases      []Phase
    Checkpoints map[string]Checkpoint
}

type Phase struct {
    Name         string
    Steps        []Step
    ApprovalGate bool // Requires user approval
}

func (e *Engine) Execute(ctx context.Context, wf *Workflow) error {
    for i, phase := range wf.Phases {
        if phase.ApprovalGate {
            if !e.waitForApproval(ctx, phase) {
                return nil // User cancelled
            }
        }

        for _, step := range phase.Steps {
            if err := e.executeStep(ctx, step); err != nil {
                if e.canRollback(step) {
                    e.rollback(ctx, step.Checkpoint)
                }
                return err
            }
            e.saveCheckpoint(ctx, step)
        }
    }
}
```

**Pre-defined Workflows:**
- Feature implementation
- Bug fix with verification
- Code review and cleanup
- Migration scripts

**Core Features:**
- Workflow definition (YAML or code)
- Checkpoint persistence
- User approval gates
- Automatic rollback on failure
- Progress visualization

#### 4.4 Session Branching
**Files:**
- `internal/session/branch.go` (NEW)
- Database migration

```go
type Branch struct {
    ParentID string
    Name     string
    Status   BranchStatus // active, selected, abandoned
}

func (s *BranchService) CreateBranch(ctx context.Context, parentID, name string) (Session, error) {
    // Copy parent session messages
    // Create new session with branch metadata
}

func (s *BranchService) SelectBranch(ctx context.Context, branchID string) error {
    // Mark as selected
    // Abandon sibling branches
    // Merge into parent
}
```

**Deliverables:**
- Database migration for branching
- Branch API implementation
- Git integration for file branches
- UI for branch selection

### Success Criteria
- [ ] Multi-modal inputs processed correctly
- [ ] Browser automation with visual feedback
- [ ] Complex workflows with checkpoint recovery
- [ ] Session branching operational

**Estimated Time:** 24-32 hours

---

## STAGE 5: POLISH (Weeks 13-14) - STABLE RELEASE

**Priority:** MEDIUM - Ensure production readiness
**Status:** NOT STARTED
**Owner:** TBD

### Overview
Final polish before stable v4.0 release.

### Tasks

#### 5.1 Documentation Complete
- [ ] User guide updated with all v4.0 features
- [ ] API documentation complete
- [ ] Migration guide from v3.x to v4.0
- [ ] Architecture documentation
- [ ] Troubleshooting guide

#### 5.2 All Tests Passing
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests
- [ ] E2E tests for critical workflows
- [ ] Performance tests meet benchmarks

#### 5.3 Performance Benchmarks Met
| Metric | Target |
|--------|--------|
| MCP uptime | 99%+ (auto-recovery) |
| Operation speed | 50%+ faster on multi-op tasks |
| Context efficiency | 2x effective context window |
| Symbol search | <100ms for typical queries |

#### 5.4 Security Audit Complete
- [ ] Dependency audit
- [ ] Code vulnerability scan
- [ ] Penetration testing
- [ ] OWASP top 10 checklist

### Success Criteria
- [ ] Documentation 100% complete
- [ ] All tests passing in CI
- [ ] All benchmarks met or exceeded
- [ ] Security audit passed with no critical issues

**Estimated Time:** 16-20 hours

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
│  Multimodal                  Stage 4.1 (vision integration)                  │
│  Long context                Stage 3.2 (smart compression)                    │
│  Computer use                Stage 4.2 (workflow engine)                     │
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

## IMPLEMENTATION SUMMARY

| Stage | Duration | Target | Key Deliverables |
|-------|----------|--------|------------------|
| 0: Foundation | 1 week | Boot optimization | Tool registry, environment state, version changelog, push hook |
| 1: Reliability | 2 weeks | v3.8 | MCP health, SUPERCACHE namespaces, SafeOps complete |
| 2: Performance | 2 weeks | v3.9 | Parallel bash, streaming progress, multi-agent headless |
| 3: Intelligence | 3 weeks | v3.95 | Symbol index, context compression, agent auto-selection |
| 4: Competitive | 4 weeks | v4.0 | Multi-modal, browser automation, workflow engine |
| 5: Polish | 2 weeks | Stable | Docs, tests, benchmarks, security audit |

**Total Time:** 14 weeks to stable v4.0

---

## DEPENDENCY GRAPH

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  DEPENDENCIES (must complete before starting dependent stage)                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  Stage 0 (Foundation)                                                          │
│    ├─> Tool registry ──> All agent intelligence features                      │
│    ├─> Environment state ──> Component integration                             │
│    └─> Version changelog ──> Feature discovery                                │
│                                                                                 │
│  Stage 1 (Reliability)                                                         │
│    ├─> MCP health ──> Stable multi-agent operations                          │
│    ├─> SUPERCACHE namespaces ──> Multi-project support                        │
│    └─> SafeOps complete ──> CI/CD integration                                 │
│                                                                                 │
│  Stage 2 (Performance)                                                         │
│    ├─> Parallel bash ──> Faster multi-agent coordination                     │
│    ├─> Streaming progress ──> Better UX for long ops                         │
│    └─> Browork + Hivemind ──> Stage 4 workflow engine                        │
│                                                                                 │
│  Stage 3 (Intelligence)                                                        │
│    ├─> Symbol index ──> Semantic search without embeddings                    │
│    ├─> Context compression ──> Longer effective sessions                      │
│    └─> Agent auto-selection ──> Stage 4 workflows                            │
│                                                                                 │
│  Stage 4 (Competitive)                                                         │
│    ├─> Multi-modal ──> Browser automation                                   │
│    ├─> Browser automation ──> Computer use                                    │
│    └─> Workflow engine ──> Complex task automation                           │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## QUICK START: BEGIN TODAY

### Stage 0 Tasks (This Week)

```bash
# Day 1-2: Tool Registry
cd /Volumes/Storage/floyd-main
# Edit: internal/agent/coordinator.go
# Add: lab_get_tool_registry call, cache_store("system:tool_registry")
# Test: ./floyd4
# Should show: "Tools available: 105+ from 18 servers"

# Day 3: Environment State
# Edit: internal/cmd/root.go
# Add: cache_store("system:environment_state", environmentState)

# Day 4: Version Changelog
# Create: internal/version/changelog.go
# Add: v4.0.0 feature list

# Day 5: Push Verification Hook
cd /Volumes/Storage/floyd-main
cat > .git/hooks/pre-push << 'SCRIPT'
#!/bin/bash
CODE="PUSH-$(head /dev/urandom | LC_ALL=C tr -dc A-Z0-9 | head -c 6)"
echo "Enter code to confirm push: $CODE"
read -p "Verification code: " INPUT
if [ "$INPUT" != "$CODE" ]; then
    echo "Code mismatch. Push cancelled."
    exit 1
fi
echo "Verified. Proceeding with push."
exit 0
SCRIPT
chmod +x .git/hooks/pre-push

# Repeat for FloydDesktopWeb-v2, MCP, FloydDeployable
```

---

## NEXT STEPS

1. **Review this plan** and confirm the approach aligns with your goals
2. **Assign ownership** for each stage
3. **Set up tracking** (create GitHub issues/project board)
4. **Begin Stage 0** immediately (boot optimization)
5. **Establish checkpoints** after each stage

---

*Ideal Implementation Path - Generated 2026-02-22*
*Consolidates: FULL_UTILIZATION_ROADMAP, safeops-complete-summary, phase2-progress-report,*
*refactoring-templates, phase2-context-singularity-integration, FLOYD_EVOLUTION_ROADMAP,*
*V4_TECHNICAL_SPECS, PUSH_VERIFICATION_HOOK*
