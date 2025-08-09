package repo

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
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/marcbran/jpoet/internal/terminal"
	"io"
	"os"
	"path"
)

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

func checkoutBranch(ctx context.Context, r *git.Repository, w *git.Worktree, branch string, authMethod transport.AuthMethod) error {
	terminal.Infof("Checking out branch %s into work tree...", branch)
	err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		if !errors.Is(err, plumbing.ErrReferenceNotFound) {
			return err
		}
		err := r.FetchContext(ctx, &git.FetchOptions{
			Auth:     authMethod,
			RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch))},
		})
		if err != nil {
			return err
		}
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.NewBranchReferenceName(branch),
		})
		if err != nil {
			return err
		}
	}
	terminal.Successf("Checked out branch %s into work tree", branch)
	return nil
}

func deleteBranch(ctx context.Context, r *git.Repository, branch string, authMethod transport.AuthMethod) error {
	terminal.Space()
	terminal.Infof("Deleting branch %s...", branch)
	err := r.PushContext(ctx, &git.PushOptions{
		Auth: authMethod,
		RefSpecs: []config.RefSpec{
			config.RefSpec(":refs/heads/" + branch),
		},
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			terminal.Successf("Branch %s already does not exist", branch)
			return nil
		}
		return err
	}
	terminal.Successf("Deleted branch %s", branch)
	return err
}

func copyFile(source string, targetFs billy.Filesystem, target string, filename string) error {
	terminal.Space()
	terminal.Infof("Copying files from %s into %s...", source, target)

	sourceFile, err := os.Open(path.Join(source, filename))
	if err != nil {
		return nil
	}
	defer func() {
		if ferr := sourceFile.Close(); ferr != nil {
			err = ferr
		}
	}()
	targetFile, err := targetFs.Create(path.Join(target, filename))
	if err != nil {
		return nil
	}
	defer func() {
		if ferr := targetFile.Close(); ferr != nil {
			err = ferr
		}
	}()
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return nil
	}

	terminal.Successf("Copied files from %s into %s", source, target)
	return err
}

func removeFile(fs billy.Filesystem, filename string) error {
	terminal.Space()
	terminal.Infof("Removing file %s...", filename)

	err := fs.Remove(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			terminal.Successf("File %s already does not exist", filename)
			return nil
		}
		return err
	}

	terminal.Successf("Removed file %s", filename)
	return nil
}

func addAllCommitPush(ctx context.Context, r *git.Repository, w *git.Worktree, message string, authMethod transport.AuthMethod) error {
	terminal.Space()
	terminal.Info("Adding files to index...")
	err := w.AddGlob("*")
	if err != nil {
		return err
	}
	terminal.Success("Added files to index")
	terminal.Infof("Making commit to %s...", message)
	_, err = w.Commit(message, &git.CommitOptions{})
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			terminal.Warn("No new changes! Won't push to remote")
			return nil
		}
		return err
	}
	terminal.Successf("Made commit to %s", message)
	terminal.Info("Pushing commit to remote...")
	err = r.PushContext(ctx, &git.PushOptions{
		Auth: authMethod,
	})
	if err != nil {
		return err
	}
	terminal.Success("Pushed commit to remote")
	return nil
}
