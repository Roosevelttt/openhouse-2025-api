# openhouse-2025-api

### Windows: Scoop quickstart (install make)
Open a PowerShell terminal (v5.1+) and run:

```
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
```

Then install make:
```
scoop install make
```


## Quick start
1) Configure environment
```
cp .env.example .env
# edit DB_* (default DB: openhouse_2025)
```

2) Install tools (one-time)
```
make tools
```

3) Migrate database
```
make migrate-up
```

4) Seed UKMs
```
make seed-ukm
```

5) Run server
```
make run-server
# Server on :8080 (configurable via HTTP_PORT)
```

## Make targets
- tools: install goose CLI
- goose-status: show migration status
- migrate-up | migrate-down | migrate-redo
- migrate-create name=<snake_case>: create a new SQL migration in db/migrations
- seed-ukm: run the UKM seeder
- run-server: run the API server
- tidy: `go mod tidy`
