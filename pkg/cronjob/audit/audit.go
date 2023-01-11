package audit

import (
	"reflect"

	"github.com/dstotijn/go-notion"
	"github.com/k0kubun/pp/v3"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/cronjob/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

type cronjob struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) ICronjob {
	return &cronjob{store: store, repo: repo, service: service, logger: logger, config: cfg}
}

const (
	AuditCycleBDID  = "6de879f5-1554-4d25-beaa-b5b44b46f629"
	AuditActionBDID = "2baf09fc-a67b-4f3f-aa90-ea1eb6392109"
	VoconicAudit1   = "0fde875d-67b1-4c23-b524-4b541b301c62"
	NghenhanAudit1  = "23d6b63b-3a2b-49fc-8af7-5ffca1379a29"
)

func (c *cronjob) SyncAuditCycle() {
	// Get all audit cycle from notion
	database, err := c.service.Notion.GetDatabase(AuditCycleBDID)
	if err != nil {
		c.logger.Error(err, "failed to get database from notion")
		return
	}

	// Map notion page
	mapPage := NotionDatabaseToMap(database)

	// Get all audit cycle from database
	auditCycles, err := c.store.AuditCycle.All(c.repo.DB())
	if err != nil {
		c.logger.Error(err, "failed to get audit cycle from database")
		return
	}

	auditCycleMap := model.AuditCycleToMap(auditCycles)

	// Compare and update
	// Create audit cycle if not exist in auditCycleMap
	for _, page := range mapPage {
		// If auditCycle not exist in mapPage
		if _, ok := auditCycleMap[model.MustGetUUIDFromString(page.ID)]; !ok {
			pp.Println("Create audit cycle ID:", page.ID)
			auditCycle := model.NewAuditCycleFromNotionPage(&page)

			// Create audits
			if err := c.createAudit(&page, auditCycle); err != nil {
				c.logger.Error(err, "failed to create audit")
				return
			}

			// Create audit cycle
			auditCycle, err = c.store.AuditCycle.Create(c.repo.DB(), auditCycle)
		}
	}

	return
}

func (c *cronjob) createAudit(page *notion.Page, auditCycle *model.AuditCycle) error {
	audit, err := c.service.Notion.GetBlockChildren(page.ID)
	if err != nil {
		c.logger.Error(err, "failed to get audit from notion")
		return err
	}

	if page.ID == VoconicAudit1 || page.ID == NghenhanAudit1 {
		return nil
	}
	// Find the audit checklist block index
	auditChecklistIndex := -1
	for index, block := range audit.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			auditChecklistIndex = index
		}
	}

	properties := page.Properties.(notion.DatabasePageProperties)
	auditChecklist, err := c.service.Notion.GetDatabase(audit.Results[auditChecklistIndex].ID())
	if err != nil {
		c.logger.Error(err, "failed to get database from notion")
		return err
	}

	// Create audit record for each row
	for _, row := range auditChecklist.Results {
		pp.Println("Create audit ID:", row.ID)

		// get auditor from notion id
		checklistProperties := row.Properties.(notion.DatabasePageProperties)
		auditor, err := c.store.Employee.OneByNotionID(c.repo.DB(), checklistProperties["Auditor"].People[0].ID)
		if err != nil {
			c.logger.Error(err, "failed to get auditor from notion id")
			return err
		}

		// Create new audit object
		flag, err := c.getFlag(page, &row)
		if err != nil {
			c.logger.Error(err, "failed to get flag")
			return errs.ErrFailedToGetFlag
		}

		newAudit := model.NewAuditFromNotionPage(row, properties["Project"].Relation[0].ID, auditor.ID, flag, audit.Results[auditChecklistIndex].ID())
		newAudit, err = c.store.Audit.Create(c.repo.DB(), newAudit)

		// Create new audit participant
		if len(checklistProperties["Participants"].People) > 0 {
			for _, p := range checklistProperties["Participants"].People {
				participant, err := c.store.Employee.OneByNotionID(c.repo.DB(), p.ID)
				if err != nil {
					c.logger.Error(err, "failed to get participant by notion id")
					return err
				}

				_, err = c.store.AuditParticipant.Create(c.repo.DB(), &model.AuditParticipant{
					AuditID:    newAudit.ID,
					EmployeeID: participant.ID,
				})

				if err != nil {
					c.logger.Error(err, "failed to create audit participant")
					return err
				}
			}
		}

		// Update audit cycle
		c.updateAuditIDForAuditCycle(newAudit, auditCycle)

		// Create Audit Items
		if err := c.createAuditItem(page, &row); err != nil {
			c.logger.Error(err, "failed to create audit item")
			return err
		}
	}
	return nil
}

func (c *cronjob) updateAuditIDForAuditCycle(audit *model.Audit, auditCycle *model.AuditCycle) {
	switch audit.Type {
	case model.AuditTypeHealth:
		auditCycle.HealthAuditID = audit.ID
	case model.AuditTypeProcess:
		auditCycle.ProcessAuditID = audit.ID
	case model.AuditTypeBackend:
		auditCycle.BackendAuditID = audit.ID
	case model.AuditTypeFrontend:
		auditCycle.FrontendAuditID = audit.ID
	case model.AuditTypeSystem:
		auditCycle.SystemAuditID = audit.ID
	case model.AuditTypeMobile:
		auditCycle.MobileAuditID = audit.ID
	case model.AuditTypeBlockchain:
		auditCycle.BlockchainAuditID = audit.ID
	}

	if audit.Score != 0 {
		auditCycle.Status = model.AuditStatusAudited
	}

}

func (c *cronjob) getFlag(page *notion.Page, row *notion.Page) (model.AuditFlag, error) {
	countPoor, countAcceptable := 0, 0
	checklistPage, err := c.service.Notion.GetBlockChildren(row.ID)
	if err != nil {
		c.logger.Error(err, "failed to get audit from notion")
		return "", err
	}

	// Find the audit checklist block index
	checklistDatabaseIndex := -1
	for index, block := range checklistPage.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			checklistDatabaseIndex = index
		}
	}

	checklistDatabase, err := c.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID())
	if err != nil {
		c.logger.Error(err, "failed to get database from notion")
		return "", err
	}

	// Create audit item record for each row
	for _, checklist := range checklistDatabase.Results {
		// Create new audit object
		properties := checklist.Properties.(notion.DatabasePageProperties)
		if properties["Grade"].Select != nil {
			switch properties["Grade"].Select.Name {
			case "Poor":
				countPoor++
			case "Acceptable":
				countAcceptable++
			}
		}
	}

	if countPoor >= 3 {
		return model.AuditFlagRed, nil
	} else if countAcceptable >= 3 {
		return model.AuditFlagYellow, nil
	}

	return model.AuditFlagGreen, nil
}

func (c *cronjob) createAuditItem(page *notion.Page, row *notion.Page) error {
	pp.Println("Create audit item", row.ID)
	checklistPage, err := c.service.Notion.GetBlockChildren(row.ID)
	if err != nil {
		c.logger.Error(err, "failed to get audit from notion")
		return err
	}

	// Find the audit checklist block index
	checklistDatabaseIndex := -1
	for index, block := range checklistPage.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			checklistDatabaseIndex = index
		}
	}

	checklistDatabase, err := c.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID())
	if err != nil {
		c.logger.Error(err, "failed to get database from notion")
		return err
	}

	// Create audit item record for each row
	for _, checklist := range checklistDatabase.Results {
		// Create new audit object
		newAuditItem := model.NewAuditItemFromNotionPage(checklist, row.ID, checklistPage.Results[checklistDatabaseIndex].ID())
		newAuditItem, err = c.store.AuditItem.Create(c.repo.DB(), newAuditItem)
	}

	return nil
}

func NotionDatabaseToMap(database *notion.DatabaseQueryResponse) map[string]notion.Page {
	result := make(map[string]notion.Page)
	for _, r := range database.Results {
		result[r.ID] = r
	}

	return result
}

func (c *cronjob) SyncActionItem() {
	// Get audit action database
	database, err := c.service.Notion.GetDatabase(AuditActionBDID)
	if err != nil {
		c.logger.Error(err, "failed to get database from notion")
		return
	}

	// Map notion page
	mapPage := NotionDatabaseToMap(database)

	// Get all audit cycle from database
	actionItems, err := c.store.ActionItem.All(c.repo.DB())
	if err != nil {
		c.logger.Error(err, "failed to get audit cycle from database")
		return
	}

	actionItemMap := model.ActionItemToMap(actionItems)

	// Compare and update
	// Create audit cycle if not exist in auditCycleMap
	for _, page := range mapPage {
		pp.Println("Create audit cycle ID:", page.ID)
		// If auditCycle not exist in mapPage
		actionItempProperties := page.Properties.(notion.DatabasePageProperties)
		if _, ok := actionItemMap[model.MustGetUUIDFromString(page.ID)]; !ok {
			// Create audits
			pic := &model.Employee{}

			if actionItempProperties["Auditor"].People != nil && len(actionItempProperties["Auditor"].People) > 0 {
				pic, err = c.store.Employee.OneByNotionID(c.repo.DB(), actionItempProperties["Auditor"].People[0].ID)
				if err != nil {
					c.logger.Error(err, "failed to get pic from notion id")
					return
				}
			}

			newActionItem := model.NewActionItemFromNotionPage(page, pic.ID)
			newActionItem, err = c.store.ActionItem.Create(c.repo.DB(), newActionItem)
			if err != nil {
				c.logger.Error(err, "failed to create action item")
				return
			}

		}
	}
}
