# cb (Codeberg CLI)

`cb` is a minimal, `gh`-style CLI for Codeberg focused on practical daily workflows.

Current scope:
- auth login/logout/status
- repo list/create/migrate
- self update

## Install / Update / Uninstall

Install:

```bash
curl -fsSL https://raw.githubusercontent.com/rishabyd/codeberg-cli/main/install.sh | sh
```

Update:

```bash
cb update
```

Uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/rishabyd/codeberg-cli/main/uninstall.sh | sh
```

From local checkout:

```bash
./install.sh
./uninstall.sh
```

## Command Syntax

```bash
cb auth <action>
cb repo <action> [target]
cb update
```

Auth:

```bash
cb auth login
cb auth status
cb auth logout
```

Repo:

```bash
cb repo list [--limit <n>]
cb repo create <name> [flags]
cb repo migrate <owner/repo> [--clone]
```

Create flags:
- `-d, --description <text>`
- `--public`
- `--private`
- `--add-readme`
- `-c, --clone`

## Common Workflows

Create repository:

```bash
cb repo create my-project --public
cb repo create my-project --private --description "internal tools"
cb repo create my-project --public --add-readme --clone
```

Migrate repository:

```bash
cb repo migrate owner/repo
cb repo migrate owner/repo --clone
```

`cb repo migrate` performs public GitHub -> public Codeberg migration and includes:
- source code
- issues
- pull requests
- wiki
- labels
- milestones
- releases

## Troubleshooting

- DNS failure (`lookup codeberg.org ... server misbehaving`) is a local network resolver issue, not token expiry.
- If refresh token is revoked or expired, run `cb auth login` again.
- Verify auth and API reachability with `cb auth status`.
- Ensure git HTTPS helper is configured: `git config --global --get credential.https://codeberg.org.helper`.

## Exit Codes

- `0`: success
- `1`: runtime/auth/network/api error
- `2`: usage or argument error

## Build from Source

```bash
go build -o cb ./cmd/cb
./cb --help
```

## Platform Compatibility

Official prebuilt release binaries are provided for:
- Linux x86_64 (`linux_amd64`)
- Linux ARM64 (`linux_arm64`)

Other platforms can use source build (`go build -o cb ./cmd/cb`).

## Versioning

This project uses Go-native tag-based versioning.

- Local/dev builds show version `dev`.
- Release builds inject version from git tag using Go linker flags.

Create and publish a release:

```bash
git status
git tag -a v0.4.13 -m "v0.4.13"
git push origin main --tags
```

To verify locally:

```bash
go build -ldflags "-X main.version=0.4.13" -o cb ./cmd/cb
./cb --version
```

Required secret for releases:
- `GITHUB_TOKEN` (provided automatically by GitHub Actions)
