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
pnpm start                              # https://listus-app.dev.localhost:4315
pnpm exec nx build ext-listus-ui       # build the public UI library
pnpm exec nx run-many -t lint test build
pnpm exec nx e2e listus-app-e2e        # end-to-end tests
```

The development server uses HTTPS on `listus-app.dev.localhost`. When Firebase
emulators are enabled, the app calls the shared Sneat platform API directly at
`https://sneat-api.dev.localhost:4300`. Run
`scripts/setup-localhost-tls` once from the Workbench repo to create and trust
the `*.dev.localhost` certificate; `pnpm start` uses its standard location.

## Layout

```
frontend/
├── apps/
│   └── listus-app/                  # standalone listus.app (Ionic shell)
└── libs/
    └── extensions/listus/
        ├── ui/                      # @sneat/extension-listus-ui
        └── runtime/                 # @sneat/extension-listus
```

> Projects are generated incrementally during the extraction; see the repo
> root README for the overall plan.
