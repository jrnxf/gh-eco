package user

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coloradocolby/ghx/api"
)

type Model struct {
	User api.User
	err  error
}

func NewModel() Model {
	return Model{
		User: api.User{},
		err:  nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case api.SearchUserResponse:
		m.User = msg.User
	}
	return m, cmd
}

func (m Model) View() string {
	return m.User.Name
}
