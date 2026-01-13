package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/pp/octl/internal/auth"
	"github.com/pp/octl/internal/config"
	"github.com/spf13/cobra"
)

var (
	clientIDFlag string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Manage authentication with Microsoft Outlook.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to Microsoft Outlook",
	Long: `Log in to Microsoft Outlook using device code flow.

You'll be prompted to open a URL in your browser and enter a code.
After authentication, your credentials are stored securely for future use.

Before first login, you need to create an Azure app registration:
1. Go to https://portal.azure.com → App registrations
2. Create a new registration with "personal + work accounts" support
3. Enable "Allow public client flows" in Authentication settings
4. Add API permissions: User.Read, Mail.Read, Mail.ReadWrite, Mail.Send, Calendars.Read, Calendars.ReadWrite
5. Copy the Application (client) ID and use it with --client-id`,
	RunE: runLogin,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out of Microsoft Outlook",
	Long:  `Log out and remove stored credentials.`,
	RunE:  runLogout,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Show the current authentication status and logged-in account.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)

	loginCmd.Flags().StringVar(&clientIDFlag, "client-id", "", "Azure app client ID (saved for future use)")
}

func runLogin(cmd *cobra.Command, args []string) error {
	clientID := clientIDFlag

	// If not provided via flag, try config/environment
	if clientID == "" {
		clientID = config.GetClientID()
	}

	if clientID == "" {
		return fmt.Errorf("client ID required. Provide via --client-id flag, OCTL_CLIENT_ID env var, or config file.\n\nTo create an Azure app:\n1. Go to https://portal.azure.com → App registrations\n2. Create a new registration\n3. Copy the Application (client) ID")
	}

	// Save client ID for future use if provided via flag
	if clientIDFlag != "" {
		if err := config.SetClientID(clientIDFlag); err != nil {
			PrintError("failed to save client ID: %v", err)
		}
	}

	mgr := auth.NewManager(clientID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Starting authentication...")

	if err := mgr.Login(ctx); err != nil {
		return err
	}

	username, _ := mgr.GetUserInfo()
	fmt.Println()
	fmt.Printf("Successfully logged in as: %s\n", username)
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	clientID := config.GetClientID()
	if clientID == "" {
		// Still try to logout even without client ID
		clientID = "dummy"
	}

	mgr := auth.NewManager(clientID)

	if err := mgr.Logout(); err != nil {
		return err
	}

	fmt.Println("Logged out successfully")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	clientID := config.GetClientID()
	if clientID == "" {
		fmt.Println("Status: Not configured")
		fmt.Println("\nNo client ID configured. Run 'octl auth login --client-id <your-id>' to configure.")
		return nil
	}

	mgr := auth.NewManager(clientID)

	if mgr.IsLoggedIn() {
		if err := mgr.LoadCredential(); err == nil {
			username, accountID := mgr.GetUserInfo()
			fmt.Println("Status: Logged in")
			fmt.Printf("Account: %s\n", username)
			if accountID != "" {
				fmt.Printf("Account ID: %s\n", accountID)
			}
		} else {
			fmt.Println("Status: Logged in")
		}
	} else {
		fmt.Println("Status: Not logged in")
		fmt.Println("\nRun 'octl auth login' to authenticate.")
	}

	return nil
}
