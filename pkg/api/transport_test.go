package kvstore

/* import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
	"github.com/sprectza/go-kvstore/pkg/model"
)

func TestIntegration(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	service := NewService(kvs, qs, )
	endpoints := MakeEndpoints(service)
	handler := MakeHTTPHandler(endpoints)

	server := httptest.NewServer(handler)
	defer server.Close()

	// Test Set
	setReq := model.SetRequest{
		Key:       "key1",
		Value:     "value1",
		ExpiresAt: time.Time{},
		Condition: "",
	}
	setReqBody, _ := json.Marshal(setReq)
	resp, err := http.Post(server.URL+"/api/commands/set", "application/json", bytes.NewBuffer(setReqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Get
	getReq := model.GetRequest{
		Key: "key1",
	}
	getReqBody, _ := json.Marshal(getReq)
	resp, err = http.Post(server.URL+"/api/commands/get", "application/json", bytes.NewBuffer(getReqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getResp model.GetResponse
	err = json.NewDecoder(resp.Body).Decode(&getResp)
	assert.NoError(t, err)
	assert.Equal(t, "value1", getResp.Value)

	// Test QPush
	qPushReq := model.QPushRequest{
		Key:    "queue1",
		Values: []interface{}{"item1", "item2"},
	}
	qPushReqBody, _ := json.Marshal(qPushReq)
	resp, err = http.Post(server.URL+"/api/commands/qpush", "application/json", bytes.NewBuffer(qPushReqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test QPop
	qPopReq := model.QPopRequest{
		Key: "queue1",
	}
	qPopReqBody, _ := json.Marshal(qPopReq)
	resp, err = http.Post(server.URL+"/api/commands/qpop", "application/json", bytes.NewBuffer(qPopReqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var qPopResp model.QPopResponse
	err = json.NewDecoder(resp.Body).Decode(&qPopResp)
	assert.NoError(t, err)
	assert.Equal(t, "item1", qPopResp.Value)

	// Test BQPop
	bqPopReq := model.BQPopRequest{
		Key:     "queue1",
		Timeout: 1,
	}
	bqPopReqBody, _ := json.Marshal(bqPopReq)
	resp, err = http.Post(server.URL+"/api/commands/bqpop", "application/json", bytes.NewBuffer(bqPopReqBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var bqPopResp model.BQPopResponse
	err = json.NewDecoder(resp.Body).Decode(&bqPopResp)
	assert.NoError(t, err)
	assert.Equal(t, "item2", bqPopResp.Value)
}
*/
