package zapchi

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

var (
	// SugaredLogFormat is the format the Chi logs will use when
	// a sugared Zap logger is passed. Uses fmt.Printf templating.
	SugaredLogFormat = `%s Request: {Method: %s, Path: %s, ReqID: %s, RemoteIP: %s, Protocol: %s} Response: {Status: %d, Elapsed: %s, Size: %d bytes}`
)

// Logger is a Chi middleware that logs each request recived using
// the provided Zap logger, sugared or not.
// Provide a name if you want to set the caller (`.Named()`)
// otherwise leave blank.
func Logger(l interface{}, name string) func(next http.Handler) http.Handler {
	switch logger := l.(type) {
	case *zap.Logger:
		logger = logger.Named(name)
		logger.Debug("zap.logger detected for chi")
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
				t1 := time.Now()
				defer func() {
					logger.Info("served",
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Int("status", ww.Status()),
						zap.String("reqId", middleware.GetReqID(r.Context())),
						zap.String("remoteAddr", r.RemoteAddr),
						zap.String("proto", r.Proto),
						zap.Duration("latency", time.Since(t1)),
						zap.Int("size", ww.BytesWritten()))
				}()
				next.ServeHTTP(ww, r)
			}
			return http.HandlerFunc(fn)
		}

	case *zap.SugaredLogger:
		logger = logger.Named(name)
		logger.Debug("zap.SugaredLogger logger detected for chi")
		return func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
				t1 := time.Now()
				defer func() {
					logger.Infof(SugaredLogFormat,
						// Header
						statusLabel(ww.Status()), // "200 OK"
						// Request
						r.Method,                         // Method
						r.URL.Path,                       // Path
						middleware.GetReqID(r.Context()), // RequestID (if set)
						r.RemoteAddr,                     // RemoteAddr
						r.Proto,                          // Protocol
						// Response
						ww.Status(),       // Status
						time.Since(t1),    // Elapsed
						ww.BytesWritten(), // Bytes Written
					)
				}()
				next.ServeHTTP(ww, r)
			}
			return http.HandlerFunc(fn)
		}
	default:
		// Log error and exit
		log.Fatalf("Unknown logger passed in. Please provide *Zap.Logger or *Zap.SugaredLogger")
	}
	return nil
}

func statusLabel(status int) string {
	switch {
	case status >= 100 && status < 300:
		return fmt.Sprintf("%d OK", status)
	case status >= 300 && status < 400:
		return fmt.Sprintf("%d Redirect", status)
	case status >= 400 && status < 500:
		return fmt.Sprintf("%d Client Error", status)
	case status >= 500:
		return fmt.Sprintf("%d Server Error", status)
	default:
		return fmt.Sprintf("%d Unknown", status)
	}
}
