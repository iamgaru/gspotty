package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iamgaru/gspotty/internal/cli"
	"github.com/iamgaru/gspotty/internal/menu"
	"github.com/iamgaru/gspotty/internal/profile"
	"golang.org/x/net/context"
)

// checkEnvironmentVariables verifies that required Spotify API credentials are set
func checkEnvironmentVariables() {
	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")

	if clientID == "" || clientSecret == "" {
		fmt.Println("=================================================================")
		fmt.Println("ERROR: Spotify API credentials not properly configured")
		fmt.Println("=================================================================")

		if clientID == "" {
			fmt.Println("Missing SPOTIFY_ID environment variable")
		}

		if clientSecret == "" {
			fmt.Println("Missing SPOTIFY_SECRET environment variable")
		}

		fmt.Println("\nTo set up your credentials:")
		fmt.Println("1. Go to https://developer.spotify.com/dashboard/")
		fmt.Println("2. Log in and create a new app")
		fmt.Println("3. Set the redirect URI to http://localhost:8888/callback in your app settings")
		fmt.Println("4. Set these environment variables with your credentials:")
		fmt.Println("   export SPOTIFY_ID=your_client_id")
		fmt.Println("   export SPOTIFY_SECRET=your_client_secret")
		fmt.Println("=================================================================")
		os.Exit(1)
	}
}

func main() {
	// Check environment variables first
	checkEnvironmentVariables()

	// Define command line flags
	var (
		searchType   = flag.String("t", "track", "Type of search: track, album, or playlist")
		searchQuery  = flag.String("q", "", "Search query")
		artistName   = flag.String("a", "", "Artist name to filter results (only for track search)")
		limit        = flag.Int("l", 5, "Number of results to display")
		showDetails  = flag.Bool("d", false, "Show detailed information about the results")
		interactive  = flag.Bool("i", false, "Run in interactive mode with a menu interface")
		returnToMenu = flag.Bool("r", false, "Return to interactive menu after viewing search results")
		keepPlaying  = flag.Bool("k", false, "Keep music playing when exiting the player interface")
		autoPlay     = flag.Bool("p", false, "Automatically play the first result and exit")
		stopPlayback = flag.Bool("s", false, "Stop the currently playing track")
		userID       = flag.String("u", "", "Spotify user ID to look up profile information")
	)

	// Add long flag alternatives (kept for backward compatibility but not documented)
	flag.StringVar(searchType, "type", "track", "")
	flag.StringVar(searchQuery, "query", "", "")
	flag.StringVar(artistName, "artist", "", "")
	flag.IntVar(limit, "limit", 5, "")
	flag.BoolVar(showDetails, "details", false, "")
	flag.BoolVar(interactive, "interactive", false, "")
	flag.BoolVar(returnToMenu, "return-to-menu", false, "")
	flag.BoolVar(keepPlaying, "keep-playing", false, "")
	flag.BoolVar(autoPlay, "auto-play", false, "")
	flag.BoolVar(stopPlayback, "stop", false, "")
	flag.StringVar(userID, "user", "", "")

	// Define usage information
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "A CLI tool to search and play Spotify tracks, albums, and playlists.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")

		// Only print flags that have descriptions (the single-letter flags)
		flag.VisitAll(func(f *flag.Flag) {
			if f.Usage != "" {
				fmt.Fprintf(os.Stderr, "  -%s", f.Name)
				name, usage := flag.UnquoteUsage(f)
				if len(name) > 0 {
					fmt.Fprintf(os.Stderr, " %s", name)
				}
				// Boolean flags of one ASCII letter are so common we
				// treat them specially, putting their usage on the same line.
				if len(f.Name) <= 1 && f.DefValue == "false" {
					fmt.Fprintf(os.Stderr, "\t%s", usage)
				} else {
					fmt.Fprintf(os.Stderr, "\n\t%s", usage)
					if f.DefValue != "" {
						if f.DefValue != "0" && f.DefValue != "false" {
							fmt.Fprintf(os.Stderr, " (default %v)", f.DefValue)
						}
					}
				}
				fmt.Fprint(os.Stderr, "\n")
			}
		})

		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -t track -q \"Bohemian Rhapsody\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t track -q \"Bohemian Rhapsody\" -a \"Queen\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t album -q \"Dark Side of the Moon\" -l 3\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -t playlist -q \"workout\" -d\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -i\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -q \"Bohemian Rhapsody\" -r\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -q \"Bohemian Rhapsody\" -k\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -q \"Bohemian Rhapsody\" -p\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -u spotify\n", os.Args[0])
	}

	flag.Parse()

	// Initialize Spotify client
	ctx := context.Background()
	client := cli.GetSpotifyClient(ctx)

	// Check if user profile lookup is requested
	if *userID != "" {
		profile.GetProfile(*userID)
		return
	}

	// Check if stop playback is requested
	if *stopPlayback {
		cli.StopCurrentlyPlaying(ctx, client)
		return
	}

	// Check if interactive mode is enabled
	if *interactive {
		// Start interactive menu
		interactiveMenu := menu.NewInteractiveMenu(ctx, client)
		interactiveMenu.SetKeepPlayingFlag(*keepPlaying)
		if err := interactiveMenu.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running interactive menu: %v\n", err)
		}
		return
	}

	// Validate search type for non-interactive mode
	validTypes := map[string]bool{
		"track":    true,
		"album":    true,
		"playlist": true,
	}

	if !validTypes[*searchType] {
		fmt.Fprintf(os.Stderr, "Error: invalid search type '%s'. Must be one of: track, album, playlist\n", *searchType)
		flag.Usage()
		return
	}

	// Validate search query for non-interactive mode
	if *searchQuery == "" {
		fmt.Fprintf(os.Stderr, "Error: missing search query\n")
		flag.Usage()
		return
	}

	// Perform search based on type and returnToMenu flag
	if *returnToMenu {
		switch *searchType {
		case "track":
			cli.SearchTracksWithMenu(ctx, client, *searchQuery, *artistName, *limit, *showDetails, *keepPlaying, *autoPlay)
		case "album":
			cli.SearchAlbumsWithMenu(ctx, client, *searchQuery, *limit, *showDetails, *keepPlaying, *autoPlay)
		case "playlist":
			cli.SearchPlaylistsWithMenu(ctx, client, *searchQuery, *limit, *showDetails, *keepPlaying, *autoPlay)
		}
	} else {
		switch *searchType {
		case "track":
			cli.SearchTracks(ctx, client, *searchQuery, *artistName, *limit, *showDetails, *keepPlaying, *autoPlay)
		case "album":
			cli.SearchAlbums(ctx, client, *searchQuery, *limit, *showDetails, *keepPlaying, *autoPlay)
		case "playlist":
			cli.SearchPlaylists(ctx, client, *searchQuery, *limit, *showDetails, *keepPlaying, *autoPlay)
		}
	}
}
