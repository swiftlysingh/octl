package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Format represents the output format type
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
)

// Formatter handles output formatting
type Formatter struct {
	format Format
	writer io.Writer
}

// New creates a new Formatter with the specified format
func New(format string) *Formatter {
	f := Format(format)
	if f != FormatTable && f != FormatJSON && f != FormatPlain {
		f = FormatTable
	}
	return &Formatter{
		format: f,
		writer: os.Stdout,
	}
}

// SetWriter sets the output writer
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// Print outputs data in the configured format
func (f *Formatter) Print(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.printJSON(data)
	case FormatPlain:
		return f.printPlain(data)
	default:
		return f.printTable(data)
	}
}

// printJSON outputs data as JSON
func (f *Formatter) printJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// printPlain outputs data as plain text (tab-separated)
func (f *Formatter) printPlain(data interface{}) error {
	switch v := data.(type) {
	case [][]string:
		for _, row := range v {
			if _, err := fmt.Fprintln(f.writer, strings.Join(row, "\t")); err != nil {
				return err
			}
		}
	case []string:
		if _, err := fmt.Fprintln(f.writer, strings.Join(v, "\t")); err != nil {
			return err
		}
	case string:
		if _, err := fmt.Fprintln(f.writer, v); err != nil {
			return err
		}
	default:
		// For complex types, fall back to JSON
		return f.printJSON(data)
	}
	return nil
}

// printTable outputs data as a formatted table
func (f *Formatter) printTable(data interface{}) error {
	switch v := data.(type) {
	case *Table:
		return v.Render(f.writer)
	default:
		// For non-table data, just print as string
		_, err := fmt.Fprintln(f.writer, data)
		return err
	}
}

// Table represents tabular data
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a new table with the given headers
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells ...string) {
	t.rows = append(t.rows, cells)
}

// Render outputs the table to the writer
func (t *Table) Render(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Print headers
	if len(t.headers) > 0 {
		if _, err := fmt.Fprintln(tw, strings.Join(t.headers, "\t")); err != nil {
			return err
		}
		// Print separator
		sep := make([]string, len(t.headers))
		for i, h := range t.headers {
			sep[i] = strings.Repeat("-", len(h))
		}
		if _, err := fmt.Fprintln(tw, strings.Join(sep, "\t")); err != nil {
			return err
		}
	}

	// Print rows
	for _, row := range t.rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}

	return tw.Flush()
}

// ToPlain returns the table as a 2D string array for plain output
func (t *Table) ToPlain() [][]string {
	return t.rows
}

// ToJSON returns a JSON-serializable representation
func (t *Table) ToJSON() []map[string]string {
	result := make([]map[string]string, len(t.rows))
	for i, row := range t.rows {
		result[i] = make(map[string]string)
		for j, cell := range row {
			if j < len(t.headers) {
				result[i][t.headers[j]] = cell
			}
		}
	}
	return result
}

// PrintTable is a convenience function for printing tables
func PrintTable(format string, table *Table) error {
	f := New(format)
	switch Format(format) {
	case FormatJSON:
		return f.printJSON(table.ToJSON())
	case FormatPlain:
		return f.printPlain(table.ToPlain())
	default:
		return table.Render(f.writer)
	}
}
