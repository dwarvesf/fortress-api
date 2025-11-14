package webhook

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/consts"
	"github.com/dwarvesf/fortress-api/pkg/store/invoice"
)

func (h *handler) GetInvoiceViaBasecampTitle(msg *model.BasecampWebhookMessage) (*model.Invoice, *taskprovider.InvoiceTaskRef, error) {
	if msg.Creator.ID == consts.AutoBotID {
		return nil, nil, nil
	}

	invoiceNumber, err := parseInvoiceNumberFromTitle(msg.Recording.Title)
	if err != nil {
		return nil, nil, err
	}

	invoice, err := h.store.Invoice.One(h.repo.DB(), &invoice.Query{Number: invoiceNumber})
	if err != nil {
		return nil, nil, fmt.Errorf(`Can't get invoice %v`, err.Error())
	}

	if invoice.Status != model.InvoiceStatusSent && invoice.Status != model.InvoiceStatusOverdue {
		return nil, nil, fmt.Errorf(`Update invoice failed, invoice has status %s`, invoice.Status)
	}

	comments, err := h.service.Basecamp.Comment.Gets(msg.Recording.Bucket.ID, msg.Recording.ID)
	if err != nil {
		return nil, nil, fmt.Errorf(`can't get basecamp comment %v`, err.Error())
	}

	reCmt := regexp.MustCompile(fmt.Sprintf(`(^Paid|^<div>Paid).*#%s`, invoiceNumber))
	for i := range comments {
		// TODO: wtf
		if reCmt.MatchString(strings.ReplaceAll(comments[i].Content, "\n", "")) &&
			!(!(msg.Creator.ID == consts.HanBasecampID) && h.config.Env == "prod") {
			ref := &taskprovider.InvoiceTaskRef{
				Provider:   taskprovider.ProviderBasecamp,
				ExternalID: strconv.Itoa(msg.Recording.ID),
				BucketID:   msg.Recording.Bucket.ID,
				TodoID:     msg.Recording.ID,
			}
			return invoice, ref, nil
		}
	}

	return nil, nil, fmt.Errorf("missing confirm comment")
}

func parseInvoiceNumberFromTitle(title string) (string, error) {
	reTitle := regexp.MustCompile(`.*([1-9]|0[1-9]|1[0-2])/(20[0-9]{2}) - #(20[0-9]+-[A-Z0-9]+-[0-9]+)`)
	invoiceInfo := reTitle.FindStringSubmatch(title)
	if len(invoiceInfo) != 4 {
		return "", fmt.Errorf(`Todo title have wrong format`)
	}
	return invoiceInfo[3], nil
}
