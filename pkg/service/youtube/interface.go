package youtube

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// IService interface contain related google calendar method
type IService interface {
	CreateBroadcast(*model.Event) (err error)
}
