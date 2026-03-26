package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"gophermart/internal/auth"
	"gophermart/internal/handler"
	"gophermart/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func NewServer(
	userSvc *service.UserService,
	orderSvc *service.OrderService,
	balanceSvc *service.BalanceService,
	withdrawalSvc *service.WithdrawalService,
	log zerolog.Logger,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(gzipMiddleware)
	r.Use(loggingMiddleware(log))
	r.Use(auth.AuthMiddleware(jwtSecret))

	h := handler.NewHandler(userSvc, orderSvc, balanceSvc, withdrawalSvc, log, jwtSecret)
	r.Mount("/", h.Router())
	return r
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := gzip.NewWriter(w)
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func loggingMiddleware(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Msg("request")
		})
	}
}
