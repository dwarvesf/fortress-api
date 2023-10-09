package audit

import (
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/audit/errs"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

const (
	VoconicAudit1  = "0fde875d-67b1-4c23-b524-4b541b301c62"
	NghenhanAudit1 = "23d6b63b-3a2b-49fc-8af7-5ffca1379a29"
)

type handler struct {
	store   *store.Store
	service *service.Service
	logger  logger.Logger
	repo    store.DBRepo
	config  *config.Config
}

// New returns a handler
func New(store *store.Store, repo store.DBRepo, service *service.Service, logger logger.Logger, cfg *config.Config) IHandler {
	return &handler{
		store:   store,
		repo:    repo,
		service: service,
		logger:  logger,
		config:  cfg,
	}
}

// Sync godoc
// @Summary Sync audit info from Notion to database
// @Description Sync audit info from Notion to database
// @Tags Audit
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} MessageResponse
// @Router /cronjobs/audits [post]
func (h *handler) Sync(c *gin.Context) {
	h.SyncAuditCycle()

	c.JSON(http.StatusOK, view.CreateResponse[any](nil, nil, nil, nil, "sync audit from notion successfully"))
}

// SyncAuditCycle sync audit cycle from notion to database
func (h *handler) SyncAuditCycle() {
	l := h.logger.Fields(logger.Fields{
		"method": "SyncAuditCycle",
	})

	l.Infof("Sync audit cycle started")

	// Start Transaction
	tx, done := h.repo.NewTransaction()

	// Get all audit cycle from notion
	database, err := h.service.Notion.GetDatabase(h.config.Notion.Databases.AuditCycle, nil, nil, 0)
	if err != nil {
		l.Error(done(err), "failed to get database from notion")
		return
	}

	// Map notion page
	mapPage := NotionDatabaseToMap(tx.DB(), database)

	// Get all audit cycle from database
	auditCycles, err := h.store.AuditCycle.All(tx.DB())
	if err != nil {
		l.Error(done(err), "failed to get audit cycle from database")
		return
	}

	auditCycleMap := model.AuditCycleToMap(auditCycles)

	// Compare and update
	// Create audit cycle if not exist in auditCycleMap
	for _, page := range mapPage {
		// If auditCycle not exist in mapPage
		if ac, ok := auditCycleMap[model.MustGetUUIDFromString(page.ID)]; !ok {
			l.Infof("Create audit cycle ID: %s", page.ID)
			auditCycle := model.NewAuditCycleFromNotionPage(&page, h.config.Notion.Databases.AuditCycle)

			// Check project_is map with audit_notion_id in db or not
			exists, err := h.store.ProjectNotion.IsExistByAuditNotionID(tx.DB(), auditCycle.ProjectID.String())
			if err != nil {
				l.Error(done(err), "failed to check project_notion")
				return
			}

			if !exists {
				l.Infof("Project %s not exist in project_notion", auditCycle.ProjectID.String())
				continue
			}

			// Create audits
			if err := h.createAudits(tx.DB(), &page, auditCycle); err != nil {
				l.Error(done(err), "failed to create audit")
				return
			}

			// Create audit cycle
			_, err = h.store.AuditCycle.Create(tx.DB(), auditCycle)
			if err != nil {
				l.Error(done(err), "failed to create audit cycle")
				return
			}
		} else {
			l.Infof("Sync audit cycle ID: %s", page.ID)
			delete(auditCycleMap, model.MustGetUUIDFromString(page.ID))
			if err := h.syncAuditCycle(tx.DB(), &page, ac); err != nil {
				l.Error(done(err), "failed to sync audit cycle")
				return
			}
		}
	}

	// Delete audit cycle if not exist in mapPage
	for _, auditCycle := range auditCycleMap {
		l.Infof("Delete audit cycle ID: %s", auditCycle.ID.String())
		if err := h.deleteAuditCycle(tx.DB(), auditCycle); err != nil {
			l.Error(done(err), "failed to delete audit cycle")
			return
		}
	}

	// Set value action item high, action item to 0
	if err := h.store.AuditCycle.ResetActionItem(tx.DB()); err != nil {
		l.Error(done(err), "failed to reset action item")
		return
	}

	if err := h.store.Audit.ResetActionItem(tx.DB()); err != nil {
		l.Error(done(err), "failed to reset action item")
		return
	}

	// Sync Action Item
	if err := h.SyncActionItem(tx.DB()); err != nil {
		l.Error(done(err), "failed to sync action item")
		return
	}

	if err := done(nil); err != nil {
		l.Error(err, "failed to commit txn")
	}

	l.Infof("Sync audit cycle finished")
}

// SyncActionItem sync action item from notion to database
func (h *handler) SyncActionItem(db *gorm.DB) error {
	l := h.logger.Fields(logger.Fields{
		"method": "SyncActionItem",
	})

	l.Infof("Syncing action item started")
	// Get all audit cycle from database
	actionItems, err := h.store.ActionItem.All(db)
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	actionItemMap := model.ActionItemToMap(actionItems)

	// Sync action item
	if err := h.syncActionItemPage(db, h.config.Notion.Databases.AuditActionItem, false, actionItemMap); err != nil {
		l.Error(err, "failed to run function syncActionItemPage")
		return err
	}

	// Delete non-exist action item
	for _, actionItem := range actionItemMap {
		if err := h.store.ActionItem.Delete(db, actionItem.ID.String()); err != nil {
			l.Error(err, "failed to delete action item")
			return err
		}
	}

	// Sync audit action item
	if err := h.syncAuditActionItem(db); err != nil {
		l.Error(err, "failed to run function syncAuditActionItem")
		return err
	}

	if err := h.snapShot(db); err != nil {
		l.Error(err, "failed to run function snapShot")
		return err
	}

	l.Infof("Syncing action item finished")

	return nil
}

// createAudits create audit record for each audit cycle
func (h *handler) createAudits(db *gorm.DB, page *notion.Page, auditCycle *model.AuditCycle) error {
	l := h.logger.Fields(logger.Fields{
		"method":       "createAudits",
		"pageID":       page.ID,
		"auditCycleID": auditCycle.ID.String(),
	})

	audit, err := h.service.Notion.GetBlockChildren(page.ID)
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
	auditChecklist, err := h.service.Notion.GetDatabase(audit.Results[auditChecklistIndex].ID(), nil, nil, 0)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Create audit record for each row
	for _, row := range auditChecklist.Results {
		if err := h.createAudit(db, page, row, properties, auditChecklistIndex, audit, auditCycle); err != nil {
			l.Error(err, "failed to create audit")
			return err
		}
	}

	return nil
}

// syncActionItemPage sync a action item include create action item if not exist, dekete, udpate exist item and update action item high in audit cycle
func (h *handler) syncActionItemPage(db *gorm.DB, databaseID string, withStartCursor bool, actionItemMap map[model.UUID]*model.ActionItem) error {
	l := h.logger.Fields(logger.Fields{
		"method":          "syncActionItemPage",
		"databaseID":      databaseID,
		"withStartCursor": withStartCursor,
	})

	// Get audit action database
	var database *notion.DatabaseQueryResponse
	var err error

	if withStartCursor {
		database, err = h.service.Notion.GetDatabaseWithStartCursor(h.config.Notion.Databases.AuditActionItem, databaseID)
		if err != nil {
			l.Error(err, "failed to get database from notion")
			return err
		}
	} else {
		database, err = h.service.Notion.GetDatabase(databaseID, nil, nil, 0)
		if err != nil {
			l.Error(err, "failed to get database from notion")
			return err
		}
	}

	if database.HasMore {
		err = h.syncActionItemPage(db, *database.NextCursor, true, actionItemMap)
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
		picID := model.UUID{}

		if actionItempProperties["PIC"].People != nil && len(actionItempProperties["PIC"].People) > 0 {
			pic, err = h.store.Employee.OneByNotionID(db, actionItempProperties["PIC"].People[0].ID)
			if err != nil {
				l.Error(err, "failed to get pic from notion id")
			}
		}

		if pic != nil {
			picID = pic.ID
		}

		newActionItem := model.NewActionItemFromNotionPage(page, picID, h.config.Notion.Databases.AuditActionItem)

		// Check whether project_id map with audit_notion_id in db or not
		if !newActionItem.ProjectID.IsZero() {
			exits, err := h.store.ProjectNotion.IsExistByAuditNotionID(db, newActionItem.ProjectID.String())
			if err != nil {
				l.Error(err, "failed to check project_notion_id exits in db")
				return err
			}

			if !exits {
				l.Infof("Project ID: %s not exits in project_notion_id table", newActionItem.ProjectID.String())
				continue
			}
		}

		if !newActionItem.AuditCycleID.IsZero() {
			auditCycle, err := h.store.AuditCycle.One(db, newActionItem.AuditCycleID.String())
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					l.Infof("Audit cycle ID: %s not exits in db", newActionItem.AuditCycleID.String())
					continue
				}

				l.Error(err, "failed to get audit cycle from database")
				return err
			}

			newActionItem.ProjectID = auditCycle.ProjectID
		}

		if ai, ok := actionItemMap[model.MustGetUUIDFromString(page.ID)]; !ok {
			// Create action item
			newActionItem, err = h.store.ActionItem.Create(db, newActionItem)
			if err != nil {
				l.Error(err, "failed to create action item")
				return err
			}
		} else {
			// Update action item
			if !model.CompareActionItem(ai, newActionItem) {
				// Update action item
				newActionItem.ID = ai.ID
				newActionItem, err = h.store.ActionItem.UpdateSelectedFieldsByID(db, newActionItem.ID.String(), *newActionItem,
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
			auditCycle, err := h.store.AuditCycle.One(db, newActionItem.AuditCycleID.String())
			if err != nil {
				l.Error(err, "failed to get audit cycle from database")
				return err
			}

			if newActionItem.Priority != nil && newActionItem.Status != model.ActionItemStatusDone {
				// Update audit cycle
				switch *newActionItem.Priority {
				case model.ActionItemPriorityLow:
					auditCycle.ActionItemLow++
				case model.ActionItemPriorityMedium:
					auditCycle.ActionItemMedium++
				case model.ActionItemPriorityHigh:
					auditCycle.ActionItemHigh++
				}

				if _, err = h.store.AuditCycle.Update(db, auditCycle); err != nil {
					l.Error(err, "failed to update audit cycle")
					return err
				}
			}
		}
	}

	return nil
}

// syncAuditCycle sync audit cycle if it already exists
func (h *handler) syncAuditCycle(db *gorm.DB, page *notion.Page, auditCycle *model.AuditCycle) error {
	l := h.logger.Fields(logger.Fields{
		"method": "syncAuditCycle",
		"pageID": page.ID,
	})

	newAuditCycle := model.NewAuditCycleFromNotionPage(page, h.config.Notion.Databases.AuditCycle)

	cloneAuditCycle, err := h.store.AuditCycle.One(db, newAuditCycle.ID.String())
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	if err := h.syncAudit(db, page, cloneAuditCycle); err != nil {
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

		_, err := h.store.AuditCycle.UpdateSelectedFieldsByID(db, newAuditCycle.ID.String(), *newAuditCycle,
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
func (h *handler) deleteAuditCycle(db *gorm.DB, auditCycle *model.AuditCycle) error {
	l := h.logger.Fields(logger.Fields{
		"method":       "deleteAuditCycle",
		"auditCycleID": auditCycle.ID.String(),
	})

	auditMap := model.AuditMap(*auditCycle)

	for auditID := range auditMap {
		if err := h.deleteAudit(db, auditID); err != nil {
			l.Error(err, "failed to delete audit")
			return err
		}
	}

	if err := h.store.AuditCycle.Delete(db, auditCycle.ID.String()); err != nil {
		l.Error(err, "failed to delete audit cycle")
		return err
	}

	return nil
}

// syncAudit sync audit if it already exists
func (h *handler) syncAudit(db *gorm.DB, page *notion.Page, auditCycle *model.AuditCycle) error {
	l := h.logger.Fields(logger.Fields{
		"method":       "syncAudit",
		"pageID":       page.ID,
		"auditCycleID": auditCycle.ID.String(),
	})

	audit, err := h.service.Notion.GetBlockChildren(page.ID)
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
	auditChecklist, err := h.service.Notion.GetDatabase(audit.Results[auditChecklistIndex].ID(), nil, nil, 0)
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
			if err := h.createAudit(db, page, row, properties, auditChecklistIndex, audit, auditCycle); err != nil {
				l.Error(err, "failed to create audit")
				return err
			}
		} else {
			// Delete audit from audit map
			delete(auditMap, model.MustGetUUIDFromString(row.ID))

			// Check audit info with info in the database
			// Get audit in database
			auditDB, err := h.store.Audit.One(db, row.ID)
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

			auditor, err := h.store.Employee.OneByNotionID(db, checklistProperties["Auditor"].People[0].ID)
			if err != nil {
				l.Error(err, "failed to get auditor from notion id")
			}

			auditorID := model.UUID{}
			if auditor != nil {
				auditorID = auditor.ID
			}

			flag, err := h.getFlag(db, page, &row)
			if err != nil {
				l.Error(err, "failed to get flag")
				return errs.ErrFailedToGetFlag
			}

			if len(properties["Project"].Relation) == 0 {
				l.Error(errs.ErrMissingProjectInAudit, "missing project in audit")
				return errs.ErrMissingProjectInAudit
			}

			newAudit := model.NewAuditFromNotionPage(row, properties["Project"].Relation[0].ID, auditorID, flag, audit.Results[auditChecklistIndex].ID())

			// compare new audit object
			if !model.CompareAudit(*auditDB, *newAudit) {
				// Update audit
				newAudit.ID = auditDB.ID
				if _, err := h.store.Audit.UpdateSelectedFieldsByID(db, newAudit.ID.String(), *newAudit,
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
			if err := h.syncParticipant(db, checklistProperties, newAudit.ID); err != nil {
				l.Error(err, "failed to sync audit participant")
				return err
			}

			// Sync Audit Items - DONE
			if err := h.syncAuditItem(db, page, &row); err != nil {
				l.Error(err, "failed to sync audit item")
				return err
			}
		}
	}

	// Delete non-existent audits
	for key := range auditMap {
		deletedAudit, err := h.store.Audit.One(db, key.String())
		if err != nil {
			l.Error(err, "failed to get audit from database")
			return err
		}

		h.deleteAuditIDForAuditCycle(db, deletedAudit, auditCycle)
		if err := h.deleteAudit(db, key); err != nil {
			l.Error(err, "failed to delete audit")
			return err
		}
	}

	return nil
}

func (h *handler) deleteAudit(db *gorm.DB, auditID model.UUID) error {
	l := h.logger.Fields(logger.Fields{
		"method":  "deleteAudit",
		"auditID": auditID.String(),
	})

	// Delete audit item
	if err := h.store.AuditItem.DeleteByAuditID(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit item")
		return err
	}

	// Delete audit participant
	if err := h.store.AuditParticipant.DeleteByAuditID(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit participant")
		return err
	}

	// Delete audit action item by audit id
	if err := h.store.AuditActionItem.DeleteByAuditID(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit action item")
		return err
	}

	// Delete audit
	if err := h.store.Audit.Delete(db, auditID.String()); err != nil {
		l.Error(err, "failed to delete audit")
		return err
	}

	return nil
}

func (h *handler) createAudit(db *gorm.DB, page *notion.Page, row notion.Page, properties notion.DatabasePageProperties, auditChecklistIndex int, audit *notion.BlockChildrenResponse, auditCycle *model.AuditCycle) error {
	l := h.logger.Fields(logger.Fields{
		"method": "createAudit",
	})
	// get auditor from notion id
	checklistProperties := row.Properties.(notion.DatabasePageProperties)
	if len(checklistProperties["Auditor"].People) == 0 {
		l.Error(errs.ErrMissingAuditorInAudit, "auditor is missing in audit")
		return errs.ErrMissingAuditorInAudit
	}

	auditor, err := h.store.Employee.OneByNotionID(db, checklistProperties["Auditor"].People[0].ID)
	if err != nil {
		l.Error(err, "failed to get auditor from notion id")
		return err
	}

	// Create new audit object
	flag, err := h.getFlag(db, page, &row)
	if err != nil {
		l.Error(err, "failed to get flag")
		return errs.ErrFailedToGetFlag
	}

	projectID := ""
	if properties["Project"].Relation != nil && len(properties["Project"].Relation) > 0 {
		projectID = properties["Project"].Relation[0].ID
	}

	newAudit := model.NewAuditFromNotionPage(row, projectID, auditor.ID, flag, audit.Results[auditChecklistIndex].ID())
	// if something wrong with the audit object, return nil
	if newAudit == nil {
		return nil
	}

	newAudit, _ = h.store.Audit.Create(db, newAudit)

	// Create new audit participant
	if len(checklistProperties["Participants"].People) > 0 {
		for _, p := range checklistProperties["Participants"].People {
			participant, err := h.store.Employee.OneByNotionID(db, p.ID)
			if err != nil {
				l.Error(err, "failed to get participant by notion id")
				continue
			}

			_, err = h.store.AuditParticipant.Create(db, &model.AuditParticipant{
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
	h.updateAuditIDForAuditCycle(db, newAudit, auditCycle)

	// Create Audit Items
	if err := h.createAuditItem(db, page, &row); err != nil {
		l.Error(err, "failed to create audit item")
		return err
	}

	return nil
}

func (h *handler) syncParticipant(db *gorm.DB, checklistProperties notion.DatabasePageProperties, auditID model.UUID) error {
	l := h.logger.Fields(logger.Fields{
		"method":  "syncParticipant",
		"auditID": auditID.String(),
	})

	// Get all audit participant
	auditParticipants, err := h.store.AuditParticipant.AllByAuditID(db, auditID.String())
	if err != nil {
		l.Error(err, "failed to get all audit participant")
		return err
	}

	apMap := model.AuditParticipantToMap(auditParticipants)

	if len(checklistProperties["Participants"].People) > 0 {
		for _, p := range checklistProperties["Participants"].People {
			participant, err := h.store.Employee.OneByNotionID(db, p.ID)
			if err != nil {
				l.Error(err, "failed to get participant by notion id")
				continue
			}

			if _, ok := apMap[participant.ID]; !ok {
				_, err = h.store.AuditParticipant.Create(db, &model.AuditParticipant{
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
			if err := h.store.AuditParticipant.Delete(db, p.ID.String()); err != nil {
				l.Error(err, "failed to delete audit participant")
				return err
			}
		}
	}

	return nil
}

func (h *handler) deleteAuditIDForAuditCycle(db *gorm.DB, audit *model.Audit, auditCycle *model.AuditCycle) {
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

func (h *handler) updateAuditIDForAuditCycle(db *gorm.DB, audit *model.Audit, auditCycle *model.AuditCycle) {
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

func (h *handler) getFlag(db *gorm.DB, page *notion.Page, row *notion.Page) (model.AuditFlag, error) {
	l := h.logger.Fields(logger.Fields{
		"method": "getFlag",
		"pageID": page.ID,
		"rowID":  row.ID,
	})

	countPoor, countAcceptable := 0, 0
	checklistPage, err := h.service.Notion.GetBlockChildren(row.ID)
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

	if checklistDatabaseIndex == -1 {
		return model.AuditFlagGreen, nil
	}
	checklistDatabase, err := h.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil, 0)
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

func (h *handler) createAuditItem(db *gorm.DB, page *notion.Page, row *notion.Page) error {
	l := h.logger.Fields(logger.Fields{
		"method": "createAuditItem",
		"pageID": page.ID,
		"rowID":  row.ID,
	})

	l.Infof("Create audit item %s", row.ID)
	checklistPage, err := h.service.Notion.GetBlockChildren(row.ID)
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

	checklistDatabase, err := h.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil, 0)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Create audit item record for each row
	for _, checklist := range checklistDatabase.Results {
		// Create new audit object
		newAuditItem := model.NewAuditItemFromNotionPage(checklist, row.ID, checklistPage.Results[checklistDatabaseIndex].ID())
		_, err := h.store.AuditItem.Create(db, newAuditItem)
		if err != nil {
			l.Error(err, "failed to create audit item")
		}
	}

	return nil
}

func (h *handler) syncAuditItem(db *gorm.DB, page *notion.Page, row *notion.Page) error {
	l := h.logger.Fields(logger.Fields{
		"method": "syncAuditItem",
		"pageID": page.ID,
		"rowID":  row.ID,
	})

	// Get all audit items
	auditItems, err := h.store.AuditItem.AllByAuditID(db, row.ID)
	if err != nil {
		l.Error(err, "failed to get audit items from database")
		return err
	}

	aiMap := model.AuditItemToMap(auditItems)

	l.Infof("Sync audit item %s", row.ID)
	checklistPage, err := h.service.Notion.GetBlockChildren(row.ID)
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

	checklistDatabase, err := h.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil, 0)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Sync audit item record for each row
	for _, checklist := range checklistDatabase.Results {
		if _, ok := aiMap[model.MustGetUUIDFromString(checklist.ID)]; !ok {
			// Create new audit object
			newAuditItem := model.NewAuditItemFromNotionPage(checklist, row.ID, checklistPage.Results[checklistDatabaseIndex].ID())
			_, err = h.store.AuditItem.Create(db, newAuditItem)
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
				if _, err := h.store.AuditItem.UpdateSelectedFieldsByID(db, newAuditItem.ID.String(), *newAuditItem,
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
		if err := h.store.AuditItem.Delete(db, auditItem.ID.String()); err != nil {
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
func (h *handler) syncAuditActionItem(db *gorm.DB) error {
	l := h.logger.Fields(logger.Fields{
		"method": "syncAuditActionItem",
	})

	l.Infof("Syncing audit action item started")

	// Get all audit cycle from database
	auditCycles, err := h.store.AuditCycle.All(db)
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	// Sync audit action item for all audit cycles
	for _, auditCycle := range auditCycles {
		if err := h.syncAuditActionItemInAuditCycle(db, auditCycle); err != nil {
			l.Error(err, "failed to run syncAuditActionItemInAuditCycle function")
			return err
		}
	}

	l.Infof("Syncing audit action item finished")
	return err
}

func (h *handler) syncAuditActionItemInAuditCycle(db *gorm.DB, auditCycle *model.AuditCycle) error {
	l := h.logger.Fields(logger.Fields{
		"method":       "syncAuditActionItemInAuditCycle",
		"auditCycleID": auditCycle.ID.String(),
	})

	l.Infof("Sync Audit Action Item for audit cycle: %s", auditCycle.ID.String())
	// check if audit cycle has audit
	if auditCycle.HealthAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.HealthAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.HealthAuditID.String())
			return err
		}
	}

	if auditCycle.ProcessAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.ProcessAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.ProcessAuditID.String())
			return err
		}
	}

	if auditCycle.FrontendAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.FrontendAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.FrontendAuditID.String())
			return err
		}
	}

	if auditCycle.BackendAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.BackendAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.BackendAuditID.String())
			return err
		}
	}

	if auditCycle.MobileAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.MobileAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.MobileAuditID.String())
			return err
		}
	}

	if auditCycle.BlockchainAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.BlockchainAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.BlockchainAuditID.String())
			return err
		}
	}

	if auditCycle.SystemAuditID != nil {
		if err := h.syncAuditActionItemInAudit(db, auditCycle, *auditCycle.SystemAuditID); err != nil {
			l.Errorf(err, "failed to sync audit action item in audit %s", auditCycle.SystemAuditID.String())
			return err
		}
	}

	return nil
}

func (h *handler) syncAuditActionItemInAudit(db *gorm.DB, auditCycle *model.AuditCycle, auditID model.UUID) error {
	l := h.logger.Fields(logger.Fields{
		"method":       "syncAuditActionItemInAudit",
		"auditID":      auditID.String(),
		"auditCycleID": auditCycle.ID.String(),
	})

	l.Infof("Sync Audit Action Item for audit: %s", auditID.String())
	audit, err := h.store.Audit.One(db, auditID.String())
	if err != nil {
		l.Error(err, "failed to get audit from database")
		return err
	}

	// call api to get data of the audit from notion
	checklistPage, err := h.service.Notion.GetBlockChildren(audit.ID.String())
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

	checklistDatabase, err := h.service.Notion.GetDatabase(checklistPage.Results[checklistDatabaseIndex].ID(), nil, nil, 0)
	if err != nil {
		l.Error(err, "failed to get database from notion")
		return err
	}

	// Get all audit action item in the database
	auditActionItem, err := h.store.AuditActionItem.AllByAuditID(db, auditID.String())
	if err != nil {
		l.Error(err, "failed to get audit action item from database")
		return err
	}

	aaiMap := model.AuditActionItemToMap(auditActionItem)

	// Loop through all audit items and update to database if it have value in Audit action items field
	for _, checklist := range checklistDatabase.Results {
		checklistProperties := checklist.Properties.(notion.DatabasePageProperties)
		// Get audit item
		auditItem, err := h.store.AuditItem.One(db, checklist.ID)
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
			if _, err = h.store.AuditItem.UpdateSelectedFieldsByID(db, auditItem.ID.String(), *auditItem, "action_item_id"); err != nil {
				l.Error(err, "failed to update audit item")
				return err
			}
		} else if len(checklistProperties["☎️ Audit Action Items"].Relation) > 0 {
			// update audit_items(action_item_id)->audit(action_item)-> create record trong audit_action_items
			// Get action item
			actionItem, err := h.store.ActionItem.One(db, checklistProperties["☎️ Audit Action Items"].Relation[0].ID)
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
				if _, err = h.store.AuditItem.UpdateSelectedFieldsByID(db, auditItem.ID.String(), *auditItem, "action_item_id"); err != nil {
					l.Error(err, "failed to update audit item")
					return err
				}
			}

			// Update audit
			audit.ActionItem++
			if audit, err = h.store.Audit.Update(db, audit); err != nil {
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
				if _, err = h.store.AuditActionItem.Create(db, auditActionItem); err != nil {
					l.Error(err, "failed to create audit action item")
					return err
				}
			}
		}
	}

	// Delete audit action item
	for _, aai := range aaiMap {
		if err := h.store.AuditActionItem.Delete(db, aai.ID.String()); err != nil {
			l.Error(err, "failed to delete audit action item")
			return err
		}
	}

	return nil
}

func (h *handler) snapShot(db *gorm.DB) error {
	l := h.logger.Fields(logger.Fields{
		"method": "snapShot",
	})

	// Get all audit cycle from database
	auditCycles, err := h.store.AuditCycle.All(db)
	if err != nil {
		l.Error(err, "failed to get audit cycle from database")
		return err
	}

	// Create action item snapshot record for each audit cycle
	for _, auditCycle := range auditCycles {
		projectNotion, err := h.store.ProjectNotion.OneByAuditNotionID(db, auditCycle.ProjectID.String())
		if err != nil {
			l.Error(err, "failed to get project notion from database")
			return err
		}

		if projectNotion.Project.Status != model.ProjectStatusClosed && projectNotion.Project.Status != model.ProjectStatusPaused {
			actionItemSnapshot := &model.ActionItemSnapshot{
				ProjectID:    auditCycle.ProjectID,
				AuditCycleID: auditCycle.ID,
				High:         auditCycle.ActionItemHigh,
				Medium:       auditCycle.ActionItemMedium,
				Low:          auditCycle.ActionItemLow,
			}

			// Check record exists in database
			today := time.Now().Format("2006-01-02")

			if snapShot, err := h.store.ActionItemSnapshot.OneByAuditCycleIDAndTime(db, auditCycle.ID.String(), today); err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					if _, err := h.store.ActionItemSnapshot.Create(db, actionItemSnapshot); err != nil {
						l.Error(err, "failed to create action item snapshot")
						return err
					}
				} else {
					l.Error(err, "failed to get action item snapshot from database")
					return err
				}
			} else {
				// Update if record exists
				if !model.CompareActionItemSnapshot(snapShot, actionItemSnapshot) {
					if _, err := h.store.ActionItemSnapshot.UpdateSelectedFieldsByID(db, snapShot.ID.String(), *actionItemSnapshot, "high", "medium", "low"); err != nil {
						l.Error(err, "failed to update action item snapshot")
						return err
					}
				}
			}
		}
	}

	return nil
}
