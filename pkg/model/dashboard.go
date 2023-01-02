package model

import "time"

type ResourceUtilization struct {
	Date      time.Time `json:"date"`
	Official  int       `json:"official"`
	Shadow    int       `json:"shadow"`
	Available int       `json:"available"`
}
