package todo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	mModel "github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/client"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type service struct {
	client client.ClientService
}

func New(c client.ClientService) TodoService {
	return &service{
		client: c,
	}
}

func (t *service) CreateList(todoSetID int64, projectID int64, todoList model.TodoList) (*model.TodoList, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todosets/%v/todolists.json", model.BasecampAPIEndpoint, model.CompanyBasecampID, projectID, todoSetID)
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

func (t *service) CreateGroup(projectID int64, todoListID int64, group model.TodoGroup) (*model.TodoGroup, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todolists/%v/groups.json", model.BasecampAPIEndpoint, model.CompanyBasecampID, projectID, todoListID)
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

func (t *service) Create(projectID int64, todoListID int64, todo model.Todo) (*model.Todo, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todolists/%v/todos.json", model.BasecampAPIEndpoint, model.CompanyBasecampID, projectID, todoListID)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *service) Get(url string) (*model.Todo, error) {
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

func (t *service) GetAllInList(todoListID, projectID int64, query ...string) ([]model.Todo, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/buckets/%d/todolists/%d/todos.json`,
		model.CompanyBasecampID,
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

func (t *service) GetGroups(todoListID, projectID int64) ([]model.TodoGroup, error) {
	url := fmt.Sprintf(`https://3.basecampapi.com/%d/buckets/%d/todolists/%d/groups.json`,
		model.CompanyBasecampID,
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

func (t *service) GetLists(todoSetsID, projectID int64) ([]model.TodoList, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todosets/%v/todolists.json", model.BasecampAPIEndpoint, model.CompanyBasecampID, projectID, todoSetsID)
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

func (t *service) GetList(url string) (*model.TodoList, error) {
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

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	rs := &model.TodoList{}
	if err := json.Unmarshal(b, rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (t *service) FirstOrCreateList(projectID, todoSetID int64, todoListName string) (*model.TodoList, error) {
	todoLists, err := t.GetLists(todoSetID, projectID)
	if err != nil {
		return nil, err
	}
	for i := range todoLists {
		if todoLists[i].Title == todoListName {
			return &todoLists[i], nil
		}
	}
	return t.CreateList(todoSetID, projectID, model.TodoList{Name: todoListName})
}

func (t *service) FirstOrCreateGroup(todoListID, projectID int64, todoGroupName string) (*model.TodoGroup, error) {
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

func (t *service) FirstOrCreateTodo(projectID, todoListID int64, todoName string) (*model.Todo, error) {
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

func (t *service) Update(projectID int64, todo model.Todo) (*model.Todo, error) {
	url := fmt.Sprintf("%v/%v/buckets/%v/todos/%v.json", model.BasecampAPIEndpoint, model.CompanyBasecampID, projectID, todo.ID)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, res); err != nil {
		return nil, err
	}

	return res, nil
}

func (t *service) Complete(projectID, todoID int64) error {
	url := fmt.Sprintf("%v/%v/buckets/%v/todos/%v/completion.json", model.BasecampAPIEndpoint, model.CompanyBasecampID, projectID, todoID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(`complete req failed with code: %v`, resp.StatusCode)
	}

	return nil
}

func (t *service) FirstOrCreateInvoiceTodo(todoListID, projectID int64, invoice *mModel.Invoice) (result *model.Todo, err error) {
	invoiceTodoName := fmt.Sprintf(`%v %v/%v - #%v`, invoice.Project.Name, invoice.Month, invoice.Year, invoice.Number)
	todos, err := t.GetAllInList(todoListID, projectID)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(strings.Replace(fmt.Sprintf(`%v (%v|0%v)\/%v$`, invoice.Project.Name, invoice.Month, invoice.Month, invoice.Year), " ", "", -1))
	if err != nil {
		return nil, err
	}

	for i := range todos {
		if todos[i].Title == invoiceTodoName {
			return &todos[i], nil
		}
		// if same project, proceed
		if re.MatchString(strings.Replace(todos[i].Title, " ", "", -1)) {
			todos[i].Content = invoiceTodoName
			todos[i].Description = fmt.Sprintf(`<div>%v%v</div>`, todos[i].Description, invoice.TodoAttachment)
			return t.Update(projectID, todos[i])
		}
	}
	return t.Create(projectID, todoListID, model.Todo{Content: invoiceTodoName, Description: fmt.Sprintf(`<div>%v</div>`, invoice.TodoAttachment)})
}
