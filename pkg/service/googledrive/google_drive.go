package googledrive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

const FullDriveAccessScope = "https://www.googleapis.com/auth/drive"

type googleService struct {
	config    *oauth2.Config
	token     *oauth2.Token
	service   *drive.Service
	appConfig *config.Config
}

// New function return Google service
func New(config *oauth2.Config, appConfig *config.Config) Service {
	return &googleService{
		config:    config,
		appConfig: appConfig,
	}
}

func (g *googleService) UploadInvoicePDF(invoice *model.Invoice, dirName string) error {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	dir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), dirName)
	if err != nil {
		return err
	}

	_, err = g.newFile(fmt.Sprintf("#%s.pdf", invoice.Number), "application/pdf", bytes.NewReader(invoice.InvoiceFileContent), dir.Id)
	if err != nil {
		return err
	}

	return nil
}

func (g *googleService) MoveInvoicePDF(invoice *model.Invoice, fromDirName, toDirName string) error {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	fromDir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), fromDirName)
	if err != nil {
		return err
	}
	if fromDir == nil {
		return fmt.Errorf(`%v directory not found`, fromDirName)
	}

	toDir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), toDirName)
	if err != nil {
		return err
	}
	if toDir == nil {
		return fmt.Errorf(`%v directory not found`, toDirName)
	}

	invoicePdf, err := g.searchFile(fmt.Sprintf("#%s.pdf", invoice.Number), fromDir.Id, false)
	if err != nil {
		return err
	}
	if invoicePdf == nil {
		return ErrInvoicePDFNotFound
	}

	return g.updateInvoiceDir(invoicePdf.Id, fromDir.Id, toDir.Id)
}

func (g *googleService) updateInvoiceDir(fileID, oldDirID, newDirID string) error {
	_, err := g.service.Files.Update(fileID, nil).
		AddParents(newDirID).
		RemoveParents(oldDirID).
		Do()
	return err
}

func (g *googleService) findInvoiceDir(year, status string) (*drive.File, error) {
	yearDir, err := g.getDirID(year, g.appConfig.Invoice.DirID)
	if err != nil {
		return nil, err
	}

	return g.getDirID(status, yearDir.Id)
}

func (g *googleService) getDirID(dirName, parentDirID string) (*drive.File, error) {
	dir, err := g.searchFile(dirName, parentDirID, true)
	if err != nil {
		return nil, err
	}

	if dir != nil {
		return dir, nil
	}

	return g.newDir(dirName, parentDirID)
}

func (g *googleService) ensureToken(rToken string) error {
	token := &oauth2.Token{
		RefreshToken: rToken,
	}

	if !token.Valid() {
		tks := g.config.TokenSource(context.Background(), token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}
		g.token = tok
	}

	return nil
}

func (g *googleService) prepareService() error {
	client := g.config.Client(context.Background(), g.token)
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return errors.New("Get Drive Service Failed " + err.Error())
	}
	g.service = service
	return nil
}

func (g *googleService) searchFile(name, parentId string, isFolder bool) (*drive.File, error) {
	var folderQuery string
	if isFolder {
		folderQuery = "mimeType='application/vnd.google-apps.folder' and "
	}

	var parentQuery string
	if parentId != "root" {
		parentQuery = fmt.Sprintf("'%s' in parents and ", parentId)
	}
	r, err := g.service.Files.List().
		Q(parentQuery + folderQuery + fmt.Sprintf("name='%s'", name)).
		Fields("nextPageToken, files(id, name)").
		Do()
	if err != nil {
		return nil, err
	}
	if len(r.Files) == 0 {
		return nil, nil
	}

	return r.Files[0], nil
}

func (g *googleService) newDir(name string, parentId string) (*drive.File, error) {
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentId},
	}

	return g.service.Files.Create(d).Do()
}

func (g *googleService) newFile(name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}

	return g.service.Files.Create(f).Media(content).Do()
}

// func (g *googleService) deleteFile(id string) error {
// 	return g.service.Files.Delete(id).Do()
// }

// func getDrivePreviewLink(fileID string) string {
// 	return fmt.Sprintf(`https://drive.google.com/file/d/%s/view`, fileID)
// }

// func getFileIDFromLink(url string) string {
// 	s := strings.Replace(url, "https://drive.google.com/file/d/", "", 1)
// 	return strings.Replace(s, "/view", "", 1)
// }

func (g *googleService) DownloadInvoicePDF(invoice *model.Invoice, dirName string) ([]byte, error) {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return nil, err
	}

	if err := g.prepareService(); err != nil {
		return nil, err
	}

	dir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), dirName)
	if err != nil {
		return nil, err
	}

	f, err := g.searchFile(fmt.Sprintf("#%s.pdf", invoice.Number), dir.Id, false)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, fmt.Errorf(`file not found`)
	}

	resp, err := g.service.Files.Get(f.Id).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
