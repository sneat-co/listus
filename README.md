# listus

Listus — todo, shopping, watch and other lists. A standalone full-stack
product extracted from [sneat-apps](https://github.com/sneat-co/sneat-apps).

**License:** [AGPL-3.0](LICENSE)

## Repository layout

This repo hosts two independent toolchains in subdirectories — neither
`package.json` nor `go.mod` lives at the repo root:

| Directory | Stack | Description |
|-----------|-------|-------------|
| [`frontend/`](frontend) | Nx · Angular 21 · Ionic 8 · pnpm | The `listus-app` standalone app and the `@sneat/extension-listus` library |
| [`backend/`](backend) | Go 1.26 | Backend service (scaffold — health endpoint only for now) |

## Packages

- **`@sneat/extension-listus`** — the listus Angular extension library
  (`frontend/libs/ext-listus`), consumed both by `listus-app` and by
  `sneat-apps`.

## Running locally

### Frontend (`listus-app`)

```bash
cd frontend
pnpm install
pnpm exec nx serve listus-app        # standalone app (Vite dev server)
pnpm exec nx run-many -t lint test build
pnpm exec nx e2e listus-app-e2e      # Playwright smoke
```

`listus-app` reuses the sneat space framework and Firebase config, so for real
**auth + list data** it needs the same backing services as `sneat-app`:
Firebase emulators (`auth :9099`, `firestore :8080`) and `sneat-go-server`
(`:4300`), all under project `local-sneat-app`. See
[`sneat-apps/docs/RUN-LOCAL.md`](https://github.com/sneat-co/sneat-apps/blob/main/docs/RUN-LOCAL.md)
for the full stack. (A dedicated listus backend is scaffolded under
[`backend/`](backend) but currently only serves `/health`.)

### Backend (scaffold)

```bash
cd backend
go run ./cmd/listusd                 # :8080  (override with LISTUS_ADDR), GET /health -> 200
go test ./...
```

See [`frontend/README.md`](frontend/README.md) and
[`backend/README.md`](backend/README.md) for more per-stack detail.
