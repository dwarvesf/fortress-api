package store

import (
	"github.com/dwarvesf/fortress-api/pkg/config"
)

type Store struct {
}

func New(cfg *config.Config) *Store {
	// db := connDb(cfg)
	return &Store{}
}
