package discord

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type IController interface {
	Log(in model.LogDiscordInput) error
	PublicAdvanceSalaryLog(in model.LogDiscordInput) error
	PublishIcyActivityLog() error
}

type controller struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	config  *config.Config
	repo    store.DBRepo
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IController {
	return &controller{
		store:   store,
		service: service,
		logger:  logger,
		config:  cfg,
		repo:    repo,
	}
}

func (c *controller) Log(in model.LogDiscordInput) error {
	// Get discord template
	template, err := c.store.DiscordLogTemplate.GetTemplateByType(c.repo.DB(), in.Type)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Get Discord Template failed")
		return err
	}

	data := in.Data.(map[string]interface{})

	// get employee_id in discord format if any
	if employeeID, ok := data["employee_id"]; ok {
		employee, err := c.store.Employee.One(c.repo.DB(), employeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}

		accountID := employee.DisplayName
		if employee.DiscordAccount != nil && employee.DiscordAccount.DiscordID != "" {
			accountID = fmt.Sprintf("<@%s>", employee.DiscordAccount.DiscordID)
		}

		data["employee_id"] = accountID
	}

	if updatedEmployeeID, ok := data["updated_employee_id"]; ok {
		updatedEmployee, err := c.store.Employee.One(c.repo.DB(), updatedEmployeeID.(string), false)
		if err != nil {
			c.logger.Field("err", err.Error()).Warn("Get Employee failed")
			return err
		}

		accountID := updatedEmployee.DisplayName
		if updatedEmployee.DiscordAccount != nil && updatedEmployee.DiscordAccount.DiscordID != "" {
			accountID = fmt.Sprintf("<@%s>", updatedEmployee.DiscordAccount.DiscordID)
		}

		data["updated_employee_id"] = accountID
	}

	// Replace template
	content := template.Content
	for k, v := range data {
		content = strings.ReplaceAll(content, fmt.Sprintf("{{ %s }}", k), fmt.Sprintf("%v", v))
	}

	// log discord
	_, err = c.service.Discord.SendMessage(model.DiscordMessage{
		Content: content,
	}, c.config.Discord.Webhooks.AuditLog)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Log failed")
		return err
	}

	return nil
}

func (c *controller) PublicAdvanceSalaryLog(in model.LogDiscordInput) error {
	data := in.Data.(map[string]interface{})

	icyAmount := data["icy_amount"]
	usdAmount := data["usd_amount"]

	desc := fmt.Sprintf("🧊 %v ICY (%v) has been sent to an anonymous peep as a salary advance.\n", icyAmount, usdAmount)
	desc += "\nFull-time peeps can use `?salary advance` to take a short-term credit benefit."

	embedMessage := model.DiscordMessageEmbed{
		Author:      model.DiscordMessageAuthor{},
		Title:       "💸 New ICY Payment 💸",
		URL:         "",
		Description: desc,
		Color:       3447003,
		Fields:      nil,
		Thumbnail:   model.DiscordMessageImage{},
		Image:       model.DiscordMessageImage{},
		Footer: model.DiscordMessageFooter{
			IconURL: "https://cdn.discordapp.com/avatars/564764617545482251/9c9bd4aaba164fc0b92f13f052405b4d.webp?size=160",
			Text:    "?help to see all commands",
		},
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000+07:00"),
	}

	// log discord
	_, err := c.service.Discord.SendMessage(model.DiscordMessage{
		Embeds: []model.DiscordMessageEmbed{embedMessage},
	}, c.config.Discord.Webhooks.ICYPublicLog)
	if err != nil {
		c.logger.Field("err", err.Error()).Warn("Log failed")
		return err
	}

	return nil
}

func (c *controller) PublishIcyActivityLog() error {
	logger := c.logger.Field("method", "PublishIcyActivityLog")

	txns := make([]mochipay.TransactionData, 0)

	operation := func() error {
		resp, err := c.service.MochiPay.GetListTransactions(mochipay.ListTransactionsRequest{
			Size:         20,
			Status:       mochipay.TransactionStatusSuccess,
			ActionList:   []mochipay.TransactionAction{mochipay.TransactionActionVaultTransfer},
			Type:         mochipay.TransactionTypeReceive,
			TokenAddress: mochipay.ICYAddress,
			ChainIDs:     []string{mochipay.BASEChainID},
			SortBy:       "created_at-", // sort by created_at desc
		})
		if err != nil {
			logger.Error(err, "GetListTransactions failed")
			return err
		}

		// reverse txns to send notification from oldest to newest
		for i := len(resp.Data) - 1; i >= 0; i-- {
			txns = append(txns, resp.Data[i])
		}

		return nil
	}

	// Configure backoff parameters
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 2 * time.Second // Initial sleep interval after an error
	expBackoff.Multiplier = 2                    // Next interval multiplier
	expBackoff.MaxElapsedTime = 30 * time.Second // Maximum cumulative time for retries, retry 4 times = 2 + 4 + 8 + 16 = 30 seconds

	if err := backoff.Retry(operation, expBackoff); err != nil {
		logger.Error(err, "GetListTransactions failed after all retry")
		return err
	}

	now := time.Now()

	for _, transaction := range txns {
		// Just publish transaction in 3 minutes
		if transaction.CreatedAt.Before(now.Add(-3 * time.Minute)) {
			continue
		}

		txRawMetadata, err := json.Marshal(transaction.Metadata)
		if err != nil {
			logger.Error(err, "Marshal metadata failed")
			continue
		}

		var txMetadata mochipay.TransactionMetadata
		if err := json.Unmarshal(txRawMetadata, &txMetadata); err != nil {
			logger.Error(err, "Unmarshal transaction metadata failed")
			continue
		}

		if txMetadata.VaultRequest == nil {
			logger.Info("Skip transaction without vault request")
			continue
		}

		if !strings.EqualFold(txMetadata.VaultRequest.TokenInfo.Address, mochipay.ICYAddress) ||
			txMetadata.VaultRequest.TokenInfo.ChainID != mochipay.BASEChainID {
			logger.Info("Skip transaction without ICY token")
			continue
		}

		receiverProfileID := txMetadata.VaultRequest.Receiver
		receiverProfile, err := c.service.MochiProfile.GetProfile(receiverProfileID)
		if err != nil {
			logger.Error(err, "GetProfile failed")
			continue
		}

		var receiverDiscordID string
		for _, assoc := range receiverProfile.AssociatedAccounts {
			if assoc.Platform == mochiprofile.ProfilePlatformDiscord {
				receiverDiscordID = assoc.PlatformIdentifier
				break
			}
		}

		if receiverDiscordID == "" {
			logger.Info("Skip transaction has profile without discord account")
			continue
		}

		tokenAmount := txMetadata.VaultRequest.Amount

		if txMetadata.VaultRequest.TokenInfo == nil {
			logger.Info("Skip transaction without token info")
			continue
		}

		tokenDecimal := txMetadata.VaultRequest.TokenInfo.Decimal

		tokenAmountDec := utils.ConvertFromString(tokenAmount, int64(tokenDecimal))
		tokenAmountUSD := big.NewFloat(0).Mul(tokenAmountDec, big.NewFloat(1.5))

		transferReason := txMetadata.Message
		if strings.EqualFold(transferReason, mochipay.RewardDefaultMsg) {
			transferReason = fmt.Sprintf("Reward from **%s** vault", txMetadata.VaultRequest.Name)
		}

		desc := `
<:badge5:1058304281775710229> **Receiver:** <@` + receiverDiscordID + `>
<:money:1080757975649624094> **Amount:** <:ICY:1049620715374133288> ` + tokenAmountDec.String() + ` ($` + tokenAmountUSD.String() + `)
<:pepenote:885515949673951282> **Reason:** ` + transferReason + `

Head to [note.d.foundation/earn](https://note.d.foundation/earn) to see list of open quests and r&d topics
		`

		embedMessage := model.DiscordMessageEmbed{
			Author:      model.DiscordMessageAuthor{},
			Title:       "<a:money:1049621199468105758> ICY Reward <a:money:1049621199468105758>",
			Description: desc,
			URL:         "",
			Color:       3447003,
			Fields:      nil,
			Thumbnail:   model.DiscordMessageImage{},
			Image:       model.DiscordMessageImage{},
			Timestamp:   time.Now().Format("2006-01-02T15:04:05.000+07:00"),
		}

		_, err = c.service.Discord.SendMessage(model.DiscordMessage{
			Embeds: []model.DiscordMessageEmbed{embedMessage},
		}, c.config.Discord.Webhooks.ICYPublicLog)
		if err != nil {
			logger.Error(err, "Send ICY log activity failed")
			return err
		}

		// Sleep 2 seconds each time
		time.Sleep(2 * time.Second)
	}

	return nil
}
