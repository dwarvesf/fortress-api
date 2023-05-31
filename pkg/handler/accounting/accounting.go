package accounting

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
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

func (h handler) CreateAccountingTodo(c *gin.Context) {
	l := h.logger.Fields(logger.Fields{
		"handler": "Accounting",
		"method":  "CreateAccountingTodo",
	})

	month, year := timeutil.GetMonthAndYearOfNextMonth()

	l.Info(fmt.Sprintf("Creating accounting todo for %s-%v", time.Month(month), year))
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
		l.Errorf(err, "failed to find operation by month", "month", time.Month(month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, time.Month(month), ""))
		return
	}

	createTodo, err := h.service.Basecamp.Todo.CreateList(accountingTodo, todoSetID, todoList)
	if err != nil {
		l.Errorf(err, "failed to create todo list", "accountingTodo", accountingTodo, "todoSetID", todoSetID, "todoList", todoList)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, accountingTodo, ""))
		return
	}

	//Create In group
	inGroup, err := h.service.Basecamp.Todo.CreateGroup(accountingTodo, createTodo.ID, todoGroupInFoundation)
	if err != nil {
		l.Errorf(err, "failed to create todo list", "accountingTodo", accountingTodo, "createTodo.ID", createTodo.ID, "todoGroupInFoundation", todoGroupInFoundation)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, accountingTodo, ""))
		return
	}

	// Create Out group
	outGroup, err := h.service.Basecamp.Todo.CreateGroup(accountingTodo, createTodo.ID, todoGroupOut)
	if err != nil {
		l.Errorf(err, "failed to create group", "accountingTodo", accountingTodo, "createTodo.ID", createTodo.ID, "todoGroupOut", todoGroupOut)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, accountingTodo, ""))
		return
	}

	// Create todoList for each accounting template into out Group
	err = h.createTodoInOutGroup(outGroup.ID, accountingTodo, outTodoTemplates, month, year)
	if err != nil {
		l.Errorf(err, "failed to create In Out todo group", "accountingTodo", accountingTodo, "outTodoTemplates", outTodoTemplates, "month", month, "year", year)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, accountingTodo, ""))
		return
	}

	// Create Salary to do and add into out group
	err = h.createSalaryTodo(outGroup.ID, accountingTodo, month, year)
	if err != nil {
		l.Errorf(err, "failed to create salary todo", "accountingTodo", accountingTodo, "month", month, "year", year)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, accountingTodo, ""))
		return
	}
	// create to do IN group
	err = h.createTodoInInGroup(inGroup.ID, accountingTodo)
	if err != nil {
		l.Errorf(err, "failed to create salary todo", "accountingTodo", accountingTodo, "inGroup.ID", inGroup.ID)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, accountingTodo, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h handler) createTodoInOutGroup(outGroupID int, projectID int, outTodoTemplates []*model.OperationalService, month int, year int) error {
	l := h.logger.Fields(logger.Fields{
		"handler": "Accounting",
		"method":  "createTodoInOutGroup",
	})
	for _, v := range outTodoTemplates {
		extraMsg := ""

		// Create CBRE management fee from `Office Rental` template
		if strings.Contains(v.Name, "Office Rental") {
			partment := strings.Replace(v.Name, "Office Rental ", "", 1)
			extraMsg = fmt.Sprintf("Hado Office Rental %s %v/%v", partment, month, year)

			s := v.Name
			contentElectric := strings.Replace(s, "Office Rental", "Tiền điện", 1)
			todo := bcModel.Todo{
				Content:     fmt.Sprintf("%s %v/%v", contentElectric, month, year),
				DueOn:       fmt.Sprintf("%v-%v-%v", timeutil.LastDayOfMonth(month, year).Day(), month, year),
				AssigneeIDs: []int{consts.QuangBasecampID},
			}
			_, err := h.service.Basecamp.Todo.Create(projectID, outGroupID, todo)
			if err != nil {
				l.Error(err, "Fail when try to create CBRE management fee")
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
			l.Error(err, "Fail when try to create out todos")
			return err
		}
	}
	return nil
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
	l := h.logger.Fields(logger.Fields{
		"handler": "Accounting",
		"method":  "createSalaryTodo",
	})
	activeProjects, _, err := h.store.Project.All(h.repo.DB(), project.GetListProjectInput{Statuses: []string{model.ProjectStatusActive.String()}}, model.Pagination{})
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
			l.Error(err, fmt.Sprint("Failed to create invoice todo on project", p.Name))
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
