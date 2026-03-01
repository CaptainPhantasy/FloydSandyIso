# Monitoring Code Documentation

This document describes the token monitoring and plan quota monitoring features that were added to the sidebar.

## Overview

Two monitoring features were added:

1. **Token Context Monitoring** - Displays token usage with stoplight color indicator
2. **Plan Quota Monitoring** - Tracks local prompt usage against estimated plan limits

---

## 1. Token Context Monitoring

### Location
`internal/ui/common/elements.go`

### Constants

```go
// Stoplight threshold constants for context usage
const (
    StoplightGreenThreshold  = 70.0 // 0-70%: Normal operation
    StoplightYellowThreshold = 84.0 // 71-84%: Caution
    StoplightRedThreshold    = 85.0 // 85%+: Critical, triggers auto-export
)

// Stoplight indicator emojis
const (
    StoplightGreen  = "🟢"
    StoplightYellow = "🟡"
    StoplightRed    = "🔴"
)
```

### ModelContextInfo Struct

```go
type ModelContextInfo struct {
    TotalTokens     int64   // Raw total (PromptTokens + CompletionTokens)
    EffectiveTokens int64   // After cache subtraction (Total - CacheReadTokens)
    CacheReadTokens int64   // How much was served from cache
    CachePercent    float64 // What percentage was cached
    ModelContext    int64   // Context window size
    Cost            float64
}
```

### Functions

#### `getStoplightIndicator(pct float64) string`
Returns emoji based on percentage:
- 85%+ → 🔴
- 71-84% → 🟡
- 0-70% → 🟢

#### `formatTokensAndCost(t *styles.Styles, info *ModelContextInfo) string`
Formats display string:
```
🟢 22% • 45K/2K • $0.24
95% cached
⚠ Near limit • Ctrl+N  (only if 85%+)
```

### Sidebar Integration
`internal/ui/model/sidebar.go` - `modelInfo(width int)` function builds `ModelContextInfo` from session data and calls `common.ModelInfo()` which includes the stoplight display.

---

## 2. Plan Quota Monitoring

### Location
`internal/ui/model/prompt_usage.go`

### Configuration
`internal/config/config.go`:

```go
type PromptQuotaConfig struct {
    Enabled       *bool `json:"enabled,omitempty"`        // default: true
    Limit5h       int   `json:"limit_5h,omitempty"`       // default: 1600
    Limit7d       int   `json:"limit_7d,omitempty"`       // default: 8000
    WarnPercent   int   `json:"warn_percent,omitempty"`   // default: 75
    DangerPercent int   `json:"danger_percent,omitempty"` // default: 90
}
```

### Functions

#### `loadPromptUsage() tea.Cmd`
Loads all user messages from database to track prompt counts.

#### `localPromptUsageCounts(now time.Time) (fiveHour, sevenDay int)`
Counts prompts in 5-hour and 7-day windows from local database.

#### `promptQuotaInfo(width int) string`
Renders sidebar section:
```
Plan Quota (est.)
● 5h prompts 45/1600 (3%)
● 7d prompts 230/8000 (3%)
Local Floyd usage only
```

### Sidebar Integration
`internal/ui/model/sidebar.go` - `drawSidebar()` calls `m.promptQuotaInfo(width)` and includes it in the sidebar.

---

## 3. Header Percentage Display

### Location
`internal/ui/model/header.go`

### Function: `renderMetadataLine()`

Displays conservative percentage in header using **total tokens** (not effective):
```go
totalTokens := session.CompletionTokens + session.PromptTokens
percentage = (float64(totalTokens) / float64(contextWindow)) * 100
```

Format: `45%` displayed in header with `ctrl+d` hint.

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/ui/common/elements.go` | Added `ModelContextInfo`, `getStoplightIndicator`, `formatTokensAndCost` |
| `internal/ui/model/prompt_usage.go` | Added quota tracking functions |
| `internal/ui/model/sidebar.go` | Integrated stoplight and quota into sidebar |
| `internal/ui/model/header.go` | Added percentage to header |
| `internal/config/config.go` | Added `PromptQuotaConfig` struct |
| `internal/config/load.go` | Added `PromptQuota()` and `DefaultPromptQuotaConfig()` |

---

## Removal Instructions

To restore sidebar to original platform defaults:

1. Remove `promptQuotaInfo()` call from `drawSidebar()` in sidebar.go
2. Remove stoplight indicator from `formatTokensAndCost()` in elements.go
3. Remove percentage from `renderMetadataLine()` in header.go
4. Remove `PromptQuotaConfig` from config files
5. Remove `prompt_usage.go` file entirely
