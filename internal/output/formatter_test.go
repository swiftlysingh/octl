package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatter(t *testing.T) {
	t.Run("New defaults to table format", func(t *testing.T) {
		f := New("")
		if f.format != FormatTable {
			t.Errorf("New(\"\").format = %v, want %v", f.format, FormatTable)
		}
	})

	t.Run("New accepts valid formats", func(t *testing.T) {
		tests := []struct {
			input    string
			expected Format
		}{
			{"json", FormatJSON},
			{"plain", FormatPlain},
			{"table", FormatTable},
			{"invalid", FormatTable}, // defaults to table
		}

		for _, tt := range tests {
			f := New(tt.input)
			if f.format != tt.expected {
				t.Errorf("New(%q).format = %v, want %v", tt.input, f.format, tt.expected)
			}
		}
	})

	t.Run("JSON output", func(t *testing.T) {
		var buf bytes.Buffer
		f := New("json")
		f.SetWriter(&buf)

		data := map[string]string{"key": "value"}
		if err := f.Print(data); err != nil {
			t.Fatalf("Print() error = %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if result["key"] != "value" {
			t.Errorf("result[\"key\"] = %q, want %q", result["key"], "value")
		}
	})

	t.Run("Plain output with string slice", func(t *testing.T) {
		var buf bytes.Buffer
		f := New("plain")
		f.SetWriter(&buf)

		data := []string{"a", "b", "c"}
		if err := f.Print(data); err != nil {
			t.Fatalf("Print() error = %v", err)
		}

		expected := "a\tb\tc\n"
		if buf.String() != expected {
			t.Errorf("output = %q, want %q", buf.String(), expected)
		}
	})
}

func TestTable(t *testing.T) {
	t.Run("NewTable creates table with headers", func(t *testing.T) {
		table := NewTable("A", "B", "C")
		if len(table.headers) != 3 {
			t.Errorf("len(headers) = %d, want 3", len(table.headers))
		}
	})

	t.Run("AddRow adds rows", func(t *testing.T) {
		table := NewTable("Col1", "Col2")
		table.AddRow("a", "b")
		table.AddRow("c", "d")

		if len(table.rows) != 2 {
			t.Errorf("len(rows) = %d, want 2", len(table.rows))
		}
	})

	t.Run("Render outputs formatted table", func(t *testing.T) {
		var buf bytes.Buffer
		table := NewTable("Name", "Value")
		table.AddRow("foo", "bar")

		if err := table.Render(&buf); err != nil {
			t.Fatalf("Render() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Name") {
			t.Error("output should contain header 'Name'")
		}
		if !strings.Contains(output, "foo") {
			t.Error("output should contain 'foo'")
		}
		if !strings.Contains(output, "bar") {
			t.Error("output should contain 'bar'")
		}
	})

	t.Run("ToPlain returns 2D slice", func(t *testing.T) {
		table := NewTable("A", "B")
		table.AddRow("1", "2")
		table.AddRow("3", "4")

		plain := table.ToPlain()
		if len(plain) != 2 {
			t.Errorf("len(ToPlain()) = %d, want 2", len(plain))
		}
		if plain[0][0] != "1" {
			t.Errorf("plain[0][0] = %q, want %q", plain[0][0], "1")
		}
	})

	t.Run("ToJSON returns map slice", func(t *testing.T) {
		table := NewTable("Name", "Age")
		table.AddRow("Alice", "30")
		table.AddRow("Bob", "25")

		jsonData := table.ToJSON()
		if len(jsonData) != 2 {
			t.Errorf("len(ToJSON()) = %d, want 2", len(jsonData))
		}
		if jsonData[0]["Name"] != "Alice" {
			t.Errorf("jsonData[0][\"Name\"] = %q, want %q", jsonData[0]["Name"], "Alice")
		}
		if jsonData[1]["Age"] != "25" {
			t.Errorf("jsonData[1][\"Age\"] = %q, want %q", jsonData[1]["Age"], "25")
		}
	})
}

func TestPrintTable(t *testing.T) {
	table := NewTable("X", "Y")
	table.AddRow("1", "2")

	// Test JSON format
	if err := PrintTable("json", table); err != nil {
		t.Errorf("PrintTable(json) error = %v", err)
	}

	// Test plain format
	if err := PrintTable("plain", table); err != nil {
		t.Errorf("PrintTable(plain) error = %v", err)
	}

	// Test table format (default)
	if err := PrintTable("table", table); err != nil {
		t.Errorf("PrintTable(table) error = %v", err)
	}
}
