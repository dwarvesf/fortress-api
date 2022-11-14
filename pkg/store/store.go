package store

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/store/chapter"
	"github.com/dwarvesf/fortress-api/pkg/store/country"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
	"github.com/dwarvesf/fortress-api/pkg/store/permission"
	"github.com/dwarvesf/fortress-api/pkg/store/position"
	"github.com/dwarvesf/fortress-api/pkg/store/project"
	"github.com/dwarvesf/fortress-api/pkg/store/role"
	"github.com/dwarvesf/fortress-api/pkg/store/seniority"
	"github.com/dwarvesf/fortress-api/pkg/store/techstack"
)

type Store struct {
	Employee   employee.IStore
	Seniority  seniority.IStore
	Chapter    chapter.IStore
	Position   position.IStore
	Permission permission.IStore
	TechStack  techstack.IStore
	Country    country.IStore
	Role       role.IStore
	Project    project.IStore
}

func New(cfg *config.Config) *Store {
	db := connDb(cfg)
	return &Store{
		Employee:   employee.New(db),
		Seniority:  seniority.New(db),
		Chapter:    chapter.New(db),
		Position:   position.New(db),
		Permission: permission.New(db),
		TechStack:  techstack.New(db),
		Country:    country.New(db),
		Role:       role.New(db),
		Project:    project.New(db),
	}
}
