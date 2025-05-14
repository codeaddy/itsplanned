package handlers_test

import (
	"encoding/json"
	"itsplanned/handlers"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGoogleOAuthURL(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		redirectURI  string
		expectedCode int
		validateFunc func(t *testing.T, response *api.GoogleOAuthURLResponse)
	}{
		{
			name:         "Get OAuth URL with default redirect URI",
			userID:       user.ID,
			redirectURI:  "",
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.GoogleOAuthURLResponse) {
				assert.Contains(t, response.URL, "https://accounts.google.com/o/oauth2/auth")
				assert.Contains(t, response.URL, "client_id=")
				assert.Contains(t, response.URL, "redirect_uri=")
				assert.Contains(t, response.URL, "response_type=code")
				assert.Contains(t, response.URL, "scope=")
				assert.Contains(t, response.URL, "state=randomState")
			},
		},
		{
			name:         "Get OAuth URL with custom redirect URI",
			userID:       user.ID,
			redirectURI:  "itsplanned://callback",
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.GoogleOAuthURLResponse) {
				assert.Contains(t, response.URL, "https://accounts.google.com/o/oauth2/auth")
				assert.Contains(t, response.URL, "redirect_uri=itsplanned%3A%2F%2Fcallback")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			req := httptest.NewRequest("GET", "/auth/google", nil)
			if tc.redirectURI != "" {
				q := req.URL.Query()
				q.Add("redirect_uri", tc.redirectURI)
				req.URL.RawQuery = q.Encode()
			}
			c.Request = req

			handlers.GetGoogleOAuthURL(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.GoogleOAuthURLResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestWebToAppRedirect(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)

	testCases := []struct {
		name         string
		userID       uint
		code         string
		state        string
		expectedCode int
		validateFunc func(t *testing.T, location string)
	}{
		{
			name:         "Web to app redirect with code and state",
			userID:       user.ID,
			code:         "test_code",
			state:        "test_state",
			expectedCode: http.StatusFound,
			validateFunc: func(t *testing.T, location string) {
				assert.Contains(t, location, "itsplanned://callback/auth")
				assert.Contains(t, location, "code=test_code")
				assert.Contains(t, location, "state=test_state")
			},
		},
		{
			name:         "Web to app redirect with code only",
			userID:       user.ID,
			code:         "test_code",
			state:        "",
			expectedCode: http.StatusFound,
			validateFunc: func(t *testing.T, location string) {
				assert.Contains(t, location, "itsplanned://callback/auth")
				assert.Contains(t, location, "code=test_code")
				assert.NotContains(t, location, "state=")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			req := httptest.NewRequest("GET", "/auth/web-to-app", nil)
			q := req.URL.Query()
			if tc.code != "" {
				q.Add("code", tc.code)
			}
			if tc.state != "" {
				q.Add("state", tc.state)
			}
			req.URL.RawQuery = q.Encode()
			c.Request = req

			handlers.WebToAppRedirect(c)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.validateFunc != nil {
				location := w.Header().Get("Location")
				tc.validateFunc(t, location)
			}
		})
	}
}
