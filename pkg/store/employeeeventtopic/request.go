package employeeeventtopic

type GetByEmployeeIDInput struct {
	Status string
}

type GetByEventIDInput struct {
	EventID string
	Keyword string
	Preload bool
	Paging  bool
}
