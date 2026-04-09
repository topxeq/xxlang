// pkg/objects/backup_task_test.go
package objects

import (
	"testing"
)

func TestBackupResultNew(t *testing.T) {
	result := NewBackupResult()
	if result == nil {
		t.Fatal("expected BackupResult instance")
	}
	if result.FilesCopied != 0 {
		t.Errorf("expected FilesCopied=0, got %d", result.FilesCopied)
	}
	if result.Success != false {
		t.Errorf("expected Success=false, got %v", result.Success)
	}
}

func TestBackupResultInspect(t *testing.T) {
	result := NewBackupResult()
	expected := "BackupResult(filesCopied=0, filesSkipped=0)"
	if result.Inspect() != expected {
		t.Errorf("expected '%s', got '%s'", expected, result.Inspect())
	}
}

func TestBackupResultType(t *testing.T) {
	result := NewBackupResult()
	if result.Type() != BackupResultType {
		t.Errorf("expected BackupResultType, got %s", result.Type())
	}
}

// ============================================================
// BackupProgress Tests
// ============================================================

func TestBackupProgressNew(t *testing.T) {
	progress := NewBackupProgress()
	if progress == nil {
		t.Fatal("expected BackupProgress instance")
	}
	if progress.TotalFiles != 0 {
		t.Errorf("expected TotalFiles=0, got %d", progress.TotalFiles)
	}
}

func TestBackupProgressType(t *testing.T) {
	progress := NewBackupProgress()
	if progress.Type() != BackupProgressType {
		t.Errorf("expected BackupProgressType, got %s", progress.Type())
	}
}

func TestBackupProgressUpdatePercent(t *testing.T) {
	progress := NewBackupProgress()
	progress.TotalFiles = 100
	progress.ProcessedFiles = 50
	percent := progress.UpdatePercent()
	if percent != 50.0 {
		t.Errorf("expected percent=50.0, got %f", percent)
	}
}