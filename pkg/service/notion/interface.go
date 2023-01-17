package notion

import (
	nt "github.com/dstotijn/go-notion"
)

type NotionService interface {
	FindClientPageForChangelog(clientId string) (clientPage nt.Page, err error)
	GetProjectInDB(pageID, projectPageID string) (project *nt.DatabasePageProperties, err error)
	GetProjectsInDB(pageIDs []string, projectPageID string) (projects map[string]nt.DatabasePageProperties, err error)
	GetBlockChildren(pageID string) (blockChildrenResponse *nt.BlockChildrenResponse, err error)
	GetDatabase(databaseID string, filter *nt.DatabaseQueryFilter) (database *nt.DatabaseQueryResponse, err error)
}
