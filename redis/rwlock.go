package redis

import (
	"context"
	"sync"
	"time"

	"github.com/J-guanghua/rwlock"
	"github.com/go-redis/redis/v8"
)

var rlock *rwLock

type rwLock struct {
	mtx   sync.Mutex
	pool  []*redis.Client
	mutex map[string]*rwRedis
}

func Init(options ...*redis.Options) {
	pools := []*redis.Client{}
	for _, o := range options {
		client := redis.NewClient(o)
		_, err := client.Ping(context.TODO()).Result()
		if err != nil {
			panic(err)
		}
		pools = append(pools, redis.NewClient(o))
	}
	rlock = &rwLock{
		pool:  pools,
		mutex: make(map[string]*rwRedis, 100),
	}
}

func (rlock *rwLock) allocation(name string, opts *rwlock.Options) rwlock.Mutex {
	rlock.mtx.Lock()
	defer rlock.mtx.Unlock()
	if rlock.mutex[name] == nil {
		index := len(rlock.mutex) % len(rlock.pool)
		rlock.mutex[name] = &rwRedis{
			name:   name,
			opts:   opts,
			client: rlock.pool[index],
			signal: make(chan struct{}, 1),
		}
	}
	return rlock.mutex[name]
}

func Mutex(name string, opts ...rwlock.Option) rwlock.Mutex {
	opt := &rwlock.Options{
		Expiry:    6 * time.Second,
		Value:     "default",
		OnRenewal: func(r *rwlock.Renewal) {},
	}
	for _, o := range opts {
		o(opt)
	}
	return rlock.allocation(name, opt)
}

func RWMutex(name string, opts ...rwlock.Option) rwlock.RWMutex { // nolint
	return nil
}
