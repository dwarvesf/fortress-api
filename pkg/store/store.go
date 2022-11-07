package store

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/store/accountstatus"
	"github.com/dwarvesf/fortress-api/pkg/store/country"
	employee_store "github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/permission"
	position_store "github.com/dwarvesf/fortress-api/pkg/store/position"
	"github.com/dwarvesf/fortress-api/pkg/store/role"
)

type Store struct {
	Employee      employee_store.IStore
	AccountStatus accountstatus.IStore
	Position      position_store.IStore
	Permission    permission.IStore
	Country       country.IStore
	Role          role.IStore
}

func New(cfg *config.Config) *Store {
	db := connDb(cfg)
	return &Store{
		Employee:      employee_store.New(db),
		AccountStatus: accountstatus.New(db),
		Position:      position_store.New(db),
		Permission:    permission.New(db),
		Country:       country.New(db),
		Role:          role.New(db),
	}
}
