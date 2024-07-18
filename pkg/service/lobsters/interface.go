package lobsters

type IService interface {
	FetchNews(tag string) ([]LobsterPost, error)
}
