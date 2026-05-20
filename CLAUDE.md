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

This is a **Terraform Plugin Framework** provider for managing BioT templates. It exposes a single resource (`biot_template`) and connects to a BioT backend via a custom HTTP SDK.

### Provider (`internal/provider/provider.go`)
Configures credentials (`base_url`, `service_id`, `service_secret_key`), validates server version compatibility, and registers resources/data sources.

### Resource (`internal/resources/template/`)
All CRUD logic lives in `biot_template.go`. The schema is defined via nested helper functions (`attributeSchema`, `builtinAttributeSchema`, etc.). Two mapper files handle bidirectional conversion:
- `from_terraform_mapper.go` — Terraform model → API request
- `to_terraform_mapper.go` — API response → Terraform state

The model types live in `model.go` and are prefixed with `Terraform` (e.g. `TerraformTemplate`, `TerraformBuiltinAttribute`) to distinguish them from API types.

### API Client (`internal/api/`)
- `biotSdk.go` — raw HTTP calls to the BioT backend
- `api_client.go` — higher-level CRUD wrapper used by the resource
- `authenticator.go` — token management with disk-based caching and `sync.RWMutex` for concurrency
- `version_validator.go` — enforces minimum compatible server version at provider init
- `validation_json.go` — custom `UnmarshalJSON`/`MarshalJSON` for `Validation`, needed to handle `defaultValue` as any JSON type. **When adding a new field to the `Validation` struct, you must also add it to both `validationAlias` structs inside this file**, otherwise the field will silently unmarshal as nil.

### Custom Plan Modifiers (`internal/resources/biot_plan_modifiers/`)
- `CopyIDFromStateByNameSetModifier` — preserves computed IDs across plan cycles by matching on `name`
- `JsonNormalizePlanModifier` — normalizes JSON strings to suppress spurious diffs on `value_json` fields

### Utilities (`internal/utils/mapping_utils.go`)
Type conversion helpers between Terraform framework types and native Go types, used throughout the mappers.

## Key Patterns

- **`value_json` fields** accept arbitrary JSON; use `jsonencode()` in Terraform configs and the `JsonNormalizePlanModifier` handles diff normalization.
- **No unit tests exist** in this repo — acceptance tests (`testacc`) require a running BioT server.
- **Releases** are automated via Jenkins + GoReleaser; see `HOWTO-PUBLISH.md` for the manual release process.
- **Provider version** is set in `main.go` and must match the version in `build.sh` and `.goreleaser.yml` when cutting a release.
