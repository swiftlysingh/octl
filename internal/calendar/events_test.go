package calendar

import (
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	t.Run("FormatTime returns 'All day' for all-day events", func(t *testing.T) {
		e := Event{IsAllDay: true}
		if got := e.FormatTime(); got != "All day" {
			t.Errorf("FormatTime() = %q, want %q", got, "All day")
		}
	})

	t.Run("FormatTime returns time range", func(t *testing.T) {
		e := Event{
			Start:    time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
			End:      time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC),
			IsAllDay: false,
		}
		got := e.FormatTime()
		want := "14:00 - 15:30"
		if got != want {
			t.Errorf("FormatTime() = %q, want %q", got, want)
		}
	})

	t.Run("FormatDate returns formatted date", func(t *testing.T) {
		e := Event{
			Start: time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
		}
		got := e.FormatDate()
		want := "Mon Jan 15"
		if got != want {
			t.Errorf("FormatDate() = %q, want %q", got, want)
		}
	})

	t.Run("Duration calculates correctly", func(t *testing.T) {
		e := Event{
			Start: time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC),
		}
		got := e.Duration()
		want := 90 * time.Minute
		if got != want {
			t.Errorf("Duration() = %v, want %v", got, want)
		}
	})
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		datetime string
		tz       string
		wantYear int
		wantHour int
	}{
		{
			name:     "RFC3339 format",
			datetime: "2024-01-15T14:30:00Z",
			tz:       "UTC",
			wantYear: 2024,
			wantHour: 14,
		},
		{
			name:     "Graph API format",
			datetime: "2024-01-15T14:30:00.0000000",
			tz:       "UTC",
			wantYear: 2024,
			wantHour: 14,
		},
		{
			name:     "Simple format",
			datetime: "2024-01-15T14:30:00",
			tz:       "UTC",
			wantYear: 2024,
			wantHour: 14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDateTime(tt.datetime, tt.tz)
			if got.Year() != tt.wantYear {
				t.Errorf("parseDateTime() year = %d, want %d", got.Year(), tt.wantYear)
			}
			if got.Hour() != tt.wantHour {
				t.Errorf("parseDateTime() hour = %d, want %d", got.Hour(), tt.wantHour)
			}
		})
	}

	t.Run("invalid format returns zero time", func(t *testing.T) {
		got := parseDateTime("not-a-date", "UTC")
		if !got.IsZero() {
			t.Errorf("parseDateTime(invalid) should return zero time, got %v", got)
		}
	})
}

func TestSafeHelpers(t *testing.T) {
	t.Run("safeString with nil", func(t *testing.T) {
		if got := safeString(nil); got != "" {
			t.Errorf("safeString(nil) = %q, want empty", got)
		}
	})

	t.Run("safeString with value", func(t *testing.T) {
		s := "test"
		if got := safeString(&s); got != "test" {
			t.Errorf("safeString(&s) = %q, want %q", got, "test")
		}
	})

	t.Run("safeBool with nil", func(t *testing.T) {
		if got := safeBool(nil); got != false {
			t.Errorf("safeBool(nil) = %v, want false", got)
		}
	})

	t.Run("safeBool with true", func(t *testing.T) {
		b := true
		if got := safeBool(&b); got != true {
			t.Errorf("safeBool(&b) = %v, want true", got)
		}
	})

	t.Run("safeBool with false", func(t *testing.T) {
		b := false
		if got := safeBool(&b); got != false {
			t.Errorf("safeBool(&b) = %v, want false", got)
		}
	})
}
