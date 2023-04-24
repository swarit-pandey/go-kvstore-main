// pkg/kvstore/transport.go

package kvstore

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Spawn a new HTTP handler
func MakeHTTPHandler(endpoints Endpoints) http.Handler {
	r := mux.NewRouter()

	// def SET
	r.Methods("POST").Path("/api/commands/set").Handler(httptransport.NewServer(
		endpoints.SetEndpoint,
		decodeSetRequest,
		encodeResponse,
	))

	// def GET
	r.Methods("POST").Path("/api/commands/get").Handler(httptransport.NewServer(
		endpoints.GetEndpoint,
		decodeGetRequest,
		encodeResponse,
	))

	// def QPUSH
	r.Methods("POST").Path("/api/commands/qpush").Handler(httptransport.NewServer(
		endpoints.QPushEndpoint,
		decodeQPushRequest,
		encodeResponse,
	))

	// def QPOP
	r.Methods("POST").Path("/api/commands/qpop").Handler(httptransport.NewServer(
		endpoints.QPopEndpoint,
		decodeQPopRequest,
		encodeResponse,
	))

	// def BQPOP
	r.Methods("POST").Path("/api/commands/bqpop").Handler(httptransport.NewServer(
		endpoints.BQPopEndpoint,
		decodeBQPopRequest,
		encodeResponse,
	))

	return r
}

func decodeSetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeQPushRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req QPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeQPopRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req QPopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeBQPopRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req BQPopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
