package git

type IService interface {
	Dest() string
	CreateBranch(branchName string) (err error)
	Commit(message string) (err error)
	Push() (err error)
	CreatePullRequest(owner, repo, head, base, title, body string) (*int, error)
	RequestReview(owner, repo string, pullRequestNumber int, reviewers []string) error
}
