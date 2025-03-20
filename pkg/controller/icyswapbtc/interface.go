package icyswapbtc

import (
	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IController interface {
	GenerateSignature(transferRequest *model.TransferRequestResponse, icyInfo *model.IcyInfo) (*model.GenerateSignature, error)
	Swap(transferRequest *model.TransferRequestResponse) (string, error)
	TransferFromVaultToUser(transferRequest *model.TransferRequestResponse) error
	DepositToVault(transferRequest *model.TransferRequestResponse) (string, error)
	WithdrawFromVault(transferRequest *model.TransferRequestResponse) error
}
