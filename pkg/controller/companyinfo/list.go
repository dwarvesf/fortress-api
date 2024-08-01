package companyinfo

import (
	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func (r *controller) List(c *gin.Context) ([]*model.CompanyInfo, error) {
	companyInfos, err := r.store.CompanyInfo.All(r.repo.DB())
	if err != nil {
		return nil, err
	}

	return companyInfos, nil
}
