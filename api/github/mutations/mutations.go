package mutations

type AddStarMutation struct {
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

type FollowUserMutation struct {
	FollowUser struct {
		User struct {
			Id                string
			ViewerIsFollowing bool
			Followers         struct {
				TotalCount int
			}
		}
	} `graphql:"followUser(input: {userId: $userId})"`
}

type UnfollowUserMutation struct {
	UnfollowUser struct {
		User struct {
			Id                string
			ViewerIsFollowing bool
			Followers         struct {
				TotalCount int
			}
		}
	} `graphql:"unfollowUser(input: {userId: $userId})"`
}
