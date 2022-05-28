package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/coloradocolby/gh-eco/api/github"
	"github.com/coloradocolby/gh-eco/ui/commands"
	"github.com/coloradocolby/gh-eco/ui/components/help"
	"github.com/coloradocolby/gh-eco/ui/components/pager"
	"github.com/coloradocolby/gh-eco/ui/components/search"
	"github.com/coloradocolby/gh-eco/ui/components/user"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/utils"
)

type Model struct {
	keys   utils.KeyMap
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
		pager:  pager.NewModel(),
		help:   help.NewModel(),
		ctx: context.ProgramContext{
			Mode: context.InsertMode,
		},
	}

	m.resetWidgets()
	m.resetCurrentFocus()

	m.syncProgramContext()

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd             tea.Cmd
		spinnerCmd      tea.Cmd
		searchCmd       tea.Cmd
		userCmd         tea.Cmd
		helpCmd         tea.Cmd
		pagerCmd        tea.Cmd
		getReadmeCmd    tea.Cmd
		focusChangeCmd  tea.Cmd
		layoutChangeCmd tea.Cmd
		cmds            []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			cmd = tea.Quit
		}
		switch m.ctx.Mode {
		case context.NormalMode:
			info := m.ctx.CurrentFocus.FocusedWidget.Info

			switch {
			case key.Matches(msg, m.keys.FocusNext):
				focusChangeCmd = m.notifyFocusNextWidget()
			case key.Matches(msg, m.keys.FocusPrev):
				focusChangeCmd = m.notifyFocusPrevWidget()
			case key.Matches(msg, m.keys.OpenGithub):
				utils.BrowserOpen(info.Url)
			case key.Matches(msg, m.keys.ToggleReadme):
				if m.ctx.View == context.UserView {
					getReadmeCmd = github.GetReadme(info.RepoName, info.Owner)
				} else {
					m.ctx.View = context.UserView
					m.onLayoutChange()
					layoutChangeCmd = m.notifyLayoutChange()
				}

			}
		}

	case commands.GetReadmeResponse:
		m.ctx.View = context.RepoView
		m.onLayoutChange()
		layoutChangeCmd = m.notifyLayoutChange()

	case commands.GetUserResponse:
		m.resetWidgets()
	case commands.FocusChange:
		m.resetWidgets()

	case tea.WindowSizeMsg:
		m.onWindowSizeChanged(msg)
		m.syncProgramContext()
	}

	m.syncProgramContext()

	m.search, searchCmd = m.search.Update(msg)
	m.pager, pagerCmd = m.pager.Update(msg)
	m.user, userCmd = m.user.Update(msg)
	m.help, helpCmd = m.help.Update(msg)
	cmds = append(cmds, cmd, spinnerCmd, searchCmd, userCmd, helpCmd, pagerCmd, getReadmeCmd, focusChangeCmd, layoutChangeCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	switch m.ctx.View {
	case context.RepoView:
		return lipgloss.JoinVertical(lipgloss.Left,
			m.pager.View(),
			m.help.View(),
		)
	case context.UserView:
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Render(m.search.View()),
			m.user.View(),
			m.help.View(),
		)
	default:
		return ""
	}
}

func (m *Model) notifyFocusNextWidget() tea.Cmd {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx + 1) % numWidgets
	cf.FocusedWidget = m.ctx.FocusableWidgets[cf.FocusIdx]

	return func() tea.Msg {
		return commands.FocusChange{}
	}
}

func (m *Model) notifyFocusPrevWidget() tea.Cmd {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx - 1 + numWidgets) % numWidgets
	cf.FocusedWidget = m.ctx.FocusableWidgets[cf.FocusIdx]

	return func() tea.Msg {
		return commands.FocusChange{}
	}
}

func (m Model) notifyLayoutChange() tea.Cmd {
	return func() tea.Msg {
		return commands.LayoutChange{}
	}
}

func (m *Model) onWindowSizeChanged(msg tea.WindowSizeMsg) {
	m.ctx.Layout.ScreenHeight = msg.Height
	m.ctx.Layout.ScreenWidth = msg.Width

	m.ctx.Layout.ContentHeight = msg.Height - lipgloss.Height(m.search.View()) - lipgloss.Height(m.help.View())
	m.ctx.Layout.ContentWidth = msg.Width
}

func (m *Model) onLayoutChange() {
	log.Println("onLayoutChange")
	contentHeight := m.ctx.Layout.ScreenHeight - lipgloss.Height(m.help.View())

	if m.ctx.View == context.UserView {
		log.Println("UserView")
		contentHeight -= lipgloss.Height(m.search.View())
	} else {
		log.Println("RepoView")
	}
	log.Println(m.ctx.Layout.ScreenHeight - contentHeight)
	m.ctx.Layout.ContentHeight = contentHeight
	m.syncProgramContext()
}

func (m *Model) resetWidgets() {
	m.ctx.FocusableWidgets = []context.FocusableWidget{
		{
			Type: "NoFocus",
		},
	}
}

func (m *Model) resetCurrentFocus() {
	m.ctx.CurrentFocus = context.CurrentFocus{
		FocusIdx: 0,
		FocusedWidget: context.FocusableWidget{
			Type: "NoFocus",
		},
	}
}

func (m *Model) syncProgramContext() {
	m.pager.UpdateProgramContext(&m.ctx)
	m.search.UpdateProgramContext(&m.ctx)
	m.user.UpdateProgramContext(&m.ctx)
	m.help.UpdateProgramContext(&m.ctx)
}
