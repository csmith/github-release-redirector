package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/csmith/envflag/v2"
	"github.com/google/go-github/v82/github"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	owner       string
	repo        string
	redirect    *string
	webhookPath *string
	ctx         context.Context
	client      *github.Client
	release     *github.RepositoryRelease
	ticker      *time.Ticker
)

func fetchLatest() {
	latest, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Println("Error retrieving latest release", err)
		return
	}
	log.Printf("Found latest release: %s\n", *latest.Name)
	release = latest
}

func temporaryRedirect(w http.ResponseWriter, url string) {
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func serveStaticPaths(w http.ResponseWriter, request *http.Request) bool {
	if "/" == request.RequestURI && len(*redirect) > 0 {
		temporaryRedirect(w, *redirect)
		return true
	}

	if *webhookPath == request.RequestURI {
		log.Println("Received webhook, starting a refresh")
		go func() {
			fetchLatest()
		}()
		w.WriteHeader(http.StatusOK)
		return true
	}

	return false
}

func serveAssets(w http.ResponseWriter, request *http.Request) bool {
	if release == nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Unknown release")
		return true
	}

	for _, asset := range release.Assets {
		if "/"+*asset.Name == request.RequestURI {
			temporaryRedirect(w, *asset.BrowserDownloadURL)
			return true
		}
	}

	return false
}

func serve(w http.ResponseWriter, request *http.Request) {
	if !serveStaticPaths(w, request) && !serveAssets(w, request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, "Asset not found in release ", *release.Name)
	}
}

func parseRepo(fullRepo *string) error {
	if len(*fullRepo) == 0 {
		return fmt.Errorf("the repository option must be specified")
	}
	if strings.Count(*fullRepo, "/") != 1 {
		return fmt.Errorf("the repository must be specified in `user/repo` format")
	}
	repoParts := strings.Split(*fullRepo, "/")
	owner = repoParts[0]
	repo = repoParts[1]
	return nil
}

func initTicker(seconds int) {
	if seconds > 0 {
		log.Printf("Starting ticker for polling once every %d seconds.\n", seconds)
		ticker = time.NewTicker(time.Duration(seconds) * time.Second)
		go func() {
			for range ticker.C {
				fetchLatest()
			}
		}()
	} else {
		log.Println("Not starting ticker; performing one off fetch and relying on webhooks.")
	}
	fetchLatest()
}

func main() {
	redirect = flag.String("redirect", "", "if specified, requests for / will be redirected to this url")
	webhookPath = flag.String("webhook", "", "full path to receive release webhooks from GitHub on")
	var fullRepo = flag.String("repo", "", "the repository to redirect releases for, in user/repo format [required]")
	var port = flag.Int("port", 8080, "the port to listen on for HTTP requests")
	var poll = flag.Int("poll", 3600, "the amount of time to wait between polling for releases; 0 to disable polling")

	envflag.Parse()

	if err := parseRepo(fullRepo); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		flag.Usage()
		return
	}

	client = github.NewClient(nil)

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		for sig := range c {
			cancel()
			if ticker != nil {
				ticker.Stop()
			}
			log.Printf("Received %s, exiting.\n", sig.String())
			os.Exit(0)
		}
	}()

	initTicker(*poll)

	log.Printf("Listing on :%d\n", *port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.HandlerFunc(serve),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Println("Error listening for requests on port ", port, err)
	}
}
