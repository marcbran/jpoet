package release

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"io"
	"os"
	"path"
)

func Run(ctx context.Context, source, repo, branch, p, token string) error {
	authMethod := &http.BasicAuth{
		Username: "github",
		Password: token,
	}
	r, fs, err := cloneBranch(ctx, repo, branch, authMethod)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(path.Join(source, "main.libsonnet"))
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	targetFile, err := fs.Create(path.Join(p, "main.libsonnet"))
	if err != nil {
		return err
	}
	defer targetFile.Close()
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	err = w.AddGlob("*")
	if err != nil {
		return err
	}
	_, err = w.Commit(fmt.Sprintf("release %s", p), &git.CommitOptions{})
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return nil
		}
		return err
	}
	err = r.Push(&git.PushOptions{
		Auth: authMethod,
	})
	if err != nil {
		return err
	}

	return nil
}

func cloneBranch(ctx context.Context, repo string, branch string, authMethod transport.AuthMethod) (*git.Repository, billy.Filesystem, error) {
	fs := memfs.New()
	r, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		Auth:          authMethod,
		URL:           repo,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return cloneAndCreateBranch(ctx, repo, branch, authMethod)
		}
		return nil, nil, err
	}
	return r, fs, nil
}

func cloneAndCreateBranch(ctx context.Context, repo string, branch string, authMethod transport.AuthMethod) (*git.Repository, billy.Filesystem, error) {
	fs := memfs.New()
	r, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		Auth:          authMethod,
		URL:           repo,
		ReferenceName: "main",
	})
	if err != nil {
		return nil, nil, err
	}
	refName := plumbing.NewBranchReferenceName(branch)
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))
	remote, err := r.Remote("origin")
	if err != nil {
		return nil, nil, err
	}
	err = remote.FetchContext(ctx, &git.FetchOptions{
		Auth:     authMethod,
		RefSpecs: []config.RefSpec{refSpec},
	})
	if err != nil && !errors.Is(err, git.NoMatchingRefSpecError{}) {
		return nil, nil, err
	}
	headRef, err := r.Head()
	if err != nil {
		return nil, nil, err
	}
	ref := plumbing.NewHashReference(refName, headRef.Hash())
	err = r.Storer.SetReference(ref)
	if err != nil {
		return nil, nil, err
	}
	return r, fs, err
}
