package mw

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestWithAuth(t *testing.T) {
	type args struct {
		testURL          string
		testTokenType    string
		testAccessToken  string
		expectedHTTPCode int
		expectedError    error
	}
	tcs := map[string]args{
		"authorization header valid": {
			testURL:          "/sample-routes",
			expectedHTTPCode: http.StatusOK,
			testTokenType:    "Bearer",
			testAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzMjExNDIsImlkIjoiMjY1NTgzMmUtZjAwOS00YjczLWE1MzUtNjRjM2EyMmU1NThmIiwiYXZhdGFyIjoiaHR0cHM6Ly9zMy1hcC1zb3V0aGVhc3QtMS5hbWF6b25hd3MuY29tL2ZvcnRyZXNzLWltYWdlcy81MTUzNTc0Njk1NjYzOTU1OTQ0LnBuZyIsImVtYWlsIjoidGhhbmhAZC5mb3VuZGF0aW9uIiwicGVybWlzc2lvbnMiOlsiZW1wbG95ZWVzLnJlYWQiXSwidXNlcl9pbmZvIjpudWxsfQ.GENGPEucSUrILN6tHDKxLMtj0M0REVMUPC7-XhDMpGM",
		},
		"authorization header invalid": {
			testURL:          "/sample-routes",
			testTokenType:    "",
			testAccessToken:  "",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    ErrAuthenticationTypeHeaderInvalid,
		},
		"authorization header invalid - missing token type": {
			testURL:          "/sample-routes",
			testTokenType:    "",
			testAccessToken:  "access_token",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    ErrAuthenticationTypeHeaderInvalid,
		},
		"authorization header invalid - none token type": {
			testURL:          "/sample-routes",
			testTokenType:    "Bearerr",
			testAccessToken:  "access_token",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    ErrAuthenticationTypeHeaderInvalid,
		},
		"invalid access token - invalid length of segments": {
			testURL:          "/sample-routes",
			testTokenType:    "Bearer",
			testAccessToken:  "access_token",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    errors.New("token contains an invalid number of segments"),
		},
		"invalid access token - invalid signature": {
			testURL:          "/sample-routes",
			testTokenType:    "Bearer",
			testAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTkzNTM4OTcsIlVzZXJJRCI6IjI2NTU4MzJlLWYwMDktNGI3My1hNTM1LTY0YzNhMjJlNTU4ZiIsImF2YXRhciI6Imh0dHBzOi8vczMtYXAtc291dGhlYXN0LTEuYW1hem9uYXdzLmNvbS9mb3J0cmVzcy1pbWFnZXMvNTE1MzU3NDY5NTY2Mzk1NTk0NC5wbmciLCJlbWFpbCI6InRoYW5oQGQuZm91bmRhdGlvbiIsInBlcm1pc3Npb25zIjpbImVtcGxveWVlcy5yZWFkIl0sInVzZXJfaW5mbyI6bnVsbH0.WoIAHchh9H6tEClULpJBPB0zmkZEOgtoWBEVlTzHZbc",
			expectedHTTPCode: http.StatusUnauthorized,
			expectedError:    errors.New("signature is invalid"),
		},
		"invalid access token - expired": {
			testURL:          "/sample-routes",
			expectedHTTPCode: http.StatusUnauthorized,
			testTokenType:    "Bearer",
			testAccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Njc4MTc4MzEsIlVzZXJJRCI6IjI2NTU4MzJlLWYwMDktNGI3My1hNTM1LTY0YzNhMjJlNTU4ZiIsImF2YXRhciI6Imh0dHBzOi8vczMtYXAtc291dGhlYXN0LTEuYW1hem9uYXdzLmNvbS9mb3J0cmVzcy1pbWFnZXMvNTE1MzU3NDY5NTY2Mzk1NTk0NC5wbmciLCJlbWFpbCI6InRoYW5oQGQuZm91bmRhdGlvbiIsInBlcm1pc3Npb25zIjpbImVtcGxveWVlcy5yZWFkIl0sInVzZXJfaW5mbyI6bnVsbH0.GLzCC6dcHRjPFGm_CQHzrD3nmSsKrqsN6Yq6BYzNRbk",
			expectedError:    errors.New("token is expired"),
		},
	}
	for desc, tc := range tcs {
		t.Run(desc, func(t *testing.T) {
			r := prepareTestDefaultRoutes()
			req, _ := http.NewRequest("GET", tc.testURL, nil)
			req.Header.Set("Authorization", fmt.Sprintf("%s %s", tc.testTokenType, tc.testAccessToken))

			// Create a response recorder
			w := httptest.NewRecorder()

			// Create the service and process the above request.
			r.ServeHTTP(w, req)
			if tc.expectedError != nil {
				require.Contains(t, w.Body.String(), tc.expectedError.Error())
			}
			require.Equal(t, tc.expectedHTTPCode, w.Code)
		})
	}
}

func prepareTestDefaultRoutes() *gin.Engine {
	r := gin.Default()
	r.GET("/sample-routes", WithAuth)

	return r
}
