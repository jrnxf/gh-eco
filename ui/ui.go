package ui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atotto/clipboard"
	"github.com/coloradocolby/gh-eco/api"
	"github.com/coloradocolby/gh-eco/ui/components/pager"
	"github.com/coloradocolby/gh-eco/ui/components/search"
	"github.com/coloradocolby/gh-eco/ui/components/user"
	"github.com/coloradocolby/gh-eco/ui/context"
	"github.com/coloradocolby/gh-eco/utils"
)

var (
	// colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
)

type Model struct {
	keys   utils.KeyMap
	err    error
	search search.Model
	pager  pager.Model
	user   user.Model
	ctx    context.ProgramContext
}

func New() Model {
	m := Model{
		keys:   utils.Keys,
		search: search.NewModel(),
		pager:  pager.NewModel(),
		user:   user.NewModel(),
		ctx: context.ProgramContext{
			Mode: context.InsertMode,
			FocusableWidgets: []context.FocusableWidget{
				{
					Name: "NoFocus",
				},
			},
			CurrentFocus: context.CurrentFocus{
				FocusIdx: 0,
				FocusedWidget: context.FocusableWidget{
					Name: "NoFocus",
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
		pagerCmd       tea.Cmd
		searchCmd      tea.Cmd
		focusChangeCmd tea.Cmd
		userCmd        tea.Cmd
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
			case key.Matches(msg, m.keys.Down):
				log.Println(m.ctx.FocusableWidgets, m.ctx.CurrentFocus.FocusIdx, m.ctx.CurrentFocus.FocusedWidget.Name)

				focusChangeCmd = m.FocusNext()
			case key.Matches(msg, m.keys.Up):
				focusChangeCmd = m.FocusPrev()
			case key.Matches(msg, m.keys.Copy):
				m.copyActiveWidgetToClipboard()
			}
		}

	// when doing paging
	// case api.SearchUserResponse:
	// 	val, _ := json.MarshalIndent(msg.User, "", "    ")
	// 	in := "```json\n" + string(val) + "\n```"
	// 	out, _ := glamour.Render(in, "dark")
	// 	m.pager.Viewport.SetContent(out)

	case api.SearchUserResponse:
	case context.FocusChange:
		m.ctx.FocusableWidgets = []context.FocusableWidget{
			{
				Name: "NoFocus",
			},
		}
	}

	m.syncProgramContext()

	m.search, searchCmd = m.search.Update(msg)
	m.pager, pagerCmd = m.pager.Update(msg)
	m.user, userCmd = m.user.Update(msg)
	cmds = append(cmds, cmd, pagerCmd, spinnerCmd, searchCmd, userCmd, focusChangeCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	// log.Println("------------")
	// log.Println(m.ctx.FocusableWidgets)
	// log.Println(m.ctx.CurrentFocus.FocusedWidget.Name)
	// log.Println("------------")

	// When doing paging
	// if m.user.User.Login != "" {
	// 	return lipgloss.JoinVertical(lipgloss.Left,
	// 		m.search.View(),
	// 		m.pager.View(),
	// 	)
	// } else {
	// 	return lipgloss.JoinVertical(lipgloss.Left,
	// 		m.search.View(),
	// 	)
	// }

	return lipgloss.JoinVertical(lipgloss.Left,
		m.search.View(),
		m.user.View(),
	)

}

func (m *Model) FocusNext() tea.Cmd {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx + 1) % numWidgets
	cf.FocusedWidget.Name = m.ctx.FocusableWidgets[cf.FocusIdx].Name

	return func() tea.Msg {
		return context.FocusChange{}
	}
}

func (m *Model) FocusPrev() tea.Cmd {
	cf := &m.ctx.CurrentFocus

	numWidgets := len(m.ctx.FocusableWidgets)
	cf.FocusIdx = (cf.FocusIdx - 1 + numWidgets) % numWidgets
	cf.FocusedWidget.Name = m.ctx.FocusableWidgets[cf.FocusIdx].Name

	return func() tea.Msg {
		return context.FocusChange{}
	}
}

func (m *Model) syncProgramContext() {
	m.pager.UpdateProgramContext(&m.ctx)
	m.search.UpdateProgramContext(&m.ctx)
	m.user.UpdateProgramContext(&m.ctx)
}

func (m Model) copyActiveWidgetToClipboard() {
	clipboard.WriteAll("todo")
	// fmt.Println(len(m.widgets))
	// for _, value := range m.widgets {
	// 	if value.isActive {
	// 		if err := clipboard.WriteAll("colby"); err != nil {
	// 			panic(err)
	// 		}
	// 		os.Exit(0)
	// 		break
	// 	}
	// }
}
