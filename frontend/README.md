# listus frontend

Nx workspace for the listus frontend: the standalone `listus-app` and the
publishable `@sneat/extension-listus-{contract,shared,internal}` libraries
(extension library-architecture convention — see the
[root README](../README.md#library-structure-extension-library-architecture-convention)).

- **Nx** 22 · **Angular** 21 · **Ionic** 8 · **pnpm**

## Setup

```bash
pnpm install
```

## Common tasks

```bash
pnpm exec nx serve listus-app          # run the app locally
pnpm exec nx build ext-listus-shared   # build a publishable tier library
pnpm exec nx run-many -t lint test build
pnpm exec nx e2e listus-app-e2e        # end-to-end tests
```

## Layout

```
frontend/
├── apps/
│   └── listus-app/                  # standalone listus.app (Ionic shell)
└── libs/
    └── extensions/listus/
        ├── contract/                # @sneat/extension-listus-contract
        ├── shared/                  # @sneat/extension-listus-shared
        └── internal/                # @sneat/extension-listus-internal
```

> Projects are generated incrementally during the extraction; see the repo
> root README for the overall plan.
