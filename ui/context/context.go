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
	ReadmeView
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

type FocusedWidgetType int

const (
	NoWidget FocusedWidgetType = iota
	UserWidget
	RepoWidget
)

type FocusableWidget struct {
	Descriptor string
	Type       FocusedWidgetType
	Repo       models.Repo
	User       models.User
}

type CurrentFocus struct {
	FocusIdx      int
	FocusedWidget FocusableWidget
}
