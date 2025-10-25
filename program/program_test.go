package program

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"
)

func TestVersionCmd(t *testing.T) {
	var program Options

	// Parse the version command
	ctx, err := program.Parse([]string{"version"})
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// Capture stdout when running version command
	out := capturer.CaptureStdout(func() {
		err = ctx.Run(&program)
		assert.NoError(t, err)
	})

	// Verify version output
	assert.Equal(t, "unknown\n", out)
	assert.Equal(t, "version", ctx.Command())
}

func TestListCmd_Run(t *testing.T) {
	// Get the path to the example vault
	vaultPath, err := filepath.Abs("../example/vault")
	if err != nil {
		t.Fatalf("Failed to get vault path: %v", err)
	}

	var program Options

	// Parse the list command with vault path
	ctx, err := program.Parse([]string{"obsidian", "--vault", vaultPath, "list"})
	assert.NoError(t, err)

	// Capture stdout
	out := capturer.CaptureStdout(func() {
		err = ctx.Run(&program)
		assert.NoError(t, err)
	})

	// Verify output contains people from the People folder
	assert.Contains(t, out, "Person: Alice")
	assert.Contains(t, out, "Person: Bob")
	assert.Contains(t, out, "Person: Carol")
	assert.Contains(t, out, "Person: David")
	assert.Contains(t, out, "Person: Emma")

	// Verify it contains URLs
	assert.Contains(t, out, "URL: https://fetlife.com/users/12345")
	assert.Contains(t, out, "URL: https://fetlife.com/users/23456")

	// Verify it doesn't list people from Bad People folder
	assert.NotContains(t, out, "Person: Frank")
	assert.NotContains(t, out, "Person: George")
}

func TestListCmd_EmptyVault(t *testing.T) {
	// Create a temporary empty vault
	tempDir := t.TempDir()

	// Create .obsidian directory to make it a valid vault
	err := os.Mkdir(filepath.Join(tempDir, ".obsidian"), 0755)
	assert.NoError(t, err)

	var program Options

	// Parse the list command with vault flag
	ctx, err := program.Parse([]string{"obsidian", "--vault", tempDir, "list"})
	assert.NoError(t, err)

	// Run the command - should not error on empty vault
	out := capturer.CaptureStdout(func() {
		err = ctx.Run(&program)
		assert.NoError(t, err)
	})

	// Should not contain any person entries
	assert.NotContains(t, out, "Person:")
}

func TestListCmd_VaultPath(t *testing.T) {
	// Get the path to the example vault
	vaultPath, err := filepath.Abs("../example/vault")
	if err != nil {
		t.Fatalf("Failed to get vault path: %v", err)
	}

	var program Options

	// Test setting vault path via command line flag
	ctx, err := program.Parse([]string{"obsidian", "--vault", vaultPath, "list"})
	assert.NoError(t, err)
	assert.Equal(t, vaultPath, program.Obsidian.Vault)

	// Run should succeed
	out := capturer.CaptureStdout(func() {
		err = ctx.Run(&program)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "Person: Alice")
}

func TestSyncCmd_Parse(t *testing.T) {
	// Create a temporary vault for the test
	tempVault := t.TempDir()

	// Create .obsidian directory to make it a valid vault
	err := os.Mkdir(filepath.Join(tempVault, ".obsidian"), 0755)
	assert.NoError(t, err)

	// Get the path to the example test-data
	dataPath, err := filepath.Abs("../example/test-data")
	if err != nil {
		t.Fatalf("Failed to get data path: %v", err)
	}

	var program Options

	// Parse the sync command with required data-dir flag and vault
	ctx, err := program.Parse([]string{"obsidian", "--vault", tempVault, "sync", "--data-dir", dataPath})
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	// Verify the sync command was selected
	assert.Equal(t, "obsidian sync", ctx.Command())
}

func TestSyncCmd_Run(t *testing.T) {
	// Create a temporary vault to avoid modifying the example vault
	tempVault := t.TempDir()

	// Get the path to the example test-data
	dataPath, err := filepath.Abs("../example/test-data")
	if err != nil {
		t.Fatalf("Failed to get data path: %v", err)
	}

	// Create .obsidian directory to make it a valid vault
	if err := os.Mkdir(filepath.Join(tempVault, ".obsidian"), 0755); err != nil {
		t.Fatalf("Failed to create .obsidian directory: %v", err)
	}

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

	var program Options

	// Parse the sync command
	ctx, err := program.Parse([]string{"obsidian", "--vault", tempVault, "sync", "--data-dir", dataPath})
	assert.NoError(t, err)

	// Run the sync command - should not error
	err = ctx.Run(&program)
	assert.NoError(t, err)

	// Verify that files were created
	peopleDir := filepath.Join(tempVault, "People")
	files, err := os.ReadDir(peopleDir)
	assert.NoError(t, err)
	assert.Greater(t, len(files), 0, "Expected at least one file to be created")
}

func TestProgramDefaults(t *testing.T) {
	var program Options

	// Parse with a subcommand to test defaults (use version since it doesn't require a vault)
	_, err := program.Parse([]string{"version"})
	assert.NoError(t, err)

	// Verify default output format is "auto"
	assert.Equal(t, "auto", program.OutputFormat)

	// Verify debug and quiet are false by default
	assert.False(t, program.Debug)
	assert.False(t, program.Quiet)
}

func TestProgramDebugFlag(t *testing.T) {
	var program Options

	// Parse with debug flag (use version since it doesn't require a vault)
	_, err := program.Parse([]string{"--debug", "version"})
	assert.NoError(t, err)

	assert.True(t, program.Debug)
}

func TestProgramQuietFlag(t *testing.T) {
	var program Options

	// Parse with quiet flag (use version since it doesn't require a vault)
	_, err := program.Parse([]string{"--quiet", "version"})
	assert.NoError(t, err)

	assert.True(t, program.Quiet)
}

func TestProgramOutputFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		valid  bool
	}{
		{"auto format", "auto", true},
		{"terminal format", "terminal", true},
		{"jsonl format", "jsonl", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var program Options

			// Use version command since it doesn't require a vault
			_, err := program.Parse([]string{"--output-format", tt.format, "version"})

			assert.NoError(t, err)
			assert.Equal(t, tt.format, program.OutputFormat)
		})
	}
}

func TestProgramOutputFormatInvalid(t *testing.T) {
	var program Options

	_, err := program.Parse([]string{"--output-format", "invalid", "obsidian", "list"})
	assert.Error(t, err)
	// Kong should reject invalid enum values
	assert.Contains(t, err.Error(), "must be one of")
}
