package context

import "github.com/coloradocolby/gh-eco/ui/models"

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type View int

const (
	UserView View = iota
	RepoView
)

type ProgramContext struct {
	User             models.User
	View             View
	Mode             Mode
	CurrentFocus     CurrentFocus
	FocusableWidgets []FocusableWidget
	Layout           struct {
		ScreenHeight  int
		ScreenWidth   int
		ContentHeight int
		ContentWidth  int
	}
}

type FocusableWidget struct {
	Type string
	Info struct {
		Url      string
		Owner    string
		RepoName string
	}
}

type CurrentFocus struct {
	FocusIdx      int
	FocusedWidget FocusableWidget
}
