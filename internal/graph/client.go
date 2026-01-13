package graph

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/pp/octl/internal/auth"
)

// Client wraps the Microsoft Graph client
type Client struct {
	graphClient *msgraph.GraphServiceClient
}

// NewClient creates a new Graph client with the given credential
func NewClient(credential azcore.TokenCredential) (*Client, error) {
	// Create auth provider using the credential
	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(
		credential,
		auth.Scopes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}

	// Create request adapter
	adapter, err := msgraph.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create request adapter: %w", err)
	}

	// Create Graph client
	client := msgraph.NewGraphServiceClient(adapter)

	return &Client{
		graphClient: client,
	}, nil
}

// Graph returns the underlying Microsoft Graph client
func (c *Client) Graph() *msgraph.GraphServiceClient {
	return c.graphClient
}
