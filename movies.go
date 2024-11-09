package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/razsteinmetz/go-ptn"
)

const (
	SUBDL_URL    = "https://api.subdl.com/api/v1/subtitles"
	DOWNLOAD_URL = "https://dl.subdl.com"
)

type series struct {
	Season  int
	Episode int
}

type Movie struct {
	Filename  string
	Title     string
	Lang      string
	Series    *series
	Source    string
	Extension string
	SDID      *string
	Year      string
	Codec     string

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
			SDID:     nil,
			Year:     strconv.Itoa(info.Year),
			Codec:    info.Codec,
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
			err := movie.searchMovie(movies.config.SDL_API_KEY, false)
			if err != nil {
				log.Printf("cannot get movie: %s", err.Error())
				return
			}
			subUrl := movie.selectSubtitle()
			if subUrl == nil {
				// TODO: better error message
				log.Printf("cannot get subtitle for this movie\n\n")
				return
			}
			err = movie.downloadSubtitle(*subUrl)
			if err != nil {
				log.Print(err.Error())
				return
			}
		}(movie)
	}

	wg.Wait()
}

func (movie *Movie) searchMovie(api_key string, found bool) error {
	req, err := http.NewRequest(http.MethodGet, SUBDL_URL, nil)
	if err != nil {
		log.Print(err.Error())
	}

	// is a movie or a tv series, default to movie
	t := "movie"

	q := req.URL.Query()

	// TODO: add full season flag to just search download url that provides full season episodes

	q.Add("api_key", api_key)
	q.Add("film_name", movie.Title)
	q.Add("languages", movie.Lang)
	q.Add("subs_per_page", "30")

	if movie.SDID != nil {
		q.Add("sd_id", *movie.SDID)
		q.Del("film_name") // delete film_name because we already have the sd_id otherwise its gonna override the result
	}

	if movie.Series != nil {
		t = "tv"
		q.Add("season_number", strconv.Itoa(movie.Series.Season))
		q.Add("episode_number", strconv.Itoa(movie.Series.Episode))
	}

	q.Add("type", t)

	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var sdlResp Response
	if err = json.NewDecoder(resp.Body).Decode(&sdlResp); err != nil {
		return errors.New(err.Error())
	}

	if found {
		movie.AvailableSubtitles = sdlResp.Subtitles
		return nil
	}

	for _, mvResults := range sdlResp.Results {
		if mvResults.Name == movie.Title {
			s := strconv.Itoa(mvResults.SdID)
			movie.SDID = &s
			return movie.searchMovie(api_key, true)
		}
	}

	return errors.New("movie not found")
}

func (movie Movie) selectSubtitle() *string {
	if len(movie.Source) == 0 {
		return nil
	}

	var matchedSubs = []Subtitle{}
	re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(movie.Source))
	if err != nil {
		return nil
	}

	for _, sub := range movie.AvailableSubtitles {
		if re.MatchString(sub.ReleaseName) {
			matchedSubs = append(matchedSubs, sub)
		}
	}

	for _, sub := range matchedSubs {
		if movie.Series != nil && movie.Series.Season == sub.Season {
			if sub.Episode != nil && movie.Series.Episode == *sub.Episode {
				return &sub.URL
			}
			continue
		} else {
			p, err := ptn.Parse(sub.ReleaseName)
			if err != nil {
				continue
			}
			if movie.Title == p.Title && movie.Year == strconv.Itoa(p.Year) {
				if len(p.Codec) > 0 && len(movie.Codec) > 0 && p.Codec != movie.Codec {
					continue
				}
				return &sub.URL
			}
		}
	}

	return nil
}

func (movie Movie) extToSRT() string {
	s := movie.Filename
	match := SearchExtension(movie.Filename)
	r := strings.NewReplacer(match, ".srt")
	s = r.Replace(s)
	return s
}

func (movie Movie) downloadSubtitle(url string) error {
	dl_url := DOWNLOAD_URL + url
	req, err := http.NewRequest(http.MethodGet, dl_url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to download the subtitle")
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New("response is unreadable")
	}

	reader := bytes.NewReader(contents)
	zipReader, err := zip.NewReader(reader, int64(len(contents)))
	if err != nil {
		return fmt.Errorf("failed to read zip file: %v", err)
	}

	for _, file := range zipReader.File {
		f, err := file.Open()
		if err != nil {
			log.Print(err.Error())
			continue
		}
		defer f.Close()

		destPath := filepath.Join(config.CWD, movie.extToSRT())
		destFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create destination file: %v", err.Error())
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, f)
		if err != nil {
			return fmt.Errorf("failed to copy subtitle to %v: %v", destFile, err.Error())
		}
	}

	return nil
}
