package mutations

type AddStarMutation struct {
	AddStar struct {
		Starrable struct {
			Id               string
			StargazerCount   int
			ViewerHasStarred bool
		}
	} `graphql:"addStar(input: $input)"`
}

type RemoveStarMutation struct {
	RemoveStar struct {
		Starrable struct {
			Id               string
			StargazerCount   int
			ViewerHasStarred bool
		}
	} `graphql:"removeStar(input: $input)"`
}

// AddStarInput and RemoveStarInput are named after their GraphQL input
// types, which the client derives the variable declaration from.
type AddStarInput struct {
	StarrableID string `json:"starrableId"`
}

type RemoveStarInput struct {
	StarrableID string `json:"starrableId"`
}
