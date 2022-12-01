package googlemail

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var (
	accountingUser = "accounting"
	teamEmail      = "team"
	spawnEmail     = "spawn"
)

type mailParseInfo struct {
	Sender           string
	TemplateFileName string
	Data             interface{}
	FuncMap          map[string]interface{}
}

func composeMailContent(m *mailParseInfo, templatePath string) (string, error) {
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

// email regex
var (
	emailRegex = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
)

// regex : validate regex
func regex(regex, sample string) error {
	c, err := regexp.Compile(regex)
	if err != nil {
		return err
	}
	if !c.MatchString(sample) {
		return errors.New("invalid input")
	}

	return nil
}

// Email validate
func email(email string) bool {
	return regex(emailRegex, email) == nil
}

func (g *googleService) getEmailThread(threadID, id string) (*googleMailThread, error) {
	thread := &googleMailThread{}
	req, err := buildGetEmailThreadRequest(threadID, id, g)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(thread); err != nil {
		return nil, err
	}
	return thread, err
}

func buildGetEmailThreadRequest(threadID, id string, g *googleService) (*http.Request, error) {
	url := getGMailURL(g.apiKey, id, "threads/"+threadID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", g.token.AccessToken))

	return req, nil
}

func getMessageIDFromThread(thread *googleMailThread) (msgID, references string, err error) {
	if len(thread.Messages) == 0 {
		return "", "", errors.New("empty message thread")
	}

	for _, v := range thread.Messages[len(thread.Messages)-1].Payload.Headers {
		if strings.ToLower(v.Name) == "message-id" {
			msgID = v.Value
		}
		if strings.ToLower(v.Name) == "references" {
			references = v.Value
		}
	}

	if msgID == "" {
		return "", "", errors.New("can't find message_id")
	}

	return msgID, fmt.Sprintf(`%s %s`, references, msgID), nil
}

func gatherAddresses(CCs []string) (string, error) {
	for _, v := range CCs {
		if v == "" {
			continue
		}
		if !email(v) {
			return "", errors.New(v + " is not an email")
		}
	}
	return strings.Join(CCs, ", "), nil
}

func buildEmailRequest(encodedEmail, id string, g *googleService) (*http.Request, error) {
	url := getGMailURL(g.apiKey, id, "messages/send")

	jsonContent, err := json.Marshal(map[string]interface{}{
		"raw": encodedEmail,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonContent))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", g.token.AccessToken))

	return req, nil
}

func getGMailURL(apiKey, id, action string) string {
	return fmt.Sprintf(
		"%v%v/%v?key=%v",
		gmailURL,
		id,
		action,
		apiKey,
	)
}
