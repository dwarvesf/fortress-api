package notion

import (
	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

type IService interface {
	GetPage(pageID string) (clientPage nt.Page, err error)
	GetDatabase(databaseID string, filter *nt.DatabaseQueryFilter, sorts []nt.DatabaseQuerySort, pageSize int) (database *nt.DatabaseQueryResponse, err error)
	GetDatabaseWithStartCursor(databaseID string, startCursor string) (*nt.DatabaseQueryResponse, error)
	FindClientPageForChangelog(clientID string) (clientPage nt.Page, err error)
	GetProjectInDB(pageID string) (project *nt.DatabasePageProperties, err error)
	GetProjectsInDB(pageIDs []string, projectPageID string) (projects map[string]nt.DatabasePageProperties, err error)
	GetBlockChildren(pageID string) (blockChildrenResponse *nt.BlockChildrenResponse, err error)
	GetBlock(pageID string) (blockResponse nt.Block, err error)
	ListProjects() ([]model.NotionProject, error)
	ListProjectsWithChangelog() ([]model.ProjectChangelogPage, error)
	GetPagePropByID(pageID, propID string, query *nt.PaginationQuery) (*nt.PagePropResponse, error)

	// CreatePage create a page in notion
	CreatePage() error
	CreateDatabaseRecord(databaseID string, properties map[string]interface{}) (pageID string, err error)
	ToChangelogMJML(blocks []nt.Block, email model.Email) (string, error)

	QueryAudienceDatabase(audienceDBId, audience string) (records []nt.Page, err error)

	// GetProjectHeadDisplayNames fetches the display names for sales person, tech lead, account managers, and deal closing for a given Notion project pageID.
	GetProjectHeadDisplayNames(pageID string) (salePersonName, techLeadName, accountManagerNames, dealClosingEmails string, err error)
}
