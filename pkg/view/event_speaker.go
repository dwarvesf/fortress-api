package view

import "github.com/dwarvesf/fortress-api/pkg/model"

// OgifStats contains list of ogif and some stats
type OgifStats struct {
	OgifList               []model.EventSpeaker `json:"ogifList"`
	UserAllTimeSpeaksCount int64                `json:"userAllTimeSpeaksCount"`
	UserAllTimeRank        int64                `json:"userAllTimeRank"`
	UserCurrentSpeaksCount int64                `json:"userCurrentSpeaksCount"`
	UserCurrentRank        int64                `json:"userCurrentRank"`
	TotalSpeakCount        int64                `json:"totalSpeakCount"`
	CurrentSpeakCount      int64                `json:"currentSpeakCount"`
}

// OgifStatsResponse return ogif stats response
type OgifStatsResponse struct {
	Data OgifStats `json:"data"`
} // @name OgifStatsResponse
