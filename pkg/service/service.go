package service

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/youtube/v3"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp"
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
	"github.com/dwarvesf/fortress-api/pkg/service/communitynft"
	"github.com/dwarvesf/fortress-api/pkg/service/currency"
	"github.com/dwarvesf/fortress-api/pkg/service/discord"
	"github.com/dwarvesf/fortress-api/pkg/service/duckdb"
	"github.com/dwarvesf/fortress-api/pkg/service/evm"
	"github.com/dwarvesf/fortress-api/pkg/service/github"
	googleauth "github.com/dwarvesf/fortress-api/pkg/service/google"
	"github.com/dwarvesf/fortress-api/pkg/service/googleadmin"
	"github.com/dwarvesf/fortress-api/pkg/service/googledrive"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/googlesheet"
	"github.com/dwarvesf/fortress-api/pkg/service/googlestorage"
	"github.com/dwarvesf/fortress-api/pkg/service/icyswap"
	"github.com/dwarvesf/fortress-api/pkg/service/improvmx"
	"github.com/dwarvesf/fortress-api/pkg/service/landingzone"
	"github.com/dwarvesf/fortress-api/pkg/service/lobsters"
	"github.com/dwarvesf/fortress-api/pkg/service/mochi"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
	"github.com/dwarvesf/fortress-api/pkg/service/mochiprofile"
	"github.com/dwarvesf/fortress-api/pkg/service/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/ogifmemosummarizer"
	"github.com/dwarvesf/fortress-api/pkg/service/openrouter"
	"github.com/dwarvesf/fortress-api/pkg/service/parquet"
	"github.com/dwarvesf/fortress-api/pkg/service/reddit"
	"github.com/dwarvesf/fortress-api/pkg/service/sendgrid"
	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
	tpbasecamp "github.com/dwarvesf/fortress-api/pkg/service/taskprovider/basecamp"
	tpnocodb "github.com/dwarvesf/fortress-api/pkg/service/taskprovider/nocodb"
	"github.com/dwarvesf/fortress-api/pkg/service/tono"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
	yt "github.com/dwarvesf/fortress-api/pkg/service/youtube"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type Service struct {
	Basecamp                     *basecamp.Service
	TaskProvider                 taskprovider.InvoiceProvider
	AccountingProvider           taskprovider.AccountingProvider
	ExpenseProvider              taskprovider.ExpenseProvider // Webhook expense provider
	PayrollExpenseProvider       basecamp.ExpenseProvider     // Payroll expense fetcher (Basecamp or NocoDB expense_submissions)
	PayrollAccountingTodoProvider basecamp.ExpenseProvider     // Payroll accounting todo fetcher (NocoDB accounting_todos)
	NocoDB                       *nocodb.Service
	Cache                   *cache.Cache
	Currency                currency.IService
	Discord                 discord.IService
	DuckDB                  duckdb.IService
	Github                  github.IService
	Google                  googleauth.IService
	GoogleStorage           googlestorage.IService
	GoogleAdmin             googleadmin.IService
	GoogleDrive             googledrive.IService
	GoogleMail              googlemail.IService
	GoogleSheet             googlesheet.IService
	ImprovMX                improvmx.IService
	Mochi                   mochi.IService
	MochiPay                mochipay.IService
	MochiProfile            mochiprofile.IService
	Notion                  *notion.Services
	OpenRouter              *openrouter.OpenRouterService
	ParquetSync             parquet.ISyncService
	Sendgrid                sendgrid.IService
	Wise                    wise.IService
	BaseClient              evm.IService
	IcySwap                 icyswap.IService
	CommunityNft            communitynft.IService
	Tono                    tono.IService
	Reddit                  reddit.IService
	Lobsters                lobsters.IService
	Youtube                 yt.IService
	Dify                    ogifmemosummarizer.IService
	LandingZone             landingzone.IService
}

func New(cfg *config.Config, store *store.Store, repo store.DBRepo) (*Service, error) {
	cch := cache.New(5*time.Minute, 10*time.Minute)

	googleAuthSvc, err := googleauth.New(
		&oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			Endpoint:     google.Endpoint,
			Scopes:       []string{"email", "profile"},
		},
	)
	if err != nil {
		return nil, err
	}

	gcsSvc, err := googlestorage.New(
		cfg.Google.GCSBucketName,
		cfg.Google.GCSProjectID,
		cfg.Google.GCSCredentials,
	)
	if err != nil {
		return nil, err
	}

	landingZoneSvc, err := landingzone.New(
		cfg.Google.GCSLandingZoneCredentials,
	)
	if err != nil {
		return nil, err
	}

	driveConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{googledrive.FullDriveAccessScope},
	}

	googleDriveSvc := googledrive.New(driveConfig, cfg, logger.L)

	googleAdminConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{admin.AdminDirectoryUserScope,
			admin.AdminDirectoryGroupScope,
		},
	}
	googleAdminSvc := googleadmin.New(googleAdminConfig, cfg)

	mailConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			gmail.MailGoogleComScope,
			"https://www.googleapis.com/auth/gmail.settings.basic",
			"https://www.googleapis.com/auth/gmail.settings.sharing",
		},
	}

	googleMailSvc := googlemail.New(
		mailConfig,
		cfg,
	)

	youtubeSvc := yt.New(&oauth2.Config{
		ClientID:     cfg.Youtube.ClientID,
		ClientSecret: cfg.Youtube.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{youtube.YoutubeForceSslScope},
	}, cfg)

	gSheetConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			sheets.DriveScope,
			sheets.DriveFileScope,
			sheets.DriveReadonlyScope,
			sheets.SpreadsheetsScope,
			sheets.SpreadsheetsReadonlyScope,
		},
	}

	gSheetSvc := googlesheet.New(
		gSheetConfig,
		cfg,
	)

	bc := model.Basecamp{
		ClientID:     cfg.Basecamp.ClientID,
		ClientSecret: cfg.Basecamp.ClientSecret,
	}

	Currency := currency.New(cfg)

	baseClient, err := evm.New(evm.DefaultBASEClient, cfg, logger.L)
	if err != nil {
		return nil, err
	}
	icySwap, err := icyswap.New(baseClient, cfg, logger.L)
	if err != nil {
		return nil, err
	}

	communityNft, err := communitynft.New(baseClient, cfg, logger.L)
	if err != nil {
		return nil, err
	}

	reddit, err := reddit.New(cfg, logger.L)
	if err != nil {
		return nil, err
	}

	difySvc := ogifmemosummarizer.New(cfg)

	duckDBSvc, err := duckdb.New(logger.L)
	if err != nil {
		return nil, err
	}

	parquetSvc := parquet.NewSyncService(cfg.Parquet, logger.L)
	wiseSvc := wise.New(cfg, logger.L)
	basecampSvc := basecamp.New(store, repo, cfg, &bc, logger.L)
	nocoSvc := nocodb.New(cfg.Noco)
	basecampTaskProvider := tpbasecamp.New(basecampSvc, cfg)
	var nocodbProvider *tpnocodb.Provider
	if nocoSvc != nil {
		nocodbProvider = tpnocodb.New(nocoSvc, cfg, store, repo, wiseSvc, logger.L)
	}

	selectedProvider := strings.ToLower(cfg.TaskProvider)

	var invoiceProvider taskprovider.InvoiceProvider
	if selectedProvider == string(taskprovider.ProviderNocoDB) && nocodbProvider != nil {
		invoiceProvider = nocodbProvider
	}
	if invoiceProvider == nil {
		invoiceProvider = basecampTaskProvider
	}

	var accountingProvider taskprovider.AccountingProvider
	if selectedProvider == string(taskprovider.ProviderNocoDB) && nocodbProvider != nil {
		accountingProvider = nocodbProvider
	}
	if accountingProvider == nil {
		accountingProvider = basecampTaskProvider
	}

	var expenseProvider taskprovider.ExpenseProvider
	if selectedProvider == string(taskprovider.ProviderNocoDB) && nocodbProvider != nil {
		expenseProvider = nocodbProvider
	}
	if expenseProvider == nil {
		expenseProvider = basecampTaskProvider
	}

	// Payroll expense fetcher provider (for fetching expense_submissions during payroll calculation)
	var payrollExpenseProvider basecamp.ExpenseProvider
	if selectedProvider == string(taskprovider.ProviderNotion) {
		// Use Notion expense service for payroll
		payrollExpenseProvider = notion.NewExpenseService(cfg, store, repo, logger.L)
	} else if selectedProvider == string(taskprovider.ProviderNocoDB) && nocoSvc != nil {
		// Use NocoDB expense service for payroll (expense_submissions table)
		payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)
	}
	if payrollExpenseProvider == nil {
		// Fallback to Basecamp adapter
		payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
	}

	// Payroll accounting todo fetcher (for fetching accounting_todos during payroll calculation)
	// Only used for NocoDB - Notion fetches all expenses from one table, Basecamp uses getAccountingExpense()
	var payrollAccountingTodoProvider basecamp.ExpenseProvider
	if selectedProvider == string(taskprovider.ProviderNocoDB) && nocoSvc != nil {
		// Use NocoDB accounting todo service for payroll (accounting_todos table)
		payrollAccountingTodoProvider = nocodb.NewAccountingTodoService(nocoSvc, cfg, store, repo, logger.L)
	} else if selectedProvider == string(taskprovider.ProviderBasecamp) && basecampSvc != nil {
		// Fallback to Basecamp adapter only if Basecamp is actually configured
		payrollAccountingTodoProvider = basecamp.NewExpenseAdapter(basecampSvc)
	}
	// For Notion provider, payrollAccountingTodoProvider remains nil (all expenses fetched via PayrollExpenseProvider)

	return &Service{
		Basecamp:                     basecampSvc,
		TaskProvider:                 invoiceProvider,
		AccountingProvider:           accountingProvider,
		ExpenseProvider:              expenseProvider,
		PayrollExpenseProvider:       payrollExpenseProvider,
		PayrollAccountingTodoProvider: payrollAccountingTodoProvider,
		NocoDB:                       nocoSvc,
		Cache:              cch,
		Currency:           Currency,
		Discord:            discord.New(cfg),
		DuckDB:             duckDBSvc,
		Github:             github.New(cfg, logger.L),
		Google:             googleAuthSvc,
		GoogleStorage:      gcsSvc,
		GoogleAdmin:        googleAdminSvc,
		GoogleDrive:        googleDriveSvc,
		GoogleMail:         googleMailSvc,
		GoogleSheet:        gSheetSvc,
		ImprovMX:           improvmx.New(cfg.ImprovMX.Token),
		Mochi:              mochi.New(cfg, logger.L),
		MochiPay:           mochipay.New(cfg, logger.L),
		MochiProfile:       mochiprofile.New(cfg, logger.L),
		Notion: &notion.Services{
			IService:          notion.New(cfg.Notion.Secret, cfg.Notion.Databases.Project, logger.L, repo.DB()),
			Timesheet:         notion.NewTimesheetService(cfg, logger.L),
			TaskOrderLog:      notion.NewTaskOrderLogService(cfg, logger.L),
			ContractorRates:   notion.NewContractorRatesService(cfg, logger.L),
			ContractorFees:    notion.NewContractorFeesService(cfg, logger.L),
			ContractorPayouts: notion.NewContractorPayoutsService(cfg, logger.L),
			RefundRequests:    notion.NewRefundRequestsService(cfg, logger.L),
			InvoiceSplit:      notion.NewInvoiceSplitService(cfg, logger.L),
		},
		OpenRouter:     openrouter.NewOpenRouterService(cfg, logger.L),
		ParquetSync:    parquetSvc,
		Sendgrid:           sendgrid.New(cfg.Sendgrid.APIKey, cfg, logger.L),
		Wise:               wiseSvc,
		BaseClient:         baseClient,
		IcySwap:            icySwap,
		CommunityNft:       communityNft,
		Tono:               tono.New(cfg, logger.L),
		Reddit:             reddit,
		Lobsters:           lobsters.New(),
		Youtube:            youtubeSvc,
		Dify:               difySvc,
		LandingZone:        landingZoneSvc,
	}, nil
}

// NewForTest returns a Service with all external services set to nil or mocked for testing purposes.
func NewForTest() *Service {
	return &Service{
		Cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}
