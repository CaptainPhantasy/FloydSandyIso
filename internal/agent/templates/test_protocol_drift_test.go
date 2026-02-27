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
}
