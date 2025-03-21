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
	signature, err := c.GenerateSignature(transferRequest)
	if err != nil || signature == nil {
		return "", errors.New("can't get signature for swap")
	}

	swapResponse, err := c.service.IcyBackend.Swap(*signature, transferRequest.Description)
	if err != nil || swapResponse == nil {
		return "", errors.New("cant' swap icy to btc")
	}

	_ = c.DmSuccessSwapMessage(transferRequest, signature.BtcAmount, swapResponse.TxHash)
	// Even if sending success message fails, we still return the txHash without error
	// since the swap itself was successful
	return swapResponse.TxHash, nil
}

func (c *controller) RevertIcyToUser(transferRequest *model.TransferRequestResponse) error {
	transaction, err := c.service.MochiPay.TransferFromVaultToUser(transferRequest.ProfileID, &mochipay.TransferFromVaultRequest{
		RecipientIDs: []string{transferRequest.ProfileID},
		Amounts:      []string{transferRequest.Amount},
		TokenID:      transferRequest.TokenID,
		Description:  "Revert icy to user",
	})
	if err != nil || len(transaction) == 0 {
		return errors.New("can't revert icy to user")
	}
	err = c.DmRevertMessage(transferRequest, transaction[0].TxId)
	if err != nil {
		return err
	}
	return nil
}

func (c *controller) GenerateSignature(transferRequest *model.TransferRequestResponse) (*model.GenerateSignature, error) {
	icyInfo, err := c.service.IcyBackend.GetIcyInfo()
	if err != nil || icyInfo == nil {
		return nil, fmt.Errorf("failed to get ICY info: %w", err)
	}

	icyAmount, err := strconv.ParseFloat(transferRequest.Amount, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount format: %w", err)
	}

	icySatoshiRate, err := strconv.ParseFloat(icyInfo.IcySatoshiRate, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ICY-Satoshi rate format: %s - %w", icyInfo.IcySatoshiRate, err)
	}

	satoshiAmount := math.Round(icyAmount * icySatoshiRate)
	icyAmountWithDecimal := stringutils.ConvertToFullDecimal(icyAmount, TokenDecimal)

	return c.service.IcyBackend.GetSignature(model.GenerateSignatureRequest{
		IcyAmount:  icyAmountWithDecimal, // No decimal places
		BtcAddress: transferRequest.Description,
		BtcAmount:  fmt.Sprintf("%.0f", satoshiAmount), // No decimal places
	})
}

func (c *controller) getDiscordId(profileId string) string {
	profile, err := c.service.MochiProfile.GetProfile(profileId)
	if err != nil || profile == nil {
		return ""
	}

	var discordID string
	for _, account := range profile.AssociatedAccounts {
		if account.Platform == PlatformDiscord {
			discordID = account.PlatformIdentifier
			break
		}
	}

	return discordID
}

func (c *controller) DmRevertMessage(transferRequest *model.TransferRequestResponse, TxId int64) error {
	discordId := c.getDiscordId(transferRequest.ProfileID)
	if discordId == "" {
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
		return err
	}

	return nil
}

func (c *controller) DmSuccessSwapMessage(transferRequest *model.TransferRequestResponse, satoshiAmount, txHash string) error {
	discordId := c.getDiscordId(transferRequest.ProfileID)
	if discordId == "" {
		return errors.New("not found discord id")
	}

	lines := []string{
		fmt.Sprintf("`TxID.           `[%s](%s/tx/%s)", stringutils.Shorten(txHash), evm.TestnetBASEExplorer, txHash),
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
		return err
	}

	return nil
}
