package model

type AuditActionItem struct {
	BaseModel

	AuditID      UUID
	ActionItemID UUID
}

type AuditAction struct {
	AuditID      UUID
	ActionItemID UUID
}

func AuditActionItemToMap(aais []*AuditActionItem) map[AuditAction]AuditActionItem {
	rs := make(map[AuditAction]AuditActionItem)
	for _, aai := range aais {
		rs[AuditAction{AuditID: aai.AuditID, ActionItemID: aai.ActionItemID}] = *aai
	}

	return rs
}
