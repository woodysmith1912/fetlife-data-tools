package main

import (
	"context"
	"fmt"
	"os"

	"github.com/woodysmith1912/fetlife-data-tools/program"
	"github.com/rs/zerolog/log"
)

func main() {

	var options program.Options

	kctx, err := options.Parse(os.Args[1:])

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := context.Background()
	kctx.BindTo(ctx, (*context.Context)(nil))

	// This ends up calling options.Run()
	if err := kctx.Run(&options); err != nil {
		log.Err(err).Msg("Program failed")
		os.Exit(1)
	}
}
