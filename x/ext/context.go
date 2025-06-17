package ext

import (
	"context"
)

func Get[T comparable, V any](ctx context.Context, key T) V {
	if ctx == nil {
		var zero V
		return zero
	}

	if ret, ok := ctx.Value(key).(V); ok {
		return ret
	}

	var zero V
	return zero
}

func Set[T comparable](ctx context.Context, key T, value any) context.Context {
	if ctx == nil {
		return ctx
	}
	return context.WithValue(ctx, key, value)
}
