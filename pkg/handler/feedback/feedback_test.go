package feedback

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_List(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/feedbacks?%s", tt.query), nil)
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.URL.RawQuery = tt.query

			h := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)
			h.List(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			res, err := utils.RemoveFieldInSliceResponse(w.Body.Bytes(), "lastUpdated")
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Feedback.List] response mismatched")
		})
	}
}

func TestHandler_Detail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)
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
			name:             "failed_topic_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/detail/404.json",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.feedbackID}, gin.Param{Key: "topicID", Value: tt.topicID}}
			ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/feedbacks/%s/topics/%s", tt.feedbackID, tt.topicID), nil)
			ctx.Request.Header.Set("Authorization", testToken)

			h := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)
			h.Detail(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Feedback.Detail] response mismatched")
		})
	}
}

func TestHandler_Submit(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	tests := []struct {
		name             string
		body             request.SubmitBody
		wantCode         int
		wantResponsePath string
		topicID          string
		eventID          string
	}{
		{
			name:             "failed_unanswer_question",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/submit/400_unanswer_question.json",
			body: request.SubmitBody{
				Answers: []request.BasicEventQuestionInput{
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
			body: request.SubmitBody{
				Answers: []request.BasicEventQuestionInput{
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
			body: request.SubmitBody{
				Answers: []request.BasicEventQuestionInput{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {

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
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInSliceResponse(w.Body.Bytes(), "lastUpdated")
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Feedback.Draft] response mismatched")
			})
		})
	}
}
