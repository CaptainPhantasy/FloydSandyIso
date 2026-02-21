# Terminal Shadow

**Black Box Recorder for Terminal Sessions**

A Python harness that captures terminal output and populates `HANDOFF.md` to maintain Single Source of Truth across AI agent session compactions.

## The Problem

When AI agents compact their context, 35-45% of knowledge is lost:
- Why a decision was made
- What error X meant
- What was almost tried but rejected
- What the user's actual priority was

## The Solution

Terminal Shadow acts as a "black box recorder" that:
1. Intercepts command executions
2. Filters for significant events (errors, builds, commits)
3. Logs structured entries to `HANDOFF.md`
4. Survives context compaction

## Quick Start

```bash
# Initialize shadow for a project
cd /path/to/your/project
python /path/to/terminal-shadow/shadow.py init

# Run commands through shadow
python shadow.py run "go build ./..."
python shadow.py run "npm test"

# Check status
python shadow.py status
```

## Integration with Floyd

### Method 1: Hook Integration (Recommended)

Call the shadow hook after each command:

```bash
# After Floyd executes a command, log it:
python floyd_shadow_hook.py log \
  --project /path/to/project \
  --command "go build ./..." \
  --exit-code 0 \
  --stdout "build succeeded" \
  --stderr ""
```

### Method 2: Go Integration

Use the Go package in your Floyd code:

```go
import "github.com/CaptainPhantasy/FloydSandyIso/terminal-shadow/shadow"

// Create logger
logger := shadow.NewLogger(&shadow.ShadowConfig{
    Enabled:     true,
    ProjectPath: "/path/to/project",
    HookPath:    "/path/to/floyd_shadow_hook.py",
})

// Log commands after execution
logger.LogCommand(shadow.CommandLog{
    Command:    "go build ./...",
    ExitCode:   0,
    Stdout:     "build succeeded",
    Stderr:     "",
    DurationMs: 1500,
})
```

### Method 3: Python Module

Use directly in Python:

```python
from shadow import ShadowSession

with ShadowSession("/path/to/project") as session:
    # Commands are automatically logged
    result = session.execute("go build ./...")
    print(result.stdout)
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  TERMINAL SHADOW ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   Command Execution ──► Event Filter ──► Handoff Updater   │
│                              │                              │
│                              ▼                              │
│                    ┌─────────────────┐                      │
│                    │ Significance?   │                      │
│                    │ - Error?        │                      │
│                    │ - Build/Test?   │                      │
│                    │ - Git commit?   │                      │
│                    └─────────────────┘                      │
│                              │                              │
│              ┌───────────────┼───────────────┐             │
│              ▼               ▼               ▼             │
│       ┌──────────┐   ┌──────────┐   ┌──────────┐          │
│       │  ERROR   │   │ SUCCESS  │   │ DECISION │          │
│       │  LOG     │   │  LOG     │   │  LOG     │          │
│       └──────────┘   └──────────┘   └──────────┘          │
│              │               │               │              │
│              └───────────────┴───────────────┘              │
│                              │                              │
│                              ▼                              │
│                       ┌───────────┐                         │
│                       │ HANDOFF.md │                        │
│                       └───────────┘                         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Event Classification

| Type | Criteria | Example |
|------|----------|---------|
| ERROR | Non-zero exit code | `go build` fails |
| SUCCESS | Build/test/commit succeeds | `npm test` passes |
| DECISION | Git operations | `git commit` |
| NOISE | Trivial commands | `ls`, `cd`, `cat` |

## HANDOFF.md Structure

The shadow populates these sections:

```markdown
## SESSION METADATA
*Auto-populated by Terminal Shadow*

## COMPLETED THIS SESSION
*Significant completions logged here*

## LOST CONTEXT INSURANCE / Debugging History
*Errors and resolutions logged here*

## LOST CONTEXT INSURANCE / Decision Log
*Key decisions and rationale logged here*
```

## Example Error Entry

```markdown
### ⚠ Error: assignment to entry in nil map

**Timestamp:** 2026-02-21 13:00:00

**Command:**
```bash
go run ./...
```

**Working Directory:** `/Volumes/Storage/floyd-sandbox/FloydDeployable`

**Exit Code:** 2

**Stderr:**
```
panic: assignment to entry in nil map

goroutine 1 [running]:
github.com/CaptainPhantasy/FloydSandyIso/internal/intelligence.(*SymbolIndex).Stats(...)
        /Volumes/Storage/floyd-sandbox/FloydDeployable/internal/intelligence/symbols.go:289
```

**Files Involved:**
- `internal/intelligence/symbols.go:289`

**Hypothesis:** [To be filled after investigation]

**Resolution:** [To be filled after fix]

---
```

## Configuration

Create `shadow_config.yaml` in your project:

```yaml
project:
  name: "MyProject"
  path: "/path/to/project"
  handoff: "HANDOFF.md"

capture:
  ignore_commands:
    - ls
    - cd
    - pwd
  capture_all_errors: true
  capture_git_commits: true
  capture_build_commands: true

heartbeat:
  enabled: true
  interval_minutes: 5

output:
  max_entry_length: 2000
  include_timestamps: true
```

## CLI Commands

```bash
# Initialize configuration
shadow init [--force]

# Run command with logging
shadow run "command" [--timeout 300]

# Show session status
shadow status

# Log a pre-executed command (for Floyd integration)
floyd_shadow_hook.py log \
  --command "go build" \
  --exit-code 0 \
  --stdout "..." \
  --stderr ""

# Extract error summary from stderr
floyd_shadow_hook.py extract-error --stderr "panic: ..."

# Check if command result is significant
floyd_shadow_hook.py check --command "ls" --exit-code 0
```

## Success Metrics

| Metric | Target |
|--------|--------|
| Context Recovery | 85%+ |
| Error Capture Rate | 100% |
| Noise Ratio | <10% |
| Setup Time | <5 min |

## Files

```
terminal-shadow/
├── shadow.py              # Main CLI entry point
├── floyd_shadow_hook.py   # Floyd integration hook
├── shadow.go              # Go integration package
├── src/
│   ├── __init__.py
│   ├── wrapper.py         # Command execution wrapper
│   ├── filter.py          # Event significance filter
│   ├── updater.py         # HANDOFF.md updater
│   ├── heartbeat.py       # Periodic status updates
│   └── config.py          # Configuration management
└── README.md
```

## Requirements

- Python 3.8+
- PyYAML (optional, for YAML config)

```bash
pip install pyyaml
```

## License

MIT

---

*Built by Floyd v4.0.0 for context continuity across session compactions.*
