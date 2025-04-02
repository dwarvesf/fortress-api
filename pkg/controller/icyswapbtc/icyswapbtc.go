package icyswapbtc

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/stringutils"
)

type controller struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	config  *config.Config
}

const (
	TokenDecimal    int = 18
	PlatformDiscord     = "discord"
)

func New(store *store.Store, service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

func (c *controller) Swap(transferRequest *model.TransferRequestResponse) (string, error) {
	icyInfo, err := c.service.IcyBackend.GetIcyInfo()
	if err != nil || icyInfo == nil {
		c.logger.Error(err, "failed to get ICY info")
		return "", fmt.Errorf("can't get ICY info, err: %v", err)
	}

	signature, err := c.GenerateSignature(transferRequest, icyInfo)
	if err != nil || signature == nil {
		c.logger.Error(err, "failed to generate signature for swap")
		return "", fmt.Errorf("can't get signature for swap, err: %w", err)
	}

	swapResponse, err := c.service.IcyBackend.Swap(*signature, transferRequest.Description)
	if err != nil || swapResponse == nil {
		c.logger.Error(err, "failed to swap icy to btc")
		return "", fmt.Errorf("cant' swap icy to btc, err: %w", err)
	}

	// Calculate satoshiAmountReceived by converting string values to float64 and subtracting fee
	btcAmount, err := strconv.ParseFloat(signature.BtcAmount, 64)
	if err != nil {
		c.logger.Error(err, "failed to parse BtcAmount")
		// Continue with the function as the swap was successful
	}

	minSatoshiFee, err := strconv.ParseFloat(icyInfo.MinSatoshiFee, 64)
	if err != nil {
		c.logger.Error(err, "failed to parse MinSatoshiFee")
		// Continue with the function as the swap was successful
	}

	satoshiAmountReceived := btcAmount - minSatoshiFee

	// Format using %0.f which rounds to nearest integer with no decimal places
	satoshiAmountReceivedStr := fmt.Sprintf("%0.f", satoshiAmountReceived)

	if err = c.DmSuccessSwapMessage(transferRequest, satoshiAmountReceivedStr, swapResponse.TxHash); err != nil {
		c.logger.Error(err, "failed to send success swap message")
	}
	//Even if sending success message fails, we still return the txHash without error
	//since the swap itself was successful
	return swapResponse.TxHash, nil
}

func (c *controller) DepositToVault(transferRequest *model.TransferRequestResponse) (string, error) {
	depositToVault, err := c.service.MochiPay.DepositToVault(&mochipay.DepositToVaultRequest{})
	if err != nil || len(depositToVault) == 0 {
		c.logger.Error(err, "failed to revert icy to user")
		return "", fmt.Errorf("can't revert icy to user, err: %v", err)
	}

	var destinationAddress string
	for _, d := range depositToVault {
		if d.Token != nil && d.Token.ID == transferRequest.TokenID {
			destinationAddress = d.Contract.Address
			break
		}
	}

	if destinationAddress == "" {
		c.logger.Error(err, "not found application evm address")
		return "", errors.New("not found application evm address")
	}

	transferResponse, err := c.service.IcyBackend.Transfer(stringutils.FloatToString(transferRequest.Amount, int64(TokenDecimal)), destinationAddress)
	if err != nil || transferResponse == nil || transferResponse.TxHash == "" {
		c.logger.Error(err, "failed to transfer icy ")
		return "", fmt.Errorf("cant' transfer icy, err: %w", err)
	}

	return transferResponse.TxHash, nil
}

func (c *controller) TransferFromVaultToUser(transferRequest *model.TransferRequestResponse) error {
	transaction, err := c.service.MochiPay.TransferFromVaultToUser(transferRequest.ProfileID, &mochipay.TransferFromVaultRequest{
		RecipientIDs: []string{transferRequest.ProfileID},
		Amounts:      []string{transferRequest.Amount},
		TokenID:      transferRequest.TokenID,
		Description:  "Revert icy to user",
	})
	if err != nil || len(transaction) == 0 {
		c.logger.Error(err, "failed to revert icy to user")
		return fmt.Errorf("can't revert icy to user, err: %v", err)
	}

	err = c.DmRevertMessage(transferRequest, transaction[0].TxId)
	if err != nil {
		c.logger.Error(err, "failed to send revert message")
	}

	return nil
}

func (c *controller) WithdrawFromVault(transferRequest *model.TransferRequestResponse) error {
	_, err := c.service.MochiPay.WithdrawFromVault(&mochipay.WithdrawFromVaultRequest{
		Amount:  stringutils.FloatToString(transferRequest.Amount, int64(TokenDecimal)),
		TokenID: transferRequest.TokenID,
	})
	if err != nil {
		c.logger.Error(err, "failed to withdraw from vault")
		return err
	}
	return nil
}

func (c *controller) GenerateSignature(transferRequest *model.TransferRequestResponse, icyInfo *model.IcyInfo) (*model.GenerateSignature, error) {
	if icyInfo == nil {
		err := errors.New("failed to get ICY info")
		c.logger.Error(err, "Icy info empty")
		return nil, err
	}

	icyAmount, err := strconv.ParseFloat(transferRequest.Amount, 64)
	if err != nil {
		c.logger.Error(err, "invalid amount format")
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	icySatoshiRate, err := strconv.ParseFloat(icyInfo.IcySatoshiRate, 64)
	if err != nil {
		c.logger.Error(err, "invalid ICY-Satoshi rate format")
		return nil, fmt.Errorf("invalid ICY-Satoshi rate format: %s - %w", icyInfo.IcySatoshiRate, err)
	}

	satoshiAmount := math.Round(icyAmount * icySatoshiRate)
	icyAmountWithDecimal := stringutils.FloatToString(transferRequest.Amount, int64(TokenDecimal))

	signature, err := c.service.IcyBackend.GetSignature(model.GenerateSignatureRequest{
		IcyAmount:  icyAmountWithDecimal, // No decimal places
		BtcAddress: transferRequest.Description,
		BtcAmount:  fmt.Sprintf("%.0f", satoshiAmount), // No decimal places
	})
	if err != nil {
		c.logger.Error(err, "failed to get signature")
	}
	return signature, err
}

func (c *controller) getDiscordId(profileId string) string {
	profile, err := c.service.MochiProfile.GetProfile(profileId)
	if err != nil || profile == nil {
		c.logger.Error(err, "failed to get profile")
		return ""
	}

	var discordID string
	for _, account := range profile.AssociatedAccounts {
		if account.Platform == PlatformDiscord {
			discordID = account.PlatformIdentifier
			break
		}
	}

	if discordID == "" {
		c.logger.Error(nil, "discord account not found for profile")
	}

	return discordID
}

func (c *controller) DmRevertMessage(transferRequest *model.TransferRequestResponse, TxId int64) error {
	discordId := c.getDiscordId(transferRequest.ProfileID)
	if discordId == "" {
		c.logger.Error(nil, "not found discord id")
		return errors.New("not found discord id")
	}

	lines := []string{
		fmt.Sprintf("`TxID.           `%d", TxId),
		fmt.Sprintf("`Icy Amount.     `**%s ICY**", transferRequest.Amount),
	}

	embedMessage := &discordgo.MessageEmbed{
		Title:       "Successful revert icy",
		Description: strings.Join(lines, "\n"),
		Color:       0x5cd97d,
	}

	err := c.service.Discord.SendDmMessage(discordId, embedMessage)
	if err != nil {
		c.logger.Error(err, "failed to send DM revert message")
		return err
	}

	return nil
}

func (c *controller) DmSuccessSwapMessage(transferRequest *model.TransferRequestResponse, satoshiAmount, txHash string) error {
	discordId := c.getDiscordId(transferRequest.ProfileID)
	if discordId == "" {
		c.logger.Error(nil, "not found discord id")
		return errors.New("not found discord id")
	}

	lines := []string{
		fmt.Sprintf("`TxID.           `[%s](%s/tx/%s)", stringutils.Shorten(txHash), evm.DefaultBaseExplorer, txHash),
		fmt.Sprintf("`Icy Amount.     `**%s ICY**", transferRequest.Amount),
		fmt.Sprintf("`Satoshi Amount. `**%s SAT**", satoshiAmount),
	}

	embedMessage := &discordgo.MessageEmbed{
		Title:       "Successful swap",
		Description: strings.Join(lines, "\n"),
		Color:       0x5cd97d,
	}

	err := c.service.Discord.SendDmMessage(discordId, embedMessage)
	if err != nil {
		c.logger.Error(err, "failed to send DM success swap message")
		return err
	}

	return nil
}
