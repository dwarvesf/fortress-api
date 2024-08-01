package ogifmemosummarizer

type IService interface {
	SummarizeOGIFMemo(youtubeURL string) (content string, err error)
}
