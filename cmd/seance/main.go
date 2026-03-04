package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/aschepis/seance/internal/api"
	"github.com/aschepis/seance/internal/conversations"
	"github.com/aschepis/seance/web"
)

func main() {
	port := flag.Int("port", 3333, "port to listen on")
	claudeDir := flag.String("claude-dir", conversations.DefaultClaudeDir(), "path to Claude config directory")
	flag.Parse()

	if *claudeDir == "" {
		fmt.Fprintln(os.Stderr, "error: could not determine Claude config directory")
		os.Exit(1)
	}

	parser := conversations.NewParser(*claudeDir)
	handler := api.NewHandler(parser)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Serve embedded frontend
	frontendFS, err := fs.Sub(web.FrontendFS, "dist")
	if err != nil {
		log.Fatalf("error: could not load frontend: %v", err)
	}
	fileServer := http.FileServer(http.FS(frontendFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file; if it doesn't exist, serve index.html for SPA routing
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if _, err := fs.Stat(frontendFS, path[1:]); err != nil {
			// Serve index.html for SPA routes
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("seance listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("error: %v", err)
	}
}
