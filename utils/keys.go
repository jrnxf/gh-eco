package utils

import (
	"github.com/charmbracelet/bubbles/key"
)

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type KeyMap struct {
	FocusNext       key.Binding
	FocusPrev       key.Binding
	PreviewPageDown key.Binding
	PreviewPageUp   key.Binding
	// Copy            key.Binding
	ExitReadme    key.Binding
	PreviewReadme key.Binding
	OpenGithub    key.Binding
	FocusInput    key.Binding
	Search        key.Binding
	Help          key.Binding
	Quit          key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.FocusNext, k.FocusPrev},
		{k.FocusInput, k.Search},
		{k.OpenGithub, k.PreviewReadme},
		{k.PreviewPageDown, k.PreviewPageUp},
		{k.Help, k.Quit},
	}
}

var Keys = KeyMap{
	FocusPrev: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	FocusNext: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	PreviewPageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "preview page up"),
	),
	PreviewPageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "preview page down"),
	),
	// Copy: key.NewBinding(
	// 	key.WithKeys("c"),
	// 	key.WithHelp("c", "copy to clipboard"),
	// ),
	ExitReadme: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "exit readme"),
	),
	PreviewReadme: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "preview readme"),
	),
	OpenGithub: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "open in github"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
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
