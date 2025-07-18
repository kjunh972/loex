package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

type Updater struct {
	owner      string
	repo       string
	httpClient *http.Client
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func New() *Updater {
	return &Updater{
		owner: "kjunh972",
		repo:  "loex",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (u *Updater) CheckForUpdate(currentVersion string) (bool, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.owner, u.repo)
	
	resp, err := u.httpClient.Get(url)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	
	// Skip version comparison if current version is "dev"
	if currentVersion == "dev" {
		return true, latestVersion, nil
	}

	currentVer, err := version.NewVersion(currentVersion)
	if err != nil {
		return false, "", fmt.Errorf("invalid current version: %v", err)
	}

	latestVer, err := version.NewVersion(latestVersion)
	if err != nil {
		return false, "", fmt.Errorf("invalid latest version: %v", err)
	}

	return latestVer.GreaterThan(currentVer), latestVersion, nil
}

func (u *Updater) Update(targetVersion string) error {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Determine the asset name based on OS and architecture
	assetName := u.getAssetName(targetVersion)
	
	// Download the new version
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s", 
		u.owner, u.repo, targetVersion, assetName)
	
	tempFile, err := u.downloadFile(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer os.Remove(tempFile)

	// Extract the binary
	binaryPath, err := u.extractBinary(tempFile)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %v", err)
	}
	defer os.Remove(binaryPath)

	// Replace the current executable
	if err := u.replaceExecutable(execPath, binaryPath); err != nil {
		return fmt.Errorf("failed to replace executable: %v", err)
	}

	return nil
}

func (u *Updater) getAssetName(version string) string {
	osName := runtime.GOOS
	archName := runtime.GOARCH
	
	return fmt.Sprintf("loex-%s-%s-%s.tar.gz", version, osName, archName)
}

func (u *Updater) downloadFile(url string) (string, error) {
	resp, err := u.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "loex-update-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Copy the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func (u *Updater) extractBinary(tarGzPath string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the loex binary
		if header.Name == "loex" || filepath.Base(header.Name) == "loex" {
			// Create temporary file for the binary
			tempBinary, err := os.CreateTemp("", "loex-binary-*")
			if err != nil {
				return "", err
			}
			defer tempBinary.Close()

			// Copy the binary content
			_, err = io.Copy(tempBinary, tarReader)
			if err != nil {
				os.Remove(tempBinary.Name())
				return "", err
			}

			// Make it executable
			if err := os.Chmod(tempBinary.Name(), 0755); err != nil {
				os.Remove(tempBinary.Name())
				return "", err
			}

			return tempBinary.Name(), nil
		}
	}

	return "", fmt.Errorf("loex binary not found in archive")
}

func (u *Updater) replaceExecutable(currentPath, newPath string) error {
	// Create a backup of the current executable
	backupPath := currentPath + ".backup"
	if err := u.copyFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Replace the current executable with the new one
	if err := u.copyFile(newPath, currentPath); err != nil {
		// Restore backup if replacement fails
		u.copyFile(backupPath, currentPath)
		os.Remove(backupPath)
		return fmt.Errorf("failed to replace executable: %v", err)
	}

	// Remove backup
	os.Remove(backupPath)

	return nil
}

func (u *Updater) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}