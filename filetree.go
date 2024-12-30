package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
)

type FileTree struct {
	root      string
	items     []FileItem
	cursor    int
	expanded  map[string]bool
	showHidden bool
}

type FileItem struct {
	path     string
	name     string
	isDir    bool
	mode     fs.FileMode
}

func NewFileTree(root string) *FileTree {
	return &FileTree{
		root:      root,
		expanded:  make(map[string]bool),
		showHidden: true,
	}
}

// LoadDirectory reads the directory contents and returns a command
func (t *FileTree) LoadDirectory(dir string) tea.Cmd {
	return func() tea.Msg {
		items := []FileItem{}
		
		// Add parent directory entry if not at root
		if dir != "/" {
			items = append(items, FileItem{
				path:  filepath.Dir(dir),
				name:  "..",
				isDir: true,
			})
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return errMsg{err}
		}

		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			name := entry.Name()
			// Skip hidden files if showHidden is false
			if !t.showHidden && name[0] == '.' && name != ".." {
				continue
			}

			items = append(items, FileItem{
				path:  filepath.Join(dir, name),
				name:  name,
				isDir: entry.IsDir(),
				mode:  info.Mode(),
			})
		}

		// Sort items: directories first, then files, both alphabetically
		sort.Slice(items, func(i, j int) bool {
			if items[i].isDir != items[j].isDir {
				return items[i].isDir
			}
			return items[i].name < items[j].name
		})

		return loadedDirectoryMsg{items}
	}
}

// MoveUp moves the cursor up
func (t *FileTree) MoveUp() {
	if t.cursor > 0 {
		t.cursor--
	}
}

// MoveDown moves the cursor down
func (t *FileTree) MoveDown() {
	if t.cursor < len(t.items)-1 {
		t.cursor++
	}
}

// ToggleExpand expands or collapses the current directory
func (t *FileTree) ToggleExpand() tea.Cmd {
	if t.cursor >= len(t.items) {
		return nil
	}

	item := t.items[t.cursor]
	if !item.isDir {
		return nil
	}

	if t.expanded[item.path] {
		delete(t.expanded, item.path)
	} else {
		t.expanded[item.path] = true
		return t.LoadDirectory(item.path)
	}

	return nil
}

// ToggleHidden toggles visibility of hidden files
func (t *FileTree) ToggleHidden() tea.Cmd {
	t.showHidden = !t.showHidden
	return t.LoadDirectory(t.root)
}

// GetSelectedItem returns the currently selected FileItem
func (t *FileTree) GetSelectedItem() *FileItem {
	if t.cursor >= len(t.items) {
		return nil
	}
	return &t.items[t.cursor]
}

// Custom messages
type loadedDirectoryMsg struct {
	items []FileItem
}

type errMsg struct {
	error
}