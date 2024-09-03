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
