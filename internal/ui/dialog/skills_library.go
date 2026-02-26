package dialog

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"

	"github.com/CaptainPhantasy/FloydSandyIso/internal/skills"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/common"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/list"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
)

const (
	// SkillsLibraryID is the identifier for the skills library dialog.
	SkillsLibraryID              = "skills_library"
	skillsLibraryDialogMaxWidth  = 70
	skillsLibraryDialogMaxHeight = 20
)

// SkillsLibrary represents a dialog for selecting skills from the library.
type SkillsLibrary struct {
	com    *common.Common
	help   help.Model
	list   *list.FilterableList
	input  textinput.Model
	skills []*skills.Skill

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// SkillsLibraryItem represents a skill list item.
type SkillsLibraryItem struct {
	skill   *skills.Skill
	t       *styles.Styles
	m       fuzzy.Match
	cache   map[int]string
	focused bool
}

var (
	_ Dialog   = (*SkillsLibrary)(nil)
	_ ListItem = (*SkillsLibraryItem)(nil)
)

// NewSkillsLibrary creates a new skills library dialog.
func NewSkillsLibrary(com *common.Common, skillsDir string) (*SkillsLibrary, error) {
	s := &SkillsLibrary{com: com}

	help := help.New()
	help.Styles = com.Styles.DialogHelpStyles()
	s.help = help

	s.list = list.NewFilterableList()
	s.list.Focus()

	s.input = textinput.New()
	s.input.SetVirtualCursor(false)
	s.input.Placeholder = "Type to filter skills"
	s.input.SetStyles(com.Styles.TextInput)
	s.input.Focus()

	s.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "select"),
	)
	s.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next item"),
	)
	s.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous item"),
	)
	s.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "choose"),
	)
	s.keyMap.Close = CloseKey

	s.loadSkills(skillsDir)

	return s, nil
}

// loadSkills loads skills from the specified directory.
func (s *SkillsLibrary) loadSkills(dir string) {
	s.skills = skills.Discover([]string{dir})

	items := make([]list.FilterableItem, 0, len(s.skills))
	for _, skill := range s.skills {
		items = append(items, &SkillsLibraryItem{
			skill: skill,
			t:     s.com.Styles,
		})
	}
	s.list.SetItems(items...)
}

// ID implements Dialog.
func (s *SkillsLibrary) ID() string {
	return SkillsLibraryID
}

// HandleMsg implements Dialog.
func (s *SkillsLibrary) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, s.keyMap.Previous):
			s.list.Focus()
			if s.list.IsSelectedFirst() {
				s.list.SelectLast()
				s.list.ScrollToBottom()
				break
			}
			s.list.SelectPrev()
			s.list.ScrollToSelected()
		case key.Matches(msg, s.keyMap.Next):
			s.list.Focus()
			if s.list.IsSelectedLast() {
				s.list.SelectFirst()
				s.list.ScrollToTop()
				break
			}
			s.list.SelectNext()
			s.list.ScrollToSelected()
		case key.Matches(msg, s.keyMap.Select):
			selectedItem := s.list.SelectedItem()
			if selectedItem == nil {
				break
			}
			skillItem, ok := selectedItem.(*SkillsLibraryItem)
			if !ok {
				break
			}
			return ActionSelectSkill{
				SkillName:        skillItem.skill.Name,
				SkillDescription: skillItem.skill.Description,
				SkillContent:     skillItem.skill.Instructions,
				SkillCategory:    skillItem.skill.Path,
			}
		default:
			var cmd tea.Cmd
			s.input, cmd = s.input.Update(msg)
			value := s.input.Value()
			s.list.SetFilter(value)
			s.list.ScrollToTop()
			s.list.SetSelected(0)
			return ActionCmd{cmd}
		}
	}
	return nil
}

// Cursor returns the cursor position.
func (s *SkillsLibrary) Cursor() *tea.Cursor {
	return InputCursor(s.com.Styles, s.input.Cursor())
}

// Draw implements Dialog.
func (s *SkillsLibrary) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := s.com.Styles
	width := max(0, min(skillsLibraryDialogMaxWidth, area.Dx()))
	height := max(0, min(skillsLibraryDialogMaxHeight, area.Dy()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()

	s.input.SetWidth(innerWidth - t.Dialog.InputPrompt.GetHorizontalFrameSize() - 1)
	s.list.SetSize(innerWidth, height-heightOffset)
	s.help.SetWidth(innerWidth)

	rc := NewRenderContext(t, width)
	rc.Title = "Skills Library"
	inputView := t.Dialog.InputPrompt.Render(s.input.View())
	rc.AddPart(inputView)

	visibleCount := len(s.list.FilteredItems())
	if s.list.Height() >= visibleCount {
		s.list.ScrollToTop()
	} else {
		s.list.ScrollToSelected()
	}

	listView := t.Dialog.List.Height(s.list.Height()).Render(s.list.Render())
	rc.AddPart(listView)
	rc.Help = s.help.View(s)

	view := rc.Render()

	cur := s.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// ShortHelp implements help.KeyMap.
func (s *SkillsLibrary) ShortHelp() []key.Binding {
	return []key.Binding{
		s.keyMap.UpDown,
		s.keyMap.Select,
		s.keyMap.Close,
	}
}

// FullHelp implements help.KeyMap.
func (s *SkillsLibrary) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{s.keyMap.Select, s.keyMap.Next, s.keyMap.Previous, s.keyMap.Close},
	}
}

// Filter implements ListItem.
func (s *SkillsLibraryItem) Filter() string {
	return s.skill.Name + " " + s.skill.Description + " " + s.skill.Path
}

// ID implements ListItem.
func (s *SkillsLibraryItem) ID() string {
	return s.skill.Name
}

// SetFocused implements ListItem.
func (s *SkillsLibraryItem) SetFocused(focused bool) {
	if s.focused != focused {
		s.cache = nil
	}
	s.focused = focused
}

// SetMatch implements ListItem.
func (s *SkillsLibraryItem) SetMatch(match fuzzy.Match) {
	s.cache = nil
	s.m = match
}

// Render implements ListItem.
func (s *SkillsLibraryItem) Render(width int) string {
	styles := ListItemStyles{
		ItemBlurred:     s.t.Dialog.NormalItem,
		ItemFocused:     s.t.Dialog.SelectedItem,
		InfoTextBlurred: s.t.Subtle,
		InfoTextFocused: s.t.Base,
	}

	// Show category (parent directory) alongside name
	displayName := s.skill.Name
	if s.skill.Path != "" {
		displayName = s.skill.Name
	}

	return renderItem(styles, displayName, s.skill.Description, s.focused, width, s.cache, &s.m)
}
