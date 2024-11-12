package main

import (
	"embed"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"strings"
)

//go:embed data/lang.json
var content embed.FS

func overrideLang(lang string) (*string, error) {
	f, err := content.Open("data/lang.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	langs := make(map[string]string)
	bytesVal, _ := io.ReadAll(f)

	err = json.Unmarshal([]byte(bytesVal), &langs)
	if err != nil {
		return nil, errors.New("error parsing json: " + err.Error())
	}

	if _, ok := langs[strings.ToUpper(lang)]; ok {
		return &lang, nil
	}

	return nil, errors.New("language not found")
}

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

	// handle lang flag if not empty
	if len(*lang) > 0 {
		ol, err := overrideLang(*lang)
		if err != nil {
			log.Fatal(err.Error())
		}
		config.PREFERRED_LANG = *ol
	}

	parsed, err := NewMovies(movies, config)
	if err != nil {
		log.Fatal(err)
	}

	parsed.GetSubtitles()
	os.Exit(0)
}
