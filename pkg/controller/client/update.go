package client

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dwarvesf/fortress-api/pkg/handler/client/request"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (r *controller) Update(c *gin.Context, clientID string, input request.UpdateClientInput) (int, error) {
	tx, done := r.repo.NewTransaction()

	// Get client by id
	client, err := r.store.Client.One(tx.DB(), clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return http.StatusNotFound, done(ErrClientNotFound)
		}

		return http.StatusInternalServerError, done(err)
	}

	// Update client
	client.Name = input.Name
	client.Description = input.Description
	client.RegistrationNumber = input.RegistrationNumber
	client.Address = input.Address
	client.Country = input.Country
	client.Industry = input.Industry
	client.Website = input.Website

	_, err = r.store.Client.UpdateSelectedFieldsByID(tx.DB(), clientID, *client,
		"name",
		"description",
		"registration_number",
		"address",
		"country",
		"industry",
		"website")

	if err != nil {
		return http.StatusInternalServerError, done(err)
	}

	// Delete all client contacts
	if err = r.store.ClientContact.DeleteByClientID(tx.DB(), clientID); err != nil {
		return http.StatusInternalServerError, done(err)
	}

	// Create client contact
	for _, clientContact := range input.Contacts {
		// parse struct email to json
		emails, err := json.Marshal(model.ClientEmail{Emails: clientContact.Emails})
		if err != nil {
			return http.StatusInternalServerError, done(err)
		}

		_, err = r.store.ClientContact.Create(tx.DB(), &model.ClientContact{
			ClientID:      model.MustGetUUIDFromString(clientID),
			Name:          clientContact.Name,
			Role:          clientContact.Role,
			Emails:        datatypes.JSON(emails),
			IsMainContact: clientContact.IsMainContact,
		})

		if err != nil {
			return http.StatusInternalServerError, done(err)
		}
	}

	return http.StatusOK, done(nil)
}
