// main_test.go
package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func InitialModel() Model {
	return Model{
		activeView: TreeView,
		input:      nil, // initially no input active
	}
}

func TestModelInitialization(t *testing.T) {
    m := InitialModel()
    if m.activeView != TreeView {
        t.Errorf("Expected initial view to be TreeView, got %v", m.activeView)
    }
}

func TestHandleInputViewKeys(t *testing.T) {
    m := Model{
        activeView: InputView,
        input:      NewInput(InputRename, "test.txt"),
    }
    
    // Test ESC key
    newModel, _ := m.handleInputViewKeys(tea.KeyMsg{Type: tea.KeyEsc})
		result := newModel.(Model)
    if result.activeView != TreeView {
			t.Error("Expected view to change to TreeView on ESC")
	}
	if result.input != nil {
			t.Error("Expected input to be nil after ESC")
	}
}

func TestHandleConfirmViewKeys(t *testing.T) {
    tests := []struct {
        name     string
        key      string
        wantView View // changed
    }{
        {"Confirm Yes", "y", TreeView},
        {"Confirm No", "n", TreeView},
        {"Quit", "q", TreeView},
        {"Escape", "esc", TreeView},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := Model{activeView: ConfirmView}
						newModel, _ := m.handleConfirmViewKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
						result := newModel.(Model) // add type assertion
            if result.activeView != tt.wantView {
                t.Errorf("Expected view %v, got %v", tt.wantView, result.activeView)
            }
        })
    }
}