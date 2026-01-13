package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pp/octl/internal/auth"
	"github.com/pp/octl/internal/config"
	"github.com/pp/octl/internal/graph"
	"github.com/pp/octl/internal/mail"
	"github.com/pp/octl/internal/output"
	"github.com/spf13/cobra"
)

var (
	// mail list flags
	mailListCount  int32
	mailListUnread bool
	mailListFolder string

	// mail send flags
	mailTo      []string
	mailCc      []string
	mailBcc     []string
	mailSubject string
	mailBody    string
	mailHTML    bool
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Manage email messages",
	Long:  `List, read, search, and send email messages.`,
}

var mailListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent emails",
	Long:  `List recent email messages from your inbox.`,
	RunE:  runMailList,
}

var mailReadCmd = &cobra.Command{
	Use:   "read <message-id>",
	Short: "Read an email message",
	Long:  `Read the full content of an email message.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMailRead,
}

var mailSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search emails",
	Long:  `Search email messages by keywords.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runMailSearch,
}

var mailFoldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "List mail folders",
	Long:  `List all mail folders in your mailbox.`,
	RunE:  runMailFolders,
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email",
	Long: `Send an email message.

Example:
  octl mail send --to user@example.com --subject "Hello" --body "Message body"`,
	RunE: runMailSend,
}

var mailDraftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Create a draft email",
	Long:  `Create a draft email message.`,
	RunE:  runMailDraft,
}

var mailMoveCmd = &cobra.Command{
	Use:   "move <message-id> <folder>",
	Short: "Move a message to a folder",
	Long: `Move an email message to a different folder.

Folder can be a folder ID or well-known name: inbox, drafts, sentitems, deleteditems, junkemail, archive`,
	Args: cobra.ExactArgs(2),
	RunE: runMailMove,
}

func init() {
	rootCmd.AddCommand(mailCmd)
	mailCmd.AddCommand(mailListCmd)
	mailCmd.AddCommand(mailReadCmd)
	mailCmd.AddCommand(mailSearchCmd)
	mailCmd.AddCommand(mailFoldersCmd)
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailDraftCmd)
	mailCmd.AddCommand(mailMoveCmd)

	// mail list flags
	mailListCmd.Flags().Int32VarP(&mailListCount, "count", "n", 25, "Number of messages to list")
	mailListCmd.Flags().BoolVarP(&mailListUnread, "unread", "u", false, "Only show unread messages")
	mailListCmd.Flags().StringVarP(&mailListFolder, "folder", "f", "", "Folder to list messages from")

	// mail search flags
	mailSearchCmd.Flags().Int32VarP(&mailListCount, "count", "n", 25, "Maximum number of results")

	// mail send flags
	mailSendCmd.Flags().StringSliceVar(&mailTo, "to", nil, "Recipient email address(es)")
	mailSendCmd.Flags().StringSliceVar(&mailCc, "cc", nil, "CC recipient(s)")
	mailSendCmd.Flags().StringSliceVar(&mailBcc, "bcc", nil, "BCC recipient(s)")
	mailSendCmd.Flags().StringVar(&mailSubject, "subject", "", "Email subject")
	mailSendCmd.Flags().StringVar(&mailBody, "body", "", "Email body")
	mailSendCmd.Flags().BoolVar(&mailHTML, "html", false, "Send body as HTML")
	mailSendCmd.MarkFlagRequired("to")
	mailSendCmd.MarkFlagRequired("subject")
	mailSendCmd.MarkFlagRequired("body")

	// mail draft flags
	mailDraftCmd.Flags().StringSliceVar(&mailTo, "to", nil, "Recipient email address(es)")
	mailDraftCmd.Flags().StringVar(&mailSubject, "subject", "", "Email subject")
	mailDraftCmd.Flags().StringVar(&mailBody, "body", "", "Email body")
	mailDraftCmd.Flags().BoolVar(&mailHTML, "html", false, "Body is HTML")
}

func getGraphClient() (*graph.Client, error) {
	clientID := config.GetClientID()
	if clientID == "" {
		return nil, fmt.Errorf("not configured - run 'octl auth login --client-id <your-id>' first")
	}

	authMgr := auth.NewManager(clientID)
	if err := authMgr.LoadCredential(); err != nil {
		return nil, fmt.Errorf("not logged in - run 'octl auth login' first")
	}

	return graph.NewClient(authMgr.GetCredential())
}

func runMailList(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := mail.ListOptions{
		Top:        mailListCount,
		UnreadOnly: mailListUnread,
		FolderID:   mailListFolder,
	}

	messages, err := mail.ListMessages(ctx, client.Graph(), opts)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		fmt.Println("No messages found")
		return nil
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(messages)
	}

	table := output.NewTable("ID", "FROM", "SUBJECT", "DATE", "READ")
	for _, msg := range messages {
		read := " "
		if msg.IsRead {
			read = "✓"
		}
		table.AddRow(
			msg.ID[:8]+"...",
			msg.FormatFrom(30),
			msg.FormatSubject(50),
			msg.FormatDate(),
			read,
		)
	}

	if format == "plain" {
		return output.New(format).Print(table.ToPlain())
	}

	return table.Render(cmd.OutOrStdout())
}

func runMailRead(cmd *cobra.Command, args []string) error {
	messageID := args[0]

	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	msg, err := mail.GetMessage(ctx, client.Graph(), messageID)
	if err != nil {
		return err
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(msg)
	}

	// Print message details
	fmt.Printf("From:    %s\n", msg.From)
	fmt.Printf("To:      %s\n", strings.Join(msg.To, ", "))
	fmt.Printf("Subject: %s\n", msg.Subject)
	fmt.Printf("Date:    %s\n", msg.ReceivedAt.Format(time.RFC1123))
	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Print body
	body := msg.Body
	if msg.BodyContentType == "html" {
		body = mail.StripHTML(body)
	}
	fmt.Println(body)

	return nil
}

func runMailSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := mail.SearchMessages(ctx, client.Graph(), query, mailListCount)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		fmt.Println("No messages found")
		return nil
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(messages)
	}

	table := output.NewTable("ID", "FROM", "SUBJECT", "DATE")
	for _, msg := range messages {
		table.AddRow(
			msg.ID[:8]+"...",
			msg.FormatFrom(30),
			msg.FormatSubject(50),
			msg.FormatDate(),
		)
	}

	if format == "plain" {
		return output.New(format).Print(table.ToPlain())
	}

	return table.Render(cmd.OutOrStdout())
}

func runMailFolders(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	folders, err := mail.ListFolders(ctx, client.Graph())
	if err != nil {
		return err
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(folders)
	}

	table := output.NewTable("ID", "NAME", "TOTAL", "UNREAD")
	for _, f := range folders {
		table.AddRow(
			f.ID[:8]+"...",
			f.DisplayName,
			fmt.Sprintf("%d", f.TotalItemCount),
			fmt.Sprintf("%d", f.UnreadItemCount),
		)
	}

	if format == "plain" {
		return output.New(format).Print(table.ToPlain())
	}

	return table.Render(cmd.OutOrStdout())
}

func runMailSend(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bodyType := "text"
	if mailHTML {
		bodyType = "html"
	}

	opts := mail.SendOptions{
		To:         mailTo,
		Cc:         mailCc,
		Bcc:        mailBcc,
		Subject:    mailSubject,
		Body:       mailBody,
		BodyType:   bodyType,
		SaveToSent: true,
	}

	if err := mail.SendMessage(ctx, client.Graph(), opts); err != nil {
		return err
	}

	fmt.Println("Message sent successfully")
	return nil
}

func runMailDraft(cmd *cobra.Command, args []string) error {
	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bodyType := "text"
	if mailHTML {
		bodyType = "html"
	}

	opts := mail.SendOptions{
		To:       mailTo,
		Subject:  mailSubject,
		Body:     mailBody,
		BodyType: bodyType,
	}

	draft, err := mail.CreateDraft(ctx, client.Graph(), opts)
	if err != nil {
		return err
	}

	format := GetOutputFormat()
	if format == "json" {
		return output.New(format).Print(draft)
	}

	fmt.Printf("Draft created: %s\n", draft.ID)
	return nil
}

func runMailMove(cmd *cobra.Command, args []string) error {
	messageID := args[0]
	folderID := args[1]

	client, err := getGraphClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := mail.MoveMessage(ctx, client.Graph(), messageID, folderID); err != nil {
		return err
	}

	fmt.Println("Message moved successfully")
	return nil
}
