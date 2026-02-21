# Terminal Shadow - Design Document v1.0

**Purpose:** A Python harness that acts as a "black box recorder" for terminal sessions, automatically populating the HANDOFF.md SSOT document with real-time progress, errors, and decisions.

**Created:** 2026-02-21
**Status:** Design Phase - Not Yet Implemented

---

## The Problem Being Solved

```
COMPACTION DRIFT TIMELINE
══════════════════════════════════════════════════════════════

Session Start    → Agent has 100% context
    ↓
Work happens     → Decisions made, errors encountered, fixes applied
    ↓
Context limit    → Compaction occurs
    ↓
Summary created  → 35-45% of context LOST
    ↓
Next session     → Agent doesn't remember:
                   - Why decision X was made
                   - What error Y meant
                   - What was almost tried but rejected
                   - What the user's actual priority was

══════════════════════════════════════════════════════════════

THE INSIGHT: The agent shouldn't have to REMEMBER.
              It should have to READ.
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  TERMINAL SHADOW ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │   User      │────▶│   Shadow    │────▶│   Shell     │   │
│  │  Command    │     │   Wrapper   │     │  Execution  │   │
│  └─────────────┘     └─────────────┘     └─────────────┘   │
│                             │                               │
│                             ▼                               │
│                      ┌─────────────┐                        │
│                      │   Event     │                        │
│                      │   Filter    │                        │
│                      └─────────────┘                        │
│                             │                               │
│              ┌──────────────┼──────────────┐               │
│              ▼              ▼              ▼               │
│       ┌──────────┐   ┌──────────┐   ┌──────────┐          │
│       │  Error   │   │ Success  │   │ Decision │          │
│       │  Logger  │   │  Logger  │   │  Logger  │          │
│       └──────────┘   └──────────┘   └──────────┘          │
│              │              │              │               │
│              └──────────────┼──────────────┘               │
│                             ▼                               │
│                      ┌─────────────┐                        │
│                      │  HANDOFF.md │                        │
│                      │   SSOT      │                        │
│                      └─────────────┘                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Shadow Wrapper

**Purpose:** Intercepts all shell commands and captures their output.

```python
# Conceptual structure
class TerminalShadow:
    def __init__(self, project_path: str):
        self.project_path = project_path
        self.handoff_path = os.path.join(project_path, "HANDOFF.md")
        self.event_log = []
        self.filter = EventFilter()

    def execute(self, command: str) -> CommandResult:
        """Execute a command and shadow the output."""
        # 1. Record the command
        event = CommandEvent(
            timestamp=datetime.now(),
            command=command,
            working_dir=os.getcwd()
        )

        # 2. Execute and capture
        result = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True
        )

        # 3. Record the result
        event.stdout = result.stdout
        event.stderr = result.stderr
        event.exit_code = result.returncode

        # 4. Filter for significance
        if self.filter.is_significant(event):
            self.log_event(event)

        # 5. Return to user
        return result
```

### 2. Event Filter

**Purpose:** Determines what's worth recording to avoid noise.

```python
class EventFilter:
    IGNORE_COMMANDS = {
        'ls', 'cd', 'pwd', 'clear', 'exit',
        'cat', 'head', 'tail', 'less', 'more',
        'echo', 'printf'
    }

    SIGNIFICANT_PATTERNS = [
        r'error', r'Error', r'ERROR',
        r'exception', r'Exception', r'EXCEPTION',
        r'fail', r'Fail', r'FAIL',
        r'traceback', r'Traceback',
        r'fatal', r'Fatal', r'FATAL',
    ]

    def is_significant(self, event: CommandEvent) -> bool:
        # Always capture errors
        if event.exit_code != 0:
            return True

        # Always capture git commits
        if event.command.startswith('git commit'):
            return True

        # Always capture build/test commands
        if any(cmd in event.command for cmd in ['build', 'test', 'deploy']):
            return True

        # Check for error patterns in output
        for pattern in self.SIGNIFICANT_PATTERNS:
            if re.search(pattern, event.stderr) or re.search(pattern, event.stdout):
                return True

        return False
```

### 3. Handoff Updater

**Purpose:** Appends structured entries to HANDOFF.md.

```python
class HandoffUpdater:
    SECTIONS = {
        'error': '## LOST CONTEXT INSURANCE / Debugging History',
        'success': '## COMPLETED THIS SESSION',
        'decision': '## LOST CONTEXT INSURANCE / Decision Log',
    }

    def append_event(self, event: CommandEvent, event_type: str):
        """Append a formatted entry to the appropriate section."""
        section_marker = self.SECTIONS.get(event_type)

        if event_type == 'error':
            entry = self._format_error_entry(event)
        elif event_type == 'success':
            entry = self._format_success_entry(event)
        else:
            entry = self._format_decision_entry(event)

        self._insert_before_section_end(section_marker, entry)

    def _format_error_entry(self, event: CommandEvent) -> str:
        return f"""
**Issue:** Command failed with exit code {event.exit_code}

**Command:** `{event.command}`

**Symptoms:**
```
{event.stderr[:500]}
```

**Discovery Path:**
1. Hypothesis: [Agent fills in]
2. Result: [Agent fills in]

**Key Insight:** [Agent fills in after resolution]

---
*Auto-logged: {event.timestamp}*
"""

    def _format_success_entry(self, event: CommandEvent) -> str:
        return f"""
### ✓ Completed: {event.command}

**Timestamp:** {event.timestamp}

**Output:**
```
{event.stdout[:200]}
```

---
"""
```

### 4. Heartbeat Service

**Purpose:** Periodic updates even without significant events.

```python
class HeartbeatService:
    def __init__(self, interval_minutes: int = 5):
        self.interval = interval_minutes * 60
        self.running = False

    def start(self):
        """Start the heartbeat timer."""
        self.running = True
        while self.running:
            time.sleep(self.interval)
            self.emit_heartbeat()

    def emit_heartbeat(self):
        """Append a heartbeat entry to HANDOFF.md."""
        entry = f"""
### [HEARTBEAT] {datetime.now().strftime("%Y-%m-%d %H:%M")}

**Active Duration:** {self.get_session_duration()}
**Commands Executed:** {self.command_count}
**Last Command:** {self.last_command}
**Current Directory:** {os.getcwd()}

**Status:** Session active. No errors since last heartbeat.

---
"""
        self.updater.append_to_section("## SESSION METADATA", entry)
```

---

## Integration Points

### With Floyd (Primary Agent)

```
FLOYD INTEGRATION FLOW
══════════════════════════════════════════════════════════════

1. Floyd starts session
2. Floyd reads HANDOFF.md
3. Floyd now has full context from Terminal Shadow logs
4. Floyd works, executing commands through Shadow wrapper
5. Shadow captures all significant events
6. Floyd compacts → summary created
7. Next session: Floyd reads HANDOFF.md → recovers lost context

══════════════════════════════════════════════════════════════
```

### With AXIOM (Project Management)

AXIOM already has:
- `state/project_map.json` - Project tracking
- `state/hooks/` - Hook infrastructure
- Quality gates and task lifecycle

**Potential Integration:**
- Terminal Shadow could be an AXIOM module
- Use AXIOM's hook infrastructure for event capture
- Feed into AXIOM's status dashboard

---

## Implementation Phases

### Phase 1: MVP (Day 1-2)

**Goal:** Basic command capture and logging

**Deliverables:**
- `shadow.py` - Command wrapper script
- Event filter with basic patterns
- Simple HANDOFF.md appender

**Test Criteria:**
- Can run commands through shadow
- Errors are captured and logged
- HANDOFF.md is updated correctly

### Phase 2: Integration (Day 3-4)

**Goal:** Integrate with Floyd workflow

**Deliverables:**
- Wrapper script for Floyd execution
- Floyd reads HANDOFF.md on startup
- Auto-section detection in HANDOFF.md

**Test Criteria:**
- Floyd commands are shadowed
- Handoff is populated automatically
- Context recovery works after simulated compaction

### Phase 3: Intelligence (Day 5-7)

**Goal:** Smart summarization

**Deliverables:**
- LLM integration for error summarization
- Pattern detection for recurring issues
- Automatic "Last 20%" blocker identification

**Test Criteria:**
- Errors are summarized intelligently
- Recurring patterns are flagged
- Blockers are auto-identified

### Phase 4: Polish (Day 8-10)

**Goal:** Production readiness

**Deliverables:**
- Configuration file support
- Multiple project support
- Web dashboard for log viewing

**Test Criteria:**
- Works across all 10 sprint projects
- Configurable per-project settings
- Dashboard shows real-time status

---

## Configuration Schema

```yaml
# shadow_config.yaml

project:
  name: "FloydDeployable"
  path: "/Volumes/Storage/floyd-sandbox/FloydDeployable"
  handoff: "HANDOFF.md"

capture:
  ignore_commands:
    - ls
    - cd
    - pwd
    - clear
  capture_all_errors: true
  capture_git_commits: true
  capture_build_commands: true

heartbeat:
  enabled: true
  interval_minutes: 5

llm:
  enabled: true
  provider: "local"  # or "openai", "anthropic"
  model: "llama-7b"  # for local
  summarize_errors: true

output:
  max_entry_length: 1000
  include_timestamps: true
  format: "markdown"
```

---

## File Structure

```
terminal-shadow/
├── shadow.py              # Main entry point
├── config.yaml            # Configuration
├── src/
│   ├── __init__.py
│   ├── wrapper.py         # Command wrapper
│   ├── filter.py          # Event filtering
│   ├── updater.py         # HANDOFF.md updater
│   ├── heartbeat.py       # Periodic updates
│   └── llm_summarizer.py  # Optional LLM integration
├── tests/
│   ├── test_wrapper.py
│   ├── test_filter.py
│   └── test_updater.py
└── README.md
```

---

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Context Recovery | 85%+ | Agent can answer questions about prior work |
| Error Capture Rate | 100% | All errors logged to HANDOFF.md |
| Noise Ratio | <10% | Insignificant events filtered out |
| Heartbeat Reliability | 99% | Heartbeats occur on schedule |
| Setup Time | <5 min | Time to configure for new project |

---

## Open Questions

1. **LLM Integration:** Should we use a local model or API? Tradeoff: privacy vs. quality.

2. **Multi-Session Support:** How to handle concurrent sessions across projects?

3. **Git Integration:** Should Shadow auto-commit the HANDOFF.md updates?

4. **Retroactive Filling:** When agent resolves an error, should it retroactively fill in the "Key Insight" field?

5. **IDE Integration:** Can we hook into VS Code or other editors?

---

## Next Steps

1. **Decision Required:** Approve this design for implementation
2. **Resource Assignment:** Who builds Phase 1 MVP?
3. **Pilot Project:** Which project will be the first to use Shadow?
4. **Integration Plan:** How does this fit into the 10-day sprint?

---

*Document Version: 1.0*
*Author: Floyd v4.0.0 + Jim/Gemini Collaboration*
*Template: Floyd Handoff Template v1.0*
