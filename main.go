package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// Styles for different elements
var (
	directoryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	fileStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	executableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	selectedStyle  = lipgloss.NewStyle().Reverse(true)
	statusStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	headerStyle    = lipgloss.NewStyle().Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func (m Model) renderTreeItem(item FileItem, i int, depth int, isLastAtLevel bool) string {
	// Calculate indentation and tree symbols
	prefix := strings.Repeat(" ", m.config.Display.IndentSize * (depth-1))
	if depth > 0 {
		if isLastAtLevel {
			prefix += m.config.treeSymbols.Corner + m.config.treeSymbols.Horizontal
		} else {
			prefix += m.config.treeSymbols.Tee + m.config.treeSymbols.Horizontal
		}
	}

	// Add cursor indicator
	if i == m.tree.cursor {
		prefix += "> "
	} else {
		prefix += "  "
	}

	// Get appropriate icon and style
	var icon string
	var itemStyle lipgloss.Style
	if item.isDir {
		itemStyle = directoryStyle
		if item.name == ".." {
			icon = m.config.icons.ParentDir
		} else if m.tree.expanded[item.path] {
			icon = m.config.icons.DirectoryOpen
		} else {
			icon = m.config.icons.Directory
		}
	} else {
		if item.mode&0111 != 0 { // Executable file
			itemStyle = executableStyle
		} else {
			itemStyle = fileStyle
		}
		icon = m.config.icons.GetFileIcon(item)
	}
	
	// Construct item text
	itemText := fmt.Sprintf("%s%s %s", prefix, icon, item.name)
	if !item.isDir {
		itemText += fmt.Sprintf(" (%s)", item.mode.String())
	}

	// Apply highlighting if selected
	if i == m.tree.cursor {
		return selectedStyle.Render(itemText)
	}
	return itemStyle.Render(itemText)
}

func (m Model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	switch m.activeView {
	case TreeView:
		// Show current directory at top
		currentDirText := fmt.Sprintf("Directory: %s", m.config.CurrentDir)
		b.WriteString(headerStyle.Render(currentDirText) + "\n\n")

		// Calculate last items at each level for proper tree lines
		lastAtLevel := make(map[int]bool)
		for i, item := range m.tree.items {
			depth := strings.Count(item.path, string(os.PathSeparator)) -
				strings.Count(m.config.CurrentDir, string(os.PathSeparator))
			if depth < 0 {
				depth = 0
			}

			// Check if this is the last item at its level
			isLast := true
			for j := i + 1; j < len(m.tree.items); j++ {
				nextDepth := strings.Count(m.tree.items[j].path, string(os.PathSeparator)) -
					strings.Count(m.config.CurrentDir, string(os.PathSeparator))
				if nextDepth <= depth {
					isLast = false
					break
				}
			}
			lastAtLevel[depth] = isLast

			// Render the item
			b.WriteString(m.renderTreeItem(item, i, depth, isLast))
			b.WriteString("\n")
		}

		// Add status and help text
		b.WriteString("\n")
		if m.status != "" {
			b.WriteString(statusStyle.Render(m.status) + "\n")
		}
		helpText := "\nj/k: move   h/l: collapse/expand   .: toggle hidden   q: quit"
		b.WriteString(helpText)

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

	case ".":
		m.tree.showHidden = !m.tree.showHidden
		return m, m.tree.LoadDirectory(m.config.CurrentDir)
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