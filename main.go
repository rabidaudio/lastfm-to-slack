package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/shkh/lastfm-go/lastfm"
)

var LfmApiKey = os.Getenv("LASTFM_API_KEY")
var LfmApiSecret = os.Getenv("LASTFM_API_SECRET")
var SlackClientID = os.Getenv("SLACK_CLIENT_ID")
var SlackClientSecret = os.Getenv("SLACK_CLIENT_SECRET")
var Debug = os.Getenv("DEBUG") == "true"

var LastfmUsername = flag.String("username", "", "The Last.FM account")
var AlbumMode = flag.Bool("album", false, "Rather than showing the current track, focus on the current album")
var SlackIcon = flag.String("icon", "", "The status icon to use for slack. Defaults to :musical_note: for track mode and :cd: for album mode")

// TODO: it would be cool to smart detect if it was an album or a single track which is playing and adjust
// behavior accordingly

const ShortSleep = 45 * time.Second
const LongSleep = 15 * time.Minute

var lfmApi *lastfm.Api
var slackApi *SlackClient

func debugPrintf(format string, v ...interface{}) {
	if !Debug {
		return
	}
	log.Printf(format, v...)
}

func main() {
	flag.Parse()
	if LfmApiKey == "" || LfmApiSecret == "" {
		log.Fatalf("LASTFM_API_KEY and LASTFM_API_SECRET are required")
	}
	if SlackClientID == "" || SlackClientSecret == "" {
		log.Fatalf("SLACK_CLIENT_ID and SLACK_CLIENT_SECRET are required")
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
	lfmApi = lastfm.New(LfmApiKey, LfmApiSecret)
	var err error
	debugPrintf("authenticating with slack")
	slackApi, err = Authenticate(SlackClientID, SlackClientSecret)
	if err != nil {
		log.Fatalf("authenticate with slack: %v", err)
	}

	status := "" // empty string represents not playing
	for {
		newStatus, err := getPlayingStatus()
		if err != nil {
			log.Printf("get playing: %v", err)
			time.Sleep(LongSleep)
			continue
		}
		if newStatus == "" {
			debugPrintf("not playing")
			if status != "" {
				liveStatus, err := slackApi.Status()
				if err != nil {
					log.Printf("check status: %v", err)
					time.Sleep(LongSleep)
					continue
				}
				if liveStatus != "" && liveStatus == status {
					debugPrintf("clear status")
					if err := slackApi.ClearStatus(); err != nil {
						log.Printf("clear status: %v", err)
					}
				}
			}
			time.Sleep(LongSleep)
			continue
		}
		debugPrintf("playing: %v", newStatus)
		if status != newStatus {
			debugPrintf("setting status")
			err = slackApi.SetStatus(*SlackIcon, newStatus)
			if err != nil {
				log.Printf("problem setting status: %v", err)
				time.Sleep(LongSleep)
				continue
			}
			status = newStatus
		}
		time.Sleep(ShortSleep)
	}
}

func getPlayingStatus() (string, error) {
	res, err := lfmApi.User.GetRecentTracksExtended(map[string]interface{}{
		"limit": 1,
		"user":  *LastfmUsername,
	})
	if err != nil {
		return "", fmt.Errorf("get recent tracks: %w", err)
	}
	if len(res.Tracks) == 0 {
		return "", nil
	}
	track := res.Tracks[0]
	if track.NowPlaying != "true" {
		return "", nil
	}
	loved := track.Loved == "true"
	status := GenerateStatus(track.Name, track.Artist.Name, track.Album.Name, loved)
	return status, nil
}
