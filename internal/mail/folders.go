package mail

import (
	"context"
	"fmt"

	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

// Folder represents a mail folder
type Folder struct {
	ID              string `json:"id"`
	DisplayName     string `json:"display_name"`
	TotalItemCount  int32  `json:"total_item_count"`
	UnreadItemCount int32  `json:"unread_item_count"`
	ParentFolderID  string `json:"parent_folder_id,omitempty"`
}

// ListFolders retrieves all mail folders
func ListFolders(ctx context.Context, client *msgraph.GraphServiceClient) ([]Folder, error) {
	result, err := client.Me().MailFolders().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	folders := make([]Folder, 0)
	for _, f := range result.GetValue() {
		folders = append(folders, convertFolder(f))
	}

	return folders, nil
}

// GetFolder retrieves a single folder by ID or well-known name
func GetFolder(ctx context.Context, client *msgraph.GraphServiceClient, folderID string) (*Folder, error) {
	f, err := client.Me().MailFolders().ByMailFolderId(folderID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	folder := convertFolder(f)
	return &folder, nil
}

// MoveMessage moves a message to a different folder
func MoveMessage(ctx context.Context, client *msgraph.GraphServiceClient, messageID, destinationFolderID string) error {
	body := users.NewItemMessagesItemMovePostRequestBody()
	body.SetDestinationId(&destinationFolderID)

	_, err := client.Me().Messages().ByMessageId(messageID).Move().Post(ctx, body, nil)
	if err != nil {
		return fmt.Errorf("failed to move message: %w", err)
	}

	return nil
}

// convertFolder converts a Graph API folder to our Folder type
func convertFolder(f models.MailFolderable) Folder {
	folder := Folder{
		ID:          safeString(f.GetId()),
		DisplayName: safeString(f.GetDisplayName()),
	}

	if count := f.GetTotalItemCount(); count != nil {
		folder.TotalItemCount = *count
	}

	if count := f.GetUnreadItemCount(); count != nil {
		folder.UnreadItemCount = *count
	}

	if parentID := f.GetParentFolderId(); parentID != nil {
		folder.ParentFolderID = *parentID
	}

	return folder
}

// Well-known folder names
const (
	FolderInbox     = "inbox"
	FolderDrafts    = "drafts"
	FolderSentItems = "sentitems"
	FolderDeleted   = "deleteditems"
	FolderJunk      = "junkemail"
	FolderArchive   = "archive"
)
