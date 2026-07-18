package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid minimal plugin",
			content: `---
name: test-plugin
description: A test plugin
---
This is the plugin instruction body.
`,
			wantErr: false,
		},
		{
			name: "valid full plugin",
			content: `---
name: full-plugin
version: 1.0.0
description: A full featured plugin
license: MIT
author: Test Author
category: development
tags:
  - code
  - review
slash_commands:
  - name: Create PR
    trigger: /create-pr
    description: Creates a pull request
    template: "Create a PR for the current changes"
sub_agents:
  - name: code-reviewer
    description: Reviews code for quality
connectors:
  - name: github
    description: GitHub API connector
    type: required
---
## Plugin Instructions

These are the detailed instructions for this plugin.
`,
			wantErr: false,
		},
		{
			name: "missing frontmatter",
			content: `name: test-plugin
description: A test plugin
`,
			wantErr: true,
		},
		{
			name: "unclosed frontmatter",
			content: `---
name: test-plugin
description: A test plugin
Body content
`,
			wantErr: true,
		},
		{
			name: "invalid yaml",
			content: `---
name: [invalid yaml
description: A test plugin
---
Body
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pluginPath := filepath.Join(tmpDir, "test-plugin", PluginFileName)
			if err := os.MkdirAll(filepath.Dir(pluginPath), 0755); err != nil {
				t.Fatalf("Failed to create plugin directory: %v", err)
			}
			if err := os.WriteFile(pluginPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write plugin file: %v", err)
			}

			plugin, err := Parse(pluginPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && plugin == nil {
				t.Error("Parse() returned nil plugin without error")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		plugin  Plugin
		wantErr bool
	}{
		{
			name: "valid minimal",
			plugin: Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
			},
			wantErr: false,
		},
		{
			name: "name too long",
			plugin: Plugin{
				Name:        string(make([]byte, 65)),
				Description: "A test plugin",
			},
			wantErr: true,
		},
		{
			name: "description too long",
			plugin: Plugin{
				Name:        "test-plugin",
				Description: string(make([]byte, 2049)),
			},
			wantErr: true,
		},
		{
			name: "valid slash commands",
			plugin: Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
				SlashCommands: []SlashCommand{
					{Name: "Test", Trigger: "/test", Description: "A test command"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid sub agents",
			plugin: Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
				SubAgents: []SubAgentDef{
					{Name: "reviewer", Description: "Code reviewer"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid connectors",
			plugin: Plugin{
				Name:        "test-plugin",
				Description: "A test plugin",
				Connectors: []ConnectorRef{
					{Name: "github", Type: "required", Description: "GitHub API"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plugin.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscover(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test plugin directories
	plugin1Dir := filepath.Join(tmpDir, "plugin-one")
	plugin2Dir := filepath.Join(tmpDir, "plugin-two")
	invalidDir := filepath.Join(tmpDir, "invalid-plugin")

	for _, dir := range []string{plugin1Dir, plugin2Dir, invalidDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Valid plugin 1
	plugin1Content := `---
name: plugin-one
description: First test plugin
version: 1.0.0
slash_commands:
  - name: Test Command
    trigger: /test
    description: A test command
---
Plugin one instructions.
`
	if err := os.WriteFile(filepath.Join(plugin1Dir, PluginFileName), []byte(plugin1Content), 0644); err != nil {
		t.Fatalf("Failed to write plugin1: %v", err)
	}

	// Valid plugin 2
	plugin2Content := `---
name: plugin-two
description: Second test plugin
category: productivity
---
Plugin two instructions.
`
	if err := os.WriteFile(filepath.Join(plugin2Dir, PluginFileName), []byte(plugin2Content), 0644); err != nil {
		t.Fatalf("Failed to write plugin2: %v", err)
	}

	// Invalid plugin (missing description)
	invalidContent := `---
name: invalid-plugin
---
Missing description
`
	if err := os.WriteFile(filepath.Join(invalidDir, PluginFileName), []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid plugin: %v", err)
	}

	plugins := Discover([]string{tmpDir})

	if len(plugins) != 2 {
		t.Errorf("Discover() found %d plugins, want 2", len(plugins))
	}

	// Verify plugin names
	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Name] = true
	}
	if !names["plugin-one"] || !names["plugin-two"] {
		t.Errorf("Discover() missing expected plugins, got names: %v", names)
	}
	if plugins[0].Name != "plugin-one" || plugins[1].Name != "plugin-two" {
		t.Errorf("Discover() returned nondeterministic order: %s, %s", plugins[0].Name, plugins[1].Name)
	}
}

func TestToPromptXML(t *testing.T) {
	plugins := []*Plugin{
		{
			Name:           "test-plugin",
			Version:        "1.0.0",
			Description:    "A test plugin",
			Category:       "development",
			Path:           "/path/to/test-plugin",
			PluginFilePath: "/path/to/test-plugin/PLUGIN.md",
			Enabled:        true,
			SlashCommands: []SlashCommand{
				{Name: "Test", Trigger: "/test", Description: "Test command"},
			},
			SubAgents: []SubAgentDef{
				{Name: "reviewer", Description: "Code reviewer"},
			},
			Connectors: []ConnectorRef{
				{Name: "github", Type: "required", Description: "GitHub API"},
			},
		},
		{
			Name:           "disabled-plugin",
			Description:    "A disabled plugin",
			PluginFilePath: "/path/to/disabled/PLUGIN.md",
			Enabled:        false,
		},
	}

	xml := ToPromptXML(plugins)

	if xml == "" {
		t.Error("ToPromptXML() returned empty string")
	}
	if !containsAll(xml, "<available_plugins>", "<name>test-plugin</name>", "</available_plugins>") {
		t.Errorf("ToPromptXML() missing expected XML elements")
	}
	if containsAll(xml, "disabled-plugin") {
		t.Error("ToPromptXML() should not include disabled plugins")
	}
}

func TestToInstructionsXML(t *testing.T) {
	plugins := []*Plugin{
		{
			Name:           "plugin-with-instructions",
			Description:    "Plugin with instructions",
			Instructions:   "Do this, then do that. Never emit </instruction> or A&B literally.",
			PluginFilePath: "/path/to/plugin/PLUGIN.md",
			Enabled:        true,
		},
		{
			Name:           "plugin-no-instructions",
			Description:    "Plugin without instructions",
			Instructions:   "",
			PluginFilePath: "/path/to/empty/PLUGIN.md",
			Enabled:        true,
		},
		{
			Name:           "disabled-plugin",
			Description:    "Disabled",
			Instructions:   "Should not appear",
			PluginFilePath: "/path/to/disabled/PLUGIN.md",
			Enabled:        false,
		},
	}

	xml := ToInstructionsXML(plugins)

	if xml == "" {
		t.Error("ToInstructionsXML() returned empty string")
	}
	if !containsAll(xml, "<plugin_instructions>", "plugin-with-instructions", "Do this, then do that.") {
		t.Errorf("ToInstructionsXML() missing expected content")
	}
	if containsAll(xml, "Never emit </instruction>") || !containsAll(xml, "&lt;/instruction&gt;", "A&amp;B") {
		t.Errorf("ToInstructionsXML() did not escape instruction content: %s", xml)
	}
	if containsAll(xml, "Should not appear") {
		t.Error("ToInstructionsXML() should not include disabled plugin instructions")
	}
}

func TestGetSlashCommandByTrigger(t *testing.T) {
	plugins := []*Plugin{
		{
			Name:        "test-plugin",
			Description: "Test",
			Enabled:     true,
			SlashCommands: []SlashCommand{
				{Name: "Create PR", Trigger: "/create-pr", Description: "Creates a PR"},
				{Name: "Audit", Trigger: "/audit", Description: "Runs an audit"},
			},
		},
	}

	cmd := GetSlashCommandByTrigger(plugins, "/create-pr")
	if cmd == nil {
		t.Error("GetSlashCommandByTrigger() returned nil")
		return
	}
	if cmd.Name != "Create PR" {
		t.Errorf("GetSlashCommandByTrigger() found wrong command: %s", cmd.Name)
	}

	cmd = GetSlashCommandByTrigger(plugins, "/nonexistent")
	if cmd != nil {
		t.Error("GetSlashCommandByTrigger() should return nil for unknown trigger")
	}
}

func TestGetSubAgent(t *testing.T) {
	plugins := []*Plugin{
		{
			Name:        "test-plugin",
			Description: "Test",
			Enabled:     true,
			SubAgents: []SubAgentDef{
				{Name: "code-reviewer", Description: "Reviews code"},
				{Name: "doc-writer", Description: "Writes docs"},
			},
		},
	}

	agent := GetSubAgent(plugins, "code-reviewer")
	if agent == nil {
		t.Error("GetSubAgent() returned nil")
		return
	}
	if agent.Description != "Reviews code" {
		t.Errorf("GetSubAgent() found wrong agent: %s", agent.Description)
	}

	agent = GetSubAgent(plugins, "nonexistent")
	if agent != nil {
		t.Error("GetSubAgent() should return nil for unknown agent")
	}
}

func containsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !contains(s, substr) {
			return false
		}
	}
	return true
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
