package attachment

type Service interface {
	Create(contentType string, fileName string, file []byte) (id string, err error)
}
