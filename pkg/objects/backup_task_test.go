// pkg/objects/backup_task_test.go
package objects

import (
	"os"
	"path/filepath"
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

// ============================================================
// BackupTask Tests
// ============================================================

func TestBackupTaskNew(t *testing.T) {
	task := NewBackupTask()
	if task == nil {
		t.Fatal("expected BackupTask instance")
	}
	if task.Mode != "incremental" {
		t.Errorf("expected default Mode='incremental', got '%s'", task.Mode)
	}
}

func TestBackupTaskNewWithOptions(t *testing.T) {
	task := NewBackupTaskWithOptions(map[string]interface{}{
		"mode":       "mirror",
		"deleteExtra": true,
	})
	if task.Mode != "mirror" {
		t.Errorf("expected Mode='mirror', got '%s'", task.Mode)
	}
	if task.DeleteExtra != true {
		t.Errorf("expected DeleteExtra=true")
	}
}

func TestBackupTaskSetSourceLocal(t *testing.T) {
	task := NewBackupTask()
	task.SetSourceLocal("/tmp/source")
	if task.Source == nil {
		t.Fatal("expected Source to be set")
	}
	if task.Source.GetBasePath() != "/tmp/source" {
		t.Errorf("expected BasePath='/tmp/source', got '%s'", task.Source.GetBasePath())
	}
}

func TestBackupTaskSetTargetLocal(t *testing.T) {
	task := NewBackupTask()
	task.SetTargetLocal("/tmp/target")
	if task.Target == nil {
		t.Fatal("expected Target to be set")
	}
}

// ============================================================
// BackupTask Run Tests
// ============================================================

func TestBackupTaskRunIncremental(t *testing.T) {
	srcDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(srcDir, "file2.txt"), []byte("content2"), 0644)
	dstDir := t.TempDir()

	task := NewBackupTask()
	task.SetSourceLocal(srcDir)
	task.SetTargetLocal(dstDir)

	result := task.Run()
	if !result.Success {
		t.Errorf("expected Success=true, errors: %v", result.Errors)
	}
	if result.FilesCopied != 2 {
		t.Errorf("expected FilesCopied=2, got %d", result.FilesCopied)
	}
	if _, err := os.Stat(filepath.Join(dstDir, "file1.txt")); err != nil {
		t.Error("file1.txt not copied")
	}
}

func TestBackupTaskRunIncrementalSkip(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "same.txt"), []byte("same content"), 0644)
	os.WriteFile(filepath.Join(dstDir, "same.txt"), []byte("same content"), 0644)

	task := NewBackupTask()
	task.SetSourceLocal(srcDir)
	task.SetTargetLocal(dstDir)

	result := task.Run()
	if !result.Success {
		t.Errorf("expected Success=true")
	}
	if result.FilesSkipped < 1 {
		t.Errorf("expected at least 1 file skipped, got %d", result.FilesSkipped)
	}
}

func TestBackupTaskRunFull(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0644)

	task := NewBackupTask()
	task.SetSourceLocal(srcDir)
	task.SetTargetLocal(dstDir)
	task.SetMode("full")

	result := task.Run()
	if result.FilesCopied != 1 {
		t.Errorf("expected FilesCopied=1, got %d", result.FilesCopied)
	}
}

func TestBackupTaskRunMirror(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(dstDir, "file2.txt"), []byte("extra"), 0644)

	task := NewBackupTask()
	task.SetSourceLocal(srcDir)
	task.SetTargetLocal(dstDir)
	task.SetMode("mirror")

	result := task.Run()
	if !result.Success {
		t.Errorf("expected Success=true")
	}
	if result.FilesDeleted != 1 {
		t.Errorf("expected FilesDeleted=1, got %d", result.FilesDeleted)
	}
	if _, err := os.Stat(filepath.Join(dstDir, "file2.txt")); !os.IsNotExist(err) {
		t.Error("file2.txt should have been deleted")
	}
}