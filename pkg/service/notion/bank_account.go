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
	client *nt.Client
	cfg    *config.Config
	logger logger.Logger
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
func NewBankAccountService(cfg *config.Config, logger logger.Logger) *BankAccountService {
	if cfg.Notion.Secret == "" {
		logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
		return nil
	}

	logger.Debug("creating new BankAccountService")

	return &BankAccountService{
		client: nt.NewClient(cfg.Notion.Secret),
		cfg:    cfg,
		logger: logger,
	}
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
		AccountHolderName:   s.extractRichText(props, "Account Holder Name"),
		BankName:            s.extractRichText(props, "Bank Name"),
		AccountNumber:       s.extractRichText(props, "Account Number"),
		Currency:            s.extractSelect(props, "Currency"),
		Country:             s.extractSelect(props, "Country"),
		AccountType:         s.extractSelect(props, "Account Type"),
		SwiftBIC:            s.extractRichText(props, "SWIFT / BIC"),
		IBAN:                s.extractRichText(props, "IBAN"),
		RoutingNumber:       s.extractRichText(props, "Routing Number"),
		SortCode:            s.extractRichText(props, "Sort Code"),
		BankCode:            s.extractRichText(props, "Bank Code"),
		BranchAddress:       s.extractRichText(props, "Branch / Address"),
		PreferredForPayouts: s.extractCheckbox(props, "Preferred for Payouts"),
		Status:              s.extractStatus(props, "Status"),
	}

	s.logger.Debug(fmt.Sprintf("successfully extracted bank account data: pageID=%s accountHolder=%s bank=%s currency=%s",
		bankAccount.PageID, bankAccount.AccountHolderName, bankAccount.BankName, bankAccount.Currency))

	return bankAccount, nil
}

// Helper functions for extracting properties

func (s *BankAccountService) extractRichText(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.RichText == nil || len(prop.RichText) == 0 {
		return ""
	}
	return prop.RichText[0].PlainText
}

func (s *BankAccountService) extractSelect(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Select == nil {
		return ""
	}
	return prop.Select.Name
}

func (s *BankAccountService) extractCheckbox(props nt.DatabasePageProperties, propName string) bool {
	prop, ok := props[propName]
	if !ok || prop.Checkbox == nil {
		return false
	}
	return *prop.Checkbox
}

func (s *BankAccountService) extractStatus(props nt.DatabasePageProperties, propName string) string {
	prop, ok := props[propName]
	if !ok || prop.Status == nil {
		return ""
	}
	return prop.Status.Name
}
