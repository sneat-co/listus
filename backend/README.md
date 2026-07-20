# listus backend

Go service for listus. Module path: `github.com/sneat-co/listus/backend`
(the module is rooted here in `backend/`, not at the repo root).

Listus domain rules, storage, and application facades live here. Bot delivery
(profiles, commands, callbacks, and rendering) lives in
`github.com/sneat-co/sneat-bots/extensions/listus`; this module must not import
bot-framework packages, Sneat-Bots, or Sneat-Go.

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
