package recording

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type Service interface {
	GetFrom(from time.Time, recordingType string) ([]model.Recording, error)
	GetEvents(from time.Time, projectID, recordingID int) ([]model.Event, error)
	Trash(projectID string, recordingID string) (err error)
	Archive(projectID, recordingID int) error
	TryToGetInvoiceImageURL(url string) (res string, err error)
}
