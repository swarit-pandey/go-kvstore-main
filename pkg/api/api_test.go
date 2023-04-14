package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
)

func TestAPICalls(t *testing.T) {
	kvs := kvstore.NewKVStore()
	qs := queue.NewQueue()
	handler := NewHandler(kvs, qs)

	server := httptest.NewServer(handler)
	defer server.Close()

	// Test SET command
	testSetCommand(t, server.URL)
	testGetCommand(t, server.URL)
	testQPushCommand(t, server.URL)
	testQPopCommand(t, server.URL)
}

func testSetCommand(t *testing.T, baseURL string) {
	requestBody := `{"command": "SET key1 value1"}`
	resp, err := http.Post(baseURL+"/api", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response string
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}

	if response != "OK" {
		t.Errorf("Expected value 'OK', got '%v'", response)
	}
}

func testGetCommand(t *testing.T, baseURL string) {
	requestBody := `{"command": "GET key1"}`
	resp, err := http.Post(baseURL+"/api", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response string
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}

	if response != "value1" {
		t.Errorf("Expected value 'value1', got '%v'", response)
	}
}

func testQPushCommand(t *testing.T, baseURL string) {
	requestBody := `{"command": "QPUSH queue1 value1 value2 value3"}`
	resp, err := http.Post(baseURL+"/api", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response string
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}

	if response != "OK" {
		t.Errorf("Expected value 'OK', got '%v'", response)
	}
}

func testQPopCommand(t *testing.T, baseURL string) {
	requestBody := `{"command": "QPOP queue1"}`
	resp, err := http.Post(baseURL+"/api", "application/json", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response string
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}

	if response != "value1" {
		t.Errorf("Expected value 'value1', got '%v'", response)
	}
}
