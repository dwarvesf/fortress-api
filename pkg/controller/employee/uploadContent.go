package employee

import (
	"fmt"
	"mime/multipart"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type UploadContentInput struct {
	ID string
}

func (r *controller) UploadContent(file *multipart.FileHeader, params UploadContentInput) (*model.Content, error) {
	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	filePath := "employees/" + params.ID
	fileType := ""

	// 2.1 validate
	if !fileExtension.Valid() {
		return nil, ErrInvalidFileExtension
	}
	if fileExtension == model.ContentExtensionJpg || fileExtension == model.ContentExtensionPng {
		if fileSize > model.MaxFileSizeImage {
			return nil, ErrInvalidFileSize
		}
		filePath = filePath + "/images"
		fileType = "image"
	}
	if fileExtension == model.ContentExtensionPdf {
		if fileSize > model.MaxFileSizePdf {
			return nil, ErrInvalidFileSize
		}
		filePath = filePath + "/docs"
		fileType = "document"
	}
	filePath = filePath + "/" + fileName

	tx, done := r.repo.NewTransaction()

	// 2.2 check file name exist
	_, err := r.store.Content.OneByPath(tx.DB(), filePath)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, done(err)
	}
	if err == nil {
		return nil, done(ErrFileAlreadyExisted)
	}

	// 2.3 check employee existed
	emp, err := r.store.Employee.One(tx.DB(), params.ID, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, done(ErrEmployeeNotFound)
		}
		return nil, done(err)
	}

	content, err := r.store.Content.Create(tx.DB(), model.Content{
		Type:      fileType,
		Extension: fileExtension.String(),
		Path:      fmt.Sprintf("https://storage.googleapis.com/%s/%s", r.config.Google.GCSBucketName, filePath),
		TargetID:  emp.ID,
		UploadBy:  emp.ID,
	})
	if err != nil {
		return nil, done(err)
	}

	multipart, err := file.Open()
	if err != nil {
		return nil, done(err)
	}

	// 3. Upload to GCS
	err = r.service.Google.UploadContentGCS(multipart, filePath)
	if err != nil {
		return nil, done(err)
	}

	return content, done(nil)
}
