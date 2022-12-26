package dashboard

type store struct{}

func New() IStore {
	return &store{}
}
