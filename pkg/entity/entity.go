package entity

import (
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Entity struct {
}

func New(store *store.Store, service *service.Service) *Entity {
	return &Entity{}
}
