// pkg/objects/backup_source.go
// BackupSource abstraction for local and remote file operations.
package objects

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// BackupFileInfo represents file metadata for backup comparison.
type BackupFileInfo struct {
	Path  string
	Size  int64
	MTime time.Time
	IsDir bool
	Hash  string // optional, for hash verification
}

// BackupSource interface for file operations.
type BackupSource interface {
	// ListFiles returns all files under the base path (recursive).
	ListFiles() ([]BackupFileInfo, error)
	// GetFileInfo returns info for a single file (relative path).
	GetFileInfo(relPath string) (BackupFileInfo, error)
	// ReadFile reads file content (relative path).
	ReadFile(relPath string) ([]byte, error)
	// WriteFile writes content to file (relative path).
	WriteFile(relPath string, content []byte) error
	// DeleteFile deletes a file (relative path).
	DeleteFile(relPath string) error
	// MkdirAll creates directory with parents (relative path).
	MkdirAll(relPath string) error
	// Exists checks if path exists (relative path).
	Exists(relPath string) bool
	// GetBasePath returns the base path.
	GetBasePath() string
	// CalculateHash computes hash of file content.
	CalculateHash(relPath string, algo string) (string, error)
}

// ============================================================
// LocalSource - Local filesystem backup source
// ============================================================

// LocalSource implements BackupSource for local filesystem.
type LocalSource struct {
	BasePath string
}

// NewLocalSource creates a new LocalSource.
func NewLocalSource(basePath string) *LocalSource {
	return &LocalSource{BasePath: basePath}
}

// GetBasePath returns the base path.
func (s *LocalSource) GetBasePath() string {
	return s.BasePath
}

// ListFiles returns all files under the base path.
func (s *LocalSource) ListFiles() ([]BackupFileInfo, error) {
	var files []BackupFileInfo
	err := filepath.Walk(s.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(s.BasePath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		files = append(files, BackupFileInfo{
			Path:  relPath,
			Size:  info.Size(),
			MTime: info.ModTime(),
			IsDir: info.IsDir(),
		})
		return nil
	})
	return files, err
}

// GetFileInfo returns info for a single file.
func (s *LocalSource) GetFileInfo(relPath string) (BackupFileInfo, error) {
	fullPath := filepath.Join(s.BasePath, relPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return BackupFileInfo{}, err
	}
	return BackupFileInfo{
		Path:  relPath,
		Size:  info.Size(),
		MTime: info.ModTime(),
		IsDir: info.IsDir(),
	}, nil
}

// ReadFile reads file content.
func (s *LocalSource) ReadFile(relPath string) ([]byte, error) {
	fullPath := filepath.Join(s.BasePath, relPath)
	return os.ReadFile(fullPath)
}

// WriteFile writes content to file.
func (s *LocalSource) WriteFile(relPath string, content []byte) error {
	fullPath := filepath.Join(s.BasePath, relPath)
	// Create parent directories if needed
	dir := filepath.Dir(fullPath)
	if dir != "." && dir != s.BasePath {
		os.MkdirAll(dir, 0755)
	}
	return os.WriteFile(fullPath, content, 0644)
}

// DeleteFile deletes a file.
func (s *LocalSource) DeleteFile(relPath string) error {
	fullPath := filepath.Join(s.BasePath, relPath)
	return os.Remove(fullPath)
}

// MkdirAll creates directory with parents.
func (s *LocalSource) MkdirAll(relPath string) error {
	fullPath := filepath.Join(s.BasePath, relPath)
	return os.MkdirAll(fullPath, 0755)
}

// Exists checks if path exists.
func (s *LocalSource) Exists(relPath string) bool {
	fullPath := filepath.Join(s.BasePath, relPath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// CalculateHash computes hash of file content.
func (s *LocalSource) CalculateHash(relPath string, algo string) (string, error) {
	content, err := s.ReadFile(relPath)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(algo) {
	case "md5":
		hash := md5.Sum(content)
		return fmt.Sprintf("%x", hash), nil
	case "sha1":
		hash := sha1.Sum(content)
		return fmt.Sprintf("%x", hash), nil
	default:
		return "", errors.New("unsupported hash algorithm: " + algo)
	}
}

// Helper function for copy
func copyFile(src io.Reader, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err
}

// ============================================================
// SSHClientInterface - Interface for SSH client operations
// ============================================================

// SSHClientInterface defines the methods needed by RemoteSource.
// The existing SSHClient in ssh_client.go already implements this interface.
type SSHClientInterface interface {
	// Exec executes a command and returns stdout.
	Exec(cmd string) (string, error)
	// ReadFile reads a remote file content.
	ReadFile(path string) (string, error)
	// WriteFile writes content to a remote file.
	WriteFile(path, content string) error
	// ListDir lists directory contents.
	ListDir(path string) ([]map[string]interface{}, error)
	// Stat returns file information.
	Stat(path string) (map[string]interface{}, error)
	// MkdirAll creates directory with parents.
	MkdirAll(path string) error
	// Remove removes a remote file.
	Remove(path string) error
	// Exists checks if a path exists.
	Exists(path string) bool
}

// ============================================================
// RemoteSource - SSH-based remote backup source
// ============================================================

// RemoteSource implements BackupSource for remote SSH filesystem.
type RemoteSource struct {
	Client   SSHClientInterface
	BasePath string
}

// NewRemoteSource creates a new RemoteSource with an SSH client.
func NewRemoteSource(client SSHClientInterface, basePath string) *RemoteSource {
	return &RemoteSource{
		Client:   client,
		BasePath: basePath,
	}
}

// GetBasePath returns the base path.
func (s *RemoteSource) GetBasePath() string {
	return s.BasePath
}

// ListFiles returns all files under the base path using find command.
func (s *RemoteSource) ListFiles() ([]BackupFileInfo, error) {
	// Use find command to get all files recursively
	cmd := fmt.Sprintf("find %s -type f -o -type d 2>/dev/null | head -n 10000", escapeRemotePath(s.BasePath))
	output, err := s.Client.Exec(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []BackupFileInfo

	for _, line := range lines {
		if line == "" || line == s.BasePath {
			continue
		}

		// Get relative path
		relPath := strings.TrimPrefix(line, s.BasePath)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			continue
		}

		// Get file info using stat
		info, err := s.GetFileInfo(relPath)
		if err != nil {
			continue // Skip files we can't stat
		}
		files = append(files, info)
	}

	return files, nil
}

// GetFileInfo returns info for a single file using stat command.
func (s *RemoteSource) GetFileInfo(relPath string) (BackupFileInfo, error) {
	fullPath := s.BasePath + "/" + relPath
	// Use stat command with custom format: size|mtime|filetype
	cmd := "stat -c '" + "%s" + "|" + "%Y" + "|" + "%F" + "' " + escapeRemotePath(fullPath) + " 2>/dev/null"
	output, err := s.Client.Exec(cmd)
	if err != nil {
		return BackupFileInfo{}, fmt.Errorf("failed to stat remote file: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(output), "|")
	if len(parts) < 3 {
		return BackupFileInfo{}, errors.New("unexpected stat output format")
	}

	size, _ := strconv.ParseInt(parts[0], 10, 64)
	mtimeUnix, _ := strconv.ParseInt(parts[1], 10, 64)
	fileType := strings.TrimSpace(parts[2])

	return BackupFileInfo{
		Path:  relPath,
		Size:  size,
		MTime: time.Unix(mtimeUnix, 0),
		IsDir: fileType == "directory",
	}, nil
}

// ReadFile reads file content from remote.
func (s *RemoteSource) ReadFile(relPath string) ([]byte, error) {
	fullPath := s.BasePath + "/" + relPath
	content, err := s.Client.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote file: %w", err)
	}
	return []byte(content), nil
}

// WriteFile writes content to remote file.
func (s *RemoteSource) WriteFile(relPath string, content []byte) error {
	fullPath := s.BasePath + "/" + relPath
	return s.Client.WriteFile(fullPath, string(content))
}

// DeleteFile deletes a remote file.
func (s *RemoteSource) DeleteFile(relPath string) error {
	fullPath := s.BasePath + "/" + relPath
	return s.Client.Remove(fullPath)
}

// MkdirAll creates directory with parents on remote.
func (s *RemoteSource) MkdirAll(relPath string) error {
	fullPath := s.BasePath + "/" + relPath
	return s.Client.MkdirAll(fullPath)
}

// Exists checks if path exists on remote.
func (s *RemoteSource) Exists(relPath string) bool {
	fullPath := s.BasePath + "/" + relPath
	return s.Client.Exists(fullPath)
}

// CalculateHash computes hash of remote file content using md5sum/sha1sum.
func (s *RemoteSource) CalculateHash(relPath string, algo string) (string, error) {
	fullPath := s.BasePath + "/" + relPath

	var cmd string
	switch strings.ToLower(algo) {
	case "md5":
		cmd = fmt.Sprintf("md5sum %s 2>/dev/null", escapeRemotePath(fullPath))
	case "sha1":
		cmd = fmt.Sprintf("sha1sum %s 2>/dev/null", escapeRemotePath(fullPath))
	default:
		return "", errors.New("unsupported hash algorithm: " + algo)
	}

	output, err := s.Client.Exec(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Output format: "hashvalue  filename"
	parts := strings.Fields(strings.TrimSpace(output))
	if len(parts) < 1 {
		return "", errors.New("unexpected hash output format")
	}

	return parts[0], nil
}

// escapeRemotePath wraps path in single quotes for shell safety.
func escapeRemotePath(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\\''") + "'"
}