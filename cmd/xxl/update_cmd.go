package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/topxeq/xxlang/pkg/version"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// newHTTPClient creates an HTTP client that respects proxy environment variables
// Supports: HTTP_PROXY, HTTPS_PROXY, http_proxy, https_proxy, NO_PROXY, no_proxy
func newHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// getProxyInfo returns information about configured proxy
func getProxyInfo() string {
	proxyURL := ""
	if p := os.Getenv("HTTPS_PROXY"); p != "" {
		proxyURL = p
	} else if p := os.Getenv("https_proxy"); p != "" {
		proxyURL = p
	} else if p := os.Getenv("HTTP_PROXY"); p != "" {
		proxyURL = p
	} else if p := os.Getenv("http_proxy"); p != "" {
		proxyURL = p
	}

	if proxyURL != "" {
		// Parse and mask password if present
		if u, err := url.Parse(proxyURL); err == nil {
			if u.User != nil {
				if _, hasPass := u.User.Password(); hasPass {
					u.User = url.UserPassword(u.User.Username(), "****")
					return u.String()
				}
			}
			return proxyURL
		}
		return proxyURL
	}
	return ""
}

// updateCmd implements the self-update command
func updateCmd(args []string) error {
	fmt.Printf("Current version: %s\n", version.Version)

	// Show proxy info if configured
	if proxyInfo := getProxyInfo(); proxyInfo != "" {
		fmt.Printf("Using proxy: %s\n", proxyInfo)
	}

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
	if latestVersion == version.Version {
		fmt.Println("You are already running the latest version!")
		return nil
	}

	// Find the appropriate asset for current platform
	assetURL, assetName, err := findAssetForPlatform(release)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s...\n", assetName)

	// Download the archive
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

	// Extract the binary from the archive
	fmt.Println("Extracting...")
	extractedFile, err := extractBinary(tempFile, assetName, targetName)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %v", err)
	}
	defer os.Remove(extractedFile)

	// Backup the current executable
	backupPath := execPath + ".backup"
	if err := copyFile(execPath, backupPath); err != nil {
		// If backup fails, try to continue anyway (might be first run)
		fmt.Printf("Warning: could not create backup: %v\n", err)
	}

	// Replace the executable
	if err := replaceExecutable(extractedFile, targetPath); err != nil {
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
// Uses the releases page directly to avoid API rate limits
func getLatestRelease() (*GitHubRelease, error) {
	client := newHTTPClient(30 * time.Second)

	// First, try the API (faster, but has rate limits)
	apiURL := "https://api.github.com/repos/topxeq/xxlang/releases/latest"
	req, err := http.NewRequest("GET", apiURL, nil)
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

	// If API returns 403 (rate limited), try alternative method
	if resp.StatusCode == http.StatusForbidden {
		return getLatestReleaseFromHTML(client)
	}

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

// getLatestReleaseFromHTML parses the releases page to get the latest version
// This is a fallback when API rate limit is hit
func getLatestReleaseFromHTML(client *http.Client) (*GitHubRelease, error) {
	// Use the redirect to get the latest tag
	// https://github.com/topxeq/xxlang/releases/latest redirects to /releases/tag/vX.Y.Z
	releasesURL := "https://github.com/topxeq/xxlang/releases/latest"

	// Don't follow redirects - we want the Location header
	clientNoRedirect := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := clientNoRedirect.Get(releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases page: %v", err)
	}
	defer resp.Body.Close()

	// We expect a redirect (301 or 302)
	if resp.StatusCode != http.StatusMovedPermanently && resp.StatusCode != http.StatusFound {
		return nil, fmt.Errorf("unexpected status %d, expected redirect", resp.StatusCode)
	}

	// Get the redirect location
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("no redirect location found")
	}

	// Parse the tag from the location URL
	// Format: https://github.com/topxeq/xxlang/releases/tag/vX.Y.Z
	var release GitHubRelease
	if idx := strings.LastIndex(location, "/tag/"); idx != -1 {
		release.TagName = location[idx+len("/tag/"):]
	} else {
		return nil, fmt.Errorf("could not parse tag from redirect URL: %s", location)
	}

	// Construct asset URLs directly
	// Format: https://github.com/topxeq/xxlang/releases/download/vX.Y.Z/xxlang-OS-ARCH.ext
	downloadBase := fmt.Sprintf("https://github.com/topxeq/xxlang/releases/download/%s/", release.TagName)

	// Add common platform assets
	platforms := []struct {
		os, arch, ext string
	}{
		{"linux", "amd64", ".tar.gz"},
		{"linux", "arm64", ".tar.gz"},
		{"darwin", "amd64", ".tar.gz"},
		{"darwin", "arm64", ".tar.gz"},
		{"windows", "amd64", ".tar.gz"},
		{"windows", "386", ".tar.gz"},
	}

	for _, p := range platforms {
		assetName := fmt.Sprintf("xxlang-%s-%s%s", p.os, p.arch, p.ext)
		release.Assets = append(release.Assets, struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		}{
			Name: assetName,
			URL:  downloadBase + assetName,
		})
	}

	return &release, nil
}

// findAssetForPlatform finds the correct asset for the current OS and architecture
// Now looks for compressed archives: .zip for Windows, .tar.gz for Linux/macOS
func findAssetForPlatform(release *GitHubRelease) (string, string, error) {
	// Build expected asset name pattern
	// Format: xxlang-{os}-{arch}.zip (Windows) or xxlang-{os}-{arch}.tar.gz (Linux/macOS)

	var patterns []string

	// Primary pattern: xxlang-{os}-{arch}.{ext}
	if runtime.GOOS == "windows" {
		patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s.zip", runtime.GOOS, runtime.GOARCH))
		patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))
	} else {
		patterns = append(patterns, fmt.Sprintf("xxlang-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH))
	}

	for _, asset := range release.Assets {
		// Check if asset matches any pattern
		for _, pattern := range patterns {
			if asset.Name == pattern {
				return asset.URL, asset.Name, nil
			}
		}

		// Also check if asset name contains OS and arch with correct extension
		assetLower := strings.ToLower(asset.Name)
		if strings.Contains(assetLower, runtime.GOOS) &&
			strings.Contains(assetLower, runtime.GOARCH) {
			if strings.HasSuffix(asset.Name, ".zip") || strings.HasSuffix(asset.Name, ".tar.gz") {
				return asset.URL, asset.Name, nil
			}
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
func downloadFile(downloadURL string) (string, error) {
	client := newHTTPClient(5 * time.Minute) // Large files may take time

	resp, err := client.Get(downloadURL)
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

// extractBinary extracts the binary from a compressed archive
func extractBinary(archivePath, archiveName, binaryName string) (string, error) {
	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "xxlang-extract-*")
	if err != nil {
		return "", err
	}

	var extractedPath string

	if strings.HasSuffix(archiveName, ".zip") {
		extractedPath, err = extractZip(archivePath, tempDir, binaryName)
	} else if strings.HasSuffix(archiveName, ".tar.gz") || strings.HasSuffix(archiveName, ".tgz") {
		extractedPath, err = extractTarGz(archivePath, tempDir, binaryName)
	} else {
		return "", fmt.Errorf("unsupported archive format: %s", archiveName)
	}

	if err != nil {
		os.RemoveAll(tempDir)
		return "", err
	}

	return extractedPath, nil
}

// extractZip extracts a zip archive and returns the path to the binary
func extractZip(zipPath, destDir, binaryName string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var extractedPath string

	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// Get the base name of the file
		baseName := filepath.Base(f.Name)

		// Look for the binary file
		if baseName == binaryName || baseName == "xxl" || baseName == "xxl.exe" {
			// Open the file in the archive
			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			// Create the destination file
			destPath := filepath.Join(destDir, binaryName)
			destFile, err := os.Create(destPath)
			if err != nil {
				rc.Close()
				return "", err
			}

			// Copy the contents
			_, err = io.Copy(destFile, rc)
			rc.Close()
			destFile.Close()

			if err != nil {
				return "", err
			}

			// Set permissions
			if err := os.Chmod(destPath, 0755); err != nil {
				return "", err
			}

			extractedPath = destPath
			break
		}
	}

	if extractedPath == "" {
		return "", fmt.Errorf("binary '%s' not found in archive", binaryName)
	}

	return extractedPath, nil
}

// extractTarGz extracts a tar.gz archive and returns the path to the binary
func extractTarGz(tarGzPath, destDir, binaryName string) (string, error) {
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

	tr := tar.NewReader(gzr)

	var extractedPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Get the base name of the file
		baseName := filepath.Base(header.Name)

		// Look for the binary file
		if baseName == binaryName || baseName == "xxl" || baseName == "xxl.exe" {
			// Create the destination file
			destPath := filepath.Join(destDir, binaryName)
			destFile, err := os.Create(destPath)
			if err != nil {
				return "", err
			}

			// Copy the contents
			_, err = io.Copy(destFile, tr)
			destFile.Close()

			if err != nil {
				return "", err
			}

			// Set permissions from tar header
			if err := os.Chmod(destPath, os.FileMode(header.Mode)); err != nil {
				return "", err
			}

			extractedPath = destPath
			break
		}
	}

	if extractedPath == "" {
		return "", fmt.Errorf("binary '%s' not found in archive", binaryName)
	}

	return extractedPath, nil
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
