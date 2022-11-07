package google

type GoogleService interface {
	GetLoginURL() string
	GetAccessToken(code string, redirectURL string) (accessToken string, err error)
	GetGoogleEmail(accessToken string) (email string, err error)
}
