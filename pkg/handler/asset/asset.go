package asset

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/asset/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/authutils"
	"github.com/dwarvesf/fortress-api/pkg/view"
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

// Upload godoc
// @Summary Upload the content
// @Description Upload the content
// @Tags Asset
// @Accept json
// @Produce json
// @Param Authorization header string true "jwt token"
// @Param file formData file true "content upload"
// @Param type formData string true "image/doc"
// @Param targetType formData string true "employees/projects/change-logs/invoices"
// @Param targetID formData string false "employeeID/projectID"
// @Success 200 {object} view.ContentDataResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /assets/upload [post]
func (h *handler) Upload(c *gin.Context) {
	// 1.1 get userID
	userInfo, err := authutils.GetLoggedInUserInfo(c, h.store, h.repo.DB(), h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	uuidUserID, err := model.UUIDFromString(userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// 1.2 get upload file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, file, ""))
		return
	}

	fType := c.PostForm("type")
	tType := c.PostForm("targetType")
	tID := c.PostForm("targetID")

	l := h.logger.Fields(logger.Fields{
		"handler": "asset",
		"method":  "Upload",
	})

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileType := model.ContentType(fType)
	targetType := model.ContentTargetType(tType)
	fileSize := file.Size

	if !fileType.Valid() {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}
	if !targetType.Valid() {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}

	if fileType == model.ContentTypeImage && !fileExtension.ImageValid() {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}

	if (fileType == model.ContentTypeImage && fileSize > model.MaxFileSizeImage) || (fileType == model.ContentTypeDoc && fileSize > model.MaxFileSizePdf) {
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	if targetType == model.ContentTargetTypeEmployee {
		if tID == "" {
			tID = uuidUserID.String()
		}
		isExisted, err := h.store.Employee.IsExist(tx.DB(), tID)
		if err != nil {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
		if !isExisted {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrEmployeeNotFound), nil, ""))
			return
		}
	}

	if targetType == model.ContentTargetTypeProject {
		isExisted, err := h.store.Project.IsExist(tx.DB(), tID)
		if err != nil {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
		if !isExisted {
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(errs.ErrProjectNotFound), nil, ""))
			return
		}
	}
	gcsPath := ""
	if targetType == model.ContentTargetTypeEmployee || targetType == model.ContentTargetTypeProject {
		gcsPath = fmt.Sprintf("%s/%s/%s/%s", targetType, tID, fileType, fileName)
	} else {
		gcsPath = fmt.Sprintf("%s/%s", targetType, fileName)
	}
	filePath := fmt.Sprintf("https://storage.googleapis.com/%s/%s", h.config.Google.GCSBucketName, gcsPath)

	var targetID model.UUID
	if targetType == model.ContentTargetTypeEmployee || targetType == model.ContentTargetTypeProject {
		targetID, err = model.UUIDFromString(tID)
		if err != nil {
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}
	authType := "Authorization"
	if authutils.IsAPIKey(c) {
		authType = "ApiKey"
	}
	content, err := h.store.Content.Create(tx.DB(), model.Content{
		Type:       fileType.String(),
		Extension:  fileExtension.String(),
		Path:       filePath,
		TargetID:   targetID,
		UploadBy:   uuidUserID,
		TargetType: targetType.String(),
		AuthType:   authType,
	})
	if err != nil {
		l.Error(err, "error create content")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	multipart, err := file.Open()
	if err != nil {
		l.Error(err, "error in open file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// 3. Upload to GCS
	err = h.service.Google.UploadContentGCS(multipart, gcsPath)
	if err != nil {
		l.Error(err, "error in upload file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData(content.Path), nil, done(nil), nil, ""))
}
