package main

import (
	"fmt"
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
	Operation FileOperation
	BackupPath string
	StartTime time.Time
	RetryCount int
	LastError error
	Stage OperationStage
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
	sourceInfo, err := os.Stat(op.Source)
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

	// Validate permissions before attempting operation
	if err := ValidatePermissions(op); err != nil {
		op.state.Stage = StageFailed
		op.state.LastError = err
		return fmt.Errorf("permission check failed: %w", err)
	}
	op.state.Stage = StageValidated

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
		defer os.Remove(backup) // Clean up backup on success
	}

	op.state.Stage = StageExecuting
	// Execute operation with retries
	err = retryOperation(func() error {
		op.state.RetryCount++
		switch op.Type {
		case OpMove:
			return os.Rename(op.Source, op.Dest)
		case OpCopy:
			return CopyFile(op.Source, op.Dest)
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
	})

	// If operation failed and we have a backup, try to restore
	if err != nil {
		op.state.Stage = StageFailed
		op.state.LastError = err

		if backup != "" {
			if restoreErr := restoreBackup(backup, op.Source); restoreErr != nil {
				return fmt.Errorf("operation failed and backup restoration failed: %v (original error: %v)", restoreErr, err)
			}
			op.state.Stage = StageRestored
			return fmt.Errorf("operation failed but backup restored: %w", err)
		}
	}

	op.state.Stage = StageCompleted
	return err
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

