package notion

// Services holds all Notion-related services
type Services struct {
	IService                            // Embedded general Notion service
	Timesheet         *TimesheetService
	TaskOrderLog      *TaskOrderLogService
	ContractorRates   *ContractorRatesService
	ContractorFees    *ContractorFeesService
	ContractorPayouts   *ContractorPayoutsService
	ContractorPayables  *ContractorPayablesService
	RefundRequests      *RefundRequestsService
	InvoiceSplit      *InvoiceSplitService
}
