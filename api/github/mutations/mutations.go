package mutations

type StarMutation struct {
	AddStar struct {
		Starrable struct {
			Id               string
			StargazerCount   int
			ViewerHasStarred bool
		}
	} `graphql:"addStar(input: {starrableId: $starrableId})"`
}

type RemoveStarMutation struct {
	RemoveStar struct {
		Starrable struct {
			Id               string
			StargazerCount   int
			ViewerHasStarred bool
		}
	} `graphql:"removeStar(input: {starrableId: $starrableId})"`
}
