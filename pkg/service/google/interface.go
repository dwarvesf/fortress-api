package google

import (
	"io"
)

type IService interface {
	GetLoginURL() string
	GetAccessToken(code string, redirectURL string) (accessToken string, err error)
	GetGoogleEmailLegacy(accessToken string) (email string, err error)
	GetGoogleEmail(accessToken string) (email string, err error)
	UploadContentGCS(file io.Reader, fileName string) error
}
