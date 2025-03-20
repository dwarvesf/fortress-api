package icyswapbtc

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IController interface {
	GenerateSignature(transferRequest *model.TransferRequestResponse, icyInfo *model.IcyInfo) (*model.GenerateSignature, error)
	Swap(transferRequest *model.TransferRequestResponse) (string, error)
	RevertIcyToUser(transferRequest *model.TransferRequestResponse) (string, error)
}
