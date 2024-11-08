package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	SDL_API_KEY = "SDL_API_KEY"
)

var (
	command string
	movies  []string
	config  Config
)

func printHelp() {
	fmt.Println("Usage:")
}

func init() {
	flag.Parse()
	flag.Usage = printHelp

	args := make([]string, 0)
	for i := len(os.Args) - len(flag.Args()) + 1; i < len(os.Args); {
		if i > 1 && os.Args[i-2] == "--" {
			break
		}
		args = append(args, flag.Arg(0))
		if err := flag.CommandLine.Parse(os.Args[i:]); err != nil {
			log.Fatal("error while parsing arguments")
		}

		i += 1 + len(os.Args[i:]) - len(flag.Args())
	}
	args = append(args, flag.Args()...)

	if len(args) < 1 {
		flag.Usage()
		os.Exit(0)
	}

	command = args[0]

	switch command {
	case "run":
		config = GetConfig()
		// noop
	case "help":
		flag.Usage()
	default:
		log.Fatal("command not found, please refer to `help` command")
		os.Exit(0)
	}
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	config.CWD = cwd

	files, err := os.ReadDir(cwd)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if !file.IsDir() && ExtIsAllowed(file.Name()) {
			movies = append(movies, file.Name())
		}
	}

	parsed, err := NewMovies(movies, config)
	if err != nil {
		log.Fatal(err)
	}

	parsed.GetSubtitles()
}
