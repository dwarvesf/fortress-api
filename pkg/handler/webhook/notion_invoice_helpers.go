package webhook

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

// extractInvoiceDataFromNotion extracts invoice data from Notion page properties
// and transforms it into the model.Invoice structure needed for PDF generation
func (h *handler) extractInvoiceDataFromNotion(l logger.Logger, page nt.Page, props nt.DatabasePageProperties) (*model.Invoice, []model.InvoiceItem, error) {
	l.Debug("extracting invoice data from notion page properties")

	// Extract invoice number from title
	invoiceNumber := ""
	if titleProp, ok := props["(auto) Invoice Number"]; ok {
		l.Debug(fmt.Sprintf("(auto) Invoice Number property found, Title length: %d", len(titleProp.Title)))
		if len(titleProp.Title) > 0 {
			// Concatenate all title segments (Notion may split styled text into multiple segments)
			var parts []string
			for _, segment := range titleProp.Title {
				parts = append(parts, segment.PlainText)
			}
			invoiceNumber = strings.Join(parts, "")
			l.Debug(fmt.Sprintf("extracted invoice number: '%s' (length: %d, bytes: %v)",
				invoiceNumber, len(invoiceNumber), []byte(invoiceNumber)))
			l.Info(fmt.Sprintf("INVOICE NUMBER: %s", invoiceNumber))
		} else {
			l.Debug("(auto) Invoice Number Title is empty")
		}
	} else {
		l.Debug("(auto) Invoice Number property not found in props")
	}
	if invoiceNumber == "" {
		return nil, nil, errors.New("invoice number not found in (auto) Invoice Number title property")
	}

	// Extract issue date
	var issueDate *time.Time
	if dateProp, ok := props["Issue Date"]; ok && dateProp.Date != nil {
		t := dateProp.Date.Start.Time
		issueDate = &t
	}

	// Extract due date (formula field: Issue Date + 7 days)
	var dueDate *time.Time
	if dueProp, ok := props["Due Date"]; ok && dueProp.Formula != nil && dueProp.Formula.Date != nil {
		t := dueProp.Formula.Date.Start.Time
		dueDate = &t
		l.Debug(fmt.Sprintf("extracted due date from formula: %v", dueDate))
	} else if issueDate != nil {
		// Fallback: if Due Date formula doesn't exist, calculate as Issue Date + 7 days
		t := issueDate.AddDate(0, 0, 7)
		dueDate = &t
		l.Debug(fmt.Sprintf("calculated due date as issue date + 7 days: %v", dueDate))
	} else {
		l.Debug("no due date available - both Due Date formula and Issue Date are missing")
	}

	// Extract currency
	currency := "USD" // default
	if currencyProp, ok := props["Currency"]; ok && currencyProp.Select != nil {
		currency = currencyProp.Select.Name
	}

	l.Debug(fmt.Sprintf("currency: %s", currency))

	// Extract totals from formulas
	var subtotal, total, taxRate, discount float64

	if subtotalProp, ok := props["Subtotal"]; ok && subtotalProp.Formula != nil && subtotalProp.Formula.Number != nil {
		subtotal = *subtotalProp.Formula.Number
	}

	if totalProp, ok := props["Final Total"]; ok && totalProp.Formula != nil && totalProp.Formula.Number != nil {
		total = *totalProp.Formula.Number
	}

	if taxProp, ok := props["Tax Rate"]; ok && taxProp.Number != nil {
		taxRate = *taxProp.Number
	}

	if discountProp, ok := props["Discount Value"]; ok && discountProp.Number != nil {
		discount = *discountProp.Number
	}

	l.Debug(fmt.Sprintf("financial data - subtotal: %.2f, total: %.2f, tax: %.2f, discount: %.2f", subtotal, total, taxRate, discount))

	// Fetch project and client data from relations
	project, err := h.extractProjectAndClientFromNotion(l, props)
	if err != nil {
		l.Error(err, "failed to extract project and client data")
		return nil, nil, fmt.Errorf("failed to extract project and client data: %w", err)
	}

	// Extract description and notes
	description := ""
	if descProp, ok := props["Description"]; ok && len(descProp.RichText) > 0 {
		description = descProp.RichText[0].PlainText
	}

	notes := ""
	if notesProp, ok := props["Notes"]; ok && len(notesProp.RichText) > 0 {
		notes = notesProp.RichText[0].PlainText
	}

	// Fetch bank account details from Notion relation
	bankAccount, err := h.extractCompanyBankAccountFromNotion(l, props, currency)
	if err != nil {
		l.Error(err, "failed to extract company bank account, using minimal structure")
		// Fallback to minimal structure
		currencySymbol := currency
		if strings.ToUpper(currency) == "USDC" {
			currencySymbol = "$"
		}
		bankAccount = &model.BankAccount{
			Currency: &model.Currency{
				Name:   currency,
				Symbol: currencySymbol,
			},
		}
	}

	// Assign bank account to project
	project.BankAccount = bankAccount

	// Extract month and year from issue date
	month := time.Now().Month()
	year := time.Now().Year()
	if issueDate != nil {
		month = issueDate.Month()
		year = issueDate.Year()
	}

	// Calculate tax amount
	taxAmount := subtotal * taxRate

	// Build Invoice model
	invoice := &model.Invoice{
		Number:      invoiceNumber,
		InvoicedAt:  issueDate,
		DueAt:       dueDate,
		Status:      model.InvoiceStatusDraft,
		Description: description,
		Note:        notes,
		SubTotal:    subtotal,
		Tax:         taxAmount,
		Discount:    discount,
		Total:       total,
		Month:       int(month),
		Year:        year,
		Project:     project,
		Bank:        bankAccount,
	}

	l.Debug(fmt.Sprintf("built invoice model: number=%s, total=%.2f", invoice.Number, invoice.Total))

	// Fetch and build line items
	lineItems, err := h.extractLineItemsFromNotion(l, page.ID, props)
	if err != nil {
		l.Error(err, "failed to extract line items")
		return nil, nil, fmt.Errorf("failed to extract line items: %w", err)
	}

	l.Info(fmt.Sprintf("extracted invoice data successfully: number=%s, items=%d", invoiceNumber, len(lineItems)))

	return invoice, lineItems, nil
}

// extractLineItemsFromNotion fetches and transforms line items from Notion
func (h *handler) extractLineItemsFromNotion(l logger.Logger, invoicePageID string, invoiceProps nt.DatabasePageProperties) ([]model.InvoiceItem, error) {
	l.Debug(fmt.Sprintf("extracting line items for invoice page: %s", invoicePageID))

	// Get line item relation IDs
	var lineItemIDs []string
	if lineItemProp, ok := invoiceProps["Line Item"]; ok && lineItemProp.Relation != nil {
		for _, rel := range lineItemProp.Relation {
			lineItemIDs = append(lineItemIDs, rel.ID)
		}
	}

	if len(lineItemIDs) == 0 {
		l.Debug("no line items found, returning empty array")
		return []model.InvoiceItem{}, nil
	}

	l.Debug(fmt.Sprintf("found %d line item IDs", len(lineItemIDs)))

	// Fetch each line item page from Notion
	var items []model.InvoiceItem
	for _, itemID := range lineItemIDs {
		l.Debug(fmt.Sprintf("fetching line item page: %s", itemID))

		itemPage, err := h.service.Notion.GetPage(itemID)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to fetch line item page: %s", itemID))
			continue // Skip failed items
		}

		itemProps, ok := itemPage.Properties.(nt.DatabasePageProperties)
		if !ok {
			l.Error(errors.New("invalid properties"), fmt.Sprintf("failed to cast line item properties: %s", itemID))
			continue
		}

		// Extract line item data
		item, err := h.transformNotionLineItem(l, itemProps)
		if err != nil {
			l.Error(err, fmt.Sprintf("failed to transform line item: %s", itemID))
			continue
		}

		items = append(items, item)
	}

	l.Info(fmt.Sprintf("extracted %d line items successfully", len(items)))

	return items, nil
}

// transformNotionLineItem transforms a Notion line item page to model.InvoiceItem
func (h *handler) transformNotionLineItem(l logger.Logger, props nt.DatabasePageProperties) (model.InvoiceItem, error) {
	var item model.InvoiceItem

	// Extract quantity
	if quantityProp, ok := props["Quantity"]; ok && quantityProp.Number != nil {
		item.Quantity = *quantityProp.Number
	} else {
		item.Quantity = 1 // default quantity
	}

	// Extract unit price - fallback to Fixed Amount if Unit Price is empty
	if unitPriceProp, ok := props["Unit Price"]; ok && unitPriceProp.Number != nil {
		item.UnitCost = *unitPriceProp.Number
		l.Debug(fmt.Sprintf("extracted unit price from Unit Price column: %f", item.UnitCost))
	} else if fixedAmountProp, ok := props["Fixed Amount"]; ok && fixedAmountProp.Number != nil {
		item.UnitCost = *fixedAmountProp.Number
		l.Debug(fmt.Sprintf("extracted unit price from Fixed Amount column: %f", item.UnitCost))
	}

	// Extract discount type (from Discount Type select)
	item.DiscountType = "None" // default
	if discountTypeProp, ok := props["Discount Type"]; ok && discountTypeProp.Select != nil {
		item.DiscountType = discountTypeProp.Select.Name
		l.Debug(fmt.Sprintf("extracted discount type: %s", item.DiscountType))
	}

	// Extract discount (from Discount Value number)
	if discountProp, ok := props["Discount Value"]; ok && discountProp.Number != nil {
		item.Discount = *discountProp.Number
		l.Debug(fmt.Sprintf("extracted discount value: %f for type: %s", item.Discount, item.DiscountType))
	}

	// Extract total cost (Line Total formula)
	if totalProp, ok := props["Line Total"]; ok && totalProp.Formula != nil && totalProp.Formula.Number != nil {
		item.Cost = *totalProp.Formula.Number
	}

	// Extract Position from rollup
	var position string
	if positionProp, ok := props["Position"]; ok && positionProp.Rollup != nil {
		if len(positionProp.Rollup.Array) > 0 {
			// Position is a rollup that returns an array, get the first value
			if positionProp.Rollup.Array[0].Select != nil {
				position = positionProp.Rollup.Array[0].Select.Name
			}
		}
	}

	// Extract Contractor relation to get full name
	var contractorName string
	if contractorProp, ok := props["Contractor"]; ok && contractorProp.Rollup != nil {
		if len(contractorProp.Rollup.Array) > 0 {
			// Contractor is a rollup that returns an array of relations
			if len(contractorProp.Rollup.Array[0].Relation) > 0 {
				contractorID := contractorProp.Rollup.Array[0].Relation[0].ID
				l.Debug(fmt.Sprintf("fetching contractor page: %s", contractorID))

				// Fetch contractor page to get Full Name
				contractorPage, err := h.service.Notion.GetPage(contractorID)
				if err != nil {
					l.Error(err, "failed to fetch contractor page")
				} else {
					contractorProps, ok := contractorPage.Properties.(nt.DatabasePageProperties)
					if ok {
						// Extract Full Name from title or name property
						if fullNameProp, ok := contractorProps["Full Name"]; ok && len(fullNameProp.Title) > 0 {
							fullName := fullNameProp.Title[0].PlainText
							l.Debug(fmt.Sprintf("contractor full name: %s", fullName))

							// Reformat name: "First Last Middle" -> "Middle First"
							// First Name = last word, Last Name = first word
							parts := strings.Fields(fullName)
							if len(parts) >= 2 {
								firstName := parts[len(parts)-1] // last word
								lastName := parts[0]             // first word
								contractorName = fmt.Sprintf("%s %s", firstName, lastName)
							} else if len(parts) == 1 {
								contractorName = parts[0]
							}
							l.Debug(fmt.Sprintf("reformatted contractor name: %s", contractorName))
						}
					}
				}
			}
		}
	}

	// Build description with fallback priority:
	// 1. Description column (if not empty)
	// 2. Position - Name format (if available)
	// 3. (auto) Invoice Number (final fallback)

	var description string

	// Check if Description column has content
	if descProp, ok := props["Description"]; ok && len(descProp.RichText) > 0 {
		description = descProp.RichText[0].PlainText
		l.Debug(fmt.Sprintf("using Description column: %s", description))
	}

	// If Description is empty, use Position - Name format
	if description == "" {
		if position != "" && contractorName != "" {
			description = fmt.Sprintf("%s - %s", position, contractorName)
			l.Debug(fmt.Sprintf("Description empty, using Position - Name: %s", description))
		} else if position != "" {
			description = position
			l.Debug(fmt.Sprintf("Description empty, using Position only: %s", description))
		} else if contractorName != "" {
			description = contractorName
			l.Debug(fmt.Sprintf("Description empty, using contractor name only: %s", description))
		}
	}

	// Final fallback to (auto) Invoice Number if still empty
	if description == "" {
		if autoInvoiceNumProp, ok := props["(auto) Invoice Number"]; ok && len(autoInvoiceNumProp.Title) > 0 {
			// Concatenate all title segments
			var parts []string
			for _, segment := range autoInvoiceNumProp.Title {
				parts = append(parts, segment.PlainText)
			}
			description = strings.Join(parts, "")
			l.Debug(fmt.Sprintf("all fallbacks empty, using (auto) Invoice Number: %s", description))
		}
	}

	item.Description = description

	l.Debug(fmt.Sprintf("transformed line item: quantity=%.2f, unit_cost=%.2f, cost=%.2f, desc=%s",
		item.Quantity, item.UnitCost, item.Cost, item.Description))

	return item, nil
}

// extractCompanyBankAccountFromNotion fetches company bank account details from Notion relation
func (h *handler) extractCompanyBankAccountFromNotion(l logger.Logger, invoiceProps nt.DatabasePageProperties, currency string) (*model.BankAccount, error) {
	l.Debug("extracting company bank account from notion relation")

	// Get bank account relation ID
	var bankAccountID string
	if bankProp, ok := invoiceProps["Bank Account"]; ok {
		l.Debug(fmt.Sprintf("Bank Account property found: %+v", bankProp))
		if len(bankProp.Relation) > 0 {
			bankAccountID = bankProp.Relation[0].ID
			l.Debug(fmt.Sprintf("Bank Account relation ID: %s", bankAccountID))
		} else {
			l.Debug("Bank Account relation is empty")
		}
	} else {
		l.Debug("Bank Account property not found in invoice properties")
	}

	if bankAccountID == "" {
		l.Debug("no bank account relation ID found on invoice")
		return nil, errors.New("no bank account relation found on invoice - please set Bank Account relation")
	}

	l.Debug(fmt.Sprintf("fetching bank account page: %s", bankAccountID))

	// Fetch bank account page from Notion
	bankPage, err := h.service.Notion.GetPage(bankAccountID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch bank account page: %s", bankAccountID))
		return nil, fmt.Errorf("failed to fetch bank account page: %w", err)
	}

	bankProps, ok := bankPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast bank account properties")
	}

	// Extract bank account fields
	bankAccount := &model.BankAccount{
		Currency: &model.Currency{
			Name:   currency,
			Symbol: currency, // Default symbol to currency name, will be overridden for USDC
		},
	}

	// Extract account name from title
	if nameProp, ok := bankProps["Name"]; ok && len(nameProp.Title) > 0 {
		bankAccount.Name = nameProp.Title[0].PlainText
	}

	// Extract currency from rich text (override if present)
	if currencyProp, ok := bankProps["Currency"]; ok && len(currencyProp.RichText) > 0 {
		bankAccount.Currency.Name = currencyProp.RichText[0].PlainText
	}

	// Extract details from rich text and parse structured information
	var details string
	if detailsProp, ok := bankProps["Details"]; ok && len(detailsProp.RichText) > 0 {
		details = detailsProp.RichText[0].PlainText
	}

	l.Debug(fmt.Sprintf("parsing bank account details: %s", details))

	// Parse details text to extract individual fields
	// Details format varies by bank, so we use flexible parsing
	if details != "" {
		lines := strings.Split(details, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Split by colon to get key-value pairs
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(strings.ToLower(parts[0]))
			value := strings.TrimSpace(parts[1])

			switch {
			case strings.Contains(key, "account number") || strings.Contains(key, "bank account number"):
				bankAccount.AccountNumber = value
			case strings.Contains(key, "address") && !strings.Contains(key, "intermediary"):
				// Cryptocurrency wallet address or bank address
				if strings.HasPrefix(value, "0x") {
					// This is a blockchain address
					bankAccount.AccountNumber = value
				} else if bankAccount.Address == nil {
					// This is a physical address
					bankAccount.Address = &value
				}
			case strings.Contains(key, "chain"):
				// Blockchain network (e.g., Base, Ethereum)
				bankAccount.BankName = value
			case strings.Contains(key, "bank name"):
				bankAccount.BankName = value
			case strings.Contains(key, "owner name") || strings.Contains(key, "account name") || strings.Contains(key, "global account name"):
				bankAccount.OwnerName = value
			case strings.Contains(key, "swift"):
				bankAccount.SwiftCode = value
			case strings.Contains(key, "routing number") || strings.Contains(key, "ach routing") || strings.Contains(key, "fedwire routing"):
				if bankAccount.RoutingNumber == "" {
					bankAccount.RoutingNumber = value
				}
			case strings.Contains(key, "sort code"):
				bankAccount.UKSortCode = value
			case strings.Contains(key, "location") || strings.Contains(key, "branch"):
				if bankAccount.Address == nil {
					bankAccount.Address = &value
				}
			case strings.Contains(key, "intermediary bank name"):
				bankAccount.IntermediaryBankName = value
			case strings.Contains(key, "intermediary bank address"):
				bankAccount.IntermediaryBankAddress = value
			}
		}
	}

	// Special handling for USDC: use "$" symbol and leave bank name empty
	if strings.ToUpper(bankAccount.Currency.Name) == "USDC" {
		bankAccount.Currency.Symbol = "$"
		bankAccount.BankName = ""
		l.Debug("detected USDC currency: set symbol to '$' and cleared bank name")
	}

	l.Debug(fmt.Sprintf("extracted company bank account: name=%s, bank=%s, account=%s, swift=%s",
		bankAccount.Name, bankAccount.BankName, bankAccount.AccountNumber, bankAccount.SwiftCode))

	return bankAccount, nil
}

// extractProjectAndClientFromNotion fetches project and client data from Notion relations
func (h *handler) extractProjectAndClientFromNotion(l logger.Logger, invoiceProps nt.DatabasePageProperties) (*model.Project, error) {
	l.Debug("extracting project and client from notion relations")

	// Get project relation ID
	var projectID string
	if projectProp, ok := invoiceProps["Project"]; ok && projectProp.Relation != nil && len(projectProp.Relation) > 0 {
		projectID = projectProp.Relation[0].ID
	}

	if projectID == "" {
		return nil, errors.New("no project relation found")
	}

	l.Debug(fmt.Sprintf("fetching project page: %s", projectID))

	// Fetch project page from Notion
	projectPage, err := h.service.Notion.GetPage(projectID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch project page: %s", projectID))
		return nil, fmt.Errorf("failed to fetch project page: %w", err)
	}

	projectProps, ok := projectPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		return nil, errors.New("failed to cast project properties")
	}

	// Extract project name
	var projectName string
	if nameProp, ok := projectProps["Project"]; ok && len(nameProp.Title) > 0 {
		projectName = nameProp.Title[0].PlainText
	}

	l.Debug(fmt.Sprintf("project name: %s", projectName))

	// Get client relation ID from project
	var clientID string
	if clientProp, ok := projectProps["Client"]; ok && clientProp.Relation != nil && len(clientProp.Relation) > 0 {
		clientID = clientProp.Relation[0].ID
	}

	if clientID == "" {
		l.Debug("no client relation found in project, using minimal client")
		// Return project with minimal client
		return &model.Project{
			Name: projectName,
			Client: &model.Client{
				Name: projectName, // Fallback to project name
			},
		}, nil
	}

	l.Debug(fmt.Sprintf("fetching client page: %s", clientID))

	// Fetch client page from Notion
	clientPage, err := h.service.Notion.GetPage(clientID)
	if err != nil {
		l.Error(err, fmt.Sprintf("failed to fetch client page: %s", clientID))
		// Return project with minimal client
		return &model.Project{
			Name: projectName,
			Client: &model.Client{
				Name: projectName,
			},
		}, nil
	}

	clientProps, ok := clientPage.Properties.(nt.DatabasePageProperties)
	if !ok {
		l.Error(errors.New("failed to cast client properties"), "client properties is not DatabasePageProperties")
		// Return project with minimal client
		return &model.Project{
			Name: projectName,
			Client: &model.Client{
				Name: projectName,
			},
		}, nil
	}

	// Extract client data
	client := &model.Client{}

	// Extract client name from title
	if clientNameProp, ok := clientProps["Client Name"]; ok && len(clientNameProp.Title) > 0 {
		client.Name = clientNameProp.Title[0].PlainText
	}

	// Extract client location (address)
	if locationProp, ok := clientProps["Client Location"]; ok && len(locationProp.RichText) > 0 {
		client.Address = locationProp.RichText[0].PlainText
	}

	// Extract contact person
	var contactPerson string
	if contactProp, ok := clientProps["Contact Person"]; ok && len(contactProp.RichText) > 0 {
		contactPerson = contactProp.RichText[0].PlainText
	}

	// Extract contact email
	var contactEmail string
	if emailProp, ok := clientProps["Contact Email"]; ok && emailProp.Email != nil {
		contactEmail = *emailProp.Email
	}

	l.Debug(fmt.Sprintf("extracted client: name=%s, location=%v, contact=%s, email=%s",
		client.Name, client.Address, contactPerson, contactEmail))

	// Build project model
	project := &model.Project{
		Name:   projectName,
		Client: client,
	}

	l.Debug(fmt.Sprintf("built project with client: project=%s, client=%s", projectName, client.Name))

	return project, nil
}

// extractRecipientsFromNotion extracts email addresses from the Recipients rollup property
// Recipients is a rollup from Project → Recipient Emails
func (h *handler) extractRecipientsFromNotion(l logger.Logger, props nt.DatabasePageProperties) ([]string, error) {
	l.Debug("extracting recipients from Recipients rollup")

	recipientsProp, ok := props["Recipients"]
	if !ok {
		l.Debug("Recipients property not found")
		return nil, errors.New("Recipients property not found")
	}

	if recipientsProp.Rollup == nil {
		l.Debug("Recipients rollup is nil")
		return nil, errors.New("Recipients rollup is nil")
	}

	l.Debug(fmt.Sprintf("Recipients rollup type: %s, array length: %d",
		recipientsProp.Rollup.Type, len(recipientsProp.Rollup.Array)))

	var recipients []string

	// The rollup contains an array of rich text values
	for _, item := range recipientsProp.Rollup.Array {
		if len(item.RichText) > 0 {
			for _, rt := range item.RichText {
				email := strings.TrimSpace(rt.PlainText)
				if email != "" {
					recipients = append(recipients, email)
					l.Debug(fmt.Sprintf("extracted recipient email: %s", email))
				}
			}
		}
	}

	l.Debug(fmt.Sprintf("extracted %d recipients total", len(recipients)))

	return recipients, nil
}

// downloadPDFFromNotionAttachment downloads the PDF file from Notion's Preview property
func (h *handler) downloadPDFFromNotionAttachment(l logger.Logger, props nt.DatabasePageProperties) ([]byte, error) {
	l.Debug("downloading PDF from Notion Preview property")

	attachmentProp, ok := props["Preview"]
	if !ok {
		l.Debug("Preview property not found")
		return nil, errors.New("Preview property not found")
	}

	if len(attachmentProp.Files) == 0 {
		l.Debug("Preview property has no files")
		return nil, errors.New("no files in Preview property")
	}

	// Get the first file (should be the PDF)
	file := attachmentProp.Files[0]
	l.Debug(fmt.Sprintf("found attachment file: name=%s, type=%s", file.Name, file.Type))

	// Get the file URL based on file type
	var fileURL string
	if file.Type == nt.FileTypeFile && file.File != nil {
		fileURL = file.File.URL
	} else if file.Type == nt.FileTypeExternal && file.External != nil {
		fileURL = file.External.URL
	} else {
		l.Error(errors.New("unknown file type"), fmt.Sprintf("unsupported file type: %s", file.Type))
		return nil, fmt.Errorf("unsupported file type: %s", file.Type)
	}

	if fileURL == "" {
		l.Debug("file URL is empty")
		return nil, errors.New("file URL is empty")
	}

	l.Debug(fmt.Sprintf("downloading PDF from URL: %s (length: %d)", fileURL[:100], len(fileURL)))

	// Download the file
	resp, err := http.Get(fileURL)
	if err != nil {
		l.Error(err, "failed to download PDF from URL")
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Error(errors.New("non-200 status code"), fmt.Sprintf("HTTP status: %d", resp.StatusCode))
		return nil, fmt.Errorf("failed to download PDF: HTTP %d", resp.StatusCode)
	}

	// Read the file content
	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read PDF content")
		return nil, fmt.Errorf("failed to read PDF content: %w", err)
	}

	l.Debug(fmt.Sprintf("successfully downloaded PDF: size=%d bytes", len(pdfBytes)))

	return pdfBytes, nil
}
