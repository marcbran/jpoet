package repo

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/marcbran/devsonnet/internal/pkg"
	"os"
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

func Push(ctx context.Context, pkgDir, buildDir string, authMethod transport.AuthMethod) error {
	cfg, err := pkg.ResolvePkgConfig(pkgDir)
	if err != nil {
		return err
	}
	repo := cfg.Coordinates.Repo
	branch := cfg.Coordinates.Branch
	p := cfg.Coordinates.Path

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
	err = copyFile(buildDir, fs, p, "main.libsonnet")
	if err != nil {
		return err
	}
	err = addAllCommitPush(ctx, r, w, fmt.Sprintf("push %s/main.libsonnet to %s", p, branch), authMethod)
	if err != nil {
		return err
	}

	err = checkoutBranch(ctx, r, w, "main")
	if err != nil {
		return err
	}
	err = copyFile(buildDir, fs, branch, "README.md")
	if err != nil {
		return err
	}
	err = addAllCommitPush(ctx, r, w, fmt.Sprintf("push %s/README.md to main", branch), authMethod)
	if err != nil {
		return err
	}

	return nil
}
