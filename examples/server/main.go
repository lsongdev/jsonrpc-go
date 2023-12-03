package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/lsongdev/jsonrpc-go/jsonrpc"
)

type PingService struct{}

func (s PingService) Echo(_ context.Context, req *jsonrpc.RequestParams) (any, error) {
	return "ok", nil
}

func (s PingService) Register() (string, jsonrpc.RequestMap) {
	return "PingService", map[string]jsonrpc.RequestFunc{
		"Ping": s.Echo,
	}
}

func main() {
	// Create an jsonrpc server
	server := jsonrpc.NewServer(jsonrpc.Opts{
		ExecutionTimeout: 15 * time.Second, // max time a function should execute for.
		MaxBytesRead:     1 << 20,          // (1mb) - the maximum size of the total request payload
	})
	// or use the default servver with
	// server := jsonrpc.NewDefaultServer()

	server.AddService(PingService{})

	mux := http.NewServeMux()
	mux.Handle("/rpc", server)
	log.Fatalln(http.ListenAndServe(":8080", mux))
}
