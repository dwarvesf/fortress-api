package todo

import (
	pkgmodel "github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type Service interface {
	CreateList(projectID int, todoSetID int, todoList model.TodoList) (result *model.TodoList, err error)
	CreateGroup(projectID int, todoListID int, group model.TodoGroup) (result *model.TodoGroup, err error)
	Create(projectID int, todoListID int, todo model.Todo) (result *model.Todo, err error)
	Get(url string) (result *model.Todo, err error)
	GetAllInList(todoListID int, projectID int, query ...string) (result []model.Todo, err error)
	GetGroups(todoListID int, projectID int) (result []model.TodoGroup, err error)
	GetLists(projectID int, todoSetID int) (result []model.TodoList, err error)
	GetList(url string) (result *model.TodoList, err error)
	GetProjectsLatestIssue(projectNames []string) (result []*pkgmodel.ProjectIssue, err error)
	CreateHiring(cv *pkgmodel.Candidate) (err error)
	FirstOrCreateList(projectID int, todoSetID int, todoListName string) (result *model.TodoList, err error)
	FirstOrCreateGroup(projectID int, todoListID int, todoGroupName string) (result *model.TodoGroup, err error)
	FirstOrCreateInvoiceTodo(projectID, todoListID int, invoice *pkgmodel.Invoice) (result *model.Todo, err error)
	Update(projectID int, todo model.Todo) (result *model.Todo, err error)
	FirstOrCreateTodo(projectID, todoListID int, todoName string) (result *model.Todo, err error)
	Complete(projectID, todoID int) (err error)
}
