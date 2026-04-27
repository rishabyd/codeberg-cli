package repository

import (
	"context"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
)

type API interface {
	FetchUserRepos(ctx context.Context, cfg *config.AuthConfig, limit int) ([]codeberg.Repo, error)
	CreateRepo(ctx context.Context, cfg *config.AuthConfig, payload codeberg.CreateRepoRequest) (*codeberg.Repo, error)
	GetCurrentUser(ctx context.Context, cfg *config.AuthConfig) (*codeberg.User, error)
	MigrateRepo(ctx context.Context, cfg *config.AuthConfig, payload codeberg.MigrateRepoRequest) (*codeberg.Repo, error)
}

type codebergAPI struct{}

func (codebergAPI) FetchUserRepos(ctx context.Context, cfg *config.AuthConfig, limit int) ([]codeberg.Repo, error) {
	return codeberg.FetchUserRepos(ctx, cfg, limit)
}

func (codebergAPI) CreateRepo(ctx context.Context, cfg *config.AuthConfig, payload codeberg.CreateRepoRequest) (*codeberg.Repo, error) {
	return codeberg.CreateRepo(ctx, cfg, payload)
}

func (codebergAPI) GetCurrentUser(ctx context.Context, cfg *config.AuthConfig) (*codeberg.User, error) {
	return codeberg.GetCurrentUser(ctx, cfg)
}

func (codebergAPI) MigrateRepo(ctx context.Context, cfg *config.AuthConfig, payload codeberg.MigrateRepoRequest) (*codeberg.Repo, error) {
	return codeberg.MigrateRepo(ctx, cfg, payload)
}
