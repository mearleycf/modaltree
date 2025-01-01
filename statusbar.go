package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar represents the status bar component that displays current path,
// operation status, and system messages
type StatusBar struct {
	width       int
	currentPath string
	message     string
	messageType MessageType
	operation   *OperationState
}

// MessageType defines the type of message being displayed in the status bar
type MessageType int

// Message type constants for different status indicators
const (
	MessageNormal MessageType = iota
	MessageError
	MessageSuccess
)

// Style definitions for the status bar and its elements
var (
	// statusBarStyle defines the main container style for the status bar
	statusBarStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		Width(100).
		Padding(0, 1)

	// normalMessageStyle defines the style for standard status messages
	normalMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))
	
	// errorMessageStyle defines the style for error messages
	errorMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1"))
	
	// successMessageStyle defines the style for success messages
	successMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2"))
)

// NewStatusBar creates and initializes a new StatusBar instance
func NewStatusBar() StatusBar {
	return StatusBar{
		messageType: MessageNormal,
	}
}

// Init initializes the status bar component
// Implements tea.Model interface
func (s StatusBar) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the status bar state
// Implements tea.Model interface
func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

// View renders the status bar component
// Implements tea.Model interface
func (s StatusBar) View() string {
	// Select appropriate message style based on message type
	var messageStyle lipgloss.Style
	switch s.messageType {
	case MessageError:
		messageStyle = errorMessageStyle
	case MessageSuccess:
		messageStyle = successMessageStyle
	default:
		messageStyle = normalMessageStyle
	}

	// Compose the status bar content with current path and message
	content := lipgloss.JoinHorizontal(
		lipgloss.Left,
		s.currentPath,
		messageStyle.Render(s.message),
	)

	return statusBarStyle.Width(s.width).Render(content)
}

// UpdateOperation updates the status bar with current operation state
func (s *StatusBar) UpdateOperation(op *OperationState) {
	s.operation = op
	if op == nil {
		s.clearMessage()
		return
	}

	// Generate appropriate message based on operation stage
	switch op.Stage {
	case StageInit:
		s.setMessage("Preparing operation...", MessageNormal)
	case StageValidated:
		s.setMessage("Permissions validated", MessageSuccess)
	case StageBackedUp:
		s.setMessage("Backup created", MessageSuccess)
	case StageExecuting:
		msg := fmt.Sprintf("Executing %s operation (attempt %d/%d)", 
			getOperationName(op.Operation.Type), 
			op.RetryCount, 
			MaxRetries)
		s.setMessage(msg, MessageNormal)
	case StageCompleted:
		s.setMessage("Operation completed successfully", MessageSuccess)
	case StageFailed:
		s.setMessage(fmt.Sprintf("Operation failed: %v", op.LastError), MessageError)
	case StageRestored:
		s.setMessage("Operation failed, backup restored", MessageError)
	}
}

// setMessage updates the status bar message and type
func (s *StatusBar) setMessage(msg string, msgType MessageType) {
	s.message = msg
	s.messageType = msgType
}

// clearMessage resets the status bar message
func (s *StatusBar) clearMessage() {
	s.message = ""
	s.messageType = MessageNormal
	s.operation = nil
}

// UpdatePath updates the current path display in the status bar
func (s *StatusBar) UpdatePath(path string) {
	s.currentPath = path
}

// getOperationName returns a human-readable operation name
func getOperationName(opType OperationType) string {
	switch opType {
	case OpMove:
		return "move"
	case OpCopy:
		return "copy"
	case OpDelete:
		return "delete"
	case OpRename:
		return "rename"
	default:
		return "unknown"
	}
}
