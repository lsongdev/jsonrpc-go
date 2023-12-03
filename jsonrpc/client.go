// Package jsonrpc provides JSON-RPC 2.0 client and server functionality.
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/lsongdev/jsonrpc-go/jsonrpc/common"
)

// JSONRPCClient represents a JSON-RPC 2.0 client.
type JSONRPCClient struct {
	transport common.Transport
	idCounter atomic.Int64
}

// NewJSONRPCClient creates a new JSON-RPC client with the given transport.
func NewJSONRPCClient(transport common.Transport) *JSONRPCClient {
	return &JSONRPCClient{
		transport: transport,
	}
}

// Call sends a JSON-RPC request and waits for the response.
func (c *JSONRPCClient) Call(method string, params any, result any) error {
	id := c.idCounter.Add(1)

	req := common.Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	// Send request
	if err := c.transport.Send(&req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read response
	data, err := c.transport.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}
	var resp common.Response
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	if result != nil && resp.Result != "" {
		if err := json.Unmarshal([]byte(resp.Result), result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Notify sends a JSON-RPC notification (no response expected).
func (c *JSONRPCClient) Notify(method string, params any) error {
	notif := common.Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	return c.transport.Send(&notif)
}

// Close closes the underlying transport.
func (c *JSONRPCClient) Close() error {
	return c.transport.Close()
}
