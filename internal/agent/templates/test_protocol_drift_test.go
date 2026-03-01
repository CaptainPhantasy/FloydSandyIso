package templates_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// criticalBlocks defines anchor phrases that MUST appear identically
// in both the root FLOYD.md and the system prompt templates.
var criticalBlocks = []string{
	"Boot Summary (MUST be 4 lines exactly):",
	"SUPERCACHE ACCESS (CANONICAL)",
	"STOP Rule (Precedence over Bias-for-Action)",
	"Banned Tools & Revocation (agentic_fetch)",
	"A) Hypothesis Gate (NO FIX WITHOUT THIS)",
	"C) Two-Failure Reset Rule",
	"E) Prediction Rule",
	"F) Error Repetition Circuit Breaker",
	"XI. ADVANCED TOOL TRIGGERS (MANDATORY)",
	"DISCOVERY GATE (MANDATORY BEFORE ACTION)",
	"DEGRADED MODE PLAYBOOK",
	"SHADOW DAEMON & HANDOFF PROTOCOL",
}

func mustRead(t *testing.T, p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

func TestProtocolTemplatesDrift(t *testing.T) {
	root := filepath.Join("..", "..", "..", "FLOYD.md")
	protoTpl := filepath.Join("..", "..", "agent", "templates", "floyd_protocol.md.tpl")
	coderTpl := filepath.Join("..", "..", "agent", "templates", "coder.md.tpl")

	rootS := mustRead(t, root)
	protoS := mustRead(t, protoTpl)
	coderS := mustRead(t, coderTpl)

	for _, anchor := range criticalBlocks {
		if !strings.Contains(rootS, anchor) {
			t.Fatalf("root FLOYD.md missing anchor: %q", anchor)
		}
		if !strings.Contains(protoS, anchor) {
			t.Fatalf("floyd_protocol.md.tpl missing anchor: %q", anchor)
		}
		if !strings.Contains(coderS, anchor) {
			t.Fatalf("coder.md.tpl missing anchor: %q", anchor)
		}
	}

	mustContain := []string{
		"I am FLOYD v4.6.1",
		"HTTP /supercache/* MUST NOT be used",
	}
	for _, m := range mustContain {
		if !strings.Contains(rootS, m) {
			t.Fatalf("root FLOYD.md missing required text: %q", m)
		}
		if !strings.Contains(protoS, m) {
			t.Fatalf("floyd_protocol.md.tpl missing required text: %q", m)
		}
	}

	mustNotContain := []string{
		"v4.0.0",
		"FORCE MCP DISCOVERY",
	}
	for _, bad := range mustNotContain {
		if strings.Contains(rootS, bad) {
			t.Fatalf("root FLOYD.md contains forbidden text: %q", bad)
		}
		if strings.Contains(protoS, bad) {
			t.Fatalf("floyd_protocol.md.tpl contains forbidden text: %q", bad)
		}
		if strings.Contains(coderS, bad) {
			t.Fatalf("coder.md.tpl contains forbidden text: %q", bad)
		}
	}
}
