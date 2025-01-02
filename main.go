package main

import (
	"context"
	"flag"
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
	activeView View
	cleanup    context.CancelFunc
	statusBar  *StatusBar
}

type View int

const (
	TreeView View = iota
	InputView
	ConfirmView
)

// Initial setup function
func initialModel() (Model, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Model{}, fmt.Errorf("failed to get working directory: %w", err)
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
		statusBar: NewStatusBar(),
	}, nil
}
func (m Model) Init() tea.Cmd {
	return m.tree.LoadDirectory(m.config.CurrentDir)
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
		icon = m.config.icons.GetFileIcon(item, m.config.Display)
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
	var b strings.Builder

	// Render main content based on active view
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

		// Preserve existing status and help text
		b.WriteString("\n")
		if m.status != "" {
			b.WriteString(statusStyle.Render(m.status) + "\n")
		}
		helpText := "\nj/k: move   h/l: collapse/expand   .: toggle hidden   n: toggle nerd fonts   q: quit"
		b.WriteString(helpText)

		// Add status bar below help text
		b.WriteString("\n")
		b.WriteString(m.statusBar.View())

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
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Update status bar width
		m.statusBar.Update(msg)

	case FileOperation:
		m.statusBar.StartProgress()
		m.statusBar.UpdateOperation(msg.state)
		m.statusBar.UpdateProgress(msg.state.Progress)
		return m, nil

	case OperationProgress:
		m.statusBar.UpdateProgress(msg.Progress)
		return m, nil

	case loadedDirectoryMsg:
		m.tree.items = msg.items
		m.statusBar.UpdatePath(m.config.CurrentDir)
		return m, nil

	case errMsg:
		m.err = msg.error
		m.status = fmt.Sprintf("Error: %v", msg.error)
		m.statusBar.setMessage(fmt.Sprintf("Error: %v", msg.error), MessageError)
		return m, nil
	}

	switch m.activeView {
	case TreeView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleTreeViewKeys(keyMsg)
		}
	case InputView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleInputViewKeys(keyMsg) // implement this function
		}
	case ConfirmView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.handleConfirmViewKeys(keyMsg) // Implement this function
		}
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

	case ".":
		m.tree.showHidden = !m.tree.showHidden
		return m, m.tree.LoadDirectory(m.config.CurrentDir)
		
		case "n": // new toggle for nerd fonts
		  m.config.Display.UseNerdFont = !m.config.Display.UseNerdFont
			if m.config.Display.UseNerdFont {
				m.config.icons = NerdFontIconSet()
				m.config.Display.VerifyNerdFont()
			} else {
				m.config.icons = DefaultIconSet()
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
	// nerd font flag
	useNerdFont := flag.Bool("nerd-font", false, "Enable Nerd Font Icons")
	flag.Parse()

	model, err := initialModel()
	if err != nil {
		fmt.Printf("Error initializing model: %v", err)
		os.Exit(1)
	}

	// Update display config based on nerd font flag
	model.config.Display.UseNerdFont = *useNerdFont
	if *useNerdFont {
		model.config.icons = NerdFontIconSet()
		model.config.Display.VerifyNerdFont()
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

// Add validation method to Config struct
func (c *Config) Validate() error {
    if c.Editor == "" {
        return fmt.Errorf("editor cannot be empty")
    }
    if c.CurrentDir == "" {
        return fmt.Errorf("current directory cannot be empty")
    }
    return nil
}

// Extract view rendering interface
type Renderer interface {
    Render() string
}

// Make FileTree implement Renderer
func (ft *FileTree) Render(config Config) string {
    // Move tree rendering logic here
		return "" // Placeholder, replace with actual rendering logic
}

const (
    DefaultIndentSize = 2
    MinWindowWidth    = 80
    MaxItemsToDisplay = 1000
)

func NewStatusBar() *StatusBar {
	return &StatusBar{
		width: MinWindowWidth,
	}
}