package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pp/octl/internal/calendar"
	"github.com/pp/octl/internal/mail"
	"github.com/pp/octl/internal/output"
	"github.com/spf13/cobra"
)

var (
	// calendar list flags
	calendarDays int

	// calendar create flags
	calEventSubject   string
	calEventStart     string
	calEventEnd       string
	calEventDuration  string
	calEventLocation  string
	calEventBody      string
	calEventAllDay    bool
	calEventAttendees []string
	calEventOnline    bool

	// calendar respond flags
	calResponseComment string
)

var calendarCmd = &cobra.Command{
	Use:     "calendar",
	Aliases: []string{"cal"},
	Short:   "Manage calendar events",
	Long:    `List, view, create, and respond to calendar events.`,
}

var calendarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List upcoming events",
	Long:  `List calendar events for the specified time range.`,
	RunE:  runCalendarList,
}

var calendarTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's events",
	Long:  `Show all calendar events for today.`,
	RunE:  runCalendarToday,
}

var calendarWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show this week's events",
	Long:  `Show all calendar events for the current week.`,
	RunE:  runCalendarWeek,
}

var calendarShowCmd = &cobra.Command{
	Use:   "show <event-id>",
	Short: "Show event details",
	Long:  `Show detailed information about a calendar event.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCalendarShow,
}

var calendarCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new event",
	Long: `Create a new calendar event.

Examples:
  # Create a simple event
  octl calendar create --subject "Team Meeting" --start "2024-01-15T14:00:00" --duration 1h

  # Create an all-day event
  octl calendar create --subject "Conference" --start "2024-01-20" --all-day

  # Create an online meeting with attendees
  octl calendar create --subject "Sync" --start "2024-01-15T10:00:00" --duration 30m --online --attendees user@example.com`,
	RunE: runCalendarCreate,
}

var calendarRespondCmd = &cobra.Command{
	Use:   "respond <event-id> <accept|decline|tentative>",
	Short: "Respond to an event invitation",
	Long:  `Accept, decline, or tentatively accept a meeting invitation.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runCalendarRespond,
}

var calendarDeleteCmd = &cobra.Command{
	Use:   "delete <event-id>",
	Short: "Delete an event",
	Long:  `Delete a calendar event.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runCalendarDelete,
}

func init() {
	rootCmd.AddCommand(calendarCmd)
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarTodayCmd)
	calendarCmd.AddCommand(calendarWeekCmd)
	calendarCmd.AddCommand(calendarShowCmd)
	calendarCmd.AddCommand(calendarCreateCmd)
	calendarCmd.AddCommand(calendarRespondCmd)
	calendarCmd.AddCommand(calendarDeleteCmd)

	// calendar list flags
	calendarListCmd.Flags().IntVarP(&calendarDays, "days", "d", 7, "Number of days to show")

	// calendar create flags
	calendarCreateCmd.Flags().StringVar(&calEventSubject, "subject", "", "Event subject/title")
	calendarCreateCmd.Flags().StringVar(&calEventStart, "start", "", "Start time (RFC3339 or YYYY-MM-DD for all-day)")
	calendarCreateCmd.Flags().StringVar(&calEventEnd, "end", "", "End time (optional if duration specified)")
	calendarCreateCmd.Flags().StringVar(&calEventDuration, "duration", "1h", "Duration (e.g., 30m, 1h, 2h30m)")
	calendarCreateCmd.Flags().StringVar(&calEventLocation, "location", "", "Event location")
	calendarCreateCmd.Flags().StringVar(&calEventBody, "body", "", "Event description")
	calendarCreateCmd.Flags().BoolVar(&calEventAllDay, "all-day", false, "Create an all-day event")
	calendarCreateCmd.Flags().StringSliceVar(&calEventAttendees, "attendees", nil, "Attendee email addresses")
	calendarCreateCmd.Flags().BoolVar(&calEventOnline, "online", false, "Create as online meeting")
	calendarCreateCmd.MarkFlagRequired("subject")
	calendarCreateCmd.MarkFlagRequired("start")

	// calendar respond flags
	calendarRespondCmd.Flags().StringVar(&calResponseComment, "comment", "", "Optional comment with response")
}

func runCalendarList(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, calendarDays)

	opts := calendar.ListOptions{
		StartTime: start,
		EndTime:   end,
	}

	events, err := calendar.ListEvents(ctx, client.Graph(), opts)
	if err != nil {
		return err
	}

	return printEvents(cmd, events)
}

func runCalendarToday(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)

	opts := calendar.ListOptions{
		StartTime: start,
		EndTime:   end,
	}

	events, err := calendar.ListEvents(ctx, client.Graph(), opts)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No events today")
		return nil
	}

	return printEvents(cmd, events)
}

func runCalendarWeek(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	// Start from beginning of current week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 7)

	opts := calendar.ListOptions{
		StartTime: start,
		EndTime:   end,
	}

	events, err := calendar.ListEvents(ctx, client.Graph(), opts)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No events this week")
		return nil
	}

	return printEvents(cmd, events)
}

func runCalendarShow(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	event, err := calendar.GetEvent(ctx, client.Graph(), eventID)
	if err != nil {
		return err
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(event)
	}

	// Print event details
	fmt.Printf("Subject:   %s\n", event.Subject)
	fmt.Printf("Date:      %s\n", event.FormatDate())
	fmt.Printf("Time:      %s\n", event.FormatTime())
	if event.Location != "" {
		fmt.Printf("Location:  %s\n", event.Location)
	}
	if event.Organizer != "" {
		fmt.Printf("Organizer: %s\n", event.Organizer)
	}
	if len(event.Attendees) > 0 {
		fmt.Printf("Attendees: %s\n", strings.Join(event.Attendees, ", "))
	}
	if event.ResponseStatus != "" {
		fmt.Printf("Response:  %s\n", event.ResponseStatus)
	}
	if event.IsOnline {
		fmt.Println("Type:      Online meeting")
		if event.OnlineMeetingURL != "" {
			fmt.Printf("Join URL:  %s\n", event.OnlineMeetingURL)
		}
	}
	if event.WebLink != "" {
		fmt.Printf("Web Link:  %s\n", event.WebLink)
	}

	if event.Body != "" {
		fmt.Println()
		fmt.Println("---")
		fmt.Println()
		body := event.Body
		if event.BodyContentType == "html" {
			body = mail.StripHTML(body)
		}
		fmt.Println(body)
	}

	return nil
}

func runCalendarCreate(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	// Parse start time
	var startTime time.Time
	if calEventAllDay {
		startTime, err = time.Parse("2006-01-02", calEventStart)
		if err != nil {
			return fmt.Errorf("invalid start date (use YYYY-MM-DD for all-day events): %w", err)
		}
	} else {
		startTime, err = time.Parse(time.RFC3339, calEventStart)
		if err != nil {
			// Try without timezone
			startTime, err = time.Parse("2006-01-02T15:04:05", calEventStart)
			if err != nil {
				return fmt.Errorf("invalid start time (use RFC3339 format): %w", err)
			}
		}
	}

	// Parse end time
	var endTime time.Time
	if calEventEnd != "" {
		endTime, err = time.Parse(time.RFC3339, calEventEnd)
		if err != nil {
			endTime, err = time.Parse("2006-01-02T15:04:05", calEventEnd)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}
		}
	} else if calEventAllDay {
		endTime = startTime.AddDate(0, 0, 1)
	} else {
		// Parse duration
		duration, err := time.ParseDuration(calEventDuration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
		endTime = startTime.Add(duration)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := calendar.CreateEventOptions{
		Subject:   calEventSubject,
		Start:     startTime,
		End:       endTime,
		Location:  calEventLocation,
		Body:      calEventBody,
		IsAllDay:  calEventAllDay,
		Attendees: calEventAttendees,
		IsOnline:  calEventOnline,
	}

	event, err := calendar.CreateEvent(ctx, client.Graph(), opts)
	if err != nil {
		return err
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(event)
	}

	fmt.Printf("Event created: %s\n", event.Subject)
	fmt.Printf("ID: %s\n", event.ID)
	fmt.Printf("Time: %s %s\n", event.FormatDate(), event.FormatTime())
	if event.WebLink != "" {
		fmt.Printf("Link: %s\n", event.WebLink)
	}

	return nil
}

func runCalendarRespond(cmd *cobra.Command, args []string) error {
	eventID := args[0]
	response := strings.ToLower(args[1])

	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := calendar.RespondToEvent(ctx, client.Graph(), eventID, response, calResponseComment); err != nil {
		return err
	}

	fmt.Printf("Response sent: %s\n", response)
	return nil
}

func runCalendarDelete(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := calendar.DeleteEvent(ctx, client.Graph(), eventID); err != nil {
		return err
	}

	fmt.Println("Event deleted")
	return nil
}

func printEvents(cmd *cobra.Command, events []calendar.Event) error {
	if len(events) == 0 {
		fmt.Println("No events found")
		return nil
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(events)
	}

	table := output.NewTable("DATE", "TIME", "SUBJECT", "LOCATION")
	for _, ev := range events {
		location := ev.Location
		if location == "" && ev.IsOnline {
			location = "Online"
		}
		if len(location) > 25 {
			location = location[:22] + "..."
		}
		subject := ev.Subject
		if len(subject) > 40 {
			subject = subject[:37] + "..."
		}
		table.AddRow(
			ev.FormatDate(),
			ev.FormatTime(),
			subject,
			location,
		)
	}

	if format == "plain" {
		return output.New(format).Print(table.ToPlain())
	}

	return table.Render(cmd.OutOrStdout())
}
