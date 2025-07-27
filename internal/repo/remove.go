package repo

import (
	"context"
	"fmt"
	"path"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/marcbran/jpoet/internal/pkg"
)

func Remove(ctx context.Context, pkgDir string, authMethod transport.AuthMethod) error {
	cfg, err := pkg.ResolvePkgConfig(pkgDir)
	if err != nil {
		return err
	}
	repo := cfg.Coordinates.Repo
	branch := cfg.Coordinates.Branch

	r, fs, err := cloneBranch(ctx, repo, "main", authMethod)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = deleteBranch(ctx, r, branch, authMethod)
	if err != nil {
		return err
	}
	err = removeFile(fs, path.Join(branch, "README.md"))
	if err != nil {
		return err
	}
	err = addAllCommitPush(ctx, r, w, fmt.Sprintf("remove %s/README.md from main", branch), authMethod)
	if err != nil {
		return err
	}

	return nil
}
