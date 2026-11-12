package server

import (
	"errors"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"

	"github.com/sektor-ai/material-review-starter/go/gen/auth/v1/authv1connect"
	"github.com/sektor-ai/material-review-starter/go/internal/service"
)

type Deps struct {
	Auth         *service.Auth
	CookieSecure bool
}

// New returns the HTTP handler for the whole API. Connect services are
// mounted under /api so the web app can reach them through one proxy prefix.
func New(d Deps) http.Handler {
	api := http.NewServeMux()
	api.Handle(authv1connect.NewAuthServiceHandler(NewAuthHandler(d.Auth, d.CookieSecure)))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	})
	mux.Handle("/api/", http.StripPrefix("/api", api))
	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.InfoContext(r.Context(), "request", "method", r.Method, "path", r.URL.Path, "status", rec.status)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func internal(err error) error {
	slog.Error("internal error", "err", err)
	return connect.NewError(connect.CodeInternal, errors.New("internal error"))
}
