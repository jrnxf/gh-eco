package commands

import "github.com/coloradocolby/gh-eco/ui/models"

type FocusChange struct{}

type LayoutChange struct{}

type ProcessArgs struct {
	Username string
}

type ProgramInitMsg struct {
	Ready bool
}

type GetUserResponse struct {
	Err  error
	User models.User
}

type GetReadmeResponse struct {
	Err    error
	Readme models.Blob
}

type StarStarrableResponse struct {
	Err       error
	Starrable models.Starrable
}

type RemoveStarStarrableResponse struct {
	Err       error
	Starrable models.Starrable
}
