# go-cli-tool-template

<!-- BEGIN TEMPLATE DESCRIPTION -->
Template repository for building CLI tools in Go with [Cobra](https://github.com/spf13/cobra), [Task](https://taskfile.dev/), `golangci-lint`, `govulncheck`, GitHub Actions CI/CD, and a cross-platform install script.
<!-- END TEMPLATE DESCRIPTION -->

<!-- BEGIN TEMPLATE USAGE -->
## Using This Template

1. Create a new repository from this GitHub template and clone it locally.
2. Run the bootstrap task interactively to configure the repository, binary name, description, and `LICENSE`, and remove `tools/bootstrap`:
   ```bash
   task bootstrap
   ```
   Or run non-interactively with flags:
   ```bash
   go run ./tools/bootstrap -repo your-org/your-repo -binary your-cli -description "Short description of your CLI tool" -author "Your Name"
   ```
   (Pass `-keep-bootstrap` if you want to keep `tools/bootstrap`, the `bootstrap` task, and this section.)
3. Replace the starter `cmd/hello.go` and `internal/hello/` package with your own commands and domain logic.
4. Run `task check` to verify everything builds, lints, and passes tests.

<!-- END TEMPLATE USAGE -->
## Installation

### Quick Install (Recommended)

One-liner that auto-detects your OS and architecture:

#### Using curl

Latest version:

```bash
curl -sL https://raw.githubusercontent.com/oxio/go-cli-tool-template/main/install.sh | bash
```

Specific version:

```bash
curl -sL https://raw.githubusercontent.com/oxio/go-cli-tool-template/main/install.sh | bash -s -- v1.0.0
```

#### Using wget

Latest version:

```bash
wget -qO- https://raw.githubusercontent.com/oxio/go-cli-tool-template/main/install.sh | bash
```

Specific version:

```bash
wget -qO- https://raw.githubusercontent.com/oxio/go-cli-tool-template/main/install.sh | bash -s -- v1.0.0
```

### Go Install

If you have Go installed:

Latest version:

```bash
go install github.com/oxio/go-cli-tool-template@latest
```

Specific version:

```bash
go install github.com/oxio/go-cli-tool-template@v1.0.0
```

### Manual Download

Download the appropriate binary for your platform from the [latest release](https://github.com/oxio/go-cli-tool-template/releases/latest) (or a specific tag on [GitHub Releases](https://github.com/oxio/go-cli-tool-template/releases)).

#### Linux

```bash
# amd64 (x86_64)
curl -sL https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-linux-amd64 -o go-cli-tool-template && chmod +x go-cli-tool-template && sudo mv go-cli-tool-template /usr/local/bin/

# arm64 (aarch64)
curl -sL https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-linux-arm64 -o go-cli-tool-template && chmod +x go-cli-tool-template && sudo mv go-cli-tool-template /usr/local/bin/

# arm (32-bit, armv7)
curl -sL https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-linux-arm -o go-cli-tool-template && chmod +x go-cli-tool-template && sudo mv go-cli-tool-template /usr/local/bin/
```

#### macOS

```bash
# Apple Silicon (M1/M2/M3/M4/M5)
curl -sL https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-darwin-arm64 -o go-cli-tool-template && chmod +x go-cli-tool-template && sudo mv go-cli-tool-template /usr/local/bin/

# Intel (x86_64)
curl -sL https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-darwin-amd64 -o go-cli-tool-template && chmod +x go-cli-tool-template && sudo mv go-cli-tool-template /usr/local/bin/
```

#### Windows

```powershell
# amd64 (x86_64)
Invoke-WebRequest -Uri "https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-windows-amd64.exe" -OutFile "go-cli-tool-template.exe"

# arm64
Invoke-WebRequest -Uri "https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-windows-arm64.exe" -OutFile "go-cli-tool-template.exe"
```

Or using PowerShell with `curl` alias:

```powershell
curl -sL https://github.com/oxio/go-cli-tool-template/releases/latest/download/go-cli-tool-template-windows-amd64.exe -o go-cli-tool-template.exe
```

### Build from Source

```bash
task build
# or
go build -o go-cli-tool-template .
```

## Usage

Show help:

```sh
go-cli-tool-template --help
```

Print a greeting:

```sh
go-cli-tool-template hello
go-cli-tool-template hello --name Alice
```

Print the version:

```sh
go-cli-tool-template version
```

### Subcommands

- `go-cli-tool-template hello [--name <name>]` — Print a Hello World greeting.
- `go-cli-tool-template version` — Print the tool version.
- `go-cli-tool-template help [command]` — Show help for any command.

## Development

Common tasks via [Task](https://taskfile.dev/):

- `task check` — Run formatting check, `go vet`, `golangci-lint`, `govulncheck`, build, and tests.
- `task test` — Run unit tests.
- `task build` — Build the binary into `bin/`.
- `task build:all` — Cross-compile binaries for macOS, Linux, and Windows.
- `task fmt` — Format Go source files in place.
- `task tidy` — Tidy `go.mod` and `go.sum`.
