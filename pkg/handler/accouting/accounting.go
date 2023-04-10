package accounting

import (
	"fmt"
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/consts"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"strings"
	"time"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func operationSubcribersID() []int {
	return []int{
		// 22658825, // huynguyen
		13492147, // autobot
		13154176, // duyen
		11452222, // an
		10558375, // tieubao
		11893188, // ly
		11649047, // quang
	}
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

func (h handler) CreateAccountingTodo(month, year int) error {
	logger.NewLogrusLogger().Info(fmt.Sprintf("Creating accouting todo for %s-%v", time.Month(month), year))
	accountingTodo := consts.PlaygroundID
	todoSetID := consts.PlaygroundTodoID
	if h.config.Env == "prod" {
		accountingTodo = consts.AccountingID
		todoSetID = consts.AccountingTodoID
	}

	todoList := bcModel.TodoList{Name: fmt.Sprintf("Accounting | %s %v", time.Month(month).String(), year)}
	todoGroupInFoundation := bcModel.TodoGroup{Name: "In"}
	todoGroupOut := bcModel.TodoGroup{Name: "Out"}

	// Get list accounting(Service table in db) template

	outTodoTemplates, err := h.store.OperationalService.FindOperationByMonth(h.repo.DB(), time.Month(month))
	if err != nil {
		return err
	}

	createTodo, err := h.service.Basecamp.Todo.CreateList(accountingTodo, todoSetID, todoList)
	if err != nil {
		return err
	}

	// Subscribe operation accounts to this BasecampTodoList
	err = h.service.Basecamp.Subscription.Subscribe(createTodo.SubscriptionURL, &bcModel.SubscriptionList{Subscriptions: operationSubcribersID()})
	if err != nil {
		return err
	}

	//Create In group
	inGroup, err := h.service.Basecamp.Todo.CreateGroup(accountingTodo, createTodo.ID, todoGroupInFoundation)
	if err != nil {
		return err
	}

	// Create Out group
	outGroup, err := h.service.Basecamp.Todo.CreateGroup(accountingTodo, createTodo.ID, todoGroupOut)
	if err != nil {
		return err
	}
	// Create todoList for each accounting template into out Group
	err = h.createTodoInOutGroup(outGroup.ID, accountingTodo, outTodoTemplates, month, year)
	if err != nil {
		return err
	}
	// Create Salary to do and add into out group
	err = h.createSalaryTodo(outGroup.ID, accountingTodo, month, year)
	if err != nil {
		return err
	}
	// create to do in IN group
	err = h.createTodoInInGroup(inGroup.ID, accountingTodo)
	if err != nil {
		return err
	}
	return nil
}

func (h handler) createTodoInOutGroup(outGroupID int, projectID int, outTodoTemplates []*model.OperationalService, month int, year int) error {
	for _, v := range outTodoTemplates {
		extraMsg := ""

		// Create CBRE management fee from `Office Rental` template
		if strings.Contains(v.Name, "Office Rental") {
			partment := strings.Replace(v.Name, "Office Rental ", "", 1)
			extraMsg = fmt.Sprintf("Skycenter Office Rental %s %v/%v", partment, month, year)

			s := v.Name
			content := strings.Replace(s, "Office Rental", "CBRE Management fee", 1)
			todo := bcModel.Todo{
				Content:     content,
				DueOn:       fmt.Sprintf("%v-%v-%v", 21, month, year),
				AssigneeIDs: []int{consts.QuangBasecampID},
				Description: getCBREDescription(partment, month, year),
			}
			_, err := h.service.Basecamp.Todo.Create(projectID, outGroupID, todo)
			if err != nil {
				logger.NewLogrusLogger().Error(err, "Fail when try to create CBRE management fee")
				return err
			}

			s = v.Name
			contentElectric := strings.Replace(s, "Office Rental", "Tiền điện", 1)
			todo = bcModel.Todo{
				Content:     fmt.Sprintf("%s %v/%v", contentElectric, month, year),
				DueOn:       fmt.Sprintf("%v-%v-%v", timeutil.LastDayOfMonth(month, year).Day(), month, year),
				AssigneeIDs: []int{consts.QuangBasecampID},
			}
			_, err = h.service.Basecamp.Todo.Create(projectID, outGroupID, todo)
			if err != nil {
				logger.NewLogrusLogger().Error(err, "Fail when try to create CBRE management fee")
				return err
			}
		}

		todo := bcModel.Todo{
			Content:     fmt.Sprintf("%s | %s | %s", v.Name, utils.FormatCurrencyAmount(v.Amount), v.Currency.Name), //nolint:govet
			DueOn:       fmt.Sprintf("%v-%v-%v", timeutil.LastDayOfMonth(month, year).Day(), month, year),
			AssigneeIDs: []int{consts.QuangBasecampID},
			Description: extraMsg,
		}
		_, err := h.service.Basecamp.Todo.Create(projectID, outGroupID, todo)
		if err != nil {
			logger.NewLogrusLogger().Error(err, "Fail when try to create out todos")
			return err
		}
	}
	return nil
}

func getCBREDescription(partment string, month int, year int) string {
	switch string(partment[len(partment)-1]) {
	case "D":
		return fmt.Sprintf("%s(D10.11) thanh toan phi thang %v/%v", partment, month, year)
	case "B":
		return fmt.Sprintf("%s(B1.04) thanh toan phi thang %v/%v", partment, month, year)
	}
	return fmt.Sprintf("%s thanh toan phi thang %v/%v", partment, month, year)
}

func (h handler) createSalaryTodo(outGroupID int, projectID int, month int, year int) error {
	//created TO DO salary 15th
	salary15 := bcModel.Todo{
		Content:     "salary 15th",
		DueOn:       fmt.Sprintf("%v-%v-%v", 12, year, month),
		AssigneeIDs: []int{consts.QuangBasecampID, consts.HanBasecampID},
	}
	_, err := h.service.Basecamp.Todo.Create(projectID, outGroupID, salary15)
	if err != nil {
		return err
	}

	// Create To do Salary 1st
	salary1 := bcModel.Todo{
		Content:     "salary 1st",
		DueOn:       fmt.Sprintf("%v-%v-%v", 27, year, month),
		AssigneeIDs: []int{consts.QuangBasecampID, consts.HanBasecampID},
	}

	_, err = h.service.Basecamp.Todo.Create(projectID, outGroupID, salary1)
	if err != nil {
		return err
	}
	return nil
}

func (h handler) createTodoInInGroup(inGroupID int, projectID int) error {
	activeProjects, err := h.store.Project.GetActiveProjects(h.repo.DB())
	if err != nil {
		return err
	}
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	for _, p := range activeProjects {
		// Default will assign to Giang Than
		assigneeIDs := []int{consts.GiangThanBasecampID}

		_, err := h.service.Basecamp.Todo.Create(projectID, inGroupID, buildInvoiceTodo(p.Name, month, year, assigneeIDs))
		if err != nil {
			logger.NewLogrusLogger().Error(err, fmt.Sprint("Failed to create invoice todo on project", p.Name))
		}
	}
	return nil
}

func buildInvoiceTodo(name string, month, year int, assigneeIDs []int) bcModel.Todo {
	dueOn := getProjectInvoiceDueOn(name, month, year)
	content := getProjectInvoiceContent(name, month, year)
	return bcModel.Todo{
		Content:     content,
		AssigneeIDs: assigneeIDs,
		DueOn:       dueOn,
	}
}
func getProjectInvoiceDueOn(name string, month, year int) string {
	var day int
	if strings.ToLower(name) == "voconic" {
		day = 23
	} else {
		a := timeutil.LastDayOfMonth(month, year)
		day = a.Day()
	}
	return fmt.Sprintf("%v-%v-%v", day, month, year)
}
func getProjectInvoiceContent(name string, month, year int) string {
	return fmt.Sprintf("%s %v/%v", name, month, year)
}
