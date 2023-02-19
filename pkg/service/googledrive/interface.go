package googledrive

import "github.com/dwarvesf/fortress-api/pkg/model"

type Service interface {
	UploadInvoicePDF(invoice *model.Invoice, dirName string) error
	MoveInvoicePDF(invoice *model.Invoice, fromDirName, toDirName string) error
	DownloadInvoicePDF(invoice *model.Invoice, dirName string) ([]byte, error)
}
