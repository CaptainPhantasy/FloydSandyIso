package common

import (
	"cmp"
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/home"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
)

// PrettyPath formats a file path with home directory shortening and applies
// muted styling.
func PrettyPath(t *styles.Styles, path string, width int) string {
	formatted := home.Short(path)
	return t.Muted.Width(width).Render(formatted)
}

// ModelContextInfo contains token usage and cost information for a model.
type ModelContextInfo struct {
	TotalTokens     int64   // Raw total (PromptTokens + CompletionTokens)
	EffectiveTokens int64   // After cache subtraction (Total - CacheReadTokens)
	CacheReadTokens int64   // How much was served from cache
	CachePercent    float64 // What percentage was cached
	ModelContext    int64   // Context window size
	Cost            float64
}

// ModelInfo renders model information including name, provider, reasoning
// settings, and optional context usage/cost.
func ModelInfo(t *styles.Styles, modelName, providerName, reasoningInfo string, context *ModelContextInfo, width int) string {
	modelIcon := t.Subtle.Render(styles.ModelIcon)
	modelName = t.Base.Render(modelName)

	// Build first line with model name and optionally provider on the same line
	var firstLine string
	if providerName != "" {
		providerInfo := t.Muted.Render(fmt.Sprintf("via %s", providerName))
		modelWithProvider := fmt.Sprintf("%s %s %s", modelIcon, modelName, providerInfo)

		// Check if it fits on one line
		if lipgloss.Width(modelWithProvider) <= width {
			firstLine = modelWithProvider
		} else {
			// If it doesn't fit, put provider on next line
			firstLine = fmt.Sprintf("%s %s", modelIcon, modelName)
		}
	} else {
		firstLine = fmt.Sprintf("%s %s", modelIcon, modelName)
	}

	parts := []string{firstLine}

	// If provider didn't fit on first line, add it as second line
	if providerName != "" && !strings.Contains(firstLine, "via") {
		providerInfo := fmt.Sprintf("via %s", providerName)
		parts = append(parts, t.Muted.PaddingLeft(2).Render(providerInfo))
	}

	if reasoningInfo != "" {
		parts = append(parts, t.Subtle.PaddingLeft(2).Render(reasoningInfo))
	}

	if context != nil {
		formattedInfo := formatTokensAndCost(t, context)
		parts = append(parts, lipgloss.NewStyle().PaddingLeft(2).Render(formattedInfo))
	}

	return lipgloss.NewStyle().Width(width).Render(
		lipgloss.JoinVertical(lipgloss.Left, parts...),
	)
}

// formatTokensAndCost formats token usage and cost with dual display showing
// both total tokens and effective tokens (after cache subtraction).
// Returns two lines: usage line and cache line (cache line empty if no caching).
func formatTokensAndCost(t *styles.Styles, info *ModelContextInfo) string {
	// Format total tokens
	var formattedTotal string
	switch {
	case info.TotalTokens >= 1_000_000:
		formattedTotal = fmt.Sprintf("%.1fM", float64(info.TotalTokens)/1_000_000)
	case info.TotalTokens >= 1_000:
		formattedTotal = fmt.Sprintf("%.1fK", float64(info.TotalTokens)/1_000)
	default:
		formattedTotal = fmt.Sprintf("%d", info.TotalTokens)
	}
	if strings.HasSuffix(formattedTotal, ".0K") {
		formattedTotal = strings.Replace(formattedTotal, ".0K", "K", 1)
	}
	if strings.HasSuffix(formattedTotal, ".0M") {
		formattedTotal = strings.Replace(formattedTotal, ".0M", "M", 1)
	}

	// Format effective tokens
	var formattedEffective string
	switch {
	case info.EffectiveTokens >= 1_000_000:
		formattedEffective = fmt.Sprintf("%.1fM", float64(info.EffectiveTokens)/1_000_000)
	case info.EffectiveTokens >= 1_000:
		formattedEffective = fmt.Sprintf("%.1fK", float64(info.EffectiveTokens)/1_000)
	default:
		formattedEffective = fmt.Sprintf("%d", info.EffectiveTokens)
	}
	if strings.HasSuffix(formattedEffective, ".0K") {
		formattedEffective = strings.Replace(formattedEffective, ".0K", "K", 1)
	}
	if strings.HasSuffix(formattedEffective, ".0M") {
		formattedEffective = strings.Replace(formattedEffective, ".0M", "M", 1)
	}

	// Calculate percentage based on TOTAL (conservative warning)
	totalPct := (float64(info.TotalTokens) / float64(info.ModelContext)) * 100

	// Format cost
	formattedCost := t.Muted.Render(fmt.Sprintf("$%.2f", info.Cost))

	// Line 1: Usage - "42% (85K total / 42K effective) $0.12"
	formattedPercentage := t.Muted.Render(fmt.Sprintf("%d%%", int(totalPct)))
	formattedTokens := t.Subtle.Render(fmt.Sprintf("(%s total / %s effective)",
		formattedTotal, formattedEffective))
	usageLine := fmt.Sprintf("%s %s %s", formattedPercentage, formattedTokens, formattedCost)

	// Add warning icon if needed
	if totalPct > 80 {
		usageLine = fmt.Sprintf("%s %s", styles.LSPWarningIcon, usageLine)
	}

	// Line 2: Cache info (only if there's caching) - "50% cached"
	var cacheLine string
	if info.CachePercent > 0 {
		cacheLine = t.Subtle.Render(fmt.Sprintf("   %.0f%% cached", info.CachePercent))
	}

	// Combine lines
	if cacheLine != "" {
		return usageLine + "\n" + cacheLine
	}
	return usageLine
}

// StatusOpts defines options for rendering a status line with icon, title,
// description, and optional extra content.
type StatusOpts struct {
	Icon             string // if empty no icon will be shown
	Title            string
	TitleColor       color.Color
	Description      string
	DescriptionColor color.Color
	ExtraContent     string // additional content to append after the description
}

// Status renders a status line with icon, title, description, and extra
// content. The description is truncated if it exceeds the available width.
func Status(t *styles.Styles, opts StatusOpts, width int) string {
	icon := opts.Icon
	title := opts.Title
	description := opts.Description

	titleColor := cmp.Or(opts.TitleColor, t.Muted.GetForeground())
	descriptionColor := cmp.Or(opts.DescriptionColor, t.Subtle.GetForeground())

	title = t.Base.Foreground(titleColor).Render(title)

	if description != "" {
		extraContentWidth := lipgloss.Width(opts.ExtraContent)
		if extraContentWidth > 0 {
			extraContentWidth += 1
		}
		description = ansi.Truncate(description, width-lipgloss.Width(icon)-lipgloss.Width(title)-2-extraContentWidth, "…")
		description = t.Base.Foreground(descriptionColor).Render(description)
	}

	content := []string{}
	if icon != "" {
		content = append(content, icon)
	}
	content = append(content, title)
	if description != "" {
		content = append(content, description)
	}
	if opts.ExtraContent != "" {
		content = append(content, opts.ExtraContent)
	}

	return strings.Join(content, " ")
}

// Section renders a section header with a title and a horizontal line filling
// the remaining width.
func Section(t *styles.Styles, text string, width int, info ...string) string {
	char := styles.SectionSeparator
	length := lipgloss.Width(text) + 1
	remainingWidth := width - length

	var infoText string
	if len(info) > 0 {
		infoText = strings.Join(info, " ")
		if len(infoText) > 0 {
			infoText = " " + infoText
			remainingWidth -= lipgloss.Width(infoText)
		}
	}

	text = t.Section.Title.Render(text)
	if remainingWidth > 0 {
		text = text + " " + t.Section.Line.Render(strings.Repeat(char, remainingWidth)) + infoText
	}
	return text
}

// DialogTitle renders a dialog title with a decorative line filling the
// remaining width.
func DialogTitle(t *styles.Styles, title string, width int, fromColor, toColor color.Color) string {
	char := "╱"
	length := lipgloss.Width(title) + 1
	remainingWidth := width - length
	if remainingWidth > 0 {
		lines := strings.Repeat(char, remainingWidth)
		lines = styles.ApplyForegroundGrad(t, lines, fromColor, toColor)
		title = title + " " + lines
	}
	return title
}
