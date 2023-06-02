package profile

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
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
	controller *controller.Controller
	store      *store.Store
	service    *service.Service
	logger     logger.Logger
	repo       store.DBRepo
	config     *config.Config
}

// New returns a handler
func New(controller *controller.Controller, store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		controller: controller,
		store:      store,
		repo:       repo,
		service:    service,
		logger:     logger,
		config:     cfg,
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

	rs, err := h.store.Employee.One(h.repo.DB(), userID, true)
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

	input.ToEmployeeModel(employee)

	if isValid := h.validateCountryAndCity(h.repo.DB(), input.Country, input.City); !isValid {
		l.Info("country or city is invalid")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, input, ""))
		return
	}

	tx, done := h.repo.NewTransaction()

	// Update social accounts
	saInput := socialAccountInput{
		GithubID:     input.GithubID,
		NotionID:     input.NotionID,
		NotionName:   input.NotionName,
		NotionEmail:  input.NotionEmail,
		DiscordName:  input.DiscordName,
		LinkedInName: input.LinkedInName,
	}
	if err := h.updateSocialAccounts(tx.DB(), saInput, employee.ID); err != nil {
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
		"wise_recipient_id",
		"wise_account_number",
		"wise_recipient_email",
		"wise_recipient_name",
		"wise_currency",
	)
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProfileInfoData(employee), nil, done(nil), nil, ""))
}

type socialAccountInput struct {
	GithubID     string
	NotionID     string
	NotionName   string
	NotionEmail  string
	DiscordID    string
	DiscordName  string
	LinkedInName string
}

func (h *handler) updateSocialAccounts(db *gorm.DB, input socialAccountInput, employeeID model.UUID) error {
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
			AccountID:  input.DiscordID,
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

// Upload godoc
// @Summary Upload image  by id
// @Description Upload image  by id
// @Tags Profile
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param file formData file true "content upload"
// @Success 200 {object} view.EmployeeContentDataResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /profile/upload [post]
func (h *handler) Upload(c *gin.Context) {
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

	fDoctype := c.PostForm("documentType")

	// 1.3 prepare the logger
	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "UploadAvatar",
		"id":      employeeID,
	})

	docType := model.DocumentType(fDoctype)

	fileName := file.Filename
	fileExtension := model.ContentExtension(filepath.Ext(fileName))
	fileSize := file.Size
	filePath := fmt.Sprintf("https://storage.googleapis.com/%s/employees/%s/images/%s", h.config.Google.GCSBucketName, employeeID, fileName)
	gcsPath := fmt.Sprintf("employees/%s/images/%s", employeeID, fileName)
	fileType := "image"

	// 2.1 validate
	if !docType.Valid() {
		l.Info("invalid document type")
		c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, errs.ErrInvalidDocumentType, nil, ""))
		return
	}
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

	switch docType {
	case model.DocumentTypeAvatar:
		_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, model.Employee{
			Avatar: filePath,
		}, "avatar")
		if err != nil {
			l.Error(err, "error update avatar from db")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	case model.DocumentTypeIDPhotoFront:
		_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, model.Employee{
			IdentityCardPhotoFront: filePath,
		}, "avatar")
		if err != nil {
			l.Error(err, "error update id card front from db")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	case model.DocumentTypeIDPhotoBack:
		_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, model.Employee{
			IdentityCardPhotoBack: filePath,
		}, "avatar")
		if err != nil {
			l.Error(err, "error update id card back from db")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), nil, ""))
			return
		}
	}
	// 3.1 update avatar link

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

// GetInvitation godoc
// @Summary Get invitation state based on token
// @Description Submit Get invitation state based on token
// @Tags Onboarding
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Success 200 {object} view.EmployeeInvitationResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /invite [get]
func (h *handler) GetInvitation(c *gin.Context) {
	employeeID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "GetInvitation",
	})

	employeeInvitation, err := h.store.EmployeeInvitation.OneByEmployeeID(h.repo.DB(), employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee invitation not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		l.Error(err, "failed to get employee invitation")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToBasicEmployeeInvitationData(employeeInvitation), nil, nil, nil, ""))
}

// SubmitOnboardingForm godoc
// @Summary Submit onboarding form
// @Description Submit Onboarding form
// @Tags Onboarding
// @Accept  json
// @Produce  json
// @Param Authorization header string true "jwt token"
// @Param Body body request.SubmitOnboardingFormRequest true "Body"
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Failure 404 {object} view.ErrorResponse
// @Failure 500 {object} view.ErrorResponse
// @Router /invite/submit [put]
func (h *handler) SubmitOnboardingForm(c *gin.Context) {
	employeeID, err := authutils.GetUserIDFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	input := request.SubmitOnboardingFormRequest{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	l := h.logger.Fields(logger.Fields{
		"handler": "profile",
		"method":  "SubmitOnboardingForm",
		"request": input,
	})

	employeeInvitation, err := h.store.EmployeeInvitation.OneByEmployeeID(h.repo.DB(), employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("employee invitation not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		l.Error(err, "failed to get employee invitation")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if employeeInvitation.IsCompleted {
		l.Errorf(errs.ErrOnboardingFormAlreadyDone, "employee invitation is expired")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrOnboardingFormAlreadyDone, nil, ""))
		return
	}

	if err := input.Validate(); err != nil {
		l.Error(err, "failed to validate input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

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

	if isValid := h.validateCountryAndCity(h.repo.DB(), input.Country, input.City); !isValid {
		l.Info("country or city is invalid")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, input, ""))
		return
	}

	updatedFields := []string{
		"address",
		"city",
		"country",
		"gender",
		"horoscope",
		"date_of_birth",
		"local_bank_branch",
		"local_bank_currency",
		"local_bank_number",
		"local_bank_recipient_name",
		"local_branch_name",
		"mbti",
		"phone_number",
		"place_of_residence",
		"working_status",
	}

	if input.Avatar != "" {
		updatedFields = append(updatedFields, "avatar")
	}

	if input.IdentityCardPhotoFront != "" {
		updatedFields = append(updatedFields, "identity_card_photo_front")
	}

	if input.IdentityCardPhotoBack != "" {
		updatedFields = append(updatedFields, "identity_card_photo_back")
	}

	if input.PassportPhotoFront != "" {
		updatedFields = append(updatedFields, "passport_photo_front")
	}

	if input.PassportPhotoBack != "" {
		updatedFields = append(updatedFields, "passport_photo_back")
	}

	employeeData := input.ToEmployeeModel()

	tx, done := h.repo.NewTransaction()
	// Update employee
	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employeeID, *employeeData,
		updatedFields...,
	)
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Get discord info
	discordMember, err := h.getDiscordInfo(input.DiscordName)
	if err != nil {
		l.Error(err, "failed to get discord info")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Update social accounts
	saInput := socialAccountInput{
		GithubID:     input.GithubID,
		NotionName:   input.NotionName,
		DiscordName:  input.DiscordName,
		DiscordID:    discordMember.User.ID,
		LinkedInName: input.LinkedInName,
	}

	if err := h.updateSocialAccounts(tx.DB(), saInput, employee.ID); err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Commit transaction update employee info
	_ = done(nil)
	employeeInvitation.IsInfoUpdated = true

	if !employeeInvitation.IsTeamEmailCreated {
		err = h.createTeamEmail(employee.TeamEmail, employee.PersonalEmail)
		if err != nil {
			l.Error(err, "failed to create create team email")
		}
		employeeInvitation.IsTeamEmailCreated = true
	}

	if !employeeInvitation.IsBasecampAccountCreated {
		err = h.createBasecampAccount(employee)
		if err != nil {
			l.Error(err, "failed to create basecamp account")
		}
		employeeInvitation.IsBasecampAccountCreated = true
	}

	if !employeeInvitation.IsDiscordRoleAssigned {
		err = h.assignDiscordRole(discordMember)
		if err != nil {
			l.Error(err, "failed to assign discord role")
		}
		employeeInvitation.IsDiscordRoleAssigned = true
	}

	if employeeInvitation.IsInfoUpdated && employeeInvitation.IsTeamEmailCreated && employeeInvitation.IsBasecampAccountCreated && employeeInvitation.IsDiscordRoleAssigned {
		employeeInvitation.IsCompleted = true
	}

	err = h.store.EmployeeInvitation.Save(h.repo.DB(), employeeInvitation)
	if err != nil {
		l.Error(err, "failed to update employee invitation")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	err = h.controller.Discord.Log(model.LogDiscordInput{
		Type: "employee_submit_onboarding_form",
		Data: map[string]interface{}{
			"employee_id": employee.ID.String(),
		},
	})
	if err != nil {
		l.Error(err, "failed to logs to discord")
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) createTeamEmail(teamEmail string, personalEmail string) error {
	if h.config.Env != "prod" {
		return nil
	}

	return h.service.ImprovMX.CreateAccount(teamEmail, personalEmail)
}

func (h *handler) createBasecampAccount(employee *model.Employee) error {
	if h.config.Env != "prod" {
		employee.BasecampID = 123456
		employee.BasecampAttachableSGID = "sample_sg_id"
		_, err := h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employee.ID.String(), *employee,
			"basecamp_id",
			"basecamp_attachable_sgid",
		)
		if err != nil {
			return err
		}

		return nil
	}

	email := employee.PersonalEmail
	if employee.TeamEmail != "" {
		email = employee.TeamEmail
	}
	bcID, sgID, err := h.service.Basecamp.People.Create(employee.DisplayName, email, model.OrganizationNameDwarves)
	if err != nil {
		return err
	}

	employee.BasecampID = int(bcID)
	employee.BasecampAttachableSGID = sgID

	_, err = h.store.Employee.UpdateSelectedFieldsByID(h.repo.DB(), employee.ID.String(), *employee,
		"basecamp_id",
		"basecamp_attachable_sgid",
	)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) getDiscordInfo(discordName string) (*discordgo.Member, error) {
	if len(discordName) == 0 {
		return nil, nil
	}

	discordNameParts := strings.Split(discordName, "#")

	guildMembers, err := h.service.Discord.SearchMember(discordNameParts[0])
	if err != nil {
		return nil, err
	}

	var discordMember *discordgo.Member
	for _, m := range guildMembers {
		if len(discordNameParts) == 1 {
			if m.User.Username == discordNameParts[0] {
				discordMember = m
			}
			break
		}
		if len(discordNameParts) > 1 {
			if m.User.Username == discordNameParts[0] && m.User.Discriminator == discordNameParts[1] {
				discordMember = m
			}
			break
		}
	}

	return discordMember, nil
}

func (h *handler) assignDiscordRole(discordMember *discordgo.Member) error {
	if discordMember != nil {
		return nil
	}
	// Get list discord role
	dRoles, err := h.service.Discord.GetRoles()
	if err != nil {
		return err
	}

	peepsRoleID := ""
	for _, r := range dRoles {
		if r.Name == model.DiscordRolePeeps.String() {
			peepsRoleID = r.ID
			break
		}
	}

	if peepsRoleID != "" {
		// Check if user already has peeps role
		if utils.Contains(discordMember.Roles, peepsRoleID) {
			return nil
		}

		err := h.service.Discord.AddRole(discordMember.User.ID, peepsRoleID)
		if err != nil {
			return err
		}
	}

	return nil
}
