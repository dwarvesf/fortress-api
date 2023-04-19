package view

type ContentData struct {
	Url string `json:"url"`
}

type ContentDataResponse struct {
	Data *ContentData `json:"data"`
}

func ToContentData(url string) *ContentData {
	return &ContentData{
		Url: url,
	}
}
