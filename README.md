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

## Development

See [`frontend/README.md`](frontend/README.md) and
[`backend/README.md`](backend/README.md) for per-stack instructions.
