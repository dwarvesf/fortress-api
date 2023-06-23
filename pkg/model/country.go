package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Country struct {
	BaseModel

	Name   string `json:"name"`
	Code   string `json:"code"`
	Cities Cities `json:"cities"`
}

type City struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Long string `json:"long"`
}

type Cities []City

func (j Cities) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *Cities) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch t := value.(type) {
	case []uint8:
		jsonData := value.([]uint8)
		if string(jsonData) == "null" {
			return nil
		}
		return json.Unmarshal(jsonData, j)
	default:
		return fmt.Errorf("could not scan type %T into json", t)
	}
}

func (j Cities) GetCity(city string) *City {
	for _, itm := range j {
		if itm.Name == city {
			return &itm
		}
	}

	return nil
}
