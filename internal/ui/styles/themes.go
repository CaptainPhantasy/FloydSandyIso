package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

// ThemeName is the display name of a theme.
type ThemeName string

const (
	ThemeDefault    ThemeName = "Default"
	ThemeSunset     ThemeName = "Sunset"
	ThemeDeepSea    ThemeName = "Deep Sea"
	ThemeNeonHacker ThemeName = "Neon Hacker"
)

// ThemePreset defines the accent palette that drives the entire UI
// appearance. Every field that is nil falls back to the default
// charmtone value.
type ThemePreset struct {
	Name ThemeName

	// Core accents — these propagate into borders, selection
	// highlights, logo gradient, dialog focus, header, etc.
	Primary     color.Color
	Secondary   color.Color
	Tertiary    color.Color
	BorderFocus color.Color

	// Logo gradient endpoints.
	LogoColorA color.Color
	LogoColorB color.Color
	LogoCharm  color.Color
	LogoField  color.Color

	// Status accent overrides (nil = keep default).
	Green     color.Color
	GreenDark color.Color
	Blue      color.Color
	BlueDark  color.Color
	Red       color.Color
}

// mustHex parses a hex color string and panics on failure.
func mustHex(hex string) colorful.Color {
	c, err := colorful.Hex(hex)
	if err != nil {
		panic("invalid hex color: " + hex)
	}
	return c
}

// themePresets is the ordered list of available themes. The first entry
// ("Default") is populated at runtime from the original palette.
var themePresets []ThemePreset

func init() {
	themePresets = []ThemePreset{
		{Name: ThemeDefault}, // filled by initThemes
		{
			Name:        ThemeSunset,
			Primary:     mustHex("#FF5F6D"),
			Secondary:   mustHex("#FFC371"),
			Tertiary:    mustHex("#FFD194"),
			BorderFocus: mustHex("#FF5F6D"),
			LogoColorA:  mustHex("#FF5F6D"),
			LogoColorB:  mustHex("#FFC371"),
			LogoCharm:   mustHex("#FFC371"),
			LogoField:   mustHex("#FF5F6D"),
			Green:       mustHex("#FFB347"),
			GreenDark:   mustHex("#FF8C42"),
			Blue:        mustHex("#FF6B6B"),
			BlueDark:    mustHex("#CC4455"),
			Red:         mustHex("#FF4444"),
		},
		{
			Name:        ThemeDeepSea,
			Primary:     mustHex("#00CED1"),
			Secondary:   mustHex("#00FFFF"),
			Tertiary:    mustHex("#7FFFD4"),
			BorderFocus: mustHex("#00CED1"),
			LogoColorA:  mustHex("#00FFFF"),
			LogoColorB:  mustHex("#0055FF"),
			LogoCharm:   mustHex("#00FFFF"),
			LogoField:   mustHex("#0055FF"),
			Green:       mustHex("#00E5CC"),
			GreenDark:   mustHex("#00B89C"),
			Blue:        mustHex("#00BFFF"),
			BlueDark:    mustHex("#0077B6"),
			Red:         mustHex("#FF6B9D"),
		},
		{
			Name:        ThemeNeonHacker,
			Primary:     mustHex("#39FF14"),
			Secondary:   mustHex("#00FF41"),
			Tertiary:    mustHex("#ADFF2F"),
			BorderFocus: mustHex("#39FF14"),
			LogoColorA:  mustHex("#39FF14"),
			LogoColorB:  mustHex("#006400"),
			LogoCharm:   mustHex("#00FF41"),
			LogoField:   mustHex("#39FF14"),
			Green:       mustHex("#39FF14"),
			GreenDark:   mustHex("#228B22"),
			Blue:        mustHex("#00FF41"),
			BlueDark:    mustHex("#006400"),
			Red:         mustHex("#FF3131"),
		},
	}
}

// ThemePresets returns all available theme presets.
func ThemePresets() []ThemePreset {
	return themePresets
}

// initThemes snapshots the current palette into the Default preset so
// it can be restored later.
func (s *Styles) initThemes() {
	themePresets[0] = ThemePreset{
		Name:        ThemeDefault,
		Primary:     s.Primary,
		Secondary:   s.Secondary,
		Tertiary:    s.Tertiary,
		BorderFocus: s.BorderColor,
		LogoColorA:  s.LogoTitleColorA,
		LogoColorB:  s.LogoTitleColorB,
		LogoCharm:   s.LogoCharmColor,
		LogoField:   s.LogoFieldColor,
		Green:       s.Green,
		GreenDark:   s.GreenDark,
		Blue:        s.Blue,
		BlueDark:    s.BlueDark,
		Red:         s.Red,
	}
}

// ActiveTheme returns the name of the currently active theme.
func (s *Styles) ActiveTheme() ThemeName {
	return s.activeTheme
}

// ApplyTheme applies a full theme preset, recoloring accents, borders,
// logo, dialogs, editor prompts, header, tool icons, and other key
// surfaces throughout the UI.
func (s *Styles) ApplyTheme(t ThemePreset) {
	s.activeTheme = t.Name

	// --- Core semantic colors ---
	s.Primary = t.Primary
	s.Secondary = t.Secondary
	s.Tertiary = t.Tertiary
	s.BorderColor = t.BorderFocus

	if t.Green != nil {
		s.Green = t.Green
	}
	if t.GreenDark != nil {
		s.GreenDark = t.GreenDark
	}
	if t.Blue != nil {
		s.Blue = t.Blue
	}
	if t.BlueDark != nil {
		s.BlueDark = t.BlueDark
	}
	if t.Red != nil {
		s.Red = t.Red
	}

	// --- Logo ---
	s.LogoTitleColorA = t.LogoColorA
	s.LogoTitleColorB = t.LogoColorB
	s.LogoCharmColor = t.LogoCharm
	s.LogoFieldColor = t.LogoField
	s.LogoVersionColor = t.Primary

	// --- Header ---
	s.Header.Charm = lipgloss.NewStyle().Foreground(t.Secondary)
	s.Header.Diagonals = lipgloss.NewStyle().Foreground(t.Primary)

	// --- Borders ---
	s.CompactDetails.View = s.Base.Padding(0, 1, 1, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus)

	// --- Dialog ---
	s.Dialog.Title = lipgloss.NewStyle().Padding(0, 1).Foreground(t.Primary)
	s.Dialog.TitleText = lipgloss.NewStyle().Foreground(t.Primary)
	s.Dialog.TitleAccent = lipgloss.NewStyle().Foreground(t.GreenDark).Bold(true)
	s.Dialog.View = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus)
	s.Dialog.PrimaryText = lipgloss.NewStyle().Padding(0, 1).Foreground(t.Primary)
	s.Dialog.SelectedItem = lipgloss.NewStyle().Padding(0, 1).
		Background(t.Primary).
		Foreground(s.FgBase)
	s.Dialog.Spinner = lipgloss.NewStyle().Foreground(t.Secondary)
	s.Dialog.ScrollbarThumb = lipgloss.NewStyle().Foreground(t.Secondary)

	// --- Editor prompt ---
	s.EditorPromptNormalFocused = lipgloss.NewStyle().Foreground(t.GreenDark).SetString("::: ")
	s.EditorPromptNormalBlurred = s.EditorPromptNormalFocused.Foreground(s.FgMuted)

	// --- Text inputs ---
	s.TextInput.Focused.Prompt = lipgloss.NewStyle().Foreground(t.Tertiary)
	s.TextInput.Cursor.Color = t.Secondary
	s.TextArea.Focused.Prompt = lipgloss.NewStyle().Foreground(t.Tertiary)
	s.TextArea.Cursor.Color = t.Secondary

	// --- Message borders ---
	s.Chat.Message.UserBlurred = s.Chat.Message.NoContent.PaddingLeft(1).
		BorderLeft(true).
		BorderForeground(t.Primary).
		BorderStyle(lipgloss.NormalBorder())
	s.Chat.Message.UserFocused = s.Chat.Message.NoContent.PaddingLeft(1).
		BorderLeft(true).
		BorderForeground(t.Primary).
		BorderStyle(s.FocusedMessageBorder)
	s.Chat.Message.AssistantFocused = s.Chat.Message.NoContent.PaddingLeft(1).
		BorderLeft(true).
		BorderForeground(t.GreenDark).
		BorderStyle(s.FocusedMessageBorder)

	// --- Tool call icons ---
	s.Tool.IconPending = lipgloss.NewStyle().Foreground(t.GreenDark)
	s.Tool.IconSuccess = lipgloss.NewStyle().Foreground(t.Green)
	s.Tool.IconError = lipgloss.NewStyle().Foreground(s.Red)
	s.Tool.NameNormal = lipgloss.NewStyle().Foreground(t.Blue)
	s.Tool.NameNested = lipgloss.NewStyle().Foreground(t.BlueDark)
	s.Tool.MCPName = lipgloss.NewStyle().Foreground(t.Blue)
	s.Tool.MCPToolName = lipgloss.NewStyle().Foreground(t.BlueDark)
	s.Tool.MCPArrow = lipgloss.NewStyle().Foreground(t.Blue).SetString(ArrowRightIcon)
	s.Tool.TodoCompletedIcon = lipgloss.NewStyle().Foreground(t.Green)
	s.Tool.TodoInProgressIcon = lipgloss.NewStyle().Foreground(t.GreenDark)
	s.Tool.TodoRatio = lipgloss.NewStyle().Foreground(t.BlueDark)

	// --- Tool call focused borders ---
	s.Chat.Message.ToolCallFocused = s.Muted.PaddingLeft(1).
		BorderStyle(s.FocusedMessageBorder).
		BorderLeft(true).
		BorderForeground(t.GreenDark)

	// --- File picker ---
	s.FilePicker.Directory = lipgloss.NewStyle().Foreground(t.Primary)
	s.FilePicker.Selected = lipgloss.NewStyle().Background(t.Primary).Foreground(s.FgBase)

	// --- Completions ---
	s.Completions.Focused = lipgloss.NewStyle().Background(t.Primary).Foreground(s.White)

	// --- Pills ---
	s.Pills.TodoSpinner = lipgloss.NewStyle().Foreground(t.GreenDark)

	// --- Status indicators ---
	s.Status.SuccessIndicator = lipgloss.NewStyle().Foreground(s.BgSubtle).
		Background(t.Green).Padding(0, 1).Bold(true).SetString("OKAY!")
	s.Status.InfoIndicator = s.Status.SuccessIndicator
	s.Status.UpdateIndicator = s.Status.SuccessIndicator.SetString("HEY!")
	s.Status.SuccessMessage = lipgloss.NewStyle().Foreground(s.BgSubtle).
		Background(t.GreenDark).Padding(0, 1)
	s.Status.InfoMessage = s.Status.SuccessMessage
	s.Status.UpdateMessage = s.Status.SuccessMessage

	// --- Radio ---
	s.RadioOn = s.HalfMuted.SetString(RadioOn)
	s.RadioOff = s.HalfMuted.SetString(RadioOff)

	// --- LSP / MCP online icon ---
	s.ItemOnlineIcon = lipgloss.NewStyle().Foreground(t.GreenDark).SetString("●")
}

// CycleTheme advances to the next theme preset, applies it, and
// returns the name of the newly active theme.
func (s *Styles) CycleTheme() ThemeName {
	presets := ThemePresets()
	idx := 0
	for i, p := range presets {
		if p.Name == s.activeTheme {
			idx = i
			break
		}
	}
	next := (idx + 1) % len(presets)
	s.ApplyTheme(presets[next])
	return presets[next].Name
}

// CycleThemeReverse moves to the previous theme preset, applies it,
// and returns the name of the newly active theme.
func (s *Styles) CycleThemeReverse() ThemeName {
	presets := ThemePresets()
	idx := 0
	for i, p := range presets {
		if p.Name == s.activeTheme {
			idx = i
			break
		}
	}
	prev := (idx - 1 + len(presets)) % len(presets)
	s.ApplyTheme(presets[prev])
	return presets[prev].Name
}
