package client

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) Detail(c *gin.Context, clientID string) (*model.Client, error) {
	client, err := r.store.Client.One(r.repo.DB(), clientID)
	if err != nil {
		return nil, err
	}
	return client, nil
}
