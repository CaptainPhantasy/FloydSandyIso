# Pattern: Edit Tool String Matching Debug

**Date:** 2026-02-28
**Category:** debugging/tool-usage
**Severity:** medium

---

## Problem

The `edit` tool repeatedly failed with "old_string not found" error despite the text appearing to match when viewing the file.

## Root Cause

**Visual confusion between rendered markdown and raw file content.** The file used single pipe characters (`|`) as table column separators, but the edit attempts used double pipes (`||`) at the start of lines. This was caused by misreading how markdown tables render in the View tool output.

## Diagnostic Steps That Worked

1. **Raw byte inspection with `xxd`**:
   ```bash
   sed -n '38p' file.md | xxd
   ```
   This revealed the exact bytes including UTF-8 emoji encoding.

2. **Line-by-line verification with `cat -n`**:
   ```bash
   cat -n file.md | sed -n '37,40p'
   ```
   This showed exact line numbers and content.

## MCP Tools Most Relevant

| Tool | Use Case |
|------|----------|
| `floyd-patch/edit_range` | Line-based editing bypasses string matching entirely |
| `floyd-explorer/smart_replace` | Surgical edits with uniqueness validation |
| `floyd-terminal/execute_command` | Raw byte inspection for encoding diagnosis |
| `floyd-supercache/cache_store_pattern` | Store this pattern for reuse |

## Prevention

- Always use `cat -n` or `sed -n` to get exact line content before editing
- For markdown tables, remember: single `|` for columns, not `||`
- Use `xxd` or `od -c` when string matching mysteriously fails

## Key Insight

> When the edit tool fails repeatedly on "matching" text, the issue is NEVER with the tool—it's with the agent's perception of the text. Always inspect raw bytes.

---

## Related Patterns

- `edit-tool-whitespace-mismatch` - Tab vs space issues
- `line-ending-mismatch` - CRLF vs LF issues
- `unicode-normalization` - Different Unicode representations

---

*Crystallized by FLOYD v4.6 after Two-Failure Reset on HANDOFF.md edit*
