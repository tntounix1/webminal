package files

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Root est le répertoire racine accessible depuis le panneau.
// Par défaut "/", modifiable via la variable d'env PANEL_ROOT.
var Root = "/"

type entry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// safePath résout un chemin demandé par le client et vérifie qu'il reste
// bien à l'intérieur de Root (protection contre les "../../etc/passwd").
func safePath(reqPath string) (string, error) {
	cleaned := filepath.Clean("/" + reqPath)
	full := filepath.Join(Root, cleaned)
	rel, err := filepath.Rel(Root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", os.ErrPermission
	}
	return full, nil
}

// List renvoie le contenu d'un répertoire en JSON.
func List(w http.ResponseWriter, r *http.Request) {
	p, err := safePath(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, "chemin invalide", http.StatusBadRequest)
		return
	}

	items, err := os.ReadDir(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := make([]entry, 0, len(items))
	for _, it := range items {
		info, err := it.Info()
		if err != nil {
			continue
		}
		out = append(out, entry{
			Name:    it.Name(),
			Path:    strings.TrimPrefix(filepath.Join(p, it.Name()), Root),
			IsDir:   it.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDir != out[j].IsDir {
			return out[i].IsDir
		}
		return out[i].Name < out[j].Name
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// Download envoie le contenu d'un fichier.
func Download(w http.ResponseWriter, r *http.Request) {
	p, err := safePath(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, "chemin invalide", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(p)+"\"")
	http.ServeFile(w, r, p)
}

// Upload écrit un fichier envoyé en multipart/form-data dans le répertoire demandé.
func Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil { // 64 Mo en mémoire max
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dir, err := safePath(r.FormValue("path"))
	if err != nil {
		http.Error(w, "chemin invalide", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create(filepath.Join(dir, filepath.Base(header.Filename)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Delete supprime un fichier ou un répertoire (récursivement).
func Delete(w http.ResponseWriter, r *http.Request) {
	p, err := safePath(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, "chemin invalide", http.StatusBadRequest)
		return
	}
	if p == Root {
		http.Error(w, "impossible de supprimer la racine", http.StatusBadRequest)
		return
	}
	if err := os.RemoveAll(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
