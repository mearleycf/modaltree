package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type InputType int

const (
	InputRename InputType = iota
	InputMove
	InputCopy
)

type Input struct {
	value string
	prompt string
	inputType InputType
	cursorPos int
}

func NewInput(inputType InputType, initialValue string) *Input {
	prompts := map[InputType]string{
		InputRename: "Rename to: ",
		InputMove: "Move to: ",
		InputCopy: "Copy to: ",
	}

	return &Input{
		value: initialValue,
		prompt: prompts[inputType],
		inputType: inputType,
		cursorPos: len(initialValue),
	}
}

func (i *Input) Update(msg tea.Msg) (string, bool, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return i.value, true, nil
		case tea.KeyEsc:
			return "", false, nil
		case tea.KeyBackspace:
			if len(i.value) > 0 && i.cursorPos > 0 {
				i.value = i.value[:i.cursorPos-1] + i.value[i.cursorPos:]
				i.cursorPos--
			}
		case tea.KeyRight:
			if i.cursorPos < len(i.value) {
				i.cursorPos++
			}
			case tea.KeyLeft:
				if i.cursorPos > 0 {
					i.cursorPos--
				}
			case tea.KeyRunes:
				before := i.value[:i.cursorPos]
				after := i.value[i.cursorPos:]
				i.value = before + string(msg.Runes) + after
				i.cursorPos+= len(msg.Runes)
		}
	}

	return i.value, false, nil
}

func (i *Input) View() string {
	var sb strings.Builder
	sb.WriteString(i.prompt)

	// Insert Cursor
	before := i.value[:i.cursorPos]
	after := i.value[i.cursorPos:]
	sb.WriteString(before)
	sb.WriteString("█")
	sb.WriteString(after)

	return sb.String()
}