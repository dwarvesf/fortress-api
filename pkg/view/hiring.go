package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type HiringResponse struct {
	Data []model.NotionHiringPosition `json:"data"`
}
