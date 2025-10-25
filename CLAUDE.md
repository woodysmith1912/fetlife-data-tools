# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go CLI tool that syncs FetLife data exports into an Obsidian vault. It reads CSV files containing blocked users and private notes, then creates or updates markdown files in the Obsidian vault with appropriate metadata and organization.

## Core Architecture

### Three-Layer Structure

1. **CLI Layer** (`program/` package):
   - Uses Kong for command parsing
   - Command hierarchy: `obsidian sync` and `obsidian list`
   - Handles logging setup (zerolog with console/JSON output)
   - Global options: `--vault`, `--debug`, `--quiet`, `--output-format`

2. **Obsidian Layer** (`obsidian/` package):
   - `Vault` type: Represents an Obsidian vault and its pages
   - `Page` type: Represents a markdown file with YAML frontmatter
   - Key metadata fields: `tags`, `url`, `url-aliases`, `web-message`, `web-badge-color`
   - `Load()`: Walks directory tree and parses all `.md` files
   - `Save()`: Writes page back with updated frontmatter

3. **Sync Logic** (`program/sync.go`):
   - Reads CSV files: `blockeds.txt` and `private_notes.txt`
   - Creates/updates pages for users based on their user ID
   - Finds existing pages by matching URLs or URL aliases

### Key Sync Behavior

**Folder Placement Logic:**

The sync command uses two mechanisms to determine where to create pages:

1. **Blocked Users** (`CreateBlockedIn` flag):
   - Default folder: "Bad People"
   - All blocked users go here regardless of other settings
   - Set via `--create-blocked-in` flag

2. **Private Notes** (`CreatePeopleIn` flag):
   - Keyword-based folder routing via syntax: `folder[:keyword1,keyword2,...]`
   - Example: `--in "People" --in "Bad People:creepy,stalker" --in "Friends:friend,cool"`
   - Case-insensitive keyword matching in private note content
   - First matching keyword determines folder, otherwise uses first folder as default

**User Identification:**
- Uses FetLife user ID from URLs (e.g., `/users/12345`)
- Checks both `url` field and `url-aliases` array in frontmatter
- Skips if multiple pages match same user ID

**Data Import Flow:**
1. Load vault pages into memory
2. Read `blockeds.txt` CSV (columns: user_id, created_at, updated_at, nickname)
3. Read `private_notes.txt` CSV (columns: member_id, created_at, updated_at, private_note)
4. For each blocked user: create/update page, add "blocked" tag, set folder per `CreateBlockedIn`
5. For each private note: create/update page, set `web-message`, determine folder via keyword matching

## Development Commands

### Building and Testing

```bash
# Build the binary
go build

# Run all tests
go test ./...

# Run specific test
go test ./program -run TestSyncCmd_Integration_KeywordMatching -v

# Run tests for specific package
go test ./obsidian -v
go test ./program -v
```

### Running the Tool

```bash
# Sync with default settings
./fetlife-data-tools obsidian sync --data-dir ./example/test-data

# Sync with custom vault path
./fetlife-data-tools --vault ./example/vault obsidian sync --data-dir ./example/test-data

# Sync with keyword-based folder routing
./fetlife-data-tools obsidian sync \
  --data-dir ./example/test-data \
  --in "People" \
  --in "Bad People:creepy,stalker,harassment" \
  --in "Friends:friend,cool"

# Sync with custom blocked folder
./fetlife-data-tools obsidian sync \
  --data-dir ./example/test-data \
  --create-blocked-in "Blocked Users"

# List people in vault
./fetlife-data-tools --vault ./example/vault obsidian list

# Enable debug logging
./fetlife-data-tools --debug obsidian sync --data-dir ./example/test-data
```

## Testing Patterns

### Test Data Structure

- `example/vault/`: Sample Obsidian vault with People, Bad People folders
- `example/test-data/`: CSV files for testing sync functionality
- `example/private-data/`: Contains actual exported FetLife data (not in tests)

### Writing Sync Tests

When testing sync functionality:
1. Create temp vault and data directories using `t.TempDir()`
2. Set up template file at `Templates/People.md`
3. Create CSV files with appropriate headers
4. Create `SyncCmd` with desired configuration
5. Verify files created in correct folders
6. Check that tags, URLs, and messages are set correctly

Example test structure:
```go
func TestSyncCmd_Something(t *testing.T) {
    tempVault := t.TempDir()
    testDataDir := t.TempDir()

    // Set up templates
    templatesDir := filepath.Join(tempVault, "Templates")
    os.MkdirAll(templatesDir, 0755)

    // Create test CSV files
    // ...

    sync := &SyncCmd{
        DataDir:         testDataDir,
        CreatePeopleIn:  []string{"People", "Bad People:creepy"},
        CreateBlockedIn: "Bad People",
    }

    err := sync.Run(&Options{Vault: tempVault})
    assert.NoError(t, err)

    // Verify files exist in expected locations
    // ...
}
```

## Key Implementation Details

### Page Creation Functions

Three related functions handle page creation:

1. `createPageInFolder(vault, userID, nickname, folder, options)` - Base function that creates page in specific folder
2. `createPageFromTemplateWithNote(vault, userID, nickname, privateNote, options)` - Determines folder via keyword matching, then delegates to `createPageInFolder`
3. `createPageFromTemplate(vault, userID, nickname, options)` - Wrapper for blocked users without notes

### Template System

- Template location: `<vault>/Templates/People.md`
- Placeholder `{{title}}` replaced with page name
- URL placeholder auto-completed with user ID
- Falls back to default template if file doesn't exist

### Folder Configuration Parsing

The `parseFolderConfig()` function splits `"Folder:keyword1,keyword2"` into:
- Folder name (before colon)
- Keywords array (after colon, comma-separated, trimmed, lowercased)

Example: `"Bad People:creepy,stalker"` → folder="Bad People", keywords=["creepy", "stalker"]
