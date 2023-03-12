package profile

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/profile/errs"
	"github.com/dwarvesf/fortress-api/pkg/handler/profile/request"
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

// GetProfile godoc
// @Summary Get profile information of employee
// @Description Get profile information of employee
// @Tags Profile
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.ProfileDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /profile [get]
func (h *handler) GetProfile(c *gin.Context) {
	userID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "GetProfile",
	})

	rs, err := h.store.Employee.One(h.repo.DB(), userID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToProfileData(rs), nil, nil, nil, ""))
}

// UpdateInfo godoc
// @Summary Update profile info by id
// @Description Update profile info by id
// @Tags Profile
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param id path string true "Employee ID"
// @Param Body body request.UpdateInfoInput true "Body"
// @Success 200 {object} view.UpdateProfileInfoResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /profile [put]
func (h *handler) UpdateInfo(c *gin.Context) {
	employeeID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	input := request.UpdateInfoInput{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "UpdateInfo",
		"request": input,
	})

	employee, err := h.store.Employee.One(h.repo.DB(), employeeID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		l.Error(err, "failed to get employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// validate personal email
	_, err = h.store.Employee.OneByEmail(h.repo.DB(), input.PersonalEmail)
	if employee.PersonalEmail != input.PersonalEmail && input.PersonalEmail != "" && !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			l.Error(err, "personal email exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEmailExisted, input, ""))
			return
		}
		l.Error(err, "failed to get employee by email")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	input.MapEmployeeInput(employee)

	if isValid := h.validateCountryAndCity(h.repo.DB(), input.Country, input.City); !isValid {
		l.Info("country or city is invalid")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, input, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	// Update social accounts
	if err := h.updateSocialAccounts(tx.DB(), input, employee.ID); err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Update employee
	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *employee,
		"personal_email",
		"phone_number",
		"place_of_residence",
		"address",
		"country",
		"city",
		"github_id",
		"notion_id",
		"notion_name",
		"notion_email",
		"discord_name",
		"linkedin_name",
	)
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProfileInfoData(employee), nil, done(nil), nil, ""))
}

func (h *handler) updateSocialAccounts(db *gorm.DB, input request.UpdateInfoInput, employeeID model.UUID) error {
	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "updateSocialAccounts",
		"request": input,
	})

	accounts, err := h.store.SocialAccount.GetByEmployeeID(db, employeeID.String())
	if err != nil {
		l.Error(err, "failed to get social accounts by employeeID")
		return err
	}

	accountsInput := map[model.SocialAccountType]model.SocialAccount{
		model.SocialAccountTypeGitHub: {
			Type:       model.SocialAccountTypeGitHub,
			EmployeeID: employeeID,
			AccountID:  input.GithubID,
			Name:       input.GithubID,
		},
		model.SocialAccountTypeNotion: {
			Type:       model.SocialAccountTypeNotion,
			EmployeeID: employeeID,
			AccountID:  input.NotionID,
			Name:       input.NotionName,
			Email:      input.NotionEmail,
		},
		model.SocialAccountTypeDiscord: {
			Type:       model.SocialAccountTypeDiscord,
			EmployeeID: employeeID,
			Name:       input.DiscordName,
		},
		model.SocialAccountTypeLinkedIn: {
			EmployeeID: employeeID,
			Type:       model.SocialAccountTypeLinkedIn,
			AccountID:  input.LinkedInName,
			Name:       input.LinkedInName,
		},
	}

	for _, account := range accounts {
		delete(accountsInput, account.Type)

		switch account.Type {
		case model.SocialAccountTypeGitHub:
			account.AccountID = input.GithubID
			account.Name = input.GithubID
		case model.SocialAccountTypeNotion:
			account.Name = input.NotionName
			account.Email = input.NotionEmail
		case model.SocialAccountTypeDiscord:
			account.Name = input.DiscordName
		case model.SocialAccountTypeLinkedIn:
			account.AccountID = input.LinkedInName
			account.Name = input.LinkedInName
		default:
			continue
		}

		if _, err := h.store.SocialAccount.UpdateSelectedFieldsByID(db, account.ID.String(), *account, "account_id", "name", "email"); err != nil {
			l.Errorf(err, "failed to update social account %s", account.ID)
			return err
		}
	}

	for _, account := range accountsInput {
		if _, err := h.store.SocialAccount.Create(db, &account); err != nil {
			l.AddField("account", account).Error(err, "failed to create social account")
			return err
		}
	}

	return nil
}

func (h *handler) validateCountryAndCity(db *gorm.DB, countryName string, city string) bool {
	if countryName == "" && city == "" {
		return true
	}

	if countryName == "" && city != "" {
		return false
	}

	l := h.logger.Fields(logger.Fields{
		"handler":     "profile",
		"method":      "validateCountryAndCity",
		"countryName": countryName,
		"city":        city,
	})

	country, err := h.store.Country.OneByName(db, countryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("country not found")
			return false
		}
		l.Error(err, "failed to get country by code")
		return false
	}

	if city != "" && !slices.Contains([]string(country.Cities), city) {
		l.Info("city does not belong to country")
		return false
	}

	return true
}

// UploadAvatar godoc
// @Summary Upload avatar  by id
// @Description Upload avatar  by id
// @Tags Profile
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param file formData file true "content upload"
// @Success 200 {object} view.EmployeeContentDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /profile/upload-avatar [post]
func (h *handler) UploadAvatar(c *gin.Context) {
	employeeID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, file, ""))
		return
	}

	// 1.3 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "UploadAvatar",
		"id":      employeeID,
	})

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	filePath := fmt.Sprintf("https://storage.googleapis.com/%s/employees/%s/images/%s", h.config.Google.GCSBucketName, employeeID, fileName)
	gcsPath := fmt.Sprintf("employees/%s/images/%s", employeeID, fileName)
	fileType := "image"

	// 2.1 validate
	if !fileExtension.ImageValid() {
		l.Info("invalid file extension")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileExtension, nil, ""))
		return
	}

	if fileSize > model.MaxFileSizeImage {
		l.Info("invalid file size")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidFileSize, nil, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	// 2.2 check file name exist
	_, err = h.store.Content.OneByPath(tx.DB(), filePath)
	if err != nil && err != gorm.ErrRecordNotFound {
		l.Error(err, "error query content from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}
	if err == nil {
		l.Info("file already existed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrFileAlreadyExisted), nil, ""))
		return
	}

	// 2.3 check employee existed
	existedEmployee, err := h.store.Employee.One(tx.DB(), employeeID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	_, err = h.store.Content.Create(tx.DB(), model.Content{
		Type:      fileType,
		Extension: fileExtension.String(),
		Path:      filePath,
		TargetID:  existedEmployee.ID,
		UploadBy:  existedEmployee.ID,
	})
	if err != nil {
		l.Error(err, "error query employee from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// 3.1 update avatar link
	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, model.Employee{
		Avatar: filePath,
	}, "avatar")
	if err != nil {
		l.Error(err, "error update avatar from db")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	multipart, err := file.Open()
	if err != nil {
		l.Error(err, "error in open file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	// 3.2 Upload to GCS
	err = h.service.Google.UploadContentGCS(multipart, gcsPath)
	if err != nil {
		l.Error(err, "error in upload file")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToContentData(filePath), nil, done(nil), nil, ""))
}
