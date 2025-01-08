package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var config Config

func init() {
	config = GetConfig()
}

func main() {
	var (
		language string
		verbose  bool
	)

	app := &cli.App{
		UseShortOptionHandling: true,
		Name:                   "sdl-go",
		Usage:                  "simple CLI application to download movie subtitles from SUBDL",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lang",
				Aliases:     []string{"l"},
				Value:       config.PREFERRED_LANG,
				Usage:       "override selected language from config",
				Destination: &language,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "more detailed information about what's going on",
				Destination: &verbose,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "run",
				Aliases: []string{"r"},
				Usage:   "download subtitle from the current working directory",
				Action: func(ctx *cli.Context) error {
					return GetSubtitles(language, verbose)
				},
			},
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "options for config",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "show current config",
						Action: func(ctx *cli.Context) error {
							return ListConfig()
						},
					},
					{
						Name:  "reset",
						Usage: "reset config",
						Action: func(ctx *cli.Context) error {
							return RemoveConfig()
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
