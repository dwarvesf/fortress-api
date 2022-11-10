package utils

import (
	"encoding/json"
)

func RemoveFieldInResponse(data []byte, field string) ([]byte, error) {
	var res map[string]map[string]interface{}

	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if res["data"] != nil && res["data"][field] != nil {
		res["data"][field] = ""
	}

	return json.Marshal(res)
}

func RemoveFieldInSliceResponse(data []byte, field string) ([]byte, error) {
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	if res["data"] != nil {
		if resData, ok := res["data"].([]interface{}); ok {
			for i, v := range resData {
				if v, ok := v.(map[string]interface{}); ok {
					v[field] = ""
				}
				resData[i] = v
			}
			res["data"] = resData
		}
	}

	return json.Marshal(res)
}
