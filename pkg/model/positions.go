package model

type Position struct {
	BaseModel

	Name string `json:"name"`
	Code string `json:"code"`
}

// ToPositionMap create map from position
func ToPositionMap(positions []*Position) map[UUID]Position {
	rs := map[UUID]Position{}
	for _, s := range positions {
		rs[s.ID] = *s
	}

	return rs
}

type Positions []Position

func (p Positions) IsSales() bool {
	for _, position := range p {
		if position.Code == "sales" {
			return true
		}
	}
	return false
}
