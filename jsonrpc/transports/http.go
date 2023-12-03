package transports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// HTTPTransport implements Transport using HTTP POST for JSON-RPC.
// It handles the request-response cycle synchronously.
type HTTPTransport struct {
	url      string
	client   *http.Client
	headers  map[string]string
	timeout  time.Duration
	lastResp []byte
	mu       sync.Mutex
}

// HTTPOptions holds configuration for HTTP transport.
type HTTPOptions struct {
	// Timeout for HTTP requests (default: 30s)
	Timeout time.Duration
	// Custom HTTP client (if not provided, a default client will be created)
	Client *http.Client
	// Custom headers to include in requests
	Headers map[string]string
}

// NewHTTPTransport creates a new HTTP transport with the given URL and options.
func NewHTTPTransport(url string, opts *HTTPOptions) *HTTPTransport {
	if opts == nil {
		opts = &HTTPOptions{}
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.Client == nil {
		opts.Client = &http.Client{
			Timeout: opts.Timeout,
		}
	}
	return &HTTPTransport{
		url:     url,
		client:  opts.Client,
		timeout: opts.Timeout,
		headers: opts.Headers,
	}
}

// Send sends a JSON-RPC request via HTTP POST and stores the response.
func (t *HTTPTransport) Send(msg any) error {
	resp, err := t.Call(msg)
	if err != nil {
		return err
	}
	t.mu.Lock()
	t.lastResp = resp
	t.mu.Unlock()
	return nil
}

// Recv returns the response from the last Send call.
func (t *HTTPTransport) Recv() ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.lastResp == nil {
		return nil, fmt.Errorf("no response available")
	}
	resp := t.lastResp
	t.lastResp = nil
	return resp, nil
}

func (t *HTTPTransport) Call(msg any) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// Close closes the HTTP transport.
func (t *HTTPTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}
