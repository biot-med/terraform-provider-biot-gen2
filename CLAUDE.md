# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
make build        # go build -v ./...
make install      # build + go install (also run via build.sh for local Terraform testing)
make test         # go test -v -cover -timeout=120s -parallel=10 ./...
make testacc      # TF_ACC=1 go test -v -cover -timeout 120m ./...  (requires live BioT server)
make lint         # golangci-lint run
make fmt          # gofmt -s -w -e .
make generate     # regenerate docs via terraform-plugin-docs (run from tools/)
```

Local install for manual Terraform testing:
```sh
./build.sh        # builds and copies binary to ~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/<version>/darwin_arm64/ (version read from main.go)
```

## Architecture

This is a **Terraform Plugin Framework** provider for managing BioT configuration. It exposes `biot_template`, `biot_abac_condition` and `biot_abac_action` (with ABAC rules to follow), and connects to a BioT backend via a custom HTTP client.

### Provider (`internal/provider/provider.go`)
Configures credentials (`base_url`, `service_id`, `service_secret_key`), validates server version compatibility, and registers resources/data sources.

### Resource (`internal/resources/template/`)
All CRUD logic lives in `biot_template.go`. The schema is defined via nested helper functions (`attributeSchema`, `builtinAttributeSchema`, etc.). Two mapper files handle bidirectional conversion:
- `from_terraform_mapper.go` — Terraform model → API request
- `to_terraform_mapper.go` — API response → Terraform state

The model types live in `model.go` and are prefixed with `Terraform` (e.g. `TerraformTemplate`, `TerraformBuiltinAttribute`) to distinguish them from API types.

### API Client (`internal/api/`)
One package per BioT domain, over shared plumbing. Resources receive `*api.APIClient` and reach a domain through its field: `client.Template.Get(...)`, `client.Abac.CreateCondition(...)`.
- `client.go` — `api.New(...)` wires everything together; `APIClient` holds `Template`, `Abac`, `Versions`
- `transport/` — domain-agnostic HTTP: `transport.Do[T]` / `DoNoContent` / `DoUnauthenticated[T]`, BioT's `APIError` envelope. Attaches the bearer token itself, so API methods never handle tokens.
  - A 404 returns `ErrNotFound` **wrapping** the parsed `APIError`, so check drift with `transport.IsNotFound(err)` and still recover the code with `transport.AsAPIError(err)`. This matters because access-control uses 404 for both "missing" and "you referenced something missing".
  - `APIError.Details` is raw JSON because its shape differs per service; decode it with `apiError.DecodeDetails(&domainpkg.ErrorDetails{})`.
- `auth/` — service login and token management, with disk-based caching and `sync.RWMutex` for concurrency. Implements `transport.TokenSource`.
- `template/` — settings-service template client and models
  - `validation_json.go` — custom `UnmarshalJSON`/`MarshalJSON` for `Validation`, needed to handle `defaultValue` as any JSON type. **When adding a new field to the `Validation` struct, you must also add it to both `validationAlias` structs inside this file**, otherwise the field will silently unmarshal as nil.
- `abac/` — access-control client and models. Import it as `abacapi`, since `internal/resources/abac` is also named `abac`.
- `version_validator.go` — enforces minimum compatible server version at provider init

To add a new domain: create `internal/api/<domain>/` with a `Client` over `*transport.Client`, then add a field to `APIClient` in `client.go`.

### ABAC resources (`internal/resources/abac/`)
Each access-control resource has its own subpackage (`condition/`, `action/`, and `rule/` to come), each exporting `NewResource` and an `Entity`. The parent `abac` package holds only what they share:
- `errors.go` — `abac.AddError` maps service error codes to actionable diagnostics. Each resource declares its codes on its `Entity`, because the service names them inconsistently (`CONDITIONS_NOT_FOUND` vs `RULE_NOT_FOUND`). Leave a code empty if the entity has no such error.
- `tags.go` — the service re-adds `<<BuiltIn>>` on every update of a built-in object, so it is stripped from `tags` and surfaced as a read-only `built_in` attribute. Call `abac.RejectBuiltInTag` from `ValidateConfig`. Tags are otherwise sent exactly as configured.

### Custom Plan Modifiers (`internal/resources/biot_plan_modifiers/`)
- `CopyIDFromStateByNameSetModifier` — preserves computed IDs across plan cycles by matching on `name`
- `JsonNormalizePlanModifier` — normalizes JSON strings to suppress spurious diffs on `value_json` fields

### Utilities (`internal/utils/mapping_utils.go`)
Type conversion helpers between Terraform framework types and native Go types, used throughout the mappers.

## Key Patterns

- **`value_json` fields** accept arbitrary JSON; use `jsonencode()` in Terraform configs and the `JsonNormalizePlanModifier` handles diff normalization.
- **No unit tests exist** in this repo — acceptance tests (`testacc`) require a running BioT server.
- **Releases** are automated via Jenkins — pushing to `master` triggers the pipeline which builds and publishes via GoReleaser. No manual release steps needed.
- **Provider version** is set in `main.go` and must match the version in `build.sh` and `.goreleaser.yml` when cutting a release.
