package notion

import (
	"context"
	"errors"
	"fmt"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// BankAccountService handles bank account operations with Notion
type BankAccountService struct {
	*baseService
}

// BankAccountData represents bank account data from Notion
type BankAccountData struct {
	PageID             string
	AccountHolderName  string
	BankName           string
	AccountNumber      string
	Currency           string
	Country            string
	AccountType        string // "Local", "Borderless", "International"
	SwiftBIC           string
	IBAN               string
	RoutingNumber      string
	SortCode           string
	BankCode           string
	BranchAddress      string
	PreferredForPayouts bool
	Status             string
}

// NewBankAccountService creates a new Notion bank account service
func NewBankAccountService(cfg *config.Config, l logger.Logger) *BankAccountService {
	base := newBaseService(cfg, l)
	if base == nil {
		return nil
	}

	l.Debug("creating new BankAccountService")

	return &BankAccountService{baseService: base}
}

// QueryBankAccountByDiscord queries bank account by Discord username
// Returns the preferred bank account for payouts, or the first active account if no preferred one is found
func (s *BankAccountService) QueryBankAccountByDiscord(ctx context.Context, discord string) (*BankAccountData, error) {
	bankAccountsDBID := s.cfg.Notion.Databases.BankAccounts
	if bankAccountsDBID == "" {
		return nil, errors.New("bank accounts database ID not configured")
	}

	s.logger.Debug(fmt.Sprintf("querying bank account: discord=%s", discord))

	// Build filter for Discord username and Active status
	query := &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: "Discord",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Rollup: &nt.RollupDatabaseQueryFilter{
							Any: &nt.DatabaseQueryPropertyFilter{
								RichText: &nt.TextPropertyFilter{
									Contains: discord,
								},
							},
						},
					},
				},
				{
					Property: "Status",
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Status: &nt.StatusDatabaseQueryFilter{
							Equals: "Active",
						},
					},
				},
			},
		},
	}

	s.logger.Debug(fmt.Sprintf("executing database query: dbID=%s", bankAccountsDBID))

	resp, err := s.client.QueryDatabase(ctx, bankAccountsDBID, query)
	if err != nil {
		s.logger.Error(err, fmt.Sprintf("failed to query bank accounts database: dbID=%s", bankAccountsDBID))
		return nil, fmt.Errorf("failed to query bank accounts: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("query returned %d results", len(resp.Results)))

	if len(resp.Results) == 0 {
		s.logger.Debug(fmt.Sprintf("no bank account found for discord=%s", discord))
		return nil, fmt.Errorf("no bank account found for contractor: %s", discord)
	}

	// Find preferred account or use first active one
	var selectedPage *nt.Page
	for i := range resp.Results {
		page := &resp.Results[i]
		props := page.Properties.(nt.DatabasePageProperties)

		// Check if this is the preferred account
		if prop, ok := props["Preferred for Payouts"]; ok && prop.Checkbox != nil && *prop.Checkbox {
			s.logger.Debug(fmt.Sprintf("found preferred bank account: pageID=%s", page.ID))
			selectedPage = page
			break
		}

		// If no preferred found yet, use first active account
		if selectedPage == nil {
			selectedPage = page
		}
	}

	if selectedPage == nil {
		s.logger.Debug(fmt.Sprintf("no suitable bank account found for discord=%s", discord))
		return nil, fmt.Errorf("no suitable bank account found for contractor: %s", discord)
	}

	// Extract bank account data
	props := selectedPage.Properties.(nt.DatabasePageProperties)

	bankAccount := &BankAccountData{
		PageID:              selectedPage.ID,
		AccountHolderName:   ExtractRichTextFirst(props, "Account Holder Name"),
		BankName:            ExtractRichTextFirst(props, "Bank Name"),
		AccountNumber:       ExtractRichTextFirst(props, "Account Number"),
		Currency:            ExtractSelect(props, "Currency"),
		Country:             ExtractSelect(props, "Country"),
		AccountType:         ExtractSelect(props, "Account Type"),
		SwiftBIC:            ExtractRichTextFirst(props, "SWIFT / BIC"),
		IBAN:                ExtractRichTextFirst(props, "IBAN"),
		RoutingNumber:       ExtractRichTextFirst(props, "Routing Number"),
		SortCode:            ExtractRichTextFirst(props, "Sort Code"),
		BankCode:            ExtractRichTextFirst(props, "Bank Code"),
		BranchAddress:       ExtractRichTextFirst(props, "Branch / Address"),
		PreferredForPayouts: ExtractCheckbox(props, "Preferred for Payouts"),
		Status:              ExtractStatus(props, "Status"),
	}

	s.logger.Debug(fmt.Sprintf("successfully extracted bank account data: pageID=%s accountHolder=%s bank=%s currency=%s",
		bankAccount.PageID, bankAccount.AccountHolderName, bankAccount.BankName, bankAccount.Currency))

	return bankAccount, nil
}

