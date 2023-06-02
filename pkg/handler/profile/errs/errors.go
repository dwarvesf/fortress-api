package errs

import "errors"

var (
	ErrEmployeeNotFound          = errors.New("employee not found")
	ErrCountryNotFound           = errors.New("country not found")
	ErrInvalidCountryOrCity      = errors.New("invalid country or city")
	ErrInvalidFileExtension      = errors.New("invalid file extension")
	ErrInvalidDocumentType       = errors.New("invalid document type")
	ErrInvalidFileType           = errors.New("invalid file type")
	ErrInvalidFileSize           = errors.New("invalid file size")
	ErrFileAlreadyExisted        = errors.New("file already existed")
	ErrEmailExisted              = errors.New("email already exists")
	ErrOnboardingFormAlreadyDone = errors.New("onboarding form already done")
	ErrMissingDocuments          = errors.New("missing id/passport documents")
	ErrInvalidDate               = errors.New("invalid date")
	ErrInvalidDiscordMemberInfo  = errors.New("invalid discord member info")
)
