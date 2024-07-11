package earn

import (
	"context"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IController interface {
	ListEarn(ctx context.Context) ([]model.Earn, error)
}
