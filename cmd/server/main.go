package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ainsleyclark/hookrelay/internal/store"
	"github.com/ainsleyclark/hookrelay/internal/webhook"
)

func main() {
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           routes(store.NewMemory()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func routes(st store.Store) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /webhooks/{provider}", webhook.NewHandler(st))
	return mux
}
