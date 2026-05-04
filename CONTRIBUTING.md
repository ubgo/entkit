# Contributing to ubgo/entkit

Thanks for your interest in `ubgo/entkit`. This repository is licensed under the **Apache License 2.0**. Pull requests are welcome.

## Repository layout

`entkit` is a multi-module Go repository. Each sub-directory at the root is its own Go module with its own `go.mod`. Sub-modules are versioned independently via tags of the form `<sub-module>/vX.Y.Z`.

```
entkit/
├── jsonmap/             # column types (own go.mod each)
├── passwordtype/
├── encryptedtype/
├── ent_jsontype/        # ent field helpers (own go.mod each)
├── ent_jsonslice/
├── ent_jsonmap/
├── ent_passwordtype/
├── ent_encryptedtype/
├── examples/            # runnable demo
├── go.work              # local-dev workspace stitching all sub-modules
└── .github/workflows/   # CI runs each sub-module's tests independently
```

## Workflow

1. Open an issue first for anything beyond a tiny fix.
2. Fork + branch named after the issue: `fix/123-null-scan`, `feat/456-typed-element`.
3. Run local checks: `task ci`.
4. Use Conventional Commits for the PR title; scope with the sub-module name (e.g. `feat(passwordtype): ...`).

## Code conventions

- **Per-sub-module dep isolation.** Adding a third-party dep to one sub-module must not pull it into the others. CI enforces this via per-module `go.mod` audits.
- **Race detector mandatory.** Every test must pass under `-race`.
- **Coverage target.** ≥ 90% line coverage on column types, ≥ 80% on ent helpers.
- **Defense in depth on credential types.** Any new path that could expose a `passwordtype` hash or `encryptedtype` plaintext needs an explicit redaction. Add a test that asserts the redacted form for it.
- **Public API stability.** Each sub-module reaches v1.0.0 independently. Once it does, breaking changes require a major bump and a strong rationale.
- **No comments explaining what the code does.** Names should make that clear. Reserve comments for the *why*.

## Testing locally

The root `Taskfile.yml` runs every sub-module:

```sh
task test           # standard tests across all modules
task test:race      # race detector across all modules
task lint           # golangci-lint everywhere
task ci             # everything
```

Or test a single module:

```sh
cd passwordtype && go test -race ./...
```

## License of contributions

By submitting a pull request, you agree that your contribution is provided under the same Apache License 2.0 as the rest of the repository.
