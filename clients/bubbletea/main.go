package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

var (
	mainStyle         = lipgloss.NewStyle().Margin(0, 2)
	mainHeaderStyle   = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)
	setupHeaderStyle  = lipgloss.NewStyle().Width(11 + 1 + 32 + 4).AlignHorizontal(lipgloss.Center)
	inputHeaderStyle  = lipgloss.NewStyle().Width(11).AlignHorizontal(lipgloss.Right).Foreground(lipgloss.Color("7"))
	errStyle          = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center).Background(lipgloss.Color("9")).Foreground(lipgloss.Color("0"))
	thoughtInputStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder())
	endOfBufferStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	thoughtStyle      = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	thoughtDateStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type keymap struct {
	// Setup
	nextSetupField, prevSetupField, setupSubmit key.Binding
	// Main
	newerThoughts, olderThoughts, submitThought, openSetup key.Binding
	// Global
	quit key.Binding
}

const (
	name = iota
	thoughtKey
)

type appState = int

const (
	loadingState appState = iota
	setupState
	mainState
)

type model struct {
	setupInputs  []textinput.Model
	setupFocused int
	thoughtInput textarea.Model
	thoughts     []Thought
	thoughtsView viewport.Model
	currentState appState
	config       Config
	keys         keymap
	help         help.Model
	err          string
	errTag       int
	width        int
	height       int
}

func initialModel() model {
	inputs := make([]textinput.Model, 2)

	inputs[name] = textinput.New()
	inputs[name].Placeholder = "user"
	inputs[name].Prompt = ""
	inputs[name].Focus()

	inputs[thoughtKey] = textinput.New()
	inputs[thoughtKey].Placeholder = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	inputs[thoughtKey].CharLimit = 36
	inputs[thoughtKey].Prompt = ""

	thoughtInput := textarea.New()
	thoughtInput.Placeholder = "think of something..."
	thoughtInput.Prompt = ""
	thoughtInput.FocusedStyle.Base = thoughtInputStyle
	thoughtInput.BlurredStyle.Base = thoughtInputStyle
	thoughtInput.FocusedStyle.EndOfBuffer = endOfBufferStyle
	thoughtInput.BlurredStyle.EndOfBuffer = endOfBufferStyle
	thoughtInput.CharLimit = 0
	thoughtInput.Focus()

	m := model{
		setupInputs:  inputs,
		setupFocused: 0,
		thoughtInput: thoughtInput,
		currentState: loadingState,
		thoughtsView: viewport.New(0, 0),
		keys: keymap{
			nextSetupField: key.NewBinding(
				key.WithKeys("tab", "ctrl+n"),
				key.WithHelp("tab", "next field"),
			),
			prevSetupField: key.NewBinding(
				key.WithKeys("shift+tab", "ctrl+p"),
				key.WithHelp("shift+tab", "prev field"),
			),
			setupSubmit: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "submit"),
			),
			newerThoughts: key.NewBinding(
				key.WithKeys("pgup"),
				key.WithHelp("pgup", "newer thoughts"),
			),
			olderThoughts: key.NewBinding(
				key.WithKeys("pgdown"),
				key.WithHelp("pgdown", "older thoughts"),
			),
			submitThought: key.NewBinding(
				key.WithKeys("ctrl+s"),
				key.WithHelp("ctrl+s", "submit thought"),
			),
			openSetup: key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp("ctrl+e", "open setup"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c", "ctrl+q"),
				key.WithHelp("ctrl+c", "quit"),
			),
		},
		help: help.New(),
	}

	m.thoughtsView.KeyMap = viewport.KeyMap{
		PageDown: m.keys.olderThoughts,
		PageUp:   m.keys.newerThoughts,
	}

	return m
}

type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type hideErrMsg struct{ tag int }

func hideErrorAfterTime(tag int) tea.Cmd {
	return tea.Tick(time.Second*3, func(t time.Time) tea.Msg {
		return hideErrMsg{tag}
	})
}

type Config struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type configMsg struct {
	config Config
}

type configSavedMsg struct{}

func loadConfig() tea.Msg {
	filePath, err := xdg.ConfigFile("rethink/config.json")
	if err != nil {
		return errMsg{err}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		content = []byte("{}")
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return errMsg{err}
	}

	return configMsg{config}
}

func saveConfig(config Config) tea.Cmd {
	return func() tea.Msg {
		filePath, err := xdg.ConfigFile("rethink/config.json")
		if err != nil {
			return errMsg{err}
		}

		content, err := json.Marshal(config)
		if err != nil {
			return errMsg{err}
		}

		err = os.WriteFile(filePath, content, 0644)
		if err != nil {
			return errMsg{err}
		}

		return configSavedMsg{}
	}
}

type thoughtSubmittedMsg struct{}

func submitThought(content string, name string, key string) tea.Cmd {
	return func() tea.Msg {
		err := PutThought(content, name, key)
		if err != nil {
			return errMsg{err}
		}
		return thoughtSubmittedMsg{}
	}
}

type thoughtsMsg struct {
	thoughts []Thought
}

func loadThoughts(name string) tea.Cmd {
	return func() tea.Msg {
		thoughts, err := GetThoughts(name)
		if err != nil {
			return errMsg{err}
		}
		return thoughtsMsg{thoughts}
	}
}

func updateThoughts(thoughts []Thought) tea.Cmd {
	return func() tea.Msg {
		return thoughtsMsg{thoughts}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadConfig, cursor.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.quit) {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case configMsg:
		m.config = msg.config
		m.setupInputs[name].SetValue(m.config.Name)
		m.setupInputs[thoughtKey].SetValue(m.config.Key)
		if m.config.Key == "" || m.config.Name == "" {
			m.currentState = setupState
			return m, nil
		}
		m.currentState = mainState
		return m, loadThoughts(m.config.Name)
	case configSavedMsg:
		m.currentState = mainState
		return m, loadThoughts(m.config.Name)
	case errMsg:
		m.thoughtInput.Focus()
		m.err = msg.Error()
		m.errTag++
		cmds = append(cmds, hideErrorAfterTime(m.errTag))
	case hideErrMsg:
		if msg.tag == m.errTag {
			m.err = ""
		}
	case thoughtsMsg:
		m.thoughts = msg.thoughts
		m.thoughtsView.SetContent(m.renderThoughts())
	}

	m.resizeElements()

	switch m.currentState {
	case setupState:
		m, cmd = m.updateSetup(msg)
	case mainState:
		m, cmd = m.updateMain(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) updateSetup(msg tea.Msg) (model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.setupInputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.setupSubmit):
			if m.setupFocused == len(m.setupInputs)-1 {
				m.config.Name = m.setupInputs[name].Value()
				m.config.Key = m.setupInputs[thoughtKey].Value()
				m.currentState = loadingState
				return m, saveConfig(m.config)
			}
			m.nextSetupField()
		case key.Matches(msg, m.keys.prevSetupField):
			m.prevSetupField()
		case key.Matches(msg, m.keys.nextSetupField):
			m.nextSetupField()
		}
		for i := range m.setupInputs {
			m.setupInputs[i].Blur()
		}
		m.setupInputs[m.setupFocused].Focus()
	}

	for i := range m.setupInputs {
		m.setupInputs[i], cmds[i] = m.setupInputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) nextSetupField() {
	m.setupFocused = (m.setupFocused + 1) % len(m.setupInputs)
}

func (m *model) prevSetupField() {
	m.setupFocused--
	if m.setupFocused < 0 {
		m.setupFocused = len(m.setupInputs) - 1
	}
}

func (m model) updateMain(msg tea.Msg) (model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.openSetup):
			m.currentState = setupState
			return m, nil
		case key.Matches(msg, m.keys.submitThought):
			m.thoughtInput.Blur()
			return m, submitThought(m.thoughtInput.Value(), m.config.Name, m.config.Key)
		}
	case thoughtSubmittedMsg:
		thoughts := append([]Thought{{
			Content: m.thoughtInput.Value(),
			Date:    time.Now(),
		}}, m.thoughts...)
		m.thoughtsView.GotoTop()
		m.thoughtInput.SetValue("")
		m.thoughtInput.Focus()
		return m, updateThoughts(thoughts)
	}

	m.thoughtInput, cmd = m.thoughtInput.Update(msg)
	cmds = append(cmds, cmd)

	m.thoughtsView, cmd = m.thoughtsView.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) resizeElements() {
	errStyle.Width(m.width)
	mainHeaderStyle.Width(m.width - 4)

	thoughtStyle.Width(m.width - 4)

	m.thoughtInput.SetWidth(m.width - 4)
	m.thoughtsView.Width = m.width - 4
	m.thoughtsView.Height = m.height - 14

	if m.err != "" {
		m.thoughtsView.Height = m.height - 15
	}
}

func (m model) View() string {
	var s string

	switch m.currentState {
	case loadingState:
		s = "Loading..."
	case setupState:
		s = m.setupView()
	case mainState:
		s = m.mainView()
	}

	notice := ""
	if m.err != "" {
		notice = errStyle.Render(m.err) + "\n"
	}

	return notice + mainStyle.Render(s)
}

func (m model) setupView() string {
	return fmt.Sprintf("%s\n\n%s %s\n%s %s\n\n%s",
		setupHeaderStyle.Render("rethink setup"),
		inputHeaderStyle.Render("Username"),
		m.setupInputs[name].View(),
		inputHeaderStyle.Render("Thought key"),
		m.setupInputs[thoughtKey].View(),
		m.help.FullHelpView([][]key.Binding{
			{m.keys.nextSetupField, m.keys.prevSetupField, m.keys.setupSubmit},
			{m.keys.quit},
		}),
	)
}

func (m model) mainView() string {
	s := ""

	s += mainHeaderStyle.Render("rethink: it's all in your head.") + "\n"
	s += m.thoughtInput.View() + "\n\n"

	// s += fmt.Sprintf("pretend there's a list of %d thoughts here\n\n", len(m.thoughts))

	s += m.thoughtsView.View() + "\n\n"

	s += m.help.FullHelpView([][]key.Binding{
		{m.keys.newerThoughts, m.keys.olderThoughts},
		{m.keys.submitThought, m.keys.openSetup, m.keys.quit},
	})

	return s
}

func (m model) renderThoughts() string {
	s := ""

	// there's probably a string builder or smth i should use for this
	for i, th := range m.thoughts {
		t := ""
		if th.Content != "" {
			t += wordwrap.String(th.Content, m.width-4) + "\n"
		}
		t += thoughtDateStyle.Render(th.Date.Format("Mon Jan 02 15:04:05 2006"))
		if i != len(m.thoughts)-1 {
			s += thoughtStyle.Render(t) + "\n"
		} else {
			s += t
		}
	}

	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
