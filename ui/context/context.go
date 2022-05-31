package context

import "github.com/coloradocolby/gh-eco/ui/models"

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type View int

const (
	VoidView View = iota
	UserView
	ReadmeView
	MessageView
)

type ProgramContext struct {
	User             models.User
	CurrentView      View
	LastView         View
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
