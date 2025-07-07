package main

import (
	"log"
	"net/http"
	"time"
)

type Middleware struct{}

func (m Middleware) LogMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		ts := time.Now()
		next.ServeHTTP(w, r)
		te := time.Now()
		log.Printf("[%s] %q %v", r.Method, r.URL.String(), te.Sub(ts))
	}

	return http.HandlerFunc(f)
}

func (m Middleware) RecoverMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recover from panic: %+v", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}
