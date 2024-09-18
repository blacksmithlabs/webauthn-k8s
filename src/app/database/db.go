package database

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"blacksmithlabs.dev/webauthn-k8s/app/config"
	"blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
)

var lock = &sync.Mutex{}
var pool *pgxpool.Pool

type DBConn interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Begin(context.Context) (pgx.Tx, error)
}

// Testing mocks
type dbPoolFuncType = func(ctx context.Context, connString string) (*pgxpool.Pool, error)

var dbPoolFunc dbPoolFuncType = pgxpool.New

func ConnectDb(ctx context.Context) (*pgxpool.Pool, error) {
	if pool == nil {
		lock.Lock()
		defer lock.Unlock()
		if pool == nil {
			if ctx == nil {
				ctx = context.Background()
			}
			// Connect to the database
			var err error
			pool, err = dbPoolFunc(ctx, config.GetPostgresUrl())
			if err != nil {
				return nil, err
			}
		}
	}
	return pool, nil
}

func GetCredentialsQueries(ctx context.Context) (*credentials.Queries, error) {
	conn, err := ConnectDb(ctx)
	if err != nil {
		return nil, err
	}
	return credentials.New(conn), nil
}

func CloseDb() {
	if pool != nil {
		pool.Close()
		pool = nil
	}
}
