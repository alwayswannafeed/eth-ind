package handlers

import (
	"context"
	"net/http"

	"github.com/alwayswannafeed/eth-ind/internal/data"
	"gitlab.com/distributed_lab/logan/v3"
)

type ctxKey int

const (
	logCtxKey ctxKey = iota
	storageCtxKey 
)

func CtxLog(entry *logan.Entry) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, logCtxKey, entry)
	}
}

func Log(r *http.Request) *logan.Entry {
	return r.Context().Value(logCtxKey).(*logan.Entry)
}

func CtxStorage(entry data.Storage) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, storageCtxKey, entry)
	}
}

func Storage(r *http.Request) data.Storage {
	return r.Context().Value(storageCtxKey).(data.Storage)
}