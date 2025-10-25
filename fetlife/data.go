package fetlife

import (
	"encoding/csv"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// BlockedRecord represents a blocked user entry from blockeds.txt
type BlockedRecord struct {
	UserID    string
	CreatedAt string
	UpdatedAt string
	Nickname  string
}

// PrivateNoteRecord represents a private note entry from private_notes.txt
type PrivateNoteRecord struct {
	MemberID    string
	CreatedAt   string
	UpdatedAt   string
	PrivateNote string
}

// ReadBlockeds reads and parses the blockeds.txt file from the specified data directory
func ReadBlockeds(dataDir string) ([]BlockedRecord, error) {
	path := filepath.Join(dataDir, "blockeds.txt")
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

// ReadPrivateNotes reads and parses the private_notes.txt file from the specified data directory
func ReadPrivateNotes(dataDir string) ([]PrivateNoteRecord, error) {
	path := filepath.Join(dataDir, "private_notes.txt")
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
