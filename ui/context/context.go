package context

import "github.com/coloradocolby/gh-eco/types/display"

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

type FocusableWidget struct {
	Type string
	Repo struct {
		Url   string
		Owner string
		Name  string
	}
}

type ProgramContext struct {
	// ScreenHeight     int
	// ScreenWidth      int
	// ContentHeight    int
	// ContentWidth     int
	User             display.User
	View             View
	Mode             Mode
	FocusableWidgets []FocusableWidget
	CurrentFocus     CurrentFocus
}

type CurrentFocus struct {
	FocusIdx      int
	FocusedWidget FocusableWidget
}

type FocusChange struct{}
