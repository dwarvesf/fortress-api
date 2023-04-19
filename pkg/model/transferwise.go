package model

// TWQuote defines a structure for quote request in transferwise
type TWQuote struct {
	SourceAmount float64 `json:"sourceAmount"`
	Fee          float64 `json:"fee"`
	Rate         float64 `json:"rate"`
}

type TWRate struct {
	Rate   float64 `json:"rate"`
	Source string  `json:"source"`
	Target string  `json:"target"`
}
