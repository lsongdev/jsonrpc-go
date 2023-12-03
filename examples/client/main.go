package main

import (
	"fmt"
	"log"

	"github.com/lsongdev/jsonrpc-go/jsonrpc"
	"github.com/lsongdev/jsonrpc-go/jsonrpc/transports"
)

func main() {
	// Create a JSON-RPC client over HTTP
	transport := transports.NewHTTPTransport("http://localhost:8080/rpc", nil)
	client := jsonrpc.NewJSONRPCClient(transport)
	defer client.Close()

	var result string
	err := client.Call("PingService.Ping", nil, &result)
	if err != nil {
		log.Fatalf("client call failed: %v", err)
	}
	fmt.Printf("result: %s\n", result)
}
