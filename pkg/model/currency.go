package model

import "github.com/dwarvesf/fortress-api/pkg/utils"

type Currency struct {
	BaseModel

	Name   string
	Symbol string
	Locale string
	Type   string
}

type VietnamDong int64

func (vnd *VietnamDong) Scan(b interface{}) error {
	if b == nil {
		return nil
	}
	*vnd = VietnamDong(b.(int64))
	vnd.format()
	return nil
}

func NewVietnamDong(i int64) VietnamDong {
	return VietnamDong(i)
}

func (vnd VietnamDong) String() string {
	return utils.FormatNumber(int64(vnd))
}

func (vnd *VietnamDong) format() {
	*vnd = NewVietnamDong(int64(*vnd) / 1000 * 1000)
}

func (vnd *VietnamDong) Format() VietnamDong {
	vnd.format()
	return *vnd
}
