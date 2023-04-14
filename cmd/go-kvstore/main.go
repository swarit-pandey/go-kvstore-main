package main

import (
	"log"
	"net/http"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
	"github.com/sprectza/go-kvstore/pkg/api"
)

func main() {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	handler := api.NewHandler(kvs, qs)

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}
