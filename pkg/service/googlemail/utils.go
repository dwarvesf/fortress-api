package googlemail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
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

// composeTaskOrderConfirmationContent generates email content from template
func composeTaskOrderConfirmationContent(appConfig *config.Config, data *model.TaskOrderConfirmationEmail) (string, error) {
	templatePath := appConfig.Invoice.TemplatePath
	if appConfig.Env == "local" || templatePath == "" {
		pwd, err := os.Getwd()
		if err != nil {
			pwd = os.Getenv("GOPATH") + "/src/github.com/dwarvesf/fortress-api"
		}
		templatePath = filepath.Join(pwd, "pkg/templates")
	}

	t, err := template.New("taskOrderConfirmation.tpl").Funcs(template.FuncMap{
		"formattedMonth": func() string {
			t, err := time.Parse("2006-01", data.Month)
			if err != nil {
				return data.Month
			}
			return t.Format("January 2006")
		},
		"contractorLastName": func() string {
			parts := strings.Fields(data.ContractorName)
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
			return data.ContractorName
		},
		"periodEndDay": func() string {
			t, err := time.Parse("2006-01", data.Month)
			if err != nil {
				return "31"
			}
			lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC)
			return fmt.Sprintf("%02d", lastDay.Day())
		},
		"monthName": func() string {
			t, err := time.Parse("2006-01", data.Month)
			if err != nil {
				return ""
			}
			return t.Format("January")
		},
		"year": func() string {
			parts := strings.Split(data.Month, "-")
			if len(parts) >= 1 {
				return parts[0]
			}
			return ""
		},
	}).ParseFiles(filepath.Join(templatePath, "taskOrderConfirmation.tpl"))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
