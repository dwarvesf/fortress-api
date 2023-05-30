package employee

import (
	"fmt"
	"mime/multipart"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UploadAvatarInput struct {
	ID string
}

func (r *controller) UploadAvatar(uuidUserID model.UUID, file *multipart.FileHeader, params UploadAvatarInput) (string, error) {
	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	fileType := "image"
	filePath := fmt.Sprintf("https://storage.googleapis.com/%s/employees/%s/images/%s", r.config.Google.GCSBucketName, params.ID, fileName)
	gcsPath := fmt.Sprintf("employees/%s/images/%s", params.ID, fileName)

	// 2.1 validate
	if !fileExtension.ImageValid() {
		return "", ErrInvalidFileExtension
	}
	if fileExtension == model.ContentExtensionJpg || fileExtension == model.ContentExtensionPng {
		if fileSize > model.MaxFileSizeImage {
			return "", ErrInvalidFileSize
		}
	}

	tx, done := r.repo.NewTransaction()

	// 2.2 check employee existed
	emp, err := r.store.Employee.One(tx.DB(), params.ID, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", done(ErrEmployeeNotFound)
		}
		return "", done(err)
	}

	// 2.3 check file name exist
	_, err = r.store.Content.OneByPath(tx.DB(), filePath)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", done(err)
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		// not found => create and upload content to GCS
		_, err = r.store.Content.Create(tx.DB(), model.Content{
			Type:      fileType,
			Extension: fileExtension.String(),
			Path:      filePath,
			TargetID:  emp.ID,
			UploadBy:  uuidUserID,
		})
		if err != nil {
			return "", done(err)
		}

		multipart, err := file.Open()
		if err != nil {
			return "", done(err)
		}

		err = r.service.Google.UploadContentGCS(multipart, gcsPath)
		if err != nil {
			return "", done(err)
		}
	}

	// 3. update avatar field
	_, err = r.store.Employee.UpdateSelectedFieldsByID(tx.DB(), emp.ID.String(), model.Employee{
		Avatar: filePath,
	}, "avatar")
	if err != nil {
		return "", done(err)
	}

	return filePath, done(nil)
}
