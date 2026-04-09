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

// ============================================================
// BackupTask - Main backup task configuration
// ============================================================

// BackupTask holds the configuration and state for a backup operation.
type BackupTask struct {
	mu sync.Mutex

	// Source and target
	Source BackupSource
	Target BackupSource

	// Configuration
	Mode            string // "incremental", "mirror", "full"
	CompareStrategy string // "sizeTime", "hash", "sizeOnly"
	HashAlgorithm   string // "md5", "sha1"
	DeleteExtra     bool   // delete files in target not in source
	ConflictPolicy  string // "overwrite", "skip", "rename"

	// Progress callback
	OnProgress func(*BackupProgress)

	// Result
	Result   *BackupResult
	Progress *BackupProgress
}

// NewBackupTask creates a new BackupTask with default values.
func NewBackupTask() *BackupTask {
	return &BackupTask{
		Mode:            "incremental",
		CompareStrategy: "sizeTime",
		HashAlgorithm:   "md5",
		DeleteExtra:     false,
		ConflictPolicy:  "overwrite",
		Result:          NewBackupResult(),
		Progress:        NewBackupProgress(),
	}
}

// NewBackupTaskWithOptions creates a BackupTask with custom options.
func NewBackupTaskWithOptions(opts map[string]interface{}) *BackupTask {
	task := NewBackupTask()

	if mode, ok := opts["mode"].(string); ok {
		task.Mode = mode
		// Mirror mode automatically sets DeleteExtra
		if mode == "mirror" {
			task.DeleteExtra = true
		}
	}
	if strategy, ok := opts["compareStrategy"].(string); ok {
		task.CompareStrategy = strategy
	}
	if algo, ok := opts["hashAlgorithm"].(string); ok {
		task.HashAlgorithm = algo
	}
	if del, ok := opts["deleteExtra"].(bool); ok {
		task.DeleteExtra = del
	}
	if policy, ok := opts["conflictPolicy"].(string); ok {
		task.ConflictPolicy = policy
	}
	if onProgress, ok := opts["onProgress"].(func(*BackupProgress)); ok {
		task.OnProgress = onProgress
	}

	return task
}

// Type returns the object type.
func (t *BackupTask) Type() ObjectType { return BackupTaskType }

// TypeTag returns the fast type tag.
func (t *BackupTask) TypeTag() TypeTag { return TagBackupTask }

// Inspect returns a string representation.
func (t *BackupTask) Inspect() string {
	return fmt.Sprintf("BackupTask(mode=%s, source=%s, target=%s)",
		t.Mode, t.sourcePath(), t.targetPath())
}

// ToBool returns true.
func (t *BackupTask) ToBool() *Bool { return TRUE }

// HashKey returns a hash key.
func (t *BackupTask) HashKey() HashKey {
	return HashKey{
		Type:  BackupTaskType,
		Value: uint64(uintptr(unsafe.Pointer(t))),
	}
}

// sourcePath returns the source path string.
func (t *BackupTask) sourcePath() string {
	if t.Source == nil {
		return "<nil>"
	}
	return t.Source.GetBasePath()
}

// targetPath returns the target path string.
func (t *BackupTask) targetPath() string {
	if t.Target == nil {
		return "<nil>"
	}
	return t.Target.GetBasePath()
}

// SetSourceLocal sets the source to a local path.
func (t *BackupTask) SetSourceLocal(path string) {
	t.Source = NewLocalSource(path)
}

// SetTargetLocal sets the target to a local path.
func (t *BackupTask) SetTargetLocal(path string) {
	t.Target = NewLocalSource(path)
}

// SetOnProgress sets the progress callback.
func (t *BackupTask) SetOnProgress(cb func(*BackupProgress)) {
	t.OnProgress = cb
}

// SetMode sets the backup mode.
func (t *BackupTask) SetMode(mode string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Mode = mode
	// Mirror mode automatically sets DeleteExtra
	if mode == "mirror" {
		t.DeleteExtra = true
	}
}

// SetCompareStrategy sets the compare strategy.
func (t *BackupTask) SetCompareStrategy(strategy string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.CompareStrategy = strategy
}

// SetHashAlgorithm sets the hash algorithm.
func (t *BackupTask) SetHashAlgorithm(algo string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.HashAlgorithm = algo
}

// SetDeleteExtra sets whether to delete extra files in target.
func (t *BackupTask) SetDeleteExtra(del bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.DeleteExtra = del
}

// SetConflictPolicy sets the conflict policy.
func (t *BackupTask) SetConflictPolicy(policy string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ConflictPolicy = policy
}