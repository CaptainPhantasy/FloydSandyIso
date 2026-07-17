# Recovered semantic compaction and plugins

This recovery preserves two coherent local feature sets that were absent from the remote `main` branch.

## Semantic compaction

- separates system, recent, historical, tool, and file context before summarization;
- tracks the cumulative number of summarized tokens in the session database;
- retains recent context while compacting older messages;
- uses a 20,000-token remaining buffer for context windows of at least 200,000 tokens and a 20% remaining buffer for smaller windows;
- displays cumulative summarized-token counts in the session header.

The database migration adds `sessions.total_tokens_summarized` with a zero default. The SQL source and generated query bindings both persist the field.

## Plugin discovery

Plugin directories are configured through `options.plugins_paths`, `FLOYD_PLUGINS_DIR`, and the standard global configuration locations. A plugin is a directory containing `PLUGIN.md` with YAML front matter and optional instruction text.

The parser validates metadata bounds, discovers enabled plugins deterministically, and exposes safe prompt metadata for available plugins and their instructions. Focused tests cover parsing, validation, discovery, prompt rendering, slash-command lookup, and sub-agent lookup.

This recovery does not claim that parsed slash commands, connector references, or sub-agent declarations are independently executed by the runtime. They are discovery metadata supplied to the agent prompt.
