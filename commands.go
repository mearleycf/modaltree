package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// FileOperation represents a file operation (move, copy, delete)
type FileOperation struct {
	Type OperationType
	Source string
	Dest string
	Selected *FileItem
}

type OperationType int

const (
	OpMove OperationType = iota
	OpCopy
	OpDelete
	OpRename
)

// ExecuteFileOperation executes the given file operation
func ExecuteFileOperation(op FileOperation) error {
	if op.Selected == nil {
		return fmt.Errorf("no file selected for operation")
	}

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

