package rwlock

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"time"
)

var ErrFailed = errors.New("Lock acquisition failure")

type Mutex interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}

type RWMutex interface {
	Mutex
	RLock(ctx context.Context) error
	RUnlock(ctx context.Context) error
}

type Options struct {
	Value     string
	Expiry    time.Duration
	Tries     int
	OnRenewal func(r *Renewal)
}
type Renewal struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Name   string
	Value  string
	Result bool
	Err    error
}

type optsKey struct{}

type Option func(options *Options)

func WithContext(ctx context.Context, opts *Options) context.Context {
	return context.WithValue(ctx, optsKey{}, opts)
}

func FromContext(ctx context.Context) (o *Options, ok bool) {
	o, ok = ctx.Value(optsKey{}).(*Options)
	return
}

func WithValue(v string) Option {
	return func(ops *Options) {
		ops.Value = v
	}
}

func WithExpiry(expiry time.Duration) Option {
	return func(ops *Options) {
		ops.Expiry = expiry
	}
}

func WithOnRenewal(f func(renewal *Renewal)) Option {
	return func(ops *Options) {
		ops.OnRenewal = f
	}
}

// 尝试加锁次数, 默认一直尝试
func WithTries(tries int) Option {
	return func(ops *Options) {
		ops.Tries = tries
	}
}

func GetGoroutineID() int64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	var id int64
	fmt.Sscanf(string(b), "goroutine %d ", &id)
	return id
}
