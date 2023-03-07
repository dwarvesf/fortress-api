package googlemail

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"text/template"

	"github.com/dwarvesf/fortress-api/pkg/config"
)

var (
	accountingUser = "accounting"
	teamEmail      = "team"
	spawnEmail     = "spawn"
)

type MailParseInfo struct {
	Sender           string
	TemplateFileName string
	Data             interface{}
	FuncMap          map[string]interface{}
}

func composeMailContent(appConfig *config.Config, m *MailParseInfo) (string, error) {
	templatePath := appConfig.Invoice.TemplatePath
	if appConfig.Env == "local" || templatePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			pwd = os.Getenv("GOPATH") + "/src/github.com/dwarvesf/fortress-api"
		}
		templatePath = filepath.Join(pwd, "pkg/templates")
	}
	m.FuncMap["signatureName"] = func() string {
		switch m.Sender {
		case accountingUser:
			return "Eddie Ng"
		case teamEmail:
			return "Dwarves Foundation"
		case spawnEmail:
			return "Team Dwarves"
		default:
			return ""
		}
	}

	m.FuncMap["signatureTitle"] = func() string {
		switch m.Sender {
		case accountingUser:
			return "Accountant"
		case teamEmail:
			return "Team Dwarves"
		case spawnEmail:
			return "Hiring"
		default:
			return ""
		}
	}

	tmpl, err := template.New("mail").Funcs(m.FuncMap).ParseFiles(filepath.Join(templatePath, m.TemplateFileName), filepath.Join(templatePath, "signature.tpl"))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, m.TemplateFileName, m.Data); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buf.Bytes()), nil
}
