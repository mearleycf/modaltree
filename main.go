package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Config holds the application configuration
type Config struct {
	ShowHidden      bool          // Show hidden files by default
	Editor          string        // Default editor command
	ConfirmActions  bool          // Whether to confirm destructive actions
	CurrentDir      string        // Current working directory
	Display         DisplayConfig // Display configuration
	icons           IconSet       // Current icon set (determined by Display.UseNerdFont)
	treeSymbols     TreeSymbols   // Current tree symbols (determined by Display.TreeStyle)
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

	display := DefaultDisplayConfig()
	
	config := Config{
		ShowHidden:     true,
		Editor:         "code",
		ConfirmActions: true,
		CurrentDir:     cwd,
		Display:        display,
		icons:          UnicodeIconSet(),
		treeSymbols:    UnicodeTreeSymbols(),
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

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	var b strings.Builder

	switch m.activeView {
	case TreeView:
		// Show current directory at top
		b.WriteString(fmt.Sprintf("Directory: %s\n\n", m.config.CurrentDir))

		// Render tree items
		for i, item := range m.tree.items {
			// Calculate the indentation level based on path depth
			depth := strings.Count(item.path, string(os.PathSeparator)) -
				strings.Count(m.config.CurrentDir, string(os.PathSeparator))
			if depth < 0 {
				depth = 0
			}

			// Build the tree structure
			prefix := ""
			for d := 0; d < depth; d++ {
				isLast := false
				if d == depth-1 {
					// Check if this is the last item at this level
					isLast = true
					for j := i + 1; j < len(m.tree.items); j++ {
						otherDepth := strings.Count(m.tree.items[j].path, string(os.PathSeparator)) -
							strings.Count(m.config.CurrentDir, string(os.PathSeparator))
						if otherDepth <= d {
							isLast = false
							break
						}
					}
				}

				if d == depth-1 {
					if isLast {
						prefix += m.config.treeSymbols.Corner + m.config.treeSymbols.Horizontal
					} else {
						prefix += m.config.treeSymbols.Tee + m.config.treeSymbols.Horizontal
					}
				} else {
					if isLast {
						prefix += "  "
					} else {
						prefix += m.config.treeSymbols.Vertical + " "
					}
				}
			}

			// Add cursor indicator
			if i == m.tree.cursor {
				prefix += "> "
			} else {
				prefix += "  "
			}

			// Get the appropriate icon
			var icon string
			if item.isDir {
				if item.name == ".." {
					icon = m.config.icons.ParentDir
				} else if m.tree.expanded[item.path] {
					icon = m.config.icons.DirectoryOpen
				} else {
					icon = m.config.icons.Directory
				}
			} else {
				icon = m.config.icons.GetFileIcon(item)
			}
			prefix += icon + " "

			// Add the item name and optional permission info
			itemText := item.name
			if !item.isDir {
				itemText += fmt.Sprintf(" (%s)", item.mode.String())
			}

			// Highlight cursor line
			if i == m.tree.cursor {
				b.WriteString(fmt.Sprintf("\x1b[7m%s%s\x1b[0m\n", prefix, itemText))
			} else {
				b.WriteString(fmt.Sprintf("%s%s\n", prefix, itemText))
			}
		}

		// Add status line at the bottom with help text
		b.WriteString("\n")
		if m.status != "" {
			b.WriteString(fmt.Sprintf("%s\n", m.status))
		}
		b.WriteString("\nj/k: move   h/l: collapse/expand  .: toggle hidden  q: quit")

	case InputView:
		if m.input != nil {
			return m.input.View()
		}

	case ConfirmView:
		if m.input != nil {
			return "Confirm View" // placeholder, we'll implement this later
		}
	}

	return b.String()
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

func (m Model) SaveConfig() {
	if err := SaveConfig(m.config); err != nil {
		m.status = fmt.Sprintf("Error saving config: %v", err)
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}