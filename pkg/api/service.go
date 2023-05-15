package kvstore

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
)

type Service interface {
	Set(key, value string, expiresAt time.Time, condition string)
	Get(key string) (string, error)
	QPush(key string, values ...interface{}) error
	QPop(key string) (interface{}, error)
	BQPop(key string, timeout time.Duration) (interface{}, error)
	FetchErrorsForSet() []error
}

type service struct {
	kvs               *kvstore.KVStore
	qs                *queue.Queue
	bufferedSetChan   chan *SetRequest
	bufferedQPushChan chan *QPushRequest
	onceSet           sync.Once
	onceQPush         sync.Once
	errorList         []error
	errorListMutex    sync.Mutex
	shards            []*kvstore.KVStore
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

const numShards = 128

func NewService(kvs *kvstore.KVStore, qs *queue.Queue) Service {
	s := &service{
		kvs:               kvs,
		qs:                qs,
		bufferedSetChan:   make(chan *SetRequest, 1000),
		bufferedQPushChan: make(chan *QPushRequest, 512),
		shards:            make([]*kvstore.KVStore, numShards),
	}

	for i := range s.shards {
		s.shards[i] = kvstore.NewKVStore()
	}

	s.onceSet.Do(s.spawnSetWorkers)
	s.onceQPush.Do(s.spawnQPushWorkers)

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

// const maxRetries = 3

func (s *service) Set(key string, value string, expiresAt time.Time, condition string) {
	go func() {
		shardIdx := shardIndex(key)
		shard := s.shards[shardIdx]

		err := shard.Set(key, value, expiresAt, condition)

		if err != nil {
			s.errorListMutex.Lock()
			s.errorList = append(s.errorList, err)
			s.errorListMutex.Unlock()
		}
	}()
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

	return <-errChan
}

func (s *service) QPop(key string) (interface{}, error) {
	return s.qs.Pop(key)
}

func (s *service) BQPop(key string, timeout time.Duration) (interface{}, error) {
	return s.qs.BPop(key, timeout)
}

func (s *service) FetchErrorsForSet() []error {
	s.errorListMutex.Lock()
	defer s.errorListMutex.Unlock()

	errors := make([]error, len(s.errorList))
	copy(errors, s.errorList)
	s.errorList = s.errorList[:0]

	return errors
}

func shardIndex(key string) int {
	return int(murmur3_32([]byte(key), 2)) % numShards
}

// Murmer3_32 implementation
func murmur3_32(data []byte, seed uint32) uint32 {
	const (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
		r1 uint32 = 15
		r2 uint32 = 13
		m  uint32 = 5
		n  uint32 = 0xe6546b64
	)

	var h uint32 = seed

	nblocks := len(data) / 4
	for i := 0; i < nblocks; i++ {
		k := binary.LittleEndian.Uint32(data[i*4:])

		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}
	tail := data[nblocks*4:]
	k1 := uint32(0)

	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (32 - r1))
		k1 *= c2
		h ^= k1
	}

	h ^= uint32(len(data))
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	return h
}
