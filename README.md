# ZapChi [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/leosunmo/zapchi)

Logging middleware for Chi using the Zap logging library from Uber.

Can take either flat or sugared logger, named or unnamed.

The Zap log level used will depend on the status code returned by
the response.

## Installation
```
go get github.com/leosunmo/zapchi
```

## Usage
```go
package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/leosunmo/zapchi"
	"go.uber.org/zap"
)

func main() {
	// Logger with caller disabled as it will only display zapchi.go
	logger, _ := zap.NewProduction(zap.WithCaller(false))
	defer logger.Sync() // Flush buffer

	// Service
	r := chi.NewRouter()

	// Panic recovery should happen first
	r.Use(middleware.Recoverer)

	// Request ID should be before logger
	r.Use(middleware.RequestID)

	// Normal Zap logger
	r.Use(zapchi.Logger(logger, ""))

	// Or Sugared logger
	// r.Use(zapchi.Logger(logger.Sugar()))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})

	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("panicing!")
	})

	r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte("info level"))
	})

	r.Get("/warn", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte("warn level"))
	})

	r.Get("/err", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("err level"))
	})

	http.ListenAndServe(":8000", r)
}
```