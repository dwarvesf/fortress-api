package store

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/store/accountstatus"
	"github.com/dwarvesf/fortress-api/pkg/store/chapter"
	"github.com/dwarvesf/fortress-api/pkg/store/country"
	employee_store "github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/permission"
	"github.com/dwarvesf/fortress-api/pkg/store/position"
	"github.com/dwarvesf/fortress-api/pkg/store/role"
	"github.com/dwarvesf/fortress-api/pkg/store/seniority"
)

type Store struct {
	Employee      employee_store.IStore
	Seniority     seniority.IStore
	Chapter       chapter.IStore
	AccountStatus accountstatus.IStore
	Position      position.IStore
	Permission    permission.IStore
	Country       country.IStore
	Role          role.IStore
}

func New(cfg *config.Config) *Store {
	db := connDb(cfg)
	return &Store{
		Employee:      employee_store.New(db),
		Seniority:     seniority.New(db),
		Chapter:       chapter.New(db),
		AccountStatus: accountstatus.New(db),
		Position:      position.New(db),
		Permission:    permission.New(db),
		Country:       country.New(db),
		Role:          role.New(db),
	}
}
