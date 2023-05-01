package main

import (
	"log"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
	kvstoreAPI "github.com/sprectza/go-kvstore/pkg/api"
	"github.com/sprectza/go-kvstore/tcpconnpool"
)

func main() {
	os.Setenv("GOGC", "200")
	tcpPool := tcpconnpool.NewConnPool("localhost:8080", 256)
	defer tcpPool.Close()

	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	service := kvstoreAPI.NewService(kvs, qs, tcpPool)

	// Instantiate the logger and wrap the service with the logging middleware
	// logger := kitlog.NewLogfmtLogger(os.Stderr)
	// service = kvstoreAPI.NewLoggingMiddleware(logger, service)

	endpoints := kvstoreAPI.MakeEndpoints(service)
	handler := kvstoreAPI.MakeHTTPHandler(endpoints)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}
