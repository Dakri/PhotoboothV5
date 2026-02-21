# PhotoboothV5

> Moderne, standalone Photobooth-Software für Raspberry Pi 3+  
> **Single Binary** Architektur – Go Backend + Vue 3 Frontend. Keine Runtime-Installation nötig.

---

## Überblick

PhotoboothV5 verwandelt einen Raspberry Pi in eine vollständige Photobooth-Station. Kamera anschließen, Service starten – fertig. Clients verbinden sich per WLAN und steuern die Booth über den Browser.

### Was macht die Software?

- **Single Binary** – Ein Go-Binary, keine Node.js/Python Runtime nötig
- **WLAN Access Point** – Pi eröffnet beim Start ein eigenes WLAN, Clients verbinden sich direkt
- **Eingebauter DNS Server** – Captive Portal ohne dnsmasq, direkt in Go implementiert
- **Kamera-Steuerung** – Canon DSLR via gphoto2/USB, Fotos werden direkt heruntergeladen
- **Echtzeit-Kommunikation** – WebSocket für Countdown, Vorschau und Galerie
- **Bildverarbeitung** – Pure Go (`disintegration/imaging`), kein cgo/ImageMagick nötig
- **Client-Modi** – 6 konfigurierbare Modi für verschiedene Geräte (Buzzer, Countdown, Galerie, etc.)
- **Verlassen-Sperre** – 10-Tap Exit-Lock verhindert versehentliches Verlassen des Client-Modus
- **Live-Dashboard** – Professionelles Admin-Panel mit Server-Logs in Echtzeit
- **Legacy-Support** – Ältere iPads/Tablets über separaten Vanilla-HTML Client
- **USB-Export** – Fotos auf USB-Stick kopieren per Knopfdruck

---

## Architektur

```
┌──────────────────────────────────────────────────┐
│                  Raspberry Pi                     │
│                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │  WiFi AP    │  │  Go Binary  │  │  gphoto2  │ │
│  │  (nmcli)    │  │  (HTTP+WS+  │──│  (Canon)  │ │
│  │             │  │   DNS)      │  │           │ │
│  └──────┬──────┘  └──────┬──────┘  └──────────┘ │
│         │                │                        │
│         │         ┌──────┴──────┐                │
│         │         │  imaging    │                │
│         │         │ (Thumbnail) │                │
│         │         └─────────────┘                │
└─────────┼────────────────┼───────────────────────┘
          │                │
    ┌─────┴────┐    ┌──────┴──────┐
    │  Clients │    │  Clients    │
    │ (modern) │    │  (legacy)   │
    │ Vue 3+WS │    │ HTML+Polling│
    └──────────┘    └─────────────┘
```

### Tech-Stack

| Komponente | Technologie | Warum |
|---|---|---|
| **Backend** | **Go (Golang)** | Single Binary, keine Runtime, performant |
| **Echtzeit** | `gorilla/websocket` | Bewährt, breite Kompatibilität |
| **Frontend** | Vite + Vue 3 + Tailwind 4 + Pinia | Moderner, reaktiver UI-Stack |
| **Legacy-Client** | Vanilla HTML/CSS/JS (ES5) | Kein Framework, funktioniert auf Safari 9+ |
| **Bildverarbeitung** | `disintegration/imaging` | Pure Go, kein cgo nötig |
| **DNS** | `miekg/dns` | Eingebauter DNS Server für Captive Portal |
| **Kamera** | gphoto2 via `os/exec` | Bewährt, kein natives Addon nötig |
| **WLAN** | NetworkManager (nmcli) | Ab Pi OS Bookworm Standard |
| **Deployment** | systemd Service | Ein `install.sh`, kein Docker |

---

## Projektstruktur

```
PhotoboothV5/
├── backend/                    # Go Backend (läuft auf Pi)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry Point
│   ├── internal/
│   │   ├── api/                # REST Handler
│   │   ├── app/                # App Controller (State Machine)
│   │   ├── camera/             # gphoto2 Wrapper + CameraInfo
│   │   ├── config/             # Config Structs
│   │   ├── dns/                # Eigener DNS Server (Captive Portal)
│   │   ├── imaging/            # Resize & Processing (pure Go)
│   │   ├── logging/            # Ring-Buffer Logger + WS Broadcast
│   │   ├── network/            # WiFi (nmcli wrapper)
│   │   ├── storage/            # Foto-Verwaltung + USB Export
│   │   └── websocket/          # Hub & Clients
│   ├── go.mod
│   └── go.sum
│
├── frontend/                   # Vue 3 Frontend (Entwicklung lokal)
│   ├── src/
│   │   ├── composables/        # useFullscreen, useExitLock
│   │   ├── stores/             # Pinia: photobooth, gallery, clientMode
│   │   ├── views/              # Dashboard, ClientView, ModeSelect, Gallery
│   │   └── router/
│   ├── vite.config.ts
│   └── package.json
│
├── legacy/                     # Legacy Client (Statisch, ES5)
│   ├── index.html
│   └── js/
│
├── scripts/
│   ├── build-pi.sh             # Cross-Compile (Go → ARM) + Frontend Build
│   ├── deploy.sh               # SSH ControlMaster Deploy
│   ├── install.sh              # systemd Service + Dependencies
│   └── photobooth.service      # systemd Unit
│
├── config/
│   └── default.json            # Standard-Konfiguration
│
└── data/                       # Laufzeit-Daten (gitignored)
    └── photos/
        ├── original/
        ├── preview/
        └── thumb/
```

---

## Voraussetzungen

### Hardware
- Raspberry Pi 3 oder neuer (ARM64)
- Canon DSLR (oder kompatible Kamera, siehe [gphoto2 Kompatibilitätsliste](http://www.gphoto.org/proj/libgphoto2/support.php))
- USB-Kabel (Kamera → Pi)
- Optional: USB-Stick für Foto-Export

### Software (auf dem Pi)
- **Raspberry Pi OS Bookworm** (empfohlen) oder neuer
- **gphoto2** (`sudo apt install gphoto2`)
- **dnsmasq-base** (`sudo apt install dnsmasq-base`) – für DHCP im Hotspot
- **Kein** Node.js nötig!
- **Kein** dnsmasq/hostapd nötig (DNS Server ist eingebaut)

> **Hinweis:** NetworkManager ist ab Pi OS Bookworm vorinstalliert.

---

## Installation

### Schnellstart (Cross-Compile & Deploy)

*** Dies erfordert NodeJs 20+ und Go 1.25+ ***
*** Ich empfehle WSL2 für das deployment ***

```bash
# Auf dem Entwicklungs-PC (nicht auf dem Pi!)
./scripts/deploy.sh pi@192.168.x.x
```

Das Deploy-Script:
1. Baut das Frontend (Vite)
2. Cross-Compiliert das Go Backend (ARM64)
3. Überträgt alles per rsync zum Pi
4. Installiert den systemd Service

### Manuell auf dem Pi

```bash
# Binary + Frontend nach /opt/photobooth kopieren
# Dann:
sudo bash /opt/photobooth/scripts/install.sh
```

---

## Client-Modi

PhotoboothV5 bietet ein flexibles Client-Mode-System. Clients wählen über `/modes` ihren Modus und werden dann im Vollbild-Modus gesperrt.

### Verfügbare Modi

| Modus | Buzzer | Countdown | Vorschau | Galerie |
|---|:---:|:---:|:---:|:---:|
| **Vollständig** | ✅ | ✅ | ✅ | — |
| **Auslöser + Countdown** | ✅ | ✅ | — | — |
| **Monitor + Vorschau** | — | ✅ | ✅ | — |
| **Nur Vorschau** | — | — | ✅ | — |
| **Nur Countdown** | — | ✅ | — | — |
| **Galerie** | — | — | — | ✅ |

### Exit-Lock

- **10× schnell tippen** → Entsperr-Overlay erscheint
- 3 Optionen: Modus wechseln, Client verlassen, Abbrechen
- Tap-Counter wird nach 3s Inaktivität zurückgesetzt
- Subtile Punkte oben rechts zeigen Fortschritt

### Fullscreen

- Wird automatisch beim Moduswahl aktiviert
- Cross-Browser-kompatibel (webkit, moz, ms Prefixes)

---

## Dashboard

Das Admin-Dashboard (`/`) bietet:

- **System-State** mit farbigem Status-Indikator
- **Trigger-Button** zum manuellen Auslösen
- **Letztes Foto** als Vorschau
- **Kamera-Info** – Modell, Akku-Level (Farbbalken), Objektiv, freier Speicher
- **Live-Logs** – Monospace Log-Viewer mit farbcodierten Levels, Auto-Scroll
- **Uptime & Client-Count** im Header

---

## API Referenz

### REST Endpoints

| Methode | Route | Beschreibung |
|---|---|---|
| `GET` | `/api/status` | Server-Status (State, Clients, Uptime) |
| `POST` | `/api/trigger` | Foto auslösen |
| `GET` | `/api/photos` | Foto-Liste |
| `GET` | `/api/photos/latest` | Letztes Foto |
| `GET` | `/api/logs?limit=100` | Server-Logs (Ring-Buffer) |
| `GET` | `/api/legacy/poll` | Kombinierter Status für Legacy-Client |

### WebSocket Events

Verbindung über `ws://192.168.4.1/ws`

**Client → Server:**

| Event | Daten | Beschreibung |
|---|---|---|
| `register` | `{ role: "buzzer-countdown-preview" }` | Client-Modus registrieren |
| `trigger` | – | Foto auslösen |

**Server → Client:**

| Event | Daten | Beschreibung |
|---|---|---|
| `status` | `{ state: "idle" }` | Zustandsänderung |
| `countdown` | `{ remaining: 3, total: 5 }` | Countdown-Tick |
| `photo_ready` | `{ filename, url, thumbUrl }` | Foto bereit |
| `log` | `{ level, source, message, timestamp }` | Log-Eintrag (Live) |
| `error` | `{ message }` | Fehler |

### Zustandsmaschine

```
idle → countdown → capturing → processing → preview → idle
                                                 ↑
                                           (nach 8 Sekunden)
```

---

## Entwicklung

### Backend entwickeln

```bash
cd backend
go run ./cmd/server
```

> **Tipp:** `camera.mock: true` in `config/default.json` setzen um ohne Kamera zu entwickeln.

### Frontend entwickeln

```bash
cd frontend
npm install
npm run dev          # Vite Dev Server (Port 5173)
```

> Der Vite Dev Server hat einen Proxy auf den Backend-Server konfiguriert.

### Für Produktion bauen

```bash
./scripts/build-pi.sh   # Frontend + Go Cross-Compile → dist/
```

---

## Fehlerbehebung

| Problem | Lösung |
|---|---|
| WLAN: Client bekommt keine IP | `dnsmasq-base` installieren: `sudo apt install dnsmasq-base` |
| Kamera wird nicht erkannt | `gphoto2 --auto-detect` prüfen, USB-Kabel testen |
| WLAN startet nicht | `nmcli device status` prüfen |
| Deploy-Script hängt | SSH-Key einrichten oder Script nutzt SSH ControlMaster |
| Service startet nicht | `sudo journalctl -u photobooth -f` für Logs |
| Captive Portal zeigt nichts | DNS Server braucht Port 53 → `sudo` nötig |

---

## Lizenz

MIT
