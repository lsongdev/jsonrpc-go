package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lsongdev/jsonrpc-go/jsonrpc/common"
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

var defaultReq = common.Request{JSONRPC: Version}

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

type RequestParams struct {
	Payload []byte
}

type methodResp struct {
	err  error
	resp any
}

func errorResponse(req *common.Request, err error) common.Response {
	res := common.Response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Error:   &common.Error{},
	}
	switch err.(type) {
	case common.Error:
		e := err.(common.Error)
		res.Error.Code = e.Code
		res.Error.Message = e.Message
		if res.Error.Code == common.InvalidRpcVersion.Code {
			res.JSONRPC = ""
		}
		return res
	default:
		res.Error.Code = common.InternalError.Code
		res.Error.Message = err.Error()
		return res
	}
}

func successResponse(req common.Request, body []byte) common.Response {
	return common.Response{
		JSONRPC: req.JSONRPC,
		ID:      req.ID,
		Result:  body,
	}
}

func writeResponse(w http.ResponseWriter, response any) {
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(response)
}

func parseRequest(payload map[string]any) common.Request {
	req := defaultReq
	version, ok := payload["jsonrpc"]
	if ok {
		if v, ok := version.(string); ok {
			req.JSONRPC = v
		} else {
			req.JSONRPC = ""
		}
	}
	req.ID = payload["id"]
	req.Method = payload["method"].(string)
	req.Params = payload["params"]
	return req
}
