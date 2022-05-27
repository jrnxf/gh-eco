package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/coloradocolby/gh-eco/api"
	"github.com/coloradocolby/gh-eco/ui/components/help"
	"github.com/coloradocolby/gh-eco/ui/components/pager"
	"github.com/coloradocolby/gh-eco/ui/components/search"
	"github.com/coloradocolby/gh-eco/ui/components/user"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/utils"
)

type Model struct {
	keys   utils.KeyMap
	err    error
	search search.Model
	user   user.Model
	pager  pager.Model
	help   help.Model
	ctx    context.ProgramContext
}

func New() Model {
	m := Model{
		keys:   utils.Keys,
		search: search.NewModel(),
		user:   user.NewModel(),
		help:   help.NewModel(),
		pager:  pager.NewModel(),
		ctx: context.ProgramContext{
			Mode: context.InsertMode,
			FocusableWidgets: []context.FocusableWidget{
				{
					Type: "NoFocus",
				},
			},
			CurrentFocus: context.CurrentFocus{
				FocusIdx: 0,
				FocusedWidget: context.FocusableWidget{
					Type: "NoFocus",
				},
			},
		},
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
		textinput.Blink,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd            tea.Cmd
		spinnerCmd     tea.Cmd
		searchCmd      tea.Cmd
		focusChangeCmd tea.Cmd
		userCmd        tea.Cmd
		helpCmd        tea.Cmd
		pagerCmd       tea.Cmd
		getReadmeCmd   tea.Cmd
		cmds           []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			cmd = tea.Quit
		}
		switch m.ctx.Mode {
		case context.NormalMode:
			switch {
			case key.Matches(msg, m.keys.FocusNext):
				focusChangeCmd = m.FocusNext()
			case key.Matches(msg, m.keys.FocusPrev):
				focusChangeCmd = m.FocusPrev()
			case key.Matches(msg, m.keys.OpenGithub):
				utils.BrowserOpen(m.ctx.CurrentFocus.FocusedWidget.Repo.Url)
			case key.Matches(msg, m.keys.PreviewReadme):
				getReadmeCmd = api.GetReadme(m.ctx.CurrentFocus.FocusedWidget.Repo.Name, m.ctx.CurrentFocus.FocusedWidget.Repo.Owner)
			case key.Matches(msg, m.keys.ExitReadme):
				m.ctx.View = context.UserView
			}
		}

	case api.GetUserResponse:
	case context.FocusChange:
		m.ctx.FocusableWidgets = []context.FocusableWidget{
			{
				Type: "NoFocus",
			},
		}
	}

	m.syncProgramContext()

	m.search, searchCmd = m.search.Update(msg)
	m.pager, pagerCmd = m.pager.Update(msg)
	m.user, userCmd = m.user.Update(msg)
	m.help, helpCmd = m.help.Update(msg)
	cmds = append(cmds, cmd, spinnerCmd, searchCmd, userCmd, helpCmd, pagerCmd, getReadmeCmd, focusChangeCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	if m.ctx.View == context.RepoView {
		return m.pager.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().PaddingTop(1).Render(m.search.View()),
		m.user.View(),
		// m.help.View(),
		// m.pager.View(),
	)
}

func (m *Model) FocusNext() tea.Cmd {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx + 1) % numWidgets
	cf.FocusedWidget = m.ctx.FocusableWidgets[cf.FocusIdx]

	return func() tea.Msg {
		return context.FocusChange{}
	}
}

func (m *Model) FocusPrev() tea.Cmd {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx - 1 + numWidgets) % numWidgets
	cf.FocusedWidget = m.ctx.FocusableWidgets[cf.FocusIdx]

	return func() tea.Msg {
		return context.FocusChange{}
	}
}

func (m *Model) syncProgramContext() {
	m.pager.UpdateProgramContext(&m.ctx)
	m.search.UpdateProgramContext(&m.ctx)
	m.user.UpdateProgramContext(&m.ctx)
	m.help.UpdateProgramContext(&m.ctx)
}
