package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/coloradocolby/ghx/api"
	"github.com/coloradocolby/ghx/ui/components/pager"
	"github.com/coloradocolby/ghx/ui/components/spinner"
	"github.com/coloradocolby/ghx/ui/components/tabs"
	"github.com/coloradocolby/ghx/ui/components/textinput"
	"github.com/coloradocolby/ghx/ui/context"
	"github.com/coloradocolby/ghx/utils"
)

type Model struct {
	keys          utils.KeyMap
	err           error
	tabs          tabs.Model
	textInput     textinput.Model
	pager         pager.Model
	spinner       spinner.Model
	ctx           context.ProgramContext
	displayedUser api.User
	fetching      bool
}

func New() Model {
	m := Model{
		keys:          utils.Keys,
		tabs:          tabs.NewModel(),
		textInput:     textinput.NewModel(),
		spinner:       spinner.NewModel(),
		pager:         pager.NewModel(),
		displayedUser: api.User{},
		ctx: context.ProgramContext{
			Mode: context.InsertMode,
		},
		fetching: false,
	}

	return m
}

type initMsg struct {
	ready bool
}

func initScreen() tea.Msg {
	return initMsg{ready: true}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		initScreen,
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd           tea.Cmd
		spinnerCmd    tea.Cmd
		pagerCmd      tea.Cmd
		textInputCmd  tea.Cmd
		searchUserCmd tea.Cmd
		cmds          []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			cmd = tea.Quit
		}
		switch m.ctx.Mode {
		case context.InsertMode:
			switch {
			case key.Matches(msg, m.keys.Search):
				m.ctx.Mode = context.NormalMode

				m.fetching = true
				searchUserCmd = api.SearchUser(m.textInput.TextInput.Value())
			}

			m.textInput, textInputCmd = m.textInput.Update(msg)

		case context.NormalMode:
			switch {
			case key.Matches(msg, m.keys.FocusInput):
				// m.textInput.TextInput.Reset()
				m.ctx.Mode = context.InsertMode
			case key.Matches(msg, m.keys.PrevTab):
				m.tabs.PrevTab()
			case key.Matches(msg, m.keys.NextTab):
				m.tabs.NextTab()
			}

		}
	case api.SearchUserResponse:
		m.fetching = false
		m.displayedUser = msg.User

		if m.displayedUser.Login != "" {
			// s := strings.Builder{}
			// s.WriteString(fmt.Sprintf("Login: %s\n", m.displayedUser.Login))
			// s.WriteString(fmt.Sprintf("Name: %s\n", m.displayedUser.Name))
			// s.WriteString(fmt.Sprintf("Location: %s\n", m.displayedUser.Location))
			// s.WriteString(fmt.Sprintf("Bio: %s\n", m.displayedUser.Bio))
			// s.WriteString(fmt.Sprintf("Company: %s\n", m.displayedUser.Company))
			// m.pager.Viewport.SetContent(s.String())

			userJson, _ := json.MarshalIndent(m.displayedUser, "", "    ")
			md := fmt.Sprintf("\n# %s\n%s\n```json\n%s\n```", m.displayedUser.Name, m.displayedUser.WebsiteUrl, string(userJson))
			out, _ := glamour.Render(md, "dark")
			m.pager.Viewport.SetContent(out)
		}

	}
	m.syncProgramContext()

	m.pager, pagerCmd = m.pager.Update(msg)
	m.spinner, spinnerCmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd, pagerCmd, spinnerCmd, textInputCmd, searchUserCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	s := strings.Builder{}
	if m.fetching {
		s.WriteString(fmt.Sprintf("%s %s\n", m.textInput.View(), m.spinner.View()))
	} else {
		s.WriteString(m.textInput.View() + "\n")
	}

	s.WriteString(m.pager.View())
	return s.String()
}

func (m *Model) syncProgramContext() {
	m.pager.UpdateProgramContext(&m.ctx)
}
