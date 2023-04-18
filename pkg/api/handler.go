package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/sprectza/go-kvstore/internal/kvstore"
	"github.com/sprectza/go-kvstore/internal/queue"
)

type Handler struct {
	kvs *kvstore.KVStore
	qs  *queue.Queue
}

func NewHandler(kvs *kvstore.KVStore, qs *queue.Queue) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/commands", cors.Default().Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCommand(w, r, kvs, qs)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	return mux
}

func handleCommand(w http.ResponseWriter, r *http.Request, kvs *kvstore.KVStore, qs *queue.Queue) {
	type CommandRequest struct {
		Command string `json:"command"`
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	handler := Handler{kvs: kvs, qs: qs}
	response, err := handler.processCommand(req.Command)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Command(w http.ResponseWriter, r *http.Request) {
	type CommandRequest struct {
		Command string `json:"command"`
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON paylod", http.StatusBadRequest)
		return
	}

	response, err := h.processCommand(req.Command)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) processCommand(command string) (interface{}, error) {
	parts := strings.Fields(command)

	if len(parts) == 0 {
		return nil, errors.New("empty command")
	}

	switch strings.ToUpper(parts[0]) {
	case "SET":
		if len(parts) < 3 {
			return nil, errors.New("invalid SET command")
		}

		key := parts[1]
		value := parts[2]
		var expiresAt time.Time
		condition := ""

		for i := 3; i < len(parts); i++ {
			switch strings.ToUpper(parts[i]) {
			case "EX":
				i++
				if i < len(parts) {
					expiry, err := strconv.Atoi(parts[i])
					if err != nil {
						return nil, errors.New("invalid expiry time")
					}
					expiresAt = time.Now().Add(time.Duration(expiry) * time.Second)
				} else {
					return nil, errors.New("invalid SET command")
				}
			case "NX", "XX":
				condition = strings.ToUpper(parts[i])
			default:
				return nil, errors.New("invalid SET command")
			}
		}

		err := h.kvs.Set(key, value, expiresAt, condition)
		if err != nil {
			return nil, err
		}

		return "OK", nil
	case "GET":
		if len(parts) != 2 {
			return nil, errors.New("invalid GET command")
		}

		key := parts[1]
		value, err := h.kvs.Get(key)
		if err != nil {
			return nil, err
		}
		return value, nil
	case "QPUSH":
		if len(parts) < 3 {
			return nil, errors.New("invalid QPUSH command")
		}

		key := parts[1]
		values := make([]interface{}, len(parts[2:]))
		for i, v := range parts[2:] {
			values[i] = v
		}
		h.qs.Push(key, values...)
		return "OK", nil
	case "QPOP":
		if len(parts) != 2 {
			return nil, errors.New("invalid QPOP command")
		}

		key := parts[1]
		value, err := h.qs.Pop(key)
		if err != nil {
			return nil, err
		}
		return value, nil
	case "BQPOP":
		if len(parts) != 3 {
			return nil, errors.New("invalid BQPOP command")
		}

		key := parts[1]
		timeout, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, errors.New("invalid timeout value")
		}

		value, err := h.qs.BPop(key, time.Duration(timeout*float64(time.Second)))
		if err != nil {
			return nil, err
		}
		return value, nil
	default:
		return nil, errors.New("unknown command")
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Command(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}
}
