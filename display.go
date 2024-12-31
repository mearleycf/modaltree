package main

import "path/filepath"

// DisplayConfig holds the configuration for the display
type DisplayConfig struct {
	UseNerdFont bool // whether to use nerd font icons
	IndentSize int // number of spaces to indent
	TreeStyle string // "unicode" or "ascii"
}

// DefaultDisplayConfig returns the default display configuration
func DefaultDisplayConfig() DisplayConfig {
	return DisplayConfig{
		UseNerdFont: false,
		IndentSize: 2,
		TreeStyle: "unicode",
	}
}

// IconSet defines the icons used for different file types
type IconSet struct {
	Directory, File, Executable, Symlink, Pipe, Socket, BlockDevice, CharDevice, Special, Missing string
	DirectoryOpen string
	ParentDir string
	DefaultFile string
	FileTypeIcons map[string]string
}

// UnicodeIconSet returns the default Unicode tree icons
func UnicodeIconSet() IconSet {
	return IconSet{
		Directory:     "▸",
		DirectoryOpen: "▾",
		ParentDir:     "▴",
		DefaultFile:   "•",
		FileTypeIcons: make(map[string]string),
	}
}

// NerdFontIconSet returns Nerd Font icons
func NerdFontIconSet() IconSet {
	return IconSet{
		Directory:     "",
		DirectoryOpen: "",
		ParentDir:     "",
		DefaultFile:   "",
		FileTypeIcons: map[string]string{
			".go":     "",
			".mod":    "",
			".sum":    "",
			".js":     "",
			".jsx":    "",
			".ts":     "",
			".tsx":    "",
			".py":     "",
			".rb":     "",
			".html":   "",
			".css":    "",
			".scss":   "",
			".json":   "",
			".yml":    "",
			".yaml":   "",
			".toml":   "",
			".md":     "",
			".txt":    "",
			".sh":     "",
			".bash":   "",
			".zsh":    "",
			".fish":   "",
			".git":    "",
			".gitignore": "",
			".env":    "",
			".lock":   "",
			".zip":    "",
			".tar":    "",
			".gz":     "",
			".pdf":    "",
			".doc":    "",
			".docx":   "",
			".xls":    "",
			".xlsx":   "",
			".ppt":    "",
			".pptx":   "",
			".jpg":    "",
			".jpeg":   "",
			".png":    "",
			".gif":    "",
			".svg":    "",
			".mp3":    "",
			".mp4":    "",
			".wav":    "",
			".mov":    "",
		},
	}
}

// GetFileIcon returns the appropriate icon for a file
func (is IconSet) GetFileIcon(item FileItem) string {
	if item.isDir {
		if item.name == ".." {
			return is.ParentDir
		}
		return is.Directory
	}

	// get file extension
	ext := filepath.Ext(item.name)
	if icon, ok := is.FileTypeIcons[ext]; ok {
		return icon
	}

	return is.DefaultFile
}

// TreeSymbols contains the symbols used to draw the tree
type TreeSymbols struct {
	Vertical, Corner, Tee, Horizontal string
}

// UnicodeTreeSymbols returns the default Unicode tree symbols
func UnicodeTreeSymbols() TreeSymbols {
	return TreeSymbols{
		Vertical:   "│",
		Corner:     "└",
		Tee:        "├",
		Horizontal: "─",
	}
}

// AsciiTreeSymbols returns ASCII tree symbols
func AsciiTreeSymbols() TreeSymbols {
	return TreeSymbols{
		Vertical:   "|",
		Corner:     "`",
		Tee:        "|",
		Horizontal: "-",
	}
}