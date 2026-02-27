# AGENTIC_FETCH BAN (MANDATORY)

Effective: 2026-02-27

Summary
- The `agentic_fetch` tool is banned due to session hang/freeze issues.
- Use alternatives:
  - `fetch` for raw content retrieval
  - `sourcegraph` for public code search and symbol queries
  - `web-search-prime` for general web search tasks

Scope
- Applies to all Floyd instances operating from this repository.
- Remains in effect until explicitly revoked by an update to SUPERCACHE and this document.

Rationale
- Multiple incidents recorded in SUPERCACHE (floyddesktopweb:known_issues:agentic_fetch and global bans) indicate reliability problems causing agent stalls.

Enforcement
- FLOYD.md (Tool/Hook Safety) includes a mandatory clause prohibiting `agentic_fetch`.
- Code reviews should flag any PRs introducing `agentic_fetch`.

References
- FLOYD.md → VII. TOOL / HOOK SAFETY (MANDATORY) → AGENTIC_FETCH BAN (MANDATORY)
- SUPERCACHE keys:
  - floyddesktopweb:known_issues:agentic_fetch
  - reasoning:floyddesktopweb:agentic_fetch_ban
  - floyd:global:known_issues:agentic_fetch
