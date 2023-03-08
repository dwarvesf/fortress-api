package subscription

import "github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"

type Service interface {
	Subscribe(url string, list *model.SubscriptionList) (err error)
}
