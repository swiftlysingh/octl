package calendar

import (
	"context"
	"fmt"
	"time"

	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// Event represents a calendar event
type Event struct {
	ID               string    `json:"id"`
	Subject          string    `json:"subject"`
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	Location         string    `json:"location,omitempty"`
	IsAllDay         bool      `json:"is_all_day"`
	Organizer        string    `json:"organizer,omitempty"`
	Attendees        []string  `json:"attendees,omitempty"`
	Body             string    `json:"body,omitempty"`
	BodyContentType  string    `json:"body_content_type,omitempty"`
	WebLink          string    `json:"web_link,omitempty"`
	ResponseStatus   string    `json:"response_status,omitempty"`
	IsOnline         bool      `json:"is_online"`
	OnlineMeetingURL string    `json:"online_meeting_url,omitempty"`
}

// ListOptions configures event listing
type ListOptions struct {
	StartTime time.Time
	EndTime   time.Time
	Top       int32
}

// ListEvents retrieves calendar events within a time range
func ListEvents(ctx context.Context, client *msgraph.GraphServiceClient, opts ListOptions) ([]Event, error) {
	// Use calendar view for time range queries
	startStr := opts.StartTime.Format(time.RFC3339)
	endStr := opts.EndTime.Format(time.RFC3339)

	top := opts.Top
	if top == 0 {
		top = 50
	}

	requestConfig := &users.ItemCalendarCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarCalendarViewRequestBuilderGetQueryParameters{
			StartDateTime: &startStr,
			EndDateTime:   &endStr,
			Top:           &top,
			Orderby:       []string{"start/dateTime"},
			Select:        []string{"id", "subject", "start", "end", "location", "isAllDay", "organizer", "attendees", "webLink", "responseStatus", "isOnlineMeeting", "onlineMeetingUrl"},
		},
	}

	result, err := client.Me().Calendar().CalendarView().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]Event, 0)
	for _, ev := range result.GetValue() {
		events = append(events, convertEvent(ev))
	}

	return events, nil
}

// GetEvent retrieves a single event by ID
func GetEvent(ctx context.Context, client *msgraph.GraphServiceClient, eventID string) (*Event, error) {
	requestConfig := &users.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemEventsEventItemRequestBuilderGetQueryParameters{
			Select: []string{"id", "subject", "start", "end", "location", "isAllDay", "organizer", "attendees", "body", "webLink", "responseStatus", "isOnlineMeeting", "onlineMeetingUrl"},
		},
	}

	ev, err := client.Me().Events().ByEventId(eventID).Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	event := convertEvent(ev)

	// Get full body
	if body := ev.GetBody(); body != nil {
		if content := body.GetContent(); content != nil {
			event.Body = *content
		}
		if contentType := body.GetContentType(); contentType != nil {
			event.BodyContentType = contentType.String()
		}
	}

	return &event, nil
}

// CreateEventOptions configures event creation
type CreateEventOptions struct {
	Subject   string
	Start     time.Time
	End       time.Time
	Location  string
	Body      string
	IsAllDay  bool
	Attendees []string
	IsOnline  bool
}

// CreateEvent creates a new calendar event
func CreateEvent(ctx context.Context, client *msgraph.GraphServiceClient, opts CreateEventOptions) (*Event, error) {
	ev := models.NewEvent()
	ev.SetSubject(&opts.Subject)

	// Set start time
	start := models.NewDateTimeTimeZone()
	startStr := opts.Start.Format("2006-01-02T15:04:05")
	start.SetDateTime(&startStr)
	tz := "UTC"
	start.SetTimeZone(&tz)
	ev.SetStart(start)

	// Set end time
	end := models.NewDateTimeTimeZone()
	endStr := opts.End.Format("2006-01-02T15:04:05")
	end.SetDateTime(&endStr)
	end.SetTimeZone(&tz)
	ev.SetEnd(end)

	ev.SetIsAllDay(&opts.IsAllDay)

	// Set location
	if opts.Location != "" {
		loc := models.NewLocation()
		loc.SetDisplayName(&opts.Location)
		ev.SetLocation(loc)
	}

	// Set body
	if opts.Body != "" {
		body := models.NewItemBody()
		body.SetContent(&opts.Body)
		bodyType := models.TEXT_BODYTYPE
		body.SetContentType(&bodyType)
		ev.SetBody(body)
	}

	// Set attendees
	if len(opts.Attendees) > 0 {
		attendees := make([]models.Attendeeable, len(opts.Attendees))
		for i, email := range opts.Attendees {
			attendee := models.NewAttendee()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&email)
			attendee.SetEmailAddress(emailAddr)
			attendeeType := models.REQUIRED_ATTENDEETYPE
			attendee.SetTypeEscaped(&attendeeType)
			attendees[i] = attendee
		}
		ev.SetAttendees(attendees)
	}

	// Set online meeting
	ev.SetIsOnlineMeeting(&opts.IsOnline)

	created, err := client.Me().Events().Post(ctx, ev, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	event := convertEvent(created)
	return &event, nil
}

// RespondToEvent responds to a meeting invitation
func RespondToEvent(ctx context.Context, client *msgraph.GraphServiceClient, eventID string, response string, comment string) error {
	sendResponse := true

	switch response {
	case "accept":
		body := users.NewItemEventsItemAcceptPostRequestBody()
		body.SetComment(&comment)
		body.SetSendResponse(&sendResponse)
		err := client.Me().Events().ByEventId(eventID).Accept().Post(ctx, body, nil)
		if err != nil {
			return fmt.Errorf("failed to accept event: %w", err)
		}
	case "decline":
		body := users.NewItemEventsItemDeclinePostRequestBody()
		body.SetComment(&comment)
		body.SetSendResponse(&sendResponse)
		err := client.Me().Events().ByEventId(eventID).Decline().Post(ctx, body, nil)
		if err != nil {
			return fmt.Errorf("failed to decline event: %w", err)
		}
	case "tentative":
		body := users.NewItemEventsItemTentativelyAcceptPostRequestBody()
		body.SetComment(&comment)
		body.SetSendResponse(&sendResponse)
		err := client.Me().Events().ByEventId(eventID).TentativelyAccept().Post(ctx, body, nil)
		if err != nil {
			return fmt.Errorf("failed to tentatively accept event: %w", err)
		}
	default:
		return fmt.Errorf("invalid response: %s (use accept, decline, or tentative)", response)
	}

	return nil
}

// DeleteEvent deletes a calendar event
func DeleteEvent(ctx context.Context, client *msgraph.GraphServiceClient, eventID string) error {
	err := client.Me().Events().ByEventId(eventID).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// convertEvent converts a Graph API event to our Event type
func convertEvent(ev models.Eventable) Event {
	event := Event{
		ID:       safeString(ev.GetId()),
		Subject:  safeString(ev.GetSubject()),
		WebLink:  safeString(ev.GetWebLink()),
		IsAllDay: safeBool(ev.GetIsAllDay()),
		IsOnline: safeBool(ev.GetIsOnlineMeeting()),
	}

	if url := ev.GetOnlineMeetingUrl(); url != nil {
		event.OnlineMeetingURL = *url
	}

	if start := ev.GetStart(); start != nil {
		if dt := start.GetDateTime(); dt != nil {
			tz := "UTC"
			if start.GetTimeZone() != nil {
				tz = *start.GetTimeZone()
			}
			event.Start = parseDateTime(*dt, tz)
		}
	}

	if end := ev.GetEnd(); end != nil {
		if dt := end.GetDateTime(); dt != nil {
			tz := "UTC"
			if end.GetTimeZone() != nil {
				tz = *end.GetTimeZone()
			}
			event.End = parseDateTime(*dt, tz)
		}
	}

	if loc := ev.GetLocation(); loc != nil {
		event.Location = safeString(loc.GetDisplayName())
	}

	if org := ev.GetOrganizer(); org != nil {
		if email := org.GetEmailAddress(); email != nil {
			event.Organizer = safeString(email.GetAddress())
		}
	}

	if attendees := ev.GetAttendees(); attendees != nil {
		for _, a := range attendees {
			if email := a.GetEmailAddress(); email != nil {
				event.Attendees = append(event.Attendees, safeString(email.GetAddress()))
			}
		}
	}

	if resp := ev.GetResponseStatus(); resp != nil {
		if r := resp.GetResponse(); r != nil {
			event.ResponseStatus = r.String()
		}
	}

	return event
}

func parseDateTime(dt string, tz string) time.Time {
	// Try parsing with timezone
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	// Try various formats
	formats := []string{
		"2006-01-02T15:04:05.0000000",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dt, loc); err == nil {
			return t
		}
	}

	return time.Time{}
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// FormatTime formats time for display
func (e *Event) FormatTime() string {
	if e.IsAllDay {
		return "All day"
	}
	return fmt.Sprintf("%s - %s", e.Start.Format("15:04"), e.End.Format("15:04"))
}

// FormatDate formats date for display
func (e *Event) FormatDate() string {
	if e.IsAllDay {
		return e.Start.Format("Mon Jan 02")
	}
	return e.Start.Format("Mon Jan 02")
}

// Duration returns the event duration
func (e *Event) Duration() time.Duration {
	return e.End.Sub(e.Start)
}
