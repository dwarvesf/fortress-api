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
	DefaultCommunityNftContractAddress = "0x3150825A8b9990567790B22a4F987b6A82d89d54"
)

func New(evm evm.IService, cfg *config.Config, l logger.Logger) (IService, error) {
	addr := cfg.CommunityNft.ContractAddress
	if addr == "" {
		addr = DefaultCommunityNftContractAddress
	}
	instance, err := erc721abi.NewERC721(common.HexToAddress(addr), evm.Client())
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
