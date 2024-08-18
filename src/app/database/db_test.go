package database

// TODO we need to fork https://github.com/driftprogramming/pgxpoolmock
// or find an up-to-date version that is using the uber fork of gomock
// since the google version is no longer maintained
// And that support pgx/v5

// import (
//     "context"
//     "testing"

//     "github.com/jackc/pgx/v5/pgxpool"
//     "blacksmithlabs.dev/webauthn-k8s/app/config"
//     "blacksmithlabs.dev/webauthn-k8s/shared/models/credentials"
// )

// func setupTest(t *testing.T) func(t *testing.T) {
// 	origDbPoolFunc := dbPoolFunc

// 	return func(t *testing.T) {
// 		dbPoolFunc = origDbPoolFunc
// 	}
// }

// func TestConnectDb(t *testing.T) {
//     // Mock the config.GetPostgresUrl function
//     originalGetPostgresUrl := config.GetPostgresUrl
//     config.GetPostgresUrl = func() string {
//         return "postgres://user:password@localhost:5432/testdb"
// 	}
//     defer func() { config.GetPostgresUrl = originalGetPostgresUrl }()

//     ctx := context.Background()

//     // Test that ConnectDb returns a non-nil pool
//     pool1, err := ConnectDb(ctx)
//     if err != nil {
//         t.Fatalf("ConnectDb() error = %v", err)
//     }
//     if pool1 == nil {
//         t.Errorf("ConnectDb() returned nil, expected non-nil pool")
//     }

//     // Test that ConnectDb returns the same pool instance on multiple calls
//     pool2, err := ConnectDb(ctx)
//     if err != nil {
//         t.Fatalf("ConnectDb() error = %v", err)
//     }
//     if pool1 != pool2 {
//         t.Errorf("ConnectDb() returned different instances, expected the same instance")
//     }
// }

// func TestGetCredentialsQueries(t *testing.T) {
//     // Mock the config.GetPostgresUrl function
//     originalGetPostgresUrl := config.GetPostgresUrl
//     config.GetPostgresUrl = func() string {
//         return "postgres://user:password@localhost:5432/testdb"
//     }
//     defer func() { config.GetPostgresUrl = originalGetPostgresUrl }()

//     ctx := context.Background()

//     // Test that GetCredentialsQueries returns a non-nil credentials.Queries instance
//     queries, err := GetCredentialsQueries(ctx)
//     if err != nil {
//         t.Fatalf("GetCredentialsQueries() error = %v", err)
//     }
//     if queries == nil {
//         t.Errorf("GetCredentialsQueries() returned nil, expected non-nil credentials.Queries")
//     }

//     // Test that GetCredentialsQueries handles errors from ConnectDb
//     // Mock the ConnectDb function to return an error
//     originalConnectDb := ConnectDb
//     ConnectDb = func(ctx context.Context) (*pgxpool.Pool, error) {
//         return nil, fmt.Errorf("mock error")
//     }
//     defer func() { ConnectDb = originalConnectDb }()

//     _, err = GetCredentialsQueries(ctx)
//     if err == nil {
//         t.Errorf("GetCredentialsQueries() expected error, got nil")
//     }
// }
