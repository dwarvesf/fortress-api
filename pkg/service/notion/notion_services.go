package notion

// Services holds all Notion-related services
type Services struct {
	IService           // Embedded general Notion service
	Timesheet          *TimesheetService
	TaskOrderLog       *TaskOrderLogService
	ContractorRates    *ContractorRatesService
	ContractorPayouts  *ContractorPayoutsService
	ContractorPayables *ContractorPayablesService
	RefundRequests     *RefundRequestsService
	InvoiceSplit       *InvoiceSplitService
}
