package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jrnxf/gh-eco/api/github"
	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/components/help"
	"github.com/jrnxf/gh-eco/ui/components/markdown"
	"github.com/jrnxf/gh-eco/ui/components/message"
	"github.com/jrnxf/gh-eco/ui/components/search"
	"github.com/jrnxf/gh-eco/ui/components/user"
	"github.com/jrnxf/gh-eco/ui/context"
	"github.com/jrnxf/gh-eco/utils"
)

type Model struct {
	keys     utils.KeyMap
	search   search.Model
	user     user.Model
	markdown markdown.Model
	help     help.Model
	message  message.Model
	ctx      context.ProgramContext
}

func New() Model {
	m := Model{
		keys:     utils.Keys,
		search:   search.NewModel(),
		user:     user.NewModel(),
		markdown: markdown.NewModel(),
		help:     help.NewModel(),
		message:  message.NewModel(),
		ctx: context.ProgramContext{
			Mode:        context.InsertMode,
			CurrentView: context.UserView,
			LastView:    context.VoidView,
		},
	}

	m.resetWidgets()
	m.resetCurrentFocus()

	m.syncProgramContext()

	return m
}

func (m Model) Init() tea.Cmd {
	var (
		cmds []tea.Cmd
	)

	cmds = append(cmds, textinput.Blink)
	return tea.Batch(
		cmds...,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd             tea.Cmd
		spinnerCmd      tea.Cmd
		searchCmd       tea.Cmd
		userCmd         tea.Cmd
		helpCmd         tea.Cmd
		messageCmd      tea.Cmd
		markdownCmd     tea.Cmd
		getReadmeCmd    tea.Cmd
		starRepoCmd     tea.Cmd
		unstarRepoCmd   tea.Cmd
		focusChangeCmd  tea.Cmd
		layoutChangeCmd tea.Cmd
		cmds            []tea.Cmd
	)
	fw := &m.ctx.CurrentFocus.FocusedWidget

	switch msg := msg.(type) {

	case tea.KeyMsg:

		if key.Matches(msg, m.keys.Quit) {
			cmd = tea.Quit
		}

		if m.ctx.Mode == context.NormalMode && m.ctx.CurrentView != context.MessageView {

			switch fw.Type {
			case context.UserWidget:

				switch {
				case key.Matches(msg, m.keys.OpenGithub):
					utils.BrowserOpen(fw.User.Url)
				}

			case context.RepoWidget:

				switch {
				case key.Matches(msg, m.keys.OpenGithub):
					utils.BrowserOpen(fw.Repo.Url)

				case key.Matches(msg, m.keys.StarRepo):
					if fw.Repo.ViewerHasStarred {
						unstarRepoCmd = github.RemoveStarStarrable(fw.Repo.Id)
						cmds = append(cmds, unstarRepoCmd)
					} else {
						starRepoCmd = github.StarStarrable(fw.Repo.Id)
						cmds = append(cmds, starRepoCmd)
					}
				}
			}

			if m.ctx.CurrentView == context.UserView {

				switch {
				case key.Matches(msg, m.keys.FocusNext):
					m.focusNextWidget()
					focusChangeCmd = m.notifyFocusChange()

				case key.Matches(msg, m.keys.FocusPrev):
					m.focusPrevWidget()
					focusChangeCmd = m.notifyFocusChange()

				case key.Matches(msg, m.keys.FocusInput):
					if m.ctx.CurrentView == context.UserView {
						m.resetCurrentFocus()
						focusChangeCmd = m.notifyFocusChange()
					}
				}
			}

			if key.Matches(msg, m.keys.ToggleReadme) && (fw.Type == context.UserWidget || fw.Type == context.RepoWidget) {

				switch m.ctx.CurrentView {
				case context.UserView:
					switch fw.Type {
					case context.UserWidget:
						// get the focused users personal readme
						getReadmeCmd = github.GetReadme(fw.User.Login, fw.User.Login)

					case context.RepoWidget:
						// get the focused repos readme
						getReadmeCmd = github.GetReadme(fw.Repo.Name, fw.Repo.Owner.Login)
					}

				case context.ReadmeView:
					m.ctx.CurrentView = context.UserView
					m.onLayoutChange()
					layoutChangeCmd = m.notifyLayoutChange()
				}
			}

			if key.Matches(msg, m.keys.StarGhEco) {
				starRepoCmd = github.StarStarrable(github.GH_ECO_REPO_ID)
				messageCmd = m.message.TriggerMessage("tysm 🥹", 2)
				m.ctx.LastView = m.ctx.CurrentView
				cmds = append(cmds, messageCmd, starRepoCmd)
			}
		}

	case commands.GetReadmeResponse:
		m.ctx.CurrentView = context.ReadmeView
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
	m.markdown, markdownCmd = m.markdown.Update(msg)
	m.user, userCmd = m.user.Update(msg)
	m.help, helpCmd = m.help.Update(msg)
	m.message, messageCmd = m.message.Update(msg)
	cmds = append(cmds, cmd, spinnerCmd, searchCmd, userCmd, helpCmd, messageCmd, markdownCmd, getReadmeCmd, focusChangeCmd, layoutChangeCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	if m.message.Content != "" {
		return m.message.View()
	}

	switch m.ctx.CurrentView {
	case context.ReadmeView:
		return lipgloss.JoinVertical(lipgloss.Left,
			m.markdown.View(),
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

func (m *Model) focusNextWidget() {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx + 1) % numWidgets
	cf.FocusedWidget = m.ctx.FocusableWidgets[cf.FocusIdx]
}

func (m *Model) focusPrevWidget() {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx - 1 + numWidgets) % numWidgets
	cf.FocusedWidget = m.ctx.FocusableWidgets[cf.FocusIdx]
}

func (m Model) notifyFocusChange() tea.Cmd {
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
	contentHeight := m.ctx.Layout.ScreenHeight - lipgloss.Height(m.help.View())

	if m.ctx.CurrentView == context.UserView {
		contentHeight -= lipgloss.Height(m.search.View())
	}

	m.ctx.Layout.ContentHeight = contentHeight
	m.syncProgramContext()
}

func (m *Model) resetWidgets() {
	m.ctx.FocusableWidgets = []context.FocusableWidget{
		{
			Descriptor: "NoFocus",
		},
	}
}

func (m *Model) resetCurrentFocus() {
	m.ctx.CurrentFocus = context.CurrentFocus{
		FocusIdx: 0,
		FocusedWidget: context.FocusableWidget{
			Descriptor: "NoFocus",
			Type:       context.NoWidget,
		},
	}
}

func (m *Model) syncProgramContext() {
	m.markdown.UpdateProgramContext(&m.ctx)
	m.search.UpdateProgramContext(&m.ctx)
	m.user.UpdateProgramContext(&m.ctx)
	m.help.UpdateProgramContext(&m.ctx)
	m.message.UpdateProgramContext(&m.ctx)
}
