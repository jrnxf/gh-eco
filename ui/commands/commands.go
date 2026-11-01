// Package commands holds the messages exchanged between the root model and
// its components.
package commands

import "github.com/jrnxf/gh-eco/models"

type SetMessage struct {
	Content          string
	SecondsDisplayed int
}

// SearchSubmitted asks the root model to look up Login, or the viewer's
// own profile when Login is empty.
type SearchSubmitted struct {
	Login string
}

type GetUserResponse struct {
	// Req identifies the lookup; responses to superseded lookups are dropped.
	Req   int
	Login string
	User  models.User
	Err   error
}

type GetPeopleResponse struct {
	// Req identifies the request; responses to superseded requests are dropped.
	Req  int
	Page models.PeoplePage
	Err  error
}

type GetReadmeResponse struct {
	// Req identifies the request; responses to superseded requests are dropped.
	Req    int
	Readme string
	Err    error
}

type StarResponse struct {
	// Id is the starrable the toggle was for, set even when it failed.
	Id string
	// Unstar is true when the request removed a star.
	Unstar bool
	// Prev is whether the viewer had starred it before the toggle, which a
	// failure restores.
	Prev      bool
	Starrable models.Starrable
	Err       error
}
