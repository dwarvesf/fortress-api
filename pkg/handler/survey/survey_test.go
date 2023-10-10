package survey

import (
	"bytes"
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
	"github.com/dwarvesf/fortress-api/pkg/handler/survey/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk1ODMzMzA5NDUsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIn0.oIdlwWGBy4E1CbSoEX6r2B6NQLbew_J-RttpAcg6w8M"

func TestHandler_ListSurvey(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()

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
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/list_survey/list_survey.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys?%s", tt.query), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.Request.URL.RawQuery = tt.query

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.ListSurvey(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.ListSurvey] response mismatched")
			})
		})
	}
}

func TestHandler_GetSurveyDetail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
			name:             "ok_work",
			id:               "d97ee823-f7d5-418b-b281-711cb1d8e947",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_survey_detail/200_work.json",
		},
		{
			name:             "event_id_not_found",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce123",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_survey_detail/404_event_id_not_found.json",
		},
		{
			name:             "empty_event_id",
			id:               "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_detail/invalid_event_id.json",
		},
		{
			name:             "invalid_event_id_format",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce1234",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_detail/invalid_event_id.json",
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
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.GetSurveyDetail] response mismatched")
			})
		})
	}
}

func TestHandler_SendPerformanceReview(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		wantCode         int
		wantResponsePath string
		body             request.SendSurveyInput
	}{
		{
			name:             "happy_case",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/send_survey/200_performance_review.json",
			body: request.SendSurveyInput{
				TopicIDs: []model.UUID{
					model.MustGetUUIDFromString("e4a33adc-2495-43cf-b816-32feb8d5250d"),
				},
				Type: model.EventSubtypePeerReview.String(),
			},
		},
		{
			name:             "ok_send_engagement",
			id:               "53546ea4-1d9d-4216-96b2-75f84ec6d750",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/send_survey/200_send_engagement.json",
			body: request.SendSurveyInput{
				Type: model.EventSubtypeEngagement.String(),
			},
		},
		{
			name:             "not_found_case",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e1",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/send_survey/404.json",
			body: request.SendSurveyInput{
				TopicIDs: []model.UUID{
					model.MustGetUUIDFromString("e4a33adc-2495-43cf-b816-32feb8d5250d"),
				},
				Type: model.EventSubtypePeerReview.String(),
			},
		},
		{
			name:             "empty_event_id",
			id:               "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/send_survey/invalid_event_id.json",
		},
		{
			name:             "invalid_event_id_format",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce1234",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/send_survey/invalid_event_id.json",
		},
		{
			name:             "invalid_subtype",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/send_survey/invalid_subtype.json",
			body: request.SendSurveyInput{
				TopicIDs: []model.UUID{
					model.MustGetUUIDFromString("e4a33adc-2495-43cf-b816-32feb8d5250d"),
				},
				Type: "a",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/send_survey/send_survey.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				byteReq, _ := json.Marshal(tt.body)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys/%s/send", tt.id), bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.SendSurvey(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.SendSurvey] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteSurvey(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
		{
			name:             "empty_event_id",
			id:               "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/invalid_event_id.json",
		},
		{
			name:             "invalid_event_id_format",
			id:               "8a5bfedb-6e11-4f5c-82d9-2635cfcce1234",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/invalid_event_id.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_survey/delete_survey.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/surveys/%s", tt.id), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.id)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeleteSurvey(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.DeleteSurvey] response mismatched")
			})
		})
	}
}

func TestHandler_GetPeerReviewDetail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
		{
			name:             "empty_feedback_id",
			feedbackID:       "",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_peer_review_detail/invalid_feedback_id.json",
		},
		{
			name:             "invalid_feedback_id_format",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e23",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_peer_review_detail/invalid_feedback_id.json",
		},
		{
			name:             "empty_topic_id",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_peer_review_detail/invalid_topic_id.json",
		},
		{
			name:             "invalid_topic_id_format",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250de",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_peer_review_detail/invalid_topic_id.json",
		},
		{
			name:             "topic_not_found",
			feedbackID:       "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_peer_review_detail/topic_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_peer_review_detail/get_peer_review_detail.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Params = gin.Params{gin.Param{Key: "id", Value: tt.feedbackID}, gin.Param{Key: "topicID", Value: tt.topicID}}
				ctx.Request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/surveys/%s/topics/%s", tt.feedbackID, tt.topicID), nil)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetSurveyTopicDetail(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.GetSurveyTopicDetail] response mismatched")
			})
		})
	}
}

func TestHandler_UpdateTopicReviewers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
		{
			name: "empty_event_id",
			input: request.UpdateTopicReviewersInput{
				EventID: "",
				TopicID: "e4a33adc-2495-43cf-b816-32feb8d5250d",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_topic_participants/invalid_event_id.json",
		},
		{
			name: "invalid_event_id_format",
			input: request.UpdateTopicReviewersInput{
				EventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e23",
				TopicID: "e4a33adc-2495-43cf-b816-32feb8d5250d",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_topic_participants/invalid_event_id.json",
		},
		{
			name: "empty_topic_id",
			input: request.UpdateTopicReviewersInput{
				EventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
				TopicID: "",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_topic_participants/invalid_topic_id.json",
		},
		{
			name: "invalid_topic_id_format",
			input: request.UpdateTopicReviewersInput{
				EventID: "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
				TopicID: "e4a33adc-2495-43cf-b816-32feb8d5250de",
			},
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/update_topic_participants/invalid_topic_id.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/update_topic_participants/update_topic_participants.sql")
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
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.UpdateTopicReviewers] response mismatched")
			})
		})
	}
}

func TestHandler_MarkDone(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
			name:             "not_found",
			eventID:          "9b3480be-86a2-4ff9-84d8-545a4146122a",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/mark_done/404.json",
		},
		{
			name:             "empty_event_id",
			eventID:          "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/mark_done/invalid_event_id.json",
		},
		{
			name:             "invalid_event_id_format",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e23",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/mark_done/invalid_event_id.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/mark_done/mark_done.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/surveys/%s/done", tt.eventID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.eventID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.MarkDone(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.MarkDone] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteTopicReviewers(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_topic_participants/delete_topic_participants.sql")
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
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.DeleteTopicReviewers] response mismatched")
			})
		})
	}
}

func TestHandler_DeleteSurveyTopic(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
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
		{
			name:             "empty_event_id",
			eventID:          "",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/invalid_event_id.json",
		},
		{
			name:             "invalid_event_id_format",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e23",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/invalid_event_id.json",
		},
		{
			name:             "empty_topic_id",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/invalid_topic_id.json",
		},
		{
			name:             "invalid_topic_id_format",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250de",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/delete_survey/invalid_topic_id.json",
		},
		{
			name:             "topic_not_found",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/delete_survey/topic_not_found.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/delete_survey/delete_survey.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/surveys/%s/topics/%s", tt.eventID, tt.topicID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.eventID)
				ctx.AddParam("topicID", tt.topicID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.DeleteSurveyTopic(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.DeleteSurveyTopic] response mismatched")
			})
		})
	}
}

func TestHandler_GetSurveyReviewDetail(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()

	tests := []struct {
		name             string
		eventID          string
		topicID          string
		reviewerID       string
		wantCode         int
		wantResponsePath string
	}{
		{
			name:             "valid_id",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e1",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250e",
			reviewerID:       "bc9a5715-9723-4a2f-ad42-0d0f19a80b4d",
			wantCode:         http.StatusNotFound,
			wantResponsePath: "testdata/get_survey_review_detail/404_not_found.json",
		},
		{
			name:             "happy_case",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			reviewerID:       "bc9a5715-9723-4a2f-ad42-0d0f19a80b4d",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_survey_review_detail/200_happy_case.json",
		},
		{
			name:             "ok_work",
			eventID:          "d97ee823-f7d5-418b-b281-711cb1d8e947",
			topicID:          "9cf93fc1-5a38-4e2a-87de-41634b65fc87",
			reviewerID:       "789f1163-f157-4df3-9764-8100277cacba",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/get_survey_review_detail/200_work.json",
		},
		{
			name:             "empty_event_id",
			eventID:          "",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			reviewerID:       "bc9a5715-9723-4a2f-ad42-0d0f19a80b4d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_review_detail/invalid_event_id.json",
		},
		{
			name:             "invalid_event_id_format",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e23",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			reviewerID:       "bc9a5715-9723-4a2f-ad42-0d0f19a80b4d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_review_detail/invalid_event_id.json",
		},
		{
			name:             "empty_topic_id",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "",
			reviewerID:       "bc9a5715-9723-4a2f-ad42-0d0f19a80b4d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_review_detail/invalid_topic_id.json",
		},
		{
			name:             "invalid_topic_id_format",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250de",
			reviewerID:       "bc9a5715-9723-4a2f-ad42-0d0f19a80b4d",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_review_detail/invalid_topic_id.json",
		},
		{
			name:             "empty_reviewer_id",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			reviewerID:       "",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_review_detail/invalid_review_id.json",
		},
		{
			name:             "empty_reviewer_id",
			eventID:          "8a5bfedb-6e11-4f5c-82d9-2635cfcce3e2",
			topicID:          "e4a33adc-2495-43cf-b816-32feb8d5250d",
			reviewerID:       "e4a33adc-2495-43cf-b816-32feb8d5250de",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/get_survey_review_detail/invalid_review_id.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/get_survey_review_detail/get_survey_review_detail.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				ctx.Request = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/surveys/%s/topics/%s/reviews/%s", tt.eventID, tt.topicID, tt.reviewerID), nil)
				ctx.Request.Header.Set("Authorization", testToken)
				ctx.AddParam("id", tt.eventID)
				ctx.AddParam("topicID", tt.topicID)
				ctx.AddParam("reviewID", tt.reviewerID)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.GetSurveyReviewDetail(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.GetSurveyReviewDetail] response mismatched")
			})
		})
	}
}

func TestHandler_CreateSurvey(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg, nil, nil)
	storeMock := store.New()

	tests := []struct {
		name             string
		id               string
		wantCode         int
		wantResponsePath string
		body             request.CreateSurveyFeedbackInput
	}{
		{
			name:             "happy_case",
			wantCode:         http.StatusOK,
			wantResponsePath: "testdata/create_survey/200_work.json",
			body: request.CreateSurveyFeedbackInput{
				Quarter:  "q3,q4",
				Year:     2023,
				Type:     "peer-review",
				FromDate: "2023-11-28",
				ToDate:   "2023-11-29",
			},
		},
		{
			name:             "invalid_range_date",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_survey/400.json",
			body: request.CreateSurveyFeedbackInput{
				Quarter:  "q3,q4",
				Year:     2023,
				Type:     "work",
				FromDate: "2023-11-30",
				ToDate:   "2023-11-29",
			},
		},
		{
			name:             "invalid_subtype",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_survey/invalid_subtype.json",
			body: request.CreateSurveyFeedbackInput{
				Quarter:  "q3,q4",
				Year:     2023,
				Type:     "peer-revieww",
				FromDate: "2023-11-28",
				ToDate:   "2023-11-29",
			},
		},
		{
			name:             "invalid_date",
			wantCode:         http.StatusBadRequest,
			wantResponsePath: "testdata/create_survey/invalid_date.json",
			body: request.CreateSurveyFeedbackInput{
				Quarter:  "q3,q4",
				Year:     2023,
				Type:     "work",
				FromDate: "2023-13-28",
				ToDate:   "2023-11-29",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				testhelper.LoadTestSQLFile(t, txRepo, "./testdata/create_survey/create_survey.sql")
				w := httptest.NewRecorder()
				ctx, _ := gin.CreateTestContext(w)
				byteReq, _ := json.Marshal(tt.body)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/surveys", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)

				h := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)
				h.CreateSurvey(ctx)
				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := os.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.Survey.SendSurvey] response mismatched")
			})
		})
	}
}
