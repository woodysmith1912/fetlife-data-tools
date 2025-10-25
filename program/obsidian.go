package program

type ObsidianCmd struct {
	Sync SyncCmd `name:"sync" cmd:"" help:"Sync data between Obsidian and remote source"`
	List ListCmd `name:"list" cmd:"" help:"List data from vault"`
}

func (obsidian *ObsidianCmd) Run(options *Options) error {
	return nil
}
