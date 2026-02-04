package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ServiceRegistrar interface {
	Register() (string, RequestMap)
}

type Service struct {
	methodMap        map[string]func(context.Context, *RequestParams) (any, error)
	executionTimeout time.Duration
	maxBytesRead     int64
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		writeResponse(w, errorResponse(&defaultReq, RequestBodyIsEmpty))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, s.maxBytesRead)
	buf := bytes.NewBuffer([]byte{})
	n, err := io.Copy(buf, r.Body)
	if err != nil {
		if n > MaxBytesRead {
			writeResponse(w, errorResponse(&defaultReq, RequestBodyTooLargeError))
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var reqPayload any
	err = json.NewDecoder(buf).Decode(&reqPayload)
	if err != nil {
		writeResponse(w, errorResponse(&defaultReq, ParseError))
		return
	}

	switch reqPayload.(type) {
	case []any:
		payloads := reqPayload.([]any)
		resp := make([]Response, len(payloads))
		wg := sync.WaitGroup{}
		wg.Add(len(payloads))
		for i, payload := range payloads {
			go func(index int, p any) {
				defer wg.Done()
				resp[index] = s.handle(parseRequest(p.(map[string]any)))
			}(i, payload)
		}
		wg.Wait()
		writeResponse(w, resp)
	case map[string]any:
		payload := reqPayload.(map[string]any)
		writeResponse(w, s.handle(parseRequest(payload)))
	default:
		writeResponse(w, errorResponse(&defaultReq, InvalidRequest))
	}
}

func (s *Service) handle(req Request) Response {
	if req.JsonRpc != Version {
		return errorResponse(&req, InvalidRpcVersion)
	}
	if strings.TrimSpace(req.Method) == "" {
		return errorResponse(&req, InvalidMethodParam)
	}
	res, err := s.handleMethod(req)
	if err != nil {
		return errorResponse(&req, err)
	}
	return successResponse(req, res)
}

func (s *Service) handleMethod(req Request) (any, error) {
	fn, ok := s.methodMap[req.Method]
	if !ok {
		return nil, MethodNotFound
	}
	payload, err := json.Marshal(req.Params)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", InvalidRequest, err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan methodResp)
	go func() {
		params := &RequestParams{Payload: payload}
		res := methodResp{}
		res.resp, res.err = fn(ctx, params)
		result <- res
	}()
	delay := time.NewTimer(s.executionTimeout)
	select {
	case <-delay.C:
		return nil, ExecutionTimeoutError
	case r := <-result:
		if !delay.Stop() {
			<-delay.C
		}
		return r.resp, r.err
	}
}

func (s Service) AddService(services ...ServiceRegistrar) {
	for _, srv := range services {
		name, requestMap := srv.Register()
		nameFmt := "%s"
		if name != "" {
			nameFmt = "%s.%s"
		}
		for methodName, fn := range requestMap {
			s.methodMap[fmt.Sprintf(nameFmt, name, methodName)] = fn
		}
	}
}

func NewService(opts Opts) *Service {
	return &Service{
		methodMap:        map[string]func(context.Context, *RequestParams) (any, error){},
		executionTimeout: opts.ExecutionTimeout,
		maxBytesRead:     opts.MaxBytesRead,
	}
}
