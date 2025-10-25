# FetLife Data Tools

A command-line tool for working with FetLife data exports. 

Sync blocked users and private notes into an Obsidian vault with automatic organization, 
or generate CSV/Excel spreadsheets for analysis and backup.

## Quick Start

### 1. Export Your FetLife Data

1. Go to FetLife and export your data
2. Extract the exported archive
3. Locate the `blockeds.txt` and `private_notes.txt` CSV files

### 2. Choose Your Workflow

**Option A: Sync to Obsidian** (recommended for knowledge management)
- Creates markdown files in your Obsidian vault
- Supports keyword-based organization
- Integrates with your existing notes and links

**Option B: Generate Spreadsheet** (recommended for analysis/sharing)
- Creates CSV or Excel files
- Easy to sort, filter, and analyze
- Portable format for sharing or backup

### 3. Run the Tool

#### For Obsidian Sync

```bash
# Basic usage - sync to current directory
./fetlife-data-tools obsidian sync --data-dir /path/to/fetlife/export

# Specify a different vault location
./fetlife-data-tools --vault /path/to/vault obsidian sync --data-dir /path/to/fetlife/export
```

**Result:** Markdown files created in your vault:
- Blocked users → `Bad People/` folder (by default)
- Users with private notes → `People/` folder (by default, or routed by keywords)

#### For Spreadsheet Generation

```bash
# Generate CSV file
./fetlife-data-tools spreadsheet generate --data-dir /path/to/fetlife/export

# Generate Excel file
./fetlife-data-tools spreadsheet generate --data-dir /path/to/fetlife/export --format xlsx
```

**Result:** A spreadsheet file (`fetlife-export.csv` or `fetlife-export.xlsx`) in the current directory containing all your blocked users and private notes in a tabular format.

## Installation

### From Source

```bash
git clone https://github.com/woodysmith1912/fetlife-data-tools
cd fetlife-data-tools
go build
```

### Download Binary

Download the latest release from the [releases page](https://github.com/woodysmith1912/fetlife-data-tools/releases).

## Usage

### Commands

```bash
# Sync FetLife data to Obsidian vault
fetlife-data-tools obsidian sync --data-dir <path>

# List people in vault
fetlife-data-tools obsidian list

# Generate spreadsheet from FetLife data
fetlife-data-tools spreadsheet generate --data-dir <path>

# Show version
fetlife-data-tools version
```

### Sync Options

#### Required Flags

- `--data-dir` - Path to directory containing `blockeds.txt` and `private_notes.txt`

#### Optional Flags

- `--vault` - Path to Obsidian vault (default: current directory, env: `VAULT_PATH`)
- `--create-people-in` - Folders for creating people with keyword routing (default: `People`)
- `--create-blocked-in` - Folder for blocked users (default: `Bad People`)
- `--debug` - Enable debug logging
- `--quiet` - Reduce log verbosity
- `--output-format` - Output format: `auto`, `terminal`, or `jsonl`

### Spreadsheet Generation

Generate CSV or Excel spreadsheets from your FetLife data exports without syncing to an Obsidian vault.

#### Basic Usage

```bash
# Generate CSV file (default)
./fetlife-data-tools spreadsheet generate --data-dir /path/to/fetlife/export

# Generate Excel file
./fetlife-data-tools spreadsheet generate --data-dir /path/to/fetlife/export --format xlsx

# Generate both CSV and Excel
./fetlife-data-tools spreadsheet generate --data-dir /path/to/fetlife/export --format both
```

#### Options

- `--data-dir` - (Required) Path to directory containing `blockeds.txt` and `private_notes.txt`
- `--output-dir` - Directory for generated files (default: current directory)
- `--basename` - Base name for output files without extension (default: `fetlife-export`)
- `--format` - Output format: `csv`, `xlsx`, or `both` (default: `csv`)

#### Examples

```bash
# Generate Excel file with custom name
./fetlife-data-tools spreadsheet generate \
  --data-dir ~/Downloads/fetlife-export \
  --format xlsx \
  --basename my-fetlife-data

# Generate both formats in specific directory
./fetlife-data-tools spreadsheet generate \
  --data-dir ~/Downloads/fetlife-export \
  --output-dir ~/Documents/Spreadsheets \
  --format both \
  --basename fetlife-2025
```

#### Output Format

The generated spreadsheets include the following columns:

- **User ID** - FetLife user ID
- **Nickname** - User's nickname (from blocked list)
- **URL** - Direct link to user's profile
- **Blocked** - Whether the user is blocked (Yes/No)
- **Blocked At** - When the user was blocked
- **Private Note** - Your private note about the user
- **Note Created** - When the note was created
- **Note Updated** - When the note was last updated

The data combines both blocked users and private notes, showing all information for each user in a single row.

### Advanced Usage

#### Keyword-Based Folder Routing

Automatically organize people into different folders based on keywords in their private notes:

```bash
./fetlife-data-tools obsidian sync \
  --data-dir /path/to/data \
  --create-people-in "People" \
  --create-people-in "Bad People:creepy,stalker,harassment" \
  --create-people-in "Friends:friend,cool,awesome"
```

**How it works:**
- Private notes are scanned for keywords (case-insensitive)
- First matching keyword determines the folder
- If no keywords match, uses the first folder as default
- Syntax: `folder_name:keyword1,keyword2,keyword3`

**Example:**
```bash
# Private note: "This person is really creepy and sent harassing messages"
# → Goes to "Bad People" (matches "creepy" and "harassment")

# Private note: "My friend from the workshop, really cool person"
# → Goes to "Friends" (matches "friend" and "cool")

# Private note: "Met at coffee shop, interesting conversation"
# → Goes to "People" (no keywords matched, uses default)
```

#### Custom Blocked User Folder

```bash
./fetlife-data-tools obsidian sync \
  --data-dir /path/to/data \
  --create-blocked-in "Blocked Users"
```

## How It Works

### Data Processing

1. **Load Vault** - Scans your Obsidian vault for existing markdown files
2. **Read Data** - Parses `blockeds.txt` and `private_notes.txt` CSV files
3. **Match Users** - Identifies existing pages by matching FetLife user IDs in URLs
4. **Create/Update Pages** - Creates new pages or updates existing ones with:
   - Proper YAML frontmatter
   - FetLife user URL
   - Tags (`blocked` tag for blocked users)
   - Private notes (in `web-message` field)

### Page Creation

When creating new pages, the tool:
1. Uses your `Templates/People.md` template if it exists
2. Falls back to a default template if no template is found
3. Replaces `{{title}}` placeholder with the user's nickname or `user-<id>`
4. Sets the FetLife URL: `https://fetlife.com/users/<id>`
5. Places the page in the appropriate folder based on rules

### Template Format

Create a template at `<vault>/Templates/People.md`:

```yaml
---
aliases:
url: https://fetlife.com/users/
tags:
url-aliases:
  - https://fetlife.com/{{title}}
---
# Relationships
# Notes
# Log
```

The tool will:
- Replace `{{title}}` with the person's name
- Fill in the user ID in the `url` field
- Preserve any other frontmatter fields

## Data Files

### blockeds.txt

CSV format with headers:

```csv
blocked_user_id,created_at,updated_at,blocked_nickname
12345,2024-01-15 10:30:00 UTC,2024-01-15 10:30:00 UTC,UserName
```

### private_notes.txt

CSV format with headers:

```csv
member_id,created_at,updated_at,private_note
12345,2024-01-15 10:30:00 UTC,2024-01-15 10:30:00 UTC,Note text here
```

## Page Metadata

Created pages include YAML frontmatter:

```yaml
---
tags:
  - person
  - blocked  # Only for blocked users
url: https://fetlife.com/users/12345
url-aliases:
  - https://fetlife.com/UserName
web-message: Private note content here
---
```

## Examples

### Basic Sync

```bash
# Sync FetLife data to vault in current directory
./fetlife-data-tools obsidian sync --data-dir ~/Downloads/fetlife-export
```

### Sync to Specific Vault

```bash
# Sync to a vault in a different location
./fetlife-data-tools --vault ~/Documents/MyVault obsidian sync \
  --data-dir ~/Downloads/fetlife-export
```

### Advanced Organization

```bash
# Organize people into multiple folders with keyword routing
./fetlife-data-tools --vault ~/Documents/MyVault obsidian sync \
  --data-dir ~/Downloads/fetlife-export \
  --create-people-in "People" \
  --create-people-in "Potential Friends:friend,cool,nice,awesome" \
  --create-people-in "Red Flags:creepy,weird,uncomfortable" \
  --create-people-in "Events:met at,event,party" \
  --create-blocked-in "Blocked & Reported"
```

### Debug Mode

```bash
# Enable debug logging to see detailed processing
./fetlife-data-tools --debug obsidian sync --data-dir ~/Downloads/fetlife-export
```

### List People in Vault

```bash
# List all people with FetLife URLs
./fetlife-data-tools obsidian list

# List people from specific vault
./fetlife-data-tools --vault ~/Documents/MyVault obsidian list
```

### Generate Spreadsheets

```bash
# Quick CSV export
./fetlife-data-tools spreadsheet generate --data-dir ~/Downloads/fetlife-export

# Generate Excel for better formatting
./fetlife-data-tools spreadsheet generate \
  --data-dir ~/Downloads/fetlife-export \
  --format xlsx \
  --basename my-fetlife-backup

# Generate both formats for archiving
./fetlife-data-tools spreadsheet generate \
  --data-dir ~/Downloads/fetlife-export \
  --output-dir ~/Backups/FetLife \
  --format both \
  --basename fetlife-backup-2025-01
```

## Project Structure

```
fetlife-data-tools/
├── fetlife/          # FetLife data parsing (CSV readers)
├── obsidian/         # Obsidian vault and page management
├── program/          # CLI commands and sync logic
├── example/
│   ├── vault/        # Example vault structure
│   └── test-data/    # Example CSV files for testing
└── main.go           # Application entry point
```

## Development

### Building

```bash
go build
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./fetlife -v
go test ./obsidian -v
go test ./program -v
```

### Architecture

The project follows a three-layer architecture:

1. **CLI Layer** (`program/`) - Command parsing and option handling using Kong
2. **Obsidian Layer** (`obsidian/`) - Vault and page management with YAML frontmatter
3. **FetLife Layer** (`fetlife/`) - CSV parsing for blocked users and private notes

See [CLAUDE.md](CLAUDE.md) for detailed developer documentation.

## Troubleshooting

### Template Not Found Warning

```
WRN Template not found, using default error="open .../Templates/People.md: no such file or directory"
```

**Solution:** The tool will use a default template. To use a custom template:
1. Ensure you're passing the correct vault path with `--vault`
2. Create `Templates/People.md` in your vault
3. Use the template format shown above

### Multiple Pages Match Same User

```
WRN Multiple pages found for user ID, skipping userID=12345 matchCount=2
```

**Solution:** The tool found multiple pages with the same FetLife user ID. Manually consolidate the duplicate pages.

### No Files Created

**Checklist:**
- Verify `blockeds.txt` and `private_notes.txt` exist in `--data-dir`
- Check CSV files have proper headers
- Ensure vault path is correct (use `--vault` flag or `VAULT_PATH` env var)
- Run with `--debug` to see detailed processing

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please open an issue or pull request on GitHub.
