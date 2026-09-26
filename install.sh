#!/usr/bin/env bash
set -euo pipefail

# ====== À adapter avec ton propre dépôt une fois publié sur GitHub ======
REPO="TON_USER/webminal"
# ==========================================================================

INSTALL_DIR="/opt/webminal"
BIN_PATH="$INSTALL_DIR/webminal"
SERVICE_PATH="/etc/systemd/system/webminal.service"

if [ "$(id -u)" -ne 0 ]; then
  echo "❌ Lance ce script en root (ou via sudo)."
  exit 1
fi

# Un port est "libre" si rien n'écoute dessus (TCP).
port_is_free() {
  ! ss -Htln 2>/dev/null | awk '{print $4}' | grep -q ":$1\$"
}

if [ -n "${PANEL_PORT:-}" ]; then
  # Port choisi explicitement par l'utilisateur : on respecte son choix,
  # même s'il est déjà pris (il verra l'erreur au démarrage du service).
  PORT="$PANEL_PORT"
else
  PORT=8080
  if ! port_is_free "$PORT"; then
    echo "⚠️  Le port $PORT est déjà utilisé, recherche d'un port libre..."
    for candidate in 8081 8082 8090 8888 9090 9091; do
      if port_is_free "$candidate"; then
        PORT="$candidate"
        break
      fi
    done
    if ! port_is_free "$PORT"; then
      echo "❌ Aucun port libre trouvé parmi les ports par défaut."
      echo "   Relance avec : sudo PANEL_PORT=<ton_port> bash install.sh"
      exit 1
    fi
  fi
  echo "ℹ️  Port sélectionné automatiquement : $PORT"
  echo "   (pour en imposer un toi-même : sudo PANEL_PORT=<port> bash install.sh)"
fi

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) GOARCH="amd64" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *) echo "❌ Architecture non supportée: $ARCH"; exit 1 ;;
esac

echo "📦 Récupération de la dernière version pour linux-$GOARCH..."
mkdir -p "$INSTALL_DIR"

LATEST_URL="https://github.com/$REPO/releases/latest/download/webminal-linux-$GOARCH"
curl -fsSL "$LATEST_URL" -o "$BIN_PATH"
chmod +x "$BIN_PATH"

echo "⚙️  Configuration du service systemd..."
cat > "$SERVICE_PATH" <<EOF
[Unit]
Description=Webminal
After=network.target

[Service]
ExecStart=$BIN_PATH
Environment=PANEL_PORT=$PORT
Restart=on-failure
RestartSec=3
User=root

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable webminal
systemctl restart webminal

sleep 1
IP="$(curl -fsSL ifconfig.me || hostname -I | awk '{print $1}')"
TOKEN="$(cat /etc/webminal/token 2>/dev/null || echo '(voir les logs : journalctl -u webminal -n 20)')"

echo ""
echo "========================================"
echo " ✅ Webminal est installé et lancé !"
echo ""
echo " Ouvre ce lien dans le navigateur de ton iPad/phone :"
echo "   http://$IP:$PORT"
echo ""
echo " Token d'accès (à coller dans l'écran de connexion) :"
echo "   $TOKEN"
echo "========================================"
echo ""
echo "⚠️  Pense à sécuriser l'accès : mets ce port derrière un reverse"
echo "   proxy HTTPS (ex: Caddy) si tu veux l'utiliser en dehors d'un VPN."
