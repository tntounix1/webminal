package logs

import (
	"bufio"
	"log"
	"net/http"
	"os/exec"
	"regexp"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var unitNameRe = regexp.MustCompile(`^[a-zA-Z0-9@._-]+\.service$`)

// Stream ouvre "journalctl -f" (tout le système, ou -u <service> si précisé)
// et pousse chaque nouvelle ligne au client via WebSocket.
func Stream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("logs: upgrade error:", err)
		return
	}
	defer conn.Close()

	args := []string{"-f", "-n", "100", "--no-pager"}
	if unit := r.URL.Query().Get("unit"); unit != "" {
		if !unitNameRe.MatchString(unit) {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("nom de service invalide"))
			return
		}
		args = append(args, "-u", unit)
	}

	cmd := exec.Command("journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("logs: pipe error:", err)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Println("logs: start error:", err)
		return
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := conn.WriteMessage(websocket.TextMessage, scanner.Bytes()); err != nil {
			return
		}
	}
}
