// pkg/objects/backup_task.go
// Backup module objects for Xxlang.
package objects

import (
	"fmt"
	"path/filepath"
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

	// Exclude patterns
	excludePatterns []string

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

// SetSourceRemote sets the source to a remote path via SSH.
func (t *BackupTask) SetSourceRemote(client *SSHClient, path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Source = NewRemoteSource(client, path)
}

// SetTargetRemote sets the target to a remote path via SSH.
func (t *BackupTask) SetTargetRemote(client *SSHClient, path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Target = NewRemoteSource(client, path)
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

// SetExcludePatterns sets patterns to exclude.
func (t *BackupTask) SetExcludePatterns(patterns []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Store exclude patterns - this will be used by shouldExclude
	t.excludePatterns = patterns
}

// GetExcludePatterns returns the exclude patterns.
func (t *BackupTask) GetExcludePatterns() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.excludePatterns
}

// ============================================================
// Run - Main backup execution
// ============================================================

// Run executes the backup task and returns the result.
func (t *BackupTask) Run() *BackupResult {
	// Create a fresh result
	t.Result = NewBackupResult()
	t.Progress = NewBackupProgress()

	// Validate source and target are set
	if t.Source == nil {
		t.Result.AddError("source is not set")
		return t.Result
	}
	if t.Target == nil {
		t.Result.AddError("target is not set")
		return t.Result
	}

	// Ensure target base directory exists
	t.Target.MkdirAll("")

	// List files from source
	sourceFiles, err := t.Source.ListFiles()
	if err != nil {
		t.Result.AddError("failed to list source files: " + err.Error())
		return t.Result
	}

	// Filter out directories and excluded files
	var filesToProcess []BackupFileInfo
	for _, file := range sourceFiles {
		if file.IsDir {
			continue
		}
		if t.shouldExclude(file.Path) {
			continue
		}
		filesToProcess = append(filesToProcess, file)
	}

	// Initialize progress
	t.Progress.TotalFiles = len(filesToProcess)
	t.notifyProgress(t.Progress)

	// Build target file map for quick lookup
	targetFiles, err := t.Target.ListFiles()
	if err != nil {
		// Target might not exist yet - that's OK for initial backup
		targetFiles = []BackupFileInfo{}
	}
	targetMap := make(map[string]BackupFileInfo)
	for _, file := range targetFiles {
		if !file.IsDir {
			targetMap[file.Path] = file
		}
	}

	// Process each source file
	for _, srcFile := range filesToProcess {
		t.Progress.SetCurrentFile(srcFile.Path, "check")
		t.notifyProgress(t.Progress)

		needCopy, _ := t.needCopyFile(srcFile, targetMap)
		if needCopy {
			t.Progress.SetCurrentFile(srcFile.Path, "copy")
			t.notifyProgress(t.Progress)

			// Read from source
			content, err := t.Source.ReadFile(srcFile.Path)
			if err != nil {
				t.Result.AddError("failed to read source file " + srcFile.Path + ": " + err.Error())
				t.Progress.IncrementProcessed()
				t.Progress.UpdatePercent()
				t.notifyProgress(t.Progress)
				continue
			}

			// Write to target
			err = t.Target.WriteFile(srcFile.Path, content)
			if err != nil {
				t.Result.AddError("failed to write target file " + srcFile.Path + ": " + err.Error())
				t.Progress.IncrementProcessed()
				t.Progress.UpdatePercent()
				t.notifyProgress(t.Progress)
				continue
			}

			t.Result.FilesCopied++
			t.Result.BytesTransferred += int64(len(content))
		} else {
			t.Progress.SetCurrentFile(srcFile.Path, "skip")
			t.notifyProgress(t.Progress)
			t.Result.FilesSkipped++
		}

		t.Result.FilesChecked++
		t.Progress.IncrementProcessed()
		t.Progress.UpdatePercent()
		t.notifyProgress(t.Progress)
	}

	// Delete extra files in target if mirror mode or DeleteExtra is set
	if t.DeleteExtra || t.Mode == "mirror" {
		t.deleteExtraFiles(targetMap, sourceFiles)
	}

	// Mark as success if no errors
	t.Result.Success = !t.Result.HasErrors()
	return t.Result
}

// needCopyFile determines if a file needs to be copied.
// Returns true and reason if copy is needed, false and empty string otherwise.
func (t *BackupTask) needCopyFile(srcFile BackupFileInfo, targetMap map[string]BackupFileInfo) (bool, string) {
	// Full mode always copies
	if t.Mode == "full" {
		return true, "full"
	}

	// Check if file exists in target
	targetFile, exists := targetMap[srcFile.Path]
	if !exists {
		return true, "new"
	}

	// Compare based on strategy
	switch t.CompareStrategy {
	case "sizeTime":
		// Compare size first
		if srcFile.Size != targetFile.Size {
			return true, "size"
		}
		// Same size, check modification time
		if srcFile.MTime.After(targetFile.MTime) {
			return true, "newer"
		}
		// Same size and not newer - skip
		return false, ""

	case "hash":
		// Calculate hashes and compare
		srcHash, err := t.Source.CalculateHash(srcFile.Path, t.HashAlgorithm)
		if err != nil {
			// If we can't compute hash, fall back to size/time
			if srcFile.Size != targetFile.Size {
				return true, "size"
			}
			if srcFile.MTime.After(targetFile.MTime) {
				return true, "newer"
			}
			return false, ""
		}
		targetHash, err := t.Target.CalculateHash(srcFile.Path, t.HashAlgorithm)
		if err != nil {
			// Target hash failed - assume different
			return true, "hash"
		}
		if srcHash != targetHash {
			return true, "hash"
		}
		return false, ""

	case "sizeOnly":
		// Only compare size
		if srcFile.Size != targetFile.Size {
			return true, "size"
		}
		return false, ""

	default:
		// Default to size/time comparison
		if srcFile.Size != targetFile.Size {
			return true, "size"
		}
		if srcFile.MTime.After(targetFile.MTime) {
			return true, "newer"
		}
		return false, ""
	}
}

// shouldExclude checks if a path should be excluded based on patterns.
func (t *BackupTask) shouldExclude(path string) bool {
	patterns := t.GetExcludePatterns()
	if len(patterns) == 0 {
		return false
	}

	for _, pattern := range patterns {
		// Try filepath.Match first
		matched, err := filepathMatch(pattern, path)
		if err == nil && matched {
			return true
		}
		// Also check if pattern is a substring
		if containsSubstring(path, pattern) {
			return true
		}
	}
	return false
}

// deleteExtraFiles deletes files in target that don't exist in source.
func (t *BackupTask) deleteExtraFiles(targetMap map[string]BackupFileInfo, sourceFiles []BackupFileInfo) {
	// Build source map for quick lookup
	sourceMap := make(map[string]bool)
	for _, file := range sourceFiles {
		if !file.IsDir {
			sourceMap[file.Path] = true
		}
	}

	// Find and delete extra files
	for path := range targetMap {
		if !sourceMap[path] {
			t.Progress.SetCurrentFile(path, "delete")
			t.notifyProgress(t.Progress)

			err := t.Target.DeleteFile(path)
			if err != nil {
				t.Result.AddError("failed to delete extra file " + path + ": " + err.Error())
				continue
			}
			t.Result.FilesDeleted++
		}
	}
}

// notifyProgress calls the progress callback if set.
func (t *BackupTask) notifyProgress(progress *BackupProgress) {
	if t.OnProgress != nil {
		t.OnProgress(progress)
	}
}

// CheckConflicts detects potential conflicts without modifying files.
func (t *BackupTask) CheckConflicts() []string {
	conflicts := []string{}

	if t.Source == nil || t.Target == nil {
		return conflicts
	}

	// List files from both source and target
	sourceFiles, err := t.Source.ListFiles()
	if err != nil {
		return conflicts
	}

	targetFiles, err := t.Target.ListFiles()
	if err != nil {
		return conflicts
	}

	// Build target map
	targetMap := make(map[string]BackupFileInfo)
	for _, file := range targetFiles {
		if !file.IsDir {
			targetMap[file.Path] = file
		}
	}

	// Check for conflicts
	for _, srcFile := range sourceFiles {
		if srcFile.IsDir {
			continue
		}
		if t.shouldExclude(srcFile.Path) {
			continue
		}

		_, exists := targetMap[srcFile.Path]
		if exists {
			// Check if files differ
			needCopy, _ := t.needCopyFile(srcFile, targetMap)
			if needCopy && t.ConflictPolicy == "skip" {
				conflicts = append(conflicts, srcFile.Path)
			}
		}
	}

	return conflicts
}

// filepathMatch is a helper for pattern matching (windows-safe).
func filepathMatch(pattern, path string) (bool, error) {
	// Use filepath.Match which handles platform-specific separators
	return filepath.Match(pattern, path)
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsAt(s, substr)))
}

// containsAt checks if s contains substr at any position.
func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}