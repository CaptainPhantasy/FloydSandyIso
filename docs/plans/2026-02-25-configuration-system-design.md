# FLOYD Configuration System Design

**Created:** 2026-02-25
**Status:** DRAFT
**Goal:** Configurable runtime settings for Z.AI concurrency, prompt quotas, and shadow daemon with live TUI updates.

---

## 1. PROBLEM STATEMENT

Current implementation has hardcoded values:

```go
// internal/agent/agent.go
zaiConcurrencyCooldown = 2 * time.Second  // Hardcoded

// internal/ui/model/prompt_usage.go
localMaxPlan5HourPromptLimit = 1600  // Hardcoded
localMaxPlan7DayPromptLimit  = 8000  // Hardcoded
localWarnThresholdPercent    = 75    // Hardcoded
localDangerThresholdPercent  = 90    // Hardcoded
```

**Issues:**
1. Users cannot adjust rate limiting for their API tier
2. Prompt quotas are guesses, not configurable
3. No visibility into current settings in TUI
4. Changes require recompilation

---

## 2. CONFIGURATION SCHEMA

### 2.1 Physical Storage Location

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ Level          │ Path                              │ Priority              │
├────────────────┼───────────────────────────────────┼───────────────────────┤
│ Global         │ ~/.config/floyd/floyd.json        │ Base (lowest)         │
│ Project        │ ./.floyd/floyd.json               │ Override (highest)    │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Schema Extension

Add to `internal/config/config.go`:

```go
// TUIOptions extended with rate limiting and quota settings
type TUIOptions struct {
    CompactMode  bool         `json:"compact_mode,omitempty"`
    DiffMode     string       `json:"diff_mode,omitempty"`
    Transparent  *bool        `json:"transparent,omitempty"`
    Completions  Completions  `json:"completions,omitzero"`

    // NEW: Runtime behavior settings
    RateLimit    *RateLimitConfig  `json:"rate_limit,omitempty"`
    PromptQuota  *PromptQuotaConfig `json:"prompt_quota,omitempty"`
    Shadow       *ShadowConfig     `json:"shadow,omitempty"`
}

// RateLimitConfig controls API rate limiting behavior
type RateLimitConfig struct {
    // Enabled controls whether rate limiting is active
    Enabled *bool `json:"enabled,omitempty" jsonschema:"default=true"`

    // CooldownMs is the milliseconds to wait between LLM requests
    CooldownMs int `json:"cooldown_ms,omitempty" jsonschema:"default=2000,description=Milliseconds to wait between LLM requests"`

    // Providers is a list of provider names to apply rate limiting to
    // If empty, applies to all providers
    Providers []string `json:"providers,omitempty" jsonschema:"description=Provider names to rate limit (e.g., zai, openai). Empty = all"`

    // MaxConcurrent is the maximum simultaneous LLM requests (default: 1)
    MaxConcurrent int `json:"max_concurrent,omitempty" jsonschema:"default=1,description=Maximum simultaneous LLM requests"`
}

// PromptQuotaConfig controls prompt usage tracking and display
type PromptQuotaConfig struct {
    // Enabled controls whether quota tracking is shown
    Enabled *bool `json:"enabled,omitempty" jsonschema:"default=true"`

    // Limit5h is the 5-hour prompt limit for warnings
    Limit5h int `json:"limit_5h,omitempty" jsonschema:"default=1600,description=5-hour prompt limit"`

    // Limit7d is the 7-day prompt limit for warnings
    Limit7d int `json:"limit_7d,omitempty" jsonschema:"default=8000,description=7-day prompt limit"`

    // WarnPercent is the percentage at which to show warning color (yellow)
    WarnPercent int `json:"warn_percent,omitempty" jsonschema:"default=75,description=Warning threshold percentage"`

    // DangerPercent is the percentage at which to show danger color (red)
    DangerPercent int `json:"danger_percent,omitempty" jsonschema:"default=90,description=Danger threshold percentage"`
}

// ShadowConfig controls shadow daemon integration
type ShadowConfig struct {
    // Enabled controls whether shadow status is shown in header
    Enabled *bool `json:"enabled,omitempty" jsonschema:"default=true"`

    // StatePath is the path to shadow daemon state files (supports ~ expansion)
    // Default: ~/.local/share/floyd/shadow/{hash}/state.json
    StatePath string `json:"state_path,omitempty" jsonschema:"description=Override path to shadow state directory"`

    // CacheTTLSeconds is how long to cache shadow status (default: 5)
    CacheTTLSeconds int `json:"cache_ttl_seconds,omitempty" jsonschema:"default=5,description=Seconds to cache shadow status"`
}
```

### 2.3 Example Configuration

```json
{
  "$schema": "https://charm.land/floyd.json",
  "providers": { ... },
  "models": { ... },
  "tui": {
    "compact_mode": false,
    "rate_limit": {
      "enabled": true,
      "cooldown_ms": 2000,
      "providers": ["zai"],
      "max_concurrent": 1
    },
    "prompt_quota": {
      "enabled": true,
      "limit_5h": 1600,
      "limit_7d": 8000,
      "warn_percent": 75,
      "danger_percent": 90
    },
    "shadow": {
      "enabled": true,
      "cache_ttl_seconds": 5
    }
  }
}
```

---

## 3. STATE MANAGEMENT ARCHITECTURE

### 3.1 State Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CONFIGURATION STATE FLOW                          │
└─────────────────────────────────────────────────────────────────────────────┘

                    ┌──────────────────┐
                    │  floyd.json      │
                    │  (disk)          │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │  config.Load()   │
                    │  (startup)       │
                    └────────┬─────────┘
                             │
                             ▼
              ┌──────────────────────────────┐
              │      Config struct          │
              │  (in-memory, read-only)     │
              └──────────────┬───────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │ SessionAgent │ │  UI Model    │ │ ShadowCache  │
    │ (rate limit) │ │ (quotas)     │ │ (daemon)     │
    └──────────────┘ └──────────────┘ └──────────────┘
```

### 3.2 Runtime State vs Persistent Config

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│ Component        │ Source        │ Mutability │ Persistence              │
├──────────────────┼───────────────┼────────────┼──────────────────────────┤
│ RateLimitConfig  │ floyd.json    │ Read-only  │ File edit + restart      │
│ PromptQuotaConfig│ floyd.json    │ Read-only  │ File edit + restart      │
│ ShadowConfig     │ floyd.json    │ Read-only  │ File edit + restart      │
│ localUserMessages│ SQLite DB     │ Runtime    │ Per-session, persisted   │
│ shadowCache      │ Memory        │ Runtime    │ TTL-based expiration     │
└──────────────────┴───────────────┴────────────┴──────────────────────────┘
```

### 3.3 Config Access Pattern

```go
// internal/config/config.go - Add accessor methods

func (o *Options) RateLimit() RateLimitConfig {
    if o.TUI == nil || o.TUI.RateLimit == nil {
        return DefaultRateLimitConfig()
    }
    cfg := *o.TUI.RateLimit
    // Apply defaults for unset fields
    if cfg.Enabled == nil {
        cfg.Enabled = ptr(true)
    }
    if cfg.CooldownMs == 0 {
        cfg.CooldownMs = 2000
    }
    if cfg.MaxConcurrent == 0 {
        cfg.MaxConcurrent = 1
    }
    return cfg
}

func (o *Options) PromptQuota() PromptQuotaConfig {
    if o.TUI == nil || o.TUI.PromptQuota == nil {
        return DefaultPromptQuotaConfig()
    }
    // Similar default application...
}

func (o *Options) Shadow() ShadowConfig {
    if o.TUI == nil || o.TUI.Shadow == nil {
        return DefaultShadowConfig()
    }
    // Similar default application...
}
```

---

## 4. UI CONTROLS PLACEMENT

### 4.1 Sidebar - Prompt Quota Display (EXISTING)

```
┌─────────────────────────────────┐
│  ░░░ FLOYD ░░░                  │  <- Logo
│  Session Title                  │
│                                 │
│  ~/projects/my-app              │  <- Working dir
│                                 │
│  Model: glm-5                   │  <- Model info
│  Provider: Z.AI                 │
│  Tokens: 45,230 / 200,000       │
│                                 │
│  ┌─────────────────────────────┐│
│  │ Plan Quota (est.)           ││  <- PromptQuotaConfig
│  │ ● 5h prompts 234/1600 (15%) ││     Controls these values
│  │ ● 7d prompts 1,823/8000 (23%)││     and threshold colors
│  │ Local Floyd usage only      ││
│  └─────────────────────────────┘│
│                                 │
│  Files (3)                      │
│  • main.go                      │
│  • config.go                    │
│                                 │
│  LSP (1)                        │
│  ● gopls ✓                      │
│                                 │
│  MCP (5)                        │
│  ● floyd-supercache ✓           │
│  ● floyd-runner ✓               │
└─────────────────────────────────┘
```

### 4.2 Header - Shadow Status (EXISTING)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ ░░░ FLOYD ░░░              ♥127           ●0 errors   ⚠ 3 warnings        │
│                             ▲              ▲            ▲                   │
│                             │              │            │                   │
│                        ShadowConfig    LSP diag     LSP diag                │
│                        (enabled)        (always)    (always)                │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.3 NEW: Settings Dialog (Keyboard Shortcut: Ctrl+O)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           FLOYD Settings                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Rate Limiting                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ [✓] Enabled                                                         │   │
│  │ Cooldown (ms):     [2000    ]                                       │   │
│  │ Max Concurrent:    [1       ]                                       │   │
│  │ Providers:         [zai     ] (comma-separated, empty = all)        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Prompt Quota                                                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ [✓] Enabled                                                         │   │
│  │ 5h Limit:          [1600    ]                                       │   │
│  │ 7d Limit:          [8000    ]                                       │   │
│  │ Warn %:            [75      ]                                       │   │
│  │ Danger %:          [90      ]                                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  Shadow Daemon                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │ [✓] Show status in header                                           │   │
│  │ Cache TTL (sec):   [5       ]                                       │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ────────────────────────────────────────────────────────────────────────   │
│  Config file: ~/.config/floyd/floyd.json                                   │
│  [Cancel]                                                      [Save]      │
│                                                                             │
│  esc close • tab next • shift+tab prev • enter confirm                     │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.4 Alternative: Command-Based Settings

Add to command dialog (Ctrl+K):

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ > set rate_limit.cooldown 3000                                             │
│ > set prompt_quota.limit_5h 2000                                           │
│ > set shadow.enabled false                                                 │
│ > show config                                                              │
│ > edit config                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 5. IMPLEMENTATION PHASES

### Phase 1: Schema & Defaults (No UI Changes)

**Files to modify:**
- `internal/config/config.go` - Add structs
- `internal/config/load.go` - Add defaults

**Changes:**
1. Add `RateLimitConfig`, `PromptQuotaConfig`, `ShadowConfig` structs
2. Add accessor methods with defaults
3. Replace hardcoded constants with config reads

**Testing:**
- Config loads with defaults when fields absent
- Config overrides work when present
- Backward compatible (existing configs work)

### Phase 2: Runtime Integration

**Files to modify:**
- `internal/agent/agent.go` - Use rate limit config
- `internal/ui/model/prompt_usage.go` - Use quota config
- `internal/ui/model/header.go` - Use shadow config + add cache

**Changes:**
1. `withZAIConcurrencyGate` becomes `withConcurrencyGate` (provider-aware)
2. Add shadow status caching with TTL
3. Prompt quota uses config values

**Testing:**
- Rate limiting only applies to specified providers
- Shadow status cached correctly
- Quota thresholds match config

### Phase 3: Settings Dialog

**Files to create:**
- `internal/ui/dialog/settings.go` - New dialog

**Files to modify:**
- `internal/ui/model/ui.go` - Wire up dialog
- `internal/config/load.go` - Add save function

**Changes:**
1. Create settings dialog with form fields
2. Add keyboard shortcut (Ctrl+O)
3. Implement config save with validation

**Testing:**
- Dialog opens/closes correctly
- Values persist to floyd.json
- Invalid values rejected with message

### Phase 4: Command Interface (Optional)

**Files to modify:**
- `internal/ui/model/commands.go` - Add set/show commands

**Changes:**
1. Add `set <path> <value>` command
2. Add `show config` command
3. Add `edit config` command (opens $EDITOR)

---

## 6. DETAILED FILE CHANGES

### 6.1 internal/config/config.go

```go
// Add after TUIOptions (around line 196)

// RateLimitConfig controls API rate limiting behavior.
type RateLimitConfig struct {
    Enabled       *bool    `json:"enabled,omitempty" jsonschema:"description=Enable rate limiting,default=true"`
    CooldownMs    int      `json:"cooldown_ms,omitempty" jsonschema:"description=Milliseconds between requests,default=2000"`
    MaxConcurrent int      `json:"max_concurrent,omitempty" jsonschema:"description=Max simultaneous requests,default=1"`
    Providers     []string `json:"providers,omitempty" jsonschema:"description=Providers to rate limit (empty = all)"`
}

// PromptQuotaConfig controls prompt usage display.
type PromptQuotaConfig struct {
    Enabled      *bool `json:"enabled,omitempty" jsonschema:"description=Show quota in sidebar,default=true"`
    Limit5h      int   `json:"limit_5h,omitempty" jsonschema:"description=5-hour prompt limit,default=1600"`
    Limit7d      int   `json:"limit_7d,omitempty" jsonschema:"description=7-day prompt limit,default=8000"`
    WarnPercent  int   `json:"warn_percent,omitempty" jsonschema:"description=Warning threshold %,default=75"`
    DangerPercent int  `json:"danger_percent,omitempty" jsonschema:"description=Danger threshold %,default=90"`
}

// ShadowConfig controls shadow daemon integration.
type ShadowConfig struct {
    Enabled         *bool  `json:"enabled,omitempty" jsonschema:"description=Show shadow status in header,default=true"`
    StatePath       string `json:"state_path,omitempty" jsonschema:"description=Override shadow state path"`
    CacheTTLSeconds int    `json:"cache_ttl_seconds,omitempty" jsonschema:"description=Cache TTL in seconds,default=5"`
}
```

### 6.2 internal/config/config.go - Extend TUIOptions

```go
type TUIOptions struct {
    CompactMode bool        `json:"compact_mode,omitempty"`
    DiffMode    string      `json:"diff_mode,omitempty"`
    Transparent *bool       `json:"transparent,omitempty"`
    Completions Completions `json:"completions,omitzero"`

    // Runtime behavior configuration
    RateLimit   *RateLimitConfig   `json:"rate_limit,omitempty" jsonschema:"description=API rate limiting settings"`
    PromptQuota *PromptQuotaConfig `json:"prompt_quota,omitempty" jsonschema:"description=Prompt usage quota settings"`
    Shadow      *ShadowConfig      `json:"shadow,omitempty" jsonschema:"description=Shadow daemon settings"`
}
```

### 6.3 internal/config/load.go - Add Accessors

```go
// DefaultRateLimitConfig returns default rate limit settings.
func DefaultRateLimitConfig() RateLimitConfig {
    return RateLimitConfig{
        Enabled:       ptr(true),
        CooldownMs:    2000,
        MaxConcurrent: 1,
        Providers:     []string{}, // All providers
    }
}

func (c *Config) RateLimit() RateLimitConfig {
    if c.Options.TUI == nil || c.Options.TUI.RateLimit == nil {
        return DefaultRateLimitConfig()
    }
    cfg := *c.Options.TUI.RateLimit
    def := DefaultRateLimitConfig()
    if cfg.Enabled == nil {
        cfg.Enabled = def.Enabled
    }
    if cfg.CooldownMs == 0 {
        cfg.CooldownMs = def.CooldownMs
    }
    if cfg.MaxConcurrent == 0 {
        cfg.MaxConcurrent = def.MaxConcurrent
    }
    return cfg
}

// Similar for PromptQuota() and Shadow()...
```

### 6.4 internal/agent/agent.go - Use Config

```go
// Replace global limiter with config-aware version
func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error) {
    // ... existing code ...

    // Get rate limit config
    rateLimit := a.cfg.RateLimit()
    if *rateLimit.Enabled && a.shouldRateLimit(largeModel.Provider) {
        result, err = withConcurrencyGate(rateLimit, func() (*fantasy.AgentResult, error) {
            return agent.Stream(genCtx, fantasy.AgentStreamCall{...})
        })
    } else {
        result, err = agent.Stream(genCtx, fantasy.AgentStreamCall{...})
    }
}

func (a *sessionAgent) shouldRateLimit(provider string) bool {
    rateLimit := a.cfg.RateLimit()
    if len(rateLimit.Providers) == 0 {
        return true // All providers
    }
    for _, p := range rateLimit.Providers {
        if strings.EqualFold(p, provider) {
            return true
        }
    }
    return false
}
```

### 6.5 internal/ui/model/header.go - Add Cache

```go
type header struct {
    logo        string
    compactLogo string

    // Shadow cache
    shadowCache      *shadowCacheData
    shadowCacheTime  time.Time
}

type shadowCacheData struct {
    active     bool
    heartbeats int
}

func (h *header) getShadowStatus(projectPath string, ttl time.Duration) (bool, int) {
    now := time.Now()

    // Check cache
    if h.shadowCache != nil && now.Sub(h.shadowCacheTime) < ttl {
        return h.shadowCache.active, h.shadowCache.heartbeats
    }

    // Read from file
    active, heartbeats := checkShadowStatus(projectPath)
    h.shadowCache = &shadowCacheData{active: active, heartbeats: heartbeats}
    h.shadowCacheTime = now

    return active, heartbeats
}
```

---

## 7. TESTING CHECKLIST

### Config Loading
- [ ] Default values applied when config absent
- [ ] Custom values override defaults
- [ ] Invalid values logged, defaults used
- [ ] Backward compatible with old configs

### Rate Limiting
- [ ] Disabled when enabled=false
- [ ] Applies to all providers when providers=[]
- [ ] Applies only to listed providers
- [ ] Respects cooldown_ms value
- [ ] Respects max_concurrent value

### Prompt Quota
- [ ] Hidden when enabled=false
- [ ] Shows correct limits from config
- [ ] Warning color at warn_percent
- [ ] Danger color at danger_percent

### Shadow Status
- [ ] Hidden when enabled=false
- [ ] Respects cache TTL
- [ ] Handles missing state file gracefully

### Settings Dialog
- [ ] Opens with Ctrl+O
- [ ] Shows current values
- [ ] Validates input
- [ ] Saves to correct file
- [ ] Applies on next prompt (no restart needed?)

---

## 8. RISKS & MITIGATIONS

| Risk | Mitigation |
|------|------------|
| Config file corruption | Write to temp file, then atomic rename |
| Invalid user input | Validate before save, show error message |
| Breaking existing configs | Use omitempty, provide sensible defaults |
| Performance impact | Cache config reads, not per-call |
| UI complexity | Keep dialog simple, power users edit JSON directly |

---

## 9. FUTURE ENHANCEMENTS

1. **Hot reload** - Watch floyd.json for changes, apply without restart
2. **Profile system** - Multiple configs (work, personal, etc.)
3. **Import/export** - Share configs between machines
4. **Cloud sync** - Optional encrypted sync to cloud storage
5. **Per-project overrides** - .floyd/floyd.json extends global config
