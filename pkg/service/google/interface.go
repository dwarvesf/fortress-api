package google

import "mime/multipart"

type IService interface {
	GetLoginURL() string
	GetAccessToken(code string, redirectURL string) (accessToken string, err error)
	GetGoogleEmail(accessToken string) (email string, err error)
	UploadContentGCS(file multipart.File, fileName string) error
}
