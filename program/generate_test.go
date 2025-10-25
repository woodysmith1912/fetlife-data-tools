package program

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woodysmith1912/fetlife-data-tools/fetlife"
	"github.com/xuri/excelize/v2"
)

func TestMergeUserData(t *testing.T) {
	tests := []struct {
		name         string
		blockeds     []fetlife.BlockedRecord
		privateNotes []fetlife.PrivateNoteRecord
		expectedLen  int
		validate     func(*testing.T, []MergedUser)
	}{
		{
			name: "merge blocked users only",
			blockeds: []fetlife.BlockedRecord{
				{UserID: "123", Nickname: "BadUser", CreatedAt: "2024-01-01", UpdatedAt: "2024-01-01"},
				{UserID: "456", Nickname: "AnotherBad", CreatedAt: "2024-01-02", UpdatedAt: "2024-01-02"},
			},
			privateNotes: []fetlife.PrivateNoteRecord{},
			expectedLen:  2,
			validate: func(t *testing.T, users []MergedUser) {
				for _, user := range users {
					assert.True(t, user.Blocked, "User should be marked as blocked")
					assert.NotEmpty(t, user.Nickname, "Nickname should be set")
					assert.NotEmpty(t, user.BlockedAt, "BlockedAt should be set")
					assert.Empty(t, user.PrivateNote, "PrivateNote should be empty")
				}
			},
		},
		{
			name:     "merge private notes only",
			blockeds: []fetlife.BlockedRecord{},
			privateNotes: []fetlife.PrivateNoteRecord{
				{MemberID: "789", PrivateNote: "Nice person", CreatedAt: "2024-01-01", UpdatedAt: "2024-01-01"},
				{MemberID: "101", PrivateNote: "Met at event", CreatedAt: "2024-01-02", UpdatedAt: "2024-01-02"},
			},
			expectedLen: 2,
			validate: func(t *testing.T, users []MergedUser) {
				for _, user := range users {
					assert.False(t, user.Blocked, "User should not be blocked")
					assert.NotEmpty(t, user.PrivateNote, "PrivateNote should be set")
					assert.Empty(t, user.Nickname, "Nickname should be empty")
					assert.Empty(t, user.BlockedAt, "BlockedAt should be empty")
				}
			},
		},
		{
			name: "merge both blocked and notes for same user",
			blockeds: []fetlife.BlockedRecord{
				{UserID: "123", Nickname: "BlockedUser", CreatedAt: "2024-01-01", UpdatedAt: "2024-01-01"},
			},
			privateNotes: []fetlife.PrivateNoteRecord{
				{MemberID: "123", PrivateNote: "This person is creepy", CreatedAt: "2024-01-02", UpdatedAt: "2024-01-02"},
			},
			expectedLen: 1,
			validate: func(t *testing.T, users []MergedUser) {
				assert.Len(t, users, 1)
				user := users[0]
				assert.Equal(t, "123", user.UserID)
				assert.True(t, user.Blocked)
				assert.Equal(t, "BlockedUser", user.Nickname)
				assert.Equal(t, "This person is creepy", user.PrivateNote)
				assert.NotEmpty(t, user.BlockedAt)
				assert.NotEmpty(t, user.NoteCreated)
			},
		},
		{
			name: "merge multiple users with mixed data",
			blockeds: []fetlife.BlockedRecord{
				{UserID: "111", Nickname: "User1", CreatedAt: "2024-01-01", UpdatedAt: "2024-01-01"},
				{UserID: "222", Nickname: "User2", CreatedAt: "2024-01-02", UpdatedAt: "2024-01-02"},
			},
			privateNotes: []fetlife.PrivateNoteRecord{
				{MemberID: "222", PrivateNote: "Also has note", CreatedAt: "2024-01-03", UpdatedAt: "2024-01-03"},
				{MemberID: "333", PrivateNote: "Only note", CreatedAt: "2024-01-04", UpdatedAt: "2024-01-04"},
			},
			expectedLen: 3,
			validate: func(t *testing.T, users []MergedUser) {
				blockedCount := 0
				withNotesCount := 0
				for _, user := range users {
					if user.Blocked {
						blockedCount++
					}
					if user.PrivateNote != "" {
						withNotesCount++
					}
					// All users should have URLs
					assert.Contains(t, user.URL, "https://fetlife.com/users/")
				}
				assert.Equal(t, 2, blockedCount, "Should have 2 blocked users")
				assert.Equal(t, 2, withNotesCount, "Should have 2 users with notes")
			},
		},
		{
			name:         "empty input",
			blockeds:     []fetlife.BlockedRecord{},
			privateNotes: []fetlife.PrivateNoteRecord{},
			expectedLen:  0,
			validate: func(t *testing.T, users []MergedUser) {
				assert.Empty(t, users)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeUserData(tt.blockeds, tt.privateNotes)
			assert.Len(t, result, tt.expectedLen)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestWriteCSV(t *testing.T) {
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "test.csv")

	users := []MergedUser{
		{
			UserID:      "123",
			Nickname:    "TestUser",
			URL:         "https://fetlife.com/users/123",
			Blocked:     true,
			BlockedAt:   "2024-01-01",
			PrivateNote: "Test note",
			NoteCreated: "2024-01-02",
			NoteUpdated: "2024-01-03",
		},
		{
			UserID:      "456",
			URL:         "https://fetlife.com/users/456",
			Blocked:     false,
			PrivateNote: "Another note",
			NoteCreated: "2024-01-04",
			NoteUpdated: "2024-01-05",
		},
	}

	gen := &GenerateCmd{}
	err := gen.writeCSV(csvPath, users)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(csvPath)
	assert.NoError(t, err)

	// Read and verify CSV content
	file, err := os.Open(csvPath)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Check header
	assert.Len(t, records, 3) // header + 2 data rows
	assert.Equal(t, []string{"User ID", "Nickname", "URL", "Blocked", "Blocked At", "Private Note", "Note Created", "Note Updated"}, records[0])

	// Check first user
	assert.Equal(t, "123", records[1][0])
	assert.Equal(t, "TestUser", records[1][1])
	assert.Equal(t, "https://fetlife.com/users/123", records[1][2])
	assert.Equal(t, "Yes", records[1][3])
	assert.Equal(t, "2024-01-01", records[1][4])
	assert.Equal(t, "Test note", records[1][5])

	// Check second user
	assert.Equal(t, "456", records[2][0])
	assert.Equal(t, "", records[2][1])   // No nickname
	assert.Equal(t, "No", records[2][3]) // Not blocked
}

func TestWriteXLSX(t *testing.T) {
	tempDir := t.TempDir()
	xlsxPath := filepath.Join(tempDir, "test.xlsx")

	users := []MergedUser{
		{
			UserID:      "123",
			Nickname:    "TestUser",
			URL:         "https://fetlife.com/users/123",
			Blocked:     true,
			BlockedAt:   "2024-01-01",
			PrivateNote: "Test note",
			NoteCreated: "2024-01-02",
			NoteUpdated: "2024-01-03",
		},
	}

	gen := &GenerateCmd{}
	err := gen.writeXLSX(xlsxPath, users)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(xlsxPath)
	assert.NoError(t, err)

	// Open and verify XLSX content
	f, err := excelize.OpenFile(xlsxPath)
	assert.NoError(t, err)
	defer f.Close()

	// Check sheet exists
	sheets := f.GetSheetList()
	assert.Contains(t, sheets, "FetLife Data")

	// Verify headers
	headers := []string{"User ID", "Nickname", "URL", "Blocked", "Blocked At", "Private Note", "Note Created", "Note Updated"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		value, err := f.GetCellValue("FetLife Data", cell)
		assert.NoError(t, err)
		assert.Equal(t, header, value)
	}

	// Verify first row data
	userID, _ := f.GetCellValue("FetLife Data", "A2")
	assert.Equal(t, "123", userID)

	nickname, _ := f.GetCellValue("FetLife Data", "B2")
	assert.Equal(t, "TestUser", nickname)

	blocked, _ := f.GetCellValue("FetLife Data", "D2")
	assert.Equal(t, "Yes", blocked)

	note, _ := f.GetCellValue("FetLife Data", "F2")
	assert.Equal(t, "Test note", note)
}

func TestGenerateCmd_Run_CSV(t *testing.T) {
	// Create test data directory
	testDataDir := t.TempDir()
	outputDir := t.TempDir()

	// Create blockeds.txt
	blockedsContent := `user_id,created_at,updated_at,nickname
123,2024-01-01,2024-01-01,TestUser
456,2024-01-02,2024-01-02,AnotherUser
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644)
	assert.NoError(t, err)

	// Create private_notes.txt
	notesContent := `member_id,created_at,updated_at,private_note
123,2024-01-03,2024-01-03,Has a note too
789,2024-01-04,2024-01-04,Only has note
`
	notesPath := filepath.Join(testDataDir, "private_notes.txt")
	err = os.WriteFile(notesPath, []byte(notesContent), 0644)
	assert.NoError(t, err)

	// Run generate command for CSV
	gen := &GenerateCmd{
		DataDir:   testDataDir,
		OutputDir: outputDir,
		Basename:  "test-output",
		Format:    "csv",
	}

	err = gen.Run(&Options{})
	assert.NoError(t, err)

	// Verify CSV was created
	csvPath := filepath.Join(outputDir, "test-output.csv")
	_, err = os.Stat(csvPath)
	assert.NoError(t, err)

	// Verify XLSX was NOT created
	xlsxPath := filepath.Join(outputDir, "test-output.xlsx")
	_, err = os.Stat(xlsxPath)
	assert.True(t, os.IsNotExist(err), "XLSX file should not exist")

	// Verify CSV content
	file, err := os.Open(csvPath)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 4) // header + 3 users (2 blocked, 1 note-only)
}

func TestGenerateCmd_Run_XLSX(t *testing.T) {
	// Create test data directory
	testDataDir := t.TempDir()
	outputDir := t.TempDir()

	// Create blockeds.txt
	blockedsContent := `user_id,created_at,updated_at,nickname
123,2024-01-01,2024-01-01,TestUser
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644)
	assert.NoError(t, err)

	// Create private_notes.txt (empty)
	notesContent := `member_id,created_at,updated_at,private_note
`
	notesPath := filepath.Join(testDataDir, "private_notes.txt")
	err = os.WriteFile(notesPath, []byte(notesContent), 0644)
	assert.NoError(t, err)

	// Run generate command for XLSX only
	gen := &GenerateCmd{
		DataDir:   testDataDir,
		OutputDir: outputDir,
		Basename:  "test-output",
		Format:    "xlsx",
	}

	err = gen.Run(&Options{})
	assert.NoError(t, err)

	// Verify XLSX was created
	xlsxPath := filepath.Join(outputDir, "test-output.xlsx")
	_, err = os.Stat(xlsxPath)
	assert.NoError(t, err)

	// Verify CSV was NOT created
	csvPath := filepath.Join(outputDir, "test-output.csv")
	_, err = os.Stat(csvPath)
	assert.True(t, os.IsNotExist(err), "CSV file should not exist")
}

func TestGenerateCmd_Run_Both(t *testing.T) {
	// Create test data directory
	testDataDir := t.TempDir()
	outputDir := t.TempDir()

	// Create blockeds.txt
	blockedsContent := `user_id,created_at,updated_at,nickname
123,2024-01-01,2024-01-01,TestUser
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644)
	assert.NoError(t, err)

	// Create private_notes.txt
	notesContent := `member_id,created_at,updated_at,private_note
456,2024-01-02,2024-01-02,Has note
`
	notesPath := filepath.Join(testDataDir, "private_notes.txt")
	err = os.WriteFile(notesPath, []byte(notesContent), 0644)
	assert.NoError(t, err)

	// Run generate command for both formats
	gen := &GenerateCmd{
		DataDir:   testDataDir,
		OutputDir: outputDir,
		Basename:  "test-output",
		Format:    "both",
	}

	err = gen.Run(&Options{})
	assert.NoError(t, err)

	// Verify both files were created
	csvPath := filepath.Join(outputDir, "test-output.csv")
	_, err = os.Stat(csvPath)
	assert.NoError(t, err)

	xlsxPath := filepath.Join(outputDir, "test-output.xlsx")
	_, err = os.Stat(xlsxPath)
	assert.NoError(t, err)
}

func TestGenerateCmd_Run_MissingFiles(t *testing.T) {
	testDataDir := t.TempDir()
	outputDir := t.TempDir()

	gen := &GenerateCmd{
		DataDir:   testDataDir,
		OutputDir: outputDir,
		Basename:  "test-output",
		Format:    "csv",
	}

	// Run without creating input files - should error
	err := gen.Run(&Options{})
	assert.Error(t, err)
}

func TestGenerateCmd_Run_EmptyData(t *testing.T) {
	testDataDir := t.TempDir()
	outputDir := t.TempDir()

	// Create empty blockeds.txt
	blockedsContent := `user_id,created_at,updated_at,nickname
`
	blockedsPath := filepath.Join(testDataDir, "blockeds.txt")
	err := os.WriteFile(blockedsPath, []byte(blockedsContent), 0644)
	assert.NoError(t, err)

	// Create empty private_notes.txt
	notesContent := `member_id,created_at,updated_at,private_note
`
	notesPath := filepath.Join(testDataDir, "private_notes.txt")
	err = os.WriteFile(notesPath, []byte(notesContent), 0644)
	assert.NoError(t, err)

	gen := &GenerateCmd{
		DataDir:   testDataDir,
		OutputDir: outputDir,
		Basename:  "test-output",
		Format:    "csv",
	}

	err = gen.Run(&Options{})
	assert.NoError(t, err)

	// Verify CSV was created even with no data
	csvPath := filepath.Join(outputDir, "test-output.csv")
	_, err = os.Stat(csvPath)
	assert.NoError(t, err)

	// Verify it has only headers
	file, err := os.Open(csvPath)
	assert.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 1, "Should only have header row")
}
