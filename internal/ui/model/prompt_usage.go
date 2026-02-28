package model

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/config"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/message"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

func (m *UI) loadPromptUsage() tea.Cmd {
	return func() tea.Msg {
		messages, err := m.com.App.Messages.ListAllUserMessages(context.Background())
		if err != nil {
			// Log error but preserve existing data - don't wipe it
			slog.Error("Failed to load prompt usage", "error", err)
			return promptUsageLoadedMsg{userMessages: m.localUserMessages}
		}

		userMessages := make(map[string]int64, len(messages))
		for _, msg := range messages {
			if msg.Role != message.User {
				continue
			}
			userMessages[msg.ID] = msg.CreatedAt
		}
		return promptUsageLoadedMsg{userMessages: userMessages}
	}
}

// startPromptUsageRefreshTicker starts a periodic ticker to refresh prompt usage.
func (m *UI) startPromptUsageRefreshTicker() tea.Cmd {
	return tea.Tick(promptUsageRefreshInterval, func(time.Time) tea.Msg {
		return promptUsageRefreshTickMsg{}
	})
}

func (m *UI) localPromptUsageCounts(now time.Time) (fiveHour, sevenDay int) {
	// Database stores timestamps in SECONDS (using strftime('%s', 'now'))
	// so use Unix() not UnixMilli() for comparison
	sevenDayCutoff := now.Add(-7 * 24 * time.Hour).Unix()
	fiveHourCutoff := now.Add(-5 * time.Hour).Unix()

	for id, ts := range m.localUserMessages {
		if ts < sevenDayCutoff {
			delete(m.localUserMessages, id)
			continue
		}
		sevenDay++
		if ts >= fiveHourCutoff {
			fiveHour++
		}
	}

	return fiveHour, sevenDay
}

func (m *UI) promptQuotaInfo(width int) string {
	t := m.com.Styles

	// Get config values
	quotaCfg := m.promptQuotaConfig()

	// If quota display is disabled, return empty string
	if quotaCfg.Enabled == nil || !*quotaCfg.Enabled {
		return ""
	}

	fiveHourCount, sevenDayCount := m.localPromptUsageCounts(time.Now())

	fiveHourPct := usagePercent(fiveHourCount, quotaCfg.Limit5h)
	sevenDayPct := usagePercent(sevenDayCount, quotaCfg.Limit7d)

	fiveHourLine := fmt.Sprintf("● 5h prompts %d/%d (%d%%)", fiveHourCount, quotaCfg.Limit5h, fiveHourPct)
	sevenDayLine := fmt.Sprintf("● 7d prompts %d/%d (%d%%)", sevenDayCount, quotaCfg.Limit7d, sevenDayPct)

	fiveHourLine = usageSeverityStyle(t, fiveHourPct, quotaCfg).Render(fiveHourLine)
	sevenDayLine = usageSeverityStyle(t, sevenDayPct, quotaCfg).Render(sevenDayLine)

	note := t.Muted.Render("Local Floyd usage only")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		t.Subtle.Render("Plan Quota (est.)"),
		fiveHourLine,
		sevenDayLine,
		note,
	)

	return lipgloss.NewStyle().Width(width).PaddingLeft(2).Render(content)
}

// promptQuotaConfig returns the prompt quota configuration with defaults.
func (m *UI) promptQuotaConfig() config.PromptQuotaConfig {
	if m.com == nil || m.com.Config() == nil {
		return config.DefaultPromptQuotaConfig()
	}
	return m.com.Config().PromptQuota()
}

func usagePercent(current, limit int) int {
	if limit <= 0 {
		return 0
	}
	pct := int((float64(current) / float64(limit)) * 100)
	if pct < 0 {
		return 0
	}
	return pct
}

func usageSeverityStyle(t *styles.Styles, percent int, cfg config.PromptQuotaConfig) lipgloss.Style {
	switch {
	case percent >= cfg.DangerPercent:
		return t.Base.Foreground(t.Red)
	case percent >= cfg.WarnPercent:
		return t.Base.Foreground(t.Yellow)
	default:
		return t.Base.Foreground(t.Green)
	}
}
