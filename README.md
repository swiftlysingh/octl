# octl - Outlook CLI

A command-line interface for Microsoft Outlook. Access your email and calendar from the terminal using Microsoft Graph API.

## Features

- **Email Management**: List, read, search, send, and organize emails
- **Calendar**: View, create, and respond to calendar events
- **Multiple Output Formats**: Table (default), JSON, plain text for scripting
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Secure Authentication**: OAuth 2.0 device code flow with OS keychain storage
- **Multi-Account Support**: Works with both personal (outlook.com) and work/school (Microsoft 365) accounts

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap swiftlysingh/tap
brew install octl
```

### Download Binary

Download the latest release from the [Releases page](https://github.com/swiftlysingh/octl/releases).

Available for:
- macOS (Intel and Apple Silicon)
- Linux (amd64 and arm64)
- Windows (amd64 and arm64)

### From Source

```bash
# Clone the repository
git clone https://github.com/swiftlysingh/octl.git
cd octl

# Build (requires Go 1.21+)
make build

# Or install to GOPATH/bin
make install
```

### Requirements

- Azure App Registration (see Setup below)

## Setup

Before using octl, you need to create an Azure App Registration:

1. Go to [Azure Portal](https://portal.azure.com) → App registrations → New registration

2. Configure the registration:
   - **Name**: `octl` (or any name you prefer)
   - **Supported account types**: "Accounts in any organizational directory and personal Microsoft accounts"
   - **Redirect URI**: Leave blank (not needed for device code flow)

3. After creation, copy the **Application (client) ID**

4. Go to **Authentication** → **Advanced settings**:
   - Enable **"Allow public client flows"** → Yes

5. Go to **API permissions** → Add permissions → Microsoft Graph → Delegated permissions:
   - `User.Read`
   - `Mail.Read`
   - `Mail.ReadWrite`
   - `Mail.Send`
   - `Calendars.Read`
   - `Calendars.ReadWrite`
   - `offline_access`

6. Log in with octl:
   ```bash
   octl auth login --client-id <your-client-id>
   ```

## Usage

### Authentication

```bash
# Log in (first time - saves client ID for future use)
octl auth login --client-id <your-azure-client-id>

# Log in (after first time)
octl auth login

# Check authentication status
octl auth status

# Log out
octl auth logout
```

### Email Commands

```bash
# List recent emails
octl mail list

# List unread emails
octl mail list --unread

# List emails from a specific folder
octl mail list --folder inbox

# Read an email
octl mail read <message-id>

# Search emails
octl mail search "quarterly report"

# Send an email
octl mail send --to user@example.com --subject "Hello" --body "Message body"

# Send HTML email
octl mail send --to user@example.com --subject "Hello" --body "<h1>Hello</h1>" --html

# Create a draft
octl mail draft --to user@example.com --subject "Draft" --body "Work in progress"

# List mail folders
octl mail folders

# Move email to folder
octl mail move <message-id> <folder-id>
```

### Calendar Commands

```bash
# List upcoming events (next 7 days)
octl calendar list

# List events for the next 14 days
octl calendar list --days 14

# Show today's events
octl calendar today

# Show this week's events
octl calendar week

# View event details
octl calendar show <event-id>

# Create an event
octl calendar create --subject "Team Meeting" --start "2024-01-15T14:00:00" --duration 1h

# Create an all-day event
octl calendar create --subject "Conference" --start "2024-01-20" --all-day

# Create an online meeting with attendees
octl calendar create \
  --subject "Project Sync" \
  --start "2024-01-15T10:00:00" \
  --duration 30m \
  --online \
  --attendees alice@example.com,bob@example.com

# Respond to a meeting invitation
octl calendar respond <event-id> accept
octl calendar respond <event-id> decline --comment "Sorry, I have a conflict"
octl calendar respond <event-id> tentative

# Delete an event
octl calendar delete <event-id>
```

### Output Formats

```bash
# Default: formatted table
octl mail list

# JSON output (for scripting)
octl mail list --json

# Plain text (tab-separated, for piping)
octl mail list --plain
```

## Configuration

octl stores configuration in `~/.config/octl/`:

- `config.json` - Client ID and other settings
- `auth_record.json` - Authentication record (non-secret)

Tokens are stored securely in the OS keychain:
- **macOS**: Keychain
- **Linux**: Kernel key retention service or libsecret

### Environment Variables

- `OCTL_CLIENT_ID` - Azure App Client ID (overrides config file)

## Development

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Build for all platforms
make build-all
```

## License

MIT License - see [LICENSE](LICENSE) for details.
