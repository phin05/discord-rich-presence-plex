package exc

import (
	"drpp/internal/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
)

var codeSC = map[string]int{
	codeInternal:  http.StatusInternalServerError,
	codeMalformed: http.StatusBadRequest,
	codeInvalid:   http.StatusBadRequest,
}

func WithErrorHandling(next func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			handleError(w, r, err)
		}
	}
}

func WithPanicRecovery(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				handleError(w, r, fmt.Errorf("panic: %v", err))
				fmt.Print(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, r)
	}
}

type errorResponse struct {
	Error *exc `json:"error"`
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	exc, ok := err.(*exc)
	if !ok {
		exc = internal(err)
	}
	if exc.err != nil {
		logger.Error(exc.err, "%s %s - %s", r.Method, r.URL.Path, exc.Code)
	}
	w.Header().Set("Content-Type", "application/json")
	sc, ok := codeSC[exc.Code]
	if !ok {
		sc = http.StatusInternalServerError
	}
	w.WriteHeader(sc)
	json.NewEncoder(w).Encode(&errorResponse{exc})
}
