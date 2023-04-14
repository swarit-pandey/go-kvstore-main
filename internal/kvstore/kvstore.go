package kvstore

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrInvalidCondition = errors.New("invalid condition")
)

type KeyValue struct {
	Value     interface{}
	ExpiresAt time.Time
}

type KVStore struct {
	store map[string]KeyValue
	mu    sync.RWMutex
}

func NewKVStore() *KVStore {
	return &KVStore{
		store: make(map[string]KeyValue),
	}
}

func (kvs *KVStore) Set(key string, value interface{}, expiresAt time.Time, condition string) error {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()

	if condition != "" && condition != "NX" && condition != "XX" {
		return ErrInvalidCondition
	}

	_, exists := kvs.store[key]
	if condition == "NX" && exists {
		return nil
	} else if condition == "XX" && !exists {
		return nil
	}

	kvs.store[key] = KeyValue{
		Value:     value,
		ExpiresAt: expiresAt,
	}

	return nil
}

func (kvs *KVStore) Get(key string) (interface{}, error) {
	kvs.mu.RLock()
	defer kvs.mu.RLocker().Unlock()

	if keyValue, exists := kvs.store[key]; exists {
		if keyValue.ExpiresAt.IsZero() || time.Now().Before(keyValue.ExpiresAt) {
			return keyValue.Value, nil
		}
	}

	return nil, ErrKeyNotFound
}
