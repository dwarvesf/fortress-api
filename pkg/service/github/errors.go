package github

import "errors"

var (
	ErrFailedToGetGithubAccount  = errors.New("failed to get github account")
	ErrFoundOneMoreGithubAccount = errors.New("failed to get github account due to more than 1 github account found")
)
