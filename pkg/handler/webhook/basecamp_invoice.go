package webhook

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/store/invoice"
)

func (h *handler) GetInvoiceViaBasecampTitle(msg *model.BasecampWebhookMessage) (*model.Invoice, error) {
	if msg.Creator.ID == consts.AutoBotID {
		return nil, nil
	}

	reTitle := regexp.MustCompile(`.*([1-9]|0[1-9]|1[0-2])/(20[0-9]{2}) - #(20[0-9]+-[A-Z]+-[0-9]+)`)
	invoiceInfo := reTitle.FindStringSubmatch(msg.Recording.Title)
	if len(invoiceInfo) != 4 {
		return nil, fmt.Errorf(`Todo title have wrong format`)
	}

	invoiceNumber := invoiceInfo[3]

	invoice, err := h.store.Invoice.One(h.repo.DB(), &invoice.Query{Number: invoiceNumber})
	if err != nil {
		return nil, fmt.Errorf(`Can't get invoice %v`, err.Error())
	}

	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		return nil, fmt.Errorf(`Update invoice failed, invoice has status %s`, invoice.Status)
	}

	comments, err := h.service.Basecamp.Comment.Gets(msg.Recording.Bucket.ID, msg.Recording.ID)
	if err != nil {
		return nil, fmt.Errorf(`can't get basecamp comment %v`, err.Error())
	}

	reCmt := regexp.MustCompile(fmt.Sprintf(`(^Paid|^<div>Paid).*#%s`, invoiceNumber))
	for i := range comments {
		// TODO: wtf
		if reCmt.MatchString(strings.ReplaceAll(comments[i].Content, "\n", "")) &&
			!(!(msg.Creator.ID == consts.HanBasecampID) && h.config.Env == "prod") {
			return invoice, nil
		}
	}

	return nil, fmt.Errorf("missing confirm comment")
}
