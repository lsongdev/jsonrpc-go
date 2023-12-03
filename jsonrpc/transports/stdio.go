package transports

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// StdioTransport implements Transport using stdio communication.
type StdioTransport struct {
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

// NewStdioTransport creates a new stdio transport.
func NewStdioTransport(stdin io.WriteCloser, stdout io.Reader) *StdioTransport {
	return &StdioTransport{
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}
}

// Send sends a JSON-RPC message.
func (t *StdioTransport) Send(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// MCP uses newline-delimited JSON
	data = append(data, '\n')

	_, err = t.stdin.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Recv receives a JSON-RPC response.
func (t *StdioTransport) Recv() ([]byte, error) {
	line, err := t.stdout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	return line, nil
}

// Close closes the transport and waits for the command to exit.
func (t *StdioTransport) Close() error {
	if closer, ok := t.stdin.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}
