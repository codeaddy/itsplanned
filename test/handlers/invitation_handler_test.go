package handlers_test

import (
	"bytes"
	"encoding/json"
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/models/api"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInviteLink(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	participant := test.CreateTestUser(t)

	event := test.CreateTestEvent(t, organizer.ID)

	test.AddEventParticipant(t, event.ID, participant.ID)

	testCases := []struct {
		name         string
		userID       uint
		request      api.GenerateInviteLinkRequest
		expectedCode int
		validateFunc func(t *testing.T, response *api.GenerateInviteLinkResponse)
	}{
		{
			name:   "Generate invite link for event",
			userID: organizer.ID,
			request: api.GenerateInviteLinkRequest{
				EventID: event.ID,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.GenerateInviteLinkResponse) {
				assert.Contains(t, response.InviteLink, "http://localhost:8080/events/join/")
				assert.Len(t, response.InviteLink, len("http://localhost:8080/events/join/")+16)

				var invitation models.EventInvitation
				err := test.TestDB.Where("event_id = ?", event.ID).First(&invitation).Error
				assert.NoError(t, err)
				assert.Equal(t, event.ID, invitation.EventID)
			},
		},
		{
			name:   "Generate invite link for non-existent event",
			userID: organizer.ID,
			request: api.GenerateInviteLinkRequest{
				EventID: 9999,
			},
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
		{
			name:   "Generate invite link for event as participant",
			userID: participant.ID,
			request: api.GenerateInviteLinkRequest{
				EventID: event.ID,
			},
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.GenerateInviteLinkResponse) {
				assert.Contains(t, response.InviteLink, "http://localhost:8080/events/join/")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			requestJSON, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			c.Request = httptest.NewRequest("POST", "/events/invite", bytes.NewBuffer(requestJSON))
			c.Request.Header.Set("Content-Type", "application/json")

			handlers.GenerateInviteLink(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.GenerateInviteLinkResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}

func TestJoinEvent(t *testing.T) {
	cleanup := test.SetupTestDB(t)
	defer cleanup()

	organizer := test.CreateTestUser(t)
	participant := test.CreateTestUser(t)
	newUser := test.CreateTestUser(t)

	event := test.CreateTestEvent(t, organizer.ID)

	test.AddEventParticipant(t, event.ID, participant.ID)

	inviteCode := models.GenerateUniqueInviteCode(test.TestDB)
	invitation := models.EventInvitation{
		EventID:    event.ID,
		InviteCode: inviteCode,
	}
	err := test.TestDB.Create(&invitation).Error
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		userID       uint
		inviteCode   string
		expectedCode int
		validateFunc func(t *testing.T, response *api.JoinEventResponse)
	}{
		{
			name:         "Join event with valid invite code",
			userID:       newUser.ID,
			inviteCode:   inviteCode,
			expectedCode: http.StatusOK,
			validateFunc: func(t *testing.T, response *api.JoinEventResponse) {
				var participation models.EventParticipation
				err := test.TestDB.Where("event_id = ? AND user_id = ?", event.ID, newUser.ID).First(&participation).Error
				assert.NoError(t, err)
			},
		},
		{
			name:         "Join event with invalid invite code",
			userID:       newUser.ID,
			inviteCode:   "invalid",
			expectedCode: http.StatusNotFound,
			validateFunc: nil,
		},
		{
			name:         "Join event as existing participant",
			userID:       participant.ID,
			inviteCode:   inviteCode,
			expectedCode: http.StatusBadRequest,
			validateFunc: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := test.CreateTestContext(t, tc.userID)

			c.Request = httptest.NewRequest("GET", "/events/join/"+tc.inviteCode, nil)
			c.Params = []gin.Param{{Key: "invite_code", Value: tc.inviteCode}}

			handlers.JoinEvent(c, test.TestDB)

			assert.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK && tc.validateFunc != nil {
				var response api.JoinEventResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				tc.validateFunc(t, &response)
			}
		})
	}
}
