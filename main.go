// This is a a quick and dirty program to export user's playlists from Spotify.
// Heavily based on examples from https://github.com/zmb3/spotify/tree/master/examples.
//
// In order for this program to work one needs to:
//
//  1. Register an application at: https://developer.spotify.com/my-applications/
//     - Use "http://localhost:8080/callback" as the redirect URI
//  2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
//  3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
//
//
// Example usage:
//
//	export SPOTIFY_ID=xxxxx
//  export SPOTIFY_SECRET=xxxxx
//
//  spexport > playlists.json

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"github.com/zmb3/spotify/v2"
)

// redirURI is the OAuth redirect URI for the application.
// It must be registered in Spotify's developer portal for the application.
const redirURI = "http://localhost:8080/callback"

var (
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirURI),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistReadPrivate),
	)
	clientCh = make(chan *spotify.Client)
	state    = "myexporter"
)

func main() {
	ctx := context.Background()

	// HTTP server that will handle OAuth redirect callback.
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(_ http.ResponseWriter, r *http.Request) {
		log.Println("Unexpected request to:", r.URL.String())
	})
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	log.Println("If browser window does not open, then visit the following site to complete auth:", url)
	openbrowser(url)

	// wait for auth to complete
	client := <-clientCh

	// Now that we have client, we can get the list of all playlists.
	playlistPage, err := client.CurrentUsersPlaylists(ctx)
	dieOn(err)

	// Map playlist IDs to names for later use.
	playlistIDs := make(map[spotify.ID]string)

	for _, p := range playlistPage.Playlists {
		playlistIDs[p.ID] = p.Name
	}

	playlists := []Playlist{}

	for id, name := range playlistIDs {
		item, err := client.GetPlaylistItems(ctx, id, spotify.Market("ES"))
		dieOn(err)

		pl := Playlist{Name: name}

		for _, v := range item.Items {
			// First artist from the list should be good.
			artistName := v.Track.Track.Artists[0].Name
			albumName := v.Track.Track.Album.Name
			trackName := v.Track.Track.Name
			releaseData := v.Track.Track.Album.ReleaseDate

			t := Track{
				Name:        trackName,
				ArtistName:  artistName,
				AlbumName:   albumName,
				ReleaseDate: releaseData,
			}
			pl.Tracks = append(pl.Tracks, t)
		}

		playlists = append(playlists, pl)
	}

	// Marshal as JSON to STDOUT.
	if err := json.NewEncoder(os.Stdout).Encode(playlists); err != nil {
		log.Panic(err)
	}
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed!")
	clientCh <- client
}

func dieOn(e error) {
	if e != nil {
		log.Panicf("ERROR: %s", e)
	}
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	dieOn(err)
}

type Playlist struct {
	Name   string
	Tracks []Track
}

type Track struct {
	Name        string
	ArtistName  string
	AlbumName   string
	ReleaseDate string
}
