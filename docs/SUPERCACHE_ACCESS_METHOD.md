# SUPERCACHE Access Method (Canonical)

Last updated: 2026-02-27 09:51:26 UTC

Purpose: Eliminate confusion between MCP stdio tools and optional HTTP sidecar for SUPERCACHE.

Authoritative method:
- Use MCP stdio tools exposed by floyd-supercache-server:
  - cache_retrieve
  - cache_store
  - cache_delete
  - cache_list / cache_search / cache_stats
- Do NOT use HTTP routes (/supercache/*) for core operations. They are optional diagnostics only. The only allowed HTTP check is GET /health to confirm the sidecar is alive.

Authority preference when multiple entries exist:
- Prefer global keys when both global and project-tier stubs exist.
  - global:system:project_registry (authoritative registry)
  - global:system:directive_llm_optimization (full directive v2.5)
- Treat project-tier duplicates with empty/sparse values as stubs.

Shadow daemon boot integration (for continuity):
- Ensure floyd-shadowd is running per directive v2.5:
  - Start: ~/.local/bin/floyd-shadowd start -p {project_path} -i 3
  - Verify: ~/.local/bin/floyd-shadowd status -p {project_path}
- Log boot event to HANDOFF.md under "BOOT LOG" with UTC timestamp and PID.

Rationale:
- MCP stdio is the stable, universal channel used in Floyd CLI environments.
- HTTP sidecar may vary by host and is not guaranteed to expose /supercache/* routes.

References:
- FLOYD.md → I. CORE INITIALIZATION → SUPERCACHE ACCESS RULE (MANDATORY)
- ~/.floyd/supercache/index.json (underlying storage)
- /Volumes/Storage/MCP/floyd-supercache-server/src/index.ts (stdio)
- /Volumes/Storage/MCP/floyd-supercache-server/src/http-server.ts (optional HTTP)
