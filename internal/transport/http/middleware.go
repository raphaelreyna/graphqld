package http

import (
	"context"
	"net/http"
)

type key uint

const (
	keyHeaderFunc key = iota
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, keyHeaderFunc, w.Header)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetHeader(ctx context.Context) http.Header {
	headerFunc := ctx.Value(keyHeaderFunc).(func() http.Header)
	return headerFunc()
}
