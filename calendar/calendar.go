package calendar

import (
	"context"
	"log"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Fetching user events from Google Calendar
func FetchGoogleCalendarEvents(accessToken string, startDate, endDate time.Time) ([]*calendar.Event, error) {
	ctx := context.Background()

	srv, err := calendar.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
		return nil, err
	}

	tStart := startDate.Format(time.RFC3339)
	tEnd := endDate.Format(time.RFC3339)

	events, err := srv.Events.List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(tStart).
		TimeMax(tEnd).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events: %v", err)
		return nil, err
	}

	return events.Items, nil
}
