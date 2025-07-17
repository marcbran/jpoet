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
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/marcbran/devsonnet/internal/terminal"
	"io"
	"os"
	"path"
)

func NewAuthMethodFromEnv() (transport.AuthMethod, error) {
	privateKey := os.Getenv("GIT_PRIVATE_KEY")
	password := os.Getenv("GIT_PASSWORD")
	if privateKey != "" {
		return ssh.NewPublicKeys("git", []byte(privateKey), password)
	}
	privateKeyFile := os.Getenv("GIT_PRIVATE_KEY_FILE")
	if privateKeyFile != "" {
		return ssh.NewPublicKeysFromFile("git", privateKeyFile, password)
	}
	username := os.Getenv("GIT_USERNAME")
	if username != "" && password != "" {
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return &http.BasicAuth{
			Username: "github",
			Password: token,
		}, nil
	}
	return nil, nil
}

func Push(ctx context.Context, source, repo, branch, p string, authMethod transport.AuthMethod) error {
	r, fs, err := cloneBranch(ctx, repo, branch, authMethod)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = checkoutBranch(ctx, r, w, branch)
	if err != nil {
		return err
	}
	err = copyFile(source, fs, p, "main.libsonnet")
	if err != nil {
		return err
	}
	err = addAllCommitPush(ctx, r, w, fmt.Sprintf("push %s main.libsonnet", p), authMethod)
	if err != nil {
		return err
	}

	err = checkoutBranch(ctx, r, w, "main")
	if err != nil {
		return err
	}
	err = copyFile(source, fs, branch, "README.md")
	if err != nil {
		return err
	}
	err = addAllCommitPush(ctx, r, w, fmt.Sprintf("push %s README.md", p), authMethod)
	if err != nil {
		return err
	}

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

func checkoutBranch(ctx context.Context, r *git.Repository, w *git.Worktree, branch string) error {
	terminal.Infof("Checking out branch %s into work tree...", branch)
	err := w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		if !errors.Is(err, plumbing.ErrReferenceNotFound) {
			return err
		}
		err := r.FetchContext(ctx, &git.FetchOptions{
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
