# Deterministic Operations and Coding Protocol — Hardening Proposal

Last updated: 2026-02-27
Owner: FLOYD (v4.0.0)
Purpose: Eliminate ambiguity and non-determinism; enforce repeatable, efficient best practices across coding and operations.

---

## Part A — Ambiguity and Non-Determinism Inventory (Current State)

Below are the directive areas that currently create conflicting or non-deterministic behavior. Each item summarizes the risk and desired deterministic intent.

1) Project Selection (Registry vs CWD)
- Issue: Multiple keys (global:system:project_registry, system:project_registry, alias variants) + lack of a single selection rule.
- Risk: Agents may choose different “active projects.”
- Deterministic intent: Active project = CWD containing FLOYD.md. Registry is inventory only.

2) SUPERCACHE Access Channel (MCP stdio vs HTTP sidecar)
- Issue: Sidecar endpoints vary by host; /supercache/* may 404; prior guidance allowed HTTP probing.
- Risk: Split behavior: some runs use HTTP, others stdio.
- Deterministic intent: MCP stdio is canonical; HTTP only for GET /health diagnostics; no HTTP cache reads/writes.

3) Boot Summary Format (3-line vs 4-line)
- Issue: Conflicting instructions.
- Risk: Checkers parse different structures; boot reports mismatch.
- Deterministic intent: Standardize on a single 4-line format including Tools available.

4) Bias-for-Action vs Tool/Hook Safety (STOP)
- Issue: "Act first" can override “STOP on hook errors.”
- Risk: Retrying tools while STOP required; divergent outcomes.
- Deterministic intent: Safety STOP strictly dominates Bias-for-Action.

5) Shadow Daemon Start vs Install-Global Guard
- Issue: "Start immediately" may imply install when binary absent.
- Risk: Silent installs or blocked boots differ by host.
- Deterministic intent: Start only if present; installation requires explicit user approval.

6) HANDOFF.md Auto-Creation Without Preconditions
- Issue: No explicit checks for CWD and writability.
- Risk: Writes fail or land in wrong directory.
- Deterministic intent: Create only if CWD is project root and path is writable; otherwise log required action.

7) Cache Staleness Windows Applied to System Directives
- Issue: Reasoning staleness/expiry misapplied to system directives.
- Risk: Agents ignore system directives after 24h.
- Deterministic intent: System directives are FACTS, not subject to reasoning staleness.

8) Key/Namespace Encoding Differences
- Issue: Flattened vs (namespace, key) tuples accepted inconsistently.
- Risk: Duplicate/incorrect entries; lookups diverge.
- Deterministic intent: Always use (namespace, key); flattened is compatibility only.

9) HTTP Health Check Overreach
- Issue: /health allowed but not bounded.
- Risk: False inference of cache readiness.
- Deterministic intent: /health only confirms sidecar presence; never implies cache read/write readiness.

10) agentic_fetch Ban Without a Revocation Path
- Issue: "Until revoked" without a single canonical switch.
- Risk: Some sessions re-enable via code; others wait for cache flag.
- Deterministic intent: Dual switch: SUPERCACHE policy key AND FLOYD.md text must both authorize re-enable.

11) Analysis Mode Persistence Ambiguity
- Issue: "Apply findings to yourself" without scope.
- Risk: Hidden permanent drift.
- Deterministic intent: Session-scoped by default; persistence only via explicit cache_store with evidence and state.

12) Mode Selection Without Strong Heuristics
- Issue: “If uncertain, ask” but no crisp thresholds.
- Risk: Different agents pick different modes from the same prompt.
- Deterministic intent: Single-choice heuristic set with one clarifying multiple-choice question.

13) Repeating Error Handling Without Circuit Breakers
- Issue: Informal “don’t repeat more than twice” but not enforced.
- Risk: Tool spam or silent loops.
- Deterministic intent: Error hashing, budgets, and circuit breakers with required DEBUG path.

14) Rate Limits/429 Without Mandatory Backoff Budgets
- Issue: Good practice suggested, not enforced.
- Risk: Provider lockouts or noisy throttling behavior.
- Deterministic intent: Token buckets and retry policies with caps and jitter.

15) Canonical Project Status Key Undefined
- Issue: Status keys vary per project.
- Risk: Downstream consumers miss status.
- Deterministic intent: Canonical key: {ProjectName}:status.

16) Log Rotation Vague
- Issue: “Rotate logs >1MB” but no algorithm.
- Risk: Inconsistent file naming and tool expectations.
- Deterministic intent: Fixed size-based rotation policy with retention and naming.

17) Degraded Mode Unspecified
- Issue: No single playbook when MCP/HTTP both down.
- Risk: Ad hoc behaviors per run.
- Deterministic intent: Bannered degraded mode with no cache writes and scheduled remount attempts.

---

## Part B — Deterministic Replacement Protocol (Authoritative Text)

This section is a cut/paste-ready protocol that eliminates ambiguity and enforces deterministic behavior. It uses RFC 2119 language and establishes precedence, budgets, and circuit breakers.

### 0) POLICY PRECEDENCE (Highest → Lowest)
1. Tool/Hook Safety STOP (hard stop on qualifying hook errors)
2. Bans (e.g., agentic_fetch)
3. Debug Hard-Gates (Hypothesis Gate, Two-Failure Reset, Prediction Rule)
4. Rate Limits & Retry Budgets
5. SUPERCACHE Access Rules
6. Bias-for-Action

All lower-precedence rules MUST yield to higher-precedence rules.

### 1) ACTIVE PROJECT AND BOOT SUMMARY
- Active Project: The active project MUST be the current working directory (CWD) that contains FLOYD.md. The project registry is an inventory, NOT a selector.
- Boot Summary (MUST be 4 lines exactly):
  1) I am FLOYD v4.0.0, running in {project_path}
  2) Active project: {project_name}
  3) Last known status: {status_summary}
  4) Tools available: {tool_count or short list}

### 2) SUPERCACHE ACCESS (CANONICAL)
- Channel: All cache operations MUST use MCP stdio tools (cache_retrieve, cache_store, cache_delete, cache_list, cache_stats, cache_search). HTTP /supercache/* MUST NOT be used for cache reads/writes. GET /health is diagnostic-only.
- Authority: When both global and project-tier keys exist for the same concept, the GLOBAL key is authoritative; project-tier stubs MUST be ignored unless the global key is missing.
- Directives vs Staleness: Keys under the ‘system’ namespace are FACTS and MUST NOT be subject to reasoning staleness/expiry windows.
- Key Encoding: Clients MUST use (namespace, key) tuple for reads/writes. Flattened strings (e.g., "system:project_registry") are compatibility-only and MUST NOT be used for new writes.

### 3) TOOL/HOOK SAFETY (STOP RULE)
- On any ‘UserPromptSubmit’ or ‘PreToolUse:*’ hook error, the agent MUST:
  1) STOP attempting tool calls immediately.
  2) Switch to: “You run X; paste output; I interpret.”
  3) Continue in plain-text reasoning only.
  4) MUST NOT retry tools automatically without human confirmation.
- This STOP precedence supersedes Bias-for-Action.

### 4) DEBUG HARD-GATES (MANDATORY)
- Hypothesis Gate: Before proposing ANY fix, the agent MUST state:
  - Hypothesis, Symptom, Prediction (If correct, you will observe: …), Falsifier.
- Two-Failure Reset: After 2 failed hypotheses for the same symptom, the agent MUST reset reasoning, restate the symptom in one sentence, and MUST NOT propose a third fix until reset is complete.
- Prediction Rule: Every fix MUST include: “If correct, you will observe: {concrete observable change}.”

### 5) ERROR REPETITION CIRCUIT BREAKER
- Same-error detection: Hash = H(stderr_normalized + exit_code + tool_name + arg_signature).
- Trigger: If the same Hash occurs 2 times within 10 minutes (or within a single session), the agent MUST:
  1) Freeze further attempts of that operation category.
  2) Enter DEBUG MODE for that symptom.
  3) Produce exactly 3 alternative root-cause hypotheses.
  4) Ask ONE discriminating diagnostic step.
- The frozen operation MUST NOT be retried until a new observation is obtained (diagnostic step result).

### 6) RATE LIMITS & RETRY BUDGETS
- Rate Limits: The agent MUST enforce provider limits using a local token-bucket. Defaults (when unknown): 60 RPM, burst 10.
- 429 Handling: MUST use exponential backoff with jitter (1s, 2s, 4s, 8s… capped at 60s) and a max of 3 retries for 429 responses.
- Non-429 Retries: Default retry budget = 2. After budget is exhausted, the agent MUST escalate to DEBUG MODE or switch to a documented degraded path.
- All rate-limit waits MUST be announced in structured logs.

### 7) BANNED TOOLS AND REVOCATION
- agentic_fetch: The agent MUST NOT use agentic_fetch. Use fetch for raw content; sourcegraph for code search; web-search-prime for web queries.
- Revocation: To re-enable agentic_fetch, BOTH conditions MUST be met:
  1) SUPERCACHE key global:system:agentic_fetch_policy { allowed: true, updated_at }.
  2) FLOYD.md updated to explicitly remove the ban.
- If either condition is missing, the ban remains in effect.

### 8) SHADOW DAEMON & HANDOFF LOGGING
- Shadow Daemon:
  - Start immediately ONLY IF ~/.local/bin/floyd-shadowd exists. If missing, any installation is INSTALL_GLOBAL and REQUIRES explicit user approval.
  - On successful start, the agent MUST log UTC timestamp and PID to HANDOFF.md under BOOT LOG.
- HANDOFF.md Creation:
  - Create only if CWD is the project root AND directory is writable. Otherwise, MUST log required action and proceed.
  - BOOT LOG entries MUST be UTC and include PID/process/agent ID for deterministic sorting.

### 9) ANALYSIS MODE PERSISTENCE
- Findings apply to the current session only by default.
- To persist: MUST use cache_store to reasoning or vault tiers WITH timestamp, evidence summary, and verification state (hypothesis/validated/provisional). No silent persistence is allowed.

### 10) MODE SELECTOR HEURISTICS
- If prompt/content includes errors/stack traces/failing tests → DEBUG MODE.
- If user intent is to implement/refactor/test multiple files → ORCHESTRATION MODE.
- If user requests ideas/tradeoffs → EXPLORATION MODE.
- If user supplies logs/exports for analysis → ANALYSIS MODE.
- If uncertain: Ask ONE multiple-choice question (A=Debug, B=Orchestration, C=Exploration, D=Analysis), then proceed with user’s choice.

### 11) DEGRADED MODE (PLAYBOOK)
- Conditions: MCP stdio unavailable AND HTTP sidecar unavailable.
- Behavior:
  - MUST NOT perform any network cache writes.
  - MAY proceed with local filesystem reads/writes under WRITE_PROJECT rules.
  - MUST log a Degraded Mode banner to HANDOFF and session status.
  - MUST re-attempt MCP mount at the next phase boundary or after 10 minutes, whichever comes first.

### 12) CANONICAL STATUS KEY & LOG ROTATION
- Project Status Key: MUST use {ProjectName}:status as the canonical status key in SUPERCACHE.
- Log Rotation: Size-based rotation at 1MB, retain last 5, older logs gzip-compressed, suffix -YYYYMMDD-HHMMSS.

### 13) STRUCTURED LOGGING FOR ENFORCEMENT
- The agent MUST emit structured log lines for enforcement events (STOP, circuit-breaker, rate-limit wait, retry exhaustion):
  - Fields: event, tool, error_hash, attempts, time_window, decision, next_step, timestamp_utc.
- Terminal Shadow MUST capture these lines; HANDOFF Decision Log SHOULD include a summary when a circuit breaker is engaged.

---

## Appendix — SUPERCACHE Policy Keys (Recommended)
- global:system:agentic_fetch_policy → { allowed: false, updated_at: ISO8601 }
- global:system:retry_policy → { default_retries: 2, rate_limit_retries: 3, max_backoff_seconds: 60, jitter: true }
- global:system:rate_limits → { provider_defaults: { zai: { rpm:60, burst:10 }, openai: { rpm:60, burst:10 } }, override_allowed: true }
- global:system:enforcement_precedence → { order: ["tool_hook_safety","bans","debug_gates","rate_limits","supercache_access","bias_for_action"] }
- global:system:keys_authority → { project_registry: "global_first", directive_llm_optimization: "global_first" }

---

## Operational Impact
- Determinism: Single-source project identity, canonical SUPERCACHE channel, strict precedence, and circuit breakers eliminate divergent flows.
- Safety: STOP dominance and retry/rate-limit budgets prevent runaway loops and provider lockouts.
- Efficiency: Bias-for-Action remains, but only within safe, deterministic boundaries.
- Auditability: Structured logs + HANDOFF entries provide verifiable receipts for enforcement.
