package attachment

type AttachmentService interface {
	Create(contentType string, fileName string, file []byte) (id string, err error)
}
