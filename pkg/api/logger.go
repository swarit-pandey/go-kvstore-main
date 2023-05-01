package kvstore

/* import (
	"time"

	"github.com/go-kit/kit/log"
)

type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

func NewLoggingMiddleware(logger log.Logger, next Service) Service {
	return &loggingMiddleware{
		logger: logger,
		next:   next,
	}
}

func (mw loggingMiddleware) Set(key string, value string, expiresAt time.Time, condition string) (err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "Set",
			"key", key,
			"value", value,
			"expiresAt", expiresAt,
			"condition", condition,
			"took", time.Since(begin),
		)
	}(time.Now())

	err = mw.next.Set(key, value, expiresAt, condition)
	return
}

func (mw loggingMiddleware) Get(key string) (value string, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "Get",
			"key", key,
			"value", value,
			"took", time.Since(begin),
		)
	}(time.Now())

	value, err = mw.next.Get(key)
	return
}

func (mw loggingMiddleware) QPush(key string, values ...interface{}) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "QPush",
			"key", key,
			"values", values,
			"took", time.Since(begin),
		)
	}(time.Now())

	mw.next.QPush(key, values...)
}

func (mw loggingMiddleware) QPop(key string) (value interface{}, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "QPop",
			"key", key,
			"value", value,
			"took", time.Since(begin),
		)
	}(time.Now())

	value, err = mw.next.QPop(key)
	return
}

func (mw loggingMiddleware) BQPop(key string, timeout time.Duration) (value interface{}, err error) {
	defer func(begin time.Time) {
		_ = mw.logger.Log(
			"method", "BQPop",
			"key", key,
			"value", value,
			"timeout", timeout,
			"took", time.Since(begin),
		)
	}(time.Now())

	value, err = mw.next.BQPop(key, timeout)
	return
}
*/
