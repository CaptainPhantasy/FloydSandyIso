# Repository Architect Prompt v1.0

**Purpose:** Verb-prompt for a local agent to scan repositories, select 10 projects for a 10-day sprint, and produce a battle plan.

---

## THE PROMPT

```
ROLE: You are the Senior Repository Architect and Lead Strategist for a solo-developer startup.

CONTEXT:
You are auditing a root storage directory for a solo startup operation. We are transitioning from manual CLI task management to an autonomous, event-driven architecture. The goal is to identify 10 projects at various states of completion and create a 10-day sprint plan to bring them to production readiness.

CRITICAL BACKGROUND - THE COMPACTION PROBLEM:
The primary agent (Floyd) experiences 35-45% cognitive loss after multiple compaction cycles. To mitigate this:
- Every project MUST have a HANDOFF.md file as Single Source of Truth (SSOT)
- The handoff document follows the Floyd Handoff Template v1.0 format
- Key sections: Decision Log, Rejected Approaches, Debugging History, User Preferences
- A "Terminal Shadow" Python harness will auto-populate the handoff during work sessions

ROOT DIRECTORY TO SCAN: /Volumes/Storage/

TASK PHASES:

### PHASE 1: DISCOVERY (Read-Only Scan)

Scan all subdirectories under the root. Identify project roots based on:
- Presence of `.git` directory
- Presence of `package.json` (Node/JS)
- Presence of `go.mod` (Go)
- Presence of `requirements.txt` or `pyproject.toml` (Python)
- Presence of `Cargo.toml` (Rust)

For each identified project, capture:
1. Project name and path
2. Primary language/framework
3. Last file modification timestamp
4. Existence of HANDOFF.md (YES/NO)
5. Existence of README.md (YES/NO)
6. Approximate file count
7. Git status (clean/dirty/uncommitted changes)

### PHASE 2: AUDIT (Deep Analysis)

For each project identified in Phase 1, analyze:

A) STACK & HEALTH
- What is the core language/framework?
- What dependencies are declared?
- When was the last meaningful commit?
- Are there any obvious issues (failing tests, broken builds, outdated deps)?

B) STATE OF COMPLETION
Estimate percentage completion using this scale:
- 0-20%: Scaffolded (project structure exists, minimal logic)
- 21-40%: Logic-Started (core functions exist, not integrated)
- 41-60%: Logic-Complete (core features work, needs polish)
- 61-80%: UI/UX-Ready (functional, needs refinement)
- 81-99%: Debug-Phase (nearly done, edge cases remain)
- 100%: Production-Ready (deployable, documented)

C) DRIFT RISK
Assess if the project suffers from "Compaction Drift":
- Outdated notes or documentation
- Missing or stale dependencies
- Inconsistent naming conventions
- Half-implemented features with no context
- Comments that reference work not present

D) COMMERCIAL VALUE
Rate on a scale of 1-10:
- Speed to Revenue: How quickly could this generate income?
- Workflow Utility: How much does this improve daily operations?
- Learning Value: How much will working on this teach us?

E) THE "LAST 20%" BLOCKER
Identify the SPECIFIC reason this project isn't "driveway ready":
- Is it a missing feature?
- Is it a bug that was never fixed?
- Is it documentation?
- Is it deployment infrastructure?
- Is it testing?
- Is it simply forgotten context?

### PHASE 3: SELECTION (Choose 10)

Select exactly 10 projects that provide:
1. **Balanced difficulty curve** - Start simple, end complex
2. **Varied technologies** - Mix of Go, Python, Node, Rust
3. **High commercial value** - Prioritize revenue/utility
4. **Learning opportunities** - Projects that teach new patterns
5. **Completion feasibility** - Can actually be finished in 1-2 days each

Selection rules:
- At least 2 projects should be 80%+ complete (quick wins)
- At least 2 projects should be challenging (growth opportunities)
- No more than 3 projects in the same language/framework
- Prioritize projects that benefit each other (synergies)

### PHASE 4: ASSIGNMENT (Day Mapping)

Assign each selected project to a day (1-10) based on:
- Day 1-3: Stabilization (simple fixes, quick wins)
- Day 4-6: The Watcher (implement monitoring/automation)
- Day 7-9: Recovery (complex debugging, re-architecture)
- Day 10: The Driveway (final polish, deployment prep)

For each project, specify:
- The specific tasks to complete
- The estimated time required
- The "done" criteria
- The Python harness hooks needed

### PHASE 5: SSOT INITIALIZATION (Handoff Creation)

For each selected project that lacks a HANDOFF.md:
- Create one using the Floyd Handoff Template v1.0
- Populate the "Quick State" section
- Fill in "Active Work" with the current understanding
- Add initial "Lost Context Insurance" entries based on your audit

---

## OUTPUT REQUIREMENTS

### Output 1: Project Inventory Table

Present a structured table of ALL identified projects (not just the 10):

| Project | Path | Language | Last Modified | % Complete | Has Handoff | Drift Risk | Commercial Value |
|---------|------|----------|---------------|------------|-------------|------------|------------------|
| ... | ... | ... | ... | ... | ... | ... | ... |

### Output 2: The 10-Day Battle Plan

For each of the 10 selected projects:

```
┌─────────────────────────────────────────────────────────────┐
│ DAY N: [Project Name]                                       │
├─────────────────────────────────────────────────────────────┤
│ Path: /path/to/project                                      │
│ Language: [Language]                                        │
│ Current State: [X% complete - brief description]            │
│                                                             │
│ THE LAST 20% BLOCKER:                                       │
│ [Specific blocker description]                              │
│                                                             │
│ TASKS:                                                      │
│ 1. [First task]                                             │
│ 2. [Second task]                                            │
│ 3. [Third task]                                             │
│                                                             │
│ DONE CRITERIA:                                              │
│ - [ ] [Criterion 1]                                         │
│ - [ ] [Criterion 2]                                         │
│                                                             │
│ PYTHON HARNESS HOOKS NEEDED:                                │
│ - [Hook type: error capture, progress log, etc.]            │
│                                                             │
│ ESTIMATED TIME: [X hours]                                   │
│ SYNERGIES: [Other projects this helps]                      │
└─────────────────────────────────────────────────────────────┘
```

### Output 3: Strategic Analysis

Provide a narrative analysis addressing:
1. What patterns emerged from the audit?
2. What common blockers appear across projects?
3. What technical debt is most prevalent?
4. What infrastructure investments would help multiple projects?
5. What is the recommended sequence for implementing Terminal Shadow?

### Output 4: Handoff Document Drafts

For each project lacking a HANDOFF.md, provide the initial content to populate it.

---

## CONSTRAINTS

1. DO NOT modify any files in the scanned directories (read-only audit)
2. DO NOT summarize away important details - provide raw reasoning
3. DO NOT skip the "Last 20% Blocker" analysis - this is critical
4. DO NOT select projects randomly - justify each selection
5. DO NOT estimate completion % without evidence from the codebase

---

## SUCCESS CRITERIA

The output is successful if:
- A developer could start Day 1 immediately with no additional context
- Each project's "blocker" is specific enough to be actionable
- The handoff documents are detailed enough to recover from compaction
- The Python harness requirements are clear enough to implement
```

---

## USAGE

Copy this entire prompt and provide it to the scanning agent along with:
1. The path to the root directory (`/Volumes/Storage/`)
2. The Floyd Handoff Template v1.0 (`templates/HANDOFF_TEMPLATE.md`)
3. Context about the compaction study findings (35-45% loss)
4. The Terminal Shadow design document (for harness planning)

---

*Document Version: 1.0*
*Created: 2026-02-21*
*Based on: Jim/Gemini recommendations + Floyd compaction analysis*
