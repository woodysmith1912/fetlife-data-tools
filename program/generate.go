package program

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/woodysmith1912/fetlife-data-tools/fetlife"
	"github.com/xuri/excelize/v2"
)

type GenerateCmd struct {
	DataDir   string `help:"Path to data directory containing blockeds.txt and private_notes.txt" env:"DATA_DIR" type:"existingdir" required:"true"`
	OutputDir string `help:"Path to output directory for generated spreadsheets" default:"." type:"existingdir"`
	Basename  string `help:"Base name for output files (without extension)" default:"fetlife-export"`
	Format    string `help:"Output format: csv, xlsx, or both" enum:"csv,xlsx,both" default:"csv"`
}

// MergedUser represents combined data from blocked users and private notes
type MergedUser struct {
	UserID       string
	Nickname     string
	URL          string
	Blocked      bool
	BlockedAt    string
	PrivateNote  string
	NoteCreated  string
	NoteUpdated  string
}

// Run generates CSV and XLSX spreadsheets from FetLife data
func (generate *GenerateCmd) Run(options *Options) error {
	log.Info().
		Str("dataDir", generate.DataDir).
		Str("outputDir", generate.OutputDir).
		Msg("Starting spreadsheet generation")

	// Read FetLife data
	blockeds, err := fetlife.ReadBlockeds(generate.DataDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read blockeds.txt")
		return err
	}
	log.Info().Int("blockedCount", len(blockeds)).Msg("Loaded blocked users")

	privateNotes, err := fetlife.ReadPrivateNotes(generate.DataDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read private_notes.txt")
		return err
	}
	log.Info().Int("privateNoteCount", len(privateNotes)).Msg("Loaded private notes")

	// Merge data by user ID
	merged := mergeUserData(blockeds, privateNotes)
	log.Info().Int("totalUsers", len(merged)).Msg("Merged user data")

	// Generate CSV if requested
	if generate.Format == "csv" || generate.Format == "both" {
		csvPath := filepath.Join(generate.OutputDir, generate.Basename+".csv")
		if err := generate.writeCSV(csvPath, merged); err != nil {
			log.Error().Err(err).Msg("Failed to write CSV")
			return err
		}
		log.Info().Str("path", csvPath).Msg("Generated CSV file")
	}

	// Generate XLSX if requested
	if generate.Format == "xlsx" || generate.Format == "both" {
		xlsxPath := filepath.Join(generate.OutputDir, generate.Basename+".xlsx")
		if err := generate.writeXLSX(xlsxPath, merged); err != nil {
			log.Error().Err(err).Msg("Failed to write XLSX")
			return err
		}
		log.Info().Str("path", xlsxPath).Msg("Generated XLSX file")
	}

	log.Info().Msg("Spreadsheet generation completed successfully")
	return nil
}

// mergeUserData combines blocked users and private notes into a single dataset
func mergeUserData(blockeds []fetlife.BlockedRecord, privateNotes []fetlife.PrivateNoteRecord) []MergedUser {
	// Create a map to hold merged data
	userMap := make(map[string]*MergedUser)

	// Add blocked users
	for _, blocked := range blockeds {
		userMap[blocked.UserID] = &MergedUser{
			UserID:    blocked.UserID,
			Nickname:  blocked.Nickname,
			URL:       fmt.Sprintf("https://fetlife.com/users/%s", blocked.UserID),
			Blocked:   true,
			BlockedAt: blocked.CreatedAt,
		}
	}

	// Add/merge private notes
	for _, note := range privateNotes {
		if existing, ok := userMap[note.MemberID]; ok {
			// User already exists (blocked user with a note)
			existing.PrivateNote = note.PrivateNote
			existing.NoteCreated = note.CreatedAt
			existing.NoteUpdated = note.UpdatedAt
		} else {
			// New user from private notes only
			userMap[note.MemberID] = &MergedUser{
				UserID:      note.MemberID,
				URL:         fmt.Sprintf("https://fetlife.com/users/%s", note.MemberID),
				Blocked:     false,
				PrivateNote: note.PrivateNote,
				NoteCreated: note.CreatedAt,
				NoteUpdated: note.UpdatedAt,
			}
		}
	}

	// Convert map to slice
	result := make([]MergedUser, 0, len(userMap))
	for _, user := range userMap {
		result = append(result, *user)
	}

	return result
}

// writeCSV writes merged user data to a CSV file
func (generate *GenerateCmd) writeCSV(path string, users []MergedUser) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"User ID",
		"Nickname",
		"URL",
		"Blocked",
		"Blocked At",
		"Private Note",
		"Note Created",
		"Note Updated",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, user := range users {
		blocked := "No"
		if user.Blocked {
			blocked = "Yes"
		}

		record := []string{
			user.UserID,
			user.Nickname,
			user.URL,
			blocked,
			user.BlockedAt,
			user.PrivateNote,
			user.NoteCreated,
			user.NoteUpdated,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// writeXLSX writes merged user data to an Excel file
func (generate *GenerateCmd) writeXLSX(path string, users []MergedUser) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close Excel file")
		}
	}()

	sheetName := "FetLife Data"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// Set header with bold style
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	if err != nil {
		return err
	}

	headers := []string{"User ID", "Nickname", "URL", "Blocked", "Blocked At", "Private Note", "Note Created", "Note Updated"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 12) // User ID
	f.SetColWidth(sheetName, "B", "B", 20) // Nickname
	f.SetColWidth(sheetName, "C", "C", 35) // URL
	f.SetColWidth(sheetName, "D", "D", 10) // Blocked
	f.SetColWidth(sheetName, "E", "E", 20) // Blocked At
	f.SetColWidth(sheetName, "F", "F", 50) // Private Note
	f.SetColWidth(sheetName, "G", "G", 20) // Note Created
	f.SetColWidth(sheetName, "H", "H", 20) // Note Updated

	// Write data
	for i, user := range users {
		row := i + 2 // Start at row 2 (row 1 is header)

		blocked := "No"
		if user.Blocked {
			blocked = "Yes"
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), user.UserID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), user.Nickname)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user.URL)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), blocked)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), user.BlockedAt)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), user.PrivateNote)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), user.NoteCreated)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), user.NoteUpdated)
	}

	// Delete default Sheet1 if it exists
	f.DeleteSheet("Sheet1")

	// Save the file
	if err := f.SaveAs(path); err != nil {
		return err
	}

	return nil
}
