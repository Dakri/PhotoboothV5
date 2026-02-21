# PhotoboothV5

> Die professionelle Photobooth-Lösung für den Raspberry Pi.  
> Schnell, stabil und bereit für dein nächstes Event.

---

## Was macht die Software?

PhotoboothV5 verwandelt deinen Raspberry Pi in eine vollständige, professionelle Photobooth-Station. Es ist kein einfacher Foto-Viewer, sondern ein komplettes System für Events:

*   **Vollautomatisch**: Kamera anschließen, Pi starten – das System regelt den Rest.
*   **WLAN Hotspot inklusive**: Der Pi eröffnet ein eigenes WLAN. Verbinde dich und die Clients (Tablets, Smartphones) und sieh sofort die Galerie oder steuer die Booth.
*   **Plug & Play Kamera-Support**: Unterstützt gängige Canon DSLR Kameras direkt über USB. Fotos werden in Echtzeit heruntergeladen und verarbeitet. Siehe gphoto2 für unterstützte Modelle: http://www.gphoto.org/proj/libgphoto2/support.php
*   **Interaktive Client-Modi**: Nutze iPads oder Tablets als Auslöser (Buzzer), Countdown-Monitor oder Live-Galerie.
*   **Hochwertige Vorschau**: Fotos werden sofort optimiert und auf allen verbundenen Geräten blitzschnell angezeigt.
*   **USB-Export**: Am Ende des Events einfach einen Stick reinstecken und alle Fotos per Knopfdruck exportieren.
*   **Admin-Dashboard**: Volle Kontrolle über alle Einstellungen, Live-Logs und System-Status über eine moderne Weboberfläche.

---

## Installation

> [!IMPORTANT]
> Das Installationsskript muss auf dem Raspberry Pi heruntergeladen und manuell ausgeführt werden.

1.  **Verzeichnis erstellen und Berechtigungen setzen**
    ```bash
    sudo mkdir -p /opt/photobooth
    sudo chown -R $USER:$USER /opt/photobooth
    cd /opt/photobooth
    ```

2.  **Aktuelles Release herunterladen**
    Lade das neueste Paket von der [Release-Seite](https://github.com/Dakri/PhotoboothV5/releases) herunter:
    ```bash
    # Beispiel (bitte URL vom aktuellsten Release kopieren):
    wget -O pbv5.tar.gz https://github.com/Dakri/PhotoboothV5/releases/latest/download/release.v5.0.0.20260221.tar.gz
    ```

3.  **Archiv entpacken**
    ```bash
    tar -xzf pbv5.tar.gz
    ```

4.  **Installationsskript ausführbar machen und starten**
    ```bash
    chmod +x scripts/install.sh
    sudo ./scripts/install.sh
    ```

---

## Architektur & Tech-Stack

Für die Technik-Begeisterten: PhotoboothV5 ist auf maximale Performance und minimale Abhängigkeiten ausgelegt.

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

| Komponente | Technologie | Warum |
|---|---|---|
| **Backend** | **Go (Golang)** | Single Binary, keine Runtime, performant |
| **Echtzeit** | `gorilla/websocket` | Bewährt, breite Kompatibilität |
| **Frontend** | Vite + Vue 3 + Tailwind 4 | Moderner, reaktiver UI-Stack |
| **Legacy-Client** | Vanilla HTML/CSS/JS | Funktioniert auf extrem alten Tablets (Safari 9+) |
| **Bildverarbeitung** | `disintegration/imaging` | Pure Go, kein cgo nötig |
| **WLAN** | NetworkManager (nmcli) | Ab Pi OS Bookworm Standard |
| **Deployment** | systemd Service | Ein `install.sh`, kein Docker |

---

## Entwicklung & Deployment

### Schnellstart (Cross-Compile & Deploy)

> [!TIP]
> **Voraussetzung**: NodeJs 20+ und Go 1.25+.  
> Wir empfehlen **WSL2** (Windows Subsystem for Linux) für das Deployment von Windows aus.

```bash
# Auf dem Entwicklungs-PC (nicht auf dem Pi!)
./scripts/deploy.sh pi@192.168.x.x
```

Das Deploy-Script baut das Frontend, cross-compiliert das Backend für ARMv8 und überträgt alles per SSH zum Pi.

### Lokale Entwicklung

```bash
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && npm install && npm run dev
```

---

## Lizenz

MIT
