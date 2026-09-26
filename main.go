package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"server-panel/internal/auth"
	"server-panel/internal/files"
	"server-panel/internal/logs"
	"server-panel/internal/services"
	"server-panel/internal/terminal"
)

//go:embed web
var webFS embed.FS

func main() {
	port := getEnv("PANEL_PORT", "8080")
	dataDir := getEnv("PANEL_DATA_DIR", "/etc/server-panel")
	files.Root = getEnv("PANEL_ROOT", "/")

	token, err := auth.LoadOrCreateToken(dataDir)
	if err != nil {
		log.Fatalf("impossible de charger/créer le token: %v", err)
	}

	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	// Frontend statique (page de login incluse : elle-même demande le token via API)
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API protégée par token
	protected := http.NewServeMux()
	protected.HandleFunc("/api/terminal", terminal.Handler)
	protected.HandleFunc("/api/files/list", files.List)
	protected.HandleFunc("/api/files/download", files.Download)
	protected.HandleFunc("/api/files/upload", files.Upload)
	protected.HandleFunc("/api/files/delete", files.Delete)
	protected.HandleFunc("/api/services/list", services.List)
	protected.HandleFunc("/api/services/action", services.Action)
	protected.HandleFunc("/api/logs", logs.Stream)

	mux.Handle("/api/", auth.Middleware(token, protected))

	addr := fmt.Sprintf(":%s", port)
	fmt.Println("========================================")
	fmt.Println(" Server Panel démarré ✅")
	fmt.Printf(" URL locale   : http://localhost:%s\n", port)
	fmt.Printf(" Token d'accès: %s\n", token)
	fmt.Println(" (stocké dans", auth.TokenPath(dataDir), ")")
	fmt.Println("========================================")

	log.Fatal(http.ListenAndServe(addr, mux))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
