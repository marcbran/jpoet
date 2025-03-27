package release

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/marcbran/jsonnet-kit/internal/terminal"
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
	terminal.Infof("Checking out branch %s into work tree...", branch)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return err
	}
	terminal.Successf("Checked out branch %s into work tree", branch)

	terminal.Space()
	terminal.Infof("Copying files from %s into %s...", source, p)
	sourceFile, err := os.Open(path.Join(source, "main.libsonnet"))
	if err != nil {
		return err
	}
	defer func() {
		if ferr := sourceFile.Close(); ferr != nil {
			err = ferr
		}
	}()
	targetFile, err := fs.Create(path.Join(p, "main.libsonnet"))
	if err != nil {
		return err
	}
	defer func() {
		if ferr := targetFile.Close(); ferr != nil {
			err = ferr
		}
	}()
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}
	terminal.Successf("Copied files from %s into %s", source, p)

	terminal.Space()
	terminal.Info("Adding files to index...")
	err = w.AddGlob("*")
	if err != nil {
		return err
	}
	terminal.Success("Added files to index")
	terminal.Infof("Making commit to release %s...", p)
	_, err = w.Commit(fmt.Sprintf("release %s", p), &git.CommitOptions{})
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			terminal.Warn("No new changes! Won't push to remote")
			return nil
		}
		return err
	}
	terminal.Successf("Made commit to release %s", p)
	terminal.Info("Pushing commit to remote...")
	err = r.Push(&git.PushOptions{
		Auth: authMethod,
	})
	if err != nil {
		return err
	}
	terminal.Success("Pushed commit to remote")

	return nil
}

func cloneBranch(ctx context.Context, repo string, branch string, authMethod transport.AuthMethod) (*git.Repository, billy.Filesystem, error) {
	fs := memfs.New()
	terminal.Infof("Cloning branch %s from %s...", branch, repo)
	r, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		Auth:          authMethod,
		URL:           repo,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			terminal.Warn("Branch not found! Will need to clone default branch and create new branch")
			return cloneAndCreateBranch(ctx, repo, branch, authMethod)
		}
		return nil, nil, err
	}
	terminal.Successf("Cloned branch %s from %s", branch, repo)
	return r, fs, nil
}

func cloneAndCreateBranch(ctx context.Context, repo string, branch string, authMethod transport.AuthMethod) (*git.Repository, billy.Filesystem, error) {
	fs := memfs.New()
	defaultBranch := "main"
	terminal.Infof("Cloning branch %s from %s...", defaultBranch, repo)
	r, err := git.CloneContext(ctx, memory.NewStorage(), fs, &git.CloneOptions{
		Auth:          authMethod,
		URL:           repo,
		ReferenceName: plumbing.NewBranchReferenceName(defaultBranch),
	})
	if err != nil {
		return nil, nil, err
	}
	terminal.Successf("Cloned branch %s from %s", defaultBranch, repo)

	terminal.Infof("Creating branch %s...", branch)
	headRef, err := r.Head()
	if err != nil {
		return nil, nil, err
	}
	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branch), headRef.Hash())
	err = r.Storer.SetReference(ref)
	if err != nil {
		return nil, nil, err
	}
	terminal.Successf("Created branch %s...", branch)
	return r, fs, err
}
