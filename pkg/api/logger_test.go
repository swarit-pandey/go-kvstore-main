package kvstore

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func Test_loggingMiddleware_BQPop(t *testing.T) {
	type fields struct {
		logger log.Logger
		next   Service
	}
	type args struct {
		key     string
		timeout time.Duration
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue interface{}
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := loggingMiddleware{
				logger: tt.fields.logger,
				next:   tt.fields.next,
			}
			gotValue, err := mw.BQPop(tt.args.key, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("loggingMiddleware.BQPop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("loggingMiddleware.BQPop() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}
