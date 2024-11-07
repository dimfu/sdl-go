package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/razsteinmetz/go-ptn"
)

const (
	SUBDL_URL = "https://api.subdl.com/api/v1/subtitles"
)

type Movie struct {
	Filename string
	Title    string
	Lang     string
}

type Movies struct {
	List   []Movie
	config Config
}

func NewMovies(movies []string, config Config) (*Movies, error) {
	var list = []Movie{}
	for _, filename := range movies {
		info, err := ptn.Parse(filename)
		if err != nil {
			log.Printf("cannot parse %s\n", filename)
			continue
		}
		movie := &Movie{Filename: filename, Title: info.Title, Lang: config.PREFERRED_LANG}
		list = append(list, *movie)
	}

	if len(list) == 0 {
		return nil, errors.New("no movies present")
	}

	return &Movies{
		List:   list,
		config: config,
	}, nil
}

func (movies Movies) GetSubtitles() {
	for _, movie := range movies.List {
		// TODO: download the subtitle frfr then extract and rename it with the original file name
		movie.downloadSubtitle(movies.config.SDL_API_KEY)
	}
}

func (movie Movie) downloadSubtitle(api_key string) {
	req, err := http.NewRequest(http.MethodGet, SUBDL_URL, nil)
	if err != nil {
		log.Print(err.Error())
	}
	q := req.URL.Query()
	q.Add("api_key", api_key)
	q.Add("film_name", movie.Title)
	q.Add("languages", movie.Lang)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
