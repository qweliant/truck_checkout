package calendar

import (
    "context"
    "os"

    "golang.org/x/oauth2/google"
    "google.golang.org/api/calendar/v3"
    "google.golang.org/api/option"
)

// NewCalendarService creates and returns an authenticated Google Calendar service client.
func NewCalendarService(jsonKeyPath string) (*calendar.Service, error) {
    ctx := context.Background()

    // Read the service account key file.
    jsonCredentials, err := os.ReadFile(jsonKeyPath)
    if err != nil {
        return nil, err
    }

    // Configure the JWT config
    config, err := google.JWTConfigFromJSON(jsonCredentials, calendar.CalendarScope)
    if err != nil {
        return nil, err
    }

    // Create an HTTP client with the JWT config
    client := config.Client(ctx)

    // Create a new calendar service client.
    srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
    if err != nil {
        return nil, err
    }

    return srv, nil
}