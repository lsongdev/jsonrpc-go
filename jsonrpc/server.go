package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const (
	Version          = "2.0"            // JSON RPC Version
	MaxBytesRead     = 1 << 20          // 1mb
	ExecutionTimeout = 15 * time.Second // execution timout
)

type Opts struct {
	// MaxBytesRead is the maximum bytes a request object can contain
	MaxBytesRead int64
	// ExecutionTimeout is the maximum time a method should execute for. If the
	// execution exceeds the timeout, and ExectutionTimeout Error is returned for
	// that request
	ExecutionTimeout time.Duration
}

var defaultReq = Request{JsonRpc: Version}

var DefaultOpts = Opts{
	MaxBytesRead:     MaxBytesRead,
	ExecutionTimeout: ExecutionTimeout,
}

// NewServer creates a new JSON RPC Server that can handle requests.
func NewServer(opts Opts) *Service {
	return NewService(opts)
}

// NewServer creates a new JSON RPC Server that can handle requests.
func NewDefaultServer() *Service {
	return NewService(DefaultOpts)
}

type RequestFunc = func(context.Context, *RequestParams) (any, error)
type RequestMap = map[string]RequestFunc

type Request struct {
	JsonRpc string `json:"jsonrpc"` // must always be 2.0
	Id      any    `json:"id"`      // should be a string, number or null.
	Method  string `json:"method"`  // the method being called
	Params  any    `json:"params"`  // the params for the method being called
}

type Response struct {
	JsonRpc string `json:"jsonrpc,omitempty"` // must always be 2.0
	Id      any    `json:"id"`                // the id passed in the request object
	Result  any    `json:"result,omitempty"`  // required when the request is successful
	Error   *Error `json:"error,omitempty"`   // required when the request is a failure
}

type RequestParams struct {
	Payload []byte
}

type methodResp struct {
	err  error
	resp any
}

func errorResponse(req *Request, err error) Response {
	res := Response{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Error:   &Error{},
	}
	switch err.(type) {
	case Error:
		e := err.(Error)
		res.Error.Code = e.Code
		res.Error.Message = e.Message
		if res.Error.Code == InvalidRpcVersion.Code {
			res.JsonRpc = ""
		}
		return res
	default:
		res.Error.Code = InternalError.Code
		res.Error.Message = err.Error()
		return res
	}
}

func successResponse(req Request, body any) Response {
	return Response{
		JsonRpc: req.JsonRpc,
		Id:      req.Id,
		Result:  body,
	}
}

func writeResponse(w http.ResponseWriter, response any) {
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(response)
}

func parseRequest(payload map[string]any) Request {
	req := defaultReq
	version, ok := payload["jsonrpc"]
	if ok {
		if v, ok := version.(string); ok {
			req.JsonRpc = v
		} else {
			req.JsonRpc = ""
		}
	}
	req.Id = payload["id"]
	req.Method = payload["method"].(string)
	req.Params = payload["params"]
	return req
}
