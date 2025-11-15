package basecamp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/service/taskprovider"
)

func TestParseAccountingWebhook_Basecamp(t *testing.T) {
	body := []byte(`{
		"creator": {"name": "Ops Bot"},
		"recording": {
			"id": 98765,
			"title": "Office Rental | 1.500.000 | VND"
		}
	}`)

	p := &Provider{}
	payload, err := p.ParseAccountingWebhook(
		context.TODO(),
		taskprovider.AccountingWebhookRequest{Body: body},
	)
	require.NoError(t, err)
	require.Equal(t, taskprovider.ProviderBasecamp, payload.Provider)
	require.Equal(t, "Office Rental | 1.500.000 | VND", payload.Title)
	require.Equal(t, float64(1500000), payload.Amount)
	require.Equal(t, "VND", payload.Currency)
	require.Equal(t, "completed", payload.Status)
	require.Equal(t, "98765", payload.TodoID)
}
