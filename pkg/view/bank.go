package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type ListBankResponse struct {
	Data []Bank `json:"data"`
} // @name ListBankResponse

type Bank struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Bin       string `json:"bin"`
	ShortName string `json:"shortName"`
	Logo      string `json:"logo"`
	SwiftCode string `json:"swiftCode"`
} // @name Bank

func ToBank(in *model.Bank) *Bank {
	return &Bank{
		ID:        in.ID.String(),
		Name:      in.Name,
		Code:      in.Code,
		Bin:       in.Bin,
		ShortName: in.ShortName,
		Logo:      in.Logo,
		SwiftCode: in.SwiftCode,
	}
}

func ToListBank(banks []*model.Bank) []Bank {
	res := make([]Bank, 0)
	for _, bank := range banks {
		b := ToBank(bank)
		res = append(res, *b)
	}

	return res
}
