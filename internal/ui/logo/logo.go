// Package logo renders a Floyd wordmark in a stylized way.
package logo

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/slice"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

// letterform represents a letterform. It can be stretched horizontally by
// a given amount via the boolean argument.
type letterform func(bool) string

const diag = `╱`

// Opts are the options for rendering the Floyd title art.
type Opts struct {
	FieldColor   color.Color // diagonal lines
	TitleColorA  color.Color // left gradient ramp point
	TitleColorB  color.Color // right gradient ramp point
	CharmColor   color.Color // Charm™ text color
	VersionColor color.Color // Version text color
	Width        int         // width of the rendered logo, used for truncation
}

// Render renders the Floyd logo. Set the argument to true to render the narrow
// version, intended for use in a sidebar.
//
// The compact argument determines whether it renders compact for the sidebar
// or wider for the main pane.
func Render(s *styles.Styles, version string, compact bool, o Opts) string {
	const legacyai = " LEGACY AI™"

	fg := func(c color.Color, s string) string {
		return lipgloss.NewStyle().Foreground(c).Render(s)
	}

	// FLOYD ASCII art
	floydASCII := `    __/\\\\\\\\\\\\\\\___/\\\____________________/\\\\\________/\\\________/\\\___/\\\\\\\\\\\\_______________
     _\/\\\///////////___\/\\\__________________/\\\///\\\_____\///\\\____/\\\/___\/\\\////////\\\_____________
      _\/\\\______________\/\\\________________/\\\/__\///\\\_____\///\\\/\\\/_____\/\\\______\//\\\____________
       _\/\\\\\\\\\\\______\/\\\_______________/\\\______\//\\\______\///\\\/_______\/\\\_______\/\\\____________
        _\/\\\///////_______\/\\\______________\/\\\_______\/\\\________\/\\\________\/\\\_______\/\\\____________
         _\/\\\______________\/\\\______________\//\\\______/\\\_________\/\\\________\/\\\_______\/\\\____________
          _\/\\\______________\/\\\_______________\///\\\__/\\\___________\/\\\________\/\\\_______/\\\_____________
           _\/\\\______________\/\\\\\\\\\\\\\\\_____\///\\\\\/____________\/\\\________\/\\\\\\\\\\\\/______________
            _\///_______________\///////////////________\/////______________\///_________\////////////________________`

	floydWidth := lipgloss.Width(floydASCII)

	// Apply gradient coloring to the ASCII art
	b := new(strings.Builder)
	for r := range strings.SplitSeq(floydASCII, "\n") {
		fmt.Fprintln(b, styles.ApplyForegroundGrad(s, r, o.TitleColorA, o.TitleColorB))
	}
	floyd := b.String()

	// Meta row with legacy AI and version.
	metaRowGap := 1
	maxVersionWidth := floydWidth - lipgloss.Width(legacyai) - metaRowGap
	version = ansi.Truncate(version, maxVersionWidth, "…") // truncate version if too long.
	gap := max(0, floydWidth-lipgloss.Width(legacyai)-lipgloss.Width(version))
	metaRow := fg(o.CharmColor, legacyai) + strings.Repeat(" ", gap) + fg(o.VersionColor, version)

	// Join the meta row and FLOYD title.
	floyd = strings.TrimSpace(metaRow + "\n" + floyd)

	// Narrow version.
	if compact {
		fieldWidth := floydWidth
		if o.Width > 0 {
			fieldWidth = o.Width
		}
		field := fg(o.FieldColor, strings.Repeat(diag, fieldWidth))
		result := strings.Join([]string{field, field, floyd, field, ""}, "\n")
		if o.Width > 0 {
			lines := strings.Split(result, "\n")
			for i, line := range lines {
				lines[i] = ansi.Truncate(line, o.Width, "")
			}
			result = strings.Join(lines, "\n")
		}
		return result
	}

	fieldHeight := lipgloss.Height(floyd)

	// Left field.
	const leftWidth = 6
	leftFieldRow := fg(o.FieldColor, strings.Repeat(diag, leftWidth))
	leftField := new(strings.Builder)
	for range fieldHeight {
		fmt.Fprintln(leftField, leftFieldRow)
	}

	// Right field.
	rightWidth := max(15, o.Width-floydWidth-leftWidth-2) // 2 for the gap.
	const stepDownAt = 0
	rightField := new(strings.Builder)
	for i := range fieldHeight {
		width := rightWidth
		if i >= stepDownAt {
			width = rightWidth - (i - stepDownAt)
		}
		fmt.Fprint(rightField, fg(o.FieldColor, strings.Repeat(diag, width)), "\n")
	}

	// Return the wide version.
	const hGap = " "
	logo := lipgloss.JoinHorizontal(lipgloss.Top, leftField.String(), hGap, floyd, hGap, rightField.String())
	if o.Width > 0 {
		// Truncate the logo to the specified width.
		lines := strings.Split(logo, "\n")
		for i, line := range lines {
			lines[i] = ansi.Truncate(line, o.Width, "")
		}
		logo = strings.Join(lines, "\n")
	}
	return logo
}

// SmallRender renders a smaller version of the Floyd logo, suitable for
// smaller windows or sidebar usage.
func SmallRender(t *styles.Styles, width int) string {
	title := t.Base.Foreground(t.Secondary).Render("LEGACY AI™")
	title = fmt.Sprintf("%s %s", title, styles.ApplyBoldForegroundGrad(t, "FLOYD", t.Secondary, t.Primary))
	remainingWidth := width - lipgloss.Width(title) - 1 // 1 for the space after "FLOYD"
	if remainingWidth > 0 {
		lines := strings.Repeat("╱", remainingWidth)
		title = fmt.Sprintf("%s %s", title, t.Base.Foreground(t.Primary).Render(lines))
	}
	return title
}

// SidebarRender renders a compact block-character version of the Floyd logo
// for the sidebar, with gradient coloring and diagonal field lines.
func SidebarRender(s *styles.Styles, width int, o Opts) string {
	fg := func(c color.Color, str string) string {
		return lipgloss.NewStyle().Foreground(c).Render(str)
	}

	legacyai := fg(o.CharmColor, "Legacy AI")

	artLines := []string{
		"░█▀▀░█░░░█▀█░█░█░█▀▄",
		"░█▀▀░█░░░█░█░░█░░█░█",
		"░▀░░░▀▀▀░▀▀▀░░▀░░▀▀░",
		"░█▀▀░█▀█░█▀▄░█▀▀░░░░",
		"░█░░░█░█░█░█░█▀▀░░░░",
		"░▀▀▀░▀▀▀░▀▀░░▀▀▀░░░░",
	}

	// Pad each art line with diagonal field characters to fill the width.
	fieldChar := diag
	b := new(strings.Builder)
	for _, line := range artLines {
		lineWidth := lipgloss.Width(line)
		padWidth := max(0, width-lineWidth)
		padded := line + strings.Repeat(fieldChar, padWidth)
		fmt.Fprintln(b, styles.ApplyForegroundGrad(s, padded, o.TitleColorA, o.TitleColorB))
	}
	coloredArt := strings.TrimRight(b.String(), "\n")

	// Diagonal field lines above and below.
	field := fg(o.FieldColor, strings.Repeat(diag, width))

	// Also pad the "Legacy AI" line with diagonals.
	legacyaiWidth := lipgloss.Width(legacyai)
	legacyaiPad := max(0, width-legacyaiWidth)
	legacyaiRow := legacyai + fg(o.FieldColor, strings.Repeat(diag, legacyaiPad))

	result := strings.Join([]string{legacyaiRow, field, coloredArt, field}, "\n")

	// Truncate each line to the available width.
	if width > 0 {
		lines := strings.Split(result, "\n")
		for i, line := range lines {
			lines[i] = ansi.Truncate(line, width, "")
		}
		result = strings.Join(lines, "\n")
	}
	return result
}

// renderWord renders letterforms to fork a word. stretchIndex is the index of
// the letter to stretch, or -1 if no letter should be stretched.
func renderWord(spacing int, stretchIndex int, letterforms ...letterform) string {
	if spacing < 0 {
		spacing = 0
	}

	renderedLetterforms := make([]string, len(letterforms))

	// pick one letter randomly to stretch
	for i, letter := range letterforms {
		renderedLetterforms[i] = letter(i == stretchIndex)
	}

	if spacing > 0 {
		// Add spaces between the letters and render.
		renderedLetterforms = slice.Intersperse(renderedLetterforms, strings.Repeat(" ", spacing))
	}
	return strings.TrimSpace(
		lipgloss.JoinHorizontal(lipgloss.Top, renderedLetterforms...),
	)
}

// letterC renders the letter C in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, it defaults to no stretching.
func letterC(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▀
	// █
	//	▀▀▀▀

	left := heredoc.Doc(`
		▄
		█
	`)
	right := heredoc.Doc(`
		▀

		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(right, letterformProps{
			stretch:    stretch,
			width:      4,
			minStretch: 7,
			maxStretch: 12,
		}),
	)
}

// letterH renders the letter H in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, it defaults to no stretching.
func letterH(stretch bool) string {
	// Here's what we're making:
	//
	// █   █
	// █▀▀▀█
	// ▀   ▀

	side := heredoc.Doc(`
		█
		█
		▀`)
	middle := heredoc.Doc(`

		▀
	`)
	return joinLetterform(
		side,
		stretchLetterformPart(middle, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 8,
			maxStretch: 12,
		}),
		side,
	)
}

// letterR renders the letter R in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, it defaults to no stretching.
func letterR(stretch bool) string {
	// Here's what we're making:
	//
	// █▀▀▀▄
	// █▀▀▀▄
	// ▀   ▀

	left := heredoc.Doc(`
		█
		█
		▀
	`)
	center := heredoc.Doc(`
		▀
		▀
	`)
	right := heredoc.Doc(`
		▄
		▄
		▀
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(center, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterSStylized renders the letter S in a stylized way, more so than
// [letterS]. It takes an integer that determines how many cells to stretch the
// letter. If the stretch is less than 1, it defaults to no stretching.
func letterSStylized(stretch bool) string {
	// Here's what we're making:
	//
	// ▄▀▀▀▀▀
	// ▀▀▀▀▀█
	// ▀▀▀▀▀

	left := heredoc.Doc(`
		▄
		▀
		▀
	`)
	center := heredoc.Doc(`
		▀
		▀
		▀
	`)
	right := heredoc.Doc(`
		▀
		█
	`)
	return joinLetterform(
		left,
		stretchLetterformPart(center, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		right,
	)
}

// letterU renders the letter U in a stylized way. It takes an integer that
// determines how many cells to stretch the letter. If the stretch is less than
// 1, it defaults to no stretching.
func letterU(stretch bool) string {
	// Here's what we're making:
	//
	// █   █
	// █   █
	//	▀▀▀

	side := heredoc.Doc(`
		█
		█
	`)
	middle := heredoc.Doc(`


		▀
	`)
	return joinLetterform(
		side,
		stretchLetterformPart(middle, letterformProps{
			stretch:    stretch,
			width:      3,
			minStretch: 7,
			maxStretch: 12,
		}),
		side,
	)
}

func joinLetterform(letters ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, letters...)
}

// letterformProps defines letterform stretching properties.
// for readability.
type letterformProps struct {
	width      int
	minStretch int
	maxStretch int
	stretch    bool
}

// stretchLetterformPart is a helper function for letter stretching. If randomize
// is false the minimum number will be used.
func stretchLetterformPart(s string, p letterformProps) string {
	if p.maxStretch < p.minStretch {
		p.minStretch, p.maxStretch = p.maxStretch, p.minStretch
	}
	n := p.width
	if p.stretch {
		n = cachedRandN(p.maxStretch-p.minStretch) + p.minStretch //nolint:gosec
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = s
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}
