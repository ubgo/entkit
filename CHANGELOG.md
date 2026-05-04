# Changelog

All notable changes to this project will be documented in this file. Each sub-module is versioned independently via tags of the form `<sub-module>/vX.Y.Z`. Entries are grouped by sub-module.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and each sub-module adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Initial release — 2026-05-05

### `jsonmap`

- `JsonMap map[string]any` with `encoding/json` round-trip and explicit handling of the JSON `null` value.
- `database/sql/driver` integration via `Value()` and `Scan()` for storing the map in JSON / JSONB columns.
- gqlgen-compatible `MarshalGQL` / `UnmarshalGQL` methods detected by duck typing — no gqlgen import required.

### `passwordtype`

- `HashedPassword` struct backed by argon2id via `github.com/ubgo/crypt`.
- `New(plaintext)` and `Verify(plaintext)` for the canonical lifecycle.
- `database/sql/driver` integration via `Value()` and `Scan()`.
- Defense-in-depth redaction: `String`, `GoString`, `MarshalJSON`, `LogValue`, `MarshalGQL` all return non-leaking forms.
- `UnmarshalGQL` and `UnmarshalJSON` hash plaintext on input.

### `encryptedtype`

- `EncryptedString` SQL column type backed by `github.com/ubgo/crypt`.
- Writes use AES-256-GCM (`crypt.Sealer.Seal`).
- Reads use `crypt.OpenAuto` so both AES-256-GCM and AES-CBC ciphertexts decrypt transparently.
- `SetKey` / `SetSealer` boot wiring; `Reset` for tests.

### `ent_jsontype`, `ent_jsonslice`, `ent_jsonmap`, `ent_passwordtype`, `ent_encryptedtype`

- Per-type `Field(name)` constructor returning an `ent.Field` for the matching column type.
- `passwordtype` and `encryptedtype` helpers set `Sensitive()` so ent debug output never prints the credential.
