# proto2md

Tool for converting Protobuf file definitions into Markdown documentation

## Installation

### Quick Install (Recommended)

One-liner that auto-detects your OS and architecture:

#### Using curl

Latest version:

```bash
curl -sL https://raw.githubusercontent.com/oxio/proto2md/main/install.sh | bash
```

Specific version:

```bash
curl -sL https://raw.githubusercontent.com/oxio/proto2md/main/install.sh | bash -s -- v1.0.0
```

#### Using wget

Latest version:

```bash
wget -qO- https://raw.githubusercontent.com/oxio/proto2md/main/install.sh | bash
```

Specific version:

```bash
wget -qO- https://raw.githubusercontent.com/oxio/proto2md/main/install.sh | bash -s -- v1.0.0
```

### Go Install

If you have Go installed:

Latest version:

```bash
go install github.com/oxio/proto2md@latest
```

Specific version:

```bash
go install github.com/oxio/proto2md@v1.0.0
```

### Manual Download

Download the appropriate binary for your platform from the [latest release](https://github.com/oxio/proto2md/releases/latest) (or a specific tag on [GitHub Releases](https://github.com/oxio/proto2md/releases)).

#### Linux

```bash
# amd64 (x86_64)
curl -sL https://github.com/oxio/proto2md/releases/latest/download/proto2md-linux-amd64 -o proto2md && chmod +x proto2md && sudo mv proto2md /usr/local/bin/

# arm64 (aarch64)
curl -sL https://github.com/oxio/proto2md/releases/latest/download/proto2md-linux-arm64 -o proto2md && chmod +x proto2md && sudo mv proto2md /usr/local/bin/

# arm (32-bit, armv7)
curl -sL https://github.com/oxio/proto2md/releases/latest/download/proto2md-linux-arm -o proto2md && chmod +x proto2md && sudo mv proto2md /usr/local/bin/
```

#### macOS

```bash
# Apple Silicon (M1/M2/M3/M4/M5)
curl -sL https://github.com/oxio/proto2md/releases/latest/download/proto2md-darwin-arm64 -o proto2md && chmod +x proto2md && sudo mv proto2md /usr/local/bin/

# Intel (x86_64)
curl -sL https://github.com/oxio/proto2md/releases/latest/download/proto2md-darwin-amd64 -o proto2md && chmod +x proto2md && sudo mv proto2md /usr/local/bin/
```

#### Windows

```powershell
# amd64 (x86_64)
Invoke-WebRequest -Uri "https://github.com/oxio/proto2md/releases/latest/download/proto2md-windows-amd64.exe" -OutFile "proto2md.exe"

# arm64
Invoke-WebRequest -Uri "https://github.com/oxio/proto2md/releases/latest/download/proto2md-windows-arm64.exe" -OutFile "proto2md.exe"
```

Or using PowerShell with `curl` alias:

```powershell
curl -sL https://github.com/oxio/proto2md/releases/latest/download/proto2md-windows-amd64.exe -o proto2md.exe
```

### Build from Source

```bash
task build
# or
go build -o proto2md .
```

## Usage

Show help:

```sh
proto2md --help
```

Print a greeting:

```sh
proto2md hello
proto2md hello --name Alice
```

Print the version:

```sh
proto2md version
```

### Subcommands

- `proto2md hello [--name <name>]` — Print a Hello World greeting.
- `proto2md version` — Print the tool version.
- `proto2md help [command]` — Show help for any command.

## Development

Common tasks via [Task](https://taskfile.dev/):

- `task check` — Run formatting check, `go vet`, `golangci-lint`, `govulncheck`, build, and tests.
- `task test` — Run unit tests.
- `task build` — Build the binary into `bin/`.
- `task build:all` — Cross-compile binaries for macOS, Linux, and Windows.
- `task fmt` — Format Go source files in place.
- `task tidy` — Tidy `go.mod` and `go.sum`.
