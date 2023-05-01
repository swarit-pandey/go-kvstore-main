package kvstore

import (
	"encoding/binary"
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

type SetOperation struct {
	Key       string
	Value     KeyValue
	Condition string
}

type SetBatch struct {
	Ops []SetOperation
}

type shard struct {
	store   map[string]KeyValue
	mu      sync.RWMutex
	setChan chan SetBatch
}

type KVStore struct {
	shards []*shard
}

const (
	numShards      = 256
	batchThreshold = 200
)

func NewKVStore() *KVStore {
	shards := make([]*shard, numShards)
	for i := range shards {
		shards[i] = &shard{
			store:   make(map[string]KeyValue),
			setChan: make(chan SetBatch, batchThreshold),
		}
		go func(s *shard) {
			ops := make([]SetOperation, 0, batchThreshold)
			for batch := range s.setChan {
				ops = append(ops, batch.Ops...)

				if len(ops) >= batchThreshold {
					s.mu.Lock()
					for _, op := range ops {
						key := op.Key
						value := op.Value

						existingValue, exists := s.store[key]
						if op.Condition == "NX" && exists {
							continue
						} else if op.Condition == "XX" && !exists {
							continue
						}

						s.store[key] = value
						if !value.ExpiresAt.IsZero() && existingValue.ExpiresAt.Before(value.ExpiresAt) {
							delete(s.store, key)
						}
					}
					s.mu.Unlock()
					ops = ops[:0]
				}
			}
		}(shards[i])
	}
	return &KVStore{shards: shards}
}

func (kvs *KVStore) Set(key string, value interface{}, expiresAt time.Time, condition string) error {
	return kvs.SetBatch([]SetOperation{{
		Key:       key,
		Value:     KeyValue{Value: value, ExpiresAt: expiresAt},
		Condition: condition,
	}})
}

func (kvs *KVStore) SetBatch(ops []SetOperation) error {
	for _, op := range ops {
		shard := kvs.shards[shardIndex(op.Key)]
		shard.setChan <- SetBatch{Ops: []SetOperation{op}}
	}

	return nil
}

func (kvs *KVStore) Get(key string) (interface{}, error) {
	shard := kvs.shards[shardIndex(key)]

	shard.mu.RLock()
	entry, exists := shard.store[key]
	shard.mu.RUnlock()

	if !exists {
		return nil, ErrKeyNotFound
	}

	keyValue := entry
	if keyValue.ExpiresAt.IsZero() || time.Now().Before(keyValue.ExpiresAt) {
		return keyValue.Value, nil
	}

	return nil, ErrKeyNotFound
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
