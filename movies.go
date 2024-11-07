package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"

	"github.com/razsteinmetz/go-ptn"
)

const (
	SUBDL_URL     = "https://api.subdl.com/api/v1/subtitles"
	DOWNLOAD_LINK = "https://dl.subdl.com"
)

type series struct {
	Season  int
	Episode int
}

type Movie struct {
	Filename string
	Title    string
	Lang     string
	Series   *series
	Source   string

	AvailableSubtitles []Subtitle
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

		movie := &Movie{
			Filename: filename,
			Title:    info.Title,
			Lang:     config.PREFERRED_LANG,
			Series:   nil,
			Source:   info.Quality,
		}

		if info.Season > 0 && info.Episode > 0 {
			movie.Series = &series{
				Season:  info.Season,
				Episode: info.Episode,
			}
		}

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
	wg := sync.WaitGroup{}
	wg.Add(len(movies.List))

	for _, movie := range movies.List {
		go func(movie Movie) {
			defer wg.Done()
			err := movie.searchMovie(movies.config.SDL_API_KEY)
			if err != nil {
				log.Printf("cannot get movie, ERR: %s", err.Error())
				return
			}
			subUrl := movie.selectSubtitle()
			if subUrl == nil {
				// TODO: better error message
				log.Printf("cannot get subtitle for this movie")
				return
			}
			fmt.Println(*subUrl)
		}(movie)
	}

	wg.Wait()
}

func (movie *Movie) searchMovie(api_key string) error {
	req, err := http.NewRequest(http.MethodGet, SUBDL_URL, nil)
	if err != nil {
		log.Print(err.Error())
	}

	q := req.URL.Query()

	// TODO: add full season flag to just search download url that provides full season episodes

	q.Add("api_key", api_key)
	q.Add("film_name", movie.Title)
	q.Add("languages", movie.Lang)

	if movie.Series != nil {
		q.Add("season_number", strconv.Itoa(movie.Series.Season))
		q.Add("episode_number", strconv.Itoa(movie.Series.Episode))
	}

	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()

	var sdlResp Response
	if err = json.NewDecoder(resp.Body).Decode(&sdlResp); err != nil {
		return errors.New(err.Error())
	}

	movie.AvailableSubtitles = sdlResp.Subtitles
	return nil
}

func (movie Movie) selectSubtitle() *string {
	if len(movie.Source) == 0 {
		// TODO: if movie source empty, manually select subtitles
		return nil
	}

	var matchedSubs = []Subtitle{}

	// ? get the matched source so it wont mess up when the subtitle is being used
	re, _ := regexp.Compile("(?)" + movie.Source)
	for _, sub := range movie.AvailableSubtitles {
		if re.MatchString(movie.Filename) {
			matchedSubs = append(matchedSubs, sub)
		}
	}

	if movie.Series != nil {
		// ? TODO: maybe increment subtitle limit 10 by 10 if not found, search until the very last subtitle
		for _, sub := range matchedSubs {
			if movie.Series.Season == sub.Season {
				if sub.Episode != nil && movie.Series.Episode == *sub.Episode {
					return &sub.URL
				} else {
					// ? TODO: maybe search in filename string if it contains S00E00 format, if it matched then return that subtitle
					continue
				}
			}
		}
	}

	return nil
}
