package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// Config holds the application configuration
type Config struct {
	ShowHidden      bool   // Show hidden files by default
	Editor          string // Default editor command
	ConfirmActions  bool   // Whether to confirm destructive actions
	CurrentDir      string // Current working directory
}

// Model represents the application state
type Model struct {
	config     Config
	tree       *FileTree
	input      *Input
	status     string
	err        error
	activeView View // current view (tree, input, confirm)
}

type View int

const (
	TreeView View = iota
	InputView
	ConfirmView
)

// Initial setup function
func initialModel() Model {
	cwd, err := os.Getwd()
	if err != nil {
		// Handle error gracefully
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}

	config := Config{
		ShowHidden:     true,
		Editor:         "code",
		ConfirmActions: true,
		CurrentDir:     cwd,
	}

	return Model{
		config:     config,
		tree:       NewFileTree(cwd),
		activeView: TreeView,
	}
}

func (m Model) Init() tea.Cmd {
	// Initial command to load directory contents
	return m.tree.LoadDirectory(m.config.CurrentDir)
}

// Update handles all the key events and updates the model accordingly
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.activeView {
		case TreeView:
			return m.handleTreeViewKeys(msg)
		case InputView:
			return m.handleInputViewKeys(msg)
		case ConfirmView:
			return m.handleConfirmViewKeys(msg)
		}
	
	case loadedDirectoryMsg:
		m.tree.items = msg.items
		return m, nil
	
	case errMsg:
		m.err = msg.error
		m.status = fmt.Sprintf("Error: %v", msg.error)
		return m, nil
	}

	return m, nil
}

func (m Model) handleTreeViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	
	case "up", "k":
		m.tree.MoveUp()
	
	case "down", "j":
		m.tree.MoveDown()
	
	case "enter", "right", "l":
		if item := m.tree.GetSelectedItem(); item != nil {
			if item.isDir {
				return m, m.tree.ToggleExpand()
			}
		}
	
	case "left", "h":
    if item := m.tree.GetSelectedItem(); item != nil {
        if item.name == ".." {
            // Move up one directory level
            m.config.CurrentDir = filepath.Dir(m.config.CurrentDir)
            return m, m.tree.LoadDirectory(m.config.CurrentDir)
        } else if item.isDir && m.tree.expanded[item.path] {
            // If directory is expanded, collapse it
            delete(m.tree.expanded, item.path)
            return m, nil
        } else {
            // If it's a file or collapsed directory, try to move to parent directory
            parentDir := filepath.Dir(item.path)
            if parentDir != m.config.CurrentDir {
                m.config.CurrentDir = parentDir
                return m, m.tree.LoadDirectory(m.config.CurrentDir)
            }
        }
		}

	case "p":
		if item := m.tree.GetSelectedItem(); item != nil {
			m.input = NewInput(InputRename, item.name)
			m.activeView = InputView
			return m, nil
		}
	}
	return m, nil
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}