package message

import (
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jrnxf/gh-eco/ui/commands"
	"github.com/jrnxf/gh-eco/ui/context"
	"github.com/jrnxf/gh-eco/utils"
)

type Model struct {
	keys    utils.KeyMap
	Content string
	timer   timer.Model
	ctx     *context.ProgramContext
}

func NewModel() Model {
	return Model{
		keys: utils.Keys,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {

	case commands.SetMessage:
		m.ctx.CurrentView = context.MessageView
		m.Content = msg.Content
		m.timer = timer.NewWithInterval(time.Second*time.Duration(msg.SecondsDisplayed), time.Second)
		cmd = m.timer.Init()

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.ctx.CurrentView = m.ctx.LastView
		m.ctx.LastView = context.VoidView
		m.Content = ""
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	text := lipgloss.NewStyle().Width(m.ctx.Layout.ScreenWidth).PaddingTop(m.ctx.Layout.ScreenHeight / 2).Align(lipgloss.Center).Render(
		// fmt.Sprintf("%s %s", m.Content, m.timer.View()),
		m.Content,
	)
	return text
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	if ctx == nil {
		return
	}
	m.ctx = ctx
}

func (m Model) TriggerMessage(content string, secondsDisplayed int) tea.Cmd {
	return func() tea.Msg {
		return commands.SetMessage{
			Content:          content,
			SecondsDisplayed: secondsDisplayed,
		}
	}
}
