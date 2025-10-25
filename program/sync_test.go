package program

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woodysmith1912/fetlife-data-tools/obsidian"
)

func TestParseFolderConfig(t *testing.T) {
	tests := []struct {
		name             string
		config           string
		expectedFolder   string
		expectedKeywords []string
	}{
		{
			name:             "folder without keywords",
			config:           "People",
			expectedFolder:   "People",
			expectedKeywords: nil,
		},
		{
			name:             "folder with single keyword",
			config:           "Bad People:creepy",
			expectedFolder:   "Bad People",
			expectedKeywords: []string{"creepy"},
		},
		{
			name:             "folder with multiple keywords",
			config:           "Bad People:creepy,stalker,harassment",
			expectedFolder:   "Bad People",
			expectedKeywords: []string{"creepy", "stalker", "harassment"},
		},
		{
			name:             "folder with keywords with spaces",
			config:           "Bad People: creepy , stalker , harassment ",
			expectedFolder:   "Bad People",
			expectedKeywords: []string{"creepy", "stalker", "harassment"},
		},
		{
			name:             "folder with empty keyword list",
			config:           "People:",
			expectedFolder:   "People",
			expectedKeywords: nil,
		},
		{
			name:             "folder with mixed case keywords (should be lowercased)",
			config:           "Bad People:Creepy,STALKER,HaRaSsMeNt",
			expectedFolder:   "Bad People",
			expectedKeywords: []string{"creepy", "stalker", "harassment"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder, keywords := parseFolderConfig(tt.config)
			assert.Equal(t, tt.expectedFolder, folder)
			assert.Equal(t, tt.expectedKeywords, keywords)
		})
	}
}

func TestDetermineFolderForUser(t *testing.T) {
	tests := []struct {
		name           string
		createPeopleIn []string
		userID         string
		privateNote    string
		expectedFolder string
	}{
		{
			name:           "empty CreatePeopleIn defaults to People",
			createPeopleIn: []string{},
			userID:         "12345",
			privateNote:    "",
			expectedFolder: "People",
		},
		{
			name:           "single folder without keywords",
			createPeopleIn: []string{"People"},
			userID:         "12345",
			privateNote:    "",
			expectedFolder: "People",
		},
		{
			name:           "match first keyword in bad people",
			createPeopleIn: []string{"People", "Bad People:creepy,stalker"},
			userID:         "12345",
			privateNote:    "This person is really creepy",
			expectedFolder: "Bad People",
		},
		{
			name:           "match second keyword in bad people",
			createPeopleIn: []string{"People", "Bad People:creepy,stalker"},
			userID:         "12345",
			privateNote:    "Stalker behavior observed",
			expectedFolder: "Bad People",
		},
		{
			name:           "no match defaults to first folder",
			createPeopleIn: []string{"People", "Bad People:creepy,stalker"},
			userID:         "12345",
			privateNote:    "Nice person",
			expectedFolder: "People",
		},
		{
			name:           "case insensitive keyword matching",
			createPeopleIn: []string{"People", "Bad People:creepy,stalker"},
			userID:         "12345",
			privateNote:    "CREEPY person with STALKER tendencies",
			expectedFolder: "Bad People",
		},
		{
			name:           "keyword in middle of note",
			createPeopleIn: []string{"People", "Bad People:harassment"},
			userID:         "12345",
			privateNote:    "Reported for harassment multiple times",
			expectedFolder: "Bad People",
		},
		{
			name:           "multiple folders with keywords - match second",
			createPeopleIn: []string{"Friends:friend,cool", "Bad People:creepy,stalker", "Blocked:blocked"},
			userID:         "12345",
			privateNote:    "This stalker keeps messaging",
			expectedFolder: "Bad People",
		},
		{
			name:           "multiple folders with keywords - match first",
			createPeopleIn: []string{"Friends:friend,cool", "Bad People:creepy,stalker"},
			userID:         "12345",
			privateNote:    "Really cool person",
			expectedFolder: "Friends",
		},
		{
			name:           "empty note with keywords configured",
			createPeopleIn: []string{"People", "Bad People:creepy"},
			userID:         "12345",
			privateNote:    "",
			expectedFolder: "People",
		},
		{
			name:           "first folder has keywords but doesn't match",
			createPeopleIn: []string{"Friends:friend", "People"},
			userID:         "12345",
			privateNote:    "Someone I met",
			expectedFolder: "Friends",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sync := &SyncCmd{
				CreatePeopleIn: tt.createPeopleIn,
			}
			folder := sync.determineFolderForUser(tt.userID, tt.privateNote)
			assert.Equal(t, tt.expectedFolder, folder)
		})
	}
}

func TestCreatePageFromTemplateWithNote(t *testing.T) {
	// Create a temporary vault
	tempVault := t.TempDir()

	// Create Templates directory with People.md template
	templatesDir := filepath.Join(tempVault, "Templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateContent := `---
tags:
  - person
url: https://fetlife.com/users/
---

# Notes
`
	templatePath := filepath.Join(templatesDir, "People.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	tests := []struct {
		name           string
		createPeopleIn []string
		userID         string
		nickname       string
		privateNote    string
		expectedFolder string
		expectedName   string
	}{
		{
			name:           "create in default People folder",
			createPeopleIn: []string{"People"},
			userID:         "12345",
			nickname:       "TestUser",
			privateNote:    "",
			expectedFolder: "People",
			expectedName:   "TestUser",
		},
		{
			name:           "create in Bad People folder with keyword match",
			createPeopleIn: []string{"People", "Bad People:creepy"},
			userID:         "67890",
			nickname:       "CreepyUser",
			privateNote:    "This person is creepy",
			expectedFolder: "Bad People",
			expectedName:   "CreepyUser",
		},
		{
			name:           "create with user ID as name when nickname empty",
			createPeopleIn: []string{"People"},
			userID:         "99999",
			nickname:       "",
			privateNote:    "",
			expectedFolder: "People",
			expectedName:   "user-99999",
		},
		{
			name:           "create in Friends folder with keyword match",
			createPeopleIn: []string{"People", "Friends:friend,cool"},
			userID:         "11111",
			nickname:       "CoolFriend",
			privateNote:    "Really cool person",
			expectedFolder: "Friends",
			expectedName:   "CoolFriend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh vault for each test
			vault := obsidian.NewVault(tempVault)
			err := vault.Load()
			assert.NoError(t, err)

			sync := &SyncCmd{
				CreatePeopleIn: tt.createPeopleIn,
			}

			options := &Options{
				Vault: tempVault,
			}

			page, err := sync.createPageFromTemplateWithNote(vault, tt.userID, tt.nickname, tt.privateNote, options)
			assert.NoError(t, err)
			assert.NotNil(t, page)

			// Verify page properties
			assert.Equal(t, tt.expectedName, page.Title)
			assert.Contains(t, page.Url, tt.userID)

			// Verify file was created in the correct folder
			expectedPath := filepath.Join(tempVault, tt.expectedFolder, tt.expectedName+".md")
			_, err = os.Stat(expectedPath)
			assert.NoError(t, err, "Page file should exist at %s", expectedPath)

			// Clean up the created file
			os.Remove(expectedPath)
		})
	}
}

func TestCreatePageFromTemplate(t *testing.T) {
	// Create a temporary vault
	tempVault := t.TempDir()

	// Create Templates directory with People.md template
	templatesDir := filepath.Join(tempVault, "Templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateContent := `---
tags:
  - person
url: https://fetlife.com/users/
---

# Notes
`
	templatePath := filepath.Join(templatesDir, "People.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	vault := obsidian.NewVault(tempVault)
	err := vault.Load()
	assert.NoError(t, err)

	sync := &SyncCmd{
		CreatePeopleIn: []string{"People"},
	}

	options := &Options{
		Vault: tempVault,
	}

	// Test creating a page with nickname
	page, err := sync.createPageFromTemplate(vault, "12345", "TestUser", options)
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, "TestUser", page.Title)

	// Verify file exists
	expectedPath := filepath.Join(tempVault, "People", "TestUser.md")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)
}

func TestCreatePageFromTemplateWithNote_NoTemplate(t *testing.T) {
	// Create a temporary vault without template
	tempVault := t.TempDir()

	vault := obsidian.NewVault(tempVault)
	err := vault.Load()
	assert.NoError(t, err)

	sync := &SyncCmd{
		CreatePeopleIn: []string{"People"},
	}

	options := &Options{
		Vault: tempVault,
	}

	// Should still work with default template
	page, err := sync.createPageFromTemplateWithNote(vault, "12345", "TestUser", "", options)
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Equal(t, "TestUser", page.Title)

	// Verify file exists
	expectedPath := filepath.Join(tempVault, "People", "TestUser.md")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)

	// Verify default template was used (should contain basic frontmatter)
	content, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "tags:")
	assert.Contains(t, string(content), "person")
	assert.Contains(t, string(content), "https://fetlife.com/users/12345")
}

func TestSyncCmd_Integration_KeywordMatching(t *testing.T) {
	// Create a temporary vault
	tempVault := t.TempDir()

	// Create Templates directory
	templatesDir := filepath.Join(tempVault, "Templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateContent := `---
tags:
  - person
url: https://fetlife.com/users/
---

# Notes
`
	templatePath := filepath.Join(templatesDir, "People.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Create test data directory
	testDataDir := t.TempDir()

	// Create private_notes.txt
	privateNotesContent := `member_id,created_at,updated_at,private_note
11111,2024-01-01,2024-01-01,Great photographer! Very professional.
22222,2024-01-01,2024-01-01,This person is creepy and sent harassing messages
33333,2024-01-01,2024-01-01,Stalker behavior - blocked them
44444,2024-01-01,2024-01-01,Cool person met at event
55555,2024-01-01,2024-01-01,My friend from the workshop
`
	privateNotesPath := filepath.Join(testDataDir, "private_notes.txt")
	if err := os.WriteFile(privateNotesPath, []byte(privateNotesContent), 0644); err != nil {
		t.Fatalf("Failed to create private_notes.txt: %v", err)
	}

	// Create empty blockeds.txt
	blockedsContent := `user_id,created_at,updated_at,nickname
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	if err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644); err != nil {
		t.Fatalf("Failed to create blockeds.txt: %v", err)
	}

	// Create sync command directly
	sync := &SyncCmd{
		DataDir:        testDataDir,
		CreatePeopleIn: []string{"People", "Bad People:creepy,stalker,harassing", "Friends:cool,friend"},
	}

	options := &Options{
		Vault: tempVault,
	}

	err := sync.Run(options)
	assert.NoError(t, err)

	// Verify files were created in correct folders
	peopleDir := filepath.Join(tempVault, "People")
	badPeopleDir := filepath.Join(tempVault, "Bad People")
	friendsDir := filepath.Join(tempVault, "Friends")

	// User 11111 should be in People (no keywords matched)
	user1Path := filepath.Join(peopleDir, "user-11111.md")
	_, err = os.Stat(user1Path)
	assert.NoError(t, err, "User 11111 should be in People folder")

	// User 22222 should be in Bad People (matches "creepy" and "harassing")
	user2Path := filepath.Join(badPeopleDir, "user-22222.md")
	_, err = os.Stat(user2Path)
	assert.NoError(t, err, "User 22222 should be in Bad People folder")

	// User 33333 should be in Bad People (matches "stalker")
	user3Path := filepath.Join(badPeopleDir, "user-33333.md")
	_, err = os.Stat(user3Path)
	assert.NoError(t, err, "User 33333 should be in Bad People folder")

	// User 44444 should be in Friends (matches "cool")
	user4Path := filepath.Join(friendsDir, "user-44444.md")
	_, err = os.Stat(user4Path)
	assert.NoError(t, err, "User 44444 should be in Friends folder")

	// User 55555 should be in Friends (matches "friend")
	user5Path := filepath.Join(friendsDir, "user-55555.md")
	_, err = os.Stat(user5Path)
	assert.NoError(t, err, "User 55555 should be in Friends folder")
}

func TestSyncCmd_BlockedUserInBadPeopleFolder(t *testing.T) {
	// Create a temporary vault
	tempVault := t.TempDir()

	// Create Templates directory
	templatesDir := filepath.Join(tempVault, "Templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateContent := `---
tags:
  - person
url: https://fetlife.com/users/
---

# Notes
`
	templatePath := filepath.Join(templatesDir, "People.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Create test data directory
	testDataDir := t.TempDir()

	// Create blockeds.txt with a user nicknamed "CreepyPerson"
	blockedsContent := `user_id,created_at,updated_at,nickname
66666,2024-01-01,2024-01-01,CreepyPerson
77777,2024-01-01,2024-01-01,NormalPerson
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	if err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644); err != nil {
		t.Fatalf("Failed to create blockeds.txt: %v", err)
	}

	// Create empty private_notes.txt
	privateNotesContent := `member_id,created_at,updated_at,private_note
`
	privateNotesPath := filepath.Join(testDataDir, "private_notes.txt")
	if err := os.WriteFile(privateNotesPath, []byte(privateNotesContent), 0644); err != nil {
		t.Fatalf("Failed to create private_notes.txt: %v", err)
	}

	// Create sync command with CreateBlockedIn set to "Bad People"
	sync := &SyncCmd{
		DataDir:         testDataDir,
		CreatePeopleIn:  []string{"People"},
		CreateBlockedIn: "Bad People",
	}

	options := &Options{
		Vault: tempVault,
	}

	err := sync.Run(options)
	assert.NoError(t, err)

	// Both blocked users should be in Bad People folder (CreateBlockedIn setting)
	badPeopleDir := filepath.Join(tempVault, "Bad People")

	// User 66666 (CreepyPerson) should be in Bad People folder
	user1Path := filepath.Join(badPeopleDir, "CreepyPerson.md")
	_, err = os.Stat(user1Path)
	assert.NoError(t, err, "CreepyPerson should be created in Bad People folder (CreateBlockedIn)")

	// User 77777 should also be in Bad People folder
	user2Path := filepath.Join(badPeopleDir, "NormalPerson.md")
	_, err = os.Stat(user2Path)
	assert.NoError(t, err, "NormalPerson should be in Bad People folder (CreateBlockedIn)")

	// Verify the blocked tag was added to both users
	user1, err := obsidian.LoadPage(user1Path, tempVault)
	assert.NoError(t, err)
	assert.Contains(t, user1.Tags, "blocked", "CreepyPerson should have 'blocked' tag")

	user2, err := obsidian.LoadPage(user2Path, tempVault)
	assert.NoError(t, err)
	assert.Contains(t, user2.Tags, "blocked", "NormalPerson should have 'blocked' tag")
}

func TestSyncCmd_PrivateNoteWithBlockedKeyword(t *testing.T) {
	// Create a temporary vault
	tempVault := t.TempDir()

	// Create Templates directory
	templatesDir := filepath.Join(tempVault, "Templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	templateContent := `---
tags:
  - person
url: https://fetlife.com/users/
---

# Notes
`
	templatePath := filepath.Join(templatesDir, "People.md")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template: %v", err)
	}

	// Create test data directory
	testDataDir := t.TempDir()

	// Create empty blockeds.txt
	blockedsContent := `user_id,created_at,updated_at,nickname
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	if err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644); err != nil {
		t.Fatalf("Failed to create blockeds.txt: %v", err)
	}

	// Create private_notes.txt with users who should go to Bad People
	privateNotesContent := `member_id,created_at,updated_at,private_note
88888,2024-01-01,2024-01-01,Blocked this creepy person immediately
99999,2024-01-01,2024-01-01,Harassment and inappropriate messages - BLOCKED
`
	privateNotesPath := filepath.Join(testDataDir, "private_notes.txt")
	if err := os.WriteFile(privateNotesPath, []byte(privateNotesContent), 0644); err != nil {
		t.Fatalf("Failed to create private_notes.txt: %v", err)
	}

	// Create sync command
	sync := &SyncCmd{
		DataDir:        testDataDir,
		CreatePeopleIn: []string{"People", "Bad People:creepy,harassment,blocked"},
	}

	options := &Options{
		Vault: tempVault,
	}

	err := sync.Run(options)
	assert.NoError(t, err)

	// Both users should be in Bad People folder due to keyword matching
	badPeopleDir := filepath.Join(tempVault, "Bad People")

	// User 88888 should be in Bad People (matches "creepy" and "blocked")
	user1Path := filepath.Join(badPeopleDir, "user-88888.md")
	_, err = os.Stat(user1Path)
	assert.NoError(t, err, "User 88888 should be in Bad People folder (matches 'creepy' and 'blocked')")

	// User 99999 should be in Bad People (matches "harassment" and "blocked")
	user2Path := filepath.Join(badPeopleDir, "user-99999.md")
	_, err = os.Stat(user2Path)
	assert.NoError(t, err, "User 99999 should be in Bad People folder (matches 'harassment' and 'blocked')")

	// Verify the private notes were saved
	user1, err := obsidian.LoadPage(user1Path, tempVault)
	assert.NoError(t, err)
	assert.Equal(t, "Blocked this creepy person immediately", user1.WebMessage)

	user2, err := obsidian.LoadPage(user2Path, tempVault)
	assert.NoError(t, err)
	assert.Equal(t, "Harassment and inappropriate messages - BLOCKED", user2.WebMessage)
}
