# protoc-gen-connect-go-handler

A protoc plugin that generates ConnectRPC handler skeletons from `.proto` definitions **without ever overwriting developer code**.

[![Go Reference](https://pkg.go.dev/badge/github.com/jackchuka/protoc-gen-connect-go-handler.svg)](https://pkg.go.dev/github.com/jackchuka/protoc-gen-connect-go-handler)
[![Go Report Card](https://goreportcard.com/badge/github.com/jackchuka/protoc-gen-connect-go-handler)](https://goreportcard.com/report/github.com/jackchuka/protoc-gen-connect-go-handler)
[![CI](https://github.com/jackchuka/protoc-gen-connect-go-handler/actions/workflows/test.yml/badge.svg)](https://github.com/jackchuka/protoc-gen-connect-go-handler/actions/workflows/test.yml)

## Features

- **Zero-clobbering guarantee** - never overwrites your implementation code
- **Two generation modes**: per-service (default) or per-method file organization
- **Smart regeneration** - only adds new method stubs for new RPCs
- **Flexible output directories** with placeholder patterns
- **Compile-time safety** via interface checks

## Installation

```bash
go install github.com/jackchuka/protoc-gen-connect-go-handler/cmd/protoc-gen-connect-go-handler@latest
```

## Usage

### Basic buf.gen.yaml

```yaml
plugins:
  - local: protoc-gen-connect-go-handler
    out: gen/go
    opt: out=gen/go
```

### With custom directory structure

```yaml
plugins:
  - local: protoc-gen-connect-go-handler
    out: service
    opt: out=service,mode=per_method,dir_pattern={package_path}/{service_snake},impl_suffix=_impl
```

This creates a structure like:

```
internal/handlers/
└── test/v1/test_service/
    ├── test_service_impl.gen.go  # manifest (always regenerated)
    ├── test_service_impl.go      # struct (created once)
    └── test_service_echo.go      # per-method files
```

## File Types Generated

| Purpose                         | File Pattern                                                   | Overwritten?   | Editable? |
| ------------------------------- | -------------------------------------------------------------- | -------------- | --------- |
| **Manifest** (interface checks) | `*{impl_suffix}.gen.go`                                        | Always         | ❌        |
| **Struct** (add fields here)    | `*{impl_suffix}.go`                                            | First run only | ✅        |
| **Method stubs**                | Same as struct (per\*service) or `\**{method}.go` (per_method) | New RPCs only  | ✅        |

## Options

| Flag          | Default       | Description                                           |
| ------------- | ------------- | ----------------------------------------------------- |
| `out`         | _Required_    | Output directory should match with protoc `out` field |
| `mode`        | `per_service` | `per_service` or `per_method`                         |
| `impl_suffix` | `_handler`    | Suffix for implementation files                       |
| `dir_pattern` | `""`          | Directory pattern with placeholders                   |

### Directory Pattern Placeholders

| Placeholder       | Expands to         | Example        |
| ----------------- | ------------------ | -------------- |
| `{package}`       | Full proto package | `test.v1`      |
| `{package_path}`  | Package with `/`   | `test/v1`      |
| `{service}`       | Service name       | `TestService`  |
| `{service_snake}` | snake_case service | `test_service` |

## Example Output

For a service like:

```protobuf
package test.v1;

service TestService {
  rpc Echo(EchoRequest) returns (EchoResponse);
}
```

### Per-service mode (default)

Generates:

- `test_service_handler.gen.go` - Interface definition (regenerated)
- `test_service_handler.go` - Struct + method implementations (safe to edit)

### Per-method mode

Generates:

- `test_service_handler.gen.go` - Interface definition (regenerated)
- `test_service_handler.go` - Struct only (safe to edit)
- `test_service_echo.go` - Echo method implementation (safe to edit)

## Development Workflow

```bash
# Day 0: Initial generation
buf generate
# Edit the generated handler files, add your logic

# Day N: Proto changes (new RPC added)
buf generate
# Only new RPC stubs are added, existing code untouched
```

## Example

See the [example/](example/) directory for a complete working example with:

- Proto service definition
- buf.gen.yaml configuration for both modes
- Generated output examples

## Requirements

- Go 1.24+
- protoc
- buf (recommended) or protoc with plugins

## Building

```bash
go build ./cmd/protoc-gen-connect-go-handler
```

## Testing

```bash
go test ./... -v
```

## Check with example

```bash
cd example
buf generate
```
