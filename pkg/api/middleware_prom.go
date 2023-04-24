package kvstore

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type prometheusMiddleware struct {
	next     Service
	duration *prometheus.HistogramVec
}

func NewPrometheusMiddleware(duration *prometheus.HistogramVec, next Service) Service {
	return &prometheusMiddleware{
		next:     next,
		duration: duration,
	}
}

func (mw *prometheusMiddleware) Set(key string, value string, expiresAt time.Time, condition string) error {
	start := time.Now()
	err := mw.next.Set(key, value, expiresAt, condition)
	mw.duration.WithLabelValues("set").Observe(time.Since(start).Seconds())
	return err
}

func (mw *prometheusMiddleware) Get(key string) (string, error) {
	start := time.Now()
	value, err := mw.next.Get(key)
	mw.duration.WithLabelValues("get").Observe(time.Since(start).Seconds())
	return value, err
}

func (mw *prometheusMiddleware) QPush(key string, values ...interface{}) {
	start := time.Now()
	mw.next.QPush(key, values...)
	mw.duration.WithLabelValues("qpush").Observe(time.Since(start).Seconds())
}

func (mw *prometheusMiddleware) QPop(key string) (interface{}, error) {
	start := time.Now()
	value, err := mw.next.QPop(key)
	mw.duration.WithLabelValues("qpop").Observe(time.Since(start).Seconds())
	return value, err
}

func (mw *prometheusMiddleware) BQPop(key string, timeout time.Duration) (interface{}, error) {
	start := time.Now()
	value, err := mw.next.BQPop(key, timeout)
	mw.duration.WithLabelValues("bqpop").Observe(time.Since(start).Seconds())
	return value, err
}
