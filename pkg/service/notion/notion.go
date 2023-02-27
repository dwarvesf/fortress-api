package notion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	nt "github.com/dstotijn/go-notion"
)

type notionService struct {
	notionClient *nt.Client
}

func New(secret string) IService {
	return &notionService{
		notionClient: nt.NewClient(secret),
	}
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

func (n *notionService) GetProjectInDB(pageID, projectPageID string) (*nt.DatabasePageProperties, error) {
	ctx := context.Background()

	// 1. get all project records in project page
	res, err := n.notionClient.QueryDatabase(ctx, projectPageID, &nt.DatabaseQuery{})
	if err != nil {
		return nil, err
	}

	// 2. loop through all projects to find the project by page id
	for _, r := range res.Results {
		if strings.ReplaceAll(r.ID, "-", "") == strings.ReplaceAll(pageID, "-", "") {
			p := r.Properties.(nt.DatabasePageProperties)
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
