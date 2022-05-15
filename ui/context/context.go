package context

type Mode int

const (
	InsertMode Mode = iota
	NormalMode
)

type ProgramContext struct {
	ScreenHeight  int
	ScreenWidth   int
	ContentHeight int
	ContentWidth  int
	Mode          Mode
}
