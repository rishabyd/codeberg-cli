package commands

import (
	"fmt"

	"github.com/rishabyd/codeberg-cli/internal/codeberg"
)

func printRepoResult(action string, repo *codeberg.Repo) {
	if repo == nil {
		return
	}
	fmt.Printf("%s %s\n", action, repo.FullName)
	fmt.Printf("HTTPS: %s\n", repo.CloneURL)
	fmt.Printf("SSH:   %s\n", repo.SSHURL)
	fmt.Printf("Web:   %s\n", repo.HTMLURL)
}

func printClonePath(name string) {
	if name == "" {
		return
	}
	fmt.Printf("Cloned to ./%s\n", name)
}
