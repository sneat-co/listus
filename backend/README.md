# listus backend

Go service for listus. Module path: `github.com/sneat-co/listus/backend`
(the module is rooted here in `backend/`, not at the repo root).

> **Status: scaffold.** Only a health endpoint is implemented. Listus domain
> endpoints are intentionally deferred.

## Requirements

- Go 1.26+

## Run

```bash
go run ./cmd/listusd          # listens on :8080 (override with LISTUS_ADDR)
curl localhost:8080/health    # -> 200 {"status":"ok"}
```

## Build & test

```bash
go build ./...
go test ./...
```

## Layout

```
backend/
├── cmd/listusd/        # main package — HTTP server entrypoint
└── internal/health/    # health-check handler + test
```
