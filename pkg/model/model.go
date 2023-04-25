package model

import (
	"errors"
	"time"
)

var (
	ErrKeyNotFound       = errors.New("key not found")
	ErrQueueEmpty        = errors.New("queue is empty")
	ErrInvalidValue      = errors.New("invalid value")
	ErrInvalidExpiryTime = errors.New("invalid expiry time")
	ErrInvalidCondition  = errors.New("invalid condition")
)

// Request for SET
type SetRequest struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	Condition string
}

// Response for SET
type SetResponse struct {
	Err error
}

// Request for GET
type GetRequest struct {
	Key string
}

// Response for GET
type GetResponse struct {
	Value interface{}
	Err   error
}

// Request for PUSH in the queue
type QPushRequest struct {
	Key    string
	Values []interface{}
}

// Response for PUSH in the queue
type QPushResponse struct{}

// Request for POP from the queue
type QPopRequest struct {
	Key string
}

// Response for POP from the queue
type QPopResponse struct {
	Value interface{}
	Err   error
}

// Request for BQPOP from the queue
type BQPopRequest struct {
	Key     string
	Timeout time.Duration
}

// Response from BQPOP from the queue
type BQPopResponse struct {
	Value interface{}
	Err   error
}
