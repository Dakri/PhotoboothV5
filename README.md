# PhotoboothV5

> Moderne, standalone Photobooth-Software für Raspberry Pi 3+  
> Ein einziger Service – WLAN, Kamera, Webserver und Multi-Client Steuerung in einem.

---

## Überblick

PhotoboothV5 verwandelt einen Raspberry Pi in eine vollständige Photobooth-Station. Kamera anschließen, Service starten – fertig. Clients verbinden sich per WLAN und steuern die Booth über den Browser.

### Was macht die Software?

- **WLAN Access Point** – Pi eröffnet beim Start ein eigenes WLAN, Clients verbinden sich direkt
- **Kamera-Steuerung** – Canon DSLR via gphoto2/USB, Fotos werden direkt heruntergeladen (kein SD-Karten-Umweg)
- **Echtzeit-Kommunikation** – Alle Clients sehen Countdown, Vorschau und Galerie synchron via WebSocket
- **Bildverarbeitung** – Originalfotos werden automatisch in Preview + Thumbnail umgewandelt
- **Multi-Client** – Dashboard, Buzzer, Countdown, Galerie – alles gleichzeitig auf verschiedenen Geräten
- **Legacy-Support** – Ältere iPads/Tablets werden über einen separaten Vanilla-HTML Client unterstützt
- **USB-Export** – Fotos auf USB-Stick kopieren per Knopfdruck

---

## Architektur

```
┌──────────────────────────────────────────────────┐
│                  Raspberry Pi                     │
│                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │  WiFi AP    │  │  Node.js    │  │  gphoto2  │ │
│  │  (nmcli)    │  │  Server     │──│  (Canon)  │ │
│  └──────┬──────┘  └──────┬──────┘  └──────────┘ │
│         │                │                        │
│         │         ┌──────┴──────┐                │
│         │         │   sharp     │                │
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
| **Backend** | Node.js + Express + TypeScript | Einheitlicher JS-Stack, guter WebSocket-Support |
| **Echtzeit** | `ws` (WebSocket) | Leichtgewichtig, kein Socket.IO-Overhead |
| **Frontend** | Vite + Vue 3 + Tailwind 4 + Pinia | Moderner, reaktiver UI-Stack |
| **Legacy-Client** | Vanilla HTML/CSS/JS (ES5) | Kein Framework, funktioniert auf Safari 9+ |
| **Bildverarbeitung** | sharp (libvips) | Prebuilt ARM-Binaries, kein Kompilieren nötig |
| **Kamera** | gphoto2 via child_process | Bewährt, kein natives Addon nötig |
| **WLAN** | NetworkManager (nmcli) | Ab Pi OS Bookworm Standard |
| **Deployment** | systemd Service | Ein `install.sh`, kein Docker |

---

## Projektstruktur

```
PhotoboothV5/
├── server/                     # Backend (Node.js + TypeScript)
│   ├── src/
│   │   ├── index.ts            # Entry Point
│   │   ├── config.ts           # Zentrale Konfiguration
│   │   ├── camera/
│   │   │   └── gphoto.ts       # gphoto2 Wrapper + Mock-Modus
│   │   ├── image/
│   │   │   └── processor.ts    # sharp: Preview + Thumbnail
│   │   ├── storage/
│   │   │   ├── photos.ts       # Foto-Verwaltung
│   │   │   └── usb.ts          # USB-Stick Erkennung + Export
│   │   ├── websocket/
│   │   │   ├── server.ts       # WebSocket Server + Event-Flow
│   │   │   └── events.ts       # Event-Typen & Message-Factories
│   │   ├── api/
│   │   │   └── routes.ts       # REST API (inkl. Legacy-Polling)
│   │   ├── network/
│   │   │   ├── wifi.ts         # WLAN Hotspot (nmcli)
│   │   │   └── captive.ts      # Captive Portal (dnsmasq)
│   │   └── auth/
│   │       └── session.ts      # PIN-basierte Auth
│   ├── package.json
│   └── tsconfig.json
│
├── client/                     # Frontend (Vite + Vue 3)
│   ├── src/
│   │   ├── App.vue
│   │   ├── main.ts
│   │   ├── stores/             # Pinia Stores
│   │   ├── composables/        # useWebSocket, useCountdown
│   │   ├── views/              # Dashboard, Buzzer, Countdown, Gallery
│   │   ├── components/         # Wiederverwendbare Komponenten
│   │   └── router/
│   └── package.json
│
├── legacy-client/              # Fallback für alte Browser (ES5)
│   ├── index.html
│   ├── gallery.html
│   ├── css/style.css
│   └── js/
│       ├── app.js              # Polling-basierte Logik
│       └── gallery.js
│
├── scripts/
│   ├── install.sh              # Service installieren
│   ├── uninstall.sh            # Service entfernen
│   └── photobooth.service      # systemd Unit-File
│
├── config/
│   └── default.json            # Standard-Konfiguration
│
├── data/                       # Laufzeit-Daten (gitignored)
│   ├── photos/
│   │   ├── original/
│   │   ├── preview/
│   │   └── thumb/
│   └── config.json             # Laufzeit-Überschreibungen
│
└── docs/
    └── APP_DOCUMENTATION.md    # Technische Dokumentation
```

---

## Voraussetzungen

### Hardware
- Raspberry Pi 3 oder neuer (auch x64 kompatibel)
- Canon DSLR (oder kompatible Kamera, siehe [gphoto2 Kompatibilitätsliste](http://www.gphoto.org/proj/libgphoto2/support.php))
- USB-Kabel (Kamera → Pi)
- Optional: USB-Stick für Foto-Export

### Software (auf dem Pi)
- **Raspberry Pi OS Bookworm** (empfohlen) oder neuer
- **Node.js 18+** (`sudo apt install nodejs npm`)
- **gphoto2** (`sudo apt install gphoto2`)
- **dnsmasq** (optional, für Captive Portal: `sudo apt install dnsmasq`)

> **Hinweis:** NetworkManager ist ab Pi OS Bookworm vorinstalliert.  
> sharp bringt prebuilt-Binaries für ARM mit – kein manuelles Kompilieren nötig.

---

## Installation

### Schnellstart

```bash
# Repository klonen
git clone <repo-url> /opt/photobooth
cd /opt/photobooth

# Backend Dependencies installieren
cd server && npm install && npm run build && cd ..

# Frontend bauen
cd client && npm install && npm run build && cd ..

# Service installieren & starten
sudo bash scripts/install.sh
```

### Was macht `install.sh`?

1. Prüft ob Node.js und gphoto2 installiert sind
2. Kopiert `photobooth.service` nach `/etc/systemd/system/`
3. Aktiviert den Service (startet nach jedem Reboot automatisch)
4. Richtet bei Bedarf den WLAN-Hotspot ein

**Kein Docker, kein langes Setup-Script, keine outdated Pakete.**

---

## Konfiguration

Alle Einstellungen in `config/default.json`. Zur Laufzeit können Änderungen über das Dashboard oder die API gemacht werden – diese werden in `data/config.json` gespeichert und überschreiben die Defaults.

### Wichtige Einstellungen

| Einstellung | Default | Beschreibung |
|---|---|---|
| `server.port` | `80` | HTTP Port |
| `camera.mock` | `false` | `true` für Entwicklung ohne Kamera |
| `photobooth.countdownSeconds` | `5` | Countdown-Dauer |
| `photobooth.previewDisplaySeconds` | `8` | Vorschau-Anzeige nach Aufnahme |
| `image.previewWidth` | `800` | Preview-Breite in Pixel |
| `image.thumbnailWidth` | `200` | Thumbnail-Breite in Pixel |
| `wifi.ssid` | `"Photobooth"` | WLAN-Name |
| `wifi.password` | `""` | Leer = offenes WLAN |
| `auth.dashboardPin` | `"1234"` | PIN für Dashboard-Zugriff |

---

## Client-Typen

### Moderner Client (Vue 3)
- **URL:** `http://192.168.4.1/`
- **Browser:** Chrome 60+, Safari 12+, Firefox 60+
- **Features:** WebSocket, Echtzeit-Countdown, Touch-Buzzer, Galerie mit Lightbox
- **Routen:**
  - `/` – Dashboard (PIN-geschützt)
  - `/buzzer` – Touch-Auslöser
  - `/countdown` – Vollbild-Countdown
  - `/gallery` – Foto-Galerie
  - `/preview` – Foto-Vorschau

### Legacy Client (Vanilla HTML)
- **URL:** `http://192.168.4.1/legacy/`
- **Browser:** Safari 9+ (alte iPads), jeder Browser mit JavaScript
- **Features:** Polling-basiert (kein WebSocket), einfaches CSS (kein Flexbox-Gap, keine Custom Properties)
- **Einschränkungen:** Kein Echtzeit-Countdown (500ms Polling-Delay), einfacheres UI

### Hardware Client
- **Protokoll:** WebSocket
- **Beispiel:** Raspberry Pi mit physischem Buzzer-Button
- **Registriert sich als:** `role: "hardware"`

---

## API Referenz

### REST Endpoints

| Methode | Route | Beschreibung | Auth |
|---|---|---|---|
| `GET` | `/api/status` | Server-Status + Zustand | Nein |
| `GET` | `/api/camera/status` | Kamera-Info | Nein |
| `POST` | `/api/trigger` | Foto auslösen | Nein |
| `GET` | `/api/photos` | Foto-Liste (paginiert) | Nein |
| `GET` | `/api/photos/latest` | Letztes Foto | Nein |
| `DELETE` | `/api/photos/:filename` | Foto löschen | Ja |
| `GET` | `/api/usb/drives` | USB-Sticks erkennen | Ja |
| `POST` | `/api/usb/export` | Fotos auf USB kopieren | Ja |
| `GET` | `/api/config` | Konfiguration lesen | Ja |
| `PATCH` | `/api/config` | Konfiguration ändern | Ja |
| `GET` | `/api/legacy/poll` | Status für Legacy-Client | Nein |
| `POST` | `/api/auth/login` | PIN-Login | – |
| `POST` | `/api/auth/logout` | Logout | – |
| `GET` | `/api/auth/check` | Auth-Status prüfen | – |

### WebSocket Events

Verbindung über `ws://192.168.4.1/ws`

**Client → Server:**

| Event | Daten | Beschreibung |
|---|---|---|
| `register` | `{ role: "buzzer" }` | Client-Rolle registrieren |
| `trigger` | – | Foto auslösen |
| `pong` | – | Heartbeat-Antwort |

**Server → Client:**

| Event | Daten | Beschreibung |
|---|---|---|
| `status` | `{ state: "idle" }` | Zustandsänderung |
| `countdown` | `{ remaining: 3, total: 5 }` | Countdown-Tick |
| `capturing` | – | Kamera löst aus |
| `processing` | – | Bild wird verarbeitet |
| `photo_ready` | `{ filename, urls }` | Foto bereit |
| `error` | `{ message, code }` | Fehler |
| `clients_update` | `{ clients: [...] }` | Client-Liste geändert |
| `ping` | – | Heartbeat |

### Zustandsmaschine

```
idle → countdown → capturing → processing → preview → idle
                                                ↑
                                          (nach X Sekunden)
```

---

## Entwicklung

### Backend entwickeln

```bash
cd server
npm install
npm run dev          # TypeScript mit Hot-Reload (tsx watch)
```

> **Tipp:** `camera.mock: true` in `config/default.json` setzen um ohne Kamera zu entwickeln.

### Frontend entwickeln

```bash
cd client
npm install
npm run dev          # Vite Dev Server (Port 5173)
```

> Der Vite Dev Server hat einen Proxy auf den Backend-Server konfiguriert.

### Für Produktion bauen

```bash
# Backend
cd server && npm run build

# Frontend (Output → client/dist/)
cd client && npm run build
```

---

## WLAN & Netzwerk

### Automatisch (Pi OS Bookworm+)

Die Software erstellt beim Start automatisch einen WLAN-Hotspot über NetworkManager:
- **SSID:** Konfigurierbar (Default: `Photobooth`)
- **Passwort:** Optional (Default: offen)
- **IP:** `192.168.4.1`
- **DHCP:** Automatisch über NetworkManager

### Captive Portal

Wenn `dnsmasq` installiert ist, werden alle DNS-Anfragen auf die Pi-IP umgeleitet. Dadurch öffnen die meisten Geräte automatisch das Dashboard beim Verbinden.

> ⚠ **Captive Portal funktioniert nicht auf allen Geräten zuverlässig.** Fallback: Die IP-Adresse wird dem Nutzer im WLAN-Namen angezeigt.

### Manuelles Setup (ältere Pi OS Versionen)

Falls NetworkManager nicht verfügbar ist, kann der Hotspot manuell mit `hostapd` + `dnsmasq` eingerichtet werden. Siehe [docs/APP_DOCUMENTATION.md](docs/APP_DOCUMENTATION.md) für eine Anleitung.

---

## Fehlerbehebung

| Problem | Lösung |
|---|---|
| Kamera wird nicht erkannt | `gphoto2 --auto-detect` prüfen, USB-Kabel testen |
| WLAN startet nicht | `nmcli device status` prüfen, `wifi.enabled: false` setzen für manuelles Setup |
| Fotos dauern lang | Normal auf Pi 3 (~200ms Processing), Thumbnail-Größe reduzieren |
| Legacy-Client zeigt nichts | `/legacy/` öffnen (mit Slash!), JavaScript aktiviert? |
| Dashboard PIN vergessen | In `config/default.json` ändern oder `data/config.json` löschen |
| Service startet nicht | `sudo journalctl -u photobooth -f` für Logs |

---

## Lizenz

MIT
