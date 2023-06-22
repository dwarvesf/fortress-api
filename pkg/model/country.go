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

func (c Cities) GetCity(city string) *City {
	for _, itm := range c {
		if itm.Name == city {
			return &itm
		}
	}

	return nil
}
