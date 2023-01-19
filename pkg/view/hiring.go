package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type HiringResponse struct {
	Data []model.HiringPosition `json:"data"`
}
