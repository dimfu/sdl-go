package main

import (
	"log"
	"os"
)

func GetSubtitles() {
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
	os.Exit(0)
}
