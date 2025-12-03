# Copilot Instructions for Dittoden

## Project Overview

Dittoden is a prototyping system for crowd-sourced domain modeling using Protocol Buffers as the schema language and text protobuf (`.txtpb`) files for data input. The architecture separates schema definition, validation, and data storage for collaborative curation through code review workflows.

## Architecture & Components

### Core Schema (`schema/v1/*.proto`)

- **Entity-centric design**: Everything revolves around `Entity` with typed properties (`EntityType` enum: PERSON, ORGANIZATION, LOCATION, EVENT, ARTIFACT)
- **Relationship modeling**: `Relationship` connects entities via `RelationshipType` definitions with bidirectional support
- **Flexible labeling**: Entities use `Label` messages with types (PRIMARY, ALTERNATIVE, TITLE) rather than simple strings
- **Tag-based metadata**: Both entities and relationships support arbitrary `Tag` key-value pairs for extensibility

### Data Format Conventions

- **Human-readable codes**: Entity codes like `"SS_SUNNY_LOST_FROM_LIGHT"` instead of UUIDs for discoverability
- **Relationship participant naming**: Always use `a` and `z` fields (not source/target) with role specifications
- **Timestamp handling**: Use `google.protobuf.Timestamp` for dates, stored as seconds since epoch
- **File organization**: Domain-specific folders in `examples/` (e.g., `examples/country/ph.txtpb`)

### Registry System (`pkg/registry/`)

The `Registry` acts as the central validation engine:

- **Aggregation**: Merges all `.txtpb` files from a directory into a single `Records` message
- **Reference validation**: Ensures relationships reference existing entities and relationship types
- **Duplicate detection**: Prevents code collisions across all record types
- **Structured logging**: Uses `slog` with tint for colored terminal output

## Development Workflows

### Essential Commands

```bash
# Always run after schema changes - regenerates Go bindings
make generate-proto-libs  # or just: buf generate

# Validates and formats all example data
make validate-examples

# Complete validation pipeline (use for CI/pre-commit)
make validate
```

### Schema Modification Process

1. Edit `.proto` files in `schema/v1/`
2. Run `buf generate` to update `gen/schema/v1/*.pb.go`
3. Update example data if schema changes affect existing records
4. Run `make validate` to ensure all data still validates

### Data File Patterns

- **Records structure**: Each `.txtpb` file contains a `Records` message with arrays of entities, relationships, and relationship types
- **Cross-references**: Relationships use `type_ref` to reference `RelationshipType.code` and participants reference `Entity.code`
- **Source tagging**: Add `tags { name: "source" value: "DOMAIN_NAME" }` to track data origins

## Code Generation & Dependencies

- **Buf-based toolchain**: Uses `buf.gen.yaml` for protobuf compilation, not raw `protoc`
- **Go module structure**: Generated code lives in `gen/schema/v1/` with `option go_package = "schema/v1"`
- **Import paths**: Generated types are imported as `schema "github.com/nmcapule/dittoden/gen/schema/v1"`
- **Text format parsing**: Use `google.golang.org/protobuf/encoding/prototext` for `.txtpb` files

## Testing & Validation

- **CLI validation**: `go run ./cmd/validate --dir=./examples` processes all example data
- **Format enforcement**: `txtpbfmt` tool ensures consistent text protobuf formatting
- **CI pipeline**: GitHub Actions runs full validation on main/staging pushes
- **Breaking change detection**: `buf breaking` checks against git history for schema compatibility

## Project-Specific Patterns

- **Entity codes as primary keys**: Design entities with memorable, namespace-prefixed codes
- **Bidirectional relationship handling**: Use entity with lexicographically smaller code as participant `a`
- **Reserved field usage**: Schema uses `reserved 6 to 10;` in Entity for future extension points
- **Empty package placeholder**: `cmd/dittoden/` and `pkg/{cli,export}/` exist but are empty (roadmap items)

## Common Pitfalls

- Don't edit generated files in `gen/` - always modify source `.proto` files
- Entity codes must be globally unique across all domains/files
- Relationship participants must exist as entities before relationships can be validated
- Use `txtpbfmt` formatting to avoid validation failures in CI
