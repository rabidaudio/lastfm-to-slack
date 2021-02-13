module github.com/rabidaudio/lastfm-to-slack

go 1.15

require (
	github.com/pkg/browser v0.0.0-20210115035449-ce105d075bb4
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shkh/lastfm-go v0.0.0-20191215035245-89a801c244e0
	github.com/slack-go/slack v0.8.0
	gotest.tools/v3 v3.0.3
)

replace github.com/shkh/lastfm-go => github.com/rabidaudio/lastfm-go v0.0.1-pre
