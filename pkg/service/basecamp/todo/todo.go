package todo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	pkgmodel "github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

type TodoService struct {
	client client.Service
	cfg    *config.Config
}

func NewService(c client.Service, cfg *config.Config) Service {
	return &TodoService{
		client: c,
		cfg:    cfg,
	}
}

func (t *TodoService) CreateList(projectID int, todoSetID int, todoList model.TodoList) (*model.TodoList, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todosets/%v/todolists.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, todoSetID)
	jsonTodo, err := json.Marshal(todoList)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonTodo))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result model.TodoList
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *TodoService) CreateGroup(projectID int, todoListID int, group model.TodoGroup) (*model.TodoGroup, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todolists/%v/groups.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, todoListID)
	jsonGroup, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonGroup))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rs := &model.TodoGroup{}
	if err := json.NewDecoder(res.Body).Decode(&rs); err != nil {
		return nil, err
	}

	return rs, err
}

func (t *TodoService) Create(projectID int, todoListID int, todo model.Todo) (*model.Todo, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todolists/%v/todos.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, todoListID)
	jsonGroup, err := json.Marshal(todo)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonGroup))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := &model.Todo{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *TodoService) Get(url string) (*model.Todo, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rs := &model.Todo{}
	if err := json.NewDecoder(res.Body).Decode(&rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (t *TodoService) GetAllInList(todoListID, projectID int, query ...string) ([]model.Todo, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/buckets/%d/todolists/%d/todos.json`,
		consts.CompanyBasecampID,
		projectID,
		todoListID)

	for _, v := range query {
		url += v
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rs := []model.Todo{}
	if err := json.NewDecoder(res.Body).Decode(&rs); err != nil {
		return nil, err
	}

	link := res.Header.Get("Link")
	page := 2
	for link != "" {
		request, err := http.NewRequest("GET", fmt.Sprintf("%v?page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := t.client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		ss := []model.Todo{}
		if err := json.NewDecoder(response.Body).Decode(&ss); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		rs = append(rs, ss...)

		link = response.Header.Get("Link")
		page++
	}

	return rs, nil
}

func (t *TodoService) GetGroups(todoListID, projectID int) ([]model.TodoGroup, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/buckets/%d/todolists/%d/groups.json`,
		consts.CompanyBasecampID,
		projectID,
		todoListID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rs := []model.TodoGroup{}
	if err := json.NewDecoder(res.Body).Decode(&rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (t *TodoService) GetLists(projectID, todoSetsID int) ([]model.TodoList, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todosets/%v/todolists.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, todoSetsID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	todoLists := []model.TodoList{}
	if err := json.NewDecoder(res.Body).Decode(&todoLists); err != nil {
		return nil, err
	}

	link := res.Header.Get("Link")
	page := 2
	for link != "" {
		request, err := http.NewRequest("GET", fmt.Sprintf("%v?page=%v", url, page), nil)
		if err != nil {
			return nil, err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := t.client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		ss := []model.TodoList{}
		if err := json.NewDecoder(response.Body).Decode(&ss); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		todoLists = append(todoLists, ss...)

		link = response.Header.Get("Link")
		page++
	}

	return todoLists, nil
}

func (t *TodoService) GetList(url string) (*model.TodoList, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	rs := &model.TodoList{}
	if err := json.Unmarshal(b, rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (t *TodoService) GetProjectsLatestIssue(projectNames []string) ([]*pkgmodel.ProjectIssue, error) {
	issues := make([]*pkgmodel.ProjectIssue, len(projectNames))
	todoGroups, err := t.GetGroups(consts.ProjectManagementID, consts.OperationID)
	if err != nil {
		return nil, err
	}

	for i := range todoGroups {
		for j, v := range projectNames {
			if strings.Contains(strings.ToLower(todoGroups[i].Title), strings.ToLower(v)) {
				todos, err := t.GetAllInList(todoGroups[i].ID, consts.OperationID)
				if err != nil {
					return nil, err
				}
				for k := range todos {
					if !todos[k].Completed {
						issues[j] = &pkgmodel.ProjectIssue{ID: todos[k].ID, Name: todos[k].Content, Link: todos[k].AppURL}
					}
				}
			}
		}
	}
	return issues, nil
}

func (t *TodoService) CreateHiring(cv *pkgmodel.Candidate) error {
	hiringID := consts.HiringID
	hiringTodoSetID := consts.HiringTodoSetID
	runMode := t.cfg.Env
	if runMode != "prod" {
		hiringID = consts.PlaygroundID
		hiringTodoSetID = consts.PlaygroundTodoID
	}

	now := time.Now()
	currentQuarter := fmt.Sprintf(`Q%d/%d`, timeutil.GetQuarterFromMonth(now.Month()), now.Year())
	todoList, err := t.FirstOrCreateList(hiringID, hiringTodoSetID, currentQuarter)
	if err != nil {
		return err
	}

	todoGroup, err := t.FirstOrCreateGroup(hiringID, todoList.ID, pkgmodel.GroupRole(cv.Role))
	if err != nil {
		return err
	}

	todo := model.Todo{Content: cv.Name,
		AssigneeIDs: []int{cv.FindHiringInCharge(), consts.HelenBasecampID},
		Description: cv.Note,
		Notify:      (runMode != "local"),
	}

	if cv.IsReferral {
		todo.Content = fmt.Sprintf("Referral: %v", cv.Name)
	}

	res, err := t.Create(hiringID, todoGroup.ID, todo)
	if err != nil {
		return err
	}

	cv.BasecampTodoID = res.ID

	return nil
}

func (t *TodoService) FirstOrCreateList(projectID, todoSetID int, todoListName string) (*model.TodoList, error) {
	todoLists, err := t.GetLists(projectID, todoSetID)
	if err != nil {
		return nil, err
	}
	for i := range todoLists {
		if todoLists[i].Title == todoListName {
			return &todoLists[i], nil
		}
	}
	return t.CreateList(projectID, todoSetID, model.TodoList{Name: todoListName})
}

func (t *TodoService) FirstOrCreateGroup(projectID, todoListID int, todoGroupName string) (*model.TodoGroup, error) {
	todoGroups, err := t.GetGroups(todoListID, projectID)
	if err != nil {
		return nil, err
	}
	for i := range todoGroups {
		if todoGroups[i].Title == todoGroupName {
			return &todoGroups[i], nil
		}
	}
	return t.CreateGroup(projectID, todoListID, model.TodoGroup{Name: todoGroupName})
}

func (t *TodoService) FirstOrCreateTodo(projectID, todoListID int, todoName string) (*model.Todo, error) {
	todos, err := t.GetAllInList(todoListID, projectID)
	if err != nil {
		return nil, err
	}

	for i := range todos {
		// if same project, proceed
		if todos[i].Title == todoName {
			return &todos[i], err
		}
	}
	return t.Create(projectID, todoListID, model.Todo{Content: todoName})
}

func (t *TodoService) FirstOrCreateInvoiceTodo(projectID, todoListID int, invoice *pkgmodel.Invoice) (*model.Todo, error) {
	invoiceTodoName := fmt.Sprintf(`%v %v/%v - #%v`, invoice.Project.Name, invoice.Month, invoice.Year, invoice.Number)
	todos, err := t.GetAllInList(todoListID, projectID)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(utils.RemoveAllSpace(fmt.Sprintf(`%v (%v|0%v)\/%v$`, invoice.Project.Name, invoice.Month, invoice.Month, invoice.Year)))
	if err != nil {
		return nil, err
	}

	for i := range todos {
		if todos[i].Title == invoiceTodoName {
			return &todos[i], nil
		}
		// if same project, proceed
		if re.MatchString(utils.RemoveAllSpace(todos[i].Title)) {
			todos[i].Content = invoiceTodoName
			todos[i].Description = fmt.Sprintf(`<div>%v%v</div>`, todos[i].Description, invoice.TodoAttachment)
			return t.Update(projectID, todos[i])
		}
	}
	return t.Create(projectID, todoListID, model.Todo{Content: invoiceTodoName, Description: fmt.Sprintf(`<div>%v</div>`, invoice.TodoAttachment)})
}

func (t *TodoService) Update(projectID int, todo model.Todo) (*model.Todo, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todos/%v.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, todo.ID)
	jsonGroup, err := json.Marshal(todo)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonGroup))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := &model.Todo{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *TodoService) Complete(projectID, todoID int) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/todos/%v/completion.json", model.BasecampAPIEndpoint, model.CompanyID, projectID, todoID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(`complete req failed with code: %v`, resp.StatusCode)
	}

	return nil
}
