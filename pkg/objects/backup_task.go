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