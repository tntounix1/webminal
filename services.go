package services

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

// unitNameRe valide qu'un nom d'unité systemd est "propre" avant de
// l'utiliser dans une commande shell (évite toute injection de commande).
var unitNameRe = regexp.MustCompile(`^[a-zA-Z0-9@._-]+\.service$`)

type unit struct {
	Name        string `json:"name"`
	LoadState   string `json:"loadState"`
	ActiveState string `json:"activeState"`
	SubState    string `json:"subState"`
	Description string `json:"description"`
}

// List renvoie la liste des services systemd installés.
func List(w http.ResponseWriter, r *http.Request) {
	out, err := exec.Command("systemctl", "list-units", "--type=service", "--all",
		"--no-legend", "--no-pager", "--plain").Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var units []unit
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		units = append(units, unit{
			Name:        fields[0],
			LoadState:   fields[1],
			ActiveState: fields[2],
			SubState:    fields[3],
			Description: strings.Join(fields[4:], " "),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(units)
}

// Action démarre / arrête / redémarre un service. L'action et le nom
// d'unité sont strictement validés avant exécution.
func Action(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	action := r.URL.Query().Get("action")

	if !unitNameRe.MatchString(name) {
		http.Error(w, "nom de service invalide", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{"start": true, "stop": true, "restart": true}
	if !allowed[action] {
		http.Error(w, "action non autorisée", http.StatusBadRequest)
		return
	}

	if out, err := exec.Command("systemctl", action, name).CombinedOutput(); err != nil {
		http.Error(w, string(out)+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
