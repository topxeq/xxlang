package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// updateCmd implements the self-update command
func updateCmd(args []string) error {
	fmt.Printf("Current version: %s\n", Version)
	fmt.Println("Checking for updates...")

	// Get latest release from GitHub
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}

	// Parse version from tag (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	fmt.Printf("Latest version: %s\n", latestVersion)

	// Compare versions
	if latestVersion == Version {
		fmt.Println("You are already running the latest version!")
		return nil
	}

	// Find the appropriate asset for current platform
	assetURL, assetName, err := findAssetForPlatform(release)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Download the new binary
	tempFile, err := downloadFile(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %v", err)
	}
	defer os.Remove(tempFile)

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Determine the target executable name
	var targetName string
	if runtime.GOOS == "windows" {
		targetName = "xxl.exe"
	} else {
		targetName = "xxl"
	}

	// Get the directory of the current executable
	execDir := filepath.Dir(execPath)
	targetPath := filepath.Join(execDir, targetName)

	// Backup the current executable
	backupPath := execPath + ".backup"
	if err := copyFile(execPath, backupPath); err != nil {
		// If backup fails, try to continue anyway (might be first run)
		fmt.Printf("Warning: could not create backup: %v\n", err)
	}

	// Replace the executable
	if err := replaceExecutable(tempFile, targetPath); err != nil {
		// Try to restore from backup
		if restoreErr := os.Rename(backupPath, execPath); restoreErr != nil {
			fmt.Printf("Error: failed to restore backup: %v\n", restoreErr)
		}
		return fmt.Errorf("failed to replace executable: %v", err)
	}

	// Clean up backup on success
	os.Remove(backupPath)

	fmt.Printf("\nSuccessfully updated to version %s!\n", latestVersion)
	fmt.Printf("New executable: %s\n", targetPath)

	return nil
}

// getLatestRelease fetches the latest release from GitHub
func getLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := "https://api.github.com/repos/topxeq/xxlang/releases/latest"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent (required by GitHub API)
	req.Header.Set("User-Agent", "xxlang-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// findAssetForPlatform finds the correct asset for the current OS and architecture
func findAssetForPlatform(release *GitHubRelease) (string, string, error) {
	// Build expected asset name pattern
	// Expected format: xxlang-{version}-{os}-{arch}(.exe for windows)
	// or: xxl-{os}-{arch}(.exe for windows)

	var patterns []string
	version := strings.TrimPrefix(release.TagName, "v")

	// Try multiple naming patterns
	patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH))
	patterns = append(patterns, fmt.Sprintf("xxl-%s-%s", runtime.GOOS, runtime.GOARCH))
	patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s", runtime.GOOS, runtime.GOARCH))

	if runtime.GOOS == "windows" {
		// Also try with .exe extension
		patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s-%s.exe", version, runtime.GOOS, runtime.GOARCH))
		patterns = append(patterns, fmt.Sprintf("xxl-%s-%s.exe", runtime.GOOS, runtime.GOARCH))
		patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s.exe", runtime.GOOS, runtime.GOARCH))
	}

	for _, asset := range release.Assets {
		// Check if asset matches any pattern
		for _, pattern := range patterns {
			if asset.Name == pattern || strings.Contains(asset.Name, pattern) {
				return asset.URL, asset.Name, nil
			}
		}

		// Also check if asset name contains OS and arch
		if strings.Contains(strings.ToLower(asset.Name), runtime.GOOS) &&
			strings.Contains(strings.ToLower(asset.Name), runtime.GOARCH) {
			return asset.URL, asset.Name, nil
		}
	}

	// List available assets for user information
	var available []string
	for _, asset := range release.Assets {
		available = append(available, asset.Name)
	}

	return "", "", fmt.Errorf("no release found for %s/%s. Available assets: %s",
		runtime.GOOS, runtime.GOARCH, strings.Join(available, ", "))
}

// downloadFile downloads a file to a temporary location with progress display
func downloadFile(url string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Minute, // Large files may take time
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "xxlang-update-*")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Get file size for progress
	totalSize := resp.ContentLength
	var downloaded int64

	// Create progress bar
	buf := make([]byte, 32*1024) // 32KB buffer
	lastPercent := -1

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := tempFile.Write(buf[:n])
			if writeErr != nil {
				os.Remove(tempFile.Name())
				return "", writeErr
			}
			downloaded += int64(n)

			// Update progress
			if totalSize > 0 {
				percent := int(float64(downloaded) / float64(totalSize) * 100)
				if percent != lastPercent {
					printProgressBar(percent, downloaded, totalSize)
					lastPercent = percent
				}
			} else {
				// Unknown size, show downloaded bytes
				printProgressBar(-1, downloaded, 0)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempFile.Name())
			return "", err
		}
	}

	// Print newline after progress bar
	fmt.Println()

	return tempFile.Name(), nil
}

// printProgressBar prints a progress bar
func printProgressBar(percent int, downloaded, total int64) {
	width := 40
	var bar strings.Builder

	if percent >= 0 {
		// Known size - show percentage
		filled := percent * width / 100
		bar.WriteString("\r  [")
		for i := 0; i < width; i++ {
			if i < filled {
				bar.WriteString("=")
			} else {
				bar.WriteString(" ")
			}
		}
		bar.WriteString("] ")

		// Show percentage and size
		fmt.Printf("%s %3d%%  (%s / %s)",
			bar.String(),
			percent,
			formatBytes(downloaded),
			formatBytes(total))
	} else {
		// Unknown size - show downloaded bytes only
		bar.WriteString("\r  Downloading: ")
		fmt.Printf("%s %s", bar.String(), formatBytes(downloaded))
	}
}

// formatBytes formats bytes to human readable string
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
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
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// replaceExecutable replaces the current executable with the new one
func replaceExecutable(tempPath, targetPath string) error {
	// Make the new file executable
	if err := os.Chmod(tempPath, 0755); err != nil {
		return err
	}

	// On Windows, we cannot replace a running executable directly
	// We need to rename the old one and then move the new one
	if runtime.GOOS == "windows" {
		oldPath := targetPath + ".old"
		// Remove old backup if exists
		os.Remove(oldPath)
		// Rename current executable
		if err := os.Rename(targetPath, oldPath); err != nil {
			return fmt.Errorf("could not rename old executable: %v", err)
		}
		// Move new executable
		if err := os.Rename(tempPath, targetPath); err != nil {
			// Try to restore
			os.Rename(oldPath, targetPath)
			return fmt.Errorf("could not move new executable: %v", err)
		}
		// Remove old executable (might fail if still in use, that's ok)
		os.Remove(oldPath)
	} else {
		// On Unix systems, we can overwrite directly
		if err := os.Rename(tempPath, targetPath); err != nil {
			return err
		}
	}

	return nil
}
