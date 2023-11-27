package notion

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/Boostport/mjml-go"
	nt "github.com/dstotijn/go-notion"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/view"
	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	projectColumn             = "Project"
	clientsColumn             = "Client"
	groupEmailColumn          = "Group Email"
	changelogRecipientsColumn = "Changelog Recipients"
	leadColumn                = "Lead"
	changelogColumn           = "Changelog"
	portalColumn              = "Project Portal"
	emailSubject              = "EmailSubject"
)

type singleChangelogError struct {
	ProjectName string
	Err         error
}

func (e singleChangelogError) Error() string {
	return fmt.Sprintf("%s: %s", e.ProjectName, e.Err.Error())
}

func parseProjectChangelogNotionMessageFromCtx(c *gin.Context) (ProjectChangelog, error) {
	msg := ProjectChangelog{}
	err := c.ShouldBindJSON(&msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

// GetAvailableProjectsChangelog godoc
// @Summary get available projects changelog
// @Description get available projects changelog
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/changelogs/projects/available [get]
func (h *handler) GetAvailableProjectsChangelog(c *gin.Context) {
	projects, err := h.service.Notion.ListProject()
	if err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}
	c.JSON(http.StatusOK, view.CreateResponse[any](projects, nil, nil, nil, ""))
}

// SendProjectChangelog godoc
// @Summary send project changelog
// @Description send project changelog
// @Tags Notion
// @Accept  json
// @Produce  json
// @Success 200 {object} view.MessageResponse
// @Failure 400 {object} view.ErrorResponse
// @Router /notion/changelogs/project [post]
func (h *handler) SendProjectChangelog(c *gin.Context) {
	msg, err := parseProjectChangelogNotionMessageFromCtx(c)
	if err != nil {
		h.logger.Error(err, "failed to parse project changelog message")
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	if err = h.sendProjectChangelog(msg); err != nil {
		c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
		return
	}

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "ok"))
}

func (h *handler) sendProjectChangelog(changelog ProjectChangelog) error {
	key := "project changelog"
	if changelog.IsPreview {
		key = "project changelog preview"
	}
	categories := []string{key, changelog.ProjectPageID}
	values, err := h.service.Notion.GetProjectInDB(changelog.ProjectPageID)
	if err != nil {
		h.logger.Errorf(err, "failed to get project in db", "project", changelog.ProjectPageID)
		return err
	}
	return h.sendSingleProjectChangelog(changelog.ProjectPageID, *values, &mail.Email{
		Name:    changelog.From.Name,
		Address: changelog.From.Email},
		categories,
		changelog.IsPreview,
	)
}

func (h *handler) sendSingleProjectChangelog(
	id string,
	values nt.DatabasePageProperties,
	from *mail.Email,
	categories []string,
	isPreview bool,
) error {
	m, _, err := h.generateEmailChangelog(id, values, from, categories, isPreview)
	if err != nil {
		return err
	}

	if err := h.service.Sendgrid.SendEmail(m); err != nil {
		return err
	}

	h.logger.Info(fmt.Sprintf("Send %s successfully", m.Subject))
	return nil
}

func (h *handler) generateEmailChangelog(
	id string,
	values nt.DatabasePageProperties,
	from *mail.Email,
	categories []string,
	isPreview bool,
) (*model.Email, *model.ProjectChangelogPage, error) {
	m := model.Email{From: from, Categories: categories}
	var changelogBlocks []nt.Block
	projectName := ""
	archiveURL := ""
	p := model.ProjectChangelogPage{RowID: id}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			h.logger.Error(fmt.Errorf("%v", r), "Recoverd "+projectName)
		}
	}()

	// process
	for i, v := range values {
		switch i {
		case projectColumn:
			if len(v.Title) < 1 || v.Title[0].Text == nil {
				continue
			}
			projectName = v.Title[0].Text.Content
			p.Name = projectName
		case portalColumn:
			if v.URL == nil {
				continue
			}
			archiveURL = *v.URL
			if archiveURL != "" {
				if !strings.HasPrefix(archiveURL, "http") || !strings.HasPrefix(archiveURL, "https") {
					archiveURL = "https://" + archiveURL
				}
			}
		case changelogRecipientsColumn:
			if len(v.Relation) < 1 {
				continue
			}
			for _, c := range v.Relation {
				clientPage, err := h.service.Notion.FindClientPageForChangelog(c.ID)
				if err != nil {
					h.logger.Errorf(err, "failed to find client page for changelog", "clientPage", c.ID)
					return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
				}
				props := clientPage.Properties.(nt.DatabasePageProperties)
				var name string
				if len(props["First Name"].RichText) > 0 {
					name = props["First Name"].RichText[0].Text.Content
				}
				address := props["Email"].Email
				if name == "" || *address == "" {
					continue
				}
				m.To = append(m.To, &mail.Email{Name: name, Address: *address})
			}
		case groupEmailColumn:
			if len(v.RichText) < 1 || v.RichText[0].Text == nil {
				continue
			}
			groupEmails := v.RichText[0].Text.Content
			gms := strings.Split(groupEmails, ",")
			for _, groupEmail := range gms {
				bccMail := strings.TrimSpace(groupEmail)
				m.Bcc = append(m.Bcc, &mail.Email{Name: bccMail, Address: bccMail})
			}
		case leadColumn:
			if len(v.Relation) < 1 {
				continue
			}
			for _, c := range v.Relation {
				recipientsPage, err := h.service.Notion.FindClientPageForChangelog(c.ID)
				if err != nil {
					h.logger.Errorf(err, "failed to find page for changelogs", "clientPage", c.ID)
					return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
				}
				props := recipientsPage.Properties.(nt.DatabasePageProperties)
				var name string
				if len(props["Full Name"].Title) > 0 {
					name = props["Full Name"].Title[0].Text.Content
				}
				address := props["Team Email"].Email
				if name == "" || *address == "" {
					continue
				}
				m.Bcc = append(m.Bcc, &mail.Email{Name: name, Address: *address})
			}
		case changelogColumn:
			if v.URL == nil {
				continue
			}

			changelogsID := ""
			changelogsURL := *v.URL
			fields := strings.Split(changelogsURL, "/")
			changelogsID = strings.Split(fields[len(fields)-1], "?")[0]

			// timeFilter is one month ago from now
			timeFilter := time.Now().AddDate(0, -1, 0)

			resp, err := h.service.Notion.GetDatabase(changelogsID, &nt.DatabaseQueryFilter{
				And: []nt.DatabaseQueryFilter{
					{
						Property: "Created",
						DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
							Date: &nt.DatePropertyFilter{
								OnOrAfter: &timeFilter,
							},
						},
					},
				},
			}, nil, 0)
			if err != nil {
				h.logger.Error(err, "download page")
				return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
			}
			var pages = resp.Results

			// get latest changelog
			sort.Slice(pages, func(i, j int) bool {
				propsI := pages[i].Properties.(nt.DatabasePageProperties)
				propsJ := pages[j].Properties.(nt.DatabasePageProperties)
				return propsI["Created"].Date.Start.After(propsJ["Created"].Date.Start.Time)
			})

			var latestChangelogPage = pages[0]

			pageContent, err := h.service.Notion.GetBlockChildren(latestChangelogPage.ID)
			if err != nil {
				h.logger.Errorf(err, "failed to download page", "pageID", latestChangelogPage.ID)
				return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
			}
			changelogBlocks = pageContent.Results
		}
	}

	// pageTitle := changelogPage.Root().Title
	if isPreview {
		// m.Bcc = []*mail.Email{mail.NewEmail("Minh Luu", "leo@dwarvesv.com")}
		m.Categories = []string{}
		m.To = []*mail.Email{mail.NewEmail("Minh Luu", "leo@d.foundation")}
	}
	m.Subject = values[emailSubject].Title[0].Text.Content

	// Get children blocks of changelogBlocks
	if err := h.getChildrenBlocks(changelogBlocks); err != nil {
		h.logger.Error(err, "failed to get block children")
		return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
	}

	// upload temp image from notion s3 to gcs
	for i, block := range changelogBlocks {
		switch v := block.(type) {
		case *nt.ImageBlock:
			var isExternalFile = v.External != nil
			var imgURL string

			if isExternalFile {
				imgURL = v.External.URL
			} else {
				imgURL = v.File.URL
			}

			// parse the url
			u, err := url.Parse(imgURL)
			if err != nil {
				return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
			}

			// get the file extension
			extension := filepath.Ext(u.Path)

			fPath := fmt.Sprintf("https://storage.googleapis.com/%s/projects/change-logs-images/%s", h.config.Google.GCSBucketName, v.ID())
			gcsPath := fmt.Sprintf("projects/change-logs-images/%s", v.ID()+extension)

			response, err := http.Get(imgURL)
			if err != nil {
				return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
			}

			defer response.Body.Close()

			if err := h.service.Google.UploadContentGCS(response.Body, gcsPath); err != nil {
				return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
			}
			if isExternalFile {
				changelogBlocks[i].(*nt.ImageBlock).External.URL = fPath + extension
			} else {
				changelogBlocks[i].(*nt.ImageBlock).File.URL = fPath + extension
			}
		}
	}

	content, err := h.service.Notion.ToChangelogMJML(changelogBlocks, m)
	if err != nil {
		h.logger.Error(err, "To Changelog MJML")
		return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
	}

	mjmlContent := fmt.Sprintf(notion.MJMLChangelogTemplate, content, archiveURL, archiveURL)

	m.HTMLContent, err = mjml.ToHTML(context.Background(), mjmlContent)
	if err != nil {
		h.logger.Error(err, "To HTML")
		return nil, nil, singleChangelogError{ProjectName: projectName, Err: err}
	}

	return &m, &p, nil
}

func (h *handler) getChildrenBlocks(blocks []nt.Block) error {
	for _, block := range blocks {
		if block.HasChildren() {
			switch v := block.(type) {
			case *nt.BulletedListItemBlock:
				children, err := h.service.Notion.GetBlockChildren(block.ID())
				if err != nil {
					return err
				}

				v.Children = children.Results
				if err := h.getChildrenBlocks(v.Children); err != nil {
					return err
				}
			default:
				continue
			}
		}
	}

	return nil
}
