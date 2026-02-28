package model

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

// shadowState represents the shadow daemon state file structure
type shadowState struct {
	Status         string `json:"status"`
	HeartbeatCount int    `json:"heartbeat_count"`
}

// shadowCacheData holds cached shadow status
type shadowCacheData struct {
	active     bool
	heartbeats int
	timestamp  time.Time
}

// shadowCache is a package-level cache for shadow status
var shadowCache struct {
	mu    sync.RWMutex
	data  map[string]*shadowCacheData
}

func init() {
	shadowCache.data = make(map[string]*shadowCacheData)
}

// checkShadowStatus checks if shadow daemon is running for the given project path.
// Results are cached based on the TTL from config.
func checkShadowStatus(projectPath string, ttl time.Duration) (bool, int) {
	cacheKey := projectPath

	// Check cache first
	shadowCache.mu.RLock()
	if cached, ok := shadowCache.data[cacheKey]; ok {
		if time.Since(cached.timestamp) < ttl {
			shadowCache.mu.RUnlock()
			return cached.active, cached.heartbeats
		}
	}
	shadowCache.mu.RUnlock()

	// Cache miss or expired - read from file
	active, heartbeats := readShadowStateFile(projectPath)

	// Update cache
	shadowCache.mu.Lock()
	shadowCache.data[cacheKey] = &shadowCacheData{
		active:     active,
		heartbeats: heartbeats,
		timestamp:  time.Now(),
	}
	shadowCache.mu.Unlock()

	return active, heartbeats
}

// readShadowStateFile reads the shadow daemon state file directly.
func readShadowStateFile(projectPath string) (bool, int) {
	// Generate the same hash as the Python daemon: md5(path)[:12]
	hasher := md5.New()
	hasher.Write([]byte(projectPath))
	hash := hex.EncodeToString(hasher.Sum(nil))[:12]

	// Check shadow state file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, 0
	}

	stateFile := filepath.Join(homeDir, ".local", "share", "floyd", "shadow", hash, "state.json")

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return false, 0
	}

	var state shadowState
	if err := json.Unmarshal(data, &state); err != nil {
		return false, 0
	}

	return state.Status == "running", state.HeartbeatCount
}

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

	// Check shadow status (if enabled in config)
	cfg := com.Config()
	shadowCfg := cfg.Shadow()
	if shadowCfg.Enabled != nil && *shadowCfg.Enabled {
		ttl := time.Duration(shadowCfg.CacheTTLSeconds) * time.Second
		shadowActive, heartbeats := checkShadowStatus(cfg.WorkingDir(), ttl)
		if shadowActive {
			shadowIndicator := t.Header.Percentage.Render(fmt.Sprintf("♥%d", heartbeats))
			parts = append(parts, shadowIndicator)
		}
	}

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
		// Use TOTAL tokens for percentage (conservative - shows actual context pressure)
		// The sidebar shows the detailed dual-display with cache info
		totalTokens := session.CompletionTokens + session.PromptTokens
		percentage = (float64(totalTokens) / float64(contextWindow)) * 100
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
	cwd := fsext.DirTrim(fsext.PrettyPath(cfg.WorkingDir()), dirTrimLimit)
	cwd = ansi.Truncate(cwd, max(0, availWidth-lipgloss.Width(metadata)), "…")
	cwd = t.Header.WorkingDir.Render(cwd)

	return cwd + metadata
}
