package client

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"

	"github.com/dwarvesf/fortress-api/pkg/handler/client/request"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) Create(c *gin.Context, input request.CreateClientInput) (*model.Client, error) {

	tx, done := r.repo.NewTransaction()

	// Create client
	client, err := r.store.Client.Create(tx.DB(), &model.Client{
		Name:               input.Name,
		Description:        input.Description,
		RegistrationNumber: input.RegistrationNumber,
		Address:            input.Address,
		Country:            input.Country,
		Industry:           input.Industry,
		Website:            input.Website,
	})

	if err != nil {
		return nil, done(err)
	}

	// Create client contact
	for _, clientContact := range input.Contacts {
		// parse struct email to json
		emails, err := json.Marshal(model.ClientEmail{Emails: clientContact.Emails})
		if err != nil {
			return nil, done(err)
		}

		contact, err := r.store.ClientContact.Create(tx.DB(), &model.ClientContact{
			ClientID:      client.ID,
			Name:          clientContact.Name,
			Role:          clientContact.Role,
			Emails:        datatypes.JSON(emails),
			IsMainContact: clientContact.IsMainContact,
		})

		if err != nil {
			return nil, done(err)
		}

		client.Contacts = append(client.Contacts, *contact)
	}

	return client, done(nil)
}
