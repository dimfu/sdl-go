package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	SDL_API_KEY    string `json:"sdl_api_key"`
	OMDB_API_KEY   string `json:"omdb_api_key"`
	PREFERRED_LANG string `json:"preferred_lang"`
	CWD            string
}

func configFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("error while getting home dir")
	}
	cfgFilePath := homeDir + "/.config/sdl-go.json"
	return cfgFilePath
}

func ListConfig() error {
	file, err := os.ReadFile(configFilePath())
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	err = json.Indent(&buffer, file, "", "\t")
	if err != nil {
		return err
	}
	fmt.Print(buffer.String())
	return nil
}

func RemoveConfig() error {
	err := os.Remove(configFilePath())
	if err != nil {
		return fmt.Errorf("error deleting config: %v", err.Error())
	}

	log.Println("successfully deleting config")
	return nil
}

func GetConfig() Config {
	config := Config{
		SDL_API_KEY:    "",
		PREFERRED_LANG: "",
	}
	cfgFile, err := os.OpenFile(configFilePath(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer cfgFile.Close()
	parser := json.NewDecoder(cfgFile)
	parser.Decode(&config)

	if config.SDL_API_KEY == "" || config.PREFERRED_LANG == "" || config.OMDB_API_KEY == "" {
		reader := bufio.NewReader(os.Stdin)
		if config.SDL_API_KEY == "" {
			fmt.Print("SDL API KEY not assigned, please input your API KEY:\n-> ")
			text, _ := reader.ReadString('\n')
			config.SDL_API_KEY = strings.TrimSpace(text)
		}
		if config.OMDB_API_KEY == "" {
			fmt.Print("OMDB API KEY not assigned, please input your API KEY:\n-> ")
			text, _ := reader.ReadString('\n')
			config.OMDB_API_KEY = strings.TrimSpace(text)
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
