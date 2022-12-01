package profile

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/handler/profile/request"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
)

const testToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM"

func TestHandler_GetProfile(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()
	testRepoMock := store.NewPostgresStore(&cfg)

	tests := []struct {
		name             string
		wantCode         int
		wantErr          error
		wantResponsePath string
	}{
		{
			name:             "happy_case",
			wantCode:         200,
			wantErr:          nil,
			wantResponsePath: "testdata/get_profile/200.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/api/v1/profile", nil)
			ctx.Request.Header.Set("Authorization", testToken)
			metadataHandler := New(storeMock, testRepoMock, serviceMock, loggerMock, &cfg)

			metadataHandler.GetProfile(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
			require.NoError(t, err)

			require.JSONEq(t, string(expRespRaw), w.Body.String(), "[Handler.GetProfile] response mismatched")
		})
	}
}

func TestHandler_UpdateProfileInfo(t *testing.T) {
	// load env and test data
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger()
	serviceMock := service.New(&cfg)
	storeMock := store.New()

	tests := []struct {
		name             string
		wantCode         int
		wantErr          bool
		wantResponsePath string
		input            request.UpdateInfoInput
	}{
		{
			name:             "ok_get_profile",
			wantCode:         200,
			wantErr:          false,
			wantResponsePath: "testdata/update_info/200.json",
			input: request.UpdateInfoInput{
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "thanhpham123@gmail.com",
				PhoneNumber:   "0123456788",
			},
		},
		{
			name:             "invalid_phone_number",
			wantCode:         400,
			wantErr:          true,
			wantResponsePath: "testdata/update_info/400.json",
			input: request.UpdateInfoInput{
				TeamEmail:     "thanh@d.foundation",
				PersonalEmail: "thanhpham123@gmail.com",
				PhoneNumber:   "123456788",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				byteReq, err := json.Marshal(tt.input)
				require.Nil(t, err)
				w := httptest.NewRecorder()

				ctx, _ := gin.CreateTestContext(w)
				bodyReader := strings.NewReader(string(byteReq))
				ctx.Request = httptest.NewRequest("PUT", "/api/v1/profile", bodyReader)
				ctx.Request.Header.Set("Authorization", testToken)
				metadataHandler := New(storeMock, txRepo, serviceMock, loggerMock, &cfg)

				metadataHandler.UpdateInfo(ctx)

				require.Equal(t, tt.wantCode, w.Code)
				expRespRaw, err := ioutil.ReadFile(tt.wantResponsePath)
				require.NoError(t, err)

				res, err := utils.RemoveFieldInResponse(w.Body.Bytes(), "updatedAt")
				require.NoError(t, err)

				require.JSONEq(t, string(expRespRaw), string(res), "[Handler.Profile.UpdateInfo] response mismatched")
			})
		})
	}
}
