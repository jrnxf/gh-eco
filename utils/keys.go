package utils

import (
	"github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
	FocusNext       key.Binding
	FocusPrev       key.Binding
	PreviewPageDown key.Binding
	PreviewPageUp   key.Binding
	ToggleReadme    key.Binding
	OpenGithub      key.Binding
	StarRepo        key.Binding
	StarGhEco       key.Binding
	FocusInput      key.Binding
	Search          key.Binding
	Quit            key.Binding
}

var Keys = KeyMap{
	FocusPrev: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "move up"),
	),
	FocusNext: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "move down"),
	),
	PreviewPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "page up"),
	),
	PreviewPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "page down"),
	),
	ToggleReadme: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "toggle readme"),
	),
	OpenGithub: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in github"),
	),
	StarRepo: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "star repo"),
	),
	StarGhEco: key.NewBinding(
		key.WithKeys("!"),
		key.WithHelp("!", "star gh-eco"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
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
