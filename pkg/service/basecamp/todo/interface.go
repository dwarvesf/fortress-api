package todo

import (
	mModel "github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type TodoService interface {
	CreateList(todoSetID int64, projectID int64, todoList model.TodoList) (result *model.TodoList, err error)
	CreateGroup(projectID int64, todoListID int64, group model.TodoGroup) (result *model.TodoGroup, err error)
	Create(projectID int64, todoListID int64, todo model.Todo) (result *model.Todo, err error)
	Get(url string) (result *model.Todo, err error)
	GetAllInList(todoListID int64, projectID int64, query ...string) (result []model.Todo, err error)
	GetGroups(todoListID int64, projectID int64) (result []model.TodoGroup, err error)
	GetLists(todoSetID int64, projectID int64) (result []model.TodoList, err error)
	GetList(url string) (result *model.TodoList, err error)
	FirstOrCreateList(projectID int64, todoSetID int64, todoListName string) (result *model.TodoList, err error)
	FirstOrCreateGroup(projectID int64, todoListID int64, todoGroupName string) (result *model.TodoGroup, err error)
	FirstOrCreateInvoiceTodo(todoListID, projectID int64, invoice *mModel.Invoice) (result *model.Todo, err error)
	Update(projectID int64, todo model.Todo) (result *model.Todo, err error)
	FirstOrCreateTodo(projectID, todoListID int64, todoName string) (result *model.Todo, err error)
	Complete(projectID, todoID int64) (err error)
}
