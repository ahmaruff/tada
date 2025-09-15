```
__/\\\\\\\\\\\\\\\_____/\\\\\\\\\_____/\\\\\\\\\\\\________/\\\\\\\\\____        
 _\///////\\\/////____/\\\\\\\\\\\\\__\/\\\////////\\\____/\\\\\\\\\\\\\__       
  _______\/\\\________/\\\/////////\\\_\/\\\______\//\\\__/\\\/////////\\\_      
   _______\/\\\_______\/\\\_______\/\\\_\/\\\_______\/\\\_\/\\\_______\/\\\_     
    _______\/\\\_______\/\\\\\\\\\\\\\\\_\/\\\_______\/\\\_\/\\\\\\\\\\\\\\\_    
     _______\/\\\_______\/\\\/////////\\\_\/\\\_______\/\\\_\/\\\/////////\\\_   
      _______\/\\\_______\/\\\_______\/\\\_\/\\\_______/\\\__\/\\\_______\/\\\_  
       _______\/\\\_______\/\\\_______\/\\\_\/\\\\\\\\\\\\/___\/\\\_______\/\\\_ 
        _______\///________\///________\///__\////////////_____\///________\///__
```

### **_The simplest way to keep track of your tasks_**  

> _Your favorite text editor + the power of Markdown = **everything you need**_

---

## Features

- **Markdown-native**: Work with plain markdown files using any editor
- **Smart consolidation**: Automatically merge task data across different sections
- **Date-range reports**: Generate clean reports with automatic filename dating
- **Flexible workflow**: Daily cleanup or full report generation

## How it works

Tada organizes your tasks in four markdown sections:

- **Backlog**: Your task inventory with consolidated status
- **Todo**: Current work organized by date
- **Done**: Completed work by date
- **Archives**: Historical tasks for reporting

Tasks are linked by unique IDs, allowing Tada to track progress across sections and consolidate information automatically.

## Installation

Download from [releases](https://github.com/ahmaruff/tada/releases) - binaries for Windows, macOS, and Linux.

Or build from source:

```bash
git clone https://github.com/ahmaruff/tada
cd tada
go build -o tada cmd/main.go
```

## Usage

### Commands

**`tada gen [file]`** - Generate dated reports
```bash
tada gen                    # Process input.md, generate report
tada gen tasks.md           # Process specific file
tada gen -o reports/        # Save report to specific directory
tada gen --dry-run          # Preview what would be processed
```

**`tada tidy [file]`** - Clean up and organize
```bash
tada tidy                   # Consolidate task data
tada tidy --archive         # Also move completed tasks to Archives
tada tidy --dry-run         # Preview changes
```

### Workflow Examples

**Daily usage**:
```bash
# Update task status in your editor, then:
tada tidy                   # Sync Backlog with current status
```

**Weekly reports**:
```bash
tada tidy --archive         # Archive completed tasks
tada gen                    # Generate report_2025-01-15_2025-01-21.md
```

**Quick reports**:
```bash
tada gen                    # Full workflow in one command
```

## File Format

### Basic Task Structure
```markdown
## Backlog
- [ ] Task title <!-- @project|#123 -->
- [x] Completed task <!-- @project|#124|2025-01-15 -->
- [-] In progress <!-- @project|#125|2025-01-14 - 2025-01-16 -->

## Todo  
### 2025-01-16 - Tuesday
- [ ] Daily task <!-- @project|#126 -->
  Additional task description
  - [ ] Subtask 1
  - [x] Subtask 2

## Done
### 2025-01-15 - Monday  
- [x] Completed work <!-- @project|#127 -->

## Archives
```

### Task Components

**Status**: `[ ]` (todo), `[x]` (done), `[-]` (in progress)

**Comments**: `<!-- @project|#id|date-range -->`
- `@project` - Project name
- `#id` - Unique task ID (required for linking)
- `date-range` - Single date or date range

**Descriptions**: Indented text under tasks

**Subtasks**: Indented task items with status

## Generated Reports

Reports use a clean format optimized for sharing:

```markdown
# PROJECT - Task Title
2025-09-14  
Desc:  
  Task description here
  - [x] Completed subtask
  - [ ] Pending subtask

```

Report files are automatically named with date ranges: `report_2025-01-15_2025-01-21.md`

## Flags

**Global flags**:
- `-i, --input` - Input file (default: input.md)
- `-v, --verbose` - Detailed output
- `--dry-run` - Preview without changes
- `--help` - Show help

**Gen-specific**:
- `-o, --output` - Output directory for reports

**Tidy-specific**:  
- `-a, --archive` - Move completed Backlog tasks to Archives

## Examples

### Daily Workflow
1. Work on tasks, update Todo/Done sections in your editor
2. Run `tada tidy` to consolidate changes into Backlog
3. Continue working with updated task status

### Weekly Reports
1. Complete tasks throughout the week
2. Run `tada tidy -a` to move finished work to Archives
3. Run `tada gen` to create a weekly report
4. Archives are cleared, ready for next week

### Project Cleanup
```bash
tada tidy -i project.md -a --dry-run    # Preview cleanup
tada tidy -i project.md -a              # Apply cleanup
tada gen -i project.md -o reports/      # Generate report
```

## Examples

Check the `example/` folder for sample files:
- `input.md` - Starting markdown with tasks
- `input_updated.md` - After running tada commands  
- `report_2025-01-15_2025-01-16.md` - Generated report format

These show the complete workflow and expected file formats.

## Why Tada?

- **Editor agnostic**: Use vim, VS Code, Obsidian, or any markdown editor
- **Plain text**: No vendor lock-in, version control friendly
- **Flexible**: Adapts to your workflow, not the other way around
- **Fast**: CLI tool with no dependencies or setup required
- **Transparent**: Everything is visible in your markdown files

## Contributing

Contributions welcome! Please read the contributing guidelines and submit pull requests.

## License

MIT License - see [LICENSE](./LICENSE) file for details
