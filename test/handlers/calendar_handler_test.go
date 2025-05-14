package handlers_test

import (
	"itsplanned/handlers"
	"itsplanned/models"
	"itsplanned/test"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestImportCalendarEvents(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") != "" {
		t.Skip("Skipping integration tests")
	}

	os.Setenv("AES_SECRET", "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=")

	cleanup := test.SetupTestDB(t)
	defer cleanup()

	user := test.CreateTestUser(t)

	token := &models.UserToken{
		UserID:       user.ID,
		AccessToken:  "encrypted_token_for_testing",
		RefreshToken: "test_refresh_token",
		Expiry:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}
	err := test.TestDB.Create(token).Error
	assert.NoError(t, err)

	t.Run("Import calendar events without token", func(t *testing.T) {
		userWithoutToken := test.CreateTestUser(t)

		c, w := test.CreateTestContext(t, userWithoutToken.ID)
		c.Request = httptest.NewRequest("GET", "/calendar/import", nil)

		handlers.ImportCalendarEvents(c, test.TestDB)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	t.Run("Import calendar events with invalid user ID", func(t *testing.T) {
		c, w := test.CreateTestContext(t, uint(9999))
		c.Request = httptest.NewRequest("GET", "/calendar/import", nil)

		handlers.ImportCalendarEvents(c, test.TestDB)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
