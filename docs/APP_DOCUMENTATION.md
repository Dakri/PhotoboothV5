# PhotoboothV5 – Technische Dokumentation

> Referenz-Dokument für die gesamte Entwicklung.  
> Alle Module, Abläufe, Schnittstellen und Design-Entscheidungen.

---

## Inhaltsverzeichnis

1. [Design-Prinzipien](#design-prinzipien)
2. [Backend Module (Go)](#backend-module-go)
3. [Event-Flow & Zustandsmaschine](#event-flow--zustandsmaschine)
4. [Frontend Architektur](#frontend-architektur)
5. [Client-Mode-System](#client-mode-system)
6. [Legacy Client](#legacy-client)
7. [Netzwerk & WLAN](#netzwerk--wlan)
8. [Logging-System](#logging-system)
9. [Build & Deployment](#build--deployment)
10. [Bekannte Limitierungen](#bekannte-limitierungen)

---

## Design-Prinzipien

### 1. Standalone – Keine externen Abhängigkeiten zur Laufzeit
- Kein Internet nötig
- Kein Docker, kein externer Dienst
- Single Binary – kein Node.js, kein Python
- Alles läuft auf dem Pi alleine

### 2. Single Binary Deployment
- Go-Backend kompiliert zu einem Binary
- Frontend wird eingebettet als statische Dateien
- Nur `gphoto2` als System-Dependency nötig

### 3. Graceful Degradation
- Kamera nicht da? → Mock-Modus oder klare Fehlermeldung
- WLAN nicht konfigurierbar? → Klare Warnung, manuelle Anleitung
- Alter Browser? → Legacy-Client über `/legacy/`
- WebSocket nicht verfügbar? → REST Polling

### 4. Alles konfigurierbar
- `config.json` = Basis-Einstellungen
- Kein Hardcoded-Wert im Code

---

## Backend Module (Go)

### `cmd/server/main.go` – Entry Point

**Ablauf:**
1. Logger initialisieren (Ring-Buffer)
2. `config.json` laden
3. Verzeichnisse erstellen (`data/photos/original/`, `preview/`, `thumb/`)
4. Module initialisieren: Camera, Imaging, Storage, WebSocket Hub, App Controller
5. Log-Broadcast-Funktion registrieren (Logger → WebSocket)
6. WiFi-Hotspot starten (wenn konfiguriert)
7. DNS Server starten (Captive Portal, Port 53)
8. HTTP Server starten (API + statische Dateien)

---

### `internal/config/` – Konfiguration

**Config-Struct:**
```go
type Config struct {
    Server   ServerConfig   // Port
    Camera   CameraConfig   // Mock, RetryCount
    Image    ImageConfig    // PreviewWidth, ThumbWidth, Quality
    Booth    BoothConfig    // CountdownSeconds, PreviewDisplaySeconds
    WiFi     WiFiConfig     // SSID, Password, IP, Interface, Enabled
    Storage  StorageConfig  // PhotosDir
}
```

---

### `internal/camera/gphoto.go` – Kamera-Steuerung

**Controller mit Mutex-geschütztem Capture:**

| Methode | Beschreibung |
|---|---|
| `Capture()` | Löst Foto aus und downloadet es, gibt Dateipfad zurück |
| `GetInfo()` | Gibt `CameraInfo` Struct zurück (Modell, Akku, Speicher, etc.) |

**CameraInfo Struct:**
```go
type CameraInfo struct {
    Connected    bool   `json:"connected"`
    Model        string `json:"model"`
    Manufacturer string `json:"manufacturer"`
    SerialNumber string `json:"serialNumber"`
    LensName     string `json:"lensName"`
    BatteryLevel string `json:"batteryLevel"`
    StorageTotal string `json:"storageTotal"`
    StorageFree  string `json:"storageFree"`
}
```

**Capture-Ablauf:**
1. Mutex setzen (`busy = true`)
2. Dateiname generieren: `IMG_YYYYMMDD_HHMMSS.jpg`
3. `gphoto2 --capture-image-and-download --force-overwrite --filename <path>` ausführen
4. Prüfen ob Datei existiert
5. Mutex lösen

**GetInfo-Ablauf:**
1. `gphoto2 --summary` → Parst Model, Manufacturer, Serial Number, Lens Name, Battery Level
2. `gphoto2 --storage-info` → Parst TotalCapacity, Free
3. Bei Fehler: `Connected = false`, leere Felder

**Mock-Modus:**
- `Capture()`: Erstellt eine Dummy-Datei, 1s simulierte Verzögerung
- `GetInfo()`: Gibt statische Dummy-Daten zurück ("Canon EOS 700D (Mock)", 75% Akku, etc.)

---

### `internal/imaging/` – Bildverarbeitung

**Pure Go Implementierung mit `disintegration/imaging`:**

| Output | Breite | Format | Zweck |
|---|---|---|---|
| Preview | 800px (config) | JPEG 80% | WLAN-taugliche Vorschau |
| Thumbnail | 200px (config) | JPEG 70% | Galerie-Übersicht |

**Verhalten:**
- `Lanczos` Resampling für hohe Bildqualität
- EXIF-Rotation wird automatisch korrigiert
- Keine cgo-Abhängigkeiten

---

### `internal/storage/` – Foto-Verwaltung

| Funktion | Beschreibung |
|---|---|
| `ListPhotos()` | Alle Fotos, sortiert nach Datum (neueste zuerst) |
| `GetLatestPhoto()` | Das neueste Foto mit URLs |
| `DeletePhoto(filename)` | Löscht Original + Preview + Thumbnail |

**PhotoEntry Format:**
```go
type Photo struct {
    Filename  string `json:"filename"`
    Timestamp string `json:"timestamp"`
    URL       string `json:"url"`
    ThumbURL  string `json:"thumbUrl"`
}
```

---

### `internal/websocket/` – WebSocket Hub

**Hub verwaltet alle Client-Verbindungen:**

| Feature | Beschreibung |
|---|---|
| `Broadcast` Channel | Sendet Events an alle Clients |
| `Register` / `Unregister` | Client an-/abmelden |
| `ClientCount()` | Anzahl verbundener Clients |
| `Run()` | Goroutine: Empfängt Broadcasts + neue Clients |

**Event-Typen:**
```go
const (
    EventTypeStatus    = "status"
    EventTypeCountdown = "countdown"
    EventTypePhotoReady = "photo_ready"
    EventTypeError     = "error"
    EventTypeLog       = "log"
)
```

---

### `internal/app/controller.go` – App Controller

**Zentrale Steuerlogik mit State Machine:**

| Methode | Beschreibung |
|---|---|
| `HandleTrigger()` | Startet Capture-Sequenz (Countdown → Capture → Process → Preview) |
| `SetState(state)` | Setzt State + Broadcast an alle Clients |
| `GetUptime()` | Formatierter Uptime-String (HH:MM:SS oder MM:SS) |
| `GetState()` | Aktueller State als String |

**States:** `idle`, `countdown`, `capturing`, `processing`, `preview`, `error`

---

### `internal/logging/` – Strukturiertes Logging

**Ring-Buffer Logger mit WebSocket Broadcast:**

| Feature | Beschreibung |
|---|---|
| Max Einträge | 500 (älteste werden bei Überlauf gelöscht) |
| Levels | `info`, `warn`, `error`, `debug` |
| Sources | `app`, `camera`, `imaging`, `wifi`, `dns`, etc. |
| Broadcast | Jeder Log-Eintrag wird sofort per WebSocket gesendet |
| REST | `GET /api/logs?limit=100` für bestehende Logs |

**Log-Entry Format:**
```go
type Entry struct {
    Level     Level  `json:"level"`
    Message   string `json:"message"`
    Source    string `json:"source"`
    Timestamp int64  `json:"timestamp"`
}
```

---

### `internal/dns/` – Captive Portal DNS

**Eigener DNS Server in Go (`miekg/dns`):**
- Beantwortet alle DNS-Anfragen mit der Pi-IP
- Kein dnsmasq nötig
- Läuft als Goroutine auf Port 53
- Braucht Root-Rechte (oder `CAP_NET_BIND_SERVICE`)

---

### `internal/network/` – WiFi Manager

**nmcli Wrapper für Hotspot-Konfiguration:**

| Funktion | Beschreibung |
|---|---|
| `EnsureHotspot()` | Erstellt/aktiviert den Hotspot |
| Connection prüfen | Sucht nach bestehender `photobooth-ap` Connection |
| Erstellen | `nmcli connection add type wifi ...` |
| Aktivieren | `nmcli connection up photobooth-ap` |

---

### `internal/api/handler.go` – REST API

| Methode | Route | Beschreibung |
|---|---|---|
| `GET` | `/api/status` | State, Clients, Uptime, CameraInfo |
| `POST` | `/api/trigger` | Capture auslösen |
| `GET` | `/api/photos` | Foto-Liste |
| `GET` | `/api/photos/latest` | Letztes Foto |
| `GET` | `/api/logs` | Server-Logs (Ring-Buffer, `?limit=N`) |
| `GET` | `/api/legacy/poll` | Kombinierter Status für Legacy-Client |

---

## Event-Flow & Zustandsmaschine

### Zustandsdiagramm

```
           trigger
    ┌────────────────────┐
    │                    ▼
  IDLE              COUNTDOWN
    ▲               (n → 0)
    │                    │
    │                    ▼
    │              CAPTURING
    │               (gphoto2)
    │                    │
    │                    ▼
    │              PROCESSING
    │               (imaging)
    │                    │
    │                    ▼
    │               PREVIEW
    │            (X Sekunden)
    │                    │
    └────────────────────┘
```

### Detaillierter Ablauf mit Logging

```
1. Trigger empfangen (WebSocket oder REST)
   ├─ Log: [info] [app] "Capture sequence triggered"
   ├─ State: COUNTDOWN
   ├─ Broadcast: countdown { remaining: 5, total: 5 }
   └─ … (jede Sekunde + Log)

2. Kamera auslösen
   ├─ Log: [info] [app] "Capturing photo..."
   ├─ State: CAPTURING
   └─ gphoto2 --capture-image-and-download

3. Bild verarbeiten
   ├─ Log: [info] [app] "Processing image: IMG_xxx.jpg"
   ├─ State: PROCESSING
   └─ imaging: Original → Preview + Thumbnail

4. Vorschau zeigen
   ├─ Log: [info] [app] "Photo ready: IMG_xxx.jpg"
   ├─ State: PREVIEW
   └─ Timer: previewDisplaySeconds (8s default)

5. Zurück zu Idle
   └─ Log: [info] [app] "Returning to idle"
```

### Fehlerbehandlung

Bei Fehler in Capture/Processing:
- `error` Event an alle Clients
- Log: `[error] [app] "Capture failed: ..."`
- State zurück zu `IDLE` nach 3 Sekunden

---

## Frontend Architektur

### Pinia Stores

**`photobooth` Store:**
- `state` – Aktueller Zustand (idle, countdown, etc.)
- `countdown` – `{ remaining, total }` für Countdown-Anzeige
- `lastPhoto` – Letztes aufgenommenes Foto (`{ url, thumbUrl }`)
- `clients` – Anzahl verbundener Clients
- `uptime` – Server Uptime (formatierter String)
- `logs` – Array von Log-Einträgen (max 200 im Frontend)
- `cameraInfo` – Kamera-Infos (Modell, Akku, Speicher, etc.)
- `connected` – WebSocket Verbindungsstatus
- `trigger()` – Foto auslösen via WebSocket
- `fetchStatus()` – Status + CameraInfo laden (alle 5s)
- `fetchLogs()` – Bestehende Logs laden

**`clientMode` Store:**
- `selectedMode` – Gewählter Client-Modus
- `modeDefinition` – Modus-Definition mit Capability-Flags
- `selectMode(id)` – Modus setzen
- `clearMode()` – Modus zurücksetzen

**`gallery` Store:**
- `photos` – Array von PhotoEntry
- `loading` – Lade-Status
- `fetchPhotos()` – Fotos laden

### Composables

**`useFullscreen()`:**
- `enterFullscreen(el?)` – Fullscreen aktivieren (Cross-Browser)
- `exitFullscreen()` – Fullscreen verlassen
- `isFullscreen` – Reaktiver Zustand
- Unterstützt: standard, webkit, moz, ms Prefixes

**`useExitLock()`:**
- `tap()` – Tap registrieren (10 benötigt)
- `lock()` – Lock wieder aktivieren
- `tapCount` – Aktueller Zähler
- `isUnlocked` – Ob das Overlay angezeigt wird
- 3 Sekunden Inaktivitäts-Timeout für Reset

### Views & Routen

| Route | View | Beschreibung |
|---|---|---|
| `/` | DashboardView | Admin-Panel: Status, Trigger, Kamera-Info, Log-Viewer |
| `/modes` | ModeSelectView | Client-Modus wählen (6 Modi) |
| `/client` | ClientView | Composite View – zeigt gewählten Modus |
| `/gallery` | GalleryView | Foto-Raster |
| `/buzzer` | → `/modes` | Legacy-Redirect |
| `/countdown` | → `/modes` | Legacy-Redirect |
| `/preview` | → `/modes` | Legacy-Redirect |

### Dashboard-Karten

**State Card:** Farbiger Indikator + State-Label  
**Trigger Card:** Capture-Button (disabled wenn nicht idle)  
**Last Photo Card:** Thumbnail-Vorschau des letzten Fotos  
**Camera Card:**
- Verbindungsstatus (grün/rot Punkt)
- Modellname, Hersteller, Objektiv
- Akku-Füllstand mit Farbbalken (grün >50%, gelb >20%, rot ≤20%)
- Freier Speicherplatz

---

## Client-Mode-System

### Architektur

```
/modes (ModeSelectView)
  ├─ Modus wählen → Fullscreen aktivieren
  └─ Router: /client

/client (ClientView)
  ├─ Exit-Lock aktiv (10 Taps zum Entsperren)
  ├─ Composite View basierend auf Mode-Flags:
  │   ├─ hasBuzzer → Touch-Auslöser im Idle-State
  │   ├─ hasCountdown → SVG Ring + Countdown-Zahl
  │   ├─ hasPreview → Foto-Anzeige im Preview-State
  │   └─ hasGallery → Grid-Ansicht aller Fotos
  └─ Exit-Lock Overlay:
      ├─ "Modus wählen" → /modes
      ├─ "Client Modus verlassen" → / (Dashboard)
      └─ "Abbrechen" → Overlay schließen
```

### Mode-Definitionen

```typescript
type ClientMode =
  | 'buzzer-countdown-preview'  // Vollständig
  | 'buzzer-countdown'          // Ohne Bildvorschau
  | 'countdown-preview'         // Monitor + Vorschau (kein Auslöser)
  | 'preview-only'              // Nur Bildanzeige
  | 'countdown-only'            // Nur Countdown (kein Auslöser)
  | 'gallery'                   // Galerie-Ansicht
```

### Verhalten pro State

| State | hasBuzzer | !hasBuzzer + hasCountdown | hasPreview only |
|---|---|---|---|
| `idle` | Touch-Auslöser (Kamera-Icon) | "Warte auf Auslöser" | Letztes Foto |
| `countdown` | SVG Ring + Zahl | SVG Ring + Zahl | Pulsierender Punkt |
| `capturing` | Weißer Ping-Punkt | Weißer Ping-Punkt | Weißer Ping-Punkt |
| `processing` | Spinner | Spinner | Spinner |
| `preview` | Foto (wenn hasPreview) oder ✓ | Foto oder ✓ | Foto |

---

## Legacy Client

### Zielgruppe
- iPad Air 1 (Safari 9)
- iPad 2/3/4 (Safari 9-10)
- Ältere Android-Tablets

### Technische Einschränkungen
- Nur ES5 JavaScript (`var`, `function`, `XMLHttpRequest`)
- Kein WebSocket → Polling alle 500ms (`GET /api/legacy/poll`)
- Einfaches CSS (kein Flexbox `gap`, kein Grid, keine Custom Properties)

### Seiten-Struktur
- `index.html` – Startseite mit Buzzer + Countdown + Preview in einem
- Serviert unter `/legacy/`

---

## Netzwerk & WLAN

### NetworkManager Hotspot (empfohlen)

```bash
# Connection erstellen
nmcli connection add type wifi ifname wlan0 con-name photobooth-ap \
  autoconnect yes ssid "Photobooth"

# Als AP konfigurieren
nmcli connection modify photobooth-ap \
  802-11-wireless.mode ap \
  802-11-wireless.band bg \
  ipv4.addresses 192.168.4.1/24 \
  ipv4.method shared

# Aktivieren
nmcli connection up photobooth-ap
```

> **Wichtig:** `dnsmasq-base` muss installiert sein für `ipv4.method shared` (DHCP).

### Captive Portal (eingebaut)

Der DNS Server ist direkt in Go implementiert (`internal/dns/`):
- Alle DNS-Anfragen → Pi-IP (`192.168.4.1`)
- Kein `dnsmasq` für DNS nötig
- Läuft auf Port 53 (braucht Root)

---

## Logging-System

### Backend

**Ring-Buffer Logger** (`internal/logging/`):
- Bis zu 500 Einträge im Speicher
- Threadsafe (Mutex-geschützt)
- Broadcast-Funktion: Neue Einträge werden sofort per WebSocket an alle Clients gesendet

**Log-Quellen:**
- `app` – Capture-Sequenz, State-Changes
- `camera` – gphoto2 Aufrufe
- `wifi` – Hotspot-Status
- `dns` – DNS Server
- `main` – Startup-Meldungen

### Frontend

**Dashboard Log-Viewer:**
- Monospace-Font, Unix-Terminal-Style
- Farbcodiert nach Level (Info=grün, Warn=gelb, Error=rot, Debug=grau)
- Auto-Scroll (ein/ausschaltbar)
- Zeigt Timestamp, Level, Source, Message
- Initial: Bestehende Logs via `GET /api/logs`
- Live: Neue Logs via WebSocket `log` Events

---

## Build & Deployment

### Build Script (`scripts/build-pi.sh`)

```bash
# 1. Frontend bauen (Vite → statische Dateien)
cd frontend && npm run build

# 2. Backend Cross-Compilieren (Go → ARM64)
GOOS=linux GOARCH=arm64 go build -o dist/photobooth backend/cmd/server/main.go

# 3. Assets kopieren (Legacy Client, Config, Scripts, Public)
cp -r legacy/ dist/legacy/
cp -r public/ dist/public/
cp config.json dist/
```

### Deploy Script (`scripts/deploy.sh`)

**SSH ControlMaster** – nur einmal Passwort eingeben:

```bash
./scripts/deploy.sh pi@192.168.x.x
```

**Ablauf:**
1. SSH ControlMaster-Verbindung herstellen (einmalige Passwort-Eingabe)
2. `build-pi.sh` ausführen
3. Dateien per `rsync` übertragen (ohne `data/`)
4. `install.sh` auf dem Pi ausführen
5. ControlSocket automatisch aufräumen bei Exit

**Kein `sshpass` nötig** – ControlMaster nutzt eine einzige SSH-Session für alle Befehle.

### Install Script (`scripts/install.sh`)

Auf dem Pi ausgeführt:
1. `gphoto2` und `dnsmasq-base` installieren (falls fehlend)
2. Photobooth-Verzeichnisse erstellen
3. systemd Service installieren und starten

### systemd Service

```ini
[Unit]
Description=Photobooth V5
After=network.target NetworkManager.service

[Service]
Type=simple
User=pi
WorkingDirectory=/opt/photobooth
ExecStart=/opt/photobooth/photobooth
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

---

## Bekannte Limitierungen

### Performance (Pi 3)
- Bildverarbeitung (pure Go): ~200-400ms pro Foto
- WebSocket Broadcast: <5ms für alle Clients
- Gesamte Capture-to-Preview: ~1-3 Sekunden (exkl. Countdown)

### Captive Portal
- Nicht 100% zuverlässig auf allen Geräten
- Manche Android-Versionen zeigen ein kleines Popup statt vollem Browser
- Desktop-Browser reagieren unterschiedlich
- DNS Server braucht Port 53 (Root oder CAP_NET_BIND_SERVICE)

### Gleichzeitige Auslöser
- Nur ein Capture gleichzeitig möglich (Mutex)
- Zweiter Trigger während Countdown/Capture wird abgelehnt
- Clients im Buzzer-Modus: Button ist disabled wenn State ≠ idle

### Legacy Client
- Kein Echtzeit-Countdown (500ms Polling-Delay)
- Einfacheres UI (kein Tailwind, keine Animationen)
- Kein Client-Mode-System (fest: Buzzer + Countdown + Preview)


# Important
*** CRITICAL *** Build steps and install scripts will be manually executed by the user