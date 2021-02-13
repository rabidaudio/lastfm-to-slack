package main

import (
	"flag"
	"log"
	"os"
	"regexp"

	"github.com/shkh/lastfm-go/lastfm"
)

var ApiKey = os.Getenv("LASTFM_API_KEY")
var ApiSecret = os.Getenv("LASTFM_API_SECRET")

var LastfmUsername = flag.String("username", "", "The Last.FM username of the account")
var AlbumMode = flag.Bool("album", false, "Rather than showing the current track, focus on the current album")
var SlackIcon = flag.String("icon", "", "The status icon to use for slack. Defaults to :musical_note: for track mode and :cd: for album mode")

// TODO: it would be cool to smart detect if it was an album or a single track which is playing and adjust
// behavior accordingly

const MaxLength = 100

var api *lastfm.Api

func main() {
	flag.Parse()
	if ApiKey == "" || ApiSecret == "" {
		log.Fatalf("LASTFM_API_KEY and LASTFM_API_SECRET are required")
	}
	if *LastfmUsername == "" {
		flag.PrintDefaults()
		log.Fatalf("-username is required")
	}
	if !regexp.MustCompile("^:.*:$").MatchString(*SlackIcon) {
		*SlackIcon = ":musical_note:"
		if *AlbumMode {
			*SlackIcon = ":cd:"
		}
	}
	api = lastfm.New(ApiKey, ApiSecret)
}

func getPlayingStatus() (string, bool) {
	res, err := api.User.GetRecentTracksExtended(map[string]interface{}{
		"limit": 1,
		"user":  *LastfmUsername,
	})
	if err != nil {
		log.Fatalf("get recent tracks: %v", err)
	}
	if len(res.Tracks) == 0 {
		return "", false
	}
	track := res.Tracks[0]
	if track.NowPlaying != "true" {
		return "", false
	}
	loved := track.Loved == "true"
	status := GenerateStatus(track.Name, track.Artist.Name, track.Album.Name, loved)
	return status, true
}

const Sep = " - "
const Heart = " :heart:"
const Tail = "..."
const Min = 9

func GenerateStatus(title, artist, album string, loved bool) string {
	first := artist
	second := title
	if *AlbumMode {
		second = album
	}
	first = strip(first)
	second = strip(second)
	available := MaxLength - len(Sep)
	if loved {
		available -= len(Heart)
	}
	if len(first)+len(second) > available {
		// first try triming the second and see if it will fit
		if len(first)+Min <= available {
			second = truncate(second, available-len(first))
		} else {
			// we're going to need to trim both
			lf := len(first) * available / (len(first) + len(second))
			ls := available - lf
			// but we need to make sure if were going to truncate both
			// that each is at least Min characters
			if lf < Min {
				lf = Min
				if len(first) < Min {
					lf = len(first)
				}
				ls = available - lf
			} else if ls < Min {
				ls = Min
				if len(second) < Min {
					ls = len(second)
				}
				lf = available - ls
			}
			first = truncate(first, lf)
			second = truncate(second, ls)
		}
	}
	msg := first + Sep + second
	if loved {
		msg = msg + Heart
	}
	return msg
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	base := value[0 : limit-len(Tail)]
	return strip(base) + Tail
}

func strip(value string) string {
	for value[0] == ' ' {
		value = value[1:]
	}
	for value[len(value)-1] == ' ' {
		value = value[:len(value)-1]
	}
	return value
}
