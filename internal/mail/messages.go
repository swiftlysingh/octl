package mail

import (
	"context"
	"fmt"
	"strings"
	"time"

	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// Message represents an email message
type Message struct {
	ID              string    `json:"id"`
	Subject         string    `json:"subject"`
	From            string    `json:"from"`
	To              []string  `json:"to"`
	ReceivedAt      time.Time `json:"received_at"`
	IsRead          bool      `json:"is_read"`
	HasAttachments  bool      `json:"has_attachments"`
	BodyPreview     string    `json:"body_preview,omitempty"`
	Body            string    `json:"body,omitempty"`
	BodyContentType string    `json:"body_content_type,omitempty"`
}

// ListOptions configures message listing
type ListOptions struct {
	Top        int32
	Skip       int32
	Filter     string
	OrderBy    string
	UnreadOnly bool
	FolderID   string
}

// ListMessages retrieves messages from the user's mailbox
func ListMessages(ctx context.Context, client *msgraph.GraphServiceClient, opts ListOptions) ([]Message, error) {
	// Build query parameters
	top := opts.Top
	if top == 0 {
		top = 25
	}

	orderBy := opts.OrderBy
	if orderBy == "" {
		orderBy = "receivedDateTime desc"
	}

	filter := opts.Filter
	if opts.UnreadOnly {
		if filter != "" {
			filter = fmt.Sprintf("(%s) and isRead eq false", filter)
		} else {
			filter = "isRead eq false"
		}
	}

	requestConfig := &users.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMessagesRequestBuilderGetQueryParameters{
			Top:     &top,
			Orderby: []string{orderBy},
			Select:  []string{"id", "subject", "from", "toRecipients", "receivedDateTime", "isRead", "hasAttachments", "bodyPreview"},
		},
	}

	if filter != "" {
		requestConfig.QueryParameters.Filter = &filter
	}

	if opts.Skip > 0 {
		requestConfig.QueryParameters.Skip = &opts.Skip
	}

	var result models.MessageCollectionResponseable
	var err error

	if opts.FolderID != "" {
		result, err = client.Me().MailFolders().ByMailFolderId(opts.FolderID).Messages().Get(ctx, nil)
	} else {
		result, err = client.Me().Messages().Get(ctx, requestConfig)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	messages := make([]Message, 0)
	for _, msg := range result.GetValue() {
		messages = append(messages, convertMessage(msg))
	}

	return messages, nil
}

// GetMessage retrieves a single message by ID
func GetMessage(ctx context.Context, client *msgraph.GraphServiceClient, messageID string) (*Message, error) {
	requestConfig := &users.ItemMessagesMessageItemRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMessagesMessageItemRequestBuilderGetQueryParameters{
			Select: []string{"id", "subject", "from", "toRecipients", "receivedDateTime", "isRead", "hasAttachments", "body"},
		},
	}

	msg, err := client.Me().Messages().ByMessageId(messageID).Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	message := convertMessage(msg)

	// Get full body
	if body := msg.GetBody(); body != nil {
		if content := body.GetContent(); content != nil {
			message.Body = *content
		}
		if contentType := body.GetContentType(); contentType != nil {
			message.BodyContentType = contentType.String()
		}
	}

	return &message, nil
}

// SearchMessages searches messages with a query
func SearchMessages(ctx context.Context, client *msgraph.GraphServiceClient, query string, top int32) ([]Message, error) {
	if top == 0 {
		top = 25
	}

	// Use $search parameter for full-text search
	search := fmt.Sprintf("\"%s\"", query)

	requestConfig := &users.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMessagesRequestBuilderGetQueryParameters{
			Top:    &top,
			Search: &search,
			Select: []string{"id", "subject", "from", "toRecipients", "receivedDateTime", "isRead", "hasAttachments", "bodyPreview"},
		},
	}

	result, err := client.Me().Messages().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	messages := make([]Message, 0)
	for _, msg := range result.GetValue() {
		messages = append(messages, convertMessage(msg))
	}

	return messages, nil
}

// convertMessage converts a Graph API message to our Message type
func convertMessage(msg models.Messageable) Message {
	m := Message{
		ID:             safeString(msg.GetId()),
		Subject:        safeString(msg.GetSubject()),
		IsRead:         safeBool(msg.GetIsRead()),
		HasAttachments: safeBool(msg.GetHasAttachments()),
		BodyPreview:    safeString(msg.GetBodyPreview()),
	}

	if from := msg.GetFrom(); from != nil {
		if addr := from.GetEmailAddress(); addr != nil {
			name := safeString(addr.GetName())
			email := safeString(addr.GetAddress())
			if name != "" && name != email {
				m.From = fmt.Sprintf("%s <%s>", name, email)
			} else {
				m.From = email
			}
		}
	}

	if recipients := msg.GetToRecipients(); recipients != nil {
		for _, r := range recipients {
			if addr := r.GetEmailAddress(); addr != nil {
				m.To = append(m.To, safeString(addr.GetAddress()))
			}
		}
	}

	if received := msg.GetReceivedDateTime(); received != nil {
		m.ReceivedAt = *received
	}

	return m
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

// FormatFrom formats the sender for display (truncated)
func (m *Message) FormatFrom(maxLen int) string {
	from := m.From
	if len(from) > maxLen {
		from = from[:maxLen-3] + "..."
	}
	return from
}

// FormatSubject formats the subject for display (truncated)
func (m *Message) FormatSubject(maxLen int) string {
	subject := m.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	if len(subject) > maxLen {
		subject = subject[:maxLen-3] + "..."
	}
	return subject
}

// FormatDate formats the received date for display
func (m *Message) FormatDate() string {
	now := time.Now()
	if m.ReceivedAt.Year() == now.Year() &&
		m.ReceivedAt.Month() == now.Month() &&
		m.ReceivedAt.Day() == now.Day() {
		return m.ReceivedAt.Format("15:04")
	}
	if m.ReceivedAt.Year() == now.Year() {
		return m.ReceivedAt.Format("Jan 02")
	}
	return m.ReceivedAt.Format("2006-01-02")
}

// StripHTML removes HTML tags for plain text display
func StripHTML(html string) string {
	// Simple HTML stripping - replace tags and decode common entities
	s := html
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</div>", "\n")

	// Remove remaining tags
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}

	// Decode common HTML entities
	s = result.String()
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")

	// Clean up extra whitespace
	lines := strings.Split(s, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
}
