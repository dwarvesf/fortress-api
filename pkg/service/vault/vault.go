package vault

import (
	"fmt"
	"strconv"

	"github.com/dwarvesf/fortress-api/pkg/config"
	vault "github.com/hashicorp/vault/api"
)

type Vault struct {
	data map[string]interface{}
}

func New(cfg *config.Config) (VaultService, error) {
	config := vault.DefaultConfig()
	config.Address = cfg.Vault.Address

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	client.SetToken(cfg.Vault.Token)

	secret, err := client.Logical().Read(cfg.Vault.Path)
	if err != nil {
		return nil, fmt.Errorf("unable to read secret: %v", err)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to read secret data")
	}

	return &Vault{
		data: data,
	}, nil
}

func (v *Vault) LoadConfig() *config.Config {
	return &config.Config{
		Debug:  v.GetBool("DEBUG"),
		APIKey: v.GetString("API_KEY"),
		Env:    v.GetString("ENV"),

		ApiServer: config.ApiServer{
			Port:           v.GetString("PORT"),
			AllowedOrigins: v.GetString("ALLOWED_ORIGINS"),
		},

		Postgres: config.DBConnection{
			Host:    v.GetString("DB_HOST"),
			Port:    v.GetString("DB_PORT"),
			User:    v.GetString("DB_USER"),
			Name:    v.GetString("DB_NAME"),
			Pass:    v.GetString("DB_PASS"),
			SSLMode: v.GetString("DB_SSL_MODE"),
		},

		Google: config.Google{
			ClientSecret:   v.GetString("GOOGLE_API_CLIENT_SECRET"),
			ClientID:       v.GetString("GOOGLE_API_CLIENT_ID"),
			AppName:        v.GetString("GOOGLE_API_APP_NAME"),
			GCSBucketName:  v.GetString("GCS_BUCKET_NAME"),
			GCSProjectID:   v.GetString("GCS_PROJECT_ID"),
			GCSCredentials: v.GetString("GCS_CREDENTIALS"),

			MailApiKey:           v.GetString("GOOGLE_MAIL_API_KEY"),
			TeamEmailToken:       v.GetString("GOOGLE_TEAM_EMAIL_TOKEN"),
			TeamEmailID:          v.GetString("GOOGLE_TEAM_EMAIL_ID"),
			AccountingEmailToken: v.GetString("GOOGLE_ACCOUNTING_EMAIL_TOKEN"),
			AccountingEmailID:    v.GetString("GOOGLE_ACCOUNTING_EMAIL_ID"),
			TemplatePath:         v.GetString("GOOGLE_MAIL_TEMPLATE_PATH"),
		},

		Notion: config.Notion{
			Secret: v.GetString("NOTION_SECRET"),
		},

		Wise: config.Wise{
			ApiKey:  v.GetString("WISE_APIKEY"),
			Profile: v.GetString("WISE_PROFILE"),
		},

		Basecamp: config.Basecamp{
			ClientID:     v.GetString("BASECAMP_CLIENT_ID"),
			ClientSecret: v.GetString("BASECAMP_CLIENT_SECRET"),
			RefreshToken: v.GetString("BASECAMP_REFRESH_TOKEN"),
			BotKey:       v.GetString("BASECAMP_BOT_KEY"),
		},
	}
}

func (v *Vault) GetString(key string) string {
	value, _ := v.data[key].(string)
	return value
}

func (v *Vault) GetBool(key string) bool {
	data, _ := v.data[key].(string)
	value, _ := strconv.ParseBool(data)
	return value
}
