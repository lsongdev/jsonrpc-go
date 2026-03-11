package common

import (
	"encoding/json"
	"fmt"
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string `json:"jsonrpc"`          // must always be 2.0
	ID      any    `json:"id"`               // should be a string, number or null.
	Method  string `json:"method"`           // the method being called
	Params  any    `json:"params,omitempty"` // the params for the method being called
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`          // must always be 2.0
	ID      any             `json:"id"`               // the id passed in the request object
	Result  json.RawMessage `json:"result,omitempty"` // required when the request is successful
	Error   *Error          `json:"error,omitempty"`  // required when the request is a failure
}

// Notification represents a JSON-RPC 2.0 notification (no ID).
type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Error is a JSON RPC Spec Error Object that hold the error code and the message.
// Spec - https://www.jsonrpc.org/specification#error_object
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

var (
	ParseError               = NewError(-32700, "Invalid JSON was received by the server")
	InvalidRequest           = NewError(-32600, "The JSON sent is not a valid Request object")
	MethodNotFound           = NewError(-32601, "The method does not exist / is not available")
	InvalidMethodParam       = NewError(-32602, "Invalid method parameter(s)")
	InternalError            = NewError(-32602, "Internal JSON-RPC error")
	ExecutionTimeoutError    = NewError(-32001, "Execution Timeout")
	RequestBodyIsEmpty       = NewError(-32002, "Request body is empty")
	RequestBodyTooLargeError = NewError(-32003, "Request body too large")
	InvalidRpcVersion        = NewError(-32004, "Invalid RPC Version")
)

// NewError creates and Error from a code and a message
func NewError(code int, msg string) Error {
	return Error{
		Code:    code,
		Message: msg,
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("rpc error [%d] %s", e.Code, e.Message)
}
