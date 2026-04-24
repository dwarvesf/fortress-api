package notion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	notionsvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	mocknotion "github.com/dwarvesf/fortress-api/pkg/service/notion/mocks"
)

func makeInvoicePage(pageID, invoiceNumber, projectID string) nt.Page {
	return nt.Page{
		ID: pageID,
		Properties: nt.DatabasePageProperties{
			"Legacy Number": {
				RichText: []nt.RichText{{PlainText: invoiceNumber}},
			},
			"Project": {
				Relation: []nt.Relation{{ID: projectID}},
			},
		},
	}
}

func TestCloneMonthlyInvoices_DefaultStatusesIncludePaidAndSent(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notionMock := mocknotion.NewMockIService(ctrl)
	notionMock.EXPECT().
		QueryInvoicesByMonth(gomock.Any(), gomock.Any(), []string{"Paid", "Sent"}, "").
		Return([]nt.Page{}, nil)

	h := &handler{
		service: &service.Service{
			Notion: &notionsvc.Services{IService: notionMock},
		},
		logger: logger.NewLogrusLogger("debug"),
		config: &config.Config{},
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/cronjobs/clone-monthly-invoices?dryRun=true", nil)

	h.CloneMonthlyInvoices(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCloneMonthlyInvoices_UsesExplicitStatusesWithoutDefaults(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notionMock := mocknotion.NewMockIService(ctrl)
	notionMock.EXPECT().
		QueryInvoicesByMonth(gomock.Any(), gomock.Any(), []string{"Overdue", "Draft"}, "").
		Return([]nt.Page{}, nil)

	h := &handler{
		service: &service.Service{
			Notion: &notionsvc.Services{IService: notionMock},
		},
		logger: logger.NewLogrusLogger("debug"),
		config: &config.Config{},
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/cronjobs/clone-monthly-invoices?dryRun=true&status=Overdue&status=Draft", nil)

	h.CloneMonthlyInvoices(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestCloneMonthlyInvoices_OnlyProcessesLatestSelectedInvoicePerProject(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectID := "project-1"
	latestInvoice := makeInvoicePage("invoice-new", "INV-NEW", projectID)
	olderInvoice := makeInvoicePage("invoice-old", "INV-OLD", projectID)
	targetIssueDate := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.Local)

	notionMock := mocknotion.NewMockIService(ctrl)
	notionMock.EXPECT().
		QueryInvoicesByMonth(2026, 3, []string{"Paid", "Sent"}, "").
		Return([]nt.Page{latestInvoice, olderInvoice}, nil)
	notionMock.EXPECT().
		CheckInvoiceExistsForMonth(projectID, 2026, 4).
		Return(false, "", nil)
	notionMock.EXPECT().
		CloneInvoiceToNextMonth("invoice-new", targetIssueDate).
		Return(&notionsvc.ClonedInvoiceResult{NewInvoicePageID: "new-page", LineItemsCloned: 2}, nil)

	h := &handler{
		service: &service.Service{
			Notion: &notionsvc.Services{IService: notionMock},
		},
		logger: logger.NewLogrusLogger("debug"),
		config: &config.Config{},
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/cronjobs/clone-monthly-invoices?year=2026&month=3&targetYear=2026&targetMonth=4", nil)

	h.CloneMonthlyInvoices(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"cloned":1`)
	require.Contains(t, recorder.Body.String(), `"skipped":1`)
	require.Contains(t, recorder.Body.String(), `"another invoice for this project was already selected"`)
}
