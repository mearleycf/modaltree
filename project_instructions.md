# Project Instructions and Design Decisions

## (1) Markdown Standards

This repository follows strict markdown formatting rules based on markdownlint standards:

### (1.2) Document Structure

- Files should end with a single empty line
- No multiple consecutive blank lines
- No trailing spaces at end of lines
- Use ATX-style headers (# H1, ## H2, etc.)
- No emojis in headers or section titles
- Headers must use numerical prefixes for clear hierarchy:
  - Top-level sections use single numbers: `# (1) Section Name`
  - Subsections increment per level: `## (1.1) Subsection`
  - Further nesting continues pattern: `### (1.1.1) Detail`
  - Within date blocks, add section number: `##### (1.1.1.1) Context`
  - Numbers should be wrapped in parentheses
  - A space must follow the closing parenthesis

### (1.3) Headers

- Must have a space after the # symbols
- Should be surrounded by blank lines
- Should be properly nested (no skipping levels)
- First line should be a top-level header
- Must be plain text without emojis or special characters

### (1.4) Lists

- Must be preceded by a blank line
- Must have consistent indentation
- Nested lists should be indented by 2 spaces
- Must have a space after the list marker (-, *, or number)

### (1.5) Code Blocks

- Must be surrounded by blank lines
- Must specify a language for fenced code blocks
- Must use triple backticks (```) for fencing
- Indented code blocks should use 2 spaces

### (1.6) Links and References

- No bare URLs - use proper markdown link syntax
- Internal links should use the Obsidian double-bracket format [[link]]
- External links should use [text](url) format

### (1.7) Emphasis and Styling

- Use * or _ for emphasis
- Leave spaces around emphasis markers when used in middle of text
- No multiple consecutive emphasis markers

### (1.8) Tables

- Must have header row
- Must be preceded by a blank line
- Must have proper column alignment markers
- Must have consistent column widths

## (2) Application Functionality Progress Checklist

- [x] Default: Show hidden files
- [x] Toggle hidden files with '.' key
- [ ] Add configuration file for default settings
- [ ] Fix outstanding errors in existing files

### (2.1) File Operations Interface

- [x] Move files: Input field with path and destination
- [x] Copy files: Input field with path and destination
- [ ] Add directory creation command
- [ ] Add file creation command
- [ ] Implement path autocompletion for input fields

### (2.2) Editor Integration

- [x] Default editor: VS Code ('code' command)
- [ ] Add configuration for custom editor command
- [ ] Handle cases where editor command isn't available

### (2.3) Permissions Interface

- [ ] Implement checkbox/toggle interface for permissions
- [ ] Add visual indicator for current permissions
- [ ] Add quick permission presets (e.g., executable)

### (2.4) Key Bindings

#### (2.4.1) Implemented

- [x] Navigation: arrow keys and vim-style (h,j,k,l)
- [x] Operations: e(dit), m(ove), c(opy), p/u(permissions), r(ename), d(elete)
- [x] Utility: q(uit), .(toggle hidden)

#### (2.4.2) Pending

- [ ] Add configuration for custom key bindings
- [ ] Add help overlay (press ? to view)

### (2.5) Shell Integration

- [x] Works with sh, bash, zsh, fish
- [x] Changes directory on exit
- [ ] Add support for nushell
- [ ] Consider additional shell integration features

## (3) UI Design Decisions

- Show permissions but no other metadata (size, dates)
- No preview pane for files
- Input fields appear at bottom of screen
- Confirmation prompts for destructive actions (configurable)

### (3.1) State Management

- Always open in current directory
- Keep expanded state while navigating up
- No session persistence
- No favorites system

## (4) Questions and Answers from Initial Planning

### (4.1) File Exclusions

Q: Specific file types or directories to exclude?
A: Only standard hidden files, configurable with default show/hide and toggle

### (4.2) File Editing

Q: File editing preferences?
A: Input field at bottom for renaming, external editor for content

### (4.3) UI Layout

Q: UI layout preferences?
A: Simple tree view with permissions, no preview pane
A: Support both vim-style and arrow key navigation

### (4.4) Shell Integration

Q: Shell integration scope?
A: Support all major shells, focus on directory changing

### (4.5) State Persistence

Q: State persistence?
A: No favorites or session memory, but keep expanded states during navigation

## (5) Future Considerations

- Performance optimization for large directories
- Search functionality
- Bulk operations
- File/directory creation interface

## (6) AI Instructions

* Never do unprompted changes. Only make changes I've instructed you to make. For example, if there is code we need to refactor, and I don't tell you to refactor it, don't refactor it. If there are functions/methods we haven't implemented yet, don't implement them until I ask you to.

* Never cut out previous content and list it as [Previous content remains unchanged...]. This is a bad practice. Always keep the entire content in the document and add new content where appropriate.

* MCP Server Tools
   * Don't use these tools on this project:
      * any obsidian tool
      * flashcard-server
   * Use these tools on this project, explicitly:
      * git
      * filesystem
      * github
      * mcp-knowledge-graph
      * llm-context
      * sequential-thinking
      * *aindreyway-mcp-neurolora*
      * fetch
      * *modelcontextprotocol-server-brave-search*
      * *ai-humanizer-mcp-server*
      * shell

### (2.6) View Implementation Tasks

- [ ] Implement TreeView rendering in View() method
- [ ] Implement ConfirmView rendering in View() method
- [ ] Add status bar display
- [ ] Add visual hierarchy for tree structure