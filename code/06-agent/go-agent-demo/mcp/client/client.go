package client

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Client struct {
	client  *mcp.Client
	session *mcp.ClientSession
}

func New(
	ctx context.Context,
	endpoint string,
) (*Client, error) {
	mcpClient := mcp.NewClient(
		&mcp.Implementation{
			Name:    "go-agent-demo",
			Version: "v1.0.0",
		},
		nil,
	)

	transport := &mcp.StreamableClientTransport{
		Endpoint: endpoint,
	}

	session, err := mcpClient.Connect(
		ctx,
		transport,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"connect MCP server failed: %w",
			err,
		)
	}

	return &Client{
		client:  mcpClient,
		session: session,
	}, nil
}

func (c *Client) Close() error {
	if c.session == nil {
		return nil
	}

	return c.session.Close()
}

func (c *Client) ListTools(
	ctx context.Context,
) ([]*mcp.Tool, error) {
	result, err := c.session.ListTools(
		ctx,
		&mcp.ListToolsParams{},
	)

	if err != nil {
		return nil, err
	}

	return result.Tools, nil
}

func (c *Client) CallTool(
	ctx context.Context,
	name string,
	arguments map[string]any,
) (*mcp.CallToolResult, error) {
	return c.session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name:      name,
			Arguments: arguments,
		},
	)
}
