package view

import "github.com/dwarvesf/fortress-api/pkg/model"

func ToCompanyInfos(companies []*model.CompanyInfo) []CompanyInfo {
	rs := make([]CompanyInfo, 0, len(companies))
	for _, com := range companies {
		c := ToCompanyInfo(com)
		if c != nil {
			rs = append(rs, *c)
		}
	}

	return rs
}

type GetListCompanyInfoResponse struct {
	Data []*CompanyInfo `json:"data"`
} // @GetListCompanyInfoResponse
