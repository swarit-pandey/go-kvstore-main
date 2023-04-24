// pkg/kvstore/service.go

package kvstore

import (
	"time"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
)

type Service interface {
	Set(key, value string, expiresAt time.Time, condition string) error
	Get(key string) (string, error)
	QPush(key string, values ...interface{})
	QPop(key string) (interface{}, error)
	BQPop(key string, timeout time.Duration) (interface{}, error)
}

type service struct {
	kvs *kvstore.KVStore
	qs  *queue.Queue
}

func NewService(kvs *kvstore.KVStore, qs *queue.Queue) Service {
	return &service{kvs: kvs, qs: qs}
}

func (s *service) Set(key, value string, expiresAt time.Time, condition string) error {
	return s.kvs.Set(key, value, expiresAt, condition)
}

func (s *service) Get(key string) (string, error) {
	value, err := s.kvs.Get(key)
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (s *service) QPush(key string, values ...interface{}) {
	s.qs.Push(key, values...)
}

func (s *service) QPop(key string) (interface{}, error) {
	return s.qs.Pop(key)
}

func (s *service) BQPop(key string, timeout time.Duration) (interface{}, error) {
	return s.qs.BPop(key, timeout)
}
