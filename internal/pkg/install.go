package pkg

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/marcbran/jpoet/internal/terminal"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v74/github"
)

func Install(ctx context.Context, pkgDir string) error {
	cfg, err := ResolvePkgConfig(pkgDir)
	if err != nil {
		return err
	}
	pluginsDir := filepath.Join(pkgDir, ".jpoet", "plugins")
	err = os.MkdirAll(pluginsDir, 0755)
	if err != nil {
		return err
	}
	for _, plugin := range cfg.Plugins {
		err := plugin.install(ctx, pluginsDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Plugin) install(ctx context.Context, pluginsDir string) error {
	if p.Github != nil {
		return p.Github.Install(ctx, pluginsDir)
	}
	return nil
}

func (p GithubPlugin) Install(ctx context.Context, pluginsDir string) error {
	terminal.Infof("Fetching plugin from GitHub repository %s at version %s", p.Repo, p.Version)
	client := github.NewClient(nil)
	owner := strings.Split(p.Repo, "/")[0]
	repo := strings.Split(p.Repo, "/")[1]
	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, p.Version)
	if err != nil {
		return err
	}

	checksums, err := downloadChecksums(ctx, client, release, owner, repo, p.Version)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}

	tempFile, assetName, err := downloadAsset(ctx, client, release, owner, repo)
	if err != nil {
		return err
	}

	if len(checksums) == 0 {
		return fmt.Errorf("no checksums found for %s", assetName)
	}
	err = verifyChecksum(tempFile, assetName, checksums)
	if err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	pluginDir := filepath.Join(pluginsDir, repo)
	err = os.MkdirAll(pluginDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	err = extractTarGz(tempFile, pluginDir)
	if err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	return nil
}

func downloadChecksums(
	ctx context.Context,
	client *github.Client,
	release *github.RepositoryRelease,
	owner, repo, version string,
) (map[string]string, error) {
	checksumsName := fmt.Sprintf("%s_%s_checksums.txt", repo, strings.TrimPrefix(version, "v"))
	var checksumsAsset *github.ReleaseAsset
	for _, asset := range release.Assets {
		if strings.EqualFold(*asset.Name, checksumsName) {
			checksumsAsset = asset
			break
		}
	}
	if checksumsAsset == nil {
		return nil, fmt.Errorf("checksums file %s not found in release", checksumsName)
	}
	rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *checksumsAsset.ID, http.DefaultClient)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr := rc.Close(); rerr != nil {
			err = rerr
		}
	}()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	checksums := parseChecksums(string(b))
	if err != nil {
		return nil, err
	}
	return checksums, nil
}

func parseChecksums(checksumsContent string) map[string]string {
	checksums := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(checksumsContent), "\n")

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			hash := parts[0]
			filename := parts[1]
			checksums[filename] = hash
		}
	}

	return checksums
}

func downloadAsset(ctx context.Context, client *github.Client, release *github.RepositoryRelease, owner, repo string) (string, string, error) {
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH
	assetName := fmt.Sprintf("%s_%s_%s.tar.gz", repo, targetOS, targetArch)

	var targetAsset *github.ReleaseAsset
	for _, asset := range release.Assets {
		if strings.EqualFold(*asset.Name, assetName) {
			targetAsset = asset
			break
		}
	}

	if targetAsset == nil {
		return "", "", fmt.Errorf("no asset found matching %s", assetName)
	}

	rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, *targetAsset.ID, http.DefaultClient)
	if err != nil {
		return "", "", fmt.Errorf("failed to download asset: %w", err)
	}
	defer func() {
		if rerr := rc.Close(); rerr != nil {
			err = rerr
		}
	}()

	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.tar.gz", repo))
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		cerr := tempFile.Close()
		if cerr != nil {
			err = cerr
			return
		}
	}()

	_, err = io.Copy(tempFile, rc)
	if err != nil {
		return "", "", fmt.Errorf("failed to save downloaded asset: %w", err)
	}
	return tempFile.Name(), *targetAsset.Name, nil
}

func verifyChecksum(tempFile, assetName string, checksums map[string]string) error {
	expectedHash, exists := checksums[assetName]

	if !exists {
		return fmt.Errorf("no checksum found for asset %s", assetName)
	}

	file, err := os.Open(tempFile)
	if err != nil {
		return err
	}
	defer func() {
		if ferr := file.Close(); ferr != nil {
			err = ferr
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hash.Sum(nil))

	if actualHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}

func extractTarGz(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if ferr := file.Close(); ferr != nil {
			err = ferr
		}
	}()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() {
		if rerr := gzr.Close(); rerr != nil {
			err = rerr
		}
	}()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				ferr := f.Close()
				if ferr != nil {
					err = ferr
				}
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
