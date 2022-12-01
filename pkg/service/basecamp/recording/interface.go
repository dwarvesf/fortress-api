package recording

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

type RecordingService interface {
	GetFromTime(recordingType string, t time.Time) ([]model.Recording, error)
	GetEventsFromTime(bucketID int64, recordingID int64, t time.Time) ([]model.Event, error)
}
