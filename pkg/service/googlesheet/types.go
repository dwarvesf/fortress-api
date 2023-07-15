package googlesheet

type SheetData struct {
	Range          string     `json:"range"`
	MajorDimension string     `json:"majorDimension"`
	Values         [][]string `json:"values"`
}

type DeliveryMetricRawData struct {
	Person        string `json:"person"`
	Weight        string `json:"weight"`
	Effort        string `json:"effort"`
	Effectiveness string `json:"effectiveness"`
	Date          string `json:"date"`
	Project       string `json:"project"`
	Email         string `json:"email"`
}
