package client

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *controller) Delete(c *gin.Context, clientID string) (int, error) {
	// Check client existence
	exists, err := r.store.Client.IsExist(r.repo.DB(), clientID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if !exists {
		return http.StatusNotFound, ErrClientNotFound
	}

	tx, done := r.repo.NewTransaction()

	// Delete client
	if err = r.store.Client.Delete(tx.DB(), clientID); err != nil {
		return http.StatusInternalServerError, done(err)
	}

	// Delete client contacts
	if err = r.store.ClientContact.DeleteByClientID(tx.DB(), clientID); err != nil {
		return http.StatusInternalServerError, done(err)
	}

	return http.StatusOK, done(nil)
}
