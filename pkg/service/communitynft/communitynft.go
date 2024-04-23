package communitynft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/dwarvesf/fortress-api/pkg/config"
	erc721abi "github.com/dwarvesf/fortress-api/pkg/contracts/erc721"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
)

type IService interface {
	OwnerOf(tokenId int) (string, error)
}

type nft struct {
	instance *erc721abi.ERC721
	evm      evm.IService
	cfg      *config.Config
	l        logger.Logger
}

const (
	CommunityNftContractAddress = "0x888C825C1642aC592a6cCf101542F794480f8803"
)

func New(evm evm.IService, cfg *config.Config, l logger.Logger) (IService, error) {
	instance, err := erc721abi.NewERC721(common.HexToAddress(CommunityNftContractAddress), evm.Client())
	if err != nil {
		return nil, err
	}

	return &nft{
		instance: instance,
		evm:      evm,
		cfg:      cfg,
		l:        l,
	}, nil
}

func (n *nft) OwnerOf(tokenId int) (string, error) {
	owner, err := n.instance.OwnerOf(nil, big.NewInt(int64(tokenId)))
	if err != nil {
		return "", err
	}
	return owner.String(), nil
}
