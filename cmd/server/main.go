package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ainsleyclark/hookrelay/internal/middleware"
	"github.com/ainsleyclark/hookrelay/internal/store"
	"github.com/ainsleyclark/hookrelay/internal/webhook"
)

const devSecret = "dev-secret"

func main() {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		log.Println("WEBHOOK_SECRET not set, using the dev secret")
		secret = devSecret
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           routes(store.NewMemory(), middleware.HMACVerifier{Secret: []byte(secret)}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func routes(st store.Store, v middleware.Verifier) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /webhooks/{provider}", middleware.VerifySignature(v, webhook.NewHandler(st)))
	return mux
}
