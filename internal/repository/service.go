package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
	"github.com/rishabyd/codeberg-cli/internal/config"
	"github.com/rishabyd/codeberg-cli/internal/validation"
)

type Service struct {
	Git Git
	API API
}

type ListItem struct {
	Name       string
	Desc       string
	Visibility string
	UpdatedAgo string
}

type ListResult struct {
	Items     []ListItem
	NameWidth int
}

type CreateInput struct {
	Name        string
	Description string
	Private     bool
	AddReadme   bool
	CloneAfter  bool
}

type CreateResult struct {
	Repo     *codeberg.Repo
	ClonedTo string
}

type MigrateInput struct {
	Source     string
	CloneAfter bool
}

type MigrateResult struct {
	Repo     *codeberg.Repo
	ClonedTo string
}

func NewService(git Git) *Service {
	if git == nil {
		git = ExecGit{}
	}
	return &Service{Git: git, API: codebergAPI{}}
}

func (s *Service) List(ctx context.Context, cfg *config.AuthConfig, limit int) (*ListResult, error) {
	if err := validation.PositiveLimit(limit, 1000); err != nil {
		return nil, usageError(err.Error())
	}

	repos, err := s.API.FetchUserRepos(ctx, cfg, limit)
	if err != nil {
		return nil, err
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].FullName < repos[j].FullName
	})

	nameWidth := 10
	items := make([]ListItem, 0, len(repos))
	for _, repo := range repos {
		name := repo.FullName
		if name == "" {
			name = repo.Name
		}
		if len(name) > nameWidth {
			nameWidth = len(name)
		}

		desc := strings.TrimSpace(repo.Description)
		if desc == "" {
			desc = "No description"
		}
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}

		vis := "public"
		if repo.Private {
			vis = "private"
		}

		updated := repo.UpdatedAt
		if strings.TrimSpace(updated) == "" {
			updated = repo.CreatedAt
		}

		items = append(items, ListItem{
			Name:       name,
			Desc:       desc,
			Visibility: vis,
			UpdatedAgo: timeAgo(updated),
		})
	}

	return &ListResult{Items: items, NameWidth: nameWidth}, nil
}

func (s *Service) Create(ctx context.Context, cfg *config.AuthConfig, input CreateInput) (*CreateResult, error) {
	if err := validation.RepoName(input.Name); err != nil {
		return nil, usageError(err.Error())
	}
	if err := validation.Description(input.Description); err != nil {
		return nil, usageError(err.Error())
	}

	repo, err := s.API.CreateRepo(ctx, cfg, codeberg.CreateRepoRequest{
		Name:        input.Name,
		Description: input.Description,
		Private:     input.Private,
		AutoInit:    input.AddReadme,
		DefaultRef:  "main",
	})
	if err != nil {
		return nil, err
	}

	out := &CreateResult{Repo: repo}
	if input.CloneAfter {
		if err := s.Git.Clone(ctx, repo.CloneURL, ""); err != nil {
			return nil, err
		}
		out.ClonedTo = repo.Name
	}
	return out, nil
}

func (s *Service) Migrate(ctx context.Context, cfg *config.AuthConfig, input MigrateInput) (*MigrateResult, error) {
	source := strings.TrimSpace(input.Source)
	if err := validation.OwnerRepo(source); err != nil {
		return nil, usageError("invalid repository format, expected owner/repo")
	}
	parts := strings.Split(source, "/")

	repo, err := s.API.MigrateRepo(ctx, cfg, codeberg.MigrateRepoRequest{
		Service:      "github",
		CloneAddr:    fmt.Sprintf("https://github.com/%s.git", source),
		RepoName:     parts[1],
		Private:      false,
		Issues:       true,
		PullRequests: true,
		Wiki:         true,
		Labels:       true,
		Milestones:   true,
		Releases:     true,
	})
	if err != nil {
		return nil, err
	}

	out := &MigrateResult{Repo: repo}
	if input.CloneAfter {
		if err := s.Git.Clone(ctx, repo.CloneURL, ""); err != nil {
			return nil, err
		}
		out.ClonedTo = repo.Name
	}

	return out, nil
}

func timeAgo(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return "-"
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
}
