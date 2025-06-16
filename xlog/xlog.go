// 对logs的再次包装
package xlog

import (
	"context"
	"net/http"

	"github.com/erickxeno/logs"
)

var (
	reqidKey = logs.LogIDCtxKey
)

func SetReqidKey(key string) {
	//reqidKey = logs.SetNewLogIDCtxKey(key)
}

func SetLogKey(key string) {
	//logKey = key
}

type reqIder interface {
	ReqId() string
}

type header interface {
	Header() http.Header
}

type contexter interface {
	Context() context.Context
}

type Logger struct {
	ctx   context.Context
	reqId string
}

// NewWith  create a logger with:
//   - provided req id (if @a is reqIder)
//   - provided context (if @a is contexter)
func NewWith(a interface{}) *Logger {
	if a == nil {
		l := &Logger{
			reqId: genReqId(),
		}
		l.WithContext(context.Background())
		return l
	}
	l, ok := a.(*Logger)
	if ok {
		return l
	}

	var ctx context.Context
	if g, ok := a.(contexter); ok {
		// 仅使用context, 不向下挖掘其他字段
		ctx = g.Context()
	}
	if c, ok := a.(context.Context); ok {
		ctx = c
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var reqId string
	if id, ok := ctx.Value(reqidKey).(string); ok {
		reqId = id
	}
	if g, ok := a.(string); ok {
		reqId = g
	}
	if g, ok := a.(reqIder); ok {
		reqId = g.ReqId()
	}
	if reqId == "" {
		reqId = genReqId()
	}
	// 如果context中不存在reqidKey，则赋值为reqidKey
	if _, ok := ctx.Value(reqidKey).(string); !ok {
		ctx = context.WithValue(ctx, reqidKey, reqId)
	}

	l = &Logger{
		reqId: reqId,
	}
	l.WithContext(ctx)
	return l
}

// New born a Logger instance from request, responseWriter
//   - use reqidKey header in request, if exist
//   - use req.Context as Logger.Context
//   - bind logger to req and update req.Context
func New(w http.ResponseWriter, req *http.Request) *Logger {
	reqId := req.Header.Get(string(reqidKey))
	if reqId == "" {
		if id, ok := req.Context().Value(reqidKey).(string); ok {
			reqId = id
		}
	}
	if reqId == "" {
		reqId = genReqId()
		req.Header.Set(string(reqidKey), reqId)
	}

	l := &Logger{
		reqId: reqId,
	}
	l.WithContext(req.Context())
	return l
}

// NewWithReq born a Logger instance from request, responseWriter
//   - use reqidKey header in req, if exist
//   - use req.Context as Logger.Context
func NewWithReq(req *http.Request) *Logger {
	reqId := req.Header.Get(string(reqidKey))
	if reqId == "" {
		if id, ok := req.Context().Value(reqidKey).(string); ok {
			reqId = id
		}
	}
	if reqId == "" {
		reqId = genReqId()
		req.Header.Set(string(reqidKey), reqId)
	}

	l := &Logger{
		reqId: reqId,
	}
	l.WithContext(req.Context())
	return l
}

// Born a logger with:
//  1. new random req id
func NewDummy() *Logger {
	return NewDummyWithCtx(context.Background())
}

// Born a logger with:
//  1. new random req id
//  2. provided ctx
func NewDummyWithCtx(ctx context.Context) *Logger {
	var reqId string
	if id, ok := ctx.Value(reqidKey).(string); ok {
		reqId = id
	}
	if reqId == "" {
		reqId = genReqId()
	}
	l := &Logger{
		reqId: reqId,
	}
	l.WithContext(ctx)
	return l
}

// Spawn a child logger with: same req id with parent
// 如果需要继承xlog 的生命周期，请使用 SpawnWithCtx
func (xlog *Logger) Spawn() *Logger {
	l := &Logger{
		reqId: xlog.reqId,
	}
	l.WithContext(context.Background())
	return l
}

// Spawn a child logger with:
//   - same req id with parent
//   - same ctx with parent
//     Warning: 调用这个方法，新的 routine 将会继承上游请求的生命周期——上游请求关闭，该 routine 也会关闭。
//     因此，一定要保证新的 routine 在主 routine 之前结束(可以考虑使用 waitgroup 来实现)。
//     此外，即使两个 routine 看似"同时"关闭，比如使用 multiwriter，也可能会产生问题，因为multiwriter 实际上还是有先后。
func (xlog *Logger) SpawnWithCtx() *Logger {
	l := &Logger{
		reqId: xlog.reqId,
	}
	ctx := NewContext(xlog.ctx, l)
	l.WithContext(ctx)
	return l
}

func (xlog *Logger) ReqId() string {
	return xlog.reqId
}

func (xlog *Logger) Context() context.Context {
	return xlog.ctx
}

// WithContext binding a context with Logger
// without new logger created
func (xlog *Logger) WithContext(ctx context.Context) {
	if ctx == nil {
		panic("nil context")
	}
	xlog.ctx = NewContext(ctx, xlog)
}

// ============================================================================

type key int

const (
	xlogKey key = 0
)

func NewContext(ctx context.Context, xl *Logger) context.Context {
	if xl == nil {
		panic("nil logger")
	}
	var newCtx = ctx
	newCtx = context.WithValue(newCtx, xlogKey, xl)
	newCtx = context.WithValue(newCtx, reqidKey, xl.ReqId())
	return newCtx
}

func NewContextWith(ctx context.Context, a interface{}) context.Context {
	return NewContext(ctx, NewWith(a))
}

func NewContextWithRW(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return NewContext(ctx, New(w, r))
}

func FromContext(ctx context.Context) (xl *Logger, ok bool) {
	xl, ok = ctx.Value(xlogKey).(*Logger)
	return
}

func FromContextSafe(ctx context.Context) (xl *Logger) {
	xl, ok := ctx.Value(xlogKey).(*Logger)
	if !ok {
		xl = NewWith(ctx)
	}
	return
}
