package mail

import (
	"context"
	"fmt"

	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// SendOptions configures sending a message
type SendOptions struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	BodyType    string // "text" or "html"
	SaveToSent  bool
}

// SendMessage sends an email
func SendMessage(ctx context.Context, client *msgraph.GraphServiceClient, opts SendOptions) error {
	// Create message
	msg := models.NewMessage()
	msg.SetSubject(&opts.Subject)

	// Set body
	body := models.NewItemBody()
	bodyContent := opts.Body
	body.SetContent(&bodyContent)
	if opts.BodyType == "html" {
		bodyType := models.HTML_BODYTYPE
		body.SetContentType(&bodyType)
	} else {
		bodyType := models.TEXT_BODYTYPE
		body.SetContentType(&bodyType)
	}
	msg.SetBody(body)

	// Set recipients
	toRecipients := make([]models.Recipientable, len(opts.To))
	for i, addr := range opts.To {
		recipient := models.NewRecipient()
		emailAddr := models.NewEmailAddress()
		emailAddr.SetAddress(&addr)
		recipient.SetEmailAddress(emailAddr)
		toRecipients[i] = recipient
	}
	msg.SetToRecipients(toRecipients)

	// Set CC
	if len(opts.Cc) > 0 {
		ccRecipients := make([]models.Recipientable, len(opts.Cc))
		for i, addr := range opts.Cc {
			recipient := models.NewRecipient()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&addr)
			recipient.SetEmailAddress(emailAddr)
			ccRecipients[i] = recipient
		}
		msg.SetCcRecipients(ccRecipients)
	}

	// Set BCC
	if len(opts.Bcc) > 0 {
		bccRecipients := make([]models.Recipientable, len(opts.Bcc))
		for i, addr := range opts.Bcc {
			recipient := models.NewRecipient()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&addr)
			recipient.SetEmailAddress(emailAddr)
			bccRecipients[i] = recipient
		}
		msg.SetBccRecipients(bccRecipients)
	}

	// Create send mail request
	sendMailBody := users.NewItemSendMailPostRequestBody()
	sendMailBody.SetMessage(msg)
	saveToSent := opts.SaveToSent
	sendMailBody.SetSaveToSentItems(&saveToSent)

	// Send
	err := client.Me().SendMail().Post(ctx, sendMailBody, nil)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// CreateDraft creates a draft message
func CreateDraft(ctx context.Context, client *msgraph.GraphServiceClient, opts SendOptions) (*Message, error) {
	// Create message
	msg := models.NewMessage()
	msg.SetSubject(&opts.Subject)

	// Set body
	body := models.NewItemBody()
	bodyContent := opts.Body
	body.SetContent(&bodyContent)
	if opts.BodyType == "html" {
		bodyType := models.HTML_BODYTYPE
		body.SetContentType(&bodyType)
	} else {
		bodyType := models.TEXT_BODYTYPE
		body.SetContentType(&bodyType)
	}
	msg.SetBody(body)

	// Set recipients
	if len(opts.To) > 0 {
		toRecipients := make([]models.Recipientable, len(opts.To))
		for i, addr := range opts.To {
			recipient := models.NewRecipient()
			emailAddr := models.NewEmailAddress()
			emailAddr.SetAddress(&addr)
			recipient.SetEmailAddress(emailAddr)
			toRecipients[i] = recipient
		}
		msg.SetToRecipients(toRecipients)
	}

	// Create draft
	draft, err := client.Me().Messages().Post(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	result := convertMessage(draft)
	return &result, nil
}

// MarkAsRead marks a message as read
func MarkAsRead(ctx context.Context, client *msgraph.GraphServiceClient, messageID string, isRead bool) error {
	msg := models.NewMessage()
	msg.SetIsRead(&isRead)

	_, err := client.Me().Messages().ByMessageId(messageID).Patch(ctx, msg, nil)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	return nil
}

// DeleteMessage deletes a message
func DeleteMessage(ctx context.Context, client *msgraph.GraphServiceClient, messageID string) error {
	err := client.Me().Messages().ByMessageId(messageID).Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
