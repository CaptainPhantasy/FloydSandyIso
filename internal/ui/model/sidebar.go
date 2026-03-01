package model

import (
	"cmp"
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/common"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/logo"
	uv "github.com/charmbracelet/ultraviolet"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// modelInfo renders the current model information including reasoning
// settings and context usage/cost for the sidebar.
func (m *UI) modelInfo(width int) string {
	model := m.selectedLargeModel()
	reasoningInfo := ""
	providerName := ""

	if model != nil {
		// Get provider name first
		providerConfig, ok := m.com.Config().Providers.Get(model.ModelCfg.Provider)
		if ok {
			providerName = providerConfig.Name

			// Only check reasoning if model can reason
			if model.CatwalkCfg.CanReason {
				if model.ModelCfg.ReasoningEffort == "" {
					if model.ModelCfg.Think {
						reasoningInfo = "Thinking On"
					} else {
						reasoningInfo = "Thinking Off"
					}
				} else {
					formatter := cases.Title(language.English, cases.NoLower)
					reasoningEffort := cmp.Or(model.ModelCfg.ReasoningEffort, model.CatwalkCfg.DefaultReasoningEffort)
					reasoningInfo = formatter.String(fmt.Sprintf("Reasoning %s", reasoningEffort))
				}
			}
		}
	}

	var modelContext *common.ModelContextInfo
	var modelName string
	if model != nil {
		modelName = model.CatwalkCfg.Name
		if m.session != nil {
			// Use override context window if set
			var contextWindow int64
			if model.ModelCfg.ContextWindow > 0 {
				contextWindow = model.ModelCfg.ContextWindow
			} else {
				contextWindow = int64(model.CatwalkCfg.ContextWindow)
			}
			contextUsed := m.session.CompletionTokens + m.session.PromptTokens
			modelContext = &common.ModelContextInfo{
				ContextUsed:  contextUsed,
				Cost:         m.session.Cost,
				ModelContext: contextWindow,
			}
		}
	}
	return common.ModelInfo(m.com.Styles, modelName, providerName, reasoningInfo, modelContext, width)
}

// getDynamicHeightLimits will give us the num of items to show in each section based on the height
// some items are more important than others.
// MCP section removed - tools are now auto-discovered at agent boot.
func getDynamicHeightLimits(availableHeight int) (maxFiles, maxLSPs int) {
	const (
		minItemsPerSection      = 2
		defaultMaxFilesShown    = 12
		defaultMaxLSPsShown     = 10
		minAvailableHeightLimit = 10
	)

	// If we have very little space, use minimum values
	if availableHeight < minAvailableHeightLimit {
		return minItemsPerSection, minItemsPerSection
	}

	// Distribute available height between files and LSPs
	// Give priority to files, then LSPs
	totalSections := 2
	heightPerSection := availableHeight / totalSections

	// Calculate limits for each section, ensuring minimums
	maxFiles = max(minItemsPerSection, min(defaultMaxFilesShown, heightPerSection))
	maxLSPs = max(minItemsPerSection, min(defaultMaxLSPsShown, heightPerSection))

	// If we have extra space, give it to files first
	remainingHeight := availableHeight - (maxFiles + maxLSPs)
	if remainingHeight > 0 {
		extraForFiles := min(remainingHeight, defaultMaxFilesShown-maxFiles)
		maxFiles += extraForFiles
		remainingHeight -= extraForFiles

		if remainingHeight > 0 {
			maxLSPs += min(remainingHeight, defaultMaxLSPsShown-maxLSPs)
		}
	}

	return maxFiles, maxLSPs
}

// sidebar renders the chat sidebar containing session title, working
// directory, model info, file list, LSP status, and MCP status.
func (m *UI) drawSidebar(scr uv.Screen, area uv.Rectangle) {
	if m.session == nil {
		return
	}

	const logoHeightBreakpoint = 30
	const logoWidthBreakpoint = 80

	t := m.com.Styles
	width := area.Dx()
	height := area.Dy()

	title := t.Muted.Width(width).MaxHeight(2).Render(m.session.Title)
	cwd := common.PrettyPath(t, m.com.Config().WorkingDir(), width)
	sidebarLogo := m.sidebarLogo
	if height < logoHeightBreakpoint || width < logoWidthBreakpoint {
		sidebarLogo = logo.SidebarRender(m.com.Styles, width, logo.Opts{
			FieldColor:  t.LogoFieldColor,
			TitleColorA: t.LogoTitleColorA,
			TitleColorB: t.LogoTitleColorB,
			CharmColor:  t.LogoCharmColor,
		})
	}
	blocks := []string{
		sidebarLogo,
		title,
		"",
		cwd,
		"",
		m.modelInfo(width),
		"",
	}

	sidebarHeader := lipgloss.JoinVertical(
		lipgloss.Left,
		blocks...,
	)

	_, remainingHeightArea := uv.SplitVertical(m.layout.sidebar, uv.Fixed(lipgloss.Height(sidebarHeader)))
	remainingHeight := remainingHeightArea.Dy() - 10
	maxFiles, maxLSPs := getDynamicHeightLimits(remainingHeight)

	lspSection := m.lspInfo(width, maxLSPs, true)
	filesSection := m.filesInfo(m.com.Config().WorkingDir(), width, maxFiles, true)

	uv.NewStyledString(
		lipgloss.NewStyle().
			MaxWidth(width).
			MaxHeight(height).
			Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					sidebarHeader,
					filesSection,
					"",
					lspSection,
				),
			),
	).Draw(scr, area)
}
