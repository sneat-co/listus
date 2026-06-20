# listus

Listus — todo, shopping, watch and other lists. A standalone full-stack
product extracted from [sneat-apps](https://github.com/sneat-co/sneat-apps).

**License:** [AGPL-3.0](LICENSE)

## Repository layout

This repo hosts two independent toolchains in subdirectories — neither
`package.json` nor `go.mod` lives at the repo root:

| Directory | Stack | Description |
|-----------|-------|-------------|
| [`frontend/`](frontend) | Nx · Angular 21 · Ionic 8 · pnpm | The `listus-app` standalone app and the `@sneat/extension-listus-*` libraries (see [Library structure](#library-structure-extension-library-architecture-convention)) |
| [`backend/`](backend) | Go 1.26 | Backend service (scaffold — health endpoint only for now) |

## Packages

The listus Angular extension is split into three libraries by the **extension
library-architecture** convention (see
[Library structure](#library-structure-extension-library-architecture-convention)):

- **`@sneat/extension-listus-contract`** — DTOs/types + the `LISTUS_SERVICE`
  token (`frontend/libs/extensions/listus/contract`).
- **`@sneat/extension-listus-shared`** — app-facing routing, pages, components
  (`frontend/libs/extensions/listus/shared`).
- **`@sneat/extension-listus-internal`** — service implementations +
  `provideListusInternal()` (`frontend/libs/extensions/listus/internal`).

### Library structure (extension library-architecture convention)

The listus frontend follows the **extension library-architecture** convention —
an extension is split into three libraries by *runtime weight* and *visibility*,
so other repos can depend on a light **contract** instead of the full bundle, and
cross-extension calls go through dependency-inverted `InjectionToken`s rather than
direct implementation imports. The convention is defined in
[`sneat-co/sneat-libs` → `spec/features/extension-library-architecture`](https://github.com/sneat-co/sneat-libs/tree/main/spec/features/extension-library-architecture/README.md).

| Lib | nx tags | Holds | May depend on |
|-----|---------|-------|---------------|
| [`@sneat/extension-listus-contract`](frontend/libs/extensions/listus/contract) | `type:contract` | List DTOs/types/enums + the `LISTUS_SERVICE` `InjectionToken` (`IListusService`). Runtime-light — no components/services. | other contracts + foundational `@sneat/*` |
| [`@sneat/extension-listus-shared`](frontend/libs/extensions/listus/shared) | `type:shared` | The app-facing UI: routing, pages, components, space-menu. Obtains services via the `LISTUS_SERVICE` token. | `-contract` + foundational — **never `-internal`** |
| [`@sneat/extension-listus-internal`](frontend/libs/extensions/listus/internal) | `type:internal` | `ListService` + `provideListusInternal()`. Private implementation. | `-contract` / `-shared` + foundational |

The boundary matrix is enforced by `@nx/enforce-module-boundaries` in
`frontend/eslint.config.mjs` (a `type:shared → type:internal` import fails lint).
`-internal` is consumed only by the composition-root **app**, which wires
`provideListusInternal()` at bootstrap (`frontend/apps/listus-app/src/main.ts`)
to bind `LISTUS_SERVICE` to the concrete `ListService`.

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
