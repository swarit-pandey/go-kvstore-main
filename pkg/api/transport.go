package kvstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sprectza/go-kvstore/pkg/model"
)

// Resolve status code
type StatusCodeResolver func(error) int

type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

type Endpoints struct {
	SetEndpoint   endpoint.Endpoint
	GetEndpoint   endpoint.Endpoint
	QPushEndpoint endpoint.Endpoint
	QPopEndpoint  endpoint.Endpoint
	BQPopEndpoint endpoint.Endpoint
}

// Create endpoints for each service
func MakeEndpoints(s Service) Endpoints {
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "Total time spent serving requests.",
	}, []string{"method"})
	prometheus.MustRegister(duration)

	return Endpoints{
		SetEndpoint:   PrometheusMetricsMiddleware("set", duration, statusCounter, resolveStatusCodeFromError)(makeSetEndpoint(s)),
		GetEndpoint:   PrometheusMetricsMiddleware("get", duration, statusCounter, resolveStatusCodeFromError)(makeGetEndpoint(s)),
		QPushEndpoint: PrometheusMetricsMiddleware("qpush", duration, statusCounter, resolveStatusCodeFromError)(makeQPushEndpoint(s)),
		QPopEndpoint:  PrometheusMetricsMiddleware("qpop", duration, statusCounter, resolveStatusCodeFromError)(makeQPopEndpoint(s)),
		BQPopEndpoint: PrometheusMetricsMiddleware("bqpop", duration, statusCounter, resolveStatusCodeFromError)(makeBQPopEndpoint(s)),
	}
}

// Prometheus middleware (from request to decoding)
func PrometheusMetricsMiddleware(method string, duration *prometheus.HistogramVec, statusCounter *prometheus.CounterVec, resolver StatusCodeResolver) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				latency := time.Since(begin).Seconds()
				duration.WithLabelValues(method).Observe(latency)
				fmt.Printf("Updated metric for method: %s, duration: %f\n", method, latency)

				var statusCode int
				if response != nil {
					switch resp := response.(type) {
					case model.SetResponse:
						statusCode = resolver(resp.Err)
					case model.GetResponse:
						statusCode = resolver(resp.Err)
					case model.QPopResponse:
						statusCode = resolver(resp.Err)
					case model.BQPopResponse:
						statusCode = resolver(resp.Err)
					default:
						statusCode = http.StatusOK
					}
				} else {
					statusCode = http.StatusInternalServerError
				}

				statusCounter.WithLabelValues(method, fmt.Sprint(statusCode)).Inc()
			}(time.Now())

			return next(ctx, request)
		}
	}
}

// Finalizer middleware
func PrometheusFinalizerMiddleware(method string, counter *prometheus.CounterVec, resolveStatusCode func(error) int) func(context.Context, int, *http.Request) {
	return func(ctx context.Context, code int, r *http.Request) {
		err, _ := ctx.Value("error").(error)
		statusCode := resolveStatusCode(err)
		counter.WithLabelValues(method, strconv.Itoa(statusCode)).Inc()
	}
}

func resolveStatusCodeFromError(err error) int {
	switch err {
	case model.ErrQueueEmpty, model.ErrKeyNotFound:
		return http.StatusNotFound
	case model.ErrInvalidValue, model.ErrInvalidExpiryTime, model.ErrInvalidCondition:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func customErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var customErr *CustomError
	if errors.As(err, &customErr) {
		w.WriteHeader(customErr.Code)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
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

// Encode response
func encodeResponseWithPrometheusFinalizer(encodeResponse httptransport.EncodeResponseFunc, prometheusFinalizer httptransport.ServerFinalizerFunc, resolveStatusCodeFromError StatusCodeResolver, r *http.Request) httptransport.EncodeResponseFunc {
	return func(ctx context.Context, w http.ResponseWriter, response interface{}) error {
		err := encodeResponse(ctx, w, response)

		var statusCode int
		if response != nil {
			switch resp := response.(type) {
			case model.SetResponse:
				statusCode = resolveStatusCodeFromError(resp.Err)
			case model.GetResponse:
				statusCode = resolveStatusCodeFromError(resp.Err)
			case model.QPopResponse:
				statusCode = resolveStatusCodeFromError(resp.Err)
			case model.BQPopResponse:
				statusCode = resolveStatusCodeFromError(resp.Err)
			default:
				statusCode = http.StatusOK
			}
		} else {
			statusCode = http.StatusInternalServerError
		}

		prometheusFinalizer(ctx, statusCode, r)

		return err
	}
}

// Avoiding multiple (duplicate) registers, since it leads to server panics by using sync.Once
var statusCounter *prometheus.CounterVec

func init() {
	statusCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_status_codes",
		Help: "HTTP status codes.",
	}, []string{"method", "status"})
	prometheus.MustRegister(statusCounter)
}

func prometheusFinalizer(method string) func(context.Context, int, *http.Request) {
	return func(ctx context.Context, code int, r *http.Request) {
		err, _ := ctx.Value("error").(error)
		statusCode := resolveStatusCodeFromError(err)
		statusCounter.WithLabelValues(method, strconv.Itoa(statusCode)).Inc()
	}
}

// Spawn a new HTTP handler
func MakeHTTPHandler(endpoints Endpoints) http.Handler {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(customErrorEncoder),
	}

	// def SET
	r.Methods("POST").Path("/api/commands/set").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := httptransport.NewServer(
			endpoints.SetEndpoint,
			decodeSetRequest,
			encodeResponseWithPrometheusFinalizer(
				encodeResponse,
				prometheusFinalizer("set"),
				resolveStatusCodeFromError,
				r,
			),
			options...,
		)
		handler.ServeHTTP(w, r)
	})

	// def GET
	r.Methods("POST").Path("/api/commands/get").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := httptransport.NewServer(
			endpoints.GetEndpoint,
			decodeGetRequest,
			encodeResponseWithPrometheusFinalizer(
				encodeResponse,
				prometheusFinalizer("get"),
				resolveStatusCodeFromError,
				r,
			),
			options...,
		)
		handler.ServeHTTP(w, r)
	})

	// def QPUSH
	r.Methods("POST").Path("/api/commands/qpush").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := httptransport.NewServer(
			endpoints.QPushEndpoint,
			decodeQPushRequest,
			encodeResponseWithPrometheusFinalizer(
				encodeResponse,
				prometheusFinalizer("qpush"),
				resolveStatusCodeFromError,
				r,
			),
			options...,
		)
		handler.ServeHTTP(w, r)
	})

	// def QPOP
	r.Methods("POST").Path("/api/commands/qpop").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := httptransport.NewServer(
			endpoints.QPopEndpoint,
			decodeQPopRequest,
			encodeResponseWithPrometheusFinalizer(
				encodeResponse,
				prometheusFinalizer("qpop"),
				resolveStatusCodeFromError,
				r,
			),
			options...,
		)
		handler.ServeHTTP(w, r)
	})

	// def BQPOP
	r.Methods("POST").Path("/api/commands/bqpop").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := httptransport.NewServer(
			endpoints.BQPopEndpoint,
			decodeBQPopRequest,
			encodeResponseWithPrometheusFinalizer(
				encodeResponse,
				prometheusFinalizer("bqpop"),
				resolveStatusCodeFromError,
				r,
			),
			options...,
		)
		handler.ServeHTTP(w, r)
	})

	return r
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

// decodeQPushRequest decodes the incoming HTTP request into a QPushRequest.
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

// decodeQPopRequest decodes the incoming HTTP request into a QPopRequest.
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

// decodeBQPopRequest decodes the incoming HTTP request into a BQPopRequest.
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

// validateQPushRequest checks if the QPushRequest is valid and returns an error if not.
func validateQPushRequest(req *model.QPushRequest) error {
	if req.Key == "" {
		return errors.New("key must not be emtpy")
	}
	if req.Values == nil {
		return errors.New("you must set values to be pushed")
	}

	return nil
}

// validateQPopRequest checks if the QPopRequest is valid and returns an error if not.
func validateQPopRequest(req *model.QPopRequest) error {
	if req.Key == "" {
		return errors.New("key must not be empty")
	}

	return nil
}

// validateBQPopRequest checks if the BQPopRequest is valid and returns an error if not.s
func validateBQPopRequest(req *model.BQPopRequest) error {
	if req.Key == "" {
		return errors.New("key must not be empty")
	}
	if req.Timeout < 0 {
		return errors.New("timeout must not empty")
	}

	return nil
}
