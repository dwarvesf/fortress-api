package store

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	employee_store "github.com/dwarvesf/fortress-api/pkg/store/employee"
)

type Store struct {
	Employee employee_store.IStore
}

func New(cfg *config.Config) *Store {
	db := connDb(cfg)
	return &Store{
		Employee: employee_store.New(db),
	}
}
