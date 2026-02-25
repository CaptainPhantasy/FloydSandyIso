# Superpowers Skills Installation for FLOYD

## Overview

[Superpowers](https://github.com/obra/superpowers) is a collection of high-quality agent skills that enhance AI coding assistants. This document describes how these skills are installed and used with FLOYD.

## Installation

Superpowers is installed in FLOYD's configuration directory:

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│ Location                            │ Path                                 │
├─────────────────────────────────────┼──────────────────────────────────────┤
│ Superpowers Repository              │ ~/.config/floyd/superpowers          │
│ Skills Symlink                      │ ~/.config/floyd/skills/superpowers   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Installed Skills

The following skills are available (14 total):

```text
┌────────────────────────────────────────────────────────────────────────────┐
│ Skill Name                    │ Description                               │
├───────────────────────────────┼────────────────────────────────────────────┤
│ brainstorming                 │ Explore requirements before implementation│
│ dispatching-parallel-agents   │ Coordinate multiple agents in parallel    │
│ executing-plans               │ Execute implementation plans step-by-step  │
│ finishing-a-development-branch│ Complete and merge development branches   │
│ receiving-code-review         │ Handle feedback from code reviews         │
│ requesting-code-review        │ Request effective code reviews             │
│ subagent-driven-development   │ Drive development via subagents           │
│ systematic-debugging          │ Debug systematically with root cause      │
│ test-driven-development       │ Write tests first, then implement         │
│ using-git-worktrees           │ Use git worktrees for parallel work       │
│ using-superpowers             │ Guide to using superpowers skills         │
│ verification-before-completion│ Verify work before marking complete       │
│ writing-plans                 │ Create detailed implementation plans      │
│ writing-skills                │ Write effective agent skills              │
└────────────────────────────────────────────────────────────────────────────┘
```

## Usage

FLOYD automatically discovers skills from `~/.config/floyd/skills/`. The superpowers skills are symlinked and will be available in any FLOYD session.

### Triggering Skills

Skills are triggered by context. For example:
- "Brainstorm a new authentication system" → triggers `brainstorming`
- "Debug this failing test" → triggers `systematic-debugging`
- "Review this code" → triggers `requesting-code-review`

## Updating

To update superpowers to the latest version:

```bash
cd ~/.config/floyd/superpowers
git pull
```

## Skill Priority

FLOYD uses the following priority for skills:
1. **Project skills** (`.floyd/skills/` in your project)
2. **Personal skills** (`~/.config/floyd/skills/`)
3. **Superpowers skills** (`~/.config/floyd/skills/superpowers/`)

## Creating Custom Skills

Create your own skills in `~/.config/floyd/skills/my-skill/SKILL.md`:

```markdown
---
name: my-skill
description: Use when [condition] - [what it does]
---

# My Skill

[Your skill instructions here]
```

## Troubleshooting

### Skills not found

1. Check symlink: `ls -la ~/.config/floyd/skills/superpowers`
2. Verify skill files: `ls ~/.config/floyd/skills/superpowers/*/SKILL.md`
3. Restart FLOYD to re-discover skills

### Tool Mapping

When superpowers skills reference Claude Code tools:
- `TodoWrite` → FLOYD's `todos` tool
- `Task` with subagents → FLOYD's agent spawning
- File operations → FLOYD's native tools (`view`, `edit`, `write`)

## Resources

- [Superpowers Repository](https://github.com/obra/superpowers)
- [Agent Skills Specification](https://agentskills.io)
- [FLOYD Documentation](./)
