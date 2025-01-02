package main

import (
	"path/filepath"

	"github.com/mattn/go-runewidth"
)

// DisplayConfig holds the configuration for the display
type DisplayConfig struct {
	UseNerdFont bool // whether to use nerd font icons
	IndentSize int // number of spaces to indent
	TreeStyle string // "unicode" or "ascii"
	fontVerified bool // internal state for font verification
}

// DefaultDisplayConfig returns the default display configuration
func DefaultDisplayConfig() DisplayConfig {
	return DisplayConfig{
		UseNerdFont: false,
		IndentSize: 2,
		TreeStyle: "unicode",
		fontVerified: false,
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
		Directory:     "\uf74a",     // Folder icon
		DirectoryOpen: "\uf74b",     // Open folder icon
		ParentDir:     "\uf743",     // Parent directory icon  
		DefaultFile:   "\uf723",     // Default file icon
		FileTypeIcons: map[string]string{
			".go":     "\ue724", // Go icon
			".mod":    "\ue624", // Go module icon
			".sum":    "\ue65e", // Go sum icon
			".js":     "\ue781", // JavaScript icon
			".jsx":    "\ue625", // JSX icon
			".ts":     "\ue628", // TypeScript icon
			".tsx":    "\ue625", // TSX icon
			".py":     "\ue235", // Python icon
			".rb":     "\ue605", // Ruby icon
			".html":   "\ue736", // HTML icon
			".css":    "\ue749", // CSS icon
			".scss":   "\ue603", // SCSS icon
			".json":   "\ue60b", // JSON icon
			".yml":    "\uf0f6", // YAML icon
			".yaml":   "\uf0f6", // YAML icon
			".toml":   "\ue6b2", // TOML icon
			".md":     "\ue73e", // Markdown icon
			".txt":    "\uf0f6", // Text icon
			".sh":     "\ue691", // Shell script icon
			".bash":   "\ue760", // Bash icon
			".zsh":    "\ue691", // Zsh icon
			".fish":   "\uea85", // Fish icon
			".git":    "\ue702", // Git icon
			".gitignore": "\ue65d", // Git ignore icon
			".env":    "\ueb52", // Env icon
			".lock":   "\uf023", // Lock icon
			".zip":    "\uf292", // Zip icon
			".tar":    "\ue6aa", // Tar icon
			".gz":     "\ue6aa", // Gzip icon
			".pdf":    "\uf724", // PDF icon
			".doc":    "\ue6a5", // Word icon
			".docx":   "\ue6a5", // Word icon
			".xls":    "\uf1c3", // Excel icon
			".xlsx":   "\uf1c3", // Excel icon
			".ppt":    "\uf1c4", // PowerPoint icon
			".pptx":   "\uf1c4", // PowerPoint icon
			".jpg":    "\uf03e", // JPEG icon
			".jpeg":   "\uf03e", // JPEG icon
			".png":    "\uf03e", // PNG icon
			".gif":    "\uf1c5", // GIF icon
			".svg":    "\uf1c5", // SVG icon
			".mp3":    "\uf910", // MP3 icon
			".mp4":    "\uf72f", // MP4 icon
			".wav":    "\ued81", // WAV icon
			".mov":    "\uf1c8", // MOV icon
		},
	}
}
// GetFileIcon returns the appropriate icon for a file
func (is IconSet) GetFileIcon(item FileItem, config DisplayConfig) string {
	// first check if nerd fonts are enabled and verified
	if !config.UseNerdFont || !config.fontVerified {
		// fall back to unicode icons if nerd fonts are not enabled or verified
		if item.isDir {
			if item.name == ".." {
				return "▴" // unicode fallback for parent directory
			}
			return "▸" // unicode fallback for directory
		}
		return "•" // unicode fallback for file
	}
	// use nerd fonts when verified
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

// VerifyNerdFont checks if nerd fonts are available in the terminal
func (dc *DisplayConfig) VerifyNerdFont() bool {
	if !dc.UseNerdFont {
		return false
	}

	// Test character that exists only in nerd fonts
	testChar := "\uf74a"
	
	// Use tcell to check if the character can be displayed
	width := runewidth.StringWidth(testChar)
	
	// If width is 0 or greater than 1, the font likely isn't properly supported
	dc.fontVerified = width == 1
	
	return dc.fontVerified
}

var DefaultIconSet = UnicodeIconSet()
