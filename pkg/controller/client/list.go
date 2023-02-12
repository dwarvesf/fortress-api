package client

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) List(c *gin.Context) ([]*model.Client, error) {
	clients, err := r.store.Client.All(r.repo.DB())
	if err != nil {
		return nil, err
	}

	return clients, nil
}
