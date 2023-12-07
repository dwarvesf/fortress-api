package employee

import (
	"mime/multipart"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type controller struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

type IController interface {
	List(workingStatuses []string, body GetListEmployeeInput, userInfo *model.CurrentLoggedUserInfo) (employees []*model.Employee, tptal int64, err error)
	Details(id string, userInfo *model.CurrentLoggedUserInfo) (employee *model.Employee, err error)
	UpdateEmployeeStatus(employeeID string, body UpdateWorkingStatusInput) (employee *model.Employee, err error)
	UpdateGeneralInfo(employeeID string, body UpdateEmployeeGeneralInfoInput) (employee *model.Employee, err error)
	Create(userID string, body CreateEmployeeInput) (employee *model.Employee, err error)
	UpdateSkills(l logger.Logger, employeeID string, body UpdateSkillsInput) (employee *model.Employee, err error)
	UpdatePersonalInfo(employeeID string, body UpdatePersonalInfoInput) (employee *model.Employee, err error)
	UploadAvatar(uuidUserID model.UUID, file *multipart.FileHeader, params UploadAvatarInput) (filePath string, err error)
	UpdateRole(userID string, input UpdateRoleInput) (err error)
	GetLineManagers(userInfo *model.CurrentLoggedUserInfo) (employees []*model.Employee, err error)
	UpdateBaseSalary(l logger.Logger, employeeID string, body UpdateBaseSalaryInput) (employee *model.BaseSalary, err error)
	ListWithLocation() (employees []*model.Employee, err error)
	ListByDiscordRequest(discordID, email, key string, userInfo *model.CurrentLoggedUserInfo) ([]model.Employee, error)
	SalaryAdvance(discordID string, amount int64) (*SalaryAdvanceResponse, error)
	CheckSalaryAdvance(discordID string) (string, string, error)
	ListWithMMAScore() (employees []model.EmployeeMMAScoreData, err error)
	ListAggregatedSalaryAdvance(input ListAggregatedSalaryAdvanceInput) (*model.SalaryAdvanceReport, error)
	GetEmployeeEarnTransactions(discordID string, input GetEmployeeEarnTransactionsInput) (model.EmployeeEarnTransactions, int64, error)
	GetEmployeeTotalEarn(discordID string) (string, float64, error)
}
