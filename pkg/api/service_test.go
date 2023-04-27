package kvstore

import (
	"testing"
	"time"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	svc := NewService(kvs, qs)

	err := svc.Set("key1", "value1", time.Time{}, "")
	assert.NoError(t, err)

	err = svc.Set("key2", "value2", time.Now().Add(10*time.Second), "NX")
	assert.NoError(t, err)

	err = svc.Set("key2", "value2", time.Now().Add(10*time.Second), "XX")
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	svc := NewService(kvs, qs)

	svc.Set("key1", "value1", time.Time{}, "")

	value, err := svc.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	_, err = svc.Get("nonexistent_key")
	assert.Error(t, err)
}

func TestQPushAndQPop(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	svc := NewService(kvs, qs)

	svc.QPush("queue1", "item1")
	popValue, err := svc.QPop("queue1")
	assert.NoError(t, err)
	value, ok := popValue.(string)
	assert.True(t, ok, "Expected value to be a string")
	assert.Equal(t, "item1", value)
}

func TestBQPop(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	svc := NewService(kvs, qs)

	svc.QPush("queue1", "item1")
	popValue, err := svc.BQPop("queue1", 1*time.Second)
	assert.NoError(t, err)
	value, ok := popValue.(string)
	assert.True(t, ok, "Expected value to be a string")
	assert.Equal(t, "item1", value)

	_, err = svc.BQPop("queue1", 1*time.Second)
	assert.Error(t, err)
}
