package model

type InvoiceItem struct {
	BaseModel

	InvoiceID       UUID
	ProjectMemberID UUID
	Total           int64
	Subtotal        int64
	Discount        int64
	Tax             float64
	Description     string
	Type            string
	IsExternal      bool

	ProjectMember *ProjectMember
}

type InvoiceItemRendered struct {
	Description string
	Quantity    float64
	UnitCost    int64
	Discount    int64
	Cost        int64
	IsExternal  bool
}
