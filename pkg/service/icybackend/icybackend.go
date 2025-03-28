package icybackend

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/contracts/erc20"
	"github.com/dwarvesf/fortress-api/pkg/contracts/icyswapbtc"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
)

const (
	ICYBTCSwapAddress = "0xdA3E22edf0357c781154D8DEDcfC32D7B6B0B12D"
	IcyTokenAddress   = "0xf289e3b222dd42b185b7e335fa3c5bd6d132441d"
)

type icybackend struct {
	icyswapbtc *icyswapbtc.Icyswapbtc
	evm        evm.IService
	cfg        *config.Config
	logger     logger.Logger
}

func New(evm evm.IService, cfg *config.Config, l logger.Logger) (IService, error) {
	instance, err := icyswapbtc.NewIcyswapbtc(common.HexToAddress(ICYBTCSwapAddress), evm.Client())
	if err != nil {
		return nil, err
	}
	return &icybackend{
		icyswapbtc: instance,
		evm:        evm,
		cfg:        cfg,
		logger:     l,
	}, nil
}

func (i *icybackend) Swap(signature model.GenerateSignature, btcAddress string) (*model.SwapResponse, error) {
	// 1. Get required addresses
	icyAddress := common.HexToAddress(IcyTokenAddress)
	icySwapBtcAddress := common.HexToAddress(ICYBTCSwapAddress)

	// 2. Create private key from config
	privateKey, err := crypto.HexToECDSA(i.cfg.MochiPay.IcyPoolPrivateKey)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] crypto.HexToECDSA failed")
		return nil, err
	}

	// 3. Get chain ID
	chainID, err := i.evm.Client().ChainID(context.Background())
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] client.ChainID failed")
		return nil, err
	}

	// 4. Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] NewKeyedTransactorWithChainID failed")
		return nil, err
	}

	// 5. Get current nonce for the account
	nonce, err := i.evm.Client().PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] PendingNonceAt failed")
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// 6. Create ERC20 instance for token approval
	erc20Instance, err := erc20.NewErc20(icyAddress, i.evm.Client())
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] NewErc20 failed")
		return nil, err
	}

	// 7. Convert string values to appropriate types
	icyAmount, ok := new(big.Int).SetString(signature.IcyAmount, 10)
	if !ok {
		err := fmt.Errorf("invalid icy_amount format")
		i.logger.Errorf(err, "[icybackend.Swap] invalid icy_amount format")
		return nil, err
	}

	btcAmount, ok := new(big.Int).SetString(signature.BtcAmount, 10)
	if !ok {
		err := fmt.Errorf("invalid btc_amount format")
		i.logger.Errorf(err, "[icybackend.Swap] invalid btc_amount format")
		return nil, err
	}

	sigNonce, ok := new(big.Int).SetString(signature.Nonce, 10)
	if !ok {
		err := fmt.Errorf("invalid nonce format")
		i.logger.Errorf(err, "[icybackend.Swap] invalid nonce format")
		return nil, err
	}

	deadline, ok := new(big.Int).SetString(signature.Deadline, 10)
	if !ok {
		err := fmt.Errorf("invalid deadline format")
		i.logger.Errorf(err, "[icybackend.Swap] invalid deadline format")
		return nil, err
	}

	// 8. Convert hex signature string to bytes
	signatureBytes, err := hex.DecodeString(strings.TrimPrefix(signature.Signature, "0x"))
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] hex.DecodeString failed for signature")
		return nil, err
	}

	// 9. Approve ICY token transfer to the swap contract
	txApprove, err := erc20Instance.Approve(auth, icySwapBtcAddress, icyAmount)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] erc20.Approve failed")
		return nil, err
	}

	// 10. Wait for the approval transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), i.evm.Client(), txApprove)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] WaitMined for approval failed")
		return nil, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		err := fmt.Errorf("approval transaction failed")
		i.logger.Errorf(err, "[icybackend.Swap] approval transaction failed")
		return nil, err
	}

	// 11. Update nonce for the swap transaction
	nonce++
	auth.Nonce = big.NewInt(int64(nonce))

	// 12. Execute the swap using the generated binding
	txSwap, err := i.icyswapbtc.Swap(
		auth,
		icyAmount,
		btcAddress,
		btcAmount,
		sigNonce,
		deadline,
		signatureBytes,
	)

	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] icyswapbtc.Swap failed")
		return nil, err
	}

	// 10. Wait for the approval transaction to be mined
	receiptSwap, err := bind.WaitMined(context.Background(), i.evm.Client(), txSwap)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Swap] WaitMined for swap failed")
		return nil, err
	}

	if receiptSwap.Status != types.ReceiptStatusSuccessful {
		err := fmt.Errorf("swap transaction failed")
		i.logger.Errorf(err, "[icybackend.Swap] swap transaction failed")
		return nil, err
	}

	// 13. Return transaction hash
	return &model.SwapResponse{
		TxHash: txSwap.Hash().String(),
	}, nil
}

func (i *icybackend) Transfer(icyAmount, destinationAddress string) (*model.TransferResponse, error) {
	// 1. Get required address
	icyAddress := common.HexToAddress(IcyTokenAddress)
	destinationAddr := common.HexToAddress(destinationAddress)

	// 2. Create private key from config
	privateKey, err := crypto.HexToECDSA(i.cfg.MochiPay.IcyPoolPrivateKey)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] crypto.HexToECDSA failed")
		return nil, err
	}

	// 3. Get chain ID
	chainID, err := i.evm.Client().ChainID(context.Background())
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] client.ChainID failed")
		return nil, err
	}

	// 4. Create transaction options
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] NewKeyedTransactorWithChainID failed")
		return nil, err
	}

	// 5. Get current nonce for the account
	nonce, err := i.evm.Client().PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] PendingNonceAt failed")
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))

	// 6. Create ERC20 instance for token transfer
	erc20Instance, err := erc20.NewErc20(icyAddress, i.evm.Client())
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] NewErc20 failed")
		return nil, err
	}

	// 7. Convert string amount to big.Int
	amount, ok := new(big.Int).SetString(icyAmount, 10)
	if !ok {
		err := fmt.Errorf("invalid icy_amount format")
		i.logger.Errorf(err, "[icybackend.Transfer] invalid icy_amount format")
		return nil, err
	}

	// 8. Execute the transfer transaction
	tx, err := erc20Instance.Transfer(auth, destinationAddr, amount)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] erc20.Transfer failed")
		return nil, err
	}

	// 9. Wait for the transfer transaction to be mined
	receipt, err := bind.WaitMined(context.Background(), i.evm.Client(), tx)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.Transfer] WaitMined for transfer failed")
		return nil, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		err := fmt.Errorf("transfer transaction failed")
		i.logger.Errorf(err, "[icybackend.Transfer] transfer transaction failed")
		return nil, err
	}

	// 10. Return transaction hash
	return &model.TransferResponse{
		TxHash: tx.Hash().String(),
	}, nil
}

func (i *icybackend) GetIcyInfo() (*model.IcyInfo, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/swap/info", i.cfg.IcyBackend.BaseURL)

	r, err := client.Get(url)
	if err != nil {
		i.logger.Errorf(err, "[icybackend.GetIcyInfo] client.Get failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &model.IcyInfoResponse{}
	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		i.logger.Errorf(err, "[icybackend.GetIcyInfo] decoder.Decode failed")
		return nil, err
	}

	return &res.Data, nil
}

func (i *icybackend) GetSignature(request model.GenerateSignatureRequest) (*model.GenerateSignature, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	url := fmt.Sprintf("%s/api/v1/swap/generate-signature", i.cfg.IcyBackend.BaseURL)

	requestBody, err := json.Marshal(request)

	if err != nil {
		i.logger.Errorf(err, "[icybackend.GetSignature] json.Marshal failed")
		return nil, err
	}

	r, err := client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		i.logger.Errorf(err, "[icybackend.GetSignature] client.Post failed")
		return nil, err
	}
	defer r.Body.Close()

	res := &model.GenerateSignatureResponse{}
	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		i.logger.Errorf(err, "[icybackend.GetSignature] decoder.Decode failed")
		return nil, err
	}

	if res.Error != nil {
		return nil, errors.New(*res.Error)
	}

	return &res.Data, nil
}
