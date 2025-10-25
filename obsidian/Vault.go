package obsidian

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Vault struct {
	Path string
	// Pages is a list of all of the pages in the vault
	Pages []*Page
}

// Color is an HTML color code
type Color string

type Page struct {
	// Title of the page, which is the filename without the .md
	Title string
	// Folder is the folder the page is in, relative to the vault root
	Folder string
	// Tags are taken from the `tags` metadata
	Tags []string
	// Aliases are taken from the `aliases` metadata
	Aliases []string
	// Url is taken from the `url` metadata
	Url string
	// UrlAliases are taken from the `url-aliases` metadata
	UrlAliases []string
	// WebBadgeColor is taken from the `web-badge-color` metadata and should be an HTML color code
	WebBadgeColor Color
	// WebMessage is taken from the `web-message` metadata and will be displayed by the Obsidian plugin in the browser
	WebMessage string
	// FilePath is the absolute path to the markdown file
	FilePath string
	// Content is the markdown content (body) of the page, excluding frontmatter
	Content string
}
type Person struct {
	Page
}

func NewVault(path string) *Vault {
	return &Vault{
		Path: path,
	}
}

// Load loads all of the pages in the vault
func (vault *Vault) Load() error {
	// Iterate all of the markdown files in the vault and load them into the vault
	return filepath.WalkDir(vault.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Load the page
		page, err := loadPage(path, vault.Path)
		if err != nil {
			return err
		}

		vault.Pages = append(vault.Pages, page)
		return nil
	})
}

// LoadPage loads a single page from a markdown file (exported for use in other packages)
func LoadPage(filePath string, vaultPath string) (*Page, error) {
	return loadPage(filePath, vaultPath)
}

// loadPage loads a single page from a markdown file
func loadPage(filePath string, vaultPath string) (*Page, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter
	page := &Page{FilePath: filePath}
	contentStr := string(content)

	// Check if file has frontmatter (starts with ---)
	if strings.HasPrefix(contentStr, "---\n") {
		// Find the end of frontmatter
		endIdx := strings.Index(contentStr[4:], "---\n")
		if endIdx != -1 {
			frontmatter := contentStr[4 : endIdx+4]
			// Store the markdown content (everything after the closing ---)
			page.Content = contentStr[endIdx+8:]

			// Parse YAML frontmatter
			var metadata map[string]interface{}
			if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
				return nil, err
			}

			// Extract metadata fields
			if tags, ok := metadata["tags"].([]interface{}); ok {
				for _, tag := range tags {
					if tagStr, ok := tag.(string); ok {
						page.Tags = append(page.Tags, tagStr)
					}
				}
			}

			if aliases, ok := metadata["aliases"].([]interface{}); ok {
				for _, alias := range aliases {
					if aliasStr, ok := alias.(string); ok {
						page.Aliases = append(page.Aliases, aliasStr)
					}
				}
			}

			if url, ok := metadata["url"].(string); ok {
				page.Url = url
			}

			if urlAliases, ok := metadata["url-aliases"].([]interface{}); ok {
				for _, urlAlias := range urlAliases {
					if urlAliasStr, ok := urlAlias.(string); ok {
						page.UrlAliases = append(page.UrlAliases, urlAliasStr)
					}
				}
			}

			if webBadgeColor, ok := metadata["web-badge-color"].(string); ok {
				page.WebBadgeColor = Color(webBadgeColor)
			}

			if webMessage, ok := metadata["web-message"].(string); ok {
				page.WebMessage = webMessage
			}
		}
	} else {
		// No frontmatter, store entire content
		page.Content = contentStr
	}

	// Extract title from filename (without .md extension)
	filename := filepath.Base(filePath)
	page.Title = strings.TrimSuffix(filename, ".md")

	// Extract folder relative to vault root
	relPath, err := filepath.Rel(vaultPath, filePath)
	if err != nil {
		return nil, err
	}
	page.Folder = filepath.Dir(relPath)

	return page, nil
}

// Save writes the page back to disk with updated metadata
func (page *Page) Save() error {
	// Build metadata map
	metadata := make(map[string]interface{})

	// Add fields to metadata if they have values
	if len(page.Tags) > 0 {
		metadata["tags"] = page.Tags
	}

	if len(page.Aliases) > 0 {
		metadata["aliases"] = page.Aliases
	}

	if page.Url != "" {
		metadata["url"] = page.Url
	}

	if len(page.UrlAliases) > 0 {
		metadata["url-aliases"] = page.UrlAliases
	}

	if page.WebBadgeColor != "" {
		metadata["web-badge-color"] = string(page.WebBadgeColor)
	}

	if page.WebMessage != "" {
		metadata["web-message"] = page.WebMessage
	}

	// Serialize metadata to YAML
	var fileContent strings.Builder

	if len(metadata) > 0 {
		yamlData, err := yaml.Marshal(metadata)
		if err != nil {
			return err
		}

		// Write frontmatter
		fileContent.WriteString("---\n")
		fileContent.Write(yamlData)
		fileContent.WriteString("---\n")
	}

	// Write content (should start with newline if there's frontmatter)
	fileContent.WriteString(page.Content)

	// Write to file
	return os.WriteFile(page.FilePath, []byte(fileContent.String()), 0644)
}

func (vault *Vault) InFolder(folder string) []*Page {
	if folder == "" {
		folder = "."
	}

	var pages []*Page
	for _, page := range vault.Pages {
		if page.Folder == folder {
			pages = append(pages, page)
		}
	}

	return pages
}

func (vault *Vault) WithTag(tag string) []*Page {
	var pages []*Page
	for _, page := range vault.Pages {
		for _, t := range page.Tags {
			if t == tag {
				pages = append(pages, page)
				break
			}
		}
	}
	return pages
}

// IsVaultPath checks if the given path is a valid Obsidian vault by looking for the .obsidian directory
func IsVaultPath(vault string) bool {
	info, err := os.Stat(filepath.Join(vault, ".obsidian"))
	return err == nil && info.IsDir()
}
