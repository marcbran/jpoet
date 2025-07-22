package repo

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func Index(ctx context.Context, repo string, authMethod transport.AuthMethod) error {
	r, fs, err := cloneBranch(ctx, repo, "main", authMethod)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	files, err := readAllFiles(fs, ".")
	if err != nil {
		return err
	}

	buildDir, err := manifestRepo(ctx, files)
	if err != nil {
		return err
	}
	defer func() {
		if cleanupErr := os.RemoveAll(buildDir); cleanupErr != nil {
			err = cleanupErr
		}
	}()

	err = copyFile(buildDir, fs, ".", "README.md")
	if err != nil {
		return err
	}
	err = addAllCommitPush(ctx, r, w, "push index README.md to main", authMethod)
	if err != nil {
		return err
	}

	return nil
}

func readAllFiles(fs billy.Filesystem, basePath string) (map[string]string, error) {
	files := make(map[string]string)

	entries, err := fs.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(basePath, entry.Name())

		if entry.IsDir() {
			subFiles, err := readAllFiles(fs, fullPath)
			if err != nil {
				return nil, err
			}
			for path, content := range subFiles {
				files[path] = content
			}
		} else {
			content, err := readFile(fs, fullPath)
			if err != nil {
				return nil, err
			}
			files[fullPath] = string(content)
		}
	}

	return files, nil
}

func readFile(fs billy.Filesystem, fullPath string) ([]byte, error) {
	file, err := fs.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ferr := file.Close(); ferr != nil {
			err = ferr
		}
	}()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return content, nil
}
