# listus backend extraction plan

Extract the `listus` extension from `sneat-go-backend` into this repo
(`github.com/sneat-co/listus/backend`).

## Scope

**In scope (now):** the 7 self-contained extension packages.
**Deferred:** moving `listusbot` here — that depends on extracting a shared
bots-framework library (`anybot`, `cmds4anybot`, `botinitparams`, `bothelper`,
`botsettings`) out of `sneat-go`, which is tracked with the debtus work. Until
then `listusbot` stays in `sneat-go` and just imports the new module path.

## Source

`sneat-go-backend/pkg/extensions/listus/` containing:
`const4listus`, `dbo4listus`, `dto4listus`, `dal4listus`, `facade4listus`,
`api4listus`, and `module.go` (exports `Extension()`).

## Dependencies (clean)
- `github.com/sneat-co/sneat-go-core` (extension, facade, apicore, coretypes, dbmodels, validate)
- `github.com/sneat-co/sneat-core-modules/spaceus/*` (lists live under the space hierarchy)
- `github.com/dal-go/dalgo`, `github.com/strongo/*`
- **No** dependency on `sneat-go-backend` (other extensions) or `sneat-go`.

## Target layout

Place the public packages at the module root so `listusbot` imports stay shallow:

```
listus/backend/
  go.mod                     # module github.com/sneat-co/listus/backend
  const4listus/  dbo4listus/  dto4listus/  dal4listus/  facade4listus/  api4listus/
  listusext/module.go        # Extension() entrypoint (renamed from module.go)
  cmd/listusd/   internal/health/   # existing scaffold, untouched
```

New import paths become `github.com/sneat-co/listus/backend/<pkg>`.

## Steps

1. Copy the 7 packages into this repo at the layout above.
2. Set up `go.mod`: add requires for sneat-go-core, sneat-core-modules, dalgo,
   strongo; add local `replace` directives pointing at sibling checkouts
   (`../../sneat-go-core`, `../../sneat-core-modules`) for dev.
3. Fix internal imports (`sneat-go-backend/pkg/extensions/listus/...` ->
   `listus/backend/...`).
4. `go build ./... && go test ./...` in this repo until green.
5. In `sneat-go-backend`: delete `pkg/extensions/listus/`, add a require on
   `github.com/sneat-co/listus/backend`, and update
   `pkg/extensions/standard_extensions.go` to import `listusext.Extension()` from
   the new module (least-churn: keep registering it from sneat-go-backend).
   Add a local `replace` to the sibling listus checkout.
6. In `sneat-go`: repoint `pkg/bots/botprofiles/listusbot/**` imports from
   `sneat-go-backend/pkg/extensions/listus/...` to `listus/backend/...`. Add the
   require + local replace.
7. `go build ./...` in `sneat-go-backend` and `sneat-go`.

## Verification
- `go build ./... && go test ./...` green in all three repos.
- No remaining references to `sneat-go-backend/pkg/extensions/listus`.

## Notes / decisions
- **Registration ownership:** keep sneat-go-backend registering listus (step 5)
  for minimal churn. Alternative (defer): register at the app level in sneat-go
  and drop the sneat-go-backend->listus edge entirely.
- Do not commit/push automatically — leave changes for review.
