package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/razsteinmetz/go-ptn"
)

const (
	SUBDL_URL    = "https://api.subdl.com/api/v1/subtitles"
	DOWNLOAD_URL = "https://dl.subdl.com"

	OMDB_URL = "http://www.omdbapi.com"
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
	Year      *string
	Codec     string
	IMDBID    *string

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

		year := strconv.Itoa(info.Year)
		movie := &Movie{
			Filename: filename,
			Title:    info.Title,
			Lang:     config.PREFERRED_LANG,
			Series:   nil,
			Source:   info.Quality,
			SDID:     nil,
			Year:     &year,
			Codec:    info.Codec,
		}

		// ugly fix if the torrent author setting the source into just `web` because its mixed
		if len(movie.Source) == 0 {
			r, err := regexp.Compile(`(?i)\b((?:PPV\.)?[HP]DTV|(?:HD)?CAM|B[DR]Rip|(?:HD-?)?TS|(?:PPV )?WEB(?:-?DL|Rip)?|HDRip|DVDRip|DVDRIP|CamRip|W[EB]BRip|BluRay|DvDScr|telesync)\b`)
			if err != nil {
				log.Fatal(err)
			}
			newsource := r.FindString(movie.Filename)
			movie.Source = newsource
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
	spinner := CreateSpinnerFromMethods()
	updater := &SpinnerUpdater{S: spinner, Success: 0, Failed: 0}

	// handle spinner cleanup on interrupts
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		<-sigCh

		spinner.StopFailMessage("interrupted")

		// ignoring error intentionally
		_ = spinner.StopFail()

		os.Exit(0)
	}()

	if err := spinner.Start(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)

	// pause spinner to do an "atomic" config update
	if err := spinner.Pause(); err != nil {
		exitf("failed to pause spinner: %v", err)
	}

	spinner.Suffix(" Downloading subtitles")
	spinner.Message("")

	if err := spinner.Unpause(); err != nil {
		exitf("failed to unpause spinner: %v", err)
	}

	// let spinner animation render for a bit
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(len(movies.List))

	go func() {
		for {
			spinner.Message(fmt.Sprintf("Progress: Success: %d, Failed: %d / Total: %d",
				atomic.LoadInt32(&updater.Success),
				atomic.LoadInt32(&updater.Failed),
				len(movies.List)))

			if atomic.LoadInt32(&updater.Success)+atomic.LoadInt32(&updater.Failed) == int32(len(movies.List)) {
				fmt.Println("\nAll tasks completed!")
				break
			}

			time.Sleep(200 * time.Millisecond)
		}
	}()

	for _, movie := range movies.List {
		go func(movie Movie) {
			defer wg.Done()
			err := movie.searchMovie(movies.config.SDL_API_KEY, false)
			if err != nil {
				atomic.AddInt32(&updater.Failed, 1)
				return
			}
			subUrl := movie.selectSubtitle()
			if subUrl == nil {
				atomic.AddInt32(&updater.Failed, 1)
				return
			}
			err = movie.downloadSubtitle(*subUrl)
			if err != nil {
				atomic.AddInt32(&updater.Failed, 1)
				return
			}
			atomic.AddInt32(&updater.Success, 1)
			time.Sleep(100 * time.Millisecond)
		}(movie)
	}

	wg.Wait()

	spinner.Suffix("")
	message := fmt.Sprintf(" Done downloading subtitles: %d succeeded, %d failed", updater.Success, updater.Failed)
	spinner.StopMessage(message)
	spinner.Stop()
}

func (movie Movie) getMoviesFromSDL(api_key string) (*SDLResponse, error) {
	qb := func(q url.Values) {
		t := "movie"
		q.Add("api_key", api_key)
		q.Add("film_name", movie.Title)
		q.Add("languages", movie.Lang)
		q.Add("subs_per_page", "30")

		if movie.SDID != nil {
			q.Add("sd_id", *movie.SDID)
			q.Del("film_name") // delete film_name because we already have the sd_id otherwise its gonna override the result
		}

		if movie.Year != nil {
			q.Add("year", *movie.Year)
		}

		if movie.IMDBID != nil {
			q.Add("imdb_id", *movie.IMDBID)
			q.Del("sd_id")

			q.Del("film_name") // same thing to this, apparently subdl doesnt handle it by itself
		}

		if movie.Series != nil {
			t = "tv"
			q.Add("season_number", strconv.Itoa(movie.Series.Season))
			q.Add("episode_number", strconv.Itoa(movie.Series.Episode))
		}

		q.Add("type", t)
	}

	var res SDLResponse
	if err := MakeGetRequest(SUBDL_URL, qb, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (movie Movie) getIMDB_ID() (*string, error) {
	qb := func(q url.Values) {
		q.Add("apikey", config.OMDB_API_KEY)
		q.Add("t", movie.Title)
		q.Add("year", *movie.Year)
	}

	var res OMDBResponse
	if err := MakeGetRequest(OMDB_URL, qb, &res); err != nil {
		return nil, err
	}

	return &res.ImdbID, nil
}

func (movie *Movie) searchMovie(api_key string, found bool) error {
	sdlResp, err := movie.getMoviesFromSDL(api_key)
	if err != nil {
		return err
	}

	if found {
		movie.AvailableSubtitles = sdlResp.Subtitles
		return nil
	}

	// refetch subdl request again to get the correct subtitle list
	for _, mvResults := range sdlResp.Results {
		if mvResults.Name == movie.Title {
			s := strconv.Itoa(mvResults.SdID)
			movie.SDID = &s // had to use SDID for accuracy
			return movie.searchMovie(api_key, true)
		}
	}

	imdbId, err := movie.getIMDB_ID()
	if err != nil {
		return err
	}

	if len(*imdbId) > 0 {
		movie.IMDBID = imdbId
		return movie.searchMovie(api_key, true)
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
			if (sub.Episode != nil && movie.Series.Episode == *sub.Episode) || strings.Contains(sub.Name, "complete") {
				return &sub.URL
			}
			return &sub.URL
		} else {
			p, err := ptn.Parse(sub.ReleaseName)
			if err != nil {
				continue
			}
			year := strconv.Itoa(p.Year)
			if movie.Title == p.Title && *movie.Year == year {
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
		match := SearchExtension(file.Name)
		if match != ".srt" {
			continue
		}

		subDetails, err := ptn.Parse(file.Name)
		if err != nil {
			return fmt.Errorf("failed to parse file %v. Error: %s", file.Name, err.Error())
		}

		// ensure we pick the correct subtitle if the zip we downloaded is a complete one
		if movie.Series != nil && subDetails.Episode != movie.Series.Episode {
			continue
		}

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
