package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v52/github"
	"golang.org/x/oauth2"
)

type gitService struct {
	dest   string
	repo   *git.Repository
	auth   *http.BasicAuth
	w      *git.Worktree
	client *github.Client
}

func New(url, username, password string) IService {
	repo := strings.Split(url, "/")[4]
	if repo == "" {
		fmt.Println("invalid repository url")
		return nil
	}
	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}
	dest := "/tmp/" + repo
	r, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: password},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &gitService{
		dest:   dest,
		repo:   r,
		auth:   auth,
		client: github.NewClient(tc),
	}
}

func (g *gitService) Dest() string {
	return g.dest
}

// CreateBranch creates a new branch from the current HEAD, and check out to the new branch
func (g *gitService) CreateBranch(branchName string) (err error) {
	if g.repo == nil {
		return errors.New("repository is not initialized")
	}

	headRef, err := g.repo.Head()
	if err != nil {
		return err
	}

	branchRefName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(branchRefName, headRef.Hash())
	if err := g.repo.Storer.SetReference(ref); err != nil {
		return err
	}

	w, err := g.repo.Worktree()
	if err != nil {
		return err
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Branch: branchRefName,
		Force:  true,
	}); err != nil {
		return err
	}

	g.w = w
	return nil
}

func (g *gitService) Commit(message string) (err error) {
	if g.w == nil {
		return errors.New("worktree is not initialized")
	}

	if err := g.w.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		return err
	}

	_, err = g.w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.auth.Username,
			Email: "quanglm.ops@gmail.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *gitService) Push() (err error) {
	if g.repo == nil {
		return errors.New("repository is not initialized")
	}
	if g.auth == nil {
		return errors.New("auth is not initialized")
	}
	return g.repo.Push(&git.PushOptions{Auth: g.auth})
}

func (g *gitService) CreatePullRequest(owner, repo, head, base, title, body string) (*int, error) {
	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}

	pr, _, err := g.client.PullRequests.Create(context.Background(), owner, repo, newPR)
	if err != nil {
		return nil, err
	}

	return pr.Number, nil
}

func (g *gitService) RequestReview(owner, repo string, pullRequestNumber int, reviewers []string) error {
	_, _, err := g.client.PullRequests.RequestReviewers(context.Background(), owner, repo, pullRequestNumber, github.ReviewersRequest{
		Reviewers: reviewers,
	})
	if err != nil {
		return err
	}

	return nil
}
