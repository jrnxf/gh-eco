package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/atotto/clipboard"
	"github.com/coloradocolby/ghx/ui/components/pager"
	"github.com/coloradocolby/ghx/ui/components/search"
	"github.com/coloradocolby/ghx/ui/components/user"
	"github.com/coloradocolby/ghx/ui/context"
	"github.com/coloradocolby/ghx/utils"
)

var (
	// colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
)

type Model struct {
	keys     utils.KeyMap
	err      error
	search   search.Model
	pager    pager.Model
	user     user.Model
	ctx      context.ProgramContext
	focusIdx int
}

func New() Model {
	m := Model{
		keys:   utils.Keys,
		search: search.NewModel(),
		pager:  pager.NewModel(),
		user:   user.NewModel(),
		ctx: context.ProgramContext{
			Mode: context.InsertMode,
			FocusableWidgets: context.FocusableWidgets{
				Search: context.FocusableWidgetState{},
			},
		},
		focusIdx: 0,
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
		cmd        tea.Cmd
		spinnerCmd tea.Cmd
		pagerCmd   tea.Cmd
		searchCmd  tea.Cmd
		userCmd    tea.Cmd
		cmds       []tea.Cmd
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
				m.SetFocus(m.focusIdx + 1)
			case key.Matches(msg, m.keys.Up):
				m.SetFocus(m.focusIdx - 1)
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

	}

	m.syncProgramContext()

	m.search, searchCmd = m.search.Update(msg)
	m.pager, pagerCmd = m.pager.Update(msg)
	m.user, userCmd = m.user.Update(msg)
	cmds = append(cmds, cmd, pagerCmd, spinnerCmd, searchCmd, userCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

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

func (m *Model) SetFocus(newIdx int) {
	m.focusIdx = newIdx
}

func (m *Model) syncProgramContext() {
	m.pager.UpdateProgramContext(&m.ctx)
	m.search.UpdateProgramContext(&m.ctx)
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

// iter := 1

// log.Println("__login")
// login := Widget{
// 	focusable: true,
// 	isActive:  true,
// }
// log.Println("__login2")

// login.display = boldActive.Render(m.displayedUser.Login)
// // iter++
// log.Println("__name")

// name := Widget{
// 	focusable: true,
// 	isActive:  true,
// 	display:   boldActive.Render(m.displayedUser.Name),
// }
// // iter++
// log.Println("__location")

// location := Widget{
// 	focusable: true,
// 	isActive:  true,
// 	display:   boldActive.Render(m.displayedUser.Location),
// }

// m.widgets = append(m.widgets, search, login, name, location)
// m.widgets = append(m.widgets, search)

// var activeWidget Widget

// for _, value := range m.widgets {
// 	if value.isActive {
// 		activeWidget = value
// 		break
// 	}
// }
