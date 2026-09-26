# Server Panel

Panneau d'administration web léger pour gérer un serveur Linux (VPS) depuis
n'importe quel navigateur — iPad, phone, peu importe. Pas d'app à installer
côté client, pas de config compliquée côté serveur.

## Fonctionnalités

- 🖥️ **Terminal** interactif complet dans le navigateur (xterm.js + pty)
- 📁 **Fichiers** : parcourir, uploader, télécharger, supprimer
- ⚙️ **Services** : lister, démarrer, arrêter, redémarrer les services systemd
- 📜 **Logs** : suivi en direct via `journalctl`

Un seul binaire Go, aucune dépendance à installer sur le serveur.

## Tu ne sais pas coder ? Aucun problème.

Tu n'as **jamais besoin d'installer Go ni de compiler quoi que ce soit toi-même**.
Voici le circuit complet :

### 1. Publier le projet sur GitHub

1. Crée un nouveau dépôt sur GitHub (par ex. `server-panel`), public ou privé.
2. Pousse tout ce dossier dedans :
   ```bash
   cd server-panel
   git init
   git add .
   git commit -m "Premier commit"
   git branch -M main
   git remote add origin https://github.com/TON_USER/server-panel.git
   git push -u origin main
   ```
3. Ouvre `install.sh` et remplace `TON_USER/server-panel` par le vrai nom
   de ton dépôt (ligne `REPO=...`).

### 2. Déclencher la compilation automatique

Dès que tu pousses un **tag** (une étiquette de version), GitHub Actions
compile automatiquement les binaires pour Linux (amd64 + arm64) et les
publie dans l'onglet "Releases" de ton dépôt :

```bash
git tag v1.0.0
git push origin v1.0.0
```

Va vérifier dans l'onglet **Actions** de GitHub : tu verras le build
tourner (~1 minute), puis dans **Releases** les binaires apparaîtront.

### 3. Installer sur ton VPS

Connecte-toi une seule fois en SSH à ton VPS (ou demande à quelqu'un de le
faire pour toi), et lance :

```bash
curl -fsSL https://raw.githubusercontent.com/TON_USER/server-panel/main/install.sh | sudo bash
```

Le script :
- télécharge le bon binaire depuis ta release GitHub
- l'installe comme service systemd (démarre au boot, redémarre s'il crashe)
- génère un token d'accès aléatoire
- affiche le lien + le token à copier-coller

### 4. Utiliser depuis l'iPad / phone

Ouvre Safari (ou Chrome), colle le lien `http://IP_DU_SERVEUR:8080`, entre
le token → tu es dans le panneau. Tu peux même l'ajouter à l'écran d'accueil
pour un accès en un tap.

## Mettre à jour

Modifie le code, pousse un nouveau tag (`v1.0.1`, etc.), puis relance
`install.sh` sur le serveur — il retélécharge la dernière version.

## Sécurité — à lire avant d'exposer le port publiquement

- Le token protège l'accès, mais il transite en clair si tu n'as pas HTTPS.
  Pour un usage sérieux, mets un reverse proxy comme
  [Caddy](https://caddyserver.com/) devant (HTTPS automatique et gratuit).
- Évite d'exposer le port directement sur Internet sans y réfléchir : un
  minimum est de le restreindre par pare-feu (IP autorisées) ou de passer
  par un VPN (ex. Tailscale) entre ton phone et le serveur.
- Le compte qui lance le panneau (`root` par défaut dans le service) a un
  accès total au système — c'est le but, mais sois-en conscient.

## Licence

MIT — fais-en ce que tu veux.
