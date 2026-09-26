package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const tokenFileName = "token"

// TokenPath renvoie le chemin du fichier contenant le token, dans dataDir.
func TokenPath(dataDir string) string {
	return filepath.Join(dataDir, tokenFileName)
}

// LoadOrCreateToken lit le token existant, ou en génère un nouveau
// et l'écrit sur disque (lisible uniquement par le propriétaire) s'il n'existe pas encore.
func LoadOrCreateToken(dataDir string) (string, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", err
	}
	p := TokenPath(dataDir)

	if data, err := os.ReadFile(p); err == nil {
		tok := strings.TrimSpace(string(data))
		if tok != "" {
			return tok, nil
		}
	}

	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	tok := hex.EncodeToString(buf)

	if err := os.WriteFile(p, []byte(tok), 0o600); err != nil {
		return "", err
	}
	return tok, nil
}

// Middleware protège une route en exigeant le header
// "Authorization: Bearer <token>" ou le paramètre ?token=<token>
// (utile pour la connexion WebSocket depuis le navigateur).
func Middleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		supplied := r.URL.Query().Get("token")
		if supplied == "" {
			h := r.Header.Get("Authorization")
			supplied = strings.TrimPrefix(h, "Bearer ")
		}

		if subtle.ConstantTimeCompare([]byte(supplied), []byte(token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
