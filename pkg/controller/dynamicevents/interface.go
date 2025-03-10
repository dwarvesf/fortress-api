package dynamicevents

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IController interface {
	CreateEvents(ctx context.Context, data model.DynamicEvent) error
}
