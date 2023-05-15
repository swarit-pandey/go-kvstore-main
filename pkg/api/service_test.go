package kvstore

/* func TestSet(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	tcpPool := tcpconnpool.NewConnPool("localhost:8080", 10)
	svc := NewService(kvs, qs, tcpPool)

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
	tcpPool := tcpconnpool.NewConnPool("localhost:8080", 10)
	svc := NewService(kvs, qs, tcpPool)

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
	tcpPool := tcpconnpool.NewConnPool("localhost:8080", 10)
	svc := NewService(kvs, qs, tcpPool)

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
	tcpPool := tcpconnpool.NewConnPool("localhost:8080", 10)
	svc := NewService(kvs, qs, tcpPool)

	svc.QPush("queue1", "item1")
	popValue, err := svc.BQPop("queue1", 1*time.Second)
	assert.NoError(t, err)
	value, ok := popValue.(string)
	assert.True(t, ok, "Expected value to be a string")
	assert.Equal(t, "item1", value)

	_, err = svc.BQPop("queue1", 1*time.Second)
	assert.Error(t, err)
} */
