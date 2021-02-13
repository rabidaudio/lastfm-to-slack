Set your current Last.FM playing song as your Slack status.

To run it you'll need to make a [Slack app](https://api.slack.com/apps) and a [Last.FM app](https://www.last.fm/api).

# TODO

- [ ] smart detect album vs track mode
- [ ] web UI for setting lastfm user and settings
- [ ] deploy to a server (ideally which works across slack installs?)
- [ ] store user tokens and configs in a database
- [ ] use track length to inform wait times
- [ ] error handling - token revocation, backoff
