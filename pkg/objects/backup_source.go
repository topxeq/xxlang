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
		// Normalize to forward slashes for cross-platform compatibility
		relPath = strings.ReplaceAll(relPath, "\\", "/")
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
	// Create parent directories if needed (including base path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
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
	// ReadBytes reads a remote file and returns raw bytes.
	ReadBytes(path string) ([]byte, error)
	// WriteFile writes content to a remote file.
	WriteFile(path, content string) error
	// ListDir lists directory contents.
	ListDir(path string) ([]map[string]interface{}, error)
	// WalkDir recursively lists all files in a directory.
	WalkDir(path string) ([]map[string]interface{}, error)
	// Stat returns file information.
	Stat(path string) (map[string]interface{}, error)
	// MkdirAll creates directory with parents.
	MkdirAll(path string) error
	// Remove removes a remote file.
	Remove(path string) error
	// Exists checks if a path exists.
	Exists(path string) bool
	// IsDir checks if path is a directory.
	IsDir(path string) bool
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

// ListFiles returns all files under the base path via SFTP WalkDir.
func (s *RemoteSource) ListFiles() ([]BackupFileInfo, error) {
	entries, err := s.Client.WalkDir(s.BasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote files: %w", err)
	}

	var files []BackupFileInfo
	for _, entry := range entries {
		path, _ := entry["path"].(string)
		relPath := strings.TrimPrefix(path, s.BasePath)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			continue
		}

		isDir, _ := entry["isDir"].(bool)
		size, _ := entry["size"].(int64)
		modTime, _ := entry["modTime"].(int64)

		files = append(files, BackupFileInfo{
			Path:  relPath,
			MTime: time.Unix(modTime, 0),
			Size:  size,
			IsDir: isDir,
		})
	}

	return files, nil
}

// GetFileInfo returns info for a single file using SFTP Stat.
func (s *RemoteSource) GetFileInfo(relPath string) (BackupFileInfo, error) {
	fullPath := s.BasePath + "/" + relPath
	info, err := s.Client.Stat(fullPath)
	if err != nil {
		return BackupFileInfo{}, fmt.Errorf("failed to stat remote file: %w", err)
	}

	size, _ := info["size"].(int64)
	isDir, _ := info["isDir"].(bool)
	modTime, _ := info["modTime"].(int64)

	return BackupFileInfo{
		Path:  relPath,
		MTime: time.Unix(modTime, 0),
		Size:  size,
		IsDir: isDir,
	}, nil
}

// ReadFile reads file content from remote via SFTP binary transfer.
func (s *RemoteSource) ReadFile(relPath string) ([]byte, error) {
	fullPath := s.BasePath + "/" + relPath
	return s.Client.ReadBytes(fullPath)
}

// WriteFile writes content to remote file.
func (s *RemoteSource) WriteFile(relPath string, content []byte) error {
	fullPath := s.BasePath + "/" + relPath
	// Create parent directory if needed
	dir := fullPath[:strings.LastIndex(fullPath, "/")]
	if dir != "" && dir != s.BasePath {
		s.Client.MkdirAll(dir)
	}
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

// CalculateHash computes hash of remote file content by reading via SFTP and hashing locally.
func (s *RemoteSource) CalculateHash(relPath string, algo string) (string, error) {
	data, err := s.ReadFile(relPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	switch strings.ToLower(algo) {
	case "md5":
		hash := md5.Sum(data)
		return fmt.Sprintf("%x", hash), nil
	case "sha1":
		hash := sha1.Sum(data)
		return fmt.Sprintf("%x", hash), nil
	default:
		return "", errors.New("unsupported hash algorithm: " + algo)
	}
}
