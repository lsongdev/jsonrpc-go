package common

// Transport defines the interface for JSON-RPC message transport.
// Different implementations support different communication patterns:
//   - Stdio/SSE/WebSocket: Full-duplex, use Send() + Recv()
//   - HTTP: Request-response, use Call() for synchronous calls
type Transport interface {
	// Send sends a JSON-RPC request or notification.
	// For request-response transports (HTTP), this may block until response is received.
	Send(msg any) error
	// Recv receives a JSON-RPC response.
	// For request-response transports (HTTP), this returns the response from the last Send.
	Recv() ([]byte, error)
	// Close closes the transport.
	Close() error
}
