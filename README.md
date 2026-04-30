# cb — Codeberg CLI

A minimal, `gh`-style CLI for Codeberg.

```bash
cb auth login                    # Login via OAuth
cb repo create my-project --public --clone
cb repo list --limit 10
cb health
cb update
```

## Table of Contents

- [Install](#install)
- [QuickStart](#quickstart)
- [Commands](#commands)
- [Flags](#flags)
- [Examples](#examples)
- [Troubleshooting](#troubleshooting)
- [Build from Source](#build-from-source)
- [Contributing](#contributing)
- [License](#license)

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/rishabyd/codeberg-cli/main/install.sh | sh
```

**Update:** `cb update`

**Uninstall:** `cb uninstall`

Supported platforms: Linux amd64, Linux arm64.

## QuickStart

```bash
# Login (opens browser)
cb auth login

# Create a public repo and clone it
cb repo create my-project --public --clone
cd my-project

# List your repos
cb repo list

# Check Codeberg status
cb health
```

## Commands

### `cb auth`

```bash
cb auth login     # Login via OAuth (opens browser)
cb auth status    # Show auth status
cb auth logout    # Clear session and credentials
```

### `cb repo`

Aliases: `cb repository`

```bash
cb repo list [flags]                  # List your repositories (alias: ls)
cb repo create <name> [flags]         # Create a repository
cb repo migrate <owner/repo> [flags]  # Migrate from GitHub
```

### `cb update`

Check for and install the latest release.

```bash
cb update
```

### `cb health`

Check Codeberg.org and CI/CD service status.

```bash
cb health
```

### `cb uninstall`

Remove the `cb` binary, local config, and git credential helper.

```bash
cb uninstall
```

## Flags

### Global

| Flag | Description |
|------|-------------|
| `-v`, `--version` | Show version |
| `--verbose` | Enable debug output |

### `cb repo list`

| Flag | Default | Description |
|------|---------|-------------|
| `--limit` | `30` | Number of repositories |

### `cb repo create`

| Flag | Short | Description |
|------|-------|-------------|
| `--description` | `-d` | Repository description |
| `--public` | — | Public repository (default if neither specified) |
| `--private` | — | Private repository |
| `--add-readme` | — | Initialize with README |
| `--clone` | `-c` | Clone after creating |

### `cb repo migrate`

| Flag | Description |
|------|-------------|
| `--clone` | Clone after migrating |

## Examples

### Create and push a new project

```bash
cb repo create my-project --public --clone
cd my-project
echo "# My Project" > README.md
git add README.md
git commit -m "init"
git push -u origin main
```

### Migrate from GitHub

```bash
cb repo migrate owner/repo --clone
cd repo
git remote -v
```

### List with limit

```bash
cb repo list --limit 5
```

## Troubleshooting

**DNS failure**
```
lookup codeberg.org ... server misbehaving
```
Local network resolver issue. Check your DNS settings.

**Auth expired**
```
Session expired. Run `cb auth login`
```
Run `cb auth login` to re-authenticate.

**Debug any command**
```bash
cb auth status --verbose
cb repo list --verbose
```

**Verify git credential helper**
```bash
git config --global --get credential.https://codeberg.org.helper
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Runtime / auth / network / API error |
| `2` | Usage or argument error |

## Build from Source

Requires Go 1.26+.

```bash
git clone https://github.com/rishabyd/codeberg-cli.git
cd codeberg-cli

# Format, lint, test
go tool gofumpt -w . && go vet ./... && go test ./...

# Build
go build -o cb ./cmd/cb
./cb --help
```

## Contributing

1. Fork and clone
2. Format and test: `go tool gofumpt -w . && go vet ./... && go test ./...`
3. Build: `go build -o cb ./cmd/cb`
4. Commit and push
5. Open a pull request

## License

MIT — see [LICENSE](LICENSE).