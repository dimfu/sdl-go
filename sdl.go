package main

import "time"

type SDLResponse struct {
	Status    bool       `json:"status"`
	Results   []Result   `json:"results"`
	Subtitles []Subtitle `json:"subtitles"`
}

type Result struct {
	SdID         int        `json:"sd_id"`
	Type         string     `json:"type"`
	Name         string     `json:"name"`
	ImdbID       string     `json:"imdb_id"`
	TmdbID       int        `json:"tmdb_id"`
	FirstAirDate time.Time  `json:"first_air_date"`
	Slug         string     `json:"slug"`
	ReleaseDate  *time.Time `json:"release_date"`
	Year         int        `json:"year"`
}

type Subtitle struct {
	ReleaseName  string `json:"release_name"`
	Name         string `json:"name"`
	Lang         string `json:"lang"`
	Author       string `json:"author"`
	URL          string `json:"url"`
	SubtitlePage string `json:"subtitlePage"`
	Season       int    `json:"season"`
	Episode      *int   `json:"episode"`
	Language     string `json:"language"`
	Hi           bool   `json:"hi"`
	EpisodeFrom  *int   `json:"episode_from"`
	EpisodeEnd   int    `json:"episode_end"`
	FullSeason   bool   `json:"full_season"`
}
