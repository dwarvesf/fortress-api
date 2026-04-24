package notion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	nt "github.com/dstotijn/go-notion"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	notionsvc "github.com/dwarvesf/fortress-api/pkg/service/notion"
	mocknotion "github.com/dwarvesf/fortress-api/pkg/service/notion/mocks"
)

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
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/cronjobs/clone-monthly-invoices?dryRun=true&status=Overdue&status=Draft", nil)

	h.CloneMonthlyInvoices(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
}
