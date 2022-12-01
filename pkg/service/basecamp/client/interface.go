package client

import "net/http"

type ClientService interface {
	Get(url string) (resp *http.Response, err error)
	Do(req *http.Request) (resp *http.Response, err error)
	GetAccessToken(code string, redirectURI string) (accessToken string, err error)
}
