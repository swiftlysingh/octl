package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/pp/octl/internal/config"
)

const (
	authRecordFileName = "auth_record.json"
)

// Scopes defines the Microsoft Graph API permissions we need
var Scopes = []string{
	"User.Read",
	"Mail.Read",
	"Mail.ReadWrite",
	"Mail.Send",
	"Calendars.Read",
	"Calendars.ReadWrite",
	"offline_access",
}

// Manager handles authentication state and credentials
type Manager struct {
	clientID   string
	credential *azidentity.DeviceCodeCredential
	record     *azidentity.AuthenticationRecord
}

// NewManager creates a new authentication manager
func NewManager(clientID string) *Manager {
	return &Manager{
		clientID: clientID,
	}
}

// authRecordPath returns the path to the auth record file
func authRecordPath() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, authRecordFileName), nil
}

// Login performs interactive device code authentication
func (m *Manager) Login(ctx context.Context) error {
	cred, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		ClientID: m.clientID,
		TenantID: "common", // Support both personal and work accounts
		UserPrompt: func(ctx context.Context, msg azidentity.DeviceCodeMessage) error {
			fmt.Println()
			fmt.Println("To sign in, use a web browser to open the page:")
			fmt.Printf("  %s\n", msg.VerificationURL)
			fmt.Println()
			fmt.Printf("Enter the code: %s\n", msg.UserCode)
			fmt.Println()
			fmt.Println("Waiting for authentication...")
			return nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	m.credential = cred

	// Trigger authentication by requesting a token
	record, err := cred.Authenticate(ctx, &policy.TokenRequestOptions{
		Scopes: Scopes,
	})
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	m.record = &record

	// Save the auth record for future silent auth
	if err := m.saveAuthRecord(); err != nil {
		return fmt.Errorf("failed to save auth record: %w", err)
	}

	return nil
}

// LoadCredential loads existing credentials for silent auth
func (m *Manager) LoadCredential() error {
	record, err := m.loadAuthRecord()
	if err != nil {
		return err
	}

	cred, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		ClientID:                       m.clientID,
		TenantID:                       "common",
		AuthenticationRecord:           *record,
		DisableAutomaticAuthentication: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	m.credential = cred
	m.record = record
	return nil
}

// GetCredential returns the current credential
func (m *Manager) GetCredential() azcore.TokenCredential {
	return m.credential
}

// GetAuthRecord returns the current auth record
func (m *Manager) GetAuthRecord() *azidentity.AuthenticationRecord {
	return m.record
}

// IsLoggedIn checks if we have valid credentials
func (m *Manager) IsLoggedIn() bool {
	if err := m.LoadCredential(); err != nil {
		return false
	}

	// Try to get a token silently
	ctx := context.Background()
	_, err := m.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: Scopes,
	})
	return err == nil
}

// Logout removes stored credentials
func (m *Manager) Logout() error {
	path, err := authRecordPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove auth record: %w", err)
	}

	m.credential = nil
	m.record = nil
	return nil
}

// saveAuthRecord saves the authentication record to disk
func (m *Manager) saveAuthRecord() error {
	if m.record == nil {
		return fmt.Errorf("no auth record to save")
	}

	path, err := authRecordPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(m.record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth record: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth record: %w", err)
	}

	return nil
}

// loadAuthRecord loads the authentication record from disk
func (m *Manager) loadAuthRecord() (*azidentity.AuthenticationRecord, error) {
	path, err := authRecordPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not logged in - run 'octl auth login' first")
		}
		return nil, fmt.Errorf("failed to read auth record: %w", err)
	}

	var record azidentity.AuthenticationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to parse auth record: %w", err)
	}

	return &record, nil
}

// GetUserInfo returns basic user information from the auth record
func (m *Manager) GetUserInfo() (username, homeAccountID string) {
	if m.record == nil {
		return "", ""
	}
	return m.record.Username, m.record.HomeAccountID
}
