package feedback

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/feedback/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk1ODMzMzA5NDUsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIn0.oIdlwWGBy4E1CbSoEX6r2B6NQLbew_J-RttpAcg6w8M"

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger("info")
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name             string
		query            string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "empty_query",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list/200_empty_query.json",
		},
		{
			name:             "get_draft_feedbacks",
			wantCode:         http.StatusOK,
			query:            "status=draft",
			wantResponsePath: "testdata/list/200_get_draft_feedbacks.json",
		},
		{
			name:             "get_draft_feedbacks",
			wantCode:         http.StatusBadRequest,
			query:            "status=draftf",
			wantResponsePath: "testdata/list/invalid_status.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/list/list.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/feedbacks?%s", tt.query), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.List(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInSliceResponse(w.Body.Bytes(), "lastUpdated")
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Feedback.List] response mismatched")
			})
		})
	}
}

func TestHandler_Detail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger("info")
	serviceMock := service.NewForTest()
	storeMock := store.New()
	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		feedbackID       string
		topicID          string
	}{
		{
			name:             "ok_detail",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/detail/200.json",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
		},
		{
			name:             "ok_engagement",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/detail/200_engagement.json",
			feedbackID:       "53546ea4-1d9d-4216-96b2-75f84ec6d750",
			topicID:          "ebf376a6-3d11-4cea-b464-593103258838",
		},
		{
			name:             "ok_work",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/detail/200_work.json",
			feedbackID:       "d97ee823-f7d5-418b-b281-711cb1d8e947",
			topicID:          "9cf93fc1-5a38-4e2a-87de-41634b65fc87",
		},
		{
			name:             "failed_topic_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/detail/404.json",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/detail/detail.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.feedbackID}, gin.Param{Key: "topicID", Value: tt.topicID}}
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/feedbacks/%s/topics/%s", tt.feedbackID, tt.topicID), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Detail(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Feedback.Detail] response mismatched")
			})
		})
	}
}

func TestHandler_Submit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger("info")
	serviceMock := service.NewForTest()
	storeMock := store.New()
	tests := []struct {
		name             string
		body             request.SubmitFeedbackRequest
		wantCode         int
		wantResponsePath string
		topicID          string
		eventID          string
	}{
		{
			name:             "failed_unanswer_question",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/submit/400_unanswer_question.json",
			body: request.SubmitFeedbackRequest{
				Answers: []request.BasicEventQuestionRequest{
					{
						EventQuestionID: model.MustGetUUIDFromString("7a94c0f4-81cf-4736-8628-710e25cfc4e7"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("795219e6-67d8-4611-a5cb-38fb1057e4ee"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("40177889-8098-4cb7-931e-9cfe857e56f7"),
						Answer:          "",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("7886c936-fb48-4fc0-beb6-7a8c5d723b78"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("272a2524-efe3-4386-b463-a79143ef661e"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("86aac74b-4dc4-422b-ac70-735fe247eedf"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("faab80cf-9b4c-4deb-96b6-3d7a46cc5e7d"),
						Answer:          "ok",
					},
				},
				Status: model.EventReviewerStatusDone,
			},
			topicID: "e4a33adc-2495-43cf-b816-32feb8d5250d",
			eventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
		},
		{
			name:             "ok_draft",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/submit/200.json",
			body: request.SubmitFeedbackRequest{
				Answers: []request.BasicEventQuestionRequest{
					{
						EventQuestionID: model.MustGetUUIDFromString("7a94c0f4-81cf-4736-8628-710e25cfc4e7"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("795219e6-67d8-4611-a5cb-38fb1057e4ee"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("40177889-8098-4cb7-931e-9cfe857e56f7"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("7886c936-fb48-4fc0-beb6-7a8c5d723b78"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("272a2524-efe3-4386-b463-a79143ef661e"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("86aac74b-4dc4-422b-ac70-735fe247eedf"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("faab80cf-9b4c-4deb-96b6-3d7a46cc5e7d"),
						Answer:          "ok",
					},
				},
				Status: model.EventReviewerStatusDraft,
			},
			topicID: "e4a33adc-2495-43cf-b816-32feb8d5250d",
			eventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
		},
		{
			name:             "draft_not_found_topicID",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/submit/404.json",
			body: request.SubmitFeedbackRequest{
				Answers: []request.BasicEventQuestionRequest{
					{
						EventQuestionID: model.MustGetUUIDFromString("7a94c0f4-81cf-4736-8628-710e25cfc4e7"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("795219e6-67d8-4611-a5cb-38fb1057e4ee"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("40177889-8098-4cb7-931e-9cfe857e56f7"),
						Answer:          "",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("7886c936-fb48-4fc0-beb6-7a8c5d723b78"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("272a2524-efe3-4386-b463-a79143ef661e"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("86aac74b-4dc4-422b-ac70-735fe247eedf"),
						Answer:          "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("faab80cf-9b4c-4deb-96b6-3d7a46cc5e7d"),
						Answer:          "ok",
					},
				},
				Status: model.EventReviewerStatusDraft,
			},
			topicID: "e4a33adc-2495-43cf-b816-32feb8d5250e",
			eventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
		},
		{
			name:             "ok_draft_engagement",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/submit/200_engagement.json",
			body: request.SubmitFeedbackRequest{
				Answers: []request.BasicEventQuestionRequest{
					{
						EventQuestionID: model.MustGetUUIDFromString("a9b63a36-0134-4aa3-9a9a-edb5a1d52645"),
						Answer:          "agree",
						Note:            "ok",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("ebf8ab8b-8827-49fe-a61a-40b140eba903"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("fae6d117-38c9-4634-8e20-9c012fe0f222"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("eb8e2a6a-0f86-40fb-b447-972bd8a53d15"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("c2c3e368-27b0-4154-b2ff-10d1a9f29ba1"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("ba0c090c-eefc-4e1a-b29d-a024e68c57fe"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("f508b883-2294-4f50-84e7-f43a0be89bad"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("57e4bb72-3c10-4652-b2b6-0f26be8ba936"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("c6fde31e-59c3-4161-9180-490a4a591a17"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("06c7451f-e155-409b-aebd-5082f897040a"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("16c1285b-bb45-4093-bcec-5daa8e3eaac6"),
						Answer:          "agree",
						Note:            "ok",
					}, {
						EventQuestionID: model.MustGetUUIDFromString("c9722d74-1683-4c1f-8378-75584cd25753"),
						Answer:          "agree",
						Note:            "ok",
					},
				},
				Status: model.EventReviewerStatusDraft,
			},
			topicID: "ebf376a6-3d11-4cea-b464-593103258838",
			eventID: "53546ea4-1d9d-4216-96b2-75f84ec6d750",
		},
		{
			name:             "ok_work",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/submit/200_work.json",
			body: request.SubmitFeedbackRequest{
				Answers: []request.BasicEventQuestionRequest{
					{
						EventQuestionID: model.MustGetUUIDFromString("3784e437-c7d6-4142-9007-82a7f18f7d50"),
						Answer:          model.AgreementLevelMixed.String(),
						Note:            "nothing",
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("25ae4b9c-4d44-4946-8034-29ccec02a005"),
						Answer:          model.AgreementLevelMixed.String(),
					},
					{
						EventQuestionID: model.MustGetUUIDFromString("5c49dbcd-df16-4f04-bb98-a2dbb339e4d6"),
						Answer:          model.AgreementLevelMixed.String(),
						Note:            "ok",
					},
				},
				Status: model.EventReviewerStatusDraft,
			},
			topicID: "9cf93fc1-5a38-4e2a-87de-41634b65fc87",
			eventID: "d97ee823-f7d5-418b-b281-711cb1d8e947",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/submit/submit.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				byteReq, _ := json.Marshal(tt.body)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.eventID}, gin.Param{Key: "topicID", Value: tt.topicID}}

				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/feedbacks/%s/topics/%s/submit", tt.eventID, tt.topicID), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.Submit(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInSliceResponse(w.Body.Bytes(), "lastUpdated")
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Feedback.Draft] response mismatched")
			})
		})
	}
}

func TestHandler_UnreadInbox(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger("info")
	serviceMock := service.NewForTest()
	storeMock := store.New()
	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_unread_inbox/200_happy_case.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_unread_inbox/get_unread_inbox.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/feedbacks/unreads", nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.CountUnreadFeedback(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Feedback.CountUnreadFeedback] response mismatched")
			})
		})
	}
}
