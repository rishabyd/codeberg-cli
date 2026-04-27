package commands

import (
	"sync"

	"github.com/rishabyd/codeberg-cli/internal/repository"
)

var (
	repoServiceMu sync.RWMutex
	repoSvc       = repository.NewService(repository.ExecGit{})
)

func repoService() *repository.Service {
	repoServiceMu.RLock()
	defer repoServiceMu.RUnlock()
	return repoSvc
}
