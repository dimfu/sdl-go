# sdl-go

Personal CLI app to get movie subtitles from [SUBDL](https://www.subdl.com)

## pre requisites
Before running the app you need to get 2 different API Key from [SUBDL](https://www.subdl.com) and [OMDB](http://www.omdbapi.com/apikey.aspx) by simply do registration. The app primarily fetches movie information using the SUBDL API. However, it currently falls back to the OMDB API when the SUBDL API fails to provide results.

## usage

```bash
# install the dependencies
go mod download

# build the binary
go install .

# make sure to be in the movie directory
cd (mov dir)

# run the thing, it should download all the subtitle for each movie in the directory
sdl-go run

# for more commands run
sdl-go help
`````

## credits
- [goptn](https://github.com/razsteinmetz/go-ptn) to parse torrent file names