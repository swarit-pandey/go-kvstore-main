package main

import (
	"log"
	"net/http"
	"os"

	kitlog "github.com/go-kit/kit/log"
	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
	kvstoreAPI "github.com/sprectza/go-kvstore/pkg/api"
)

func main() {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	service := kvstoreAPI.NewService(kvs, qs)

	// Instantiate the logger and wrap the service with the logging middleware
	logger := kitlog.NewLogfmtLogger(os.Stderr)
	service = kvstoreAPI.NewLoggingMiddleware(logger, service)

	endpoints := kvstoreAPI.MakeEndpoints(service)
	handler := kvstoreAPI.MakeHTTPHandler(endpoints)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}
