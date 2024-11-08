# sdl-go

Personal CLI app to get movie subtitles from [subdl](https://www.subdl.com)

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
`````

## credits
- [goptn](https://github.com/razsteinmetz/go-ptn) to parse torrent file names