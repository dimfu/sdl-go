package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	SDL_API_KEY    string `json:"sdl_api_key"`
	PREFERRED_LANG string
	CWD            string
}

func GetConfig() Config {
	config := Config{
		SDL_API_KEY:    "",
		PREFERRED_LANG: "",
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	cfgFilePath := homeDir + "/.config/sdl-go.json"
	cfgFile, err := os.OpenFile(cfgFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer cfgFile.Close()
	parser := json.NewDecoder(cfgFile)
	parser.Decode(&config)

	if config.SDL_API_KEY == "" || config.PREFERRED_LANG == "" {
		reader := bufio.NewReader(os.Stdin)
		if config.SDL_API_KEY == "" {
			fmt.Print("API KEY not assigned, please input your API KEY:\n-> ")
			text, _ := reader.ReadString('\n')
			config.SDL_API_KEY = strings.TrimSpace(text)
		}
		if config.PREFERRED_LANG == "" {
			fmt.Print("Preferred Language not assigned, please input your preferred language:\n-> ")
			text, _ := reader.ReadString('\n')
			config.PREFERRED_LANG = strings.TrimSpace(text)
		}

		cfgFile.Seek(0, 0)  // Reset file pointer to the beginning
		cfgFile.Truncate(0) // Clear existing content

		encoder := json.NewEncoder(cfgFile)
		err = encoder.Encode(config)
		if err != nil {
			log.Fatal("Failed to write config to file:", err)
		}
	}

	return config
}
