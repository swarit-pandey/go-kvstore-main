package kvstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sprectza/go-kvstore/pkg/model"
)

type Endpoints struct {
	SetEndpoint   endpoint.Endpoint
	GetEndpoint   endpoint.Endpoint
	QPushEndpoint endpoint.Endpoint
	QPopEndpoint  endpoint.Endpoint
	BQPopEndpoint endpoint.Endpoint
}

var (
	duration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "Total time spent serving requests.",
	}, []string{"method"})
	statusCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_status_codes",
		Help: "HTTP status codes.",
	}, []string{"method", "status"})
)

func init() {
	prometheus.MustRegister(duration)
	prometheus.MustRegister(statusCounter)
}

// Create endpoints for each service
func MakeEndpoints(s Service) Endpoints {
	return Endpoints{
		SetEndpoint:   PrometheusMetricsMiddleware("set", duration, statusCounter)(makeSetEndpoint(s)),
		GetEndpoint:   PrometheusMetricsMiddleware("get", duration, statusCounter)(makeGetEndpoint(s)),
		QPushEndpoint: PrometheusMetricsMiddleware("qpush", duration, statusCounter)(makeQPushEndpoint(s)),
		QPopEndpoint:  PrometheusMetricsMiddleware("qpop", duration, statusCounter)(makeQPopEndpoint(s)),
		BQPopEndpoint: PrometheusMetricsMiddleware("bqpop", duration, statusCounter)(makeBQPopEndpoint(s)),
	}
}

// Prometheus middleware
func PrometheusMetricsMiddleware(method string, duration *prometheus.HistogramVec, statusCounter *prometheus.CounterVec) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				latency := time.Since(begin).Seconds()
				duration.WithLabelValues(method).Observe(latency)
				// fmt.Printf("Updated metric for method: %s, duration: %f\n", method, latency)

				var statusCode int

				statusCounter.WithLabelValues(method, http.StatusText(statusCode)).Inc()
			}(time.Now())

			return next(ctx, request)
		}
	}
}

// SET endpoint
func makeSetEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.SetRequest)
		value, ok := req.Value.(string)
		if !ok {
			return model.SetResponse{Err: fmt.Errorf("invalid value type")}, nil
		}
		err := s.Set(req.Key, value, req.ExpiresAt, req.Condition)
		return model.SetResponse{Err: err}, nil
	}
}

// GET endpoint
func makeGetEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.GetRequest)
		value, err := s.Get(req.Key)
		return model.GetResponse{Value: value, Err: err}, nil
	}
}

// QPUSH endpoint
func makeQPushEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.QPushRequest)
		s.QPush(req.Key, req.Values...)
		return model.QPushResponse{}, nil
	}
}

// QPOP endpoint
func makeQPopEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.QPopRequest)
		value, err := s.QPop(req.Key)
		return model.QPopResponse{Value: value, Err: err}, nil
	}
}

// BQPOP endpoint
func makeBQPopEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(model.BQPopRequest)
		value, err := s.BQPop(req.Key, req.Timeout)
		return model.BQPopResponse{Value: value, Err: err}, nil
	}
}

// Spawn a new HTTP handler
func MakeHTTPHandler(endpoints Endpoints) http.Handler {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	statusCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_status_codes",
		Help: "HTTP status codes.",
	}, []string{"method", "status"})

	options := []httptransport.ServerOption{
		// httptransport.ServerErrorEncoder(errorEncoderWrapper),
		httptransport.ServerFinalizer(serverFinalizer(statusCounter)),
	}

	// def SET
	r.Methods("POST").Path("/api/commands/set").Handler(httptransport.NewServer(
		endpoints.SetEndpoint,
		decodeSetRequest,
		encodeResponse,
		options...,
	))

	// def GET
	r.Methods("POST").Path("/api/commands/get").Handler(httptransport.NewServer(
		endpoints.GetEndpoint,
		decodeGetRequest,
		encodeResponse,
		options...,
	))

	// def QPUSH
	r.Methods("POST").Path("/api/commands/qpush").Handler(httptransport.NewServer(
		endpoints.QPushEndpoint,
		decodeQPushRequest,
		encodeResponse,
		options...,
	))

	// def QPOP
	r.Methods("POST").Path("/api/commands/qpop").Handler(httptransport.NewServer(
		endpoints.QPopEndpoint,
		decodeQPopRequest,
		encodeResponse,
		options...,
	))

	// def BQPOP
	r.Methods("POST").Path("/api/commands/bqpop").Handler(httptransport.NewServer(
		endpoints.BQPopEndpoint,
		decodeBQPopRequest,
		encodeResponse,
		options...,
	))

	return r
}

func serverFinalizer(statusCounter *prometheus.CounterVec) httptransport.ServerFinalizerFunc {
	return func(ctx context.Context, code int, r *http.Request) {
		method := r.Method
		status := fmt.Sprint(code)
		statusCounter.WithLabelValues(method, status).Inc()

		// fmt.Printf("ServerFinalizer executed, status code: %d", code)
	}
}

func decodeSetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req model.SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if err := validateSetRequest(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req model.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if err := validateGetRequest(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeQPushRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req model.QPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if err := validateQPushRequest(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeQPopRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req model.QPopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	if err := validateQPopRequest(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeBQPopRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req model.BQPopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	if err := validateBQPopRequest(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func validateSetRequest(req *model.SetRequest) error {
	if req.Key == "" {
		return errors.New("key must not be empty")
	}

	if req.Value == nil {
		return errors.New("value must not be empty")
	}

	if req.Condition != "" && req.Condition != "NX" && req.Condition != "XX" {
		return errors.New("condition must be NX, XX or emtpy")
	}

	return nil
}

func validateGetRequest(req *model.GetRequest) error {
	if req.Key == "" {
		return errors.New("key must not be empty to get value")
	}

	return nil
}

func validateQPushRequest(req *model.QPushRequest) error {
	if req.Key == "" {
		return errors.New("key must not be emtpy")
	}
	if req.Values == nil {
		return errors.New("you must set values to be pushed")
	}

	return nil
}

func validateQPopRequest(req *model.QPopRequest) error {
	if req.Key == "" {
		return errors.New("key must not be empty")
	}

	return nil
}

func validateBQPopRequest(req *model.BQPopRequest) error {
	if req.Key == "" {
		return errors.New("key must not be empty")
	}
	if req.Timeout < 0 {
		return errors.New("timeout must not empty")
	}

	return nil
}

/* func errorEncoder(_ context.Context, err error, w http.ResponseWriter) int {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	statusCode := http.StatusInternalServerError

	switch err {
	case model.ErrQueueEmpty, model.ErrKeyNotFound:
		statusCode = http.StatusNotFound
	case model.ErrInvalidValue, model.ErrInvalidExpiryTime, model.ErrInvalidCondition:
		statusCode = http.StatusBadRequest
	default:
		statusCode = http.StatusInternalServerError
	}
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
	return statusCode
}

type dummyResponseWriter struct{}

func (*dummyResponseWriter) Header() http.Header       { return http.Header{} }
func (*dummyResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (*dummyResponseWriter) WriteHeader(int)           {}

// Wrapping the error encoder, for use in makeHttp
func errorEncoderWrapper(ctx context.Context, err error, w http.ResponseWriter) {
	_ = errorEncoder(ctx, err, w)
} */
