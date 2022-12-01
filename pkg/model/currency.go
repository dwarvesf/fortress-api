package model

import "strconv"

const (
	VNDCurrency = "VND"
)

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
	return formatNumber(int64(vnd))
}

func (vnd *VietnamDong) format() {
	*vnd = NewVietnamDong(int64(*vnd) / 1000 * 1000) // rounded
}

func (vnd *VietnamDong) Format() VietnamDong {
	vnd.format()
	return *vnd
}

func formatNumber(n int64) string {
	in := strconv.FormatInt(n, 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}
