package notion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	nt "github.com/dstotijn/go-notion"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/utils"
)

type notionService struct {
	notionClient *nt.Client
	projectsDBID string
	l            logger.Logger
}

func New(secret, projectID string, l logger.Logger) IService {
	return &notionService{
		notionClient: nt.NewClient(secret),
		projectsDBID: projectID,
		l:            l,
	}
}

// GetBlock implements IService
func (n *notionService) GetBlock(pageID string) (blockResponse nt.Block, err error) {
	ctx := context.Background()

	res, err := n.notionClient.FindBlockByID(ctx, pageID)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ToChangelogMJML implements Service
func (n *notionService) ToChangelogMJML(blocks []nt.Block, email model.Email) (string, error) {
	var resutl string
	for i, block := range blocks {
		switch v := block.(type) {
		case *nt.Heading1Block:
			// get array of plain text
			var plainText []string
			for _, text := range v.RichText {
				plainText = append(plainText, text.PlainText)
			}

			resutl = resutl + fmt.Sprintf(`<mj-text padding-bottom="0px" line-height="30px">
     	     <h1 style="font-weight: bold"> 
			 %s
					</h1>
        </mj-text>`, strings.Join(plainText, " "))
		case *nt.Heading2Block:
			// get array of plain text
			var plainText []string
			for _, text := range v.RichText {
				plainText = append(plainText, text.PlainText)
			}

			resutl = resutl + fmt.Sprintf(`<mj-text padding-bottom="0px" line-height="28px">
		  <h2 style="font-weight: bold"> 
		 %s
				</h2>
	</mj-text>`, strings.Join(plainText, " "))
		case *nt.Heading3Block:
			// get array of plain text
			var plainText []string
			for _, text := range v.RichText {
				plainText = append(plainText, text.PlainText)
			}

			resutl = resutl + fmt.Sprintf(`<mj-text padding-bottom="0px" font-size="16px" line-height="24px">
	  <h3 style="font-weight: bold"> 
	 %s
			</h3>
</mj-text>`, strings.Join(plainText, " "))
		case *nt.ParagraphBlock:
			// get array of plain text
			var plainText []string
			for _, text := range v.RichText {
				if text.HRef != nil {
					link := *text.HRef
					// check if text.href is a link
					if !utils.HasDomain(*text.HRef) {
						link = fmt.Sprintf("https://www.notion.so/dwarves/%s", *text.HRef)
					}

					text.PlainText = fmt.Sprintf(`<a href="%s">%s</a>`, link, text.PlainText)
				}
				plainText = append(plainText, text.PlainText)
			}

			resutl = resutl + fmt.Sprintf(`<mj-text padding-bottom="0px" padding-top="0px">
	  <p style="margin:4px 0px;"> 
	 %s
			</p>
</mj-text>`, strings.Join(plainText, " "))
		case *nt.BulletedListItemBlock:
			// get array of plain text
			var plainText []string
			for _, text := range v.RichText {
				plainText = append(plainText, text.PlainText)
			}

			// if first block
			if i == 0 {
				resutl = resutl + fmt.Sprintf(`<mj-text padding="0px 0px">
							<ul>
								  <li style="margin: 4px 0px;"> 
									%s
								  </li>`, strings.Join(plainText, " "))
				resutl = handleNestedBulletText(v, resutl)
			} else { // handle block in between first and last block
				// if block before this is a bullet list
				if _, ok := blocks[i-1].(*nt.BulletedListItemBlock); ok {
					if _, ok := blocks[i+1].(*nt.BulletedListItemBlock); ok { // and block after this is a bullet list
						resutl = resutl + fmt.Sprintf(`
								  <li style="margin: 4px 0px;"> 
									%s
								  </li>
						`, strings.Join(plainText, " "))
						resutl = handleNestedBulletText(v, resutl)
					} else { // and block after this is not a bullet list
						resutl = resutl + fmt.Sprintf(`
								  <li style="margin: 4px 0px;"> 
									%s
								  </li>
							</ul>
						</mj-text>`, strings.Join(plainText, " "))
						resutl = handleNestedBulletText(v, resutl)
					}
				} else { // if block before this is not a bullet list
					// if this is last block
					if i == len(blocks)-1 {
						resutl = resutl + fmt.Sprintf(`<mj-text padding="0px 0px">
							<ul>
								  <li style="margin: 4px 0px;"> 
									%s
								  </li>
							</ul>
						</mj-text>`, strings.Join(plainText, " "))
						resutl = handleNestedBulletText(v, resutl)
					} else { // if this is not last block
						resutl = resutl + fmt.Sprintf(`<mj-text padding="0px 0px">
							<ul>
								  <li style="margin: 4px 0px;"> 
									%s
								  </li>`, strings.Join(plainText, " "))
						resutl = handleNestedBulletText(v, resutl)
					}
				}
			}
		case *nt.ImageBlock:
			if v.External != nil {
				resutl = resutl + fmt.Sprintf(` <mj-image width="600px" padding-top="0" src="%s"></mj-image>`, v.External.URL)
			} else {
				resutl = resutl + fmt.Sprintf(` <mj-image width="600px" padding-top="0" src="%s"></mj-image>`, v.File.URL)
			}
		}
	}
	return resutl, nil
}

func handleNestedBulletText(v *nt.BulletedListItemBlock, resutl string) string {
	for _, child := range v.Children {
		switch child := child.(type) {
		case *nt.BulletedListItemBlock:
			var plainTextChild []string
			for _, textChild := range child.RichText {
				plainTextChild = append(plainTextChild, textChild.PlainText)
			}
			resutl = resutl + fmt.Sprintf(`<mj-text padding-bottom="0px" padding-top="0px" padding-left="45px">
			<ul>
			  <li> 
				%s
			  </li>
			</ul>
	</mj-text>`, strings.Join(plainTextChild, " "))
		}
	}
	return resutl
}

func (n *notionService) FindClientPageForChangelog(clientID string) (nt.Page, error) {
	ctx := context.Background()
	res, err := n.notionClient.FindPageByID(ctx, clientID)
	if err != nil {
		return nt.Page{}, err
	}

	return res, nil
}

func (n *notionService) GetDatabase(databaseID string, filter *nt.DatabaseQueryFilter, sorts []nt.DatabaseQuerySort, pageSize int) (*nt.DatabaseQueryResponse, error) {
	ctx := context.Background()

	q := &nt.DatabaseQuery{
		Filter: filter,
		Sorts:  sorts,
	}
	if pageSize > 0 {
		q.PageSize = pageSize
	}

	res, err := n.notionClient.QueryDatabase(ctx, databaseID, q)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (n *notionService) GetDatabaseWithStartCursor(databaseID string, startCursor string) (*nt.DatabaseQueryResponse, error) {
	ctx := context.Background()

	res, err := n.notionClient.QueryDatabase(ctx, databaseID, &nt.DatabaseQuery{
		StartCursor: startCursor,
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (n *notionService) GetBlockChildren(pageID string) (*nt.BlockChildrenResponse, error) {
	ctx := context.Background()

	res, err := n.notionClient.FindBlockChildrenByID(ctx, pageID, &nt.PaginationQuery{})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (n *notionService) GetPagePropByID(pageID, propID string, query *nt.PaginationQuery) (*nt.PagePropResponse, error) {
	ctx := context.Background()

	res, err := n.notionClient.FindPagePropertyByID(ctx, pageID, propID, query)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (n *notionService) GetProjectInDB(pageID string) (*nt.DatabasePageProperties, error) {
	ctx := context.Background()

	// 1. get all project records in project page
	res, err := n.notionClient.QueryDatabase(ctx, n.projectsDBID, &nt.DatabaseQuery{})
	if err != nil {
		return nil, err
	}

	// 2. loop through all projects to find the project by page id
	for _, r := range res.Results {
		if strings.ReplaceAll(r.ID, "-", "") == strings.ReplaceAll(pageID, "-", "") {
			p := r.Properties.(nt.DatabasePageProperties)
			if len(p["Project"].Title) != 0 && p["Changelog"].URL != nil {
				if *p["Changelog"].URL == "" {
					continue
				}
				clID := strings.Split(strings.Split(*p["Changelog"].URL, "/")[len(strings.Split(*p["Changelog"].URL, "/"))-1], "?")[0]
				cls, err := n.notionClient.QueryDatabase(ctx, clID, &nt.DatabaseQuery{
					Sorts: []nt.DatabaseQuerySort{
						{
							Property:  "Created",
							Direction: nt.SortDirDesc,
						},
					},
				})
				if err != nil {
					n.l.Errorf(err, "query project change log err", clID, p["Project"].Title[0].Text.Content)
					continue
				}

				if len(cls.Results) != 0 && len(cls.Results[0].Properties.(nt.DatabasePageProperties)["Title"].Title) != 0 {
					p["EmailSubject"] = nt.DatabasePageProperty{
						Title: []nt.RichText{
							{
								Type:      nt.RichTextTypeText,
								Text:      &nt.Text{Content: cls.Results[0].Properties.(nt.DatabasePageProperties)["Title"].Title[0].Text.Content},
								PlainText: cls.Results[0].Properties.(nt.DatabasePageProperties)["Title"].Title[0].Text.Content,
							},
						},
					}
				}
			}
			return &p, nil
		}
	}

	return nil, errors.New("page not found")
}

func (n *notionService) GetProjectsInDB(pageIDs []string, projectPageID string) (map[string]nt.DatabasePageProperties, error) {
	ctx := context.Background()

	// 1. get all project records in project page
	res, err := n.notionClient.QueryDatabase(ctx, projectPageID, &nt.DatabaseQuery{})
	if err != nil {
		return nil, err
	}

	// 2. loop through all projects to find the projects by page ids
	pages := map[string]nt.DatabasePageProperties{}
	for _, id := range pageIDs {
		pages[strings.ReplaceAll(id, "-", "")] = nt.DatabasePageProperties{}
	}
	for _, r := range res.Results {
		if _, ok := pages[strings.ReplaceAll(r.ID, "-", "")]; ok {
			pages[strings.ReplaceAll(r.ID, "-", "")] = r.Properties.(nt.DatabasePageProperties)
		}
	}

	return pages, nil
}

func (n *notionService) GetPage(pageID string) (nt.Page, error) {
	ctx := context.Background()

	res, err := n.notionClient.FindPageByID(ctx, pageID)
	if err != nil {
		return nt.Page{}, err
	}

	return res, nil
}

func (n *notionService) CreatePage() error {
	return nil
}

// create a record in notion database
func (n *notionService) CreateDatabaseRecord(databaseID string, properties map[string]interface{}) (string, error) {
	ctx := context.Background()

	props, err := convertMapToProperties(properties)
	if err != nil {
		return "", err
	}
	p, err := n.notionClient.CreatePage(ctx, nt.CreatePageParams{
		ParentType:             nt.ParentTypeDatabase,
		ParentID:               databaseID,
		DatabasePageProperties: &props,
	})
	if err != nil {
		return "", err
	}

	return p.ID, nil
}

func convertMapToProperties(properties map[string]interface{}) (nt.DatabasePageProperties, error) {
	props := nt.DatabasePageProperties{}

	for key, value := range properties {
		switch key {
		case "Name":
			props["Name"] = nt.DatabasePageProperty{
				Type:  nt.DBPropTypeTitle,
				Title: []nt.RichText{{Text: &nt.Text{Content: value.(string)}}},
			}
		case "Status":
			props["Status"] = nt.DatabasePageProperty{
				Type: nt.DBPropTypeSelect,
				Select: &nt.SelectOptions{
					Name: value.(string),
				},
			}
		// case "Assign":
		// 	props["Assign"] = nt.DatabasePageProperty{
		// 		Type: nt.DatabasePropertyType(nt.DBPropTypePeople),
		// 		People: []nt.User{
		// 			{
		// 				BaseUser: nt.BaseUser{
		// 					ID: value.(string),
		// 				},
		// 			},
		// 		},
		// 	}

		default:
			return nil, fmt.Errorf("unsupported property: %s", key)
		}
	}

	return props, nil
}

func (n *notionService) ListProject() ([]model.ProjectChangelogPage, error) {
	ctx := context.Background()
	prjs, err := n.notionClient.QueryDatabase(ctx, n.projectsDBID, &nt.DatabaseQuery{
		Filter: &nt.DatabaseQueryFilter{
			Property: "Status",
			DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
				Select: &nt.SelectDatabaseQueryFilter{
					Equals: "Active",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	res := []model.ProjectChangelogPage{}
	for _, r := range prjs.Results {
		p := r.Properties.(nt.DatabasePageProperties)
		if len(p["Project"].Title) != 0 && p["Changelog"].URL != nil {
			if *p["Changelog"].URL == "" {
				continue
			}
			clID := strings.Split(strings.Split(*p["Changelog"].URL, "/")[len(strings.Split(*p["Changelog"].URL, "/"))-1], "?")[0]
			cls, err := n.notionClient.QueryDatabase(ctx, clID, &nt.DatabaseQuery{
				Sorts: []nt.DatabaseQuerySort{
					{
						Property:  "Created",
						Direction: nt.SortDirDesc,
					},
				},
			})
			if err != nil {
				n.l.Errorf(err, "query project change log err", clID, p["Project"].Title[0].Text.Content)
				continue
			}
			if len(cls.Results) != 0 && len(cls.Results[0].Properties.(nt.DatabasePageProperties)["Title"].Title) != 0 {
				res = append(res, model.ProjectChangelogPage{
					RowID:        r.ID,
					Name:         p["Project"].Title[0].Text.Content,
					Title:        cls.Results[0].Properties.(nt.DatabasePageProperties)["Title"].Title[0].Text.Content,
					ChangelogURL: cls.Results[0].URL,
				})
			}
		}
	}
	return res, nil
}

func (n *notionService) QueryAudienceDatabase(audienceDBId, audience string) (records []nt.Page, err error) {
	ctx := context.Background()
	var t bool = true
	var cursor string = ""

	var filter *nt.DatabaseQueryFilter
	switch audience {
	case "Developers Only":
		filter = &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Or: []nt.DatabaseQueryFilter{
						{Property: "Personas", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{Contains: "Developer"}}},
						{Property: "Personas", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{Contains: "Engineer"}}},
						{Property: "Personas", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{Contains: "Tester"}}},
						{Property: "Personas", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{Contains: "Product Manager"}}},
					},
				},
				{
					And: []nt.DatabaseQueryFilter{
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Community"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Employee"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "CLient"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Past Client"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Fellowship"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Prospect"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Failed CV"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Failed Test"}}},
						{Property: "Tags", DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{MultiSelect: &nt.MultiSelectDatabaseQueryFilter{DoesNotContain: "Failed Interview"}}},
					},
				},
			},
		}
	case "Partner Updates", "Dwarves Updates":
		filter = &nt.DatabaseQueryFilter{
			And: []nt.DatabaseQueryFilter{
				{
					Property: audience,
					DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
						Checkbox: &nt.CheckboxDatabaseQueryFilter{
							Equals: &t,
						},
					},
				},
			},
		}

	default:
		return nil, errors.New("audience not found")
	}

	// filter out unsubscribed
	var unsubscribed bool = false
	filter.And = append(filter.And, nt.DatabaseQueryFilter{
		Property: "Unsubscribed",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			Checkbox: &nt.CheckboxDatabaseQueryFilter{
				Equals: &unsubscribed,
			},
		},
	})

	n.l.Info("start querying audience database")
	for {
		res, err := n.notionClient.QueryDatabase(ctx, audienceDBId, &nt.DatabaseQuery{
			Filter: filter,
			Sorts: []nt.DatabaseQuerySort{
				{
					Property:  "Created Time",
					Direction: nt.SortDirAsc,
				},
			},
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return nil, err
		}
		records = append(records, res.Results...)
		if !res.HasMore {
			break
		}
		cursor = *res.NextCursor
	}
	n.l.Info("finish querying audience database")
	return records, nil
}
