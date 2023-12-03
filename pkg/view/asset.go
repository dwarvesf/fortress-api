package view

type ContentData struct {
	Url string `json:"url"`
} // @name ContentData

// ContentDataResponse represent the content data
type ContentDataResponse struct {
	Data *ContentData `json:"data"`
} // @name ContentDataResponse

func ToContentData(url string) *ContentData {
	return &ContentData{
		Url: url,
	}
}
