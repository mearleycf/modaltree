// input_test.go
package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewInput(t *testing.T) {
    tests := []struct {
        name         string
        inputType    InputType
        initialValue string
        wantPrompt   string
    }{
        {"Rename", InputRename, "test.txt", "Rename to: "},
        {"Move", InputMove, "dir/", "Move to: "},
        {"Copy", InputCopy, "file.go", "Copy to: "},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            input := NewInput(tt.inputType, tt.initialValue)
            if input.prompt != tt.wantPrompt {
                t.Errorf("got prompt %q, want %q", input.prompt, tt.wantPrompt)
            }
            if input.value != tt.initialValue {
                t.Errorf("got value %q, want %q", input.value, tt.initialValue)
            }
        })
    }
}

func TestInputUpdate(t *testing.T) {
    tests := []struct {
        name     string
        msg      tea.Msg
        initial  string
        want     string
        done     bool
        cursorPos int
    }{
        {
            name:     "Type character",
            msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
            initial:  "test",
            want:     "testa",
            done:     false,
            cursorPos: 5,
        },
        {
            name:     "Backspace",
            msg:      tea.KeyMsg{Type: tea.KeyBackspace},
            initial:  "test",
            want:     "tes",
            done:     false,
            cursorPos: 3,
        },
        {
            name:     "Enter confirms",
            msg:      tea.KeyMsg{Type: tea.KeyEnter},
            initial:  "test",
            want:     "test",
            done:     true,
            cursorPos: 4,
        },
        {
            name:     "Escape cancels",
            msg:      tea.KeyMsg{Type: tea.KeyEsc},
            initial:  "test",
            want:     "",
            done:     false,
            cursorPos: 4,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            input := NewInput(InputRename, tt.initial)
            got, done, _ := input.Update(tt.msg)
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
            if done != tt.done {
                t.Errorf("got done=%v, want done=%v", done, tt.done)
            }
            if input.cursorPos != tt.cursorPos {
                t.Errorf("got cursorPos=%d, want cursorPos=%d", input.cursorPos, tt.cursorPos)
            }
        })
    }
}

func TestInputView(t *testing.T) {
    input := NewInput(InputRename, "test")
    input.cursorPos = 2
    got := input.View()
    want := "Rename to: te█st"
    if got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}