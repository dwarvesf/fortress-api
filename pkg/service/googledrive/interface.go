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

	// Contractor invoice operations
	UploadContractorInvoicePDF(contractorName, fileName string, pdfBytes []byte) (string, error)

	// FindContractorInvoiceFileID finds a contractor invoice PDF in Google Drive by contractor name and invoice ID
	// Returns the file ID if found, empty string if not found
	FindContractorInvoiceFileID(contractorName, invoiceID string) (string, error)

	// ShareFileWithEmail shares a Google Drive file with the specified email address
	// Google automatically sends a notification email to the recipient
	ShareFileWithEmail(fileID, email string) error
}
