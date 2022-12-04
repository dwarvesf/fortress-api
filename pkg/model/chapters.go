package model

type Chapter struct {
	BaseModel

	Name   string `json:"name"`
	Code   string `json:"code"`
	LeadID *UUID  `json:"lead_id"`
}

// ToChapterMap create map from chapters
func ToChapterMap(chapters []*Chapter) map[UUID]string {
	rs := map[UUID]string{}
	for _, s := range chapters {
		rs[s.ID] = s.Name
	}

	return rs
}
