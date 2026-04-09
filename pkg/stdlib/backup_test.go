// pkg/stdlib/backup_test.go
// Tests for backup module in stdlib.
package stdlib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/topxeq/xxlang/pkg/objects"
)

func TestBackupModuleExists(t *testing.T) {
	module := Get("backup")
	if module == nil {
		t.Fatal("expected backup module to be registered")
	}
	if module.Name != "backup" {
		t.Errorf("expected module name 'backup', got '%s'", module.Name)
	}
}

func TestBackupNewTask(t *testing.T) {
	module := Get("backup")
	newTask, ok := module.Exports["newTask"]
	if !ok {
		t.Fatal("expected newTask function in backup module")
	}

	result := newTask.(*objects.Builtin).Fn()
	task, ok := result.(*objects.BackupTask)
	if !ok {
		t.Fatalf("expected BackupTask, got %T", result)
	}
	if task.Mode != "incremental" {
		t.Errorf("expected default mode 'incremental', got '%s'", task.Mode)
	}
}

func TestBackupNewTaskWithOptions(t *testing.T) {
	module := Get("backup")
	newTask, ok := module.Exports["newTask"]
	if !ok {
		t.Fatal("expected newTask function in backup module")
	}

	// Create options map
	opts := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("mode").HashKey():            {Value: objects.NewString("mirror")},
			objects.NewString("compareStrategy").HashKey(): {Value: objects.NewString("hash")},
			objects.NewString("deleteExtra").HashKey():     {Value: objects.TRUE},
		},
	}

	result := newTask.(*objects.Builtin).Fn(opts)
	task, ok := result.(*objects.BackupTask)
	if !ok {
		t.Fatalf("expected BackupTask, got %T", result)
	}
	if task.Mode != "mirror" {
		t.Errorf("expected mode 'mirror', got '%s'", task.Mode)
	}
	if task.CompareStrategy != "hash" {
		t.Errorf("expected compareStrategy 'hash', got '%s'", task.CompareStrategy)
	}
	if !task.DeleteExtra {
		t.Error("expected deleteExtra to be true")
	}
}

func TestBackupIsBackupTask(t *testing.T) {
	module := Get("backup")
	isBackupTask, ok := module.Exports["isBackupTask"]
	if !ok {
		t.Fatal("expected isBackupTask function in backup module")
	}

	// Test with BackupTask
	task := objects.NewBackupTask()
	result := isBackupTask.(*objects.Builtin).Fn(task)
	if !result.(*objects.Bool).Value {
		t.Error("expected isBackupTask to return true for BackupTask")
	}

	// Test with non-BackupTask
	result = isBackupTask.(*objects.Builtin).Fn(objects.NewString("test"))
	if result.(*objects.Bool).Value {
		t.Error("expected isBackupTask to return false for non-BackupTask")
	}
}

func TestBackupIsBackupResult(t *testing.T) {
	module := Get("backup")
	isBackupResult, ok := module.Exports["isBackupResult"]
	if !ok {
		t.Fatal("expected isBackupResult function in backup module")
	}

	// Test with BackupResult
	res := objects.NewBackupResult()
	result := isBackupResult.(*objects.Builtin).Fn(res)
	if !result.(*objects.Bool).Value {
		t.Error("expected isBackupResult to return true for BackupResult")
	}

	// Test with non-BackupResult
	result = isBackupResult.(*objects.Builtin).Fn(objects.NewString("test"))
	if result.(*objects.Bool).Value {
		t.Error("expected isBackupResult to return false for non-BackupResult")
	}
}

func TestBackupLocalToLocal(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create test file in source
	os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0644)

	module := Get("backup")
	localToLocal := module.Exports["localToLocal"].(*objects.Builtin)

	result := localToLocal.Fn(
		objects.NewString(srcDir),
		objects.NewString(dstDir),
		objects.NULL,
	)

	backupResult, ok := result.(*objects.BackupResult)
	if !ok {
		t.Fatalf("expected BackupResult, got %T", result)
	}
	if !backupResult.Success {
		t.Errorf("expected Success=true, errors: %v", backupResult.Errors)
	}
	if backupResult.FilesCopied != 1 {
		t.Errorf("expected FilesCopied=1, got %d", backupResult.FilesCopied)
	}

	// Verify file was copied
	dstContent, err := os.ReadFile(filepath.Join(dstDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(dstContent) != "hello" {
		t.Errorf("expected content 'hello', got '%s'", string(dstContent))
	}
}

func TestBackupLocalToLocalWithOptions(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create test file in source
	os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0644)

	module := Get("backup")
	localToLocal := module.Exports["localToLocal"].(*objects.Builtin)

	// Create options map
	opts := &objects.Map{
		Pairs: map[objects.HashKey]objects.MapPair{
			objects.NewString("mode").HashKey(): {Value: objects.NewString("full")},
		},
	}

	result := localToLocal.Fn(
		objects.NewString(srcDir),
		objects.NewString(dstDir),
		opts,
	)

	backupResult, ok := result.(*objects.BackupResult)
	if !ok {
		t.Fatalf("expected BackupResult, got %T", result)
	}
	if !backupResult.Success {
		t.Errorf("expected Success=true, errors: %v", backupResult.Errors)
	}
}

func TestBackupLocalToLocalIncremental(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create test file in source
	os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("hello"), 0644)

	module := Get("backup")
	localToLocal := module.Exports["localToLocal"].(*objects.Builtin)

	// First backup
	result := localToLocal.Fn(
		objects.NewString(srcDir),
		objects.NewString(dstDir),
		objects.NULL,
	)

	backupResult, ok := result.(*objects.BackupResult)
	if !ok {
		t.Fatalf("expected BackupResult, got %T", result)
	}
	if backupResult.FilesCopied != 1 {
		t.Errorf("expected FilesCopied=1 on first backup, got %d", backupResult.FilesCopied)
	}

	// Second backup (incremental - should skip)
	result = localToLocal.Fn(
		objects.NewString(srcDir),
		objects.NewString(dstDir),
		objects.NULL,
	)

	backupResult, ok = result.(*objects.BackupResult)
	if !ok {
		t.Fatalf("expected BackupResult, got %T", result)
	}
	if backupResult.FilesCopied != 0 {
		t.Errorf("expected FilesCopied=0 on second backup (incremental), got %d", backupResult.FilesCopied)
	}
	if backupResult.FilesSkipped != 1 {
		t.Errorf("expected FilesSkipped=1 on second backup, got %d", backupResult.FilesSkipped)
	}
}

func TestBackupNewTaskArgCount(t *testing.T) {
	module := Get("backup")
	newTask := module.Exports["newTask"].(*objects.Builtin)

	// Test with too many arguments
	result := newTask.Fn(objects.NULL, objects.NULL, objects.NULL)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for too many arguments")
	}
}

func TestBackupLocalToLocalArgValidation(t *testing.T) {
	module := Get("backup")
	localToLocal := module.Exports["localToLocal"].(*objects.Builtin)

	// Test with no arguments
	result := localToLocal.Fn()
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for no arguments")
	}

	// Test with wrong type for source
	result = localToLocal.Fn(objects.NewInt(123), objects.NewString("/tmp"), objects.NULL)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong source type")
	}

	// Test with wrong type for destination
	result = localToLocal.Fn(objects.NewString("/tmp"), objects.NewInt(123), objects.NULL)
	if _, ok := result.(*objects.Error); !ok {
		t.Error("expected error for wrong destination type")
	}
}