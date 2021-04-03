package main

import (
	"github.com/alecthomas/kong"
)

var cli struct {
	Create struct {
		Config createConfigCmd `cmd:"" help:"Create a config"`
	} `cmd:"" help:"Create a resource"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			//Compact: true,
		}),
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
