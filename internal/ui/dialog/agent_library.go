package dialog

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/agents"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/common"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/list"
	"github.com/CaptainPhantasy/FloydSandyIso/internal/ui/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"
)

const (
	// AgentLibraryID is the identifier for the agent library dialog.
	AgentLibraryID = "agent_library"
	agentLibraryDialogMaxWidth  = 70
	agentLibraryDialogMaxHeight = 20
)

// AgentLibrary represents a dialog for selecting agents from the library.
type AgentLibrary struct {
	com    *common.Common
	help   help.Model
	list   *list.FilterableList
	input  textinput.Model
	agents []agents.AgentDefinition

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// AgentLibraryItem represents an agent list item.
type AgentLibraryItem struct {
	agent   agents.AgentDefinition
	t       *styles.Styles
	m       fuzzy.Match
	cache   map[int]string
	focused bool
}

var (
	_ Dialog   = (*AgentLibrary)(nil)
	_ ListItem = (*AgentLibraryItem)(nil)
)

// NewAgentLibrary creates a new agent library dialog.
func NewAgentLibrary(com *common.Common, agentsDir string) (*AgentLibrary, error) {
	a := &AgentLibrary{com: com}

	help := help.New()
	help.Styles = com.Styles.DialogHelpStyles()
	a.help = help

	a.list = list.NewFilterableList()
	a.list.Focus()

	a.input = textinput.New()
	a.input.SetVirtualCursor(false)
	a.input.Placeholder = "Type to filter agents"
	a.input.SetStyles(com.Styles.TextInput)
	a.input.Focus()

	a.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "select"),
	)
	a.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next item"),
	)
	a.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous item"),
	)
	a.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "choose"),
	)
	a.keyMap.Close = CloseKey

	a.loadAgents(agentsDir)

	return a, nil
}

// loadAgents loads agents from the specified directory.
func (a *AgentLibrary) loadAgents(dir string) {
	loaded, err := agents.LoadAgents(dir)
	if err != nil {
		a.agents = []agents.AgentDefinition{}
	} else {
		a.agents = loaded
	}

	items := make([]list.FilterableItem, 0, len(a.agents))
	for _, agent := range a.agents {
		items = append(items, &AgentLibraryItem{
			agent: agent,
			t:     a.com.Styles,
		})
	}
	a.list.SetItems(items...)
}

// ID implements Dialog.
func (a *AgentLibrary) ID() string {
	return AgentLibraryID
}

// HandleMsg implements Dialog.
func (a *AgentLibrary) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, a.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, a.keyMap.Previous):
			a.list.Focus()
			if a.list.IsSelectedFirst() {
				a.list.SelectLast()
				a.list.ScrollToBottom()
				break
			}
			a.list.SelectPrev()
			a.list.ScrollToSelected()
		case key.Matches(msg, a.keyMap.Next):
			a.list.Focus()
			if a.list.IsSelectedLast() {
				a.list.SelectFirst()
				a.list.ScrollToTop()
				break
			}
			a.list.SelectNext()
			a.list.ScrollToSelected()
		case key.Matches(msg, a.keyMap.Select):
			selectedItem := a.list.SelectedItem()
			if selectedItem == nil {
				break
			}
			agentItem, ok := selectedItem.(*AgentLibraryItem)
			if !ok {
				break
			}
			return ActionSelectAgent{
				AgentName:        agentItem.agent.Name,
				AgentDescription: agentItem.agent.Description,
				SystemPrompt:     agentItem.agent.SystemPrompt,
			}
		default:
			var cmd tea.Cmd
			a.input, cmd = a.input.Update(msg)
			value := a.input.Value()
			a.list.SetFilter(value)
			a.list.ScrollToTop()
			a.list.SetSelected(0)
			return ActionCmd{cmd}
		}
	}
	return nil
}

// Cursor returns the cursor position.
func (a *AgentLibrary) Cursor() *tea.Cursor {
	return InputCursor(a.com.Styles, a.input.Cursor())
}

// Draw implements Dialog.
func (a *AgentLibrary) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := a.com.Styles
	width := max(0, min(agentLibraryDialogMaxWidth, area.Dx()))
	height := max(0, min(agentLibraryDialogMaxHeight, area.Dy()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()

	a.input.SetWidth(innerWidth - t.Dialog.InputPrompt.GetHorizontalFrameSize() - 1)
	a.list.SetSize(innerWidth, height-heightOffset)
	a.help.SetWidth(innerWidth)

	rc := NewRenderContext(t, width)
	rc.Title = "Agent Library"
	inputView := t.Dialog.InputPrompt.Render(a.input.View())
	rc.AddPart(inputView)

	visibleCount := len(a.list.FilteredItems())
	if a.list.Height() >= visibleCount {
		a.list.ScrollToTop()
	} else {
		a.list.ScrollToSelected()
	}

	listView := t.Dialog.List.Height(a.list.Height()).Render(a.list.Render())
	rc.AddPart(listView)
	rc.Help = a.help.View(a)

	view := rc.Render()

	cur := a.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// ShortHelp implements help.KeyMap.
func (a *AgentLibrary) ShortHelp() []key.Binding {
	return []key.Binding{
		a.keyMap.UpDown,
		a.keyMap.Select,
		a.keyMap.Close,
	}
}

// FullHelp implements help.KeyMap.
func (a *AgentLibrary) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{a.keyMap.Select, a.keyMap.Next, a.keyMap.Previous, a.keyMap.Close},
	}
}

// Filter implements ListItem.
func (a *AgentLibraryItem) Filter() string {
	return a.agent.Name + " " + a.agent.Description + " " + a.agent.Trigger
}

// ID implements ListItem.
func (a *AgentLibraryItem) ID() string {
	return a.agent.Name
}

// SetFocused implements ListItem.
func (a *AgentLibraryItem) SetFocused(focused bool) {
	if a.focused != focused {
		a.cache = nil
	}
	a.focused = focused
}

// SetMatch implements ListItem.
func (a *AgentLibraryItem) SetMatch(match fuzzy.Match) {
	a.cache = nil
	a.m = match
}

// Render implements ListItem.
func (a *AgentLibraryItem) Render(width int) string {
	styles := ListItemStyles{
		ItemBlurred:     a.t.Dialog.NormalItem,
		ItemFocused:     a.t.Dialog.SelectedItem,
		InfoTextBlurred: a.t.Subtle,
		InfoTextFocused: a.t.Base,
	}
	return renderItem(styles, a.agent.Name, a.agent.Description, a.focused, width, a.cache, &a.m)
}
