package googledrive

import (
	"errors"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

var (
	ErrInvoicePDFNotFound = errors.New("invoice pdf not found")
)

type IService interface {
	UploadInvoicePDF(invoice *model.Invoice, dirName string) error
	MoveInvoicePDF(invoice *model.Invoice, fromDirName, toDirName string) error
	DownloadInvoicePDF(invoice *model.Invoice, dirName string) ([]byte, error)
}
