# Development Guide

This document provides instructions for setting up and developing the Dittoden project.

## Project Overview

Dittoden is a prototyping system for creating crowd-sourced and curated domain models using code review tools. The project uses Protocol Buffers for schema definition and generates code for multiple languages.

## Prerequisites

- Install [Homebrew](https://brew.sh)
- Install Go and Buf using Homebrew

   ```bash
   brew install go bufbuild/buf/buf
   ```

- (Optional) Install Podman and nektos/act for local Github Actions

   ```bash
   # If you are in WSL, execute this first:
   # $ wsl.exe -u root -e mount --make-rshared /
   $ brew install podman act
   ```

## Project Structure

```text
dittoden/
├── README.md                  # Project overview and roadmap
├── DEVELOPMENT.md             # This file
├── buf.yaml                   # Buf configuration for linting/breaking changes
├── buf.gen.yaml               # Code generation configuration
├── schema/                    # Protocol Buffer schema definitions
│   └── v1/
│       ├── entity.proto       # Entity definitions
│       ├── records.proto      # Record structures
│       └── relationship.proto # Relationship definitions
├── gen/                       # Generated code (auto-generated)
│   └── schema/
│       └── v1/                # Generated Go code
├── examples/                  # Example data files
│   └── SS_NOVEL.txtpb         # Example protobuf text format
├── cmd/                       # CLI commands
│   └── dittoden/              # Main CLI entry point
└── pkg/                       # Shared packages
    └── registry/              # Registry and validation logic
```

## Development Workflow

### 1. Code Generation

After making changes to the Protocol Buffer schema files in `schema/v1/`, regenerate the code:

```bash
# From project root
buf generate
```

This will:

- Generate Go code in `gen/schema/v1/`
- Update all language bindings as configured in `buf.gen.yaml`

### 2. Schema Validation

Validate your schema changes:

```bash
# Lint the schema
buf lint

# Check for breaking changes
buf breaking --against '.git#branch=main'
```

### 3. Building Tools

Build the CLI tool:

```bash
cd cmd/dittoden
go build -o dittoden .
```

Run the validation command:

```bash
./dittoden validate --dir=../../examples
```

### 4. Working with Example Data

The project includes example data in multiple formats:

- **Protocol Buffer Text Format**: `examples/SS_NOVEL.txtpb`

To validate example data against the schema:

```bash
# Validate protobuf text format
go run ./cmd/dittoden validate --dir=./examples
```

## Schema Development

### Adding New Fields

1. Edit the appropriate `.proto` file in `schema/v1/`
2. Follow Protocol Buffer style guide:
   - Use snake_case for field names
   - Add field numbers sequentially
   - Consider backward compatibility
3. Regenerate code: `buf generate`
4. Update example data if needed
5. Test with validation tools

### Schema Guidelines

- **Backward Compatibility**: Never remove fields or change field numbers
- **Forward Compatibility**: Use optional fields for new additions
- **Documentation**: Add comments to proto files explaining field usage
- **Validation**: Run `buf lint` before committing changes

## Testing

### Running Go Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Schema Validation Tests

```bash
# Validate all example data
go run ./cmd/dittoden validate --dir=./examples
```

## Contributing

1. **Fork and Clone**: Fork the repository and clone your fork
2. **Create Branch**: Create a feature branch for your changes
3. **Schema Changes**: Follow the schema development guidelines above
4. **Generate Code**: Run `buf generate` after schema changes
5. **Test**: Ensure all tools build and tests pass
6. **Lint**: Run `buf lint` to validate schema changes
7. **Commit**: Make atomic commits with clear messages
8. **Pull Request**: Submit a PR with description of changes

### Commit Message Format

```text
type(scope): brief description

Longer description if needed.

- Detailed change 1
- Detailed change 2
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
Scopes: `schema`, `cmd`, `pkg`, `gen`, `examples`, `docs`

## Troubleshooting

### Common Issues

1. **Code Generation Fails**

   ```bash
   # Check buf installation
   buf --version
   
   # Verify schema syntax
   buf lint
   ```

2. **Go Module Issues**

   ```bash
   # Clean and regenerate modules
   rm go.sum
   go mod tidy
   ```

3. **Import Path Issues**
   - Ensure generated code is present in `gen/` directory
   - Check go.mod replace directives point to correct paths
   - Regenerate code with `buf generate`

### Getting Help

- Check the [project README](README.md) for overview
- Review existing issues in the repository
- Create a new issue with reproduction steps

## IDE Setup

### VS Code

Recommended extensions:

- Protocol Buffer Language Support
- Go extension
- Buf extension (if available)

## Release Process

1. Update version numbers in relevant files
2. Run full test suite
3. Generate fresh code: `buf generate`
4. Update CHANGELOG.md
5. Create release tag
6. Build and publish artifacts

---

For questions or issues, please refer to the project repository or create an issue.
