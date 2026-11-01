package github

import (
	"testing"

	"github.com/jrnxf/gh-eco/api/github/queries"
)

func Test_MapUserSkipsNonRepositoryPins(t *testing.T) {
	var q queries.User
	q.Login = "octocat"
	q.PinnedItems.Nodes = make([]queries.PinnedItem, 2)
	q.PinnedItems.Nodes[0].Repository.Id = "R_1"
	q.PinnedItems.Nodes[0].Repository.Name = "hello-world"
	// Nodes[1] is a pinned gist, which decodes as an empty Repository

	u := mapUser(q)
	if len(u.PinnedRepos) != 1 || u.PinnedRepos[0].Id != "R_1" {
		t.Fatalf("pinned repos = %+v, want only hello-world", u.PinnedRepos)
	}
}
