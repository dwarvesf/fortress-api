package view

import "github.com/dwarvesf/fortress-api/pkg/model"

// Earn represents memo earn
type Earn struct {
	Title    string   `json:"title"`
	Bounty   string   `json:"bounty"`
	Status   string   `json:"status"`
	PICs     []string `json:"pics"`
	Function string   `json:"function"`
	URL      string   `json:"url"`
} // @name Earn

type ListEarnResponse struct {
	Data []Earn `json:"earns"`
} // @name ListEarnResponse

func ToEarns(earns []model.Earn) []Earn {
	res := make([]Earn, 0)
	for _, earn := range earns {
		res = append(res, Earn{
			Title:    earn.Title,
			Bounty:   earn.Bounty,
			Status:   earn.Status,
			PICs:     earn.PICs,
			Function: earn.Function,
			URL:      earn.URL,
		})
	}

	return res
}
