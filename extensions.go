package main

import "regexp"

var extensions = []string{
	// Unknown
	".webm",

	// SDTV
	".m4v", ".3gp", ".nsv", ".ty", ".strm", ".rm", ".rmvb", ".m3u", ".ifo",
	".mov", ".qt", ".divx", ".xvid", ".bivx", ".nrg", ".pva", ".wmv", ".asf",
	".asx", ".ogm", ".ogv", ".m2v", ".avi", ".bin", ".dat", ".dvr-ms", ".mpg",
	".mpeg", ".mp4", ".avc", ".vp3", ".svq3", ".nuv", ".viv", ".dv", ".fli",
	".flv", ".wpl",

	// DVD
	".img", ".iso", ".vob",

	// HD
	".mkv", ".mk3d", ".ts", ".wtv",

	// Bluray
	".m2ts",
}

func SearchExtension(filename string) string {
	pattern := regexp.MustCompile(`\.[0-9a-zA-Z]+$`)
	match := pattern.FindString(filename)
	return match
}

func ExtIsAllowed(filename string) bool {
	match := SearchExtension(filename)

	for _, ext := range extensions {
		if ext == match {
			return true
		}
	}

	return false
}
