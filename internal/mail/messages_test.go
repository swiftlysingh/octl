package mail

import (
	"testing"
	"time"
)

func TestMessage(t *testing.T) {
	t.Run("FormatFrom truncates long addresses", func(t *testing.T) {
		m := Message{From: "verylongemailaddress@verylongdomain.example.com"}
		formatted := m.FormatFrom(20)
		if len(formatted) > 20 {
			t.Errorf("FormatFrom(20) length = %d, want <= 20", len(formatted))
		}
		if formatted[len(formatted)-3:] != "..." {
			t.Error("FormatFrom should end with '...' when truncated")
		}
	})

	t.Run("FormatFrom keeps short addresses", func(t *testing.T) {
		m := Message{From: "short@test.com"}
		formatted := m.FormatFrom(50)
		if formatted != "short@test.com" {
			t.Errorf("FormatFrom(50) = %q, want %q", formatted, "short@test.com")
		}
	})

	t.Run("FormatSubject returns placeholder for empty", func(t *testing.T) {
		m := Message{Subject: ""}
		formatted := m.FormatSubject(50)
		if formatted != "(no subject)" {
			t.Errorf("FormatSubject = %q, want %q", formatted, "(no subject)")
		}
	})

	t.Run("FormatSubject truncates long subjects", func(t *testing.T) {
		m := Message{Subject: "This is a very long subject line that should be truncated"}
		formatted := m.FormatSubject(20)
		if len(formatted) > 20 {
			t.Errorf("FormatSubject(20) length = %d, want <= 20", len(formatted))
		}
	})

	t.Run("FormatDate shows time for today", func(t *testing.T) {
		now := time.Now()
		m := Message{ReceivedAt: now}
		formatted := m.FormatDate()
		// Should be in HH:MM format
		if len(formatted) != 5 || formatted[2] != ':' {
			t.Errorf("FormatDate for today = %q, expected HH:MM format", formatted)
		}
	})

	t.Run("FormatDate shows month/day for this year", func(t *testing.T) {
		now := time.Now()
		past := now.AddDate(0, -1, 0) // 1 month ago
		m := Message{ReceivedAt: past}
		formatted := m.FormatDate()
		// Should be in "Jan 02" format
		if len(formatted) < 5 {
			t.Errorf("FormatDate for this year = %q, expected 'Mon DD' format", formatted)
		}
	})

	t.Run("FormatDate shows full date for other years", func(t *testing.T) {
		past := time.Date(2020, 6, 15, 10, 30, 0, 0, time.UTC)
		m := Message{ReceivedAt: past}
		formatted := m.FormatDate()
		if formatted != "2020-06-15" {
			t.Errorf("FormatDate for 2020 = %q, want %q", formatted, "2020-06-15")
		}
	})
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "removes simple tags",
			input: "<p>Hello</p>",
			want:  "Hello",
		},
		{
			name:  "converts br to newline",
			input: "Line1<br>Line2<br/>Line3<br />Line4",
			want:  "Line1\nLine2\nLine3\nLine4",
		},
		{
			name:  "converts p end tag to double newline",
			input: "<p>Para1</p><p>Para2</p>",
			want:  "Para1\nPara2",
		},
		{
			name:  "decodes HTML entities",
			input: "&amp; &lt; &gt; &quot; &#39; &nbsp;",
			want:  "& < > \" '",
		},
		{
			name:  "handles nested tags",
			input: "<div><span><b>Bold</b></span></div>",
			want:  "Bold",
		},
		{
			name:  "handles empty input",
			input: "",
			want:  "",
		},
		{
			name:  "preserves plain text",
			input: "Just plain text",
			want:  "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHTML(tt.input)
			if got != tt.want {
				t.Errorf("StripHTML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	t.Run("returns empty string for nil", func(t *testing.T) {
		if got := safeString(nil); got != "" {
			t.Errorf("safeString(nil) = %q, want %q", got, "")
		}
	})

	t.Run("returns value for non-nil", func(t *testing.T) {
		s := "test"
		if got := safeString(&s); got != "test" {
			t.Errorf("safeString(&s) = %q, want %q", got, "test")
		}
	})
}

func TestSafeBool(t *testing.T) {
	t.Run("returns false for nil", func(t *testing.T) {
		if got := safeBool(nil); got != false {
			t.Errorf("safeBool(nil) = %v, want false", got)
		}
	})

	t.Run("returns value for non-nil", func(t *testing.T) {
		b := true
		if got := safeBool(&b); got != true {
			t.Errorf("safeBool(&b) = %v, want true", got)
		}
	})
}
