package xlog

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/erickxeno/logs"
)

// mockResponseWriter 实现http.ResponseWriter接口
type mockResponseWriter struct {
	header http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {}

func TestLogFromHttp(t *testing.T) {
	w := &mockResponseWriter{
		header: make(http.Header),
	}
	req := &http.Request{
		Header: make(http.Header),
	}
	xl := New(w, req)
	xl.WithContext(context.Background())
	xl.Info("test")
}

func TestLogs(t *testing.T) {
	xl := NewWith("XlogTest")
	t.Run("test xlog func", func(t *testing.T) {
		defer logs.Flush()

		SetOutputLevel(logs.DebugLevel)

		xl.Trace("test trace")
		xl.Debug("test debug")
		xl.Info("test info")
		xl.Warn("test warn")
		xl.Error("test error")
		//xl.Fatal("test fatal")

		xl.Tracef("test tracef %s", "formated")
		xl.Debugf("test debugf %s", "formated")
		xl.Infof("test infof %s", "formated")
		xl.Warnf("test warnf %s", "formated")
		xl.Errorf("test errorf %s", "formated")
		//xl.Fatalf("test fatalf %s", "format")

		xl.TraceStr("test trace str")
		xl.DebugStr("test debug str")
		xl.InfoStr("test info str")
		xl.WarnStr("test warn str")
		xl.ErrorStr("test error str")
		//xl.FatalStr("test fatal str")

		SetOutputLevel(logs.TraceLevel)
		xl.Trace("test trace 2 ")
		xl.Debug("test debug 2 ")
	})
	xl.Trace("test trace")
	// t.Run("test xlog panic", func(t *testing.T) {
	// 	xl := NewWith("XlogTest")
	// 	xl.Panic("test panic")
	// })
}

func testCallDepth(xl *Logger) {
	xl.Info("test trace")
}
func TestSetCallDepth(t *testing.T) {
	xl := NewWith("XlogTest")
	testCallDepth(xl)

	SetCallDepth(1)
	defer SetCallDepth(0)
	testCallDepth(xl)
}

func testContext(ctx context.Context) {
	xl := FromContextSafe(ctx)
	xl.Info("test trace")
}
func TestContext(t *testing.T) {
	xl := NewWith("XlogTest")
	xl.Info("test TestContext")
	ctx := context.Background()
	ctx = NewContext(ctx, xl)
	testContext(ctx)

	ctx = context.Background()
	testContext(ctx)
}

func TestLoggerConcurrency(t *testing.T) {
	// 测试同一个Logger实例在多个goroutine中的并发安全性
	t.Run("test same logger instance", func(t *testing.T) {
		defer logs.Flush()
		xl := NewWith("ConcurrencyTest")
		SetOutputLevel(logs.DebugLevel)

		// 使用WaitGroup来等待所有goroutine完成
		var wg sync.WaitGroup
		goroutineCount := 100

		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				xl.Infof("goroutine %d logging", id)
				xl.Debugf("goroutine %d debug", id)
				xl.Warnf("goroutine %d warning", id)
				xl.Errorf("goroutine %d error", id)
			}(i)
		}
		wg.Wait()
	})

	// 测试不同Logger实例在多个goroutine中的并发安全性
	t.Run("test different logger instances", func(t *testing.T) {
		defer logs.Flush()
		SetOutputLevel(logs.DebugLevel)

		var wg sync.WaitGroup
		goroutineCount := 100

		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				xl := NewWith(fmt.Sprintf("LoggerInstance%d", id))
				xl.Infof("goroutine %d logging with new logger", id)
				xl.Debugf("goroutine %d debug with new logger", id)
				xl.Warnf("goroutine %d warning with new logger", id)
				xl.Errorf("goroutine %d error with new logger", id)
			}(i)
		}
		wg.Wait()
	})

	// 测试Logger的context操作在多协程环境下的安全性
	t.Run("test logger context operations", func(t *testing.T) {
		defer logs.Flush()
		SetOutputLevel(logs.DebugLevel)

		baseCtx := context.Background()
		xl := NewWith("ContextTest")
		ctx := NewContext(baseCtx, xl)

		var wg sync.WaitGroup
		goroutineCount := 100

		for i := 0; i < goroutineCount; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				// 从context中获取logger
				logger, ok := FromContext(ctx)
				if !ok {
					t.Errorf("Failed to get logger from context in goroutine %d", id)
					return
				}
				logger.Infof("goroutine %d logging from context", id)

				// 创建新的context并绑定logger
				newCtx := NewContext(context.Background(), logger)
				newLogger := FromContextSafe(newCtx)
				newLogger.Infof("goroutine %d logging with new context", id)
			}(i)
		}
		wg.Wait()
	})
}
