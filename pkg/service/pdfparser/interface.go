package pdfparser

// IService defines the interface for PDF parsing operations
type IService interface {
	// ExtractText extracts text content from a PDF byte slice
	ExtractText(pdfBytes []byte) (string, error)
}
