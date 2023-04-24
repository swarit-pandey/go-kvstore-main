package kvstore

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	SetEndpoint   endpoint.Endpoint
	GetEndpoint   endpoint.Endpoint
	QPushEndpoint endpoint.Endpoint
	QPopEndpoint  endpoint.Endpoint
	BQPopEndpoint endpoint.Endpoint
}

// Create endpoints for each service
func MakeEndpoints(s Service) Endpoints {
	return Endpoints{
		SetEndpoint:   makeSetEndpoint(s),
		GetEndpoint:   makeGetEndpoint(s),
		QPushEndpoint: makeQPushEndpoint(s),
		QPopEndpoint:  makeQPopEndpoint(s),
		BQPopEndpoint: makeBQPopEndpoint(s),
	}
}

// SET endpoint
func makeSetEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SetRequest)
		value, ok := req.Value.(string)
		if !ok {
			return SetResponse{Err: fmt.Errorf("invalid value type")}, nil
		}
		err := s.Set(req.Key, value, req.ExpiresAt, req.Condition)
		return SetResponse{Err: err}, nil
	}
}

// GET endpoint
func makeGetEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetRequest)
		value, err := s.Get(req.Key)
		return GetResponse{Value: value, Err: err}, nil
	}
}

// QPUSH endpoint
func makeQPushEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(QPushRequest)
		s.QPush(req.Key, req.Values...)
		return QPushResponse{}, nil
	}
}

// QPOP endpoint
func makeQPopEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(QPopRequest)
		value, err := s.QPop(req.Key)
		return QPopResponse{Value: value, Err: err}, nil
	}
}

// BQPOP endpoint
func makeBQPopEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(BQPopRequest)
		value, err := s.BQPop(req.Key, req.Timeout)
		return BQPopResponse{Value: value, Err: err}, nil
	}
}
