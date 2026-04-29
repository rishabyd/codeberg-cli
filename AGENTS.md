## Release Flow

Releases are handled by [GoReleaser](https://goreleaser.com) via GitHub Actions.

### To release a new version

```bash
git tag v0.2.0
git push origin v0.2.0
```

GoReleaser handles the rest:
- Builds Linux amd64 + arm64 binaries (CGO_ENABLED=0, stripped)
- Creates tar.gz archives
- Generates checksums
- Creates GitHub release with auto-generated changelog from git commits
- Version is injected at build time via `-X main.version`

### Local development

```bash
go build ./cmd/cb     # binary reports "dev" version
go test ./...          # run tests
go vet ./...           # run vet
```

### Dependencies

Run dependencies: `go-resty/resty`, `golang.org/x/oauth2`, `creativeprojects/go-selfupdate`, `rodaine/table`, `spf13/cobra`
