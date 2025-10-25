package program

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/woodysmith1912/fetlife-data-tools/obsidian"
)

type SyncCmd struct {
	DataDir         string   `help:"Path to data directory containing blockeds.txt and private_notes.txt" env:"DATA_DIR" type:"existingdir" required:"true"`
	CreatePeopleIn  []string `alias:"in" help:"List of Obsidian folders to create individual people.  Syntax is folder[:keyword1,...] and this folder will be used if one of the keywords is found in the private note.  Keywords are not case sensitive" default:"People"`
	CreateBlockedIn string   `help:"Obsidian folder to create blocked people in" default:"Bad People"`
}

type BlockedRecord struct {
	UserID    string
	CreatedAt string
	UpdatedAt string
	Nickname  string
}

type PrivateNoteRecord struct {
	MemberID    string
	CreatedAt   string
	UpdatedAt   string
	PrivateNote string
}

func (sync *SyncCmd) Run(options *Options) error {
	log.Info().
		Str("vault", options.Vault).
		Str("dataDir", sync.DataDir).
		Msg("Starting sync")

	// Load the vault
	vault := obsidian.NewVault(options.Vault)
	if err := vault.Load(); err != nil {
		log.Error().Err(err).Msg("Failed to load vault")
		return err
	}

	log.Info().Int("pageCount", len(vault.Pages)).Msg("Loaded vault")

	// Read blockeds.txt
	blockeds, err := sync.readBlockeds()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read blockeds.txt")
		return err
	}
	log.Info().Int("blockedCount", len(blockeds)).Msg("Loaded blockeds")

	// Read private_notes.txt
	privateNotes, err := sync.readPrivateNotes()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read private_notes.txt")
		return err
	}
	log.Info().Int("privateNoteCount", len(privateNotes)).Msg("Loaded private notes")

	// Process blockeds
	for _, blocked := range blockeds {
		if err := sync.processBlocked(vault, blocked, options); err != nil {
			log.Error().Err(err).Str("userID", blocked.UserID).Msg("Failed to process blocked user")
			// Continue processing other records
		}
	}

	// Process private notes
	for _, note := range privateNotes {
		if err := sync.processPrivateNote(vault, note, options); err != nil {
			log.Error().Err(err).Str("memberID", note.MemberID).Msg("Failed to process private note")
			// Continue processing other records
		}
	}

	log.Info().Msg("Sync completed successfully")
	return nil
}

func (sync *SyncCmd) readBlockeds() ([]BlockedRecord, error) {
	path := filepath.Join(sync.DataDir, "blockeds.txt")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var blockeds []BlockedRecord
	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}
		if len(record) < 4 {
			log.Warn().Int("line", i+1).Msg("Skipping invalid blocked record")
			continue
		}
		blockeds = append(blockeds, BlockedRecord{
			UserID:    record[0],
			CreatedAt: record[1],
			UpdatedAt: record[2],
			Nickname:  record[3],
		})
	}

	return blockeds, nil
}

func (sync *SyncCmd) readPrivateNotes() ([]PrivateNoteRecord, error) {
	path := filepath.Join(sync.DataDir, "private_notes.txt")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var notes []PrivateNoteRecord
	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}
		if len(record) < 4 {
			log.Warn().Int("line", i+1).Msg("Skipping invalid private note record")
			continue
		}
		notes = append(notes, PrivateNoteRecord{
			MemberID:    record[0],
			CreatedAt:   record[1],
			UpdatedAt:   record[2],
			PrivateNote: record[3],
		})
	}

	return notes, nil
}

// findPageByUserID finds a page by matching the user ID in the URL or URL aliases
func (sync *SyncCmd) findPageByUserID(vault *obsidian.Vault, userID string) ([]*obsidian.Page, error) {
	var matches []*obsidian.Page

	for _, page := range vault.Pages {
		// Check main URL
		if strings.Contains(page.Url, "/users/"+userID) || strings.HasSuffix(page.Url, "/"+userID) {
			matches = append(matches, page)
			continue
		}

		// Check URL aliases
		for _, urlAlias := range page.UrlAliases {
			if strings.Contains(urlAlias, "/users/"+userID) || strings.HasSuffix(urlAlias, "/"+userID) {
				matches = append(matches, page)
				break
			}
		}
	}

	return matches, nil
}

func (sync *SyncCmd) processBlocked(vault *obsidian.Vault, blocked BlockedRecord, options *Options) error {
	pages, err := sync.findPageByUserID(vault, blocked.UserID)
	if err != nil {
		return err
	}

	if len(pages) > 1 {
		log.Warn().
			Str("userID", blocked.UserID).
			Int("matchCount", len(pages)).
			Msg("Multiple pages found for user ID, skipping")
		return nil
	}

	var page *obsidian.Page
	if len(pages) == 0 {
		// Create new page from template in the CreateBlockedIn folder
		log.Info().
			Str("userID", blocked.UserID).
			Str("nickname", blocked.Nickname).
			Str("folder", sync.CreateBlockedIn).
			Msg("Creating new page for blocked user")

		page, err = sync.createPageInFolder(vault, blocked.UserID, blocked.Nickname, sync.CreateBlockedIn, options)
		if err != nil {
			return err
		}
	} else {
		page = pages[0]
		log.Info().
			Str("userID", blocked.UserID).
			Str("page", page.Title).
			Msg("Updating existing page for blocked user")
	}

	// Ensure "blocked" tag is present
	hasBlockedTag := false
	for _, tag := range page.Tags {
		if tag == "blocked" {
			hasBlockedTag = true
			break
		}
	}
	if !hasBlockedTag {
		page.Tags = append(page.Tags, "blocked")
	}

	// Add block-date metadata (we'll need to add this field to the Page struct)
	// For now, we'll set it as a web message if not already set
	if page.WebMessage == "" {
		page.WebMessage = fmt.Sprintf("Blocked on %s", blocked.CreatedAt)
	}

	// Save the page
	if err := page.Save(); err != nil {
		return err
	}

	log.Info().
		Str("userID", blocked.UserID).
		Str("page", page.Title).
		Msg("Successfully updated blocked user page")

	return nil
}

func (sync *SyncCmd) processPrivateNote(vault *obsidian.Vault, note PrivateNoteRecord, options *Options) error {
	pages, err := sync.findPageByUserID(vault, note.MemberID)
	if err != nil {
		return err
	}

	if len(pages) > 1 {
		log.Warn().
			Str("memberID", note.MemberID).
			Int("matchCount", len(pages)).
			Msg("Multiple pages found for member ID, skipping")
		return nil
	}

	var page *obsidian.Page
	if len(pages) == 0 {
		// Create new page from template, passing the private note for folder determination
		log.Info().
			Str("memberID", note.MemberID).
			Msg("Creating new page for member with private note")

		page, err = sync.createPageFromTemplateWithNote(vault, note.MemberID, "", note.PrivateNote, options)
		if err != nil {
			return err
		}
	} else {
		page = pages[0]
		log.Info().
			Str("memberID", note.MemberID).
			Str("page", page.Title).
			Msg("Updating existing page with private note")
	}

	// Update web-message with private note
	page.WebMessage = note.PrivateNote

	// Save the page
	if err := page.Save(); err != nil {
		return err
	}

	log.Info().
		Str("memberID", note.MemberID).
		Str("page", page.Title).
		Msg("Successfully updated page with private note")

	return nil
}

// parseFolderConfig parses a folder configuration string like "People:keyword1,keyword2"
// Returns the folder name and list of keywords (all lowercase)
func parseFolderConfig(config string) (folder string, keywords []string) {
	parts := strings.SplitN(config, ":", 2)
	folder = parts[0]

	if len(parts) == 2 && parts[1] != "" {
		keywordParts := strings.Split(parts[1], ",")
		for _, kw := range keywordParts {
			trimmed := strings.TrimSpace(kw)
			if trimmed != "" {
				keywords = append(keywords, strings.ToLower(trimmed))
			}
		}
	}

	return folder, keywords
}

// determineFolderForUser determines which folder to place a user's page in
// based on the CreatePeopleIn configuration and the private note content
func (sync *SyncCmd) determineFolderForUser(userID, privateNote string) string {
	if len(sync.CreatePeopleIn) == 0 {
		return "People"
	}

	// If we have a private note, try to match keywords
	if privateNote != "" {
		lowerNote := strings.ToLower(privateNote)

		for _, config := range sync.CreatePeopleIn {
			folder, keywords := parseFolderConfig(config)

			// If this folder has keywords, check for matches
			if len(keywords) > 0 {
				for _, keyword := range keywords {
					if strings.Contains(lowerNote, keyword) {
						log.Info().
							Str("userID", userID).
							Str("folder", folder).
							Str("keyword", keyword).
							Msg("Matched keyword, placing in folder")
						return folder
					}
				}
			}
		}
	}

	// Default to the first folder
	folder, _ := parseFolderConfig(sync.CreatePeopleIn[0])
	return folder
}

// createPageInFolder creates a page in a specific folder
func (sync *SyncCmd) createPageInFolder(vault *obsidian.Vault, userID, nickname, folder string, options *Options) (*obsidian.Page, error) {
	// Determine page name
	pageName := nickname
	if pageName == "" {
		pageName = fmt.Sprintf("user-%s", userID)
	}

	folderPath := filepath.Join(options.Vault, folder)

	// Create folder if it doesn't exist
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, err
	}

	// Create file path
	filePath := filepath.Join(folderPath, pageName+".md")

	// Read template
	templatePath := filepath.Join(options.Vault, "Templates", "People.md")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		log.Warn().Err(err).Msg("Template not found, using default")
		// Use a default template
		templateContent = []byte(`---
tags:
  - person
url: https://fetlife.com/users/` + userID + `
---

# Notes
`)
	}

	// Replace {{title}} placeholder in template
	content := strings.ReplaceAll(string(templateContent), "{{title}}", pageName)

	// Update URL in template to include the user ID
	content = strings.ReplaceAll(content, "url: https://fetlife.com/users/", "url: https://fetlife.com/users/"+userID)

	// Write the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, err
	}

	// Load the newly created page
	page, err := obsidian.LoadPage(filePath, options.Vault)
	if err != nil {
		return nil, err
	}

	// Add to vault
	vault.Pages = append(vault.Pages, page)

	log.Info().
		Str("page", pageName).
		Str("path", filePath).
		Str("folder", folder).
		Msg("Created new page from template")

	return page, nil
}

// createPageFromTemplateWithNote creates a page with private note for folder determination
func (sync *SyncCmd) createPageFromTemplateWithNote(vault *obsidian.Vault, userID, nickname, privateNote string, options *Options) (*obsidian.Page, error) {
	// Determine folder based on CreatePeopleIn flag and private note
	folder := sync.determineFolderForUser(userID, privateNote)
	return sync.createPageInFolder(vault, userID, nickname, folder, options)
}

func (sync *SyncCmd) createPageFromTemplate(vault *obsidian.Vault, userID, nickname string, options *Options) (*obsidian.Page, error) {
	return sync.createPageFromTemplateWithNote(vault, userID, nickname, "", options)
}
