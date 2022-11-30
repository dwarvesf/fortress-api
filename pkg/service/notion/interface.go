package notion

import (
	nt "github.com/dstotijn/go-notion"
)

type NotionService interface {
	FindClientPageForChangelog(clientID string) (clientPage nt.Page, err error)
	GetProjectInDB(pageID, projectPageID string) (project *nt.DatabasePageProperties, err error)
	GetProjectsInDB(pageIDs []string, projectPageID string) (projects map[string]nt.DatabasePageProperties, err error)
}
