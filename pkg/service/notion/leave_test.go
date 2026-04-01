package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	nt "github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

func newLeaveTestService(client *nt.Client, contractorDBID string) *LeaveService {
	return &LeaveService{baseService: &baseService{
		client: client,
		cfg: &config.Config{
			LeaveIntegration: config.LeaveIntegration{
				Notion: config.LeaveNotionIntegration{
					ContractorDBID: contractorDBID,
				},
			},
		},
		logger: logger.NewLogrusLogger("debug"),
	}}
}

func stringPointer(value string) *string {
	return &value
}

func contractorResultPage(pageID, fullName, discordUsername, teamEmail, personalEmail string) nt.Page {
	return nt.Page{
		ID: pageID,
		Parent: nt.Parent{
			Type:       nt.ParentTypeDatabase,
			DatabaseID: "contractor-db-id",
		},
		Properties: nt.DatabasePageProperties{
			"Full Name": nt.DatabasePageProperty{
				Title: []nt.RichText{{PlainText: fullName, Text: &nt.Text{Content: fullName}}},
			},
			"Discord": nt.DatabasePageProperty{
				RichText: []nt.RichText{{PlainText: discordUsername, Text: &nt.Text{Content: discordUsername}}},
			},
			"Team Email": nt.DatabasePageProperty{
				Email: stringPointer(teamEmail),
			},
			"Personal Email": nt.DatabasePageProperty{
				Email: stringPointer(personalEmail),
			},
			"Status": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{Name: "Active"},
			},
		},
	}
}

func TestLookupContractorDetailsByEmail_UsesOrFilterAndParsesContractor(t *testing.T) {
	const contractorDBID = "contractor-db-id"
	const email = "contractor.personal@example.com"

	client := newNotionTestClient(t, func(r *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/databases/"+contractorDBID+"/query", r.URL.Path)

		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var query nt.DatabaseQuery
		require.NoError(t, json.Unmarshal(bodyBytes, &query))
		require.NotNil(t, query.Filter)
		require.Len(t, query.Filter.Or, 2)
		require.Equal(t, "Team Email", query.Filter.Or[0].Property)
		require.NotNil(t, query.Filter.Or[0].Email)
		require.Equal(t, email, query.Filter.Or[0].Email.Equals)
		require.Equal(t, "Personal Email", query.Filter.Or[1].Property)
		require.NotNil(t, query.Filter.Or[1].Email)
		require.Equal(t, email, query.Filter.Or[1].Email.Equals)
		require.Equal(t, 2, query.PageSize)

		resp := nt.DatabaseQueryResponse{
			Results: []nt.Page{contractorResultPage("contractor-page-id", "Jane Roe", "jane.roe", "jane@d.foundation", email)},
		}

		payload, err := json.Marshal(resp)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	service := newLeaveTestService(client, contractorDBID)

	contractor, err := service.LookupContractorDetailsByEmail(context.Background(), email)
	require.NoError(t, err)
	require.NotNil(t, contractor)
	require.Equal(t, "contractor-page-id", contractor.PageID)
	require.Equal(t, "Jane Roe", contractor.FullName)
	require.Equal(t, "jane.roe", contractor.DiscordUsername)
	require.Equal(t, "jane@d.foundation", contractor.TeamEmail)
	require.Equal(t, email, contractor.PersonalEmail)
	require.Equal(t, "Active", contractor.Status)
}

func TestGetContractorPageIDByEmail_NotFound(t *testing.T) {
	const contractorDBID = "contractor-db-id"
	const email = "missing@example.com"

	client := newNotionTestClient(t, func(r *http.Request) (*http.Response, error) {
		resp := nt.DatabaseQueryResponse{Results: []nt.Page{}}

		payload, err := json.Marshal(resp)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	service := newLeaveTestService(client, contractorDBID)

	pageID, err := service.GetContractorPageIDByEmail(context.Background(), email)
	require.Error(t, err)
	require.Empty(t, pageID)
	require.EqualError(t, err, "contractor not found for email: "+email)
}

func TestLookupContractorByEmail_EmptyEmail(t *testing.T) {
	service := newLeaveTestService(nil, "contractor-db-id")

	pageID, err := service.LookupContractorByEmail(context.Background(), "")
	require.NoError(t, err)
	require.Empty(t, pageID)
}
