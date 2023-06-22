package model

type Country struct {
	BaseModel

	Name   string `json:"name"`
	Code   string `json:"code"`
	Cities JSON   `json:"cities"`
}

type City struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

type Cities []City

func (c Cities) Contains(city string) bool {
	for _, itm := range c {
		if itm.Name == city {
			return true
		}
	}

	return false
}
