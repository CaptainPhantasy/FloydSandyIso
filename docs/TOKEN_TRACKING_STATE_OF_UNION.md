# State of the Union: Token Tracking, Summarization, and the Semantic Filter

## Executive Summary

You survived the weekend flying blind. Here's what happened, what's fixed, and what's still broken.

---

## The Two Bugs That Fucked You

### Bug 1: Double-Layer Reasoning (FIXED ✅)

**What it was:** The Think toggle only worked for Anthropic. Every other provider (OpenAI, Azure, OpenRouter, Vercel, Google, OpenAICompat) injected `reasoning_effort` unconditionally — even when Think was OFF.

**What it did:**
- Burned ~2x tokens invisibly
- You'd hit 200K actual when display said 100K
- Model would start "drooling on itself" at ~60% because it was actually at 120%

**Fix committed:** `coordinator.go` now checks `model.ModelCfg.Think` before injecting reasoning parameters for ALL providers.

---

### Bug 2: Token Display Shows Effective, Not Total (NOT FIXED ❌)

**Location:** `internal/ui/model/header.go:213`

```go
// CURRENT (WRONG)
contextUsed := session.CompletionTokens + session.PromptTokens - session.CacheReadTokens
```

**What it does:**
- Subtracts cached tokens from display
- Shows "effective" cost (268) instead of actual consumption (80,460)
- You see 1% when you're actually at 48%

**Why it's wrong:** Cached tokens don't cost money, but they DO consume context window. The display lies about context pressure.

**Fix needed:**
```go
// CORRECT
contextUsed := session.CompletionTokens + session.PromptTokens
```

---

## The Token Wipe During Summarization

**Location:** `internal/agent/agent.go:884-885`

```go
currentSession.CompletionTokens = usage.OutputTokens
currentSession.PromptTokens = 0  // <-- THIS WIPES YOUR TOKEN COUNT
```

**What happens:**
1. Summarization runs at 80% context
2. It's a small side-request to generate the summary
3. After it finishes, it overwrites the session's token totals with just the summary's tokens
4. Your token display jumps from 150K to 2K

**Why you might NOT want to fix this:**
- After summarization, you're starting "fresh" with a compressed context
- The old token count is no longer accurate anyway
- It could be seen as "resetting the meter" post-compaction

**Why you MIGHT want to fix this:**
- It's confusing as hell to see the number jump
- You lose visibility into actual context consumption
- The summary request's tokens are tiny compared to what you had

**My recommendation:** Fix it. Track cumulative tokens separately, or don't overwrite at all. But this is lower priority than the display bug.

---

## Is Summarization On?

**YES.** Here's the flow:

| Threshold | What Happens |
|-----------|--------------|
| 70% | Summarization threshold (defined but not used for trigger) |
| 80% | **Summarization triggers** - compresses context, extends session |
| 95% | **Hard handoff** - creates HANDOFF.md, exits gracefully |

**The `DisableAutoSummarize` flag exists but is NOT CHECKED.** It's stored but ignored at the trigger point (`agent.go:714`).

Now that double-reasoning is fixed, you WILL reach summarization instead of invisibly burning past it.

---

## The Semantic Filter: IS IT THERE?

**YES.** It exists and is functional.

### What It Is

**Tool:** `query_floyd_archive`

**Location:** `internal/agent/tools/archive_query.go`

**What it does:**
- Queries the persistent message database
- **Semantic firewall:** Only returns:
  - Tool calls and their inputs
  - Tool results
  - Code blocks
- **Excludes:** Conversational text, chitchat, persona bullshit

**Implementation:** `message.go:203-253` — `SearchTechnicalArchive()`

The SQL query explicitly filters:
```sql
WHERE parts LIKE '%query%'
AND (
    (role = 'assistant' AND parts LIKE '%"type":"tool_call"%')
    OR role = 'tool'
    OR (role = 'user' AND parts LIKE '%```%')
)
```

### Is It Being Used?

**It's registered** (`coordinator.go:513`) but **NOT in the default allowed tools list** for any agent.

The handoff instructions (`agent.go:1318`) tell the next session to use it:
> "Upon starting the new session, immediately use `query_floyd_archive` to retrieve the technical context..."

But the tool isn't pre-approved, so the model has to request permission to use it.

---

## How Far From "Index as SSOT"?

**Current state:** The pieces exist but aren't connected.

| Component | Status |
|-----------|--------|
| `query_floyd_archive` tool | ✅ Implemented |
| Semantic filter (no chitchat) | ✅ Implemented |
| Registered in coordinator | ✅ Done |
| In default allowed tools | ❌ NOT ADDED |
| Auto-queried on session start | ❌ MANUAL ONLY |
| Replaces summarization | ❌ NOT INTEGRATED |

**What's missing:**
1. Add `query_floyd_archive` to allowed tools in config
2. Auto-query on session start (or make it part of the boot sequence)
3. Decide: does this REPLACE summarization or COMPLEMENT it?

---

## This Session: Will You Survive?

**Current status:**
- You're at ~48% actual (98K of 202K)
- Display says 1% and 3,100 tokens (wrong)
- Double-reasoning is fixed in the binary you just built

**What will happen:**
1. You'll hit 80% actual context (~162K tokens)
2. Summarization will trigger
3. Token display will jump to near-zero (the wipe bug)
4. Session continues with compressed context
5. You'll eventually hit 95% and get handoff

**You will NOT experience the invisible burn anymore.** The model will make it to summarization.

---

## Priority Fixes

| Priority | Bug | Impact | Fix Difficulty |
|----------|-----|--------|----------------|
| 1 | Token display (header.go:213) | You're flying blind | 1 line change |
| 2 | Add query_floyd_archive to allowed tools | Can't auto-query archive | 1 line in config |
| 3 | Token wipe during summarization | Confusing jumps | Small refactor |

---

## The Weekend In Hindsight

You had:
- ❌ Double reasoning burning 2x tokens invisibly
- ❌ Token display showing 1/300th of reality
- ❌ No summarization (because you burned past the threshold invisibly)
- ❌ No warning before the wall

It's genuinely impressive you survived. The fixes are straightforward. The semantic filter is already built — it just needs to be turned on.

---

*Generated: 2026-03-01 22:00 UTC*
*Session token count: ~98K actual, ~4K displayed (the bug in real-time)*
