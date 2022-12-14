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
	Google Google
	Vault  Vault
	Notion Notion
	Wise   Wise

	APIKey string
	Debug  bool
	Env    string
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
	Secret string
}

type ENV interface {
	GetBool(string) bool
	GetString(string) string
}

func Generate(v ENV) *Config {
	return &Config{
		Debug:  v.GetBool("DEBUG"),
		APIKey: v.GetString("API_KEY"),
		Env:    v.GetString("ENV"),

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
			Secret: v.GetString("NOTION_SECRET"),
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
