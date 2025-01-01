package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar represents the status bar component
type StatusBar struct {
	width       int
	currentPath string
	message     string
	messageType MessageType
	operation   *OperationState
	progress    float64
	isActive    bool
	spinner     spinner.Model
	isCanceling bool
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

// stage descriptions for different operation stages
var stageDescriptions = map[OperationStage]string{
	StageInit:      "Preparing",
	StageValidated: "Validated",
	StageBackedUp:  "Backed up",
	StageExecuting: "Executing",
	StageCompleted: "Completed",
	StageFailed:    "Failed",
	StageRestored:  "Restored",
}

// Add method to handle cancellation state
func (s *StatusBar) SetCanceling(canceling bool) {
	s.isCanceling = canceling
	if canceling {
		s.setMessage("Canceling operation...", MessageNormal)
	}
}

// View renders the status bar component
func (s StatusBar) View() string {
	var content string
	if s.isActive && s.operation != nil {
		stageText := stageDescriptions[s.operation.Stage]
		if s.isCanceling {
			stageText = "Canceling"
		}
	
		progressBar := fmt.Sprintf("[%s: %.0f%%] %s", 
			stageText,
			s.progress, 
			s.spinner.View())
	
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			s.currentPath,
			" | ",
			progressBar,
			" | ",
			s.getMessageWithStyle(),
		)
	} else {
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			s.currentPath,
			" | ",
			s.getMessageWithStyle(),
		)
	}

	return statusBarStyle.Width(s.width).Render(content)
}
func (s *StatusBar) StartProgress() {
	s.isActive = true
	s.progress = 0
	s.spinner = spinner.New()
	s.spinner.Spinner = spinner.Line
}

func (s *StatusBar) UpdateProgress(progress float64) {
	s.progress = progress
}

func (s *StatusBar) StopProgress() {
	s.isActive = false
	s.progress = 0
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

// getMessageWithStyle returns the status message with appropriate styling
func (s StatusBar) getMessageWithStyle() string {
    switch s.messageType {
    case MessageError:
        return errorMessageStyle.Render(s.message)
    case MessageSuccess:
        return successMessageStyle.Render(s.message)
    default:
        return normalMessageStyle.Render(s.message)
    }
}

func (s *StatusBar) Update(msg tea.WindowSizeMsg) {
    s.width = msg.Width
}
