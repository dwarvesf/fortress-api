package invoice

import (
	"encoding/json"
	"fmt"
	"strings"

	nt "github.com/dstotijn/go-notion"
	"github.com/Rhymond/go-money"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
)

// NotionPageToInvoice transforms a Notion invoice page and its line items to API Invoice model
func NotionPageToInvoice(page nt.Page, lineItems []nt.Page, notionService notion.IService, l logger.Logger) (*model.Invoice, error) {
	l = l.Fields(logger.Fields{
		"function": "NotionPageToInvoice",
		"pageID":   page.ID,
	})

	l.Debug("transforming Notion page to Invoice model")

	// Type assert page properties
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, fmt.Errorf("failed to cast page properties to DatabasePageProperties")
	}

	invoice := &model.Invoice{}

	// Extract Invoice Number from title property
	// Concatenate all title segments (Notion may split styled text into multiple segments)
	if titleProp, ok := props["(auto) Invoice Number"]; ok {
		if len(titleProp.Title) > 0 {
			var parts []string
			for _, segment := range titleProp.Title {
				parts = append(parts, segment.PlainText)
			}
			invoice.Number = strings.Join(parts, "")
			l.Debugf("extracted invoice number: %s (from %d segments)", invoice.Number, len(titleProp.Title))
		}
	}

	// Extract Issue Date -> InvoicedAt
	if issueDateProp, ok := props["Issue Date"]; ok {
		if issueDateProp.Date != nil {
			t := issueDateProp.Date.Start.Time
			invoice.InvoicedAt = &t
			l.Debugf("extracted issue date: %v", invoice.InvoicedAt)
		}
	}

	// Extract Due Date -> DueAt (formula that returns date)
	if dueDateProp, ok := props["Due Date"]; ok {
		if dueDateProp.Formula != nil && dueDateProp.Formula.Date != nil {
			t := dueDateProp.Formula.Date.Start.Time
			invoice.DueAt = &t
			l.Debugf("extracted due date: %v", invoice.DueAt)
		}
	}

	// Extract Paid Date -> PaidAt
	if paidDateProp, ok := props["Paid Date"]; ok {
		if paidDateProp.Date != nil {
			t := paidDateProp.Date.Start.Time
			invoice.PaidAt = &t
			l.Debugf("extracted paid date: %v", invoice.PaidAt)
		}
	}

	// Extract Status and map to API status
	if statusProp, ok := props["Status"]; ok {
		if statusProp.Status != nil {
			invoice.Status = mapNotionStatusToAPI(statusProp.Status.Name)
			l.Debugf("extracted status: %s -> %s", statusProp.Status.Name, invoice.Status)
		}
	}

	// Extract Final Total -> Total (formula that returns number)
	if finalTotalProp, ok := props["Final Total"]; ok {
		if finalTotalProp.Formula != nil && finalTotalProp.Formula.Number != nil {
			invoice.Total = *finalTotalProp.Formula.Number
			l.Debugf("extracted total: %.2f", invoice.Total)
		}
	}

	// Extract Subtotal (formula that returns number)
	if subtotalProp, ok := props["Subtotal"]; ok {
		if subtotalProp.Formula != nil && subtotalProp.Formula.Number != nil {
			invoice.SubTotal = *subtotalProp.Formula.Number
			l.Debugf("extracted subtotal: %.2f", invoice.SubTotal)
		}
	}

	// Extract Tax Rate and calculate Tax amount
	if taxRateProp, ok := props["Tax Rate"]; ok {
		if taxRateProp.Number != nil {
			invoice.Tax = invoice.SubTotal * (*taxRateProp.Number)
			l.Debugf("calculated tax: rate=%.2f, amount=%.2f", *taxRateProp.Number, invoice.Tax)
		}
	}

	// Extract Discount Amount (formula that returns number)
	if discountAmountProp, ok := props["Discount Amount"]; ok {
		if discountAmountProp.Formula != nil && discountAmountProp.Formula.Number != nil {
			invoice.Discount = *discountAmountProp.Formula.Number
			l.Debugf("extracted discount: %.2f", invoice.Discount)
		}
	}

	// Extract Discount Type
	if discountTypeProp, ok := props["Discount Type"]; ok {
		if discountTypeProp.Select != nil {
			invoice.DiscountType = discountTypeProp.Select.Name
			l.Debugf("extracted discount type: %s", invoice.DiscountType)
		}
	}

	// Extract Recipients -> Email (from rollup)
	if recipientsProp, ok := props["Recipients"]; ok {
		if recipientsProp.Rollup != nil && recipientsProp.Rollup.Array != nil && len(recipientsProp.Rollup.Array) > 0 {
			// The rollup array returns DatabasePageProperty items, access first item directly
			firstItem := recipientsProp.Rollup.Array[0]
			if firstItem.Email != nil {
				invoice.Email = *firstItem.Email
				l.Debugf("extracted email: %s", invoice.Email)
			} else if len(firstItem.RichText) > 0 {
				invoice.Email = firstItem.RichText[0].PlainText
				l.Debugf("extracted email from rich text: %s", invoice.Email)
			}
		}
	}

	// Extract Currency and create minimal BankAccount
	// This ensures currency symbols display correctly in Discord/views
	if currencyProp, ok := props["Currency"]; ok {
		if currencyProp.Select != nil {
			currencyCode := currencyProp.Select.Name

			// Use go-money library to get proper currency metadata
			// Handle USDC as USD since go-money doesn't have crypto currencies
			if currencyCode == "USDC" {
				currencyCode = "USD"
			}

			moneyCurrency := money.GetCurrency(currencyCode)
			if moneyCurrency == nil {
				l.Debugf("unknown currency code: %s, using as-is", currencyCode)
				// Fallback: use currency code as both name and symbol
				invoice.Bank = &model.BankAccount{
					Currency: &model.Currency{
						Name:   currencyProp.Select.Name,
						Symbol: currencyProp.Select.Name,
					},
				}
			} else {
				l.Debugf("extracted currency: %s (symbol: %s)", moneyCurrency.Code, moneyCurrency.Grapheme)

				// Create minimal BankAccount with Currency for view layer
				invoice.Bank = &model.BankAccount{
					Currency: &model.Currency{
						Name:   moneyCurrency.Code,
						Symbol: moneyCurrency.Grapheme,
					},
				}
			}
		}
	}

	// Extract Project relation and fetch project name
	if projectProp, ok := props["Project"]; ok {
		if len(projectProp.Relation) > 0 {
			projectID := projectProp.Relation[0].ID
			l.Debugf("found project relation ID: %s", projectID)

			// Fetch the project page to get the project name
			projectPage, err := notionService.GetPage(projectID)
			if err != nil {
				l.AddField("projectID", projectID).Error(err, "failed to fetch project page from Notion")
			} else {
				projectProps, ok := projectPage.Properties.(nt.DatabasePageProperties)
				if !ok {
					l.Debug("failed to cast project properties to DatabasePageProperties")
				} else {
					// Extract project name from "Project" title property
					if projectNameProp, ok := projectProps["Project"]; ok {
						if len(projectNameProp.Title) > 0 {
							// Concatenate all title segments
							var parts []string
							for _, segment := range projectNameProp.Title {
								parts = append(parts, segment.PlainText)
							}
							projectName := strings.Join(parts, "")
							l.Debugf("extracted project name: %s", projectName)

							// Create a minimal Project model with just the name
							// This allows the view layer to access invoice.Project.Name
							invoice.Project = &model.Project{
								Name: projectName,
							}
						}
					}
				}
			}
		} else {
			l.Debug("no project relation found on invoice")
		}
	}

	// Transform line items
	apiLineItems := []model.InvoiceItem{}
	l.Debugf("transforming %d line items", len(lineItems))
	for _, li := range lineItems {
		item, err := notionLineItemToAPI(li, l)
		if err != nil {
			l.AddField("lineItemID", li.ID).Error(err, "failed to transform line item")
			continue
		}
		apiLineItems = append(apiLineItems, item)
	}

	if len(apiLineItems) > 0 {
		lineItemsJSON, err := json.Marshal(apiLineItems)
		if err != nil {
			l.Error(err, "failed to marshal line items to JSON")
		} else {
			invoice.LineItems = model.JSON(lineItemsJSON)
			l.Debugf("marshaled %d line items to JSON", len(apiLineItems))
		}
	}

	// Extract month/year from Issue Date
	if invoice.InvoicedAt != nil {
		invoice.Month = int(invoice.InvoicedAt.Month())
		invoice.Year = invoice.InvoicedAt.Year()
		l.Debugf("extracted month/year: %d/%d", invoice.Month, invoice.Year)
	}

	l.Debug("successfully transformed Notion page to Invoice model")

	return invoice, nil
}

// notionLineItemToAPI transforms a Notion line item page to API InvoiceItem model
func notionLineItemToAPI(page nt.Page, l logger.Logger) (model.InvoiceItem, error) {
	l = l.Fields(logger.Fields{
		"function":   "notionLineItemToAPI",
		"lineItemID": page.ID,
	})

	l.Debug("transforming Notion line item to InvoiceItem model")

	// Type assert page properties
	props, ok := page.Properties.(nt.DatabasePageProperties)
	if !ok {
		return model.InvoiceItem{}, fmt.Errorf("failed to cast page properties to DatabasePageProperties")
	}

	item := model.InvoiceItem{}

	// Extract Quantity
	if qtyProp, ok := props["Quantity"]; ok {
		if qtyProp.Number != nil {
			item.Quantity = *qtyProp.Number
			l.Debugf("extracted quantity: %.2f", item.Quantity)
		}
	}

	// Extract Unit Price -> UnitCost
	if unitPriceProp, ok := props["Unit Price"]; ok {
		if unitPriceProp.Number != nil {
			item.UnitCost = *unitPriceProp.Number
			l.Debugf("extracted unit cost: %.2f", item.UnitCost)
		}
	}

	// Extract Line Total -> Cost (formula that returns number)
	if lineTotalProp, ok := props["Line Total"]; ok {
		if lineTotalProp.Formula != nil && lineTotalProp.Formula.Number != nil {
			item.Cost = *lineTotalProp.Formula.Number
			l.Debugf("extracted cost: %.2f", item.Cost)
		}
	}

	// Extract Description
	if descProp, ok := props["Description"]; ok {
		if len(descProp.RichText) > 0 {
			item.Description = descProp.RichText[0].PlainText
			l.Debugf("extracted description: %s", item.Description)
		}
	}

	// Extract Discount Amount from line item (formula that returns number)
	if discountProp, ok := props["Discount Amount"]; ok {
		if discountProp.Formula != nil && discountProp.Formula.Number != nil {
			item.Discount = *discountProp.Formula.Number
			l.Debugf("extracted discount: %.2f", item.Discount)
		}
	}

	// Extract Discount Type
	if discountTypeProp, ok := props["Discount Type"]; ok {
		if discountTypeProp.Select != nil {
			item.DiscountType = discountTypeProp.Select.Name
			l.Debugf("extracted discount type: %s", item.DiscountType)
		}
	}

	l.Debug("successfully transformed Notion line item to InvoiceItem model")

	return item, nil
}

// mapNotionStatusToAPI maps Notion status to API InvoiceStatus enum
func mapNotionStatusToAPI(notionStatus string) model.InvoiceStatus {
	switch notionStatus {
	case "Draft":
		return model.InvoiceStatusDraft
	case "Sent":
		return model.InvoiceStatusSent
	case "Overdue":
		return model.InvoiceStatusOverdue
	case "Paid":
		return model.InvoiceStatusPaid
	case "Uncollectible":
		return model.InvoiceStatusUncollectible
	case "Canceled":
		return model.InvoiceStatusUncollectible // Map Canceled to Uncollectible
	default:
		return model.InvoiceStatusDraft
	}
}
