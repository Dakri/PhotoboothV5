# PhotoboothV5 – Technische Dokumentation

> Referenz-Dokument für die gesamte Entwicklung.  
> Hier sind alle Module, Abläufe, Schnittstellen und Design-Entscheidungen dokumentiert.

---

## Inhaltsverzeichnis

1. [Design-Prinzipien](#design-prinzipien)
2. [Backend Module](#backend-module)
3. [Event-Flow & Zustandsmaschine](#event-flow--zustandsmaschine)
4. [Frontend Architektur](#frontend-architektur)
5. [Legacy Client](#legacy-client)
6. [Netzwerk & WLAN](#netzwerk--wlan)
7. [Authentifizierung](#authentifizierung)
8. [Dateistruktur & Namenskonventionen](#dateistruktur--namenskonventionen)
9. [Build & Deployment](#build--deployment)
10. [Bekannte Limitierungen](#bekannte-limitierungen)

---

## Design-Prinzipien

### 1. Standalone – Keine externen Abhängigkeiten zur Laufzeit
- Kein Internet nötig
- Kein Docker, kein externer Dienst
- Alles läuft auf dem Pi alleine

### 2. Minimale Installation
- `npm install` + `npm run build` + `install.sh`
- Keine langen Setup-Scripts die apt-Pakete installieren
- Vorbedingung: Node.js + gphoto2 (2 Pakete per apt)

### 3. Graceful Degradation
- Kamera nicht da? → Mock-Modus oder klare Fehlermeldung
- WLAN nicht konfigurierbar? → Klare Warnung, manuelle Anleitung
- Alter Browser? → Legacy-Client über `/legacy/`
- WebSocket nicht verfügbar? → REST Polling

### 4. Alles konfigurierbar
- `config/default.json` = Basis-Einstellungen
- `data/config.json` = Laufzeit-Überschreibungen (via Dashboard/API)
- Kein Hardcoded-Wert im Code

---

## Backend Module

### `config.ts` – Konfigurationsmanagement

**Aufgabe:** Zentrale Konfiguration laden, mergen, und bereitstellen.

**Verhalten:**
1. Lädt `config/default.json` als Basis
2. Falls `data/config.json` existiert → Deep-Merge über Defaults
3. Exportiert `config` Singleton (Readonly zur Laufzeit)

**Wichtige Exports:**
- `config: AppConfig` – Geladene Konfiguration
- `saveRuntimeConfig(overrides)` – Überschreibungen in `data/config.json` speichern
- `getPhotosDir()` – Absoluter Pfad zum Foto-Verzeichnis
- `ensureDataDirs()` – Erstellt `original/`, `preview/`, `thumb/` falls nicht vorhanden

**Deep-Merge Logik:**
```
Default:  { image: { previewWidth: 800, thumbWidth: 200 } }
Override: { image: { previewWidth: 600 } }
Result:   { image: { previewWidth: 600, thumbWidth: 200 } }
```

---

### `camera/gphoto.ts` – Kamera-Steuerung

**Aufgabe:** Canon DSLR via gphoto2 ansteuern.

**Klasse: `CameraController`**

| Methode | Beschreibung |
|---|---|
| `Capture()` | Löst Foto aus und downloadet es, Mutex-geschützt |
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
- Für Entwicklung ohne Kamera

**gphoto2 Befehle:**
```bash
# Kamera erkennen
gphoto2 --auto-detect

# Foto aufnehmen und direkt herunterladen
gphoto2 --capture-image-and-download --force-overwrite --filename /path/to/IMG.jpg

# Kamera-Zusammenfassung (Model, Hersteller, Akku, etc.)
gphoto2 --summary

# Speicherinfo (Kapazität, freier Platz)
gphoto2 --storage-info

# Kamera-Einstellungen lesen
gphoto2 --list-config
```

---

### `image/processor.ts` – Bildverarbeitung

**Aufgabe:** Originalfotos in Preview und Thumbnail umwandeln.

**Funktion: `processPhoto(originalPath, filename)`**

| Output | Breite | Qualität | Zweck |
|---|---|---|---|
| Preview | 800px (config) | 80% (config) | WLAN-taugliche Vorschau |
| Thumbnail | 200px (config) | 70% (config) | Galerie-Übersicht |

**Verhalten:**
- Preview und Thumbnail werden **parallel** generiert (`Promise.all`)
- EXIF-Rotation wird automatisch korrigiert (`.rotate()`)
- `withoutEnlargement: true` – Kleine Bilder werden nicht hochskaliert

**Performance auf Pi 3:**
- ~150-250ms pro Foto (Preview + Thumbnail)
- sharp nutzt libvips mit ARM NEON Optimierungen

---

### `storage/photos.ts` – Foto-Verwaltung

**Aufgabe:** CRUD-Operationen für gespeicherte Fotos.

| Funktion | Beschreibung |
|---|---|
| `listPhotos()` | Alle Fotos, sortiert nach Datum (neueste zuerst) |
| `getPhoto(filename)` | Ein einzelnes Foto mit URLs |
| `getLatestPhoto()` | Das neueste Foto |
| `deletePhoto(filename)` | Löscht Original + Preview + Thumbnail |
| `getStorageStats()` | Anzahl Fotos + Gesamtgröße in MB |

**PhotoEntry Format:**
```typescript
{
  id: "IMG_20260219_143022",
  filename: "IMG_20260219_143022.jpg",
  timestamp: 1739972422000,
  urls: {
    original: "/photos/original/IMG_20260219_143022.jpg",
    preview: "/photos/preview/IMG_20260219_143022.jpg",
    thumbnail: "/photos/thumb/IMG_20260219_143022.jpg"
  }
}
```

---

### `storage/usb.ts` – USB-Export

**Aufgabe:** USB-Sticks erkennen und Fotos exportieren.

**Erkennung:** `lsblk -J -o NAME,MOUNTPOINT,LABEL,HOTPLUG,TYPE`
- Filtert nach `hotplug: true` + `type: "part"` + gemountet

**Export:**
- Erstellt `Photobooth_Export/` auf dem Stick
- Kopiert nur neue Dateien (Skip wenn bereits vorhanden)
- Gibt Statistik zurück: `{ copied, skipped, errors }`

---

### `websocket/events.ts` – Event-System

**Aufgabe:** Typdefinitionen und Message-Factories für alle WebSocket Events.

**Event-Typen:** `trigger`, `countdown`, `capturing`, `processing`, `photo_ready`, `error`, `status`, `register`, `clients_update`, `ping`, `pong`

**Client-Rollen:** `dashboard`, `buzzer`, `display`, `gallery`, `hardware`

**Message-Format:**
```typescript
{
  type: "countdown",
  data: { remaining: 3, total: 5 },
  timestamp: 1739972422000
}
```

---

### `websocket/server.ts` – WebSocket Server

**Aufgabe:** Client-Verbindungen verwalten und den Capture-Flow orchestrieren.

**Klasse: `PhotoboothWsServer`**

| Methode | Beschreibung |
|---|---|
| `attach(httpServer)` | WS Server an HTTP Server anhängen (Pfad: `/ws`) |
| `handleTrigger()` | Capture-Flow starten |
| `getState()` | Aktueller Zustand der Photobooth |
| `getLastPhoto()` | Letztes aufgenommenes Foto |
| `getConnectedClients()` | Verbundene Clients nach Rolle |
| `close()` | Sauberes Herunterfahren |

**Heartbeat:** Alle 30 Sekunden `ping` an alle Clients.

**State Change Callback:** `onStateChange` wird für die REST Legacy API genutzt.

---

### `api/routes.ts` – REST API

**Aufgabe:** HTTP Endpoints für Frontend und Legacy-Client.

**Alle Routen beginnen mit `/api/`**

**Legacy-Polling:** `GET /api/legacy/poll` gibt State + letztes Foto zurück. Client pollt alle 500ms.

**Paginierung:** `GET /api/photos?page=1&limit=50`

---

### `network/wifi.ts` – WLAN Hotspot

**Aufgabe:** WLAN Access Point über NetworkManager einrichten.

**Ablauf:**
1. Prüfen ob `nmcli` verfügbar ist
2. Prüfen ob Connection `photobooth-ap` bereits existiert
3. Falls nicht: Connection erstellen + konfigurieren (AP Mode, IP, DHCP)
4. Connection aktivieren

**Fallback:** Wenn NetworkManager nicht vorhanden → Warnung loggen, Status `'unavailable'` zurückgeben.

**nmcli Connection-Name:** `photobooth-ap` (fest)

---

### `network/captive.ts` – Captive Portal

**Aufgabe:** DNS-Redirect über dnsmasq für automatisches Dashboard-Öffnen.

**Konfigurationsdatei:** `/etc/dnsmasq.d/photobooth-captive.conf`

**Inhalt:**
```
address=/#/192.168.4.1
interface=wlan0
no-resolv
```

**Bedeutung:** Alle DNS-Anfragen (`#` = Wildcard) werden auf `192.168.4.1` aufgelöst → Browser zeigt Dashboard.

---

### `auth/session.ts` – Authentifizierung

**Aufgabe:** PIN-basierter Zugangsschutz für das Dashboard.

**Mechanismus:**
1. Client sendet PIN via `POST /api/auth/login`
2. Server prüft gegen `config.auth.dashboardPin`
3. Bei Erfolg: Session-ID als HTTP-Only Cookie setzen
4. Cookie Gültigkeit: 24 Stunden

**Kein Schutz nötig für:** Buzzer, Countdown, Gallery (konfigurierbar)

**Middleware:** `requireAuth()` – kann in beliebige Routen eingehängt werden

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
    │               (sharp)
    │                    │
    │                    ▼
    │               PREVIEW
    │            (X Sekunden)
    │                    │
    └────────────────────┘
```

### Detaillierter Ablauf

```
1. Trigger empfangen (WebSocket oder REST)
   └─ State: COUNTDOWN
   └─ Broadcast: countdown { remaining: 5, total: 5 }
   └─ … (jede Sekunde)
   └─ Broadcast: countdown { remaining: 0, total: 5 }

2. Kamera auslösen
   └─ State: CAPTURING
   └─ gphoto2 --capture-image-and-download
   └─ Foto gespeichert: data/photos/original/IMG_xxx.jpg

3. Bild verarbeiten
   └─ State: PROCESSING
   └─ sharp: Original → Preview (800px) + Thumbnail (200px)

4. Vorschau zeigen
   └─ State: PREVIEW
   └─ Broadcast: photo_ready { filename, urls }
   └─ Timer: previewDisplaySeconds (8s default)

5. Zurück zu Idle
   └─ State: IDLE
   └─ Bereit für nächsten Trigger
```

### Fehlerbehandlung

Bei Fehler in Schritt 2-3:
- `error` Event an alle Clients senden
- Zurück zu `IDLE`
- Log auf Server-Konsole

---

## Frontend Architektur

### Pinia Stores

**`photobooth` Store:**
- `state` – Aktueller Zustand (idle, countdown, etc.)
- `countdown` – `{ remaining, total }` für Countdown
- `lastPhoto` – Letztes aufgenommenes Foto
- `clients` – Anzahl verbundener Clients
- `uptime` – Server Uptime
- `logs` – Array von Log-Einträgen (max 200)
- `cameraInfo` – `CameraInfo` Objekt (connected, model, manufacturer, lensName, batteryLevel, storageFree etc.)
- `trigger()` – Foto auslösen
- `fetchStatus()` – Status inkl. Kamera-Info vom Server laden (alle 5s per Polling)
- `fetchLogs()` – Bestehende Logs laden

**`gallery` Store:**
- `photos` – Array von PhotoEntry
- `total` – Gesamtanzahl
- `loading` – Lade-Status
- `fetchPhotos(page)` – Fotos laden
- `deletePhoto(filename)` – Foto löschen

### Composables

**`useWebSocket()`:**
- Auto-Connect beim Mount
- Auto-Reconnect bei Verbindungsverlust (mit Backoff)
- Event-Dispatching an Pinia Stores
- `send(type, data)` – Nachricht senden
- `connected` – Ref<boolean>

**`useCountdown()`:**
- Reactive Countdown basierend auf Store-Daten
- Beep-Sound bei Tick (optional)

### Views & Routen

| Route | View | Beschreibung |
|---|---|---|
| `/` | DashboardView | Admin-Panel: Status, Trigger, letztes Foto, Kamera-Info, Log-Viewer |
| `/buzzer` | BuzzerView | Großer Touch-Button zum Auslösen |
| `/countdown` | CountdownView | Vollbild-Countdown mit Animation |
| `/preview` | PreviewView | Foto-Vorschau nach Aufnahme |
| `/gallery` | GalleryView | Foto-Raster mit Lightbox |

**Dashboard Camera-Karte:**
- Zeigt Verbindungsstatus (grün/rot), Modellname, Hersteller, Objektiv
- Akku-Füllstand mit farbigem Balken (grün >50%, gelb >20%, rot ≤20%)
- Freier Speicherplatz auf der SD-Karte
- Fallback: "Keine Kamera erkannt" wenn `cameraInfo.connected === false`
- Wird alle 5 Sekunden per Status-Polling aktualisiert

### Client-Rollen-System

Jede View registriert sich beim WebSocket-Server mit einer Rolle:

```javascript
// In BuzzerView.vue
onMounted(() => {
  ws.send('register', { role: 'buzzer' })
})
```

Das Dashboard zeigt an welche Client-Rollen verbunden sind.

---

## Legacy Client

### Zielgruppe
- iPad Air 1 (Safari 9)
- iPad 2/3/4 (Safari 9-10)
- Ältere Android-Tablets
- Jeder Browser ohne ES6/Proxy Support

### Technische Einschränkungen
- **Kein** `const`/`let` → nur `var`
- **Kein** `async`/`await` → Callbacks oder `.then()`
- **Kein** Template Literals → String-Konkatenation
- **Kein** `fetch` → `XMLHttpRequest`
- **Kein** Arrow Functions → `function(){}`
- **Kein** `class` → Prototyp-basiert oder Objekt-Literal
- **Kein** CSS Custom Properties
- **Kein** Flexbox `gap`
- **Kein** CSS Grid (Safari 9)
- **Kein** WebSocket (wird nicht genutzt, Polling stattdessen)

### Polling-Mechanismus

```javascript
// Legacy: Alle 500ms Status abfragen
var POLL_INTERVAL = 500;

function pollStatus() {
  var xhr = new XMLHttpRequest();
  xhr.open('GET', '/api/legacy/poll');
  xhr.onload = function() {
    if (xhr.status === 200) {
      var data = JSON.parse(xhr.responseText);
      updateUI(data.state, data.lastPhoto);
    }
    setTimeout(pollStatus, POLL_INTERVAL);
  };
  xhr.onerror = function() {
    setTimeout(pollStatus, POLL_INTERVAL * 2);
  };
  xhr.send();
}
```

### Seiten-Struktur
- `index.html` – Startseite mit Buzzer + Countdown + Preview in einem
- `gallery.html` – Einfache Galerie mit Thumbnail-Grid

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

# Offenes WLAN (kein Passwort)
nmcli connection modify photobooth-ap remove 802-11-wireless-security

# Aktivieren
nmcli connection up photobooth-ap
```

### Manuelles Setup (hostapd + dnsmasq)

Falls NetworkManager nicht verfügbar ist:

**1. hostapd installieren:**
```bash
sudo apt install hostapd dnsmasq
```

**2. `/etc/hostapd/hostapd.conf`:**
```
interface=wlan0
driver=nl80211
ssid=Photobooth
hw_mode=g
channel=7
wmm_enabled=0
macaddr_acl=0
auth_algs=1
wpa=0
```

**3. `/etc/dnsmasq.conf`:**
```
interface=wlan0
dhcp-range=192.168.4.10,192.168.4.100,255.255.255.0,24h
address=/#/192.168.4.1
```

**4. Statische IP in `/etc/dhcpcd.conf`:**
```
interface wlan0
static ip_address=192.168.4.1/24
nohook wpa_supplicant
```

**5. Services starten:**
```bash
sudo systemctl unmask hostapd
sudo systemctl enable hostapd dnsmasq
sudo reboot
```

---

## Authentifizierung

### Konzept

Da das WLAN offen ist (keine User-Hürde), findet die Zugriffskontrolle in der Software statt:

| Bereich | Auth | Grund |
|---|---|---|
| Buzzer | Optional | Gäste sollen einfach auslösen können |
| Countdown | Nein | Passive Anzeige |
| Gallery | Optional | Gäste sollen Fotos sehen können |
| Dashboard | Ja (PIN) | Admin-Funktionen schützen |
| Config API | Ja (PIN) | Einstellungen schützen |
| Foto löschen | Ja (PIN) | Versehentliches Löschen verhindern |
| USB Export | Ja (PIN) | Physische Aktion bestätigen |

### Session-Flow

```
Client                          Server
  │                               │
  │  POST /api/auth/login         │
  │  { pin: "1234" }              │
  │──────────────────────────────>│
  │                               │ PIN prüfen
  │  Set-Cookie: pb_session=xxx   │
  │<──────────────────────────────│
  │                               │
  │  GET /api/config              │
  │  Cookie: pb_session=xxx       │
  │──────────────────────────────>│
  │                               │ Session prüfen → OK
  │  { config... }                │
  │<──────────────────────────────│
```

---

## Dateistruktur & Namenskonventionen

### Fotos

```
data/photos/
├── original/IMG_20260219_143022.jpg    # Originalgröße (von Kamera)
├── preview/IMG_20260219_143022.jpg     # 800px breit, 80% JPEG
└── thumb/IMG_20260219_143022.jpg       # 200px breit, 70% JPEG
```

**Dateiname-Format:** `IMG_YYYYMMDD_HHMMSS.jpg`

### Code-Konventionen

- **Backend:** TypeScript, ES Modules (`import/export`), `.ts` Dateien
- **Frontend:** Vue 3 Composition API (`<script setup>`), TypeScript
- **Legacy:** Vanilla JavaScript, ES5, IIFE-Pattern
- **CSS (Frontend):** Tailwind 4 Utility Classes
- **CSS (Legacy):** Einfaches CSS, keine Custom Properties, keine modernen Features

---

## Build & Deployment

### Build-Reihenfolge

```bash
# 1. Frontend bauen (Vite → statische Dateien)
cd frontend && npm run build
#    Output: dist/frontend/

# 2. Backend kompilieren (Go → Binary, Cross-Compile für Pi)
#    Siehe scripts/build-pi.sh
GOOS=linux GOARCH=arm GOARM=7 go build -o dist/photobooth backend/cmd/server/main.go

# 3. Legacy-Client – kein Build nötig
#    (statische Dateien, direkt serviert)
```

### Deploy Script (`scripts/deploy.sh`)

**Verwendet SSH ControlMaster** für eine einmalige Passwort-Eingabe über alle SSH/rsync-Verbindungen hinweg.

```bash
./scripts/deploy.sh pi@192.168.4.1
```

**Ablauf:**
1. SSH ControlMaster-Verbindung herstellen (einmalige Passwort-Eingabe)
2. `build-pi.sh` ausführen (Frontend + Backend kompilieren)
3. Dateien per `rsync` zum Pi übertragen (ohne `data/` Ordner)
4. `install.sh` auf dem Pi ausführen (systemd Service einrichten)
5. ControlSocket automatisch aufräumen bei Exit

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

**Service Befehle:**
```bash
sudo systemctl start photobooth      # Starten
sudo systemctl stop photobooth       # Stoppen
sudo systemctl restart photobooth    # Neustarten
sudo systemctl status photobooth     # Status
sudo journalctl -u photobooth -f     # Live-Logs
```

---

## Bekannte Limitierungen

### Performance (Pi 3)
- Bildverarbeitung: ~150-250ms pro Foto
- WebSocket Broadcast: <5ms für alle Clients
- Gesamte Capture-to-Preview: ~1-2 Sekunden (exkl. Countdown)

### Captive Portal
- Nicht 100% zuverlässig auf allen Geräten
- Manche Android-Versionen zeigen ein kleines Popup statt vollem Browser
- Desktop-Browser (Windows/Mac) reagieren unterschiedlich
- **Workaround:** SSID enthält die IP als Hinweis

### Foto-Download
- Fotos werden direkt heruntergeladen (nicht auf SD gespeichert)
- Bei USB-Fehler während gphoto2-Capture → Foto verloren
- **Mitigation:** Retry-Logik (konfigurierbar)

### Gleichzeitige Auslöser
- Nur ein Capture gleichzeitig möglich (Mutex)
- Zweiter Trigger während Countdown/Capture wird abgelehnt
- Client erhält `error` Event mit Code `NOT_IDLE`

### Legacy Client
- Kein Echtzeit-Countdown (500ms Polling-Delay)
- Einfacheres UI (kein Tailwind, keine Animationen)
- Kein Preview-Fade, nur Hard-Switch
