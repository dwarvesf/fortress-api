package notion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	notionsvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
)

type fakeNotionService struct {
	queryInvoicesByMonthFn func(year, month int, statuses []string, projectID string) ([]nt.Page, error)
}

func (f *fakeNotionService) GetBlock(pageID string) (nt.Block, error) { return nil, nil }

func (f *fakeNotionService) ToChangelogMJML(blocks []nt.Block, email model.Email) (string, error) {
	return "", nil
}

func (f *fakeNotionService) FindClientPageForChangelog(clientID string) (nt.Page, error) {
	return nt.Page{}, nil
}

func (f *fakeNotionService) GetDatabase(databaseID string, filter *nt.DatabaseQueryFilter, sorts []nt.DatabaseQuerySort, pageSize int) (*nt.DatabaseQueryResponse, error) {
	return &nt.DatabaseQueryResponse{}, nil
}

func (f *fakeNotionService) GetDatabaseWithStartCursor(databaseID string, startCursor string) (*nt.DatabaseQueryResponse, error) {
	return &nt.DatabaseQueryResponse{}, nil
}

func (f *fakeNotionService) GetBlockChildren(pageID string) (*nt.BlockChildrenResponse, error) {
	return &nt.BlockChildrenResponse{}, nil
}

func (f *fakeNotionService) GetPagePropByID(pageID, propID string, query *nt.PaginationQuery) (*nt.PagePropResponse, error) {
	return &nt.PagePropResponse{}, nil
}

func (f *fakeNotionService) GetProjectInDB(pageID string) (*nt.DatabasePageProperties, error) {
	props := nt.DatabasePageProperties{}
	return &props, nil
}

func (f *fakeNotionService) GetProjectsInDB(pageIDs []string, projectPageID string) (map[string]nt.DatabasePageProperties, error) {
	return map[string]nt.DatabasePageProperties{}, nil
}

func (f *fakeNotionService) GetPage(pageID string) (nt.Page, error) { return nt.Page{}, nil }

func (f *fakeNotionService) CreatePage() error { return nil }

func (f *fakeNotionService) CreateDatabaseRecord(databaseID string, properties map[string]interface{}) (string, error) {
	return "", nil
}

func (f *fakeNotionService) ListProjects() ([]model.NotionProject, error) { return nil, nil }

func (f *fakeNotionService) ListProjectsWithChangelog() ([]model.ProjectChangelogPage, error) {
	return nil, nil
}

func (f *fakeNotionService) QueryAudienceDatabase(audienceDBId, audience string) ([]nt.Page, error) {
	return nil, nil
}

func (f *fakeNotionService) GetProjectHeadEmails(pageID string) (string, string, string, error) {
	return "", "", "", nil
}

func (f *fakeNotionService) UploadFile(filename, contentType string, fileData []byte) (string, error) {
	return "", nil
}

func (f *fakeNotionService) UpdatePageProperties(pageID string, properties nt.UpdatePageParams) error {
	return nil
}

func (f *fakeNotionService) UpdatePagePropertiesWithFileUpload(pageID, propertyName, fileUploadID, filename string) error {
	return nil
}

func (f *fakeNotionService) QueryInvoices(filter *notionsvc.InvoiceFilter, pagination model.Pagination) ([]nt.Page, int64, error) {
	return nil, 0, nil
}

func (f *fakeNotionService) GetInvoiceLineItems(invoicePageID string) ([]nt.Page, error) { return nil, nil }

func (f *fakeNotionService) QueryClientInvoiceByNumber(invoiceNumber string) (*nt.Page, error) {
	return nil, nil
}

func (f *fakeNotionService) UpdateClientInvoiceStatus(pageID string, status string, paidDate *time.Time) error {
	return nil
}

func (f *fakeNotionService) UpdateLineItemsStatus(invoicePageID string, status string) error { return nil }

func (f *fakeNotionService) ExtractClientInvoiceData(page *nt.Page) (*model.Invoice, error) {
	return nil, nil
}

func (f *fakeNotionService) GetNotionInvoiceStatus(page *nt.Page) (string, error) { return "", nil }

func (f *fakeNotionService) QueryLineItemsWithCommissions(invoicePageID string) ([]notionsvc.LineItemCommissionData, error) {
	return nil, nil
}

func (f *fakeNotionService) IsSplitsGenerated(invoicePageID string) (bool, error) { return false, nil }

func (f *fakeNotionService) MarkSplitsGenerated(invoicePageID string) error { return nil }

func (f *fakeNotionService) MarkLineItemsSplitsGenerated(invoicePageID string) error { return nil }

func (f *fakeNotionService) QueryInvoicesByMonth(year, month int, statuses []string, projectID string) ([]nt.Page, error) {
	if f.queryInvoicesByMonthFn != nil {
		return f.queryInvoicesByMonthFn(year, month, statuses, projectID)
	}

	return nil, nil
}

func (f *fakeNotionService) CloneInvoiceToNextMonth(sourceInvoicePageID string, targetIssueDate time.Time) (*notionsvc.ClonedInvoiceResult, error) {
	return nil, nil
}

func (f *fakeNotionService) CheckInvoiceExistsForMonth(projectPageID string, year, month int) (bool, string, error) {
	return false, "", nil
}

func TestCloneMonthlyInvoices_DefaultStatusesIncludePaidAndSent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	var gotStatuses []string
	fakeNotion := &fakeNotionService{
		queryInvoicesByMonthFn: func(year, month int, statuses []string, projectID string) ([]nt.Page, error) {
			gotStatuses = append([]string(nil), statuses...)
			return []nt.Page{}, nil
		},
	}

	h := &handler{
		service: &service.Service{
			Notion: &notionsvc.Services{IService: fakeNotion},
		},
		logger: logger.NewLogrusLogger("debug"),
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/cronjobs/clone-monthly-invoices?dryRun=true", nil)
	ctx.Request = req

	h.CloneMonthlyInvoices(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []string{"Paid", "Sent"}, gotStatuses)
}

func TestCloneMonthlyInvoices_UsesExplicitStatusesWithoutDefaults(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	var gotStatuses []string
	fakeNotion := &fakeNotionService{
		queryInvoicesByMonthFn: func(year, month int, statuses []string, projectID string) ([]nt.Page, error) {
			gotStatuses = append([]string(nil), statuses...)
			return []nt.Page{}, nil
		},
	}

	h := &handler{
		service: &service.Service{
			Notion: &notionsvc.Services{IService: fakeNotion},
		},
		logger: logger.NewLogrusLogger("debug"),
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/cronjobs/clone-monthly-invoices?dryRun=true&status=Overdue&status=Draft", nil)
	ctx.Request = req

	h.CloneMonthlyInvoices(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, []string{"Overdue", "Draft"}, gotStatuses)
}
