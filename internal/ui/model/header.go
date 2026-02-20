package model

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/csync"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/fsext"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/lsp"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/session"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/common"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

const (
	headerDiag     = "╱"
	minHeaderDiags = 3
	leftPadding    = 1
	rightPadding   = 1
)

type header struct {
	// cached logo and compact logo
	logo        string
	compactLogo string

	com     *common.Common
	width   int
	compact bool
}

// newHeader creates a new header model.
func newHeader(com *common.Common) *header {
	h := &header{
		com: com,
	}
	t := com.Styles
	h.compactLogo = t.Header.Charm.Render("LEGACY AI™") + " " +
		styles.ApplyBoldForegroundGrad(t, "FLOYD", t.Secondary, t.Primary) + " "
	return h
}

// drawHeader draws the header for the given session.
func (h *header) drawHeader(
	scr uv.Screen,
	area uv.Rectangle,
	session *session.Session,
	compact bool,
	detailsOpen bool,
	width int,
) {
	t := h.com.Styles
	if width != h.width || compact != h.compact {
		h.logo = renderLogo(h.com.Styles, compact, width)
	}

	h.width = width
	h.compact = compact

	if !compact || session == nil || h.com.App == nil {
		uv.NewStyledString(h.logo).Draw(scr, area)
		return
	}

	if session.ID == "" {
		return
	}

	var b strings.Builder
	b.WriteString(h.compactLogo)

	availDetailWidth := width - leftPadding - rightPadding - lipgloss.Width(b.String()) - minHeaderDiags
	details := renderHeaderDetails(h.com, session, h.com.App.LSPClients, detailsOpen, availDetailWidth)

	remainingWidth := width -
		lipgloss.Width(b.String()) -
		lipgloss.Width(details) -
		leftPadding -
		rightPadding

	if remainingWidth > 0 {
		b.WriteString(t.Header.Diagonals.Render(
			strings.Repeat(headerDiag, max(minHeaderDiags, remainingWidth)),
		))
		b.WriteString(" ")
	}

	b.WriteString(details)

	view := uv.NewStyledString(
		t.Base.Padding(0, rightPadding, 0, leftPadding).Render(b.String()))
	view.Draw(scr, area)
}

// renderHeaderDetails renders the details section of the header.
func renderHeaderDetails(
	com *common.Common,
	session *session.Session,
	lspClients *csync.Map[string, *lsp.Client],
	detailsOpen bool,
	availWidth int,
) string {
	t := com.Styles

	var parts []string

	errorCount := 0
	for l := range lspClients.Seq() {
		errorCount += l.GetDiagnosticCounts().Error
	}

	if errorCount > 0 {
		parts = append(parts, t.LSP.ErrorDiagnostic.Render(fmt.Sprintf("%s%d", styles.LSPErrorIcon, errorCount)))
	}

	agentCfg := config.Get().Agents[config.AgentCoder]
	contextWindow := config.Get().GetModelContextWindow(agentCfg.Model)
	var percentage float64
	if contextWindow > 0 {
		// Subtract cache read tokens - they don't consume fresh context
		contextUsed := session.CompletionTokens + session.PromptTokens - session.CacheReadTokens
		percentage = (float64(contextUsed) / float64(contextWindow)) * 100
	}
	formattedPercentage := t.Header.Percentage.Render(fmt.Sprintf("%d%%", int(percentage)))
	parts = append(parts, formattedPercentage)

	const keystroke = "ctrl+d"
	if detailsOpen {
		parts = append(parts, t.Header.Keystroke.Render(keystroke)+t.Header.KeystrokeTip.Render(" close"))
	} else {
		parts = append(parts, t.Header.Keystroke.Render(keystroke)+t.Header.KeystrokeTip.Render(" open "))
	}

	dot := t.Header.Separator.Render(" • ")
	metadata := strings.Join(parts, dot)
	metadata = dot + metadata

	const dirTrimLimit = 4
	cfg := com.Config()
	cwd := fsext.DirTrim(fsext.PrettyPath(cfg.WorkingDir()), dirTrimLimit)
	cwd = ansi.Truncate(cwd, max(0, availWidth-lipgloss.Width(metadata)), "…")
	cwd = t.Header.WorkingDir.Render(cwd)

	return cwd + metadata
}
