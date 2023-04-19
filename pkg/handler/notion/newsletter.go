package notion

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/Boostport/mjml-go"
	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type From struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

type ProjectChangelog struct {
	ProjectPageID string `json:"project_page_id,omitempty"`
	IsPreview     bool   `json:"is_preview"`
	From          From   `json:"from,omitempty"`
}

// SendNewsLetter implements IHandler
func (h *handler) SendNewsLetter(c *gin.Context) {
	contentID := c.Param("id")
	isPreview := false
	if c.Query("preview") == "true" {
		isPreview = true
	}
	categories := []string{"newsletter", contentID}
	var emails []*model.Email

	m, err := h.generateEmailNewsletter(
		contentID,
		&mail.Email{
			Name:    "Dwarves Team",
			Address: "team@d.foundation",
		},
		categories,
	)
	if err != nil {
		h.logger.Error(err, "generateEmailNewsletter() failed")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if isPreview {
		emails = []*model.Email{
			{
				HTMLContent: m.HTMLContent,
				Subject:     m.Subject,
				From:        m.From,
				To: []*mail.Email{
					mail.NewEmail("Minh Luu", "leo@d.foundation"),
					mail.NewEmail("Huy Nguyen", "huy@d.foundation"),
					mail.NewEmail("Nikki", "nikki@d.foundation"),
					mail.NewEmail("Inno", "mytx@d.foundation"),
					mail.NewEmail("Vi", "tranthiaivi.cs@gmail.com"),
				},
				Categories: categories,
			},
		}
	} else {
		// get subscribers
		subscribers, _, err := h.getSubscribers(h.config.Notion.Databases.Audience, "Dwarves Updates")
		if err != nil {
			h.logger.Error(err, "getSubscribers() failed")
			c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
			return
		}

		// get audience list
		for _, s := range subscribers {
			if err != nil {
				h.logger.Error(err, "ToNewsletterHtml() failed with "+s.Address)
				c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
				return
			}
			emails = append(emails, &model.Email{
				HTMLContent: m.HTMLContent,
				Subject:     m.Subject,
				From:        m.From,
				To:          []*mail.Email{s},
				Categories:  categories,
			})
		}

		if h.config.Env != "prod" {
			if len(emails) > 1 {
				emails = emails[:1]
			}
		}
	}

	for _, email := range emails {
		err = h.service.Sendgrid.SendEmail(email)
		if err != nil {
			h.logger.Error(err, "send email failed: "+email.To[0].Address+" "+email.Subject)
			continue
		}
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) getSubscribers(pageID, audience string) ([]*mail.Email, []string, error) {
	records, err := h.service.Notion.QueryAudienceDatabase(pageID, audience)
	if err != nil {
		h.logger.Error(err, "query audience database")
		return nil, nil, err
	}

	mails := []*mail.Email{}
	subs := []string{}
	for i := range records {
		props := records[i].Properties.(nt.DatabasePageProperties)
		var name string
		if len(props["First Name"].RichText) > 0 {
			name = props["First Name"].RichText[0].PlainText
		}
		address := props["Email"].Email
		if address == nil {
			continue
		}
		if name == "" {
			continue
		}
		id := records[i].ID

		subs = append(subs, id)
		mails = append(mails, &mail.Email{
			Name:    name,
			Address: *address,
		})
	}
	return mails, subs, nil
}

func (h *handler) generateEmailNewsletter(id string, from *mail.Email, categories []string) (*model.Email, error) {
	m := model.Email{From: from, Categories: categories}
	var changelogBlocks []nt.Block
	title := "Dwarves Updates"

	page, err := h.service.Notion.GetBlock(id)
	if err != nil {
		h.logger.Error(err, "get block")
		return nil, err
	}
	switch v := page.(type) {
	case *nt.ChildPageBlock:
		title = v.Title
	}
	m.Subject = title

	pageContent, err := h.service.Notion.GetBlockChildren(id)
	if err != nil {
		h.logger.Errorf(err, "failed to download page", "pageID", id)
		return nil, err
	}
	changelogBlocks = pageContent.Results

	// Get children blocks of changelogBlocks
	for _, block := range changelogBlocks {
		if block.HasChildren() {
			children, err := h.service.Notion.GetBlockChildren(block.ID())
			if err != nil {
				h.logger.Errorf(err, "failed to get block children", "blockID", block.ID())
				return nil, err
			}

			switch v := block.(type) {
			case *nt.BulletedListItemBlock:
				v.Children = children.Results
			default:
				continue
			}
		}
	}

	// upload temp image from notion s3 to gcs
	for i, block := range changelogBlocks {
		switch v := block.(type) {
		case *nt.ImageBlock:
			// parse the url
			u, err := url.Parse(v.File.URL)
			if err != nil {
				return nil, err
			}

			// get the file extension
			extension := filepath.Ext(u.Path)

			fPath := fmt.Sprintf("https://storage.googleapis.com/%s/projects/newsletter-images/%s", h.config.Google.GCSBucketName, v.ID())
			gcsPath := fmt.Sprintf("projects/newsletter-images/%s", v.ID()+extension)

			response, err := http.Get(v.File.URL)
			if err != nil {
				return nil, err
			}

			defer response.Body.Close()

			if err := h.service.Google.UploadContentGCS(response.Body, gcsPath); err != nil {
				return nil, err
			}
			changelogBlocks[i].(*nt.ImageBlock).File.URL = fPath + extension
		}
	}

	content, err := h.service.Notion.ToChangelogMJML(changelogBlocks, m)
	if err != nil {
		h.logger.Error(err, "To Changelog MJML")
		return nil, err
	}

	m.HTMLContent, err = mjml.ToHTML(context.Background(), fmt.Sprintf(notion.MJMLDFUpdateTemplate, content))
	if err != nil {
		h.logger.Error(err, "To HTML")
		return nil, err
	}

	return &m, nil
}
