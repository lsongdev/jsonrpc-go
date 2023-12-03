# jsonrpc-go

[![Build Status](https://github.com/bubunyo/go-rpc/workflows/test-unit/badge.svg)](https://github.com/bubunyo/go-rpc/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![GoDoc](https://godoc.org/github.com/bubunyo/go-rpc?status.svg)](https://pkg.go.dev/github.com/bubunyo/go-rpc)

A simple resource for creating JSON RPC servers and clients to comply with
the [JSON RPC Spec](https://www.jsonrpc.org/specification)

## Features

1. Bootstrap your JSON RPC Server with ease.
2. Handle multiple requests concurrently.
3. Add your own custom errors.
4. Full-featured JSON RPC Client with Call, Notify, and Batch support.
5. Multiple transport options: HTTP, SSE (Server-Sent Events), WebSocket, and Stdio.

## Example Usage

### Server Setup

1. Setup a Ping service

```go
package main

type PingService struct{}

func (s PingService) Echo(_ context.Context, req *rpc.RequestParams) (any, error) {
	return "ok", nil
}

func (s PingService) Register() (string, rpc.RequestMap) {
	return "PingService", map[string]rpc.RequestFunc{
		"Ping": s.Echo,
	}
}

func main() {
  server := rpc.NewDefaultServer()
	server.AddService(PingService{})

	mux := http.NewServeMux()
	mux.Handle("/rpc", server)
	log.Fatalln(http.ListenAndServe(":8080", mux))
```

You can consume a single ping service resource with this curl request

```shell
curl -X POST  localhost:8080/rpc \
-d '{
        "jsonrpc": "2.0",
        "id": null,
        "method": "PingService.Ping",
        "params": null
}'
```
with the following output
```shell
{
	"jsonrpc": "2.0",
	"id": null,
	"result": "ok"
}

```

or multiple requests using
```shell
~ curl -X POST  localhost:8080/rpc -d '[{
        "jsonrpc": "2.0",
        "id": null,
        "method": "PingService.Ping",
        "params": null
}, {
        "jsonrpc": "2.0",
        "id": null,
        "method": "PingService.Ping",
        "params": null
}, {
        "jsonrpc": "2.0",
        "id": null,
        "method": "PingService.Ping",
        "params": null
}]'
```
with the following output
```shell
[{
	"jsonrpc": "2.0",
	"id": null,
	"result": "ok"
}, {
	"jsonrpc": "2.0",
	"id": null,
	"result": "ok"
}, {
	"jsonrpc": "2.0",
	"id": null,
	"result": "ok"
}]
```

### Client Usage

The JSON RPC Client provides easy access to remote procedures:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lsongdev/jsonrpc-go/jsonrpc"
)

func main() {
	// Create a JSON RPC client
	client := jsonrpc.NewClient("http://localhost:8080/rpc", jsonrpc.ClientOptions{
		Timeout: 10 * time.Second,
	})

	ctx := context.Background()

	// Simple Call
	var result string
	err := client.Call(ctx, "PingService.Ping", nil, &result)
	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}
	fmt.Printf("Result: %s\n", result)

	// Call with custom ID
	var customResult string
	err = client.CallWithID(ctx, "my-custom-id", "PingService.Ping", nil, &customResult)
	if err != nil {
		log.Fatalf("Call failed: %v", err)
	}

	// Batch requests
	requests := []jsonrpc.BatchRequest{
		{JsonRpc: jsonrpc.Version, Method: "PingService.Ping", Id: 1},
		{JsonRpc: jsonrpc.Version, Method: "PingService.Ping", Id: 2},
		{JsonRpc: jsonrpc.Version, Method: "PingService.Ping", Id: 3},
	}

	responses, err := client.Batch(ctx, requests)
	if err != nil {
		log.Fatalf("Batch failed: %v", err)
	}

	// Send notification (no response expected)
	err = client.Notify(ctx, "Service.Event", map[string]any{
		"event": "user_login",
	})
	if err != nil {
		log.Printf("Notification failed: %v", err)
	}
}
```

### SSE (Server-Sent Events)

SSE transport enables server-to-client push notifications. It's ideal for real-time events, notifications, and streaming data.

#### Server Setup with SSE

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/lsongdev/jsonrpc-go/jsonrpc"
	"github.com/lsongdev/jsonrpc-go/jsonrpc/transports"
)

func main() {
	// Create JSON-RPC server
	server := jsonrpc.NewDefaultServer()

	// Create SSE handler for server-sent events
	sseHandler := transports.NewSSEHandler()
	go sseHandler.Run() // Start the broadcast loop

	// Start a goroutine to send periodic notifications
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		counter := 0
		for {
			<-ticker.C
			counter++
			// Broadcast a JSON-RPC notification to all connected SSE clients
			notif := map[string]any{
				"jsonrpc": "2.0",
				"method":  "NotificationService.OnEvent",
				"params": map[string]any{
					"event":     "heartbeat",
					"counter":   counter,
					"timestamp": time.Now().Unix(),
				},
			}
			sseHandler.Broadcast(notif)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/rpc", server)      // HTTP endpoint for request-response
	mux.Handle("/events", sseHandler) // SSE endpoint for server events

	log.Println("Server starting on :8080")
	log.Fatalln(http.ListenAndServe(":8080", mux))
}
```

#### SSE Client

```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/lsongdev/jsonrpc-go/jsonrpc/transports"
)

func main() {
	// Create SSE transport for receiving server-sent events
	sseTransport := transports.NewSSETransport(
		"http://localhost:8080/events",
		&transports.SSEOptions{
			Timeout:           30 * time.Second,
			ReconnectInterval: 3 * time.Second,
		},
	)

	// Start listening for SSE messages
	ctx := context.Background()
	sseTransport.StartConnection(ctx)

	log.Println("SSE client connected, waiting for server notifications...")

	// Continuously receive and process SSE messages
	for {
		data, err := sseTransport.Recv()
		if err != nil {
			log.Printf("recv error: %v", err)
			continue
		}

		// Parse the JSON-RPC notification
		var notif map[string]any
		json.Unmarshal(data, &notif)

		method := notif["method"].(string)
		params := notif["params"].(map[string]any)
		log.Printf("Received: %s, params: %v", method, params)
	}
}
```

#### Hybrid Client (HTTP + SSE)

For applications that need both request-response calls and server push notifications:

```go
// Use HTTP transport for RPC calls
httpClient := jsonrpc.NewHTTPClient("http://localhost:8080/rpc", nil)

// Use SSE transport for server notifications
sseTransport := transports.NewSSETransport("http://localhost:8080/events", nil)
sseTransport.StartConnection(ctx)

// Make RPC calls via HTTP
var result string
httpClient.Call("PingService.Ping", nil, &result)

// Receive server notifications via SSE
for {
	data, _ := sseTransport.Recv()
	// Process server-sent events
}
```
