package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	owner string
	repo string
	ctx context.Context
	client *github.Client
	release *github.RepositoryRelease
)

func fetchLatest() {
	latest, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		log.Println("Error retrieving latest release", err)
		return
	}
	release = latest
}

func serve(w http.ResponseWriter, request *http.Request) {
	if release == nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprint(w, "Unknown release")
		return
	}

	for _, asset := range release.Assets {
		if "/" + *asset.Name == request.RequestURI {
			w.Header().Add("Location", *asset.BrowserDownloadURL)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprint(w, "Asset not found in release ", *release.Name)
}

func main() {
	var fullRepo = flag.String("repo", "", "the repository to redirect releases for, in user/repo format")
	var port = flag.Int("port", 8080, "the port to listen on for HTTP requests")
	flag.Parse()

	if strings.Count(*fullRepo, "/") != 1 {
		log.Println("Repository must be specified in user/repo format")
		return
	}

	repoParts := strings.Split(*fullRepo, "/")
	owner = repoParts[0]
	repo = repoParts[1]

	client = github.NewClient(nil)

	ticker := time.NewTicker(time.Hour)
	go func() {
		fetchLatest()
		for range ticker.C {
			fetchLatest()
		}
	}()


	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		for sig := range c {
			cancel()
			ticker.Stop()
			log.Printf("Received %s, exiting.", sig.String())
			os.Exit(0)
		}
	}()

	log.Printf("Listing on :%d", *port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.HandlerFunc(serve),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Println("Error listening for requests on port ", port, err)
	}
}
