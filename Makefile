# Minimal Makefile for deterministic protocol ops

GO ?= go

seed-policies:
	@echo "NOTE: This seeder prints JSON payloads to apply via MCP stdio cache_store."
	@$(GO) run ./cmd/seeder

check-protocol:
	@$(GO) test ./internal/agent/templates -run TestProtocolTemplatesDrift -v

