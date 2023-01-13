package audit

import (
	"errors"
	"reflect"

	"github.com/dstotijn/go-notion"
	"gorm.io/gorm"

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
	VoconicAudit1  = "0fde875d-67b1-4c23-b524-4b541b301c62"
	NghenhanAudit1 = "23d6b63b-3a2b-49fc-8af7-5ffca1379a29"
)

// SyncAuditCycle sync audit cycle from notion to database
func (c *cronjob) SyncAuditCycle() {
	l := c.logger.Fields(logger.Fields{
		"method": "SyncAuditCycle",
	})

	l.Infof("Sync audit cycle started")

	// Start Transaction
	tx, done := c.repo.NewTransaction()

	// Get all audit cycle from notion
	database, err := c.service.Notion.GetDatabase(c.config.Notion.AuditCycleBDID, nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return
	}

	// Map notion page
	mapPage := NotionDatabaseToMap(tx.DB(), database)

	// Get all audit cycle from database
	auditCycles, err := c.store.AuditCycle.All(tx.DB())
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		done(err)
		return
	}

	auditCycleMap := model.AuditCycleToMap(auditCycles)

	// Compare and update
	// Create audit cycle if not exist in auditCycleMap
	for _, page := range mapPage {
		// If auditCycle not exist in mapPage
		if ac, ok := auditCycleMap[model.MustGetUUIDFromString(page.ID)]; !ok {
			l.Infof("Create audit cycle ID: %s", page.ID)
			auditCycle := model.NewAuditCycleFromNotionPage(&page, c.config.Notion.AuditCycleBDID)
			// Create audits
			if err := c.createAudits(tx.DB(), &page, auditCycle); err != nil {
				l.Error(err, "failed to create audit")
				done(err)
				return
			}

			// Create audit cycle
			auditCycle, err = c.store.AuditCycle.Create(tx.DB(), auditCycle)
			if err != nil {
				l.Error(err, "failed to create audit cycle")
				done(err)
				return
			}
		} else {
			l.Infof("Sync audit cycle ID: %s", page.ID)
			delete(auditCycleMap, model.MustGetUUIDFromString(page.ID))
			if err := c.syncAuditCycle(tx.DB(), &page, ac); err != nil {
				l.Error(err, "failed to sync audit cycle")
				done(err)
				return
			}
		}
	}

	// Delete audit cycle if not exist in mapPage
	for _, auditCycle := range auditCycleMap {
		l.Infof("Delete audit cycle ID: %s", auditCycle.ID.String())
		if err := c.deleteAuditCycle(tx.DB(), auditCycle); err != nil {
			l.Error(err, "failed to delete audit cycle")
			done(err)
			return
		}
	}

	// Set value action item high, action item to 0
	if err := c.store.AuditCycle.ResetActionItem(tx.DB()); err != nil {
		l.Error(err, "failed to reset action item")
		done(err)
		return
	}

	if err := c.store.Audit.ResetActionItem(tx.DB()); err != nil {
		l.Error(err, "failed to reset action item")
		done(err)
		return
	}

	// Sync Action Item
	if err := c.SyncActionItem(tx.DB()); err != nil {
		l.Error(err, "failed to sync action item")
		done(err)
		return
	}

	done(nil)

	l.Infof("Sync audit cycle finished")

	return
}

// SyncActionItem sync action item from notion to database
func (c *cronjob) SyncActionItem(db *gorm.DB) error {
	l := c.logger.Fields(logger.Fields{
		"method": "SyncActionItem",
	})

	l.Infof("Syncing action item started")
	// Get all audit cycle from database
	actionItems, err := c.store.ActionItem.All(db)
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	actionItemMap := model.ActionItemToMap(actionItems)

	// Sync action item
	if err := c.syncActionItemPage(db, c.config.Notion.AuditActionItemBDID, false, actionItemMap); err != nil {
		l.Error(err, "failed to run function syncActionItemPage")
		return err
	}

	// Delete non-exist action item
	for _, actionItem := range actionItemMap {
		if err := c.store.ActionItem.Delete(db, actionItem.ID.String()); err != nil {
			l.Error(err, "failed to delete action item")
			return err
		}
	}

	// Sync audit action item
	if err := c.syncAuditActionItem(db); err != nil {
		l.Error(err, "failed to run function syncAuditActionItem")
		return err
	}

	if err := c.snapShot(db); err != nil {
		l.Error(err, "failed to run function snapShot")
		return err
	}

	l.Infof("Syncing action item finished")

	return nil
}

// createAudits create audit record for each audit cycle
func (c *cronjob) createAudits(db *gorm.DB, page *notion.Page, auditCycle *model.AuditCycle) error {
	l := c.logger.Fields(logger.Fields{
		"method":       "createAudits",
		"pageID":       page.ID,
		"auditCycleID": auditCycle.ID.String(),
	})

	audit, err := c.service.Notion.GetBlockChildren(page.ID)
	if err != nil {
		l.Error(err, "failed to get audit from notion")
		return err
	}

	// Find the audit checklist block index
	auditChecklistIndex := -1
	for index, block := range audit.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			auditChecklistIndex = index
		}
	}

	if page.ID == NghenhanAudit1 || page.ID == VoconicAudit1 || auditChecklistIndex == -1 {
		return nil
	}

	properties := page.Properties.(notion.DatabasePageProperties)
	auditChecklist, err := c.service.Notion.GetDatabase(audit.Results[auditChecklistIndex].ID(), nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Create audit record for each row
	for _, row := range auditChecklist.Results {
		if err := c.createAudit(db, page, row, properties, auditChecklistIndex, audit, auditCycle); err != nil {
			l.Error(err, "failed to create audit")
			return err
		}
	}

	return nil
}

// syncActionItemPage sync a action item include create action item if not exist, dekete, udpate exist item and update action item high in audit cycle
func (c *cronjob) syncActionItemPage(db *gorm.DB, databaseID string, withStartCursor bool, actionItemMap map[model.UUID]*model.ActionItem) error {
	l := c.logger.Fields(logger.Fields{
		"method":          "syncActionItemPage",
		"databaseID":      databaseID,
		"withStartCursor": withStartCursor,
	})

	// Get audit action database
	var database *notion.DatabaseQueryResponse
	var err error

	if withStartCursor {
		database, err = c.service.Notion.GetDatabaseWithStartCursor(c.config.Notion.AuditActionItemBDID, databaseID)
		if err != nil {
			l.Error(err, "failed to get database from notion")
			return err
		}
	} else {
		database, err = c.service.Notion.GetDatabase(databaseID, nil, nil)
		if err != nil {
			l.Error(err, "failed to get database from notion")
			return err
		}
	}

	if database.HasMore {
		err = c.syncActionItemPage(db, *database.NextCursor, true, actionItemMap)
		if err != nil {
			l.Error(err, "failed to run function syncActionItemPage")
			return err
		}
	}

	// Map notion page
	mapPage := NotionDatabaseToMap(db, database)

	// Compare and update
	// Create action item if not exist in actionItemMap
	for _, page := range mapPage {
		l.Infof("Create action item ID: %s", page.ID)
		// If actionItem not exist in mapPage
		actionItempProperties := page.Properties.(notion.DatabasePageProperties)

		pic := &model.Employee{}

		if actionItempProperties["PIC"].People != nil && len(actionItempProperties["PIC"].People) > 0 {
			pic, err = c.store.Employee.OneByNotionID(db, actionItempProperties["PIC"].People[0].ID)
			if err != nil {
				l.Error(err, "failed to get pic from notion id")
				return err
			}
		}

		newActionItem := model.NewActionItemFromNotionPage(page, pic.ID, c.config.Notion.AuditActionItemBDID)

		if !newActionItem.AuditCycleID.IsZero() {
			auditCycle, err := c.store.AuditCycle.One(db, newActionItem.AuditCycleID.String())
			if err != nil {
				l.Error(err, "failed to get audit cycle from database")
				return err
			}

			newActionItem.ProjectID = auditCycle.ProjectID
		}

		if ai, ok := actionItemMap[model.MustGetUUIDFromString(page.ID)]; !ok {
			// Create action item
			newActionItem, err = c.store.ActionItem.Create(db, newActionItem)
			if err != nil {
				l.Error(err, "failed to create action item")
				return err
			}
		} else {
			// Update action item
			if !model.CompareActionItem(ai, newActionItem) {
				// Update action item
				newActionItem.ID = ai.ID
				newActionItem, err = c.store.ActionItem.UpdateSelectedFieldsByID(db, newActionItem.ID.String(), *newActionItem,
					"project_id",
					"notion_db_id",
					"pic_id",
					"audit_cycle_id",
					"name",
					"description",
					"need_help",
					"priority",
					"status")
				if err != nil {
					l.Error(err, "failed to update action item")
					return err
				}
			}

			delete(actionItemMap, model.MustGetUUIDFromString(page.ID))
		}

		// Update audit cycle
		if !newActionItem.AuditCycleID.IsZero() {
			auditCycle, err := c.store.AuditCycle.One(db, newActionItem.AuditCycleID.String())
			if err != nil {
				l.Error(err, "failed to get audit cycle from database")
				return err
			}

			if newActionItem.Priority != nil {
				// Update audit cycle
				switch *newActionItem.Priority {
				case model.ActionItemPriorityLow:
					auditCycle.ActionItemLow++
				case model.ActionItemPriorityMedium:
					auditCycle.ActionItemMedium++
				case model.ActionItemPriorityHigh:
					auditCycle.ActionItemHigh++
				}

				if auditCycle, err = c.store.AuditCycle.Update(db, auditCycle); err != nil {
					l.Error(err, "failed to update audit cycle")
					return err
				}
			}

		}
	}

	return nil
}

// syncAuditCycle sync audit cycle if it already exists
func (c *cronjob) syncAuditCycle(db *gorm.DB, page *notion.Page, auditCycle *model.AuditCycle) error {
	l := c.logger.Fields(logger.Fields{
		"method": "syncAuditCycle",
		"pageID": page.ID,
	})

	newAuditCycle := model.NewAuditCycleFromNotionPage(page, c.config.Notion.AuditCycleBDID)

	cloneAuditCycle, err := c.store.AuditCycle.One(db, newAuditCycle.ID.String())
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	if err := c.syncAudit(db, page, cloneAuditCycle); err != nil {
		l.Error(err, "failed to sync audit")
		return err
	}

	if !model.CompareAuditCycle(auditCycle, newAuditCycle) || !compareAuditID(auditCycle.HealthAuditID, cloneAuditCycle.HealthAuditID) || !compareAuditID(auditCycle.ProcessAuditID, cloneAuditCycle.ProcessAuditID) ||
		!compareAuditID(auditCycle.BackendAuditID, cloneAuditCycle.BackendAuditID) || !compareAuditID(auditCycle.FrontendAuditID, cloneAuditCycle.FrontendAuditID) || !compareAuditID(auditCycle.BlockchainAuditID, cloneAuditCycle.BlockchainAuditID) ||
		!compareAuditID(auditCycle.SystemAuditID, cloneAuditCycle.SystemAuditID) || !compareAuditID(auditCycle.MobileAuditID, cloneAuditCycle.MobileAuditID) {
		// Update audit cycle
		newAuditCycle.HealthAuditID = cloneAuditCycle.HealthAuditID
		newAuditCycle.ProcessAuditID = cloneAuditCycle.ProcessAuditID
		newAuditCycle.BackendAuditID = cloneAuditCycle.BackendAuditID
		newAuditCycle.FrontendAuditID = cloneAuditCycle.FrontendAuditID
		newAuditCycle.BlockchainAuditID = cloneAuditCycle.BlockchainAuditID
		newAuditCycle.SystemAuditID = cloneAuditCycle.SystemAuditID
		newAuditCycle.MobileAuditID = cloneAuditCycle.MobileAuditID

		_, err := c.store.AuditCycle.UpdateSelectedFieldsByID(db, newAuditCycle.ID.String(), *newAuditCycle,
			"project_id",
			"notion_db_id",
			"health_audit_id",
			"process_audit_id",
			"backend_audit_id",
			"frontend_audit_id",
			"blockchain_audit_id",
			"system_audit_id",
			"mobile_audit_id",
			"cycle",
			"average_score",
			"flag",
			"quarter")

		if err != nil {
			l.Error(err, "failed to update audit cycle")
			return err
		}
	}

	return nil
}

func compareAuditID(oldID, newID *model.UUID) bool {
	if oldID == nil && newID == nil {
		return true
	}

	if oldID == nil || newID == nil {
		return false
	}

	return oldID.String() == newID.String()
}

// deleteAuditCycle delete audit cycle if it not exists in notion
func (c *cronjob) deleteAuditCycle(db *gorm.DB, auditCycle *model.AuditCycle) error {
	l := c.logger.Fields(logger.Fields{
		"method":       "deleteAuditCycle",
		"auditCycleID": auditCycle.ID.String(),
	})

	auditMap := model.AuditMap(*auditCycle)

	for auditID := range auditMap {
		if err := c.deleteAudit(db, auditID); err != nil {
			l.Error(err, "failed to delete audit")
			return err
		}
	}

	if err := c.store.AuditCycle.Delete(db, auditCycle.ID.String()); err != nil {
		l.Error(err, "failed to delete audit cycle")
		return err
	}

	return nil
}

// syncAudit sync audit if it already exists
func (c *cronjob) syncAudit(db *gorm.DB, page *notion.Page, auditCycle *model.AuditCycle) error {
	l := c.logger.Fields(logger.Fields{
		"method":       "syncAudit",
		"pageID":       page.ID,
		"auditCycleID": auditCycle.ID.String(),
	})

	audit, err := c.service.Notion.GetBlockChildren(page.ID)
	if err != nil {
		l.Error(err, "failed to get audit from notion")
		return err
	}

	// Find the audit checklist block index
	auditChecklistIndex := -1
	for index, block := range audit.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			auditChecklistIndex = index
		}
	}

	if page.ID == NghenhanAudit1 || page.ID == VoconicAudit1 || auditChecklistIndex == -1 {
		return nil
	}

	properties := page.Properties.(notion.DatabasePageProperties)
	auditChecklist, err := c.service.Notion.GetDatabase(audit.Results[auditChecklistIndex].ID(), nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	auditMap := model.AuditMap(*auditCycle)

	// Sync audit record for each row
	for _, row := range auditChecklist.Results {
		// Check audit existence in database
		if _, ok := auditMap[model.MustGetUUIDFromString(row.ID)]; !ok {
			// if audit is not exist create audit
			if c.createAudit(db, page, row, properties, auditChecklistIndex, audit, auditCycle); err != nil {
				l.Error(err, "failed to create audit")
				return err
			}
		} else {
			// Delete audit from audit map
			delete(auditMap, model.MustGetUUIDFromString(row.ID))

			// Check audit info with info in the database
			// Get audit in database
			auditDB, err := c.store.Audit.One(db, row.ID)
			if err != nil {
				l.Error(err, "failed to get audit from database")
				return err
			}

			// Get auditor from notion id
			checklistProperties := row.Properties.(notion.DatabasePageProperties)
			if len(checklistProperties["Auditor"].People) == 0 {
				l.Error(errs.ErrMissingAuditorInAudit, "missing auditor in audit")
				return errs.ErrMissingAuditorInAudit
			}

			auditor, err := c.store.Employee.OneByNotionID(db, checklistProperties["Auditor"].People[0].ID)
			if err != nil {
				l.Error(err, "failed to get auditor from notion id")
				return err
			}

			flag, err := c.getFlag(db, page, &row)
			if err != nil {
				l.Error(err, "failed to get flag")
				return errs.ErrFailedToGetFlag
			}

			if len(properties["Project"].Relation) == 0 {
				l.Error(errs.ErrMissingProjectInAudit, "missing project in audit")
				return errs.ErrMissingProjectInAudit
			}

			newAudit := model.NewAuditFromNotionPage(row, properties["Project"].Relation[0].ID, auditor.ID, flag, audit.Results[auditChecklistIndex].ID())

			// compare new audit object
			if !model.CompareAudit(*auditDB, *newAudit) {
				// Update audit
				newAudit.ID = auditDB.ID
				if _, err := c.store.Audit.UpdateSelectedFieldsByID(db, newAudit.ID.String(), *newAudit,
					"project_id",
					"notion_db_id",
					"auditor_id",
					"name",
					"type",
					"score",
					"status",
					"flag",
					"action_item",
					"duration",
					"audited_at"); err != nil {
					l.Error(err, "failed to update audit")
					return err
				}
			}

			// Sync audit participant - DONE
			if err := c.syncParticipant(db, checklistProperties, newAudit.ID); err != nil {
				l.Error(err, "failed to sync audit participant")
				return err
			}

			// Sync Audit Items - DONE
			if err := c.syncAuditItem(db, page, &row); err != nil {
				l.Error(err, "failed to sync audit item")
				return err
			}
		}
	}

	// Delete non-existent audits
	for key := range auditMap {
		deletedAudit, err := c.store.Audit.One(db, key.String())
		if err != nil {
			l.Error(err, "failed to get audit from database")
			return err
		}

		c.deleteAuditIDForAuditCycle(db, deletedAudit, auditCycle)
		if err := c.deleteAudit(db, key); err != nil {
			l.Error(err, "failed to delete audit")
			return err
		}
	}

	return nil
}

func (c *cronjob) deleteAudit(db *gorm.DB, auditID model.UUID) error {
	l := c.logger.Fields(logger.Fields{
		"method":  "deleteAudit",
		"auditID": auditID.String(),
	})

	// Delete audit item
	if err := c.store.AuditItem.DeleteByAuditID(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit item")
		return err
	}

	// Delete audit participant
	if err := c.store.AuditParticipant.DeleteByAuditID(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit participant")
		return err
	}

	// Delete audit action item by audit id
	if err := c.store.AuditActionItem.DeleteByAuditID(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit action item")
		return err
	}

	// Delete audit
	if err := c.store.Audit.Delete(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit")
		return err
	}

	return nil
}

func (c *cronjob) createAudit(db *gorm.DB, page *notion.Page, row notion.Page, properties notion.DatabasePageProperties, auditChecklistIndex int, audit *notion.BlockChildrenResponse, auditCycle *model.AuditCycle) error {
	l := c.logger.Fields(logger.Fields{
		"method": "createAudit",
	})
	// get auditor from notion id
	checklistProperties := row.Properties.(notion.DatabasePageProperties)
	if len(checklistProperties["Auditor"].People) == 0 {
		l.Error(errs.ErrMissingAuditorInAudit, "auditor is missing in audit")
		return errs.ErrMissingAuditorInAudit
	}

	auditor, err := c.store.Employee.OneByNotionID(db, checklistProperties["Auditor"].People[0].ID)
	if err != nil {
		l.Error(err, "failed to get auditor from notion id")
		return err
	}

	// Create new audit object
	flag, err := c.getFlag(db, page, &row)
	if err != nil {
		l.Error(err, "failed to get flag")
		return errs.ErrFailedToGetFlag
	}

	projectID := ""
	if properties["Project"].Relation != nil && len(properties["Project"].Relation) > 0 {
		projectID = properties["Project"].Relation[0].ID
	}

	newAudit := model.NewAuditFromNotionPage(row, projectID, auditor.ID, flag, audit.Results[auditChecklistIndex].ID())
	newAudit, err = c.store.Audit.Create(db, newAudit)

	// Create new audit participant
	if len(checklistProperties["Participants"].People) > 0 {
		for _, p := range checklistProperties["Participants"].People {
			participant, err := c.store.Employee.OneByNotionID(db, p.ID)
			if err != nil {
				l.Error(err, "failed to get participant by notion id")
				return err
			}

			_, err = c.store.AuditParticipant.Create(db, &model.AuditParticipant{
				AuditID:    newAudit.ID,
				EmployeeID: participant.ID,
			})

			if err != nil {
				l.Error(err, "failed to create audit participant")
				return err
			}
		}
	}

	// Update audit cycle
	c.updateAuditIDForAuditCycle(db, newAudit, auditCycle)

	// Create Audit Items
	if err := c.createAuditItem(db, page, &row); err != nil {
		l.Error(err, "failed to create audit item")
		return err
	}

	return nil
}

func (c *cronjob) syncParticipant(db *gorm.DB, checklistProperties notion.DatabasePageProperties, auditID model.UUID) error {
	l := c.logger.Fields(logger.Fields{
		"method":  "syncParticipant",
		"auditID": auditID.String(),
	})

	// Get all audit participant
	auditParticipants, err := c.store.AuditParticipant.AllByAuditID(db, auditID.String())
	if err != nil {
		l.Error(err, "failed to get all audit participant")
		return err
	}

	apMap := model.AuditParticipantToMap(auditParticipants)

	if len(checklistProperties["Participants"].People) > 0 {
		for _, p := range checklistProperties["Participants"].People {
			participant, err := c.store.Employee.OneByNotionID(db, p.ID)
			if err != nil {
				l.Error(err, "failed to get participant by notion id")
				return err
			}

			if _, ok := apMap[participant.ID]; !ok {
				_, err = c.store.AuditParticipant.Create(db, &model.AuditParticipant{
					AuditID:    auditID,
					EmployeeID: participant.ID,
				})

				if err != nil {
					l.Error(err, "failed to create audit participant")
					return err
				}
			} else {
				delete(apMap, participant.ID)
			}
		}

		for _, p := range apMap {
			if err := c.store.AuditParticipant.Delete(db, p.ID.String()); err != nil {
				l.Error(err, "failed to delete audit participant")
				return err
			}
		}
	}

	return nil
}

func (c *cronjob) deleteAuditIDForAuditCycle(db *gorm.DB, audit *model.Audit, auditCycle *model.AuditCycle) {
	switch audit.Type {
	case model.AuditTypeHealth:
		if *auditCycle.HealthAuditID == audit.ID {
			auditCycle.HealthAuditID = nil
		}
	case model.AuditTypeProcess:
		if *auditCycle.ProcessAuditID == audit.ID {
			auditCycle.ProcessAuditID = nil
		}
	case model.AuditTypeBackend:
		if *auditCycle.BackendAuditID == audit.ID {
			auditCycle.BackendAuditID = nil
		}
	case model.AuditTypeFrontend:
		if *auditCycle.FrontendAuditID == audit.ID {
			auditCycle.FrontendAuditID = nil
		}
	case model.AuditTypeSystem:
		if *auditCycle.SystemAuditID == audit.ID {
			auditCycle.SystemAuditID = nil
		}
	case model.AuditTypeMobile:
		if *auditCycle.MobileAuditID == audit.ID {
			auditCycle.MobileAuditID = nil
		}
	case model.AuditTypeBlockchain:
		if *auditCycle.BlockchainAuditID == audit.ID {
			auditCycle.BlockchainAuditID = nil
		}
	}
}

func (c *cronjob) updateAuditIDForAuditCycle(db *gorm.DB, audit *model.Audit, auditCycle *model.AuditCycle) {
	switch audit.Type {
	case model.AuditTypeHealth:
		auditCycle.HealthAuditID = &audit.ID
	case model.AuditTypeProcess:
		auditCycle.ProcessAuditID = &audit.ID
	case model.AuditTypeBackend:
		auditCycle.BackendAuditID = &audit.ID
	case model.AuditTypeFrontend:
		auditCycle.FrontendAuditID = &audit.ID
	case model.AuditTypeSystem:
		auditCycle.SystemAuditID = &audit.ID
	case model.AuditTypeMobile:
		auditCycle.MobileAuditID = &audit.ID
	case model.AuditTypeBlockchain:
		auditCycle.BlockchainAuditID = &audit.ID
	}

	if audit.Score != 0 {
		auditCycle.Status = model.AuditStatusAudited
	}
}

func (c *cronjob) getFlag(db *gorm.DB, page *notion.Page, row *notion.Page) (model.AuditFlag, error) {
	l := c.logger.Fields(logger.Fields{
		"method": "getFlag",
		"pageID": page.ID,
		"rowID":  row.ID,
	})

	countPoor, countAcceptable := 0, 0
	checklistPage, err := c.service.Notion.GetBlockChildren(row.ID)
	if err != nil {
		l.Error(err, "failed to get audit from notion")
		return "", err
	}

	// Find the audit checklist block index
	checklistDatabaseIndex := -1
	for index, block := range checklistPage.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			checklistDatabaseIndex = index
		}
	}

	checklistDatabase, err := c.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
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

func (c *cronjob) createAuditItem(db *gorm.DB, page *notion.Page, row *notion.Page) error {
	l := c.logger.Fields(logger.Fields{
		"method": "createAuditItem",
		"pageID": page.ID,
		"rowID":  row.ID,
	})

	l.Infof("Create audit item %s", row.ID)
	checklistPage, err := c.service.Notion.GetBlockChildren(row.ID)
	if err != nil {
		l.Error(err, "failed to get audit from notion")
		return err
	}

	// Find the audit checklist block index
	checklistDatabaseIndex := -1
	for index, block := range checklistPage.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			checklistDatabaseIndex = index
		}
	}

	checklistDatabase, err := c.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Create audit item record for each row
	for _, checklist := range checklistDatabase.Results {
		// Create new audit object
		newAuditItem := model.NewAuditItemFromNotionPage(checklist, row.ID, checklistPage.Results[checklistDatabaseIndex].ID())
		newAuditItem, err = c.store.AuditItem.Create(db, newAuditItem)
	}

	return nil
}

func (c *cronjob) syncAuditItem(db *gorm.DB, page *notion.Page, row *notion.Page) error {
	l := c.logger.Fields(logger.Fields{
		"method": "syncAuditItem",
		"pageID": page.ID,
		"rowID":  row.ID,
	})

	// Get all audit items
	auditItems, err := c.store.AuditItem.AllByAuditID(db, row.ID)
	if err != nil {
		l.Error(err, "failed to get audit items from database")
		return err
	}

	aiMap := model.AuditItemToMap(auditItems)

	l.Infof("Sync audit item %s", row.ID)
	checklistPage, err := c.service.Notion.GetBlockChildren(row.ID)
	if err != nil {
		l.Error(err, "failed to get audit from notion")
		return err
	}

	// Find the audit checklist block index
	checklistDatabaseIndex := -1
	for index, block := range checklistPage.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			checklistDatabaseIndex = index
		}
	}

	checklistDatabase, err := c.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Sync audit item record for each row
	for _, checklist := range checklistDatabase.Results {
		if _, ok := aiMap[model.MustGetUUIDFromString(checklist.ID)]; !ok {
			// Create new audit object
			newAuditItem := model.NewAuditItemFromNotionPage(checklist, row.ID, checklistPage.Results[checklistDatabaseIndex].ID())
			newAuditItem, err = c.store.AuditItem.Create(db, newAuditItem)
			if err != nil {
				l.Error(err, "failed to create audit item")
				return err
			}
		} else {
			// Update audit object
			auditItem := aiMap[model.MustGetUUIDFromString(checklist.ID)]
			newAuditItem := model.NewAuditItemFromNotionPage(checklist, row.ID, checklistPage.Results[checklistDatabaseIndex].ID())
			if !model.CompareAuditItem(&auditItem, newAuditItem) {
				newAuditItem.ID = auditItem.ID
				if _, err := c.store.AuditItem.UpdateSelectedFieldsByID(db, newAuditItem.ID.String(), *newAuditItem,
					"audit_id",
					"notion_db_id",
					"name",
					"area",
					"requirements",
					"grade",
					"severity",
					"notes",
					"action_item_id"); err != nil {
					l.Error(err, "failed to update audit item")
					return err
				}
			}
			delete(aiMap, model.MustGetUUIDFromString(checklist.ID))
		}
	}

	// Delete audit items not exist
	for _, auditItem := range aiMap {
		if err := c.store.AuditItem.Delete(db, auditItem.ID.String()); err != nil {
			l.Error(err, "failed to delete audit item")
			return err
		}
	}

	return nil
}

func NotionDatabaseToMap(db *gorm.DB, database *notion.DatabaseQueryResponse) map[string]notion.Page {
	result := make(map[string]notion.Page)
	for _, r := range database.Results {
		result[r.ID] = r
	}

	return result
}

// syncAuditActionItem sync audit action item for all audit cycle, audit, audit item, audit action item
func (c *cronjob) syncAuditActionItem(db *gorm.DB) error {
	l := c.logger.Fields(logger.Fields{
		"method": "syncAuditActionItem",
	})

	l.Infof("Syncing audit action item started")

	// Get all audit cycle from database
	auditCycles, err := c.store.AuditCycle.All(db)
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	// Sync audit action item for all audit cycles
	for _, auditCycle := range auditCycles {
		if err := c.syncAuditActionItemInAuditCycle(db, auditCycle); err != nil {
			l.Error(err, "failed to create audit")
			return err
		}
	}

	l.Infof("Syncing audit action item finished")
	return err
}

func (c *cronjob) syncAuditActionItemInAuditCycle(db *gorm.DB, auditCycle *model.AuditCycle) error {
	l := c.logger.Fields(logger.Fields{
		"method":       "syncAuditActionItemInAuditCycle",
		"auditCycleID": auditCycle.ID.String(),
	})

	l.Infof("Sync Audit Action Item for audit cycle: %s", auditCycle.ID.String())
	// check if audit cycle has audit
	if auditCycle.HealthAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.HealthAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.HealthAuditID.String())
			return err
		}
	}

	if auditCycle.ProcessAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.ProcessAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.ProcessAuditID.String())
			return err
		}
	}

	if auditCycle.FrontendAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.FrontendAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.FrontendAuditID.String())
			return err
		}
	}

	if auditCycle.BackendAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.BackendAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.BackendAuditID.String())
			return err
		}
	}

	if auditCycle.MobileAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.MobileAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.MobileAuditID.String())
			return err
		}
	}

	if auditCycle.BlockchainAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.BlockchainAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.BlockchainAuditID.String())
			return err
		}
	}

	if auditCycle.SystemAuditID != nil {
		if err := c.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.SystemAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.SystemAuditID.String())
			return err
		}
	}

	return nil
}

func (c *cronjob) syncAuditActionItemInAudit(db *gorm.DB, auditCycle *model.AuditCycle, auditID model.UUID) error {
	l := c.logger.Fields(logger.Fields{
		"method":       "syncAuditActionItemInAudit",
		"auditID":      auditID.String(),
		"auditCycleID": auditCycle.ID.String(),
	})

	l.Infof("Sync Audit Action Item for audit: %s", auditID.String())
	audit, err := c.store.Audit.One(db, auditID.String())
	if err != nil {
		l.Error(err, "failed to get audit from database")
		return err
	}

	// call api to get data of the audit from notion
	checklistPage, err := c.service.Notion.GetBlockChildren(audit.ID.String())
	if err != nil {
		l.Error(err, "failed to get audit from notion")
		return err
	}

	// Find the audit checklist block index
	checklistDatabaseIndex := -1
	for index, block := range checklistPage.Results {
		if reflect.TypeOf(block) == reflect.TypeOf(&notion.ChildDatabaseBlock{}) {
			checklistDatabaseIndex = index
		}
	}

	checklistDatabase, err := c.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Get all audit action item in the database
	auditActionItem, err := c.store.AuditActionItem.AllByAuditID(db, auditID.String())
	if err != nil {
		l.Error(err, "failed to get audit action item from database")
		return err
	}

	aaiMap := model.AuditActionItemToMap(auditActionItem)

	// Loop through all audit items and update to database if it have value in Audit action items field
	for _, checklist := range checklistDatabase.Results {
		checklistProperties := checklist.Properties.(notion.DatabasePageProperties)
		// Get audit item
		auditItem, err := c.store.AuditItem.One(db, checklist.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				l.Errorf(err, "audit item with id: %s not found", checklist.ID)
				continue
			}

			l.Error(err, "failed to get audit item from database")
			return err
		}

		if checklistProperties["☎️ Audit Action Items"].Relation == nil && auditItem.ActionItemID != nil {
			// Update audit item if action item to nil
			auditItem.ActionItemID = nil
			if auditItem, err = c.store.AuditItem.UpdateSelectedFieldsByID(db, auditItem.ID.String(), *auditItem, "action_item_id"); err != nil {
				l.Error(err, "failed to update audit item")
				return err
			}
		} else if len(checklistProperties["☎️ Audit Action Items"].Relation) > 0 {
			// update audit_items(action_item_id)->audit(action_item)-> create record trong audit_action_items
			// Get action item
			actionItem, err := c.store.ActionItem.One(db, checklistProperties["☎️ Audit Action Items"].Relation[0].ID)
			if err != nil {
				// if err is not found skip update
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}

				l.Error(err, "failed to get action item from database")
				return err
			}

			// Update audit item if action item change
			if auditItem.ActionItemID == nil || (auditItem.ActionItemID != nil && *auditItem.ActionItemID != actionItem.ID) {
				auditItem.ActionItemID = &actionItem.ID
				if auditItem, err = c.store.AuditItem.UpdateSelectedFieldsByID(db, auditItem.ID.String(), *auditItem, "action_item_id"); err != nil {
					l.Error(err, "failed to update audit item")
					return err
				}
			}

			// Update audit
			audit.ActionItem++
			if audit, err = c.store.Audit.Update(db, audit); err != nil {
				l.Error(err, "failed to update audit")
				return err
			}

			// Update audit action item table
			if _, ok := aaiMap[model.AuditAction{AuditID: audit.ID, ActionItemID: actionItem.ID}]; ok {
				delete(aaiMap, model.AuditAction{AuditID: audit.ID, ActionItemID: actionItem.ID})
			} else {
				// Create audit action item
				auditActionItem := &model.AuditActionItem{
					AuditID:      audit.ID,
					ActionItemID: actionItem.ID,
				}
				if auditActionItem, err = c.store.AuditActionItem.Create(db, auditActionItem); err != nil {
					l.Error(err, "failed to create audit action item")
					return err
				}
			}
		}
	}

	// Delete audit action item
	for _, aai := range aaiMap {
		if err := c.store.AuditActionItem.Delete(db, aai.ID.String()); err != nil {
			l.Error(err, "failed to delete audit action item")
			return err
		}
	}

	return nil
}

func (c *cronjob) snapShot(db *gorm.DB) error {
	l := c.logger.Fields(logger.Fields{
		"method": "snapShot",
	})

	// Get all audit cycle from database
	auditCycles, err := c.store.AuditCycle.All(db)
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	// Create action item snapshot record for each audit cycle
	for _, auditCycle := range auditCycles {
		if auditCycle.Project.Status != model.ProjectStatusClosed && auditCycle.Project.Status != model.ProjectStatusPaused {
			actionItemSnapshot := &model.ActionItemSnapshot{
				ProjectID:    auditCycle.ProjectID,
				AuditCycleID: auditCycle.ID,
				High:         auditCycle.ActionItemHigh,
				Medium:       auditCycle.ActionItemMedium,
				Low:          auditCycle.ActionItemLow,
			}

			if _, err := c.store.ActionItemSnapshot.Create(db, actionItemSnapshot); err != nil {
				l.Error(err, "failed to create action item snapshot")
				return err
			}
		}
	}

	return nil
}
