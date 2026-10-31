// Package commands holds the messages exchanged between the root model and
// its components.
package commands

import "github.com/jrnxf/gh-eco/models"

type SetMessage struct {
	Content          string
	SecondsDisplayed int
}

// SearchSubmitted asks the root model to look up Login.
type SearchSubmitted struct {
	Login string
}

type GetUserResponse struct {
	// Req identifies the lookup; responses to superseded lookups are dropped.
	Req  int
	User models.User
	Err  error
}

type GetReadmeResponse struct {
	Readme string
	Err    error
}

type StarResponse struct {
	// Unstar is true when the request removed a star.
	Unstar    bool
	Starrable models.Starrable
	Err       error
}
