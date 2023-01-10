package model

import (
	"fmt"
	"regexp"
	"strings"
)

var maxPageSize int64 = 999

type SortOrder string

type Pagination struct {
	Page         int64  `json:"page" form:"page,default=0"`            // page index
	Size         int64  `json:"size" form:"size"`                      // page size
	Sort         string `json:"sort" form:"sort" swaggerignore:"true"` // sort by field
	Standardized bool   `json:"-" form:"-" swaggerignore:"true"`
}

const (
	SortOrderASC  SortOrder = "asc"
	SortOrderDESC SortOrder = "desc"
)

func (e SortOrder) IsValid() bool {
	switch e {
	case
		SortOrderASC,
		SortOrderDESC:
		return true
	}
	return false
}

func (e SortOrder) String() string {
	return string(e)
}

func (p *Pagination) Standardize() {
	if p.Standardized {
		return
	}

	if p.Page < 0 {
		p.Page = 0
	}

	if p.Size <= 0 || p.Size >= maxPageSize {
		p.Size = maxPageSize
	}

	if p.Sort == "" {
		return
	}

	p.Sort = standardizeSortQuery(p.Sort)
	p.Standardized = true
}

func (p *Pagination) ToLimitOffset() (limit int, offset int) {
	limit = int(p.Size)
	offset = limit * (int(p.Page) - 1)
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

func standardizeSortQuery(sortQ string) string {
	if sortQ == "" {
		return sortQ
	}

	f := func(c rune) bool {
		return c == ','
	}
	sorts := strings.FieldsFunc(sortQ, f)

	re, err := regexp.Compile(`[^\w|-]`)
	if err != nil {
		return ""
	}

	for i := range sorts {
		sort := re.ReplaceAllString(sorts[i], "")
		operator := "ASC"
		if sort[0] == '-' {
			operator = "DESC"
			sort = strings.Replace(sort, "-", "", 1)
		}
		sorts[i] = fmt.Sprintf("%s %s", sort, operator)
	}

	return strings.Join(sorts, ",")
}
