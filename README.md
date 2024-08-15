# webauthn-k8s
A webauthn implementation built for k8s

# Manually running migrations

K8S will eventually have a migration runner that runs as well.

The docker-compose file will run the migrator once when you start it.
You can run them again with
```
docker compose restart migrations
```
or manually run commands with
```
docker compose run migrations -path=/migrations/ -database postgres://localhost:5432/database [command] [options]
```

Do use Docker natively
```
docker run -v ./database/migrations:/migrations --network host migrate/migrate \
  -path=/migrations/ -database postgres://localhost:5432/database up [2]
```

# Generating database query files

```
cd app
docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate
```
