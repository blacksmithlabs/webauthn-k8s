package cache

import (
	"sync"

	"github.com/redis/go-redis/v9"

	"blacksmithlabs.dev/webauthn-k8s/app/config"
)

var lock = &sync.Mutex{}
var client *redis.Client

type CacheClient = redis.Client

const Nil = redis.Nil

func ConnectCache() *CacheClient {
	if client == nil {
		lock.Lock()
		defer lock.Unlock()
		if client == nil {
			// Connect to the cache
			client = redis.NewClient(&redis.Options{
				Addr:     config.GetRedisHost(),
				Password: config.GetRedisPassword(),
				DB:       0,
			})
		}
	}
	return client
}
