package client

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) List(c *gin.Context) ([]*model.Client, error) {
	clients, err := r.store.Client.All(r.repo.DB(), false, false)
	if err != nil {
		return nil, err
	}

	return clients, nil
}

func (r *controller) PublicList(c *gin.Context) ([]*model.Client, error) {
	clients, err := r.store.Client.All(r.repo.DB(), true, true)
	if err != nil {
		return nil, err
	}

	return clients, nil
}
