// pkg/objects/file.go
// File object type for streaming file operations in Xxlang.
package objects

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// FileMode represents the mode in which a file is opened
type FileMode string

const (
	ModeRead   FileMode = "r"   // Read-only
	ModeWrite  FileMode = "w"   // Write-only (truncate)
	ModeAppend FileMode = "a"   // Append-only
	ModeRW     FileMode = "rw"  // Read and write
	ModeRWPlus FileMode = "rw+" // Read and write (create if not exists)
)

// File represents an open file handle for streaming I/O operations.
// It wraps os.File and provides methods for reading, writing, seeking,
// and managing file state.
type File struct {
	Handle    *os.File     // The underlying OS file handle
	Path      string       // The file path
	Mode      FileMode     // The mode the file was opened with
	Open      bool         // Whether the file is currently open
	Position  int64        // Current position in the file
	LockType  FileLockType // Current lock type (0 = unlocked)
	mu        sync.Mutex   // Mutex for thread-safe operations
}

// FileLockType represents the type of file lock
type FileLockType int

const (
	LockNone     FileLockType = iota // No lock
	LockShared                       // Shared (read) lock
	LockExclusive                    // Exclusive (write) lock
)

// NewFile creates a new File object from an os.File handle.
func NewFile(handle *os.File, path string, mode FileMode) *File {
	return &File{
		Handle:   handle,
		Path:     path,
		Mode:     mode,
		Open:     true,
		Position: 0,
		LockType: LockNone,
	}
}

// Type returns the object type.
func (f *File) Type() ObjectType { return FileType }

// TypeTag returns the type tag for fast type checking.
func (f *File) TypeTag() TypeTag { return TagFile }

// Inspect returns a string representation of the File object.
func (f *File) Inspect() string {
	if f.Open {
		return fmt.Sprintf("<FILE open path=%s mode=%s>", f.Path, f.Mode)
	}
	return fmt.Sprintf("<FILE closed path=%s>", f.Path)
}

// ToBool returns true if the file is open, false otherwise.
func (f *File) ToBool() *Bool { return &Bool{Value: f.Open} }

// HashKey returns a hash key for the File object.
func (f *File) HashKey() HashKey {
	return HashKey{Type: FileType, Value: uint64(uintptr(unsafePointer(f)))}
}

// unsafePointer returns the pointer as uintptr for hashing
func unsafePointer(f *File) uintptr {
	// Use the Path string pointer as a unique identifier
	return uintptr(len(f.Path)) << 32 | uintptr(f.Position & 0xFFFFFFFF)
}

// Close closes the file handle.
// Returns nil on success, error otherwise.
func (f *File) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return nil
	}

	// Release any locks
	if f.LockType != LockNone {
		f.unlockFile()
	}

	err := f.Handle.Close()
	f.Open = false
	return err
}

// Read reads up to n bytes from the file.
// Returns the bytes read as an array of integers.
func (f *File) Read(n int) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return nil, fmt.Errorf("file is closed")
	}

	buf := make([]byte, n)
	bytesRead, err := f.Handle.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	f.Position += int64(bytesRead)
	return buf[:bytesRead], nil
}

// ReadLine reads a single line from the file.
// Returns the line as a string without the trailing newline.
func (f *File) ReadLine() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return "", fmt.Errorf("file is closed")
	}

	var result []byte
	var b [1]byte

	for {
		_, err := f.Handle.Read(b[:])
		if err != nil {
			if err == io.EOF {
				if len(result) > 0 {
					return string(result), nil
				}
				return "", io.EOF
			}
			return "", err
		}

		f.Position++

		if b[0] == '\n' {
			break
		}
		if b[0] == '\r' {
			// Peek next byte for \r\n
			next := make([]byte, 1)
			_, err := f.Handle.Read(next)
			if err == nil && next[0] == '\n' {
				f.Position++
			} else if err == nil {
				// Not \r\n, seek back
				f.Handle.Seek(-1, io.SeekCurrent)
			}
			break
		}
		result = append(result, b[0])
	}

	return string(result), nil
}

// ReadAll reads all remaining content from the file.
func (f *File) ReadAll() ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return nil, fmt.Errorf("file is closed")
	}

	data, err := io.ReadAll(f.Handle)
	if err != nil {
		return nil, err
	}

	f.Position += int64(len(data))
	return data, nil
}

// Write writes data to the file at the current position.
// Returns the number of bytes written.
func (f *File) Write(data []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return 0, fmt.Errorf("file is closed")
	}

	n, err := f.Handle.Write(data)
	if err != nil {
		return n, err
	}

	f.Position += int64(n)
	return n, nil
}

// WriteString writes a string to the file.
func (f *File) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}

// Seek sets the file position.
// Whence: 0 = start, 1 = current, 2 = end
func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return 0, fmt.Errorf("file is closed")
	}

	pos, err := f.Handle.Seek(offset, whence)
	if err != nil {
		return pos, err
	}

	f.Position = pos
	return pos, nil
}

// Tell returns the current file position.
func (f *File) Tell() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Position
}

// Flush flushes any buffered data to disk.
func (f *File) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return fmt.Errorf("file is closed")
	}

	return f.Handle.Sync()
}

// Truncate truncates the file to the specified size.
func (f *File) Truncate(size int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return fmt.Errorf("file is closed")
	}

	return f.Handle.Truncate(size)
}

// Stat returns file information.
func (f *File) Stat() (os.FileInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return nil, fmt.Errorf("file is closed")
	}

	return f.Handle.Stat()
}

// IsOpen returns whether the file is open.
func (f *File) IsOpen() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Open
}

// GetName returns the file name (base name).
func (f *File) GetName() string {
	return f.Path
}

// GetMode returns the file mode.
func (f *File) GetMode() FileMode {
	return f.Mode
}

// Lock places a lock on the file.
// lockType: 1 = shared, 2 = exclusive
// Non-blocking: returns immediately if lock cannot be acquired.
func (f *File) Lock(lockType FileLockType, blocking bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return fmt.Errorf("file is closed")
	}

	if f.LockType != LockNone {
		return fmt.Errorf("file already locked")
	}

	err := f.lockFile(lockType, blocking)
	if err == nil {
		f.LockType = lockType
	}
	return err
}

// Unlock releases the file lock.
func (f *File) Unlock() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.Open {
		return fmt.Errorf("file is closed")
	}

	if f.LockType == LockNone {
		return nil
	}

	err := f.unlockFile()
	if err == nil {
		f.LockType = LockNone
	}
	return err
}


// ============================================================
// FileInfo Object
// ============================================================

// FileInfo wraps os.FileInfo for use in Xxlang.
type FileInfo struct {
	Name     string    // Base name of the file
	Size     int64     // Length in bytes
	Mode     uint32    // File mode bits
	ModTime  time.Time // Modification time
	IsDir    bool      // Whether it's a directory
	FullPath string    // Full path to the file
}

// NewFileInfo creates a new FileInfo from os.FileInfo.
func NewFileInfo(info os.FileInfo, fullPath string) *FileInfo {
	return &FileInfo{
		Name:     info.Name(),
		Size:     info.Size(),
		Mode:     uint32(info.Mode()),
		ModTime:  info.ModTime(),
		IsDir:    info.IsDir(),
		FullPath: fullPath,
	}
}

// Type returns the object type.
func (fi *FileInfo) Type() ObjectType { return FileInfoType }

// TypeTag returns the type tag for fast type checking.
func (fi *FileInfo) TypeTag() TypeTag { return TagFileInfo }

// Inspect returns a string representation of the FileInfo.
func (fi *FileInfo) Inspect() string {
	dir := ""
	if fi.IsDir {
		dir = " [DIR]"
	}
	return fmt.Sprintf("<FILE_INFO name=%s size=%d mode=%03o%s>",
		fi.Name, fi.Size, fi.Mode, dir)
}

// ToBool always returns true for FileInfo.
func (fi *FileInfo) ToBool() *Bool { return TRUE }

// HashKey returns a hash key for the FileInfo.
func (fi *FileInfo) HashKey() HashKey {
	return HashKey{Type: FileInfoType, Value: uint64(fi.Size) ^ uint64(fi.ModTime.UnixNano())}
}

// GetModTimeString returns the modification time as a formatted string.
func (fi *FileInfo) GetModTimeString() string {
	return fi.ModTime.Format("2006-01-02 15:04:05")
}

// GetModTimeUnix returns the modification time as Unix timestamp (milliseconds).
func (fi *FileInfo) GetModTimeUnix() int64 {
	return fi.ModTime.UnixMilli()
}

// GetModeString returns the file mode as an octal string.
func (fi *FileInfo) GetModeString() string {
	return fmt.Sprintf("%03o", fi.Mode)
}

// IsRegular returns true if this is a regular file.
func (fi *FileInfo) IsRegular() bool {
	return !fi.IsDir && (fi.Mode&uint32(os.ModeType)) == 0
}

// IsSymlink returns true if this is a symbolic link.
func (fi *FileInfo) IsSymlink() bool {
	return (fi.Mode & uint32(os.ModeSymlink)) != 0
}

// ============================================================
// File Constants
// ============================================================

// Seek constants for use with File.seek()
const (
	SeekStart   = 0 // Seek relative to start of file
	SeekCurrent = 1 // Seek relative to current position
	SeekEnd     = 2 // Seek relative to end of file
)
