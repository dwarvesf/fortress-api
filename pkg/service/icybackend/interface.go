package icybackend

import "github.com/dwarvesf/fortress-api/pkg/model"

type IService interface {
	GetIcyInfo() (*model.IcyInfo, error)
	GetSignature(request model.GenerateSignatureRequest) (*model.GenerateSignature, error)
	Swap(signature model.GenerateSignature, btcAddress string) (*model.SwapResponse, error)
	Transfer(icyAmount, destinationAddress string) (*model.TransferResponse, error)
}
