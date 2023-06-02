package webhook

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/utils/timeutil"
)

func (h *handler) StoreAccountingTransactionFromBasecamp(msg model.BasecampWebhookMessage) error {
	operationInfo, err := h.getManagementTodoInfo(&msg)
	if err != nil {
		return err
	}

	data := regexp.
		MustCompile(`[S|s]alary\s*(1st|15th)|(.*)\|\s*([0-9\.]+)\s*\|\s*([a-zA-Z]{3})`).
		FindStringSubmatch(msg.Recording.Title)

	if len(data) == 0 {
		return fmt.Errorf(`unknown title format`)
	}

	err = h.storeAccountingTransaction(operationInfo, data, msg.Recording.ID)
	if err != nil {
		return err
	}

	return nil
}

type managementTodoInfo struct {
	month int
	year  int
}

func (h *handler) getManagementTodoInfo(msg *model.BasecampWebhookMessage) (*managementTodoInfo, error) {
	todoList, err := h.service.Basecamp.Todo.GetList(msg.Recording.Parent.URL)
	if err != nil {
		return nil, err
	}
	if todoList == nil || todoList.Parent == nil {
		return nil, nil
	}
	managementInfo := regexp.
		MustCompile(`Accounting \| (.+) ([0-9]{4})`).
		FindStringSubmatch(todoList.Parent.Title)

	accountingID := consts.PlaygroundID
	if h.config.Env == "prod" {
		accountingID = consts.AccountingID
	}
	if len(managementInfo) != 3 && msg.Recording.Bucket.ID == accountingID {
		return nil, nil
	}

	month, err := timeutil.GetMonthFromString(managementInfo[1])
	if err != nil {
		return nil, fmt.Errorf(`format of operation todolist title got wrong %s`, err.Error())
	}
	year, err := strconv.Atoi(managementInfo[2])
	if err != nil {
		return nil, fmt.Errorf(`format of operation todolist title got wrong %d is not a year number`, year)
	}

	return &managementTodoInfo{month, year}, nil
}

func (h *handler) storeAccountingTransaction(date *managementTodoInfo, data []string, id int) error {
	amount, err := strconv.Atoi(strings.ReplaceAll(data[3], ".", ""))
	if err != nil {
		return err
	}

	c, err := h.store.Currency.GetByName(h.repo.DB(), data[4])
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf(`unknown currency`)
		}
		return err
	}
	now := time.Now()

	m := model.AccountingMetadata{
		Source: "basecamp_accounting",
		ID:     fmt.Sprintf("%v", id),
	}
	bonusBytes, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	temp, rate, err := h.service.Wise.Convert(float64(amount), c.Name, currency.VNDCurrency)
	if err != nil {
		return nil
	}
	am := model.NewVietnamDong(int64(temp))

	transaction := &model.AccountingTransaction{
		Name:             data[2],
		Amount:           float64(amount),
		Date:             &now,
		CurrencyID:       &c.ID,
		Currency:         c.Name,
		Category:         checkCategory(strings.ToLower(data[2])),
		Type:             model.AccountingOP,
		ConversionAmount: am.Format(),
		ConversionRate:   rate,
		Metadata:         bonusBytes,
	}

	err = h.StoreOperationAccountingTransaction(transaction)
	if err != nil {
		return err
	}

	return nil
}

func (h *handler) StoreOperationAccountingTransaction(t *model.AccountingTransaction) error {
	if err := h.store.Accounting.CreateTransaction(h.repo.DB(), t); err != nil {
		return err
	}
	return nil
}

func checkCategory(content string) string {
	switch {
	case strings.Contains(content, "office rental") || strings.Contains(content, "cbre"):
		return model.AccountingOfficeSpace
	default:
		return model.AccountingOfficeServices
	}
}
