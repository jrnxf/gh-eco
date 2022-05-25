package context

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type FocusableWidget struct {
	Name string
	Type string
	Url  string
}

type ProgramContext struct {
	// ScreenHeight     int
	// ScreenWidth      int
	// ContentHeight    int
	// ContentWidth     int
	Mode             Mode
	FocusableWidgets []FocusableWidget
	CurrentFocus     CurrentFocus
}

type CurrentFocus struct {
	FocusIdx      int
	FocusedWidget FocusableWidget
}

type FocusChange struct{}
