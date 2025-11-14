package accounting

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/numfmt"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	bcModel "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
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
	ctx := c.Request.Context()

	provider := h.service.AccountingProvider
	if provider == nil {
		err := fmt.Errorf("accounting provider not configured")
		l.Error(err, "missing accounting provider")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	month, year := timeutil.GetMonthAndYearOfNextMonth()

	l.Info(fmt.Sprintf("Creating accounting todo for %s-%v", time.Month(month), year))
	label := fmt.Sprintf("Accounting | %s %v", time.Month(month).String(), year)

	// Get list accounting(Service table in db) template

	outTodoTemplates, err := h.store.OperationalService.FindOperationByMonth(h.repo.DB(), time.Month(month))
	if err != nil {
		l.Errorf(err, "failed to find operation by month", "month", time.Month(month))
		c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil, err, time.Month(month), ""))
		return
	}

	plan, err := provider.CreateMonthlyPlan(ctx, taskprovider.CreateAccountingPlanInput{
		Month: month,
		Year:  year,
		Label: label,
	})
	if err != nil {
		l.Errorf(err, "failed to create accounting plan", "label", label)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Create todoList for each accounting template into out Group
	err = h.createTodoInOutGroup(ctx, provider, plan, outTodoTemplates, month, year)
	if err != nil {
		l.Errorf(err, "failed to create In Out todo group", "month", month, "year", year)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	// Create Salary to do and add into out group
	err = h.createSalaryTodo(ctx, provider, plan, month, year)
	if err != nil {
		l.Errorf(err, "failed to create salary todo", "month", month, "year", year)
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	// create to do IN group
	err = h.createTodoInInGroup(ctx, provider, plan)
	if err != nil {
		l.Errorf(err, "failed to create salary todo")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h handler) createTodoInOutGroup(ctx context.Context, provider taskprovider.AccountingProvider, plan *taskprovider.AccountingPlanRef, outTodoTemplates []*model.OperationalService, month int, year int) error {
	l := h.logger.Fields(logger.Fields{
		"handler": "Accounting",
		"method":  "createTodoInOutGroup",
	})
	for _, v := range outTodoTemplates {
		extraMsg := ""

		// Create CBRE management fee from `Office Rental` template
		if strings.Contains(v.Name, "Office Rental") {
			extraMsg = fmt.Sprintf("Hado Office Rental %v/%v", month, year)

			s := v.Name
			templateID := v.ID
			for _, variant := range []string{"Tiền điện", "CBRE"} {
				content := strings.Replace(s, "Office Rental", variant, 1)
				input := taskprovider.CreateAccountingTodoInput{
					Group:       taskprovider.AccountingGroupOut,
					Title:       fmt.Sprintf("%s %v/%v", content, month, year),
					Description: fmt.Sprintf("I3.18.08 thanh toan %s %v/%v", variant, month, year),
					DueDate:     endOfMonthDate(month, year),
					Assignees:   accountingAssignees(consts.QuangBasecampID),
					Metadata: map[string]any{
						"template_id": templateID,
						"variant":     variant,
					},
				}
				ref, err := provider.CreateAccountingTodo(ctx, plan, input)
				if err != nil {
					l.Error(err, "Fail when try to create CBRE management fee")
					return err
				}
				h.saveAccountingTaskRef(plan, ref, taskprovider.AccountingGroupOut, input.Title, &templateID, nil, map[string]any{
					"variant":     variant,
					"description": input.Description,
				})
			}
		}

		f := &numfmt.Formatter{
			NegativeTemplate: "(n)",
			MinDecimalPlaces: 0,
		}

		input := taskprovider.CreateAccountingTodoInput{
			Group:       taskprovider.AccountingGroupOut,
			Title:       fmt.Sprintf("%s | %s | %s", v.Name, strings.Replace(f.Format(v.Amount), ",", ".", -1), v.Currency.Name), //nolint:govet
			DueDate:     endOfMonthDate(month, year),
			Assignees:   accountingAssignees(consts.QuangBasecampID),
			Description: extraMsg,
			Metadata: map[string]any{
				"template_id": v.ID,
			},
		}
		ref, err := provider.CreateAccountingTodo(ctx, plan, input)
		if err != nil {
			l.Error(err, "Fail when try to create out todos")
			return err
		}
		templateID := v.ID
		h.saveAccountingTaskRef(plan, ref, taskprovider.AccountingGroupOut, input.Title, &templateID, nil, map[string]any{
			"currency":    v.Currency.Name,
			"amount":      v.Amount,
			"description": extraMsg,
		})
	}
	return nil
}

func (h handler) createSalaryTodo(ctx context.Context, provider taskprovider.AccountingProvider, plan *taskprovider.AccountingPlanRef, month int, year int) error {
	salary15 := taskprovider.CreateAccountingTodoInput{
		Group:     taskprovider.AccountingGroupOut,
		Title:     "salary 15th",
		DueDate:   time.Date(year, time.Month(month), 12, 0, 0, 0, 0, time.UTC),
		Assignees: accountingAssignees(consts.QuangBasecampID, consts.HanBasecampID),
	}
	ref15, err := provider.CreateAccountingTodo(ctx, plan, salary15)
	if err != nil {
		return err
	}
	h.saveAccountingTaskRef(plan, ref15, taskprovider.AccountingGroupOut, salary15.Title, nil, nil, map[string]any{
		"type":  "salary",
		"cycle": "15th",
	})

	salary1 := taskprovider.CreateAccountingTodoInput{
		Group:     taskprovider.AccountingGroupOut,
		Title:     "salary 1st",
		DueDate:   time.Date(year, time.Month(month), 27, 0, 0, 0, 0, time.UTC),
		Assignees: accountingAssignees(consts.QuangBasecampID, consts.HanBasecampID),
	}

	ref1, err := provider.CreateAccountingTodo(ctx, plan, salary1)
	if err != nil {
		return err
	}
	h.saveAccountingTaskRef(plan, ref1, taskprovider.AccountingGroupOut, salary1.Title, nil, nil, map[string]any{
		"type":  "salary",
		"cycle": "1st",
	})
	return nil
}

func (h handler) createTodoInInGroup(ctx context.Context, provider taskprovider.AccountingProvider, plan *taskprovider.AccountingPlanRef) error {
	l := h.logger.Fields(logger.Fields{
		"handler": "Accounting",
		"method":  "createTodoInInGroup",
	})
	// Only create monthly invoice todos for Time & Material projects
	// Fixed-Cost projects should not receive automatic monthly invoices
	activeProjects, _, err := h.store.Project.All(h.repo.DB(), project.GetListProjectInput{
		Statuses: []string{model.ProjectStatusActive.String()},
		Types:    []string{model.ProjectTypeTimeMaterial.String()},
	}, model.Pagination{})
	if err != nil {
		return err
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	for _, p := range activeProjects {
		input := taskprovider.CreateAccountingTodoInput{
			Group:     taskprovider.AccountingGroupIn,
			Title:     getProjectInvoiceContent(p.Name, month, year),
			DueDate:   getProjectInvoiceDueDate(p.Name, month, year),
			Assignees: accountingAssignees(consts.GiangThanBasecampID),
			Metadata: map[string]any{
				"project_id": p.ID,
			},
		}
		ref, err := provider.CreateAccountingTodo(ctx, plan, input)
		if err != nil {
			l.Error(err, fmt.Sprint("Failed to create invoice todo on project", p.Name))
			continue
		}
		projectID := p.ID
		h.saveAccountingTaskRef(plan, ref, taskprovider.AccountingGroupIn, input.Title, nil, &projectID, map[string]any{
			"project_name": p.Name,
		})
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
	d := getProjectInvoiceDueDate(name, month, year)
	return fmt.Sprintf("%d-%d-%d", d.Day(), int(d.Month()), d.Year())
}
func getProjectInvoiceContent(name string, month, year int) string {
	return fmt.Sprintf("%s %v/%v", name, month, year)
}

func getProjectInvoiceDueDate(name string, month, year int) time.Time {
	if strings.EqualFold(name, "voconic") {
		return time.Date(year, time.Month(month), 23, 0, 0, 0, 0, time.UTC)
	}
	day := timeutil.LastDayOfMonth(month, year).Day()
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func endOfMonthDate(month, year int) time.Time {
	day := timeutil.LastDayOfMonth(month, year).Day()
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func accountingAssignees(ids ...int) []taskprovider.AccountingAssignee {
	assignees := make([]taskprovider.AccountingAssignee, 0, len(ids))
	for _, id := range ids {
		assignees = append(assignees, taskprovider.AccountingAssignee{
			ExternalID: strconv.Itoa(id),
		})
	}
	return assignees
}

func (h handler) saveAccountingTaskRef(plan *taskprovider.AccountingPlanRef, ref *taskprovider.AccountingTodoRef, group taskprovider.AccountingGroup, title string, templateID *model.UUID, projectID *model.UUID, metadata map[string]any) {
	if plan == nil || ref == nil || ref.ExternalID == "" {
		return
	}
	entry := &model.AccountingTaskRef{
		Month:        plan.Month,
		Year:         plan.Year,
		GroupName:    string(group),
		TaskProvider: string(ref.Provider),
		TaskRef:      ref.ExternalID,
		TaskBoard:    plan.ListID,
		Title:        title,
		Metadata:     marshalMetadata(metadata),
	}
	if templateID != nil && !templateID.IsZero() {
		tid := *templateID
		entry.TemplateID = &tid
	}
	if projectID != nil && !projectID.IsZero() {
		pid := *projectID
		entry.ProjectID = &pid
	}
	if err := h.store.AccountingTaskRef.Create(h.repo.DB(), entry); err != nil {
		h.logger.Error(err, "failed to persist accounting task ref")
	}
}

func marshalMetadata(meta map[string]any) datatypes.JSON {
	if len(meta) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(b)
}
