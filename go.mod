module github.com/rabidaudio/lastfm-to-slack

go 1.15

require (
	github.com/shkh/lastfm-go v0.0.0-20191215035245-89a801c244e0
	google.golang.org/appengine v1.6.7
	gotest.tools/v3 v3.0.3
)

replace github.com/shkh/lastfm-go => github.com/rabidaudio/lastfm-go v0.0.1-pre
