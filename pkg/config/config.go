package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	defaultBasecampAccountingProjectID = 15258324
	defaultBasecampAccountingTodoSetID = 2329633561
	defaultBasecampPlaygroundProjectID = 12984857
	defaultBasecampPlaygroundTodoSetID = 1941398075
)

// Loader load config from reader into Viper
type Loader interface {
	Load(viper.Viper) (*viper.Viper, error)
}

type Config struct {
	// db
	Postgres DBConnection

	// server
	ApiServer ApiServer

	// service
	Google                Google
	Vault                 Vault
	Notion                Notion
	Wise                  Wise
	Discord               Discord
	Basecamp              Basecamp
	CurrencyLayer         CurrencyLayer
	Mochi                 Mochi
	MochiPay              MochiPay
	MochiProfile          MochiProfile
	Tono                  Tono
	ImprovMX              ImprovMX
	CommunityNft          CommunityNft
	Reddit                Reddit
	Youtube               Youtube
	Dify                  Dify
	Parquet               Parquet
	Noco                  Noco
	TaskProvider          string
	AccountingIntegration AccountingIntegration
	ExpenseIntegration    ExpenseIntegration
	LeaveIntegration      LeaveIntegration

	Invoice         Invoice
	Sendgrid        Sendgrid
	Github          Github
	CheckIn         CheckIn
	InvoiceListener InvoiceListener

	OpenRouter OpenRouter

	APIKey                     string
	Debug                      bool
	DBMonitoringEnabled        bool
	Env                        string
	JWTSecretKey               string
	FortressURL                string
	LogLevel                   string
	TaskOrderLogWorkerPoolSize int

	// Concurrency controls for contractor payouts
	TaskOrderLogSubitemConcurrency int  // Max concurrent subitem updates (default: 10, max: 20)
	RefundProcessingWorkers        int  // Worker pool for refunds (default: 5, max: 20)
	InvoiceSplitProcessingWorkers  int  // Worker pool for splits (default: 5, max: 20)
	EnableOrderAPIConcurrency      bool // Feature flag for TIER 3 (default: true)
}

func getIntWithDefault(v ENV, key string, fallback int) int {
	if val := v.GetString(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func getStringWithDefault(v ENV, key, fallback string) string {
	if val := v.GetString(key); val != "" {
		return val
	}
	return fallback
}

func getBoolWithDefault(v ENV, key string, fallback bool) bool {
	if val := v.GetString(key); val != "" {
		return strings.ToLower(val) == "true" || val == "1"
	}
	return fallback
}

func parseKeyValuePairs(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	result := make(map[string]string)
	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// validateLogLevel validates and normalizes the log level string
// Returns the normalized level or "info" as default
func validateLogLevel(level string) string {
	// Normalize to lowercase
	normalized := strings.ToLower(strings.TrimSpace(level))

	// Valid logrus levels: trace, debug, info, warn, error, fatal, panic
	validLevels := map[string]bool{
		"trace":   true,
		"debug":   true,
		"info":    true,
		"warn":    true,
		"warning": true, // alias for warn
		"error":   true,
		"fatal":   true,
		"panic":   true,
	}

	if validLevels[normalized] {
		// Convert warning to warn for consistency with logrus
		if normalized == "warning" {
			return "warn"
		}
		return normalized
	}

	// Default to info for invalid levels
	return "info"
}

type DBConnection struct {
	Host string
	Port string
	User string
	Name string
	Pass string

	SSLMode string
}

type ApiServer struct {
	Port           string
	AllowedOrigins string
}

type Google struct {
	ClientSecret                 string
	ClientID                     string
	AppName                      string
	GCSProjectID                 string
	GCSBucketName                string
	GCSCredentials               string
	GCPProjectID                 string
	AccountingGoogleRefreshToken string
	AdminGoogleRefreshToken      string
	AccountingEmailID            string
	TeamGoogleRefreshToken       string
	TeamEmailID                  string
	GCSLandingZoneCredentials    string
	DwarvesCalendarID            string
}

type Youtube struct {
	ClientSecret string
	ClientID     string
	RefreshToken string
}

type Dify struct {
	URL   string
	Token string
}

type Parquet struct {
	LocalFilePath   string
	SyncInterval    string
	RemoteURL       string
	QuickTimeout    string
	ExtendedTimeout string
	EnableCaching   bool
}

type Wise struct {
	APIKey     string
	Profile    string
	Url        string
	UseRealAPI bool // Allow using real Wise API in non-prod environments
}

type CurrencyLayer struct {
	APIKey string
}

type Vault struct {
	Address string
	Token   string
	Path    string
}

type ImprovMX struct {
	Token string
}

type Mochi struct {
	BaseURL         string
	ApplicationID   string
	ApplicationName string
	APIKey          string
}

type MochiPay struct {
	BaseURL string
}

type MochiProfile struct {
	BaseURL string
}

type Tono struct {
	BaseURL string
}

type CommunityNft struct {
	ContractAddress string
}

type Notion struct {
	Secret    string
	Databases NotionDatabase
}

type Github struct {
	Token             string
	BraineryReviewers []string
}

type OpenRouter struct {
	APIKey string
	Model  string
}

type NotionDatabase struct {
	AuditCycle         string
	AuditActionItem    string
	Earn               string
	TechRadar          string
	Audience           string
	Event              string
	Hiring             string
	StaffingDemand     string
	Project            string
	Delivery           string
	Digest             string
	Updates            string
	Memo               string
	Issue              string
	Contractor         string
	ContractorRates    string
	ContractorPayouts  string
	ContractorPayables string
	InvoiceSplit       string
	RefundRequest      string
	DeploymentTracker  string
	Timesheet          string
	TaskOrderLog       string
	BankAccounts       string
}

type Discord struct {
	SecretToken   string
	ApplicationID string
	PublicKey     string
	Webhooks      DiscordWebhook
	IDs           DiscordID
}

type DiscordWebhook struct {
	Campfire     string
	AuditLog     string
	ICYPublicLog string
}

type DiscordID struct {
	DwarvesGuild    string
	EventsChannel   string
	GolangChannel   string
	ResearchChannel string
	OnLeaveChannel  string
}

type Invoice struct {
	TemplatePath           string
	DirID                  string
	ContractorInvoiceDirID string
	TestEmail              string
}

type Sendgrid struct {
	APIKey string
}

type Basecamp struct {
	BotKey              string
	ClientID            string
	ClientSecret        string
	OAuthRefreshToken   string
	AccountingProjectID int
	AccountingTodoSetID int
	PlaygroundProjectID int
	PlaygroundTodoSetID int
}

type Noco struct {
	BaseURL                       string
	Token                         string
	WorkspaceID                   string
	BaseID                        string
	InvoiceTableID                string
	InvoiceCommentsTableID        string
	WebhookSecret                 string
	AccountingTodosTableID        string
	AccountingTransactionsTableID string
}

type AccountingIntegration struct {
	Basecamp AccountingBasecampIntegration
	Noco     AccountingNocoIntegration
}

type AccountingBasecampIntegration struct {
	ProjectID int
	TodoSetID int
	GroupIn   string
	GroupOut  string
}

type AccountingNocoIntegration struct {
	TodosTableID        string
	TransactionsTableID string
	WebhookSecret       string
}

type ExpenseIntegration struct {
	Noco            ExpenseNocoIntegration
	Notion          ExpenseNotionIntegration
	ApproverMapping map[string]string
}

type ExpenseNocoIntegration struct {
	WorkspaceID   string
	TableID       string
	WebhookSecret string
}

type ExpenseNotionIntegration struct {
	ExpenseDBID    string
	ContractorDBID string
	DataSourceID   string // For multi-source databases, use data source ID instead of database ID
}

type LeaveIntegration struct {
	Noco   LeaveNocoIntegration
	Notion LeaveNotionIntegration
}

type LeaveNocoIntegration struct {
	TableID       string
	WebhookSecret string
}

type LeaveNotionIntegration struct {
	LeaveDBID         string
	ContractorDBID    string
	DataSourceID      string // For multi-source databases, use data source ID instead of database ID
	VerificationToken string // Token for verifying webhook signatures
}

type ENV interface {
	GetBool(string) bool
	GetString(string) string
}

type Reddit struct {
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
}

type CheckIn struct {
	WhitelistedEmployeeIDs []string
}

// InvoiceListener configuration for Gmail inbox monitoring
type InvoiceListener struct {
	Enabled        bool          // Feature toggle
	Email          string        // Email address to monitor (e.g., bill@d.foundation)
	RefreshToken   string        // OAuth refresh token for the inbox
	PollInterval   time.Duration // Poll interval (e.g., 5m)
	ProcessedLabel string        // Gmail label for processed emails
	MaxMessages    int64         // Maximum messages to process per batch
	PDFMaxSizeMB   int           // Maximum PDF size in MB
	PDFFallback    bool          // Enable PDF attachment fallback when subject match fails
}

// parseInvoiceListenerConfig parses InvoiceListener configuration from environment variables
func parseInvoiceListenerConfig(v ENV) InvoiceListener {
	// Parse poll interval with default of 5 minutes
	pollIntervalStr := getStringWithDefault(v, "INVOICE_LISTENER_POLL_INTERVAL", "5m")
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		pollInterval = 5 * time.Minute
	}

	return InvoiceListener{
		Enabled:        getBoolWithDefault(v, "INVOICE_LISTENER_ENABLED", false),
		Email:          v.GetString("INVOICE_LISTENER_EMAIL"),
		RefreshToken:   v.GetString("INVOICE_LISTENER_REFRESH_TOKEN"),
		PollInterval:   pollInterval,
		ProcessedLabel: getStringWithDefault(v, "INVOICE_LISTENER_LABEL", "fortress-api/processed"),
		MaxMessages:    int64(getIntWithDefault(v, "INVOICE_LISTENER_MAX_MESSAGES", 50)),
		PDFMaxSizeMB:   getIntWithDefault(v, "INVOICE_LISTENER_PDF_MAX_SIZE_MB", 5),
		PDFFallback:    getBoolWithDefault(v, "INVOICE_LISTENER_PDF_FALLBACK", false),
	}
}

func Generate(v ENV) *Config {
	// Variables used in multiple places
	basecampAccountingProjectID := getIntWithDefault(v, "BASECAMP_ACCOUNTING_PROJECT_ID", defaultBasecampAccountingProjectID)
	basecampAccountingTodoSetID := getIntWithDefault(v, "BASECAMP_ACCOUNTING_TODO_SET_ID", defaultBasecampAccountingTodoSetID)
	nocoWebhookSecret := v.GetString("NOCO_WEBHOOK_SECRET")
	accountingTodosTableID := v.GetString("NOCO_ACCOUNTING_TODOS_TABLE_ID")
	accountingTransactionsTableID := v.GetString("NOCO_ACCOUNTING_TRANSACTIONS_TABLE_ID")
	notionContractorDBID := v.GetString("NOTION_CONTRACTOR_DB_ID")
	logLevel := validateLogLevel(v.GetString("LOG_LEVEL"))

	// Parse and validate TaskOrderLogWorkerPoolSize
	workerPoolSize := getIntWithDefault(v, "TASK_ORDER_LOG_WORKER_POOL_SIZE", 5)
	if workerPoolSize < 1 {
		workerPoolSize = 1
	}
	if workerPoolSize > 20 {
		workerPoolSize = 20
	}

	// Parse and validate concurrency parameters
	subitemConcurrency := getIntWithDefault(v, "TASK_ORDER_LOG_SUBITEM_CONCURRENCY", 10)
	if subitemConcurrency < 1 {
		subitemConcurrency = 1
	}
	if subitemConcurrency > 20 {
		subitemConcurrency = 20
	}

	refundWorkers := getIntWithDefault(v, "REFUND_PROCESSING_WORKERS", 5)
	if refundWorkers < 1 {
		refundWorkers = 1
	}
	if refundWorkers > 20 {
		refundWorkers = 20
	}

	invoiceSplitWorkers := getIntWithDefault(v, "INVOICE_SPLIT_PROCESSING_WORKERS", 5)
	if invoiceSplitWorkers < 1 {
		invoiceSplitWorkers = 1
	}
	if invoiceSplitWorkers > 20 {
		invoiceSplitWorkers = 20
	}

	return &Config{
		Debug:                          v.GetBool("DEBUG"),
		DBMonitoringEnabled:            getBoolWithDefault(v, "DB_MONITORING_ENABLED", false),
		APIKey:                         v.GetString("API_KEY"),
		Env:                            v.GetString("ENV"),
		JWTSecretKey:                   v.GetString("JWT_SECRET_KEY"),
		FortressURL:                    v.GetString("FORTRESS_URL"),
		LogLevel:                       logLevel,
		TaskOrderLogWorkerPoolSize:     workerPoolSize,
		TaskOrderLogSubitemConcurrency: subitemConcurrency,
		RefundProcessingWorkers:        refundWorkers,
		InvoiceSplitProcessingWorkers:  invoiceSplitWorkers,
		EnableOrderAPIConcurrency:      getBoolWithDefault(v, "ENABLE_ORDER_API_CONCURRENCY", true),

		ApiServer: ApiServer{
			Port:           v.GetString("PORT"),
			AllowedOrigins: v.GetString("ALLOWED_ORIGINS"),
		},

		Postgres: DBConnection{
			Host:    v.GetString("DB_HOST"),
			Port:    v.GetString("DB_PORT"),
			User:    v.GetString("DB_USER"),
			Name:    v.GetString("DB_NAME"),
			Pass:    v.GetString("DB_PASS"),
			SSLMode: v.GetString("DB_SSL_MODE"),
		},
		Github: Github{
			Token:             v.GetString("GITHUB_ACCESS_TOKEN"),
			BraineryReviewers: strings.Split(v.GetString("BRAINERY_REVIEWERS"), ","),
		},
		OpenRouter: OpenRouter{
			APIKey: v.GetString("OPENROUTER_API_KEY"),
			Model:  v.GetString("OPENROUTER_MODEL"),
		},
		Google: Google{
			AccountingEmailID:            v.GetString("ACCOUNTING_EMAIL_ID"),
			AccountingGoogleRefreshToken: v.GetString("ACCOUNTING_GOOGLE_REFRESH_TOKEN"),
			AdminGoogleRefreshToken:      v.GetString("ADMIN_GOOGLE_REFRESH_TOKEN"),
			AppName:                      v.GetString("GOOGLE_API_APP_NAME"),
			ClientID:                     v.GetString("GOOGLE_API_CLIENT_ID"),
			ClientSecret:                 v.GetString("GOOGLE_API_CLIENT_SECRET"),
			GCPProjectID:                 v.GetString("GCP_PROJECT_ID"),
			GCSBucketName:                v.GetString("GCS_BUCKET_NAME"),
			GCSCredentials:               v.GetString("GCS_CREDENTIALS"),
			GCSProjectID:                 v.GetString("GCS_PROJECT_ID"),
			TeamEmailID:                  v.GetString("TEAM_EMAIL_ID"),
			TeamGoogleRefreshToken:       v.GetString("TEAM_GOOGLE_REFRESH_TOKEN"),
			GCSLandingZoneCredentials:    v.GetString("GCS_LD_ZONE_CREDENTIALS"),
			DwarvesCalendarID:            v.GetString("DWARVES_CALENDAR_ID"),
		},
		Youtube: Youtube{
			ClientID:     v.GetString("YOUTUBE_API_CLIENT_ID"),
			ClientSecret: v.GetString("YOUTUBE_API_CLIENT_SECRET"),
			RefreshToken: v.GetString("YOUTUBE_REFRESH_TOKEN"),
		},

		Wise: Wise{
			APIKey:     v.GetString("WISE_API_KEY"),
			Profile:    v.GetString("WISE_PROFILE"),
			Url:        v.GetString("WISE_URL"),
			UseRealAPI: v.GetBool("WISE_USE_REAL_API"),
		},
		CurrencyLayer: CurrencyLayer{
			APIKey: v.GetString("CURRENCY_LAYER_API_KEY"),
		},
		Vault: Vault{
			Address: v.GetString("VAULT_ADDR"),
			Token:   v.GetString("VAULT_TOKEN"),
			Path:    v.GetString("VAULT_PATH"),
		},
		Notion: Notion{
			Secret: v.GetString("NOTION_SECRET"),
			Databases: NotionDatabase{
				AuditCycle:         v.GetString("NOTION_AUDIT_CYCLE_DB_ID"),
				AuditActionItem:    v.GetString("NOTION_AUDIT_ACTION_ITEM_DB_ID"),
				Earn:               v.GetString("NOTION_EARN_DB_ID"),
				TechRadar:          v.GetString("NOTION_TECH_RADAR_DB_ID"),
				Audience:           v.GetString("NOTION_AUDIENCE_DB_ID"),
				Event:              v.GetString("NOTION_EVENT_DB_ID"),
				Hiring:             v.GetString("NOTION_HIRING_DB_ID"),
				StaffingDemand:     v.GetString("NOTION_STAFFING_DEMAND_DB_ID"),
				Project:            v.GetString("NOTION_PROJECT_DB_ID"),
				Delivery:           v.GetString("NOTION_DELIVERY_DB_ID"),
				Digest:             v.GetString("NOTION_DIGEST_DB_ID"),
				Updates:            v.GetString("NOTION_UPDATES_DB_ID"),
				Memo:               v.GetString("NOTION_MEMO_DB_ID"),
				Issue:              v.GetString("NOTION_ISSUE_DB_ID"),
				Contractor:         v.GetString("NOTION_CONTRACTOR_DB_ID"),
				ContractorRates:    v.GetString("NOTION_CONTRACTOR_RATES_DB_ID"),
				ContractorPayouts:  v.GetString("NOTION_CONTRACTOR_PAYOUTS_DB_ID"),
				ContractorPayables: v.GetString("NOTION_CONTRACTOR_PAYABLES_DB_ID"),
				InvoiceSplit:       v.GetString("NOTION_INVOICE_SPLIT_DB_ID"),
				RefundRequest:      v.GetString("NOTION_REFUND_REQUEST_DB_ID"),
				DeploymentTracker:  v.GetString("NOTION_DEPLOYMENT_TRACKER_DB_ID"),
				Timesheet:          v.GetString("NOTION_TIMESHEET_DB_ID"),
				TaskOrderLog:       v.GetString("NOTION_TASK_ORDER_LOG_DB_ID"),
				BankAccounts:       v.GetString("NOTION_BANK_ACCOUNTS_DB_ID"),
			},
		},
		Discord: Discord{
			Webhooks: DiscordWebhook{
				Campfire:     v.GetString("DISCORD_WEBHOOK_CAMPFIRE"),
				AuditLog:     v.GetString("DISCORD_WEBHOOK_AUDIT"),
				ICYPublicLog: v.GetString("DISCORD_WEBHOOK_ICY_PUBLIC_LOG"),
			},
			SecretToken:   v.GetString("DISCORD_SECRET_TOKEN"),
			ApplicationID: v.GetString("DISCORD_APPLICATION_ID"),
			PublicKey:     v.GetString("DISCORD_PUBLIC_KEY"),
			IDs: DiscordID{
				DwarvesGuild:    v.GetString("DISCORD_DWARVES_GUILD_ID"),
				GolangChannel:   v.GetString("DISCORD_GOLANG_CHANNEL_ID"),
				ResearchChannel: v.GetString("DISCORD_RESEARCH_CHANNEL_ID"),
				OnLeaveChannel:  v.GetString("DISCORD_ONLEAVE_CHANNEL_ID"),
			},
		},
		Basecamp: Basecamp{
			BotKey:              v.GetString("BASECAMP_BOT_KEY"),
			ClientID:            v.GetString("BASECAMP_CLIENT_ID"),
			ClientSecret:        v.GetString("BASECAMP_CLIENT_SECRET"),
			OAuthRefreshToken:   v.GetString("BASECAMP_OAUTH_REFRESH_TOKEN"),
			AccountingProjectID: basecampAccountingProjectID,
			AccountingTodoSetID: basecampAccountingTodoSetID,
			PlaygroundProjectID: getIntWithDefault(v, "BASECAMP_PLAYGROUND_PROJECT_ID", defaultBasecampPlaygroundProjectID),
			PlaygroundTodoSetID: getIntWithDefault(v, "BASECAMP_PLAYGROUND_TODO_SET_ID", defaultBasecampPlaygroundTodoSetID),
		},
		Invoice: Invoice{
			TemplatePath:           v.GetString("INVOICE_TEMPLATE_PATH"),
			DirID:                  v.GetString("INVOICE_DIR_ID"),
			ContractorInvoiceDirID: v.GetString("CONTRACTOR_INVOICE_DIR_ID"),
			TestEmail:              v.GetString("INVOICE_TEST_EMAIL"),
		},

		Sendgrid: Sendgrid{
			APIKey: v.GetString("SENDGRID_API_KEY"),
		},
		Mochi: Mochi{
			BaseURL:         v.GetString("MOCHI_BASE_URL"),
			ApplicationID:   v.GetString("MOCHI_APPLICATION_ID"),
			ApplicationName: v.GetString("MOCHI_APPLICATION_NAME"),
			APIKey:          v.GetString("MOCHI_API_KEY"),
		},
		MochiPay: MochiPay{
			BaseURL: v.GetString("MOCHI_PAY_BASE_URL"),
		},
		MochiProfile: MochiProfile{
			BaseURL: v.GetString("MOCHI_PROFILE_BASE_URL"),
		},
		Tono: Tono{
			BaseURL: v.GetString("TONO_BASE_URL"),
		},
		ImprovMX: ImprovMX{
			Token: v.GetString("IMPROVMX_API_TOKEN"),
		},
		CommunityNft: CommunityNft{
			ContractAddress: v.GetString("COMMUNITY_NFT_CONTRACT_ADDRESS"),
		},
		Reddit: Reddit{
			ClientID:     v.GetString("REDDIT_CLIENT_ID"),
			ClientSecret: v.GetString("REDDIT_CLIENT_SECRET"),
			Username:     v.GetString("REDDIT_USERNAME"),
			Password:     v.GetString("REDDIT_PASSWORD"),
		},
		Dify: Dify{
			URL:   v.GetString("DIFY_URL"),
			Token: v.GetString("DIFY_TOKEN"),
		},
		CheckIn: CheckIn{
			WhitelistedEmployeeIDs: strings.Split(v.GetString("CHECKIN_WHITELISTED_EMPLOYEE_IDS"), ","),
		},
		InvoiceListener: parseInvoiceListenerConfig(v),
		Parquet: Parquet{
			LocalFilePath:   v.GetString("PARQUET_LOCAL_FILE_PATH"),
			SyncInterval:    v.GetString("PARQUET_SYNC_INTERVAL"),
			RemoteURL:       v.GetString("PARQUET_REMOTE_URL"),
			QuickTimeout:    v.GetString("PARQUET_QUICK_TIMEOUT"),
			ExtendedTimeout: v.GetString("PARQUET_EXTENDED_TIMEOUT"),
			EnableCaching:   v.GetBool("PARQUET_ENABLE_CACHING"),
		},
		TaskProvider: getStringWithDefault(v, "TASK_PROVIDER", "basecamp"),
		Noco: Noco{
			BaseURL:                       v.GetString("NOCO_BASE_URL"),
			Token:                         v.GetString("NOCO_TOKEN"),
			WorkspaceID:                   v.GetString("NOCO_WORKSPACE_ID"),
			BaseID:                        v.GetString("NOCO_BASE_ID"),
			InvoiceTableID:                v.GetString("NOCO_INVOICE_TABLE_ID"),
			InvoiceCommentsTableID:        v.GetString("NOCO_INVOICE_COMMENTS_TABLE_ID"),
			WebhookSecret:                 nocoWebhookSecret,
			AccountingTodosTableID:        accountingTodosTableID,
			AccountingTransactionsTableID: accountingTransactionsTableID,
		},
		AccountingIntegration: AccountingIntegration{
			Basecamp: AccountingBasecampIntegration{
				ProjectID: getIntWithDefault(v, "ACCOUNTING_BASECAMP_PROJECT_ID", basecampAccountingProjectID),
				TodoSetID: getIntWithDefault(v, "ACCOUNTING_BASECAMP_TODO_SET_ID", basecampAccountingTodoSetID),
				GroupIn:   getStringWithDefault(v, "ACCOUNTING_BASECAMP_GROUP_IN_NAME", "In"),
				GroupOut:  getStringWithDefault(v, "ACCOUNTING_BASECAMP_GROUP_OUT_NAME", "Out"),
			},
			Noco: AccountingNocoIntegration{
				TodosTableID:        accountingTodosTableID,
				TransactionsTableID: accountingTransactionsTableID,
				WebhookSecret:       nocoWebhookSecret,
			},
		},
		ExpenseIntegration: ExpenseIntegration{
			Noco: ExpenseNocoIntegration{
				WorkspaceID:   v.GetString("NOCO_EXPENSE_WORKSPACE_ID"),
				TableID:       v.GetString("NOCO_EXPENSE_TABLE_ID"),
				WebhookSecret: v.GetString("NOCO_EXPENSE_WEBHOOK_SECRET"),
			},
			Notion: ExpenseNotionIntegration{
				ExpenseDBID:    v.GetString("NOTION_EXPENSE_DB_ID"),
				ContractorDBID: notionContractorDBID,
				DataSourceID:   v.GetString("NOTION_EXPENSE_DATA_SOURCE_ID"),
			},
			ApproverMapping: parseKeyValuePairs(v.GetString("NOCO_EXPENSE_APPROVER_MAPPING")),
		},
		LeaveIntegration: LeaveIntegration{
			Noco: LeaveNocoIntegration{
				TableID:       v.GetString("NOCO_LEAVE_TABLE_ID"),
				WebhookSecret: v.GetString("NOCO_LEAVE_WEBHOOK_SECRET"),
			},
			Notion: LeaveNotionIntegration{
				LeaveDBID:         v.GetString("NOTION_LEAVE_DB_ID"),
				ContractorDBID:    notionContractorDBID,
				DataSourceID:      v.GetString("NOTION_LEAVE_DATA_SOURCE_ID"),
				VerificationToken: v.GetString("NOTION_VERIFICATION_TOKEN"),
			},
		},
	}
}

func DefaultConfigLoaders() []Loader {
	var loaders []Loader
	fileLoader := NewFileLoader(".env", ".")
	loaders = append(loaders, fileLoader)
	loaders = append(loaders, NewENVLoader())

	return loaders
}

// LoadConfig load config from loader list
func LoadConfig(loaders []Loader) *Config {
	v := viper.New()
	v.SetDefault("PORT", "8080")
	v.SetDefault("ENV", "local")
	v.SetDefault("ALLOWED_ORIGINS", "*")

	// Parquet sync service defaults
	v.SetDefault("PARQUET_LOCAL_FILE_PATH", "/tmp/vault.parquet")
	v.SetDefault("PARQUET_SYNC_INTERVAL", "1h")
	v.SetDefault("PARQUET_REMOTE_URL", "https://raw.githubusercontent.com/dwarvesf/memo.d.foundation/refs/heads/main/db/vault.parquet")
	v.SetDefault("PARQUET_QUICK_TIMEOUT", "2s")
	v.SetDefault("PARQUET_EXTENDED_TIMEOUT", "60s")
	v.SetDefault("PARQUET_ENABLE_CACHING", true)
	v.SetDefault("TASK_PROVIDER", "basecamp")

	for idx := range loaders {
		newV, err := loaders[idx].Load(*v)

		if err == nil {
			v = newV
		}
	}
	return Generate(v)
}

func LoadTestConfig() Config {
	return Config{
		Debug:    true,
		LogLevel: "debug",
		ApiServer: ApiServer{
			Port: "8080",
		},
		JWTSecretKey: "JWTSecretKey",
		Postgres: DBConnection{
			Host:    "127.0.0.1",
			Port:    "35432",
			User:    "postgres",
			Pass:    "postgres",
			Name:    "fortress_local_test",
			SSLMode: "disable",
		},
	}
}
