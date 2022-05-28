package utils

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/coloradocolby/gh-eco/ui/context"
)

type KeyMap struct {
	FocusNext       key.Binding
	FocusPrev       key.Binding
	PreviewPageDown key.Binding
	PreviewPageUp   key.Binding
	ToggleReadme    key.Binding
	OpenGithub      key.Binding
	FocusInput      key.Binding
	Search          key.Binding
	Quit            key.Binding
}

func (k KeyMap) ShortHelp(ctx *context.ProgramContext) []key.Binding {
	bindings := []key.Binding{}
	switch ctx.Mode {
	case context.InsertMode:
		bindings = append(bindings, k.Search)
	case context.NormalMode:
		if ctx.User.Login == "" {
			// user has not yet loaded
			bindings = append(bindings, k.Search)
		} else {
			if ctx.View == context.UserView {
				bindings = append(bindings, k.FocusInput, k.FocusNext, k.FocusPrev, k.ToggleReadme, k.OpenGithub)
			} else {
				bindings = append(bindings, k.FocusNext, k.FocusPrev, k.PreviewPageDown, k.PreviewPageUp, k.ToggleReadme, k.OpenGithub)
			}
		}
	}
	bindings = append(bindings, k.Quit)

	return bindings
}

var Keys = KeyMap{
	FocusPrev: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "move up"),
	),
	FocusNext: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "move down"),
	),
	PreviewPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "preview page up"),
	),
	PreviewPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "preview page down"),
	),
	ToggleReadme: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "toggle readme"),
	),
	OpenGithub: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in github"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Search: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "search"),
	),
	FocusInput: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "focus input"),
	),
}
