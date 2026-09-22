package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tvbox-video-source/handlers"
)

func main() {
	port := flag.Int("port", 5000, "server port")
	flag.Parse()

	var rootDir string
	if exePath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "wwwroot")
		if _, err := os.Stat(candidate); err == nil {
			rootDir = candidate
		}
	}
	if rootDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		candidate := filepath.Join(wd, "wwwroot")
		if _, err := os.Stat(candidate); err == nil {
			rootDir = candidate
		}
	}
	if rootDir == "" {
		log.Fatal("wwwroot directory not found (checked exe directory and working directory)")
	}

	api := handlers.NewAPI(rootDir)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		api.ReadConfig(w, r)
	})

	mux.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		api.SaveConfig(w, r)
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		api.UploadConfig(w, r)
	})

	mux.HandleFunc("/page/list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		api.GetPages(w, r)
	})

	mux.HandleFunc("/page/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		api.CreatePage(w, r)
	})

	mux.HandleFunc("/page/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.NotFound(w, r)
			return
		}
		api.DeletePage(w, r)
	})

	// Static file serving: /mod-api/ -> wwwroot/mod-api/ (Swagger UI)
	mux.HandleFunc("/mod-api", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mod-api/", http.StatusMovedPermanently)
	})
	mux.Handle("/mod-api/", http.StripPrefix("/mod-api/",
		http.FileServer(http.Dir(filepath.Join(rootDir, "mod-api")))))

	// Static file serving: /mod-ce/ -> wwwroot/mod-ce/
	mux.Handle("/mod-ce/", http.StripPrefix("/mod-ce/",
		http.FileServer(http.Dir(filepath.Join(rootDir, "mod-ce")))))

	// Static file serving: /raw/ -> wwwroot/uploads/
	mux.Handle("/raw/", http.StripPrefix("/raw/",
		http.FileServer(http.Dir(filepath.Join(rootDir, "uploads")))))

	// Catch-all route for {key} and {file}/{key} redirect patterns
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "/" {
			api.Index(w, r)
			return
		}

		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 2 {
			// /{file}/{key} pattern
			api.GetFileConfig(w, r)
			return
		}

		// Check if it's a known static file or directory
		staticPath := filepath.Join(rootDir, path)
		if _, err := os.Stat(staticPath); err == nil {
			http.FileServer(http.Dir(rootDir)).ServeHTTP(w, r)
			return
		}

		// /{key} pattern - redirect based on default.json
		api.GetDefaultConfig(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
