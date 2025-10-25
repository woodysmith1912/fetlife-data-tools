package program

import (
	"fmt"

	"github.com/woodysmith1912/fetlife-data-tools/obsidian"
)

type ListCmd struct {
	// Possible options for list command
}

func (list *ListCmd) Run(vault *obsidian.Vault) error {

	// Print out all pages by title and URL
	for _, person := range vault.InFolder("People") {
		fmt.Printf("Person: %s\n", person.Title)
		fmt.Printf("  Folder: %s\n", person.Folder)
		if person.Url != "" {
			fmt.Printf("  URL: %s\n", person.Url)
		}
		if len(person.Aliases) > 0 {
			fmt.Printf("  Aliases: %s\n", person.Aliases)
		}
		if len(person.UrlAliases) > 0 {
			fmt.Printf("  URL Aliases: %s\n", person.UrlAliases)
		}
		if person.WebBadgeColor != "" {
			fmt.Printf("  Web Badge Color: %s\n", person.WebBadgeColor)
		}
		if person.WebMessage != "" {
			fmt.Printf("  Web Message: %s\n", person.WebMessage)
		}
	}

	return nil
}
