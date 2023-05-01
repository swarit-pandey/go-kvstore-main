package kvstore

import (
	"sync"
	"time"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
	"github.com/sprectza/go-kvstore/tcpconnpool"
)

type Service interface {
	Set(key, value string, expiresAt time.Time, condition string) error
	Get(key string) (string, error)
	QPush(key string, values ...interface{}) error
	QPop(key string) (interface{}, error)
	BQPop(key string, timeout time.Duration) (interface{}, error)
}

type service struct {
	kvs               *kvstore.KVStore
	qs                *queue.Queue
	tcpPool           *tcpconnpool.ConnPool
	bufferedSetChan   chan *SetRequest
	bufferedQPushChan chan *QPushRequest
	onceSet           sync.Once
	onceQPush         sync.Once
}

type SetRequest struct {
	Key       string
	Value     string
	ExpiresAt time.Time
	Condition string
	ErrChan   chan error
}

type QPushRequest struct {
	Key     string
	Values  []interface{}
	ErrChan chan error
}

func NewService(kvs *kvstore.KVStore, qs *queue.Queue, tcpPool *tcpconnpool.ConnPool) Service {
	s := &service{
		kvs:               kvs,
		qs:                qs,
		tcpPool:           tcpPool,
		bufferedSetChan:   make(chan *SetRequest, 512),
		bufferedQPushChan: make(chan *QPushRequest, 512),
	}
	return s
}

func (s *service) spawnSetWorkers() {
	for i := 0; i < 256; i++ {
		go func() {
			for req := range s.bufferedSetChan {
				err := s.kvs.Set(req.Key, req.Value, req.ExpiresAt, req.Condition)
				req.ErrChan <- err
			}
		}()
	}
}

func (s *service) spawnQPushWorkers() {
	for i := 0; i < 256; i++ {
		go func() {
			for req := range s.bufferedQPushChan {
				err := s.qs.QPush(req.Key, req.Values...)
				req.ErrChan <- err
			}
		}()
	}
}

func (s *service) Set(key string, value string, expiresAt time.Time, condition string) error {
	errChan := make(chan error, 1)
	s.bufferedSetChan <- &SetRequest{Key: key, Value: value, ExpiresAt: expiresAt, Condition: condition, ErrChan: errChan}

	s.onceSet.Do(s.spawnSetWorkers)

	return <-errChan
}

func (s *service) Get(key string) (string, error) {
	value, err := s.kvs.Get(key)
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (s *service) QPush(key string, values ...interface{}) error {
	errChan := make(chan error, 1)
	s.bufferedQPushChan <- &QPushRequest{Key: key, Values: values, ErrChan: errChan}

	s.onceQPush.Do(s.spawnQPushWorkers)

	return <-errChan
}

func (s *service) QPop(key string) (interface{}, error) {
	return s.qs.Pop(key)
}

func (s *service) BQPop(key string, timeout time.Duration) (interface{}, error) {
	return s.qs.BPop(key, timeout)
}
