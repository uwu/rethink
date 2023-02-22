package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type config struct {
	path string
	User string
	Key  string
}

var exitErr error

var (
	headerStyle      = lipgloss.NewStyle().Align(lipgloss.Center)
	inputStyle       = lipgloss.NewStyle().Margin(0, 2).Border(lipgloss.NormalBorder())
	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))
	helpStyle  = lipgloss.NewStyle().Margin(0, 3)
	setupStyle = lipgloss.NewStyle().Margin(1, 2)
)

var setupKeymap = struct {
	move, submit key.Binding
}{
	move: key.NewBinding(
		key.WithKeys("shift+tab", "tab"),
		key.WithHelp("[shift+]tab", "previous/next"),
	),
	submit: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
}

type keymap = struct {
	scrollUp, scrollDown, setup, submit, quit key.Binding
}

type model struct {
	setup       bool
	config      config
	width       int
	height      int
	thought     textarea.Model
	setupInputs []textinput.Model
	setupIndex  int
	help        help.Model
	keymap      keymap
}

func initialModel(config config) model {
	thought := textarea.New()
	thought.Placeholder = "think of something..."
	thought.Prompt = ""
	thought.FocusedStyle.Base = inputStyle
	thought.BlurredStyle.Base = inputStyle
	thought.FocusedStyle.EndOfBuffer = endOfBufferStyle
	thought.BlurredStyle.EndOfBuffer = endOfBufferStyle
	thought.Focus()

	m := model{
		setup:       config.User != "" && config.Key != "",
		config:      config,
		thought:     thought,
		setupInputs: make([]textinput.Model, 2),
		help:        help.New(),
		keymap: keymap{
			scrollUp: key.NewBinding(
				key.WithKeys("pgup"),
				key.WithHelp("pgup", "newer thoughts"),
			),
			scrollDown: key.NewBinding(
				key.WithKeys("pgdown"),
				key.WithHelp("pgdown", "older thoughts"),
			),
			setup: key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp("ctrl+e", "open setup"),
			),
			submit: key.NewBinding(
				key.WithKeys("ctrl+s"),
				key.WithHelp("ctrl+s", "submit thought"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+q", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}

	setupUsername := textinput.New()
	setupKey := textinput.New()

	setupUsername.SetValue(config.User)
	setupKey.SetValue(config.Key)

	setupUsername.Placeholder = "Username"
	setupUsername.Focus()

	setupKey.Placeholder = "Thought key"
	setupKey.CharLimit = 36
	setupKey.EchoMode = textinput.EchoPassword

	m.setupInputs[0] = setupUsername
	m.setupInputs[1] = setupKey

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, m.keymap.quit) {
			return m, tea.Quit
		}
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height
	}

	if !m.setup {
		return m.updateSetup(msg)
	}
	return m.updateMain(msg)
}

func (m model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.setupIndex == len(m.setupInputs)-1 {
				m.config.User = m.setupInputs[0].Value()
				m.config.Key = m.setupInputs[1].Value()
				content, err := json.Marshal(m.config)
				if err != nil {
					fmt.Printf("Oh no! %v\n", err)
					os.Exit(1)
				}
				os.WriteFile(m.config.path, content, 0644)
				m.setup = true
				return m, nil
			}

			if s == "shift+tab" {
				m.setupIndex--
			} else {
				m.setupIndex++
			}

			if m.setupIndex > len(m.setupInputs)-1 {
				m.setupIndex = 0
			} else if m.setupIndex < 0 {
				m.setupIndex = len(m.setupInputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.setupInputs))
			for i := 0; i <= len(m.setupInputs)-1; i++ {
				if i == m.setupIndex {
					// Set focused state
					cmds[i] = m.setupInputs[i].Focus()
					continue
				}
				// Remove focused state
				m.setupInputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateSetupInputs(msg)
	return m, cmd
}

func (m model) updateSetupInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.setupInputs))

	for i := range m.setupInputs {
		m.setupInputs[i], cmds[i] = m.setupInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.setup):
			m.setup = false
		case key.Matches(msg, m.keymap.submit):
			err := putThought(m.config, m.thought.Value())
			if err != nil {
				// make a proper error display
				exitErr = err
				return m, tea.Quit
			}
			m.thought.SetValue("")
		}
	}

	m.resizeElements()
	m.updateKeybinds()

	m.thought, cmd = m.thought.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) resizeElements() {
	headerStyle.Width(m.width)
	m.thought.SetWidth(m.width)
}

func (m *model) updateKeybinds() {
}

func (m model) View() string {
	if !m.setup {
		return setupStyle.Render(m.setupView())
	}
	return m.mainView()
}

func (m model) setupView() string {
	s := ""

	s += "rethink setup\n\n"

	for i := range m.setupInputs {
		s += m.setupInputs[i].View() + "\n"
	}

	help := m.help.ShortHelpView([]key.Binding{
		setupKeymap.move,
		setupKeymap.submit,
		m.keymap.quit,
	})

	s += "\n" + help

	return s
}

func (m model) mainView() string {
	s := ""

	s += headerStyle.Render("rethink: it's all in your head.") + "\n\n"
	s += m.thought.View() + "\n"
	s += "\n   pretend there's your pevious thoughts here"

	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.submit,
		// m.keymap.scrollUp,
		// m.keymap.scrollDown,
		m.keymap.setup,
		m.keymap.quit,
	})

	// Overwrite empty space to circumvent text from previous updates being there when resizing
	emptySpace := max(m.height-strings.Count(s, "\n")-1, 0)
	s += strings.Repeat("\n", emptySpace)

	s += helpStyle.Render(help)
	return s
}

func putThought(config config, content string) error {
	client := http.Client{}
	body := bytes.NewReader([]byte(content))
	req, err := http.NewRequest(http.MethodPut, "https://rethink.uwu.network/api/think", body)
	if err != nil {
		return err
	}
	req.Header.Add("name", config.User)
	req.Header.Add("authorization", config.Key)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated {
		return errors.New("could not create thought")
	}
	return nil
}

func main() {
	configPath, err := xdg.ConfigFile("rethink/conifg.json")
	if err != nil {
		fmt.Printf("Oh no! %v\n", err)
		os.Exit(1)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		content = []byte("{}")
	}

	var config config
	err = json.Unmarshal(content, &config)
	if err != nil {
		fmt.Printf("Oh no! %v\n", err)
		os.Exit(1)
	}

	config.path = configPath

	p := tea.NewProgram(initialModel(config), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Oh no! %v\n", err)
		os.Exit(1)
	}

	if exitErr != nil {
		fmt.Printf("an error occured: %v\n", exitErr)
	}
}

// how does go not have this function what
func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
