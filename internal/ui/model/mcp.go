package model

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/agent/tools"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/agent/tools/mcp"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/environment"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/common"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

// mcpInfo renders the MCP status section showing active MCP clients and their
// tool/prompt counts, plus a boot summary showing total tool count and environment.
func (m *UI) mcpInfo(width, maxItems int, isSection bool) string {
	var mcps []mcp.ClientInfo
	t := m.com.Styles

	for _, mcpCfg := range m.com.Config().MCP.Sorted() {
		if state, ok := m.mcpStates[mcpCfg.Name]; ok {
			mcps = append(mcps, state)
		}
	}

	title := t.Subtle.Render("MCPs")
	if isSection {
		title = common.Section(t, title, width)
	}

	// Build boot summary with total tool count and environment
	bootSummary := buildBootSummary(t, m.com.Config())

	list := t.Subtle.Render("None")
	if len(mcps) > 0 {
		list = mcpList(t, mcps, width, maxItems)
	}

	content := bootSummary + "\n\n" + list
	return lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("%s\n\n%s", title, content))
}

// buildBootSummary returns a summary line showing total tools and environment state.
func buildBootSummary(t *styles.Styles, cfg any) string {
	// Get tool counts from registry
	totalTools := tools.ToolCount()
	totalServers := tools.ServerCount()

	if totalTools == 0 && totalServers == 0 {
		return t.Subtle.Render("Initializing tool registry...")
	}

	// Build summary line
	summary := fmt.Sprintf("%d tools from %d servers", totalTools, totalServers)

	// Add environment component count if we can discover it
	// Note: We're using a simplified approach here - full discovery happens at agent boot
	components := []string{}

	// Quick check for known components in common locations
	if exists("/Volumes/Storage/FloydDesktopWeb-v2") {
		components = append(components, "desktop")
	}
	if exists("/Volumes/Storage/floyd-harness") {
		components = append(components, "harness")
	}
	if exists("/Volumes/Storage/MCP") {
		components = append(components, "mcp")
	}
	if exists("/Volumes/Storage/floyd-sandbox/FloydDeployable") {
		components = append(components, "cli")
	}

	if len(components) > 0 {
		summary += fmt.Sprintf(" | %d components", len(components)+2) // +2 for MCP and main CLI
	}

	// Add version info if available
	// TODO: Import version package without causing import cycle
	// summary += fmt.Sprintf(" | %s", "v4.0.0")

	return t.Subtle.Render(summary)
}

// exists quickly checks if a path exists without full environment discovery
func exists(path string) bool {
	return environment.PathExists(path)
}

// mcpCounts formats tool and prompt counts for display.
func mcpCounts(t *styles.Styles, counts mcp.Counts) string {
	parts := []string{}
	if counts.Tools > 0 {
		parts = append(parts, t.Subtle.Render(fmt.Sprintf("%d tools", counts.Tools)))
	}
	if counts.Prompts > 0 {
		parts = append(parts, t.Subtle.Render(fmt.Sprintf("%d prompts", counts.Prompts)))
	}
	return strings.Join(parts, " ")
}

// mcpList renders a list of MCP clients with their status and counts,
// truncating to maxItems if needed.
func mcpList(t *styles.Styles, mcps []mcp.ClientInfo, width, maxItems int) string {
	if maxItems <= 0 {
		return ""
	}
	var renderedMcps []string

	for _, m := range mcps {
		var icon string
		title := m.Name
		var description string
		var extraContent string

		switch m.State {
		case mcp.StateStarting:
			icon = t.ItemBusyIcon.String()
			description = t.Subtle.Render("starting...")
		case mcp.StateConnected:
			icon = t.ItemOnlineIcon.String()
			extraContent = mcpCounts(t, m.Counts)
		case mcp.StateError:
			icon = t.ItemErrorIcon.String()
			description = t.Subtle.Render("error")
			if m.Error != nil {
				description = t.Subtle.Render(fmt.Sprintf("error: %s", m.Error.Error()))
			}
		case mcp.StateDisabled:
			icon = t.ItemOfflineIcon.Foreground(t.Muted.GetBackground()).String()
			description = t.Subtle.Render("disabled")
		default:
			icon = t.ItemOfflineIcon.String()
		}

		renderedMcps = append(renderedMcps, common.Status(t, common.StatusOpts{
			Icon:         icon,
			Title:        title,
			Description:  description,
			ExtraContent: extraContent,
		}, width))
	}

	if len(renderedMcps) > maxItems {
		visibleItems := renderedMcps[:maxItems-1]
		remaining := len(renderedMcps) - maxItems
		visibleItems = append(visibleItems, t.Subtle.Render(fmt.Sprintf("…and %d more", remaining)))
		return lipgloss.JoinVertical(lipgloss.Left, visibleItems...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, renderedMcps...)
}
