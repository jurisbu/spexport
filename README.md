# Simple and stupid Spotify playlist exporter

Scratching my own itch to somehow get playlists out of Spotify and into some rudimentary
JSON that can hopefully later be of some use.

Most of heavy lifting is done by wonderful Spotify Web API wrapper package for Go
https://github.com/zmb3/spotify.

## Usage

First build the binary, in project root:

    go build

Or alternatively `go install...`:

    go install github.com/jurisbu/spexport@latest


Once you have binary - acquire Spotify credentials by registering an application at
https://developer.spotify.com/my-applications/.

**Note:** add a redirection URI `http://localhost:8080/callback` for your application.

Set `SPOTIFY_ID` and `SPOTIFY_SECRET` environment variables to created Spotify credentials, then it is jus a matter of:

    spexport > playlists.json

(Probably will want to tidy and pretty-print with `jq . < playlists.json`.)
