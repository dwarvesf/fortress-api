package content

import (
	"fmt"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/content/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/content/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

// UploadContent godoc
// @Summary Upload content of employee by id
// @Description Upload content of employee by id
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param id path string true "Employee ID"
// @Param Authorization header string true "jwt token"
// @Param file formData file true "content upload"
// @Success 200 {object} view.EmployeeContentDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /upload-content [post]
func (h *handler) UploadContent(c *gin.Context) {
	// 1.1 parse id from uri, validate id
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	var req request.UploadContentRequest
	err = c.Bind(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	//if err := c.ShouldBindUri(&params); err != nil {
	//	c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, params, ""))
	//	return
	//}

	// 1.2 get upload file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, file, ""))
		return
	}

	// 1.3 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "employee",
		"method":  "UploadContent",
		"params":  req,
		// "body":    body,
	})

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	filePath := ""
	fileType := ""
	switch req.Module {
	case request.UploadModuleEmployees:
		if req.EmployeeID != "" {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEmployeeIDRequired, file, ""))
			return
		}

		filePath = fmt.Sprintf("%s/%s", request.UploadModuleEmployees, req.EmployeeID)
	}

	// 2.1 validate
	if !fileExtension.Valid() {
		l.Info("invalid file extension")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}
	if fileExtension == model.ContentExtensionJpg || fileExtension == model.ContentExtensionPng {
		if fileSize > model.MaxFileSizeImage {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
			return
		}
		filePath = filePath + "/images"
		fileType = "image"
	}
	if fileExtension == model.ContentExtensionPdf {
		if fileSize > model.MaxFileSizePdf {
			l.Info("invalid file size")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
			return
		}
		filePath = filePath + "/docs"
		fileType = "document"
	}
	filePath = filePath + "/" + fileName

	tx, done := h.repo.NewTransaction()

	// 2.2 check file name exist
	_, err = h.store.Content.GetByPath(tx.DB(), filePath)
	if err != nil && err != gorm.ErrRecordNotFound {
		l.Error(err, "error query content from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	if err == nil {
		l.Info("file already existed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, errs.ErrFileAlreadyExisted, nil, ""))
		done(errs.ErrFileAlreadyExisted)
		return
	}

	// 2.3 check employee existed
	employee, err := h.store.Employee.One(tx.DB(), params.ID, false)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			done(err)
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	uploadedByUserID, err := model.UUIDFromString(userID)

	content, err := h.store.Content.Create(tx.DB(), model.Content{
		Type:       fileType,
		Extension:  fileExtension.String(),
		Path:       fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.config.Google.GCSBucketName, filePath),
		EmployeeID: employee.ID,
		UploadBy:   uploadedByUserID,
	})
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	multipart, err := file.Open()
	if err != nil {
		l.Error(err, "error in open file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}

	// 3. Upload to GCS
	err = h.service.Google.UploadContentGCS(multipart, filePath)
	if err != nil {
		l.Error(err, "error in upload file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		done(err)
		return
	}
	done(nil)

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData("content.Path"), nil, nil, nil, ""))
}
