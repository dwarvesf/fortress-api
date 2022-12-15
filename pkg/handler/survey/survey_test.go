package survey

import (
	"bytes"
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
	"github.com/dwarvesf/fortress-api/pkg/handler/survey/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_ListSurvey(t *testing.T) {
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
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/list_survey/400_empty_query.json",
		},
		{
			name:             "get_peer_review",
			query:            "subtype=peer-review",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/list_survey/200_get_peer_review.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys?%s", tt.query), nil)
			ctx.Request.Header.Set("Authorization", testToken)
			ctx.Request.URL.RawQuery = tt.query

			h := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)
			h.ListSurvey(ctx)
			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.ListSurvey] response mismatched")
		})
	}
}

func TestHandler_GetSurveyDetail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_survey_detail/200_happy_case.json",
		},
		{
			name:             "ok_engagement",
			id:               "53546ea4-1d9d-4216-96b2-75f84ec6d750",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_survey_detail/200_engagement.json",
		},
		{
			name:             "event_id_not_found",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce123",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_survey_detail/404_event_id_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_survey_detail/get_survey_detail.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetSurveyDetail(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.GetSurveyDetail] response mismatched")
			})
		})
	}
}

func TestHandler_SendPerformanceReview(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		wantCode         int
		wantResponsePath string
		body             request.SendPerformanceReviewInput
	}{
		{
			name:             "happy_case",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/send_performance_review/200.json",
			body: request.SendPerformanceReviewInput{
				Topics: []request.PerformanceReviewTopic{
					{
						TopicID: model.MustGetUUIDFromString("e4a33adc-2495-43cf-b816-32feb8d5250d"),
						Participants: []model.UUID{
							model.MustGetUUIDFromString("061820c0-bf6c-4b4a-9753-875f75d71a2c"),
						},
					},
				},
			},
		},
		{
			name:             "not_found_case",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e1",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/send_performance_review/404.json",
			body: request.SendPerformanceReviewInput{
				Topics: []request.PerformanceReviewTopic{
					{
						TopicID: model.MustGetUUIDFromString("e4a33adc-2495-43cf-b816-32feb8d5250d"),
						Participants: []model.UUID{
							model.MustGetUUIDFromString("061820c0-bf6c-4b4a-9753-875f75d71a2c"),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				byteReq, _ := json.Marshal(tt.body)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys/%s/send", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.SendPerformanceReview(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.SendPerformanceReview] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteSurvey(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "valid_id",
			id:               "163fdda2-2dce-4618-9849-7c8475dcc9c1",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/delete_survey/200_valid_id.json",
		},
		{
			name:             "event_not_found",
			id:               "163fdda2-2dce-4618-9849-7c8475dcc123",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete_survey/404_event_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/surveys/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeleteSurvey(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.DeleteSurvey] response mismatched")
			})
		})
	}
}

func TestHandler_GetPeerReviewDetail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		feedbackID       string
		topicID          string
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_peer_review_detail/200.json",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
		},
		{
			name:             "failed_topic_not_found",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_peer_review_detail/404.json",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.feedbackID}, gin.Param{Key: "topicID", Value: tt.topicID}}
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys/%s/topics/%s", tt.feedbackID, tt.topicID), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetSurveyTopicDetail(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.GetSurveyTopicDetail] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateTopicReviewers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		input            request.UpdateTopicReviewersInput
		wantCode         int
		wantResponsePath string
	}{
		{
			name: "happy_case",
			input: request.UpdateTopicReviewersInput{
				EventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
				TopicID: "e4a33adc-2495-43cf-b816-32feb8d5250d",
				Body: request.UpdateTopicReviewersBody{
					ReviewerIDs: []model.UUID{
						model.MustGetUUIDFromString("d42a6fca-d3b8-4a48-80f7-a95772abda56"),
						model.MustGetUUIDFromString("dcfee24b-306d-4609-9c24-a4021639a11b"),
						model.MustGetUUIDFromString("3f705527-0455-4e67-a585-6c1f23726fff"),
						model.MustGetUUIDFromString("a1f25e3e-cf40-4d97-a4d5-c27ee566b8c5"),
						model.MustGetUUIDFromString("498d5805-dd64-4643-902d-95067d6e5ab5"),
						model.MustGetUUIDFromString("f6ce0d0f-5794-463b-ad0b-8240ab9c49be"),
						model.MustGetUUIDFromString("7bcf4b45-0279-4da2-84e4-eec5d9d05ba3"),
					},
				},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/update_topic_participants/200_happy_case.json",
		},
		{
			name: "participant_not_ready",
			input: request.UpdateTopicReviewersInput{
				EventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
				TopicID: "e4a33adc-2495-43cf-b816-32feb8d5250d",
				Body: request.UpdateTopicReviewersBody{
					ReviewerIDs: []model.UUID{
						model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
						model.MustGetUUIDFromString("2655832e-f009-4b73-a535-64c3a22e558f"),
						model.MustGetUUIDFromString("d389d35e-c548-42cf-9f29-2a599969a8f2"),
						model.MustGetUUIDFromString("f7c6016b-85b5-47f7-8027-23c2db482197"),
						model.MustGetUUIDFromString("d42a6fca-d3b8-4a48-80f7-a95772abda56"),
						model.MustGetUUIDFromString("dcfee24b-306d-4609-9c24-a4021639a11b"),
					},
				},
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_topic_participants/400_participant_not_ready.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				body, err := json.Marshal(tt.input.Body)
				require.NoError(t, err)

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut,
					fmt.Sprintf("/api/v1/surveys/%s/topics/%s/employees", tt.input.EventID, tt.input.TopicID),
					bytes.NewBuffer(body))
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.input.EventID)
				ctx.AddParam("topicID", tt.input.TopicID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.UpdateTopicReviewers(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.UpdateTopicReviewers] response mismatched")
			})
		})
	}
}

func TestHandler_MarkDone(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantResponsePath string
		eventID          string
	}{
		{
			name:             "happy_case",
			eventID:          "9b3480be-86a2-4ff9-84d8-545a4146122b",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/mark_done/200.json",
		},
		{
			name:             "not_done_all",
			eventID:          "53546ea4-1d9d-4216-96b2-75f84ec6d750",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/mark_done/400.json",
		},
		{
			name:             "not_found",
			eventID:          "9b3480be-86a2-4ff9-84d8-545a4146122a",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/mark_done/404.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/surveys/%s/done", tt.eventID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.eventID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.MarkDone(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.MarkDone] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteTopicReviewers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		input            request.DeleteTopicReviewersInput
		wantCode         int
		wantResponsePath string
	}{
		{
			name: "happy_case",
			input: request.DeleteTopicReviewersInput{
				EventID: "53546ea4-1d9d-4216-96b2-75f84ec6d750",
				TopicID: "11121775-118f-4896-8246-d88023b22c7a",
				Body: request.DeleteTopicReviewersBody{
					ReviewerIDs: []model.UUID{
						model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98ce"),
					},
				},
			},
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/delete_topic_participants/200.json",
		},
		{
			name: "not_found_employee",
			input: request.DeleteTopicReviewersInput{
				EventID: "53546ea4-1d9d-4216-96b2-75f84ec6d750",
				TopicID: "11121775-118f-4896-8246-d88023b22c7a",
				Body: request.DeleteTopicReviewersBody{
					ReviewerIDs: []model.UUID{
						model.MustGetUUIDFromString("ecea9d15-05ba-4a4e-9787-54210e3b98cd"),
					},
				},
			},
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete_topic_participants/404.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				body, err := json.Marshal(tt.input.Body)
				require.NoError(t, err)

				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut,
					fmt.Sprintf("/api/v1/surveys/%s/topics/%s/employees", tt.input.EventID, tt.input.TopicID),
					bytes.NewBuffer(body))
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.input.EventID)
				ctx.AddParam("topicID", tt.input.TopicID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeleteTopicReviewers(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.DeleteTopicReviewers] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteSurveyTopic(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		eventID          string
		topicID          string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "valid_id",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e1",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/delete_survey/200_valid_id.json",
		},
		{
			name:             "event_not_found",
			eventID:          "163fdda2-2dce-4618-9849-7c8475dcc999",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d52999",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/404_event_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_survey/delete_survey_topic.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/surveys/%s/topics/%s", tt.eventID, tt.topicID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.eventID)
				ctx.AddParam("topicID", tt.topicID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeleteSurveyTopic(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.DeleteSurveyTopic] response mismatched")
			})
		})

	}
}
