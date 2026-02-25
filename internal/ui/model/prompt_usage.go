package model

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/message"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

const (
	localMaxPlan5HourPromptLimit = 1600
	localMaxPlan7DayPromptLimit  = 8000
	localWarnThresholdPercent    = 75
	localDangerThresholdPercent  = 90
)

func (m *UI) loadPromptUsage() tea.Cmd {
	return func() tea.Msg {
		messages, err := m.com.App.Messages.ListAllUserMessages(context.Background())
		if err != nil {
			return promptUsageLoadedMsg{userMessages: map[string]int64{}}
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

func (m *UI) localPromptUsageCounts(now time.Time) (fiveHour, sevenDay int) {
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
	fiveHourCount, sevenDayCount := m.localPromptUsageCounts(time.Now())

	fiveHourPct := usagePercent(fiveHourCount, localMaxPlan5HourPromptLimit)
	sevenDayPct := usagePercent(sevenDayCount, localMaxPlan7DayPromptLimit)

	fiveHourLine := fmt.Sprintf("● 5h prompts %d/%d (%d%%)", fiveHourCount, localMaxPlan5HourPromptLimit, fiveHourPct)
	sevenDayLine := fmt.Sprintf("● 7d prompts %d/%d (%d%%)", sevenDayCount, localMaxPlan7DayPromptLimit, sevenDayPct)

	fiveHourLine = usageSeverityStyle(t, fiveHourPct).Render(fiveHourLine)
	sevenDayLine = usageSeverityStyle(t, sevenDayPct).Render(sevenDayLine)

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

func usageSeverityStyle(t *styles.Styles, percent int) lipgloss.Style {
	switch {
	case percent >= localDangerThresholdPercent:
		return t.Base.Foreground(t.Red)
	case percent >= localWarnThresholdPercent:
		return t.Base.Foreground(t.Yellow)
	default:
		return t.Base.Foreground(t.Green)
	}
}
