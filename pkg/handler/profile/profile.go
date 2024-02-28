package profile

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
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
// @id getPofile
// @Tags Profile
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} ProfileDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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
// @id updateProfileInfo
// @Tags Profile
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Param Body body UpdateInfoInput true "Body"
// @Success 200 {object} UpdateProfileInfoResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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

	emp, err := h.store.Employee.One(h.repo.DB(), employeeID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			l.Info("emp not found")
			c.JSON(http.StatusNotFound, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}
		l.Error(err, "failed to get emp")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// validate personal email
	_, err = h.store.Employee.OneByEmail(h.repo.DB(), input.PersonalEmail)
	if emp.PersonalEmail != input.PersonalEmail && input.PersonalEmail != "" && !errors.Is(err, gorm.ErrRecordNotFound) {
		if err == nil {
			l.Error(err, "personal email exists")
			c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrEmailExisted, input, ""))
			return
		}
		l.Error(err, "failed to get emp by email")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

	city, err := h.validateAndMappingCity(h.repo.DB(), input.Country, input.City)
	if err != nil {
		l.Info("country or city is invalid")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, input, ""))
		return
	}

	input.Lat = city.Lat
	input.Long = city.Long

	input.ToEmployeeModel(emp)

	tx, done := h.repo.NewTransaction()
	// Update social accounts
	saInput := model.SocialAccountInput{
		GithubID:     input.GithubID,
		NotionID:     input.NotionID,
		NotionName:   input.NotionName,
		NotionEmail:  input.NotionEmail,
		LinkedInName: input.LinkedInName,
	}

	if err := h.updateSocialAccounts(tx.DB(), saInput, emp.ID); err != nil {
		l.Error(err, "failed to update emp")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Get discord info
	discordID := ""
	discordMember, err := h.service.Discord.GetMemberByUsername(input.DiscordName)
	if err != nil {
		l.Error(err, "failed to get discord info")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}
	if discordMember == nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrCouldNotFoundDiscordMemberInGuild), input, ""))
		return
	}

	discordID = discordMember.User.ID

	tmpE, err := h.store.Employee.GetByDiscordID(tx.DB(), discordID, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		l.Errorf(err, "failed to get emp by discordID", "discordID", discordID)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if tmpE.ID != emp.ID {
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrDiscordAccountAlreadyUsedByAnotherEmployee), input, ""))
			return
		}
	}

	discordAccountInput := &model.DiscordAccount{
		DiscordID: discordID,
		Username:  input.DiscordName,
	}

	discordAccount, err := h.store.DiscordAccount.Upsert(tx.DB(), discordAccountInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	emp.DiscordAccountID = discordAccount.ID

	// Update emp
	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *emp,
		"personal_email",
		"phone_number",
		"place_of_residence",
		"address",
		"country",
		"city",
		"lat",
		"long",
		"wise_recipient_id",
		"wise_account_number",
		"wise_recipient_email",
		"wise_recipient_name",
		"wise_currency",
		"discord_account_id",
	)
	if err != nil {
		l.Error(err, "failed to update emp")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](view.ToUpdateProfileInfoData(emp), nil, done(nil), nil, ""))
}

func (h *handler) updateSocialAccounts(db *gorm.DB, input model.SocialAccountInput, employeeID model.UUID) error {
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

func (h *handler) validateAndMappingCity(db *gorm.DB, countryName string, cityName string) (*model.City, error) {
	country, err := h.store.Country.OneByName(db, countryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrCountryNotFound
		}
		return nil, err
	}

	city := country.Cities.GetCity(cityName)
	if city == nil {
		return nil, errs.ErrCityDoesNotBelongToCountry
	}

	return city, nil
}

// UploadAvatar godoc
// @Summary Upload avatar  by id
// @Description Upload avatar  by id
// @id uploadProfileAvatar
// @Tags Profile
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param file formData file true "content upload"
// @Success 200 {object} EmployeeContentDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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
	err = h.service.GoogleStorage.UploadContentGCS(multipart, gcsPath)
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
// @id uploadImage
// @Tags Profile
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param file formData file true "content upload"
// @Success 200 {object} EmployeeContentDataResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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
	err = h.service.GoogleStorage.UploadContentGCS(multipart, gcsPath)
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
// @id getInvitation
// @Tags Onboarding
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} EmployeeInvitationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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
// @id submitOnboardingForm
// @Tags Onboarding
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param Body body SubmitOnboardingFormRequest true "Body"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
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

	if err := input.Validate(); err != nil {
		l.Error(err, "failed to validate input")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, input, ""))
		return
	}

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

	city, err := h.validateAndMappingCity(h.repo.DB(), input.Country, input.City)
	if err != nil {
		l.Info("country or city is invalid")
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, errs.ErrInvalidCountryOrCity, input, ""))
		return
	}

	input.Lat = city.Lat
	input.Long = city.Long

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

	updatedFields := []string{
		"address",
		"city",
		"lat",
		"long",
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
		"discord_account_id",
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
	// Get discord info
	discordMember, err := h.service.Discord.GetMemberByUsername(input.DiscordName)
	if err != nil {
		l.Errorf(err, "failed to get discord info", "discordName", input.DiscordName)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	discordID := ""
	if discordMember == nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrCouldNotFoundDiscordMemberInGuild), input, ""))
		return
	}

	discordID = discordMember.User.ID

	tmpE, err := h.store.Employee.GetByDiscordID(tx.DB(), discordID, false)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		if tmpE.ID != employee.ID {
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(errs.ErrDiscordAccountAlreadyUsedByAnotherEmployee), input, ""))
			return
		}
	}

	discordAccountInput := &model.DiscordAccount{
		DiscordID: discordID,
		Username:  input.DiscordName,
	}

	discordAccount, err := h.store.DiscordAccount.Upsert(tx.DB(), discordAccountInput)
	if err != nil {
		l.Error(err, "failed to get discord info")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
	}

	employeeData.DiscordAccountID = discordAccount.ID

	// Update employee
	_, err = h.store.Employee.UpdateSelectedFieldsByID(tx.DB(), employeeID, *employeeData,
		updatedFields...,
	)
	if err != nil {
		l.Error(err, "failed to update employee")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, done(err), input, ""))
		return
	}

	// Update social accounts
	saInput := model.SocialAccountInput{
		GithubID:     input.GithubID,
		NotionName:   input.NotionName,
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
	employeeInvitation.IsCompleted = true

	if !employeeInvitation.IsTeamEmailCreated {
		err = h.createTeamEmail(employee.TeamEmail, employee.PersonalEmail)
		if err != nil {
			l.Error(err, "failed to create create team email")
		} else {
			employeeInvitation.IsTeamEmailCreated = true
		}
	}

	if !employeeInvitation.IsBasecampAccountCreated {
		err = h.createBasecampAccount(employee)
		if err != nil {
			l.Error(err, "failed to create basecamp account")
		} else {
			employeeInvitation.IsBasecampAccountCreated = true
		}
	}

	if !employeeInvitation.IsDiscordRoleAssigned {
		err = h.assignDiscordRole(discordMember)
		if err != nil {
			l.Error(err, "failed to assign discord role")
		} else {
			employeeInvitation.IsDiscordRoleAssigned = true
		}
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

func (h *handler) assignDiscordRole(discordMember *discordgo.Member) error {
	if discordMember == nil {
		return errs.ErrInvalidDiscordMemberInfo
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
