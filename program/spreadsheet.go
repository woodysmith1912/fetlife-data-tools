package program

type SpreadsheetCmd struct {
	Generate GenerateCmd `name:"generate" cmd:"" help:"Generate spreadsheet from vault data"`
}

func (spreadsheet *SpreadsheetCmd) Run(options *Options) error {
	return nil

}
