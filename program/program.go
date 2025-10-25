package program

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/alecthomas/kong"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Options is the structure of program options
type Options struct {
	Debug        bool           `group:"Info" help:"Show debugging information"`
	OutputFormat string         `group:"Info" enum:"auto,jsonl,terminal" default:"auto" help:"How to show program output (auto|terminal|jsonl)"`
	Quiet        bool           `group:"Info" help:"Be less verbose than usual"`
	Version      VersionCmd     `name:"version" cmd:"" help:"Show program version"`
	Obsidian     ObsidianCmd    `name:"obsidian" cmd:"" help:"Obsidian related commands"`
	Spreadsheet  SpreadsheetCmd `name:"spreadsheet" cmd:"" help:"Spreadsheet related commands"`
}

// Parse calls the CLI parsing routines
func (program *Options) Parse(args []string) (*kong.Context, error) {
	parser, err := kong.New(program,
		kong.ShortUsageOnError(),
		// kong.Description("Brief Program Summary"),
	)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return parser.Parse(args)
}

// Run runs the program
func (program *Options) Run(options *Options) error {
	return nil
}

// AfterApply runs after the options are parsed but before anything runs
func (program *Options) AfterApply() error {
	program.initLogging()
	return nil
}

func (program *Options) initLogging() {
	switch {
	case program.Debug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case program.Quiet:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var out io.Writer = os.Stdout

	if os.Getenv("TERM") == "" && runtime.GOOS == "windows" {
		out = colorable.NewColorableStdout()
	}

	if program.OutputFormat == "terminal" ||
		(program.OutputFormat == "auto" && isTerminal(os.Stdout)) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: out})
	} else {
		log.Logger = log.Output(out)
	}

	log.Logger.Debug().
		Str("version", Version).
		Str("program", os.Args[0]).
		Msg("Starting")
}

// isTerminal returns true if the file given points to a character device (i.e. a terminal)
func isTerminal(file *os.File) bool {
	if fileInfo, err := file.Stat(); err != nil {
		log.Err(err).Msg("Error running stat")
		return false
	} else {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
}
