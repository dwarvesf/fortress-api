package accounting

type IHandler interface {
	CreateAccountingTodo(month, year int) error
}
