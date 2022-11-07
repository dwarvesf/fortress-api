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
	Postgres   DBConnection
	Clickhouse DBConnection

	// server
	ApiServer ApiServer

	Debug bool
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

type GrpcServer struct {
	Port string
	Host string
}

type ChainExplorerApiKey struct {
	Eth      string
	Ftm      string
	Bsc      string
	Optimism string
}
type RpcUrl struct {
	Eth      string
	Ftm      string
	Optimism string
}

type MarketplaceApiKey struct {
	Opensea  string
	Quixotic string
}

type Kafka struct {
	Servers       string
	IndexerTopic  string
	ConsumerGroup string
}

func generateConfigFromViper(v *viper.Viper) *Config {
	return &Config{
		Debug: v.GetBool("DEBUG"),

		ApiServer: ApiServer{
			Port:           v.GetString("PORT"),
			AllowedOrigins: v.GetString("ALLOWED_ORIGINS"),
		},

		Clickhouse: DBConnection{
			Host: v.GetString("CLICKHOUSE_HOST"),
			Port: v.GetString("CLICKHOUSE_PORT"),
			User: v.GetString("CLICKHOUSE_USER"),
			Name: v.GetString("CLICKHOUSE_NAME"),
			Pass: v.GetString("CLICKHOUSE_PASS"),
		},

		Postgres: DBConnection{
			Host:    v.GetString("DB_HOST"),
			Port:    v.GetString("DB_PORT"),
			User:    v.GetString("DB_USER"),
			Name:    v.GetString("DB_NAME"),
			Pass:    v.GetString("DB_PASS"),
			SSLMode: v.GetString("DB_SSL_MODE"),
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
	return generateConfigFromViper(v)
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
		Clickhouse: DBConnection{
			Host: "clickhouse_host",
			Port: "clickhouse_port",
			User: "clickhouse_user",
			Pass: "clickhouse_pass",
			Name: "clickhouse_name",
		},
	}
}
