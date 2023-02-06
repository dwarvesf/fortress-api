package config

import (
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
	Google      Google
	Vault       Vault
	Notion      Notion
	NotionAudit NotionAudit
	Wise        Wise
	Discord     Discord

	APIKey       string
	Debug        bool
	Env          string
	JWTSecretKey string
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
	ClientSecret   string
	ClientID       string
	AppName        string
	GCSBucketName  string
	GCSProjectID   string
	GCSCredentials string
}

type Wise struct {
	APIKey  string
	Profile string
	Url     string
}

type Vault struct {
	Address string
	Token   string
	Path    string
}

type Notion struct {
	Secret             string
	EarnDBID           string
	TechRadarDBID      string
	AudienceDBID       string
	EventDBID          string
	HiringDBID         string
	StaffingDemandDBID string
	ProjectDBID        string
	DigestDBID         string
	UpdatesDBID        string
	MemoDBID           string
}

type NotionAudit struct {
	Secret              string
	AuditCycleDBID      string
	AuditActionItemDBID string
}

type Discord struct {
	Webhook string
}

type ENV interface {
	GetBool(string) bool
	GetString(string) string
}

func Generate(v ENV) *Config {
	return &Config{
		Debug:        v.GetBool("DEBUG"),
		APIKey:       v.GetString("API_KEY"),
		Env:          v.GetString("ENV"),
		JWTSecretKey: v.GetString("JWT_SECRET_KEY"),

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

		Google: Google{
			ClientSecret:   v.GetString("GOOGLE_API_CLIENT_SECRET"),
			ClientID:       v.GetString("GOOGLE_API_CLIENT_ID"),
			AppName:        v.GetString("GOOGLE_API_APP_NAME"),
			GCSBucketName:  v.GetString("GCS_BUCKET_NAME"),
			GCSProjectID:   v.GetString("GCS_PROJECT_ID"),
			GCSCredentials: v.GetString("GCS_CREDENTIALS"),
		},

		Wise: Wise{
			APIKey:  v.GetString("WISE_API_KEY"),
			Profile: v.GetString("WISE_PROFILE"),
			Url:     v.GetString("WISE_URL"),
		},

		Vault: Vault{
			Address: v.GetString("VAULT_ADDR"),
			Token:   v.GetString("VAULT_TOKEN"),
			Path:    v.GetString("VAULT_PATH"),
		},
		Notion: Notion{
			Secret:             v.GetString("NOTION_SECRET"),
			EarnDBID:           v.GetString("NOTION_EARN_DB_ID"),
			TechRadarDBID:      v.GetString("NOTION_TECH_RADAR_DB_ID"),
			AudienceDBID:       v.GetString("NOTION_AUDIENCE_DB_ID"),
			EventDBID:          v.GetString("NOTION_EVENT_DB_ID"),
			HiringDBID:         v.GetString("NOTION_HIRING_DB_ID"),
			StaffingDemandDBID: v.GetString("NOTION_STAFFING_DEMAND_DB_ID"),
			ProjectDBID:        v.GetString("NOTION_PROJECT_DB_ID"),
			DigestDBID:         v.GetString("NOTION_DIGEST_DB_ID"),
			UpdatesDBID:        v.GetString("NOTION_UPDATES_DB_ID"),
			MemoDBID:           v.GetString("NOTION_MEMO_DB_ID"),
		},
		NotionAudit: NotionAudit{
			Secret:              v.GetString("NOTION_AUDIT_SECRET"),
			AuditCycleDBID:      v.GetString("NOTION_AUDIT_CYCLE_DB_ID"),
			AuditActionItemDBID: v.GetString("NOTION_AUDIT_ACTION_ITEM_DB_ID"),
		},
		Discord: Discord{
			Webhook: v.GetString("DISCORD_WEBHOOK"),
		},
	}
}

func DefaultConfigLoaders() []Loader {
	loaders := []Loader{}
	fileLoader := NewFileLoader(".env", ".")
	loaders = append(loaders, fileLoader)
	loaders = append(loaders, NewENVLoader())

	return loaders
}

// LoadConfig load config from loader list
func LoadConfig(loaders []Loader) *Config {
	v := viper.New()
	v.SetDefault("PORT", "8080")
	v.SetDefault("GRPC_PORT", "8081")
	v.SetDefault("ENV", "local")
	v.SetDefault("FTM_RPC", "https://rpc.fantom.network")
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
