package obsidian

import (
	"os"
	"path/filepath"
	"testing"
)

func getExampleVaultPath(t *testing.T) string {
	// Get the path to the example vault relative to this test file
	path, err := filepath.Abs("../example/vault")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	return path
}

func TestVaultLoad(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	// We created 15 total files: 5 in root, 5 in People, 5 in Bad People
	// Plus 1 template file in Templates folder
	expectedPageCount := 16
	if len(vault.Pages) != expectedPageCount {
		t.Errorf("Expected %d pages, got %d", expectedPageCount, len(vault.Pages))
	}
}

func TestVaultLoadPageMetadata(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	// Find Alice's page to test metadata parsing
	var alice *Page
	for _, page := range vault.Pages {
		if page.Title == "Alice" {
			alice = page
			break
		}
	}

	if alice == nil {
		t.Fatal("Could not find Alice's page")
	}

	// Test title
	if alice.Title != "Alice" {
		t.Errorf("Expected title 'Alice', got '%s'", alice.Title)
	}

	// Test folder
	if alice.Folder != "People" {
		t.Errorf("Expected folder 'People', got '%s'", alice.Folder)
	}

	// Test tags
	expectedTags := []string{"person", "friend"}
	if len(alice.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(alice.Tags))
	}
	for i, tag := range expectedTags {
		if i >= len(alice.Tags) || alice.Tags[i] != tag {
			t.Errorf("Expected tag '%s' at position %d, got '%s'", tag, i, alice.Tags[i])
		}
	}

	// Test aliases
	expectedAliases := []string{"Ally", "A-Train"}
	if len(alice.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(alice.Aliases))
	}

	// Test URL
	expectedURL := "https://fetlife.com/users/12345"
	if alice.Url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, alice.Url)
	}

	// Test URL aliases
	expectedURLAliases := 2
	if len(alice.UrlAliases) != expectedURLAliases {
		t.Errorf("Expected %d URL aliases, got %d", expectedURLAliases, len(alice.UrlAliases))
	}

	// Test web badge color
	expectedColor := Color("#4CAF50")
	if alice.WebBadgeColor != expectedColor {
		t.Errorf("Expected color '%s', got '%s'", expectedColor, alice.WebBadgeColor)
	}

	// Test web message
	expectedMessage := "This is Alice's profile!"
	if alice.WebMessage != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, alice.WebMessage)
	}
}

func TestVaultInFolder(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	tests := []struct {
		folder        string
		expectedCount int
		expectedNames []string
	}{
		{
			folder:        "People",
			expectedCount: 5,
			expectedNames: []string{"Alice", "Bob", "Carol", "David", "Emma"},
		},
		{
			folder:        "Bad People",
			expectedCount: 5,
			expectedNames: []string{"Frank", "George", "Helen", "Ian", "Jane"},
		},
		{
			folder:        ".",
			expectedCount: 5,
			expectedNames: []string{"Index", "Projects", "Notes", "Resources", "About"},
		},
		{
			folder:        "",
			expectedCount: 5,
			expectedNames: []string{"Index", "Projects", "Notes", "Resources", "About"},
		},
	}

	for _, tt := range tests {
		t.Run("folder_"+tt.folder, func(t *testing.T) {
			pages := vault.InFolder(tt.folder)

			if len(pages) != tt.expectedCount {
				t.Errorf("Expected %d pages in folder '%s', got %d", tt.expectedCount, tt.folder, len(pages))
			}

			// Create a map of page titles for easy lookup
			pageNames := make(map[string]bool)
			for _, page := range pages {
				pageNames[page.Title] = true
			}

			// Check that all expected names are present
			for _, name := range tt.expectedNames {
				if !pageNames[name] {
					t.Errorf("Expected to find page '%s' in folder '%s', but it was not found", name, tt.folder)
				}
			}
		})
	}
}

func TestVaultWithTag(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	tests := []struct {
		tag           string
		expectedCount int
		shouldContain []string
	}{
		{
			tag:           "person",
			expectedCount: 10, // 5 in People + 5 in Bad People
			shouldContain: []string{"Alice", "Bob", "Carol", "Frank", "George"},
		},
		{
			tag:           "friend",
			expectedCount: 3, // Alice, Carol, Emma
			shouldContain: []string{"Alice", "Carol", "Emma"},
		},
		{
			tag:           "blocked",
			expectedCount: 5, // All Bad People
			shouldContain: []string{"Frank", "George", "Helen", "Ian", "Jane"},
		},
		{
			tag:           "meta",
			expectedCount: 2, // Index and About
			shouldContain: []string{"Index", "About"},
		},
		{
			tag:           "nonexistent",
			expectedCount: 0,
			shouldContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run("tag_"+tt.tag, func(t *testing.T) {
			pages := vault.WithTag(tt.tag)

			if len(pages) != tt.expectedCount {
				t.Errorf("Expected %d pages with tag '%s', got %d", tt.expectedCount, tt.tag, len(pages))
			}

			// Create a map of page titles for easy lookup
			pageNames := make(map[string]bool)
			for _, page := range pages {
				pageNames[page.Title] = true
			}

			// Check that all expected pages are present
			for _, name := range tt.shouldContain {
				if !pageNames[name] {
					t.Errorf("Expected to find page '%s' with tag '%s', but it was not found", name, tt.tag)
				}
			}
		})
	}
}

func TestVaultLoadEmptyMetadata(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	// Find Notes page which has minimal metadata
	var notes *Page
	for _, page := range vault.Pages {
		if page.Title == "Notes" {
			notes = page
			break
		}
	}

	if notes == nil {
		t.Fatal("Could not find Notes page")
	}

	// Test that it has tags but no URL
	if len(notes.Tags) == 0 {
		t.Error("Expected Notes to have tags")
	}

	if notes.Url != "" {
		t.Errorf("Expected Notes to have empty URL, got '%s'", notes.Url)
	}

	if notes.WebBadgeColor != "" {
		t.Errorf("Expected Notes to have empty WebBadgeColor, got '%s'", notes.WebBadgeColor)
	}
}

func TestVaultLoadComplexMetadata(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	// Find Resources page which has URL aliases
	var resources *Page
	for _, page := range vault.Pages {
		if page.Title == "Resources" {
			resources = page
			break
		}
	}

	if resources == nil {
		t.Fatal("Could not find Resources page")
	}

	// Test URL aliases
	if len(resources.UrlAliases) != 2 {
		t.Errorf("Expected 2 URL aliases for Resources, got %d", len(resources.UrlAliases))
	}

	expectedURLAliases := map[string]bool{
		"https://example.com/refs":  true,
		"https://example.com/links": true,
	}

	for _, urlAlias := range resources.UrlAliases {
		if !expectedURLAliases[urlAlias] {
			t.Errorf("Unexpected URL alias: %s", urlAlias)
		}
	}
}

func TestVaultBadPeopleMetadata(t *testing.T) {
	vault := NewVault(getExampleVaultPath(t))

	err := vault.Load()
	if err != nil {
		t.Fatalf("Failed to load vault: %v", err)
	}

	// Find Frank's page to test Bad People metadata
	var frank *Page
	for _, page := range vault.Pages {
		if page.Title == "Frank" {
			frank = page
			break
		}
	}

	if frank == nil {
		t.Fatal("Could not find Frank's page")
	}

	// Test that Frank is in Bad People folder
	if frank.Folder != "Bad People" {
		t.Errorf("Expected Frank to be in 'Bad People' folder, got '%s'", frank.Folder)
	}

	// Test that Frank has blocked tag
	hasBlockedTag := false
	for _, tag := range frank.Tags {
		if tag == "blocked" {
			hasBlockedTag = true
			break
		}
	}
	if !hasBlockedTag {
		t.Error("Expected Frank to have 'blocked' tag")
	}

	// Test that Frank has a warning message
	if frank.WebMessage == "" {
		t.Error("Expected Frank to have a web message warning")
	}

	// Test that Frank has a red badge color
	if frank.WebBadgeColor != "#F44336" {
		t.Errorf("Expected Frank to have red badge color, got '%s'", frank.WebBadgeColor)
	}
}

func TestPageSaveUpdateTags(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-page.md")

	// Write initial content with frontmatter
	initialContent := `---
tags:
  - original
  - test
aliases:
  - TestAlias
url: https://example.com/test
web-badge-color: "#FF0000"
web-message: "Original message"
---

# Test Page

This is the original content.
`

	err := os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load the page
	page, err := loadPage(testFile, tempDir)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Verify initial tags
	if len(page.Tags) != 2 {
		t.Fatalf("Expected 2 initial tags, got %d", len(page.Tags))
	}

	// Update tags
	page.Tags = []string{"updated", "modified", "new-tag"}

	// Save the page
	err = page.Save()
	if err != nil {
		t.Fatalf("Failed to save page: %v", err)
	}

	// Re-load the page to verify changes were saved
	reloadedPage, err := loadPage(testFile, tempDir)
	if err != nil {
		t.Fatalf("Failed to reload page: %v", err)
	}

	// Verify tags were updated
	if len(reloadedPage.Tags) != 3 {
		t.Errorf("Expected 3 tags after update, got %d", len(reloadedPage.Tags))
	}

	expectedTags := map[string]bool{
		"updated":  true,
		"modified": true,
		"new-tag":  true,
	}

	for _, tag := range reloadedPage.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag after update: %s", tag)
		}
	}

	// Verify other metadata was preserved
	if reloadedPage.Url != "https://example.com/test" {
		t.Errorf("URL was not preserved, got: %s", reloadedPage.Url)
	}

	if len(reloadedPage.Aliases) != 1 || reloadedPage.Aliases[0] != "TestAlias" {
		t.Errorf("Aliases were not preserved")
	}

	if reloadedPage.WebBadgeColor != "#FF0000" {
		t.Errorf("WebBadgeColor was not preserved, got: %s", reloadedPage.WebBadgeColor)
	}

	// Verify content was preserved
	if reloadedPage.Content != page.Content {
		t.Errorf("Content was not preserved")
	}
}

func TestPageSaveUpdateWebMessage(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-message.md")

	// Write initial content with frontmatter
	initialContent := `---
tags:
  - person
  - friend
url: https://fetlife.com/users/99999
web-badge-color: "#00FF00"
web-message: "Hello, this is the original message!"
---

# Test Person

This person is a test subject.
Some more content here.
`

	err := os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load the page
	page, err := loadPage(testFile, tempDir)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Verify initial web message
	if page.WebMessage != "Hello, this is the original message!" {
		t.Fatalf("Expected initial web message, got: %s", page.WebMessage)
	}

	// Update web message
	page.WebMessage = "This message has been updated!"

	// Save the page
	err = page.Save()
	if err != nil {
		t.Fatalf("Failed to save page: %v", err)
	}

	// Re-load the page to verify changes were saved
	reloadedPage, err := loadPage(testFile, tempDir)
	if err != nil {
		t.Fatalf("Failed to reload page: %v", err)
	}

	// Verify web message was updated
	if reloadedPage.WebMessage != "This message has been updated!" {
		t.Errorf("Expected updated web message, got: %s", reloadedPage.WebMessage)
	}

	// Verify other metadata was preserved
	if len(reloadedPage.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(reloadedPage.Tags))
	}

	expectedTags := map[string]bool{
		"person": true,
		"friend": true,
	}

	for _, tag := range reloadedPage.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}

	if reloadedPage.Url != "https://fetlife.com/users/99999" {
		t.Errorf("URL was not preserved, got: %s", reloadedPage.Url)
	}

	if reloadedPage.WebBadgeColor != "#00FF00" {
		t.Errorf("WebBadgeColor was not preserved, got: %s", reloadedPage.WebBadgeColor)
	}

	// Verify content was preserved
	if reloadedPage.Content != page.Content {
		t.Errorf("Content was not preserved")
	}
}

func TestPageSaveUpdateBothTagsAndWebMessage(t *testing.T) {
	// Create a temporary test file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test-both.md")

	// Write initial content
	initialContent := `---
tags:
  - initial
url: https://example.com
web-message: "Initial message"
---

# Both Updates Test

Testing simultaneous updates.
`

	err := os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load the page
	page, err := loadPage(testFile, tempDir)
	if err != nil {
		t.Fatalf("Failed to load page: %v", err)
	}

	// Update both tags and web message
	page.Tags = []string{"updated", "simultaneous", "multiple"}
	page.WebMessage = "Both fields updated!"

	// Save the page
	err = page.Save()
	if err != nil {
		t.Fatalf("Failed to save page: %v", err)
	}

	// Re-load the page
	reloadedPage, err := loadPage(testFile, tempDir)
	if err != nil {
		t.Fatalf("Failed to reload page: %v", err)
	}

	// Verify both updates
	if len(reloadedPage.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(reloadedPage.Tags))
	}

	if reloadedPage.WebMessage != "Both fields updated!" {
		t.Errorf("Expected updated message, got: %s", reloadedPage.WebMessage)
	}

	// Verify URL was preserved
	if reloadedPage.Url != "https://example.com" {
		t.Errorf("URL was not preserved, got: %s", reloadedPage.Url)
	}
}
