package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	jsonOutput  bool
	plainOutput bool
)

var rootCmd = &cobra.Command{
	Use:   "octl",
	Short: "Outlook CLI - Access Microsoft Outlook from the terminal",
	Long: `octl is a command-line interface for Microsoft Outlook.

Access your email and calendar from the terminal using Microsoft Graph API.
Supports both personal (outlook.com) and work/school (Microsoft 365) accounts.

To get started:
  1. Create an Azure app registration (see README)
  2. Run: octl auth login --client-id <your-client-id>
  3. Follow the device code flow in your browser`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&plainOutput, "plain", false, "Output in plain text (for piping)")
}

// GetOutputFormat returns the current output format based on flags
func GetOutputFormat() string {
	if jsonOutput {
		return "json"
	}
	if plainOutput {
		return "plain"
	}
	return "table"
}

// PrintError prints an error message to stderr
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
