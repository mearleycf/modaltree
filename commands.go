package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// FileOperation represents a file operation (move, copy, delete)
type FileOperation struct {
	Type OperationType
	Source string
	Dest string
	Selected *FileItem
	state *OperationState
}

// NewFileOperation creates a new file operation with initialized state
func NewFileOperation(opType OperationType, source, dest string, selected *FileItem) FileOperation {
	op := FileOperation{
		Type: opType,
		Source: source,
		Dest: dest,
		Selected: selected,
	}
	op.state = &OperationState{
		Operation: op,
		StartTime: time.Now(),
		Stage: StageInit,
	}
	return op
}

// OperationState tracks the state of a file operation
type OperationState struct {
	Operation  FileOperation
	BackupPath string
	StartTime  time.Time
	RetryCount int
	LastError  error
	Stage      OperationStage
	Progress   float64    // Add this field
}

// Add this function
func handleOperationError(op FileOperation, err error, backup string) {
	op.state.Stage = StageFailed
	op.state.LastError = err
	if backup != "" {
		if restoreErr := restoreBackup(backup, op.Source); restoreErr != nil {
			op.state.LastError = fmt.Errorf("restore failed: %v (original: %v)", restoreErr, err)
		}
		op.state.Stage = StageRestored
	}
}

type OperationStage int

const (
	StageInit OperationStage = iota
	StageValidated
	StageBackedUp
	StageExecuting
	StageCompleted
	StageFailed
	StageRestored
)

type OperationType int

const (
	OpMove OperationType = iota
	OpCopy
	OpDelete
	OpRename
	MaxRetries = 3
)

// ValidatePermissions checks if we have required permissions for the operation
func ValidatePermissions(op FileOperation) error {
	// Check source permissions
	_, err := os.Stat(op.Source)
	if err != nil {
		return fmt.Errorf("cannot access source: %w", err)
	}

	// For all operations, need read permission on source
	if err := unix.Access(op.Source, unix.R_OK); err != nil {
		return fmt.Errorf("no read permission on source: %w", err)
	}

	// For delete/move operations, need write permission on source parent
	if op.Type == OpDelete || op.Type == OpMove || op.Type == OpRename {
		sourceParent := filepath.Dir(op.Source)
		if err := unix.Access(sourceParent, unix.W_OK); err != nil {
			return fmt.Errorf("no write permission on source directory: %w", err)
		}
	}

	// For move/copy operations, check destination
	if op.Type == OpMove || op.Type == OpCopy {
		destParent := filepath.Dir(op.Dest)
		
		// Check if destination parent exists
		if _, err := os.Stat(destParent); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("destination directory does not exist")
			}
			return fmt.Errorf("cannot access destination: %w", err)
		}

		// Check write permission on destination
		if err := unix.Access(destParent, unix.W_OK); err != nil {
			return fmt.Errorf("no write permission on destination directory: %w", err)
		}

		// Check for destination collision
		if _, err := os.Stat(op.Dest); err == nil {
			return fmt.Errorf("destination already exists")
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("cannot check destination: %w", err)
		}
	}

	return nil
}
// createBackup creates a backup of the file/directory being operated on
func createBackup(path string) (string, error) {
	backupPath := fmt.Sprintf("%s.bak.%d", path, time.Now().UnixNano())
	if err := CopyFile(path, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}
	return backupPath, nil
}

// restoreBackup restores from backup and cleans up
func restoreBackup(backupPath, originalPath string) error {
	if err := os.RemoveAll(originalPath); err != nil {
		return fmt.Errorf("failed to remove failed operation result: %w", err)
	}
	if err := os.Rename(backupPath, originalPath); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}
	return nil
}

// retryOperation attempts an operation with retries
func retryOperation(op func() error) error {
	var lastErr error
	for i := 0; i < MaxRetries; i++ {
		if err := op(); err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		return nil
	}
	return fmt.Errorf("operation failed after %d retries: %w", MaxRetries, lastErr)
}

// ExecuteFileOperation executes the given file operation
func ExecuteFileOperation(op FileOperation) error {
	if op.Selected == nil {
		return fmt.Errorf("no file selected for operation")
	}

	// Start progress tracking
	op.state.Stage = StageInit
	op.state.Progress = 0

	// Validate permissions before attempting operation
	if err := ValidatePermissions(op); err != nil {
		op.state.Stage = StageFailed
		op.state.LastError = err
		return fmt.Errorf("permission check failed: %w", err)
	}
	op.state.Stage = StageValidated
	op.state.Progress = 25

	// Create backup for destructive operations
	var backup string
	var err error
	if op.Type == OpMove || op.Type == OpDelete || op.Type == OpRename {
		backup, err = createBackup(op.Source)
		if err != nil {
			op.state.Stage = StageFailed
			op.state.LastError = err
			return err
		}
		op.state.BackupPath = backup
		op.state.Stage = StageBackedUp
		op.state.Progress = 50
		defer os.Remove(backup)
	}

	op.state.Stage = StageExecuting
	op.state.Progress = 75

	// Execute operation with retries and progress updates
	err = retryOperation(func() error {
		op.state.RetryCount++
		return executeWithProgress(op)
	})

	if err != nil {
		handleOperationError(op, err, backup)
		return err
	}

	op.state.Stage = StageCompleted
	op.state.Progress = 100
	return nil
}

func executeWithProgress(op FileOperation) error {
	switch op.Type {
	case OpMove:
		return os.Rename(op.Source, op.Dest)
	case OpCopy:
		return CopyFileWithProgress(op)
	case OpDelete:
		if op.Selected.isDir {
			return os.RemoveAll(op.Source)
		}
		return os.Remove(op.Source)
	case OpRename:
		return os.Rename(op.Source, op.Dest)
	default:
		return fmt.Errorf("unsupported file operation type: %v", op.Type)
	}
}

// CopyFileWithProgress copies a file from source to destination with progress tracking
func CopyFileWithProgress(op FileOperation) error {
	sourceInfo, err := os.Stat(op.Source)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		return CopyDirWithProgress(op)
	}

	totalSize := sourceInfo.Size()
	source, err := os.Open(op.Source)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(op.Dest)
	if err != nil {
		return err
	}
	defer dest.Close()

	buf := make([]byte, 32*1024)
	var bytesWritten int64

	for {
		n, err := source.Read(buf)
		if n > 0 {
			_, writeErr := dest.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			bytesWritten += int64(n)
			op.state.Progress = float64(bytesWritten) / float64(totalSize) * 100
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// CopyFile copies a file from source to destination
func CopyFile(src, dst string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		return CopyDir(src, dst)
	}

	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, sourceInfo.Mode())
}
// CopyDir recursively copies a directory
func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// LaunchEditor launches the user's preferred editor
func LaunchEditor(editor, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// ChangeShellDirectory writes a command to change the shell's directory
func ChangeShellDirectory(dir string) error {
	// create or truncate the file
	f, err := os.Create("/tmp/modaltree_lastdir")
	if err != nil {
		return err
	}
	defer f.Close()

	// write the directory path
	_, err = f.WriteString(dir)
	return err
}


// Add new message type for progress updates
type OperationProgress struct {
    Progress float64
}

// Add to existing code
func CopyDirWithProgress(op FileOperation) error {
    // Get total items to process
    var totalItems int
    filepath.Walk(op.Source, func(path string, info os.FileInfo, err error) error {
        totalItems++
        return nil
    })

    itemsProcessed := 0

    srcInfo, err := os.Stat(op.Source)
    if err != nil {
        return err
    }

    err = os.MkdirAll(op.Dest, srcInfo.Mode())
    if err != nil {
        return err
    }

    entries, err := os.ReadDir(op.Source)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        srcPath := filepath.Join(op.Source, entry.Name())
        dstPath := filepath.Join(op.Dest, entry.Name())

        if entry.IsDir() {
            subOp := op
            subOp.Source = srcPath
            subOp.Dest = dstPath
            err = CopyDirWithProgress(subOp)
        } else {
            err = CopyFileWithProgress(FileOperation{
                Type: OpCopy,
                Source: srcPath,
                Dest: dstPath,
                state: op.state,
            })
        }

        if err != nil {
            return err
        }

        itemsProcessed++
        op.state.Progress = float64(itemsProcessed) / float64(totalItems) * 100
    }

    return nil
}
