package program

import (
	"errors"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"
	"github.com/woodysmith1912/fetlife-data-tools/obsidian"
)

type ObsidianCmd struct {
	Vault string  `help:"Path to vault" env:"VAULT_PATH" default:"." type:"existingdir"`
	Sync  SyncCmd `name:"sync" cmd:"" help:"Sync data between Obsidian and remote source"`
	List  ListCmd `name:"list" cmd:"" help:"List data from vault"`
}

func (cmd *ObsidianCmd) Run(options *Options) error {
	return nil
}

func (cmd *ObsidianCmd) AfterApply(ctx *kong.Context) error {

	// Check if the path is actually a vault by looking for the .obsidian directory
	if !obsidian.IsVaultPath(cmd.Vault) {
		log.Error().
			Str("path", cmd.Vault).
			Msg("The specified path is not a valid Obsidian vault (missing .obsidian directory)")
		return errors.New("invalid Obsidian vault path")
	}
	vault := obsidian.NewVault(cmd.Vault)

	err := vault.Load()
	if err != nil {
		log.Error().Err(err).Msg("Error loading vault")
		return err
	}
	log.Info().
		Str("path", vault.Path).
		Int("pageCount", len(vault.Pages)).
		Msg("Loaded vault")

	ctx.Bind(vault)

	return nil
}
