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