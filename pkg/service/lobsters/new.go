package lobsters

type service struct {
}

func New() IService {
	return &service{}
}
