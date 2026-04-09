// pkg/objects/backup_task.go
// Backup module objects for Xxlang.
package objects

import (
	"fmt"
	"sync"
	"unsafe"
)

// ============================================================
// BackupResult - Result of backup operation
// ============================================================

// BackupResult holds the result of a backup operation.
type BackupResult struct {
	mu               sync.Mutex
	Success          bool
	FilesCopied      int
	FilesSkipped     int
	FilesDeleted     int
	FilesChecked     int
	Conflicts        []string
	Errors           []string
	BytesTransferred int64
	Duration         float64 // seconds
}

// NewBackupResult creates a new BackupResult.
func NewBackupResult() *BackupResult {
	return &BackupResult{
		Success:   false,
		Conflicts: []string{},
		Errors:    []string{},
	}
}

// Type returns the object type.
func (r *BackupResult) Type() ObjectType { return BackupResultType }

// TypeTag returns the fast type tag.
func (r *BackupResult) TypeTag() TypeTag { return TagBackupResult }

// Inspect returns a string representation.
func (r *BackupResult) Inspect() string {
	return fmt.Sprintf("BackupResult(filesCopied=%d, filesSkipped=%d)",
		r.FilesCopied, r.FilesSkipped)
}

// ToBool returns the success status.
func (r *BackupResult) ToBool() *Bool {
	return &Bool{Value: r.Success}
}

// HashKey returns a hash key.
func (r *BackupResult) HashKey() HashKey {
	return HashKey{
		Type:  BackupResultType,
		Value: uint64(uintptr(unsafe.Pointer(r))),
	}
}

// AddError adds an error message.
func (r *BackupResult) AddError(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Errors = append(r.Errors, msg)
}

// AddConflict adds a conflict file path.
func (r *BackupResult) AddConflict(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Conflicts = append(r.Conflicts, path)
}

// HasConflicts returns true if there were conflicts.
func (r *BackupResult) HasConflicts() bool {
	return len(r.Conflicts) > 0
}

// HasErrors returns true if there were errors.
func (r *BackupResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// Summary returns a summary string.
func (r *BackupResult) Summary() string {
	return fmt.Sprintf("Backup completed: %d copied, %d skipped, %d deleted, %d errors",
		r.FilesCopied, r.FilesSkipped, r.FilesDeleted, len(r.Errors))
}

// ============================================================
// BackupProgress - Progress state during backup
// ============================================================

// BackupProgress holds progress state during backup execution.
type BackupProgress struct {
	mu               sync.Mutex
	TotalFiles       int
	ProcessedFiles   int
	CurrentFile      string
	CurrentAction    string // "copy", "skip", "delete", "check"
	BytesTransferred int64
	TotalBytes       int64
	Percent          float64
}

// NewBackupProgress creates a new BackupProgress.
func NewBackupProgress() *BackupProgress {
	return &BackupProgress{}
}

// Type returns the object type.
func (p *BackupProgress) Type() ObjectType { return BackupProgressType }

// TypeTag returns the fast type tag.
func (p *BackupProgress) TypeTag() TypeTag { return TagBackupProgress }

// Inspect returns a string representation.
func (p *BackupProgress) Inspect() string {
	return fmt.Sprintf("BackupProgress(percent=%.1f, file=%s)", p.Percent, p.CurrentFile)
}

// ToBool returns true.
func (p *BackupProgress) ToBool() *Bool { return TRUE }

// HashKey returns a hash key.
func (p *BackupProgress) HashKey() HashKey {
	return HashKey{
		Type:  BackupProgressType,
		Value: uint64(uintptr(unsafe.Pointer(p))),
	}
}

// UpdatePercent calculates and updates the percentage.
func (p *BackupProgress) UpdatePercent() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.TotalFiles == 0 {
		p.Percent = 0
		return 0
	}
	p.Percent = float64(p.ProcessedFiles) * 100.0 / float64(p.TotalFiles)
	return p.Percent
}

// SetCurrentFile sets the current file being processed.
func (p *BackupProgress) SetCurrentFile(file string, action string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CurrentFile = file
	p.CurrentAction = action
}

// IncrementProcessed increments the processed files count.
func (p *BackupProgress) IncrementProcessed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ProcessedFiles++
}