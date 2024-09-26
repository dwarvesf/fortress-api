package config

import (
	"strings"

	"github.com/spf13/viper"
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
	Google        Google
	Vault         Vault
	Notion        Notion
	Wise          Wise
	Discord       Discord
	Basecamp      Basecamp
	CurrencyLayer CurrencyLayer
	Mochi         Mochi
	MochiPay      MochiPay
	MochiProfile  MochiProfile
	Tono          Tono
	ImprovMX      ImprovMX
	CommunityNft  CommunityNft
	Reddit        Reddit
	Youtube       Youtube
	Dify          Dify

	Invoice  Invoice
	Sendgrid Sendgrid
	Github   Github
	CheckIn  CheckIn

	APIKey       string
	Debug        bool
	Env          string
	JWTSecretKey string
	FortressURL  string
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

type Wise struct {
	APIKey  string
	Profile string
	Url     string
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

type NotionDatabase struct {
	AuditCycle      string
	AuditActionItem string
	Earn            string
	TechRadar       string
	Audience        string
	Event           string
	Hiring          string
	StaffingDemand  string
	Project         string
	Delivery        string
	Digest          string
	Updates         string
	Memo            string
	Issue           string
}

type Discord struct {
	SecretToken string
	Webhooks    DiscordWebhook
	IDs         DiscordID
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
}

type Invoice struct {
	TemplatePath string
	DirID        string
	TestEmail    string
}

type Sendgrid struct {
	APIKey string
}

type Basecamp struct {
	BotKey            string
	ClientID          string
	ClientSecret      string
	OAuthRefreshToken string
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

func Generate(v ENV) *Config {
	return &Config{
		Debug:        v.GetBool("DEBUG"),
		APIKey:       v.GetString("API_KEY"),
		Env:          v.GetString("ENV"),
		JWTSecretKey: v.GetString("JWT_SECRET_KEY"),
		FortressURL:  v.GetString("FORTRESS_URL"),

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
		},
		Youtube: Youtube{
			ClientID:     v.GetString("YOUTUBE_API_CLIENT_ID"),
			ClientSecret: v.GetString("YOUTUBE_API_CLIENT_SECRET"),
			RefreshToken: v.GetString("YOUTUBE_REFRESH_TOKEN"),
		},

		Wise: Wise{
			APIKey:  v.GetString("WISE_API_KEY"),
			Profile: v.GetString("WISE_PROFILE"),
			Url:     v.GetString("WISE_URL"),
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
				AuditCycle:      v.GetString("NOTION_AUDIT_CYCLE_DB_ID"),
				AuditActionItem: v.GetString("NOTION_AUDIT_ACTION_ITEM_DB_ID"),
				Earn:            v.GetString("NOTION_EARN_DB_ID"),
				TechRadar:       v.GetString("NOTION_TECH_RADAR_DB_ID"),
				Audience:        v.GetString("NOTION_AUDIENCE_DB_ID"),
				Event:           v.GetString("NOTION_EVENT_DB_ID"),
				Hiring:          v.GetString("NOTION_HIRING_DB_ID"),
				StaffingDemand:  v.GetString("NOTION_STAFFING_DEMAND_DB_ID"),
				Project:         v.GetString("NOTION_PROJECT_DB_ID"),
				Delivery:        v.GetString("NOTION_DELIVERY_DB_ID"),
				Digest:          v.GetString("NOTION_DIGEST_DB_ID"),
				Updates:         v.GetString("NOTION_UPDATES_DB_ID"),
				Memo:            v.GetString("NOTION_MEMO_DB_ID"),
				Issue:           v.GetString("NOTION_ISSUE_DB_ID"),
			},
		},
		Discord: Discord{
			Webhooks: DiscordWebhook{
				Campfire:     v.GetString("DISCORD_WEBHOOK_CAMPFIRE"),
				AuditLog:     v.GetString("DISCORD_WEBHOOK_AUDIT"),
				ICYPublicLog: v.GetString("DISCORD_WEBHOOK_ICY_PUBLIC_LOG"),
			},
			SecretToken: v.GetString("DISCORD_SECRET_TOKEN"),
			IDs: DiscordID{
				DwarvesGuild:    v.GetString("DISCORD_DWARVES_GUILD_ID"),
				GolangChannel:   v.GetString("DISCORD_GOLANG_CHANNEL_ID"),
				ResearchChannel: v.GetString("DISCORD_RESEARCH_CHANNEL_ID"),
			},
		},
		Basecamp: Basecamp{
			BotKey:            v.GetString("BASECAMP_BOT_KEY"),
			ClientID:          v.GetString("BASECAMP_CLIENT_ID"),
			ClientSecret:      v.GetString("BASECAMP_CLIENT_SECRET"),
			OAuthRefreshToken: v.GetString("BASECAMP_OAUTH_REFRESH_TOKEN"),
		},
		Invoice: Invoice{
			TemplatePath: v.GetString("INVOICE_TEMPLATE_PATH"),
			DirID:        v.GetString("INVOICE_DIR_ID"),
			TestEmail:    v.GetString("INVOICE_TEST_EMAIL"),
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
		Debug: true,
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
