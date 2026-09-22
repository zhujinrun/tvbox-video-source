package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tvbox-video-source/models"
)

type API struct {
	RootDir string
}

func NewAPI(rootDir string) *API {
	return &API{RootDir: rootDir}
}

func (a *API) respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func stripBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
}

func (a *API) readJSON(fileName string) (map[string]string, error) {
	filePath := filepath.Join(a.RootDir, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return make(map[string]string), nil
	}
	var result map[string]string
	err = json.Unmarshal(stripBOM(data), &result)
	return result, err
}

func (a *API) writeJSON(data interface{}, fileName string) error {
	filePath := filepath.Join(a.RootDir, fileName)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, 0644)
}

func (a *API) readPages(fileName string) ([]models.PageModel, error) {
	filePath := filepath.Join(a.RootDir, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return make([]models.PageModel, 0), nil
	}
	var result []models.PageModel
	err = json.Unmarshal(stripBOM(data), &result)
	return result, err
}

func (a *API) writePages(pages []models.PageModel, fileName string) error {
	return a.writeJSON(pages, fileName)
}

// GET /
func (a *API) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "success")
}

// GET /{key} - redirect based on default.json
func (a *API) GetDefaultConfig(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	if key == "" {
		http.NotFound(w, r)
		return
	}
	dict, _ := a.readJSON("default.json")
	if val, ok := dict[key]; ok && val != "" {
		http.Redirect(w, r, val, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/vip", http.StatusFound)
}

// GET /{file}/{key} - redirect based on {file}.json
func (a *API) GetFileConfig(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	file := parts[0]
	key := parts[1]
	dict, _ := a.readJSON(file + ".json")
	if val, ok := dict[key]; ok && val != "" {
		http.Redirect(w, r, val, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/"+file+"/vip", http.StatusFound)
}

// GET /Read?file=default
func (a *API) ReadConfig(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		file = "default"
	}
	dict, err := a.readJSON(file + ".json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.respondJSON(w, http.StatusOK, dict)
}

// POST /Save?file=default
func (a *API) SaveConfig(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" {
		file = "default"
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	config := make(map[string]string)
	for key, values := range r.PostForm {
		if key == "config" {
			continue
		}
		if len(values) > 0 {
			config[key] = values[0]
		}
	}
	if err := a.writeJSON(config, file+".json"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /Upload
func (a *API) UploadConfig(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	storageDir := filepath.Join(a.RootDir, "uploads")
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(filepath.Join(storageDir, handler.Filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fileURL := fmt.Sprintf("%s://%s/raw/%s", scheme, r.Host, handler.Filename)
	a.respondJSON(w, http.StatusOK, fileURL)
}

// GET /Page/List
func (a *API) GetPages(w http.ResponseWriter, r *http.Request) {
	pages, err := a.readPages("pages.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.respondJSON(w, http.StatusOK, pages)
}

// POST /Page/Create
func (a *API) CreatePage(w http.ResponseWriter, r *http.Request) {
	var model models.PageModel
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pages, _ := a.readPages("pages.json")

	// find existing page
	var existing *models.PageModel
	for i, p := range pages {
		if p.App == model.App {
			existing = &pages[i]
			break
		}
	}

	// remove old page if exists
	if existing != nil {
		a.removePageFiles(*existing)
		newPages := make([]models.PageModel, 0, len(pages)-1)
		for _, p := range pages {
			if p.App != model.App {
				newPages = append(newPages, p)
			}
		}
		pages = newPages
	}

	// generate HTML from template
	tplPath := filepath.Join(a.RootDir, "mod-ce", "tpl.html")
	if tplData, err := os.ReadFile(tplPath); err == nil {
		pageCode := strings.ReplaceAll(string(tplData), "{{app}}", model.App)
		pageCode = strings.ReplaceAll(pageCode, "{{file}}", model.File)
		tgtPath := filepath.Join(a.RootDir, "mod-ce", model.Page+".html")
		os.WriteFile(tgtPath, []byte(pageCode), 0644)
	}

	pages = append(pages, model)
	if err := a.writePages(pages, "pages.json"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /Page/Delete?app=xxx
func (a *API) DeletePage(w http.ResponseWriter, r *http.Request) {
	app := r.URL.Query().Get("app")
	if app == "" {
		http.Error(w, "app is required", http.StatusBadRequest)
		return
	}
	if app == "TVBox" || app == "AppleBox" || app == "CatBox" {
		http.Error(w, "cannot delete built-in pages", http.StatusBadRequest)
		return
	}

	pages, _ := a.readPages("pages.json")
	var target *models.PageModel
	for i, p := range pages {
		if p.App == app {
			target = &pages[i]
			break
		}
	}

	if target != nil {
		a.removePageFiles(*target)
		newPages := make([]models.PageModel, 0, len(pages)-1)
		for _, p := range pages {
			if p.App != app {
				newPages = append(newPages, p)
			}
		}
		pages = newPages
		if err := a.writePages(pages, "pages.json"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (a *API) removePageFiles(model models.PageModel) {
	pagePath := filepath.Join(a.RootDir, "mod-ce", model.Page+".html")
	os.Remove(pagePath)
	filePath := filepath.Join(a.RootDir, model.File+".json")
	os.Remove(filePath)
}
