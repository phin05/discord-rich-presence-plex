package api

import (
	"context"
	"drpp/server/logger"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

const maxReqBodyBytes = 1 * 1024 * 1024 // 1 MB

type Server struct {
	mux        *http.ServeMux
	httpServer *http.Server
	cancel     context.CancelFunc
}

func NewServer(ctx context.Context, bindAddress string, webBuildOutput fs.FS, devMode bool, allowedNetworks []string, trustedProxies []string) *Server {
	logger.Info("Web UI: http://%s", bindAddress)
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(&customFs{innerFs: webBuildOutput}))
	var handler http.Handler = mux
	if devMode {
		logger.Warning("Running in development mode, CORS enabled")
		handler = devCorsMiddleware(handler)
	}
	handler = securityMiddleware(handler)
	handler = ipCheckMiddleware(handler, allowedNetworks, trustedProxies)
	ctx, cancel := context.WithCancel(ctx)
	return &Server{
		mux: mux,
		httpServer: &http.Server{
			Addr:              bindAddress,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       5 * time.Second,
			MaxHeaderBytes:    16 * 1024, // 16 KB
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		},
		cancel: cancel,
	}
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	s.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// Leaves numeric fields as 0 if the JSON value is not a valid number
var tolerantIntUnmarshaler = json.UnmarshalFunc(func(data []byte, v *int) error {
	_ = json.Unmarshal(data, v)
	return nil
})

var (
	anyType    = reflect.TypeFor[any]()
	anyPtrType = reflect.TypeFor[*any]()
)

// https://github.com/golang/go/issues/49085
func RegisterRoute[I any, O any](s *Server, handler func(ctx context.Context, input I) (O, error), method string, path string, statusCode int) {
	if handler == nil || method == "" || path == "" || statusCode == 0 {
		logger.Error(nil, "Invalid route registration for endpoint %s %s", method, path)
		return
	}
	inputType := reflect.TypeFor[I]()
	if inputType == anyType || inputType == anyPtrType {
		inputType = nil
	}
	s.mux.HandleFunc(method+" /api/"+path, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				handleError(w, r, fmt.Errorf("panic: %v\n%s", err, strings.TrimSuffix(string(debug.Stack()), "\n")))
			}
		}()
		r.Body = http.MaxBytesReader(w, r.Body, maxReqBodyBytes)
		var input I
		if inputType != nil {
			if err := json.UnmarshalRead(r.Body, &input, json.RejectUnknownMembers(true), json.WithUnmarshalers(tolerantIntUnmarshaler)); err != nil {
				// Make the error user-facing only if it's a JSON syntax/semantic error
				_, isSynErr := errors.AsType[*jsontext.SyntacticError](err)
				_, isSemErr := errors.AsType[*json.SemanticError](err)
				if isSynErr || isSemErr {
					err = ErrBadRequest("Invalid JSON body", []string{err.Error()})
				}
				handleError(w, r, err)
				return
			}
		}
		output, err := handler(r.Context(), input)
		if err != nil {
			handleError(w, r, err)
			return
		}
		writeResponse(w, r, statusCode, output)
	})
}

func (s *Server) RegisterCustomRoute(handler http.HandlerFunc, method string, path string) {
	if handler == nil || method == "" || path == "" {
		logger.Error(nil, "Invalid custom route registration for endpoint %s %s", method, path)
		return
	}
	s.mux.HandleFunc(method+" /api/"+path, func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxReqBodyBytes)
		handler(w, r)
	})
}

type errorResponse struct {
	Error *Error `json:"error"`
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var e *Error
	if !errors.As(err, &e) {
		// If the error returned by the handler is not an *Error, log it and send a generic error instead to the user, so that internal error details are not exposed
		logger.Error(err, "%s %s", r.Method, r.URL.Path)
		e = ErrInternalServerError()
	}
	writeResponse(w, r, e.HttpStatusCode, &errorResponse{Error: e})
}

func writeResponse(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	if data == nil {
		if statusCode != http.StatusNoContent && statusCode != http.StatusAccepted {
			logger.Warning("Writing nil data for %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.MarshalWrite(w, data); err != nil {
		logger.Error(err, "Failed to write body for %s %s", r.Method, r.URL.Path)
	}
}
