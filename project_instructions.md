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
- Do not use * for unordered lists, markdownlint doesn't like it; use - instead

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

## (2) UI Design Decisions

- Show permissions but no other metadata (size, dates)
- No preview pane for files
- Input fields appear at bottom of screen
- Confirmation prompts for destructive actions (configurable)

### (2.1) State Management

- Always open in current directory
- Keep expanded state while navigating up
- No session persistence
- No favorites system

## (3) Questions and Answers from Initial Planning

### (3.1) File Exclusions

Q: Specific file types or directories to exclude?
A: Only standard hidden files, configurable with default show/hide and toggle

### (3.2) File Editing

Q: File editing preferences?
A: Input field at bottom for renaming, external editor for content

### (3.3) UI Layout

Q: UI layout preferences?
A: Simple tree view with permissions, no preview pane
A: Support both vim-style and arrow key navigation

### (3.4) Shell Integration

Q: Shell integration scope?
A: Support all major shells, focus on directory changing

### (3.5) State Persistence

Q: State persistence?
A: No favorites or session memory, but keep expanded states during navigation

## (4) Future Considerations

- Performance optimization for large directories
- Search functionality
- Bulk operations
- File/directory creation interface

## (5) AI Instructions

- Never do unprompted changes. Only make changes I've instructed you to make. For example, if there is code we need to refactor, and I don't tell you to refactor it, don't refactor it. If there are functions/methods we haven't implemented yet, don't implement them until I ask you to.

- CRITICAL: Never use placeholders like "[previous content...]" or similar. Always include the complete content of any file being modified. This is a strict requirement that helps prevent errors and maintain clarity.

- Never cut out previous content and list it as [Previous content remains unchanged...]. This is a bad practice. Always keep the entire content in the document and add new content where appropriate.

- Never make git commits unless explicitly instructed to do so.

- MCP Server Tools
  - Don't use these tools on this project:
    - any obsidian tool
    - flashcard-server
  - Use these tools on this project, explicitly:
    - git
    - filesystem
    - github
    - mcp-knowledge-graph
    - llm-context
    - sequential-thinking
    - *aindreyway-mcp-neurolora*
    - fetch
    - *modelcontextprotocol-server-brave-search*
    - *ai-humanizer-mcp-server*
    - shell