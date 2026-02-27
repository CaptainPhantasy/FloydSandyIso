package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

// Minimal MCP stdio client that sends tools/list and tools/call requests to a host
// providing floyd-supercache cache_store. This avoids HTTP and follows the deterministic protocol.
//
// Usage: go run ./cmd/seeder
// Expects to run in an environment where stdin/stdout are connected to an MCP host.

// JSON-RPC 2.0 message
type rpcReq struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type toolsListResult struct {
	Tools []struct {
		Name string `json:"name"`
	} `json:"tools"`
}

type rpcResp struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type cacheStoreArgs struct {
	Namespace string      `json:"namespace"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Tier      string      `json:"tier,omitempty"`
	Tags      []string    `json:"tags,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// writeRPC writes a single JSON-RPC request with a newline delimiter (SSE-like framing not required here)
func writeRPC(w io.Writer, req rpcReq) error {
	enc := json.NewEncoder(w)
	if err := enc.Encode(&req); err != nil {
		return err
	}
	// ensure flush
	if f, ok := w.(interface{ Flush() error }); ok {
		_ = f.Flush()
	}
	return nil
}

// readResp scans lines until it finds a JSON-RPC response with the matching id
func readResp(r *bufio.Reader, id int, timeout time.Duration) (*rpcResp, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return nil, err
		}
		if len(line) == 0 {
			continue
		}
		var resp rpcResp
		if err := json.Unmarshal(line, &resp); err != nil {
			continue // ignore non-JSON lines
		}
		if resp.ID == id {
			return &resp, nil
		}
	}
	return nil, fmt.Errorf("timeout waiting for response id=%d", id)
}

func mustOK(resp *rpcResp) error {
	if resp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	return nil
}

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)

	// 1) tools/list
	listID := 1
	if err := writeRPC(out, rpcReq{JSONRPC: "2.0", ID: listID, Method: "tools/list", Params: map[string]any{}}); err != nil {
		fmt.Fprintln(os.Stderr, "write tools/list:", err)
		os.Exit(1)
	}
	listResp, err := readResp(in, listID, 3*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read tools/list:", err)
		os.Exit(1)
	}
	if err := mustOK(listResp); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var tl toolsListResult
	if err := json.Unmarshal(listResp.Result, &tl); err != nil {
		fmt.Fprintln(os.Stderr, "decode tools/list:", err)
		os.Exit(1)
	}
	found := false
	for _, t := range tl.Tools {
		if t.Name == "cache_store" {
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintln(os.Stderr, "cache_store tool not found on MCP stdio; ensure floyd-supercache-server is wired")
		os.Exit(1)
	}

	utc := os.Getenv("SEED_UPDATED_AT")
	if utc == "" {
		utc = time.Now().UTC().Format(time.RFC3339)
	}

	payloads := []cacheStoreArgs{
		{Namespace: "global", Key: "system:agentic_fetch_policy", Value: map[string]any{"allowed": false, "updated_at": utc}, Tier: "project", Tags: []string{"policy", "deterministic_protocol"}},
		{Namespace: "global", Key: "system:retry_policy", Value: map[string]any{"default_retries": 2, "rate_limit_retries": 3, "max_backoff_seconds": 60, "jitter": true}, Tier: "project", Tags: []string{"policy", "deterministic_protocol"}},
		{Namespace: "global", Key: "system:rate_limits", Value: map[string]any{"provider_defaults": map[string]any{"zai": map[string]int{"rpm": 60, "burst": 10}, "openai": map[string]int{"rpm": 60, "burst": 10}}, "override_allowed": true}, Tier: "project", Tags: []string{"policy", "deterministic_protocol"}},
		{Namespace: "global", Key: "system:enforcement_precedence", Value: map[string]any{"order": []string{"tool_hook_safety", "bans", "debug_gates", "rate_limits", "supercache_access", "bias_for_action"}}, Tier: "project", Tags: []string{"policy", "deterministic_protocol"}},
		{Namespace: "global", Key: "system:keys_authority", Value: map[string]any{"project_registry": "global_first", "directive_llm_optimization": "global_first"}, Tier: "project", Tags: []string{"policy", "deterministic_protocol"}},
	}

	seeded := make([]map[string]any, 0, len(payloads))
	for i, p := range payloads {
		id := 100 + i
		req := rpcReq{
			JSONRPC: "2.0",
			ID:      id,
			Method:  "tools/call",
			Params: map[string]any{
				"name":     "cache_store",
				"arguments": p,
			},
		}
		if err := writeRPC(out, req); err != nil {
			fmt.Fprintln(os.Stderr, "write tools/call cache_store:", err)
			os.Exit(1)
		}
		resp, err := readResp(in, id, 5*time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, "read tools/call cache_store:", err)
			os.Exit(1)
		}
		if err := mustOK(resp); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		seeded = append(seeded, map[string]any{"key": p.Key, "ok": true})
	}

	// Final receipt
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{"seeded": seeded})
}
