# (1) ModalTree

A modal interactive directory tree TUI (Terminal User Interface) built with Go and the Charm libraries. ModalTree allows you to navigate and manipulate your file system directly from the terminal with an intuitive interface.

## (1.1) Features

- File system navigation with expandable directory tree
- Hidden file toggling
- File operations:
  - Move files/directories
  - Copy files/directories
  - Delete files/directories
  - Rename files/directories
- Integrated editor launching
- File permissions management
- Shell directory change integration

## (1.2) Installation

### (1.2.1) Prerequisites

- Go 1.16 or later
- Git

### (1.2.2) Building from source

```bash
# Clone the repository
git clone https://github.com/mearleycf/modaltree.git
cd modaltree

# Build the project
go build

# Install globally (optional)
go install
```

## (1.3) Usage

### (1.3.1) Starting the application

```bash
# Run from current directory
modaltree

# Run from specific directory
modaltree /path/to/directory
```

### (1.3.2) Keyboard Controls

Navigation:
- `↑` or `k`: Move cursor up
- `↓` or `j`: Move cursor down
- `←` or `h`: Go to parent directory/collapse directory
- `→` or `l`: Expand directory
- `Enter`: Open directory/Expand directory

File Operations:
- `e`: Open in editor (default: VS Code)
- `m`: Move file/directory
- `c`: Copy file/directory
- `u`/`p`: Change permissions
- `r`: Rename file/directory
- `d`: Delete file/directory

Other Controls:
- `.`: Toggle hidden files
- `q` or `Ctrl+C`: Quit application

### (1.3.3) Configuration

By default, ModalTree uses these settings:
- Shows hidden files (toggle with '.')
- Uses 'code' (VS Code) as the default editor
- Confirms destructive actions
- Opens in the current working directory

## (1.4) Shell Integration

ModalTree can integrate with your shell to allow changing directories when you exit. Add this to your shell configuration:

For bash (~/.bashrc):

```bash
function mt() {
    modaltree "$@"
    if [ -f /tmp/modaltree_lastdir ]; then
        cd "$(cat /tmp/modaltree_lastdir)"
        rm /tmp/modaltree_lastdir
    fi
}
```

For zsh (~/.zshrc):

```zsh
function mt() {
    modaltree "$@"
    if [ -f /tmp/modaltree_lastdir ]; then
        cd "$(cat /tmp/modaltree_lastdir)"
        rm /tmp/modaltree_lastdir
    fi
}
```

For fish (~/.config/fish/functions/mt.fish):

```fish
function mt
    modaltree $argv
    if test -f /tmp/modaltree_lastdir
        cd (cat /tmp/modaltree_lastdir)
        rm /tmp/modaltree_lastdir
    end
end
```

## (1.5) Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## (1.6) License

[MIT License](LICENSE)