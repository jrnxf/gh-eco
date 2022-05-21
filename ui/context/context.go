package context

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type FocusableWidgetState struct {
	isActive bool
	display  string
}

type ProgramContext struct {
	ScreenHeight     int
	ScreenWidth      int
	ContentHeight    int
	ContentWidth     int
	Mode             Mode
	FocusableWidgets FocusableWidgets
}

type FocusableWidgets struct {
	Search FocusableWidgetState
}
