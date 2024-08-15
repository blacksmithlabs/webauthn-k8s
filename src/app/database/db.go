package database

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"blacksmithlabs.dev/webauthn-k8s/app/config"
)

var lock = &sync.Mutex{}
var pool *pgxpool.Pool

func ConnectDb() (*pgxpool.Pool, error) {
	if pool == nil {
		lock.Lock()
		defer lock.Unlock()
		if pool == nil {
			background := context.Background()
			// Connect to the database
			var err error
			pool, err = pgxpool.New(background, config.GetPostgresUrl())
			if err != nil {
				return nil, err
			}
		}
	}
	return pool, nil
}

func CloseDb() {
	if pool != nil {
		pool.Close()
		pool = nil
	}
}
