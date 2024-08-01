package youtube

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
	"google.golang.org/api/youtube/v3"
)

// IService interface contain related google calendar method
type IService interface {
	GetLatestBroadcast() (*youtube.LiveBroadcast, error)
	CreateBroadcast(*model.Event) (err error)
}
