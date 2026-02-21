import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface LogEntry {
    level: string
    message: string
    source: string
    timestamp: number
}

export interface CameraInfo {
    connected: boolean
    model: string
    manufacturer: string
    serialNumber: string
    lensName: string
    batteryLevel: string
    batteryPercent?: number
    storageTotal: string
    storageFree: string
    storagePercent?: number
}

export interface CameraFile {
    name: string
    size: number
}

export interface DiskInfo {
    total: number
    free: number
    used: number
    usedPercent: number
}

export interface BoothSettings {
    countdownSeconds: number
    previewDisplaySeconds: number
    photosBasePath: string
    currentAlbum: string
    captureStrategy: string
    triggerDelayMs: number
}

export interface AlbumInfo {
    id: string
    name: string
    count: number
    size: number
    captureMethod: string
}

export interface UsbDevice {
    name: string
    label: string
    mountpoint: string
    size: string
    subsystems: string
    free?: string
}

export interface UsbExportProgress {
    active: boolean
    album: string
    copiedBytes: number
    totalBytes: number
    copiedFiles: number
    totalFiles: number
    etaSeconds: number
    error?: string
}

export const usePhotoboothStore = defineStore('photobooth', () => {
    // State
    const connected = ref(false)
    const state = ref('idle')
    const countdown = ref({ remaining: 0, total: 0 })
    const lastPhoto = ref<any>(null)
    const error = ref<string | null>(null)
    const clients = ref(0)
    const logs = ref<LogEntry[]>([])
    const cameraFiles = ref<CameraFile[]>([])
    const uptime = ref('00:00')
    const cameraInfo = ref<CameraInfo>({
        connected: false,
        model: '',
        manufacturer: '',
        serialNumber: '',
        lensName: '',
        batteryLevel: '',
        batteryPercent: 0,
        storageTotal: '',
        storageFree: '',
        storagePercent: 0
    })
    const diskInfo = ref<DiskInfo>({
        total: 0,
        free: 0,
        used: 0,
        usedPercent: 0
    })
    const settings = ref<BoothSettings>({
        countdownSeconds: 3,
        previewDisplaySeconds: 5,
        photosBasePath: 'data/photos',
        currentAlbum: 'default',
        captureStrategy: 'C',
        triggerDelayMs: 0
    })
    const albums = ref<AlbumInfo[]>([])
    const usbDevices = ref<UsbDevice[]>([])
    const usbExport = ref<UsbExportProgress>({ active: false, album: '', copiedBytes: 0, totalBytes: 0, copiedFiles: 0, totalFiles: 0, etaSeconds: 0 })

    // WebSocket
    let ws: WebSocket | null = null
    let reconnectTimer: any = null

    function connect() {
        if (ws) return

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host
        const url = `${protocol}//${host}/ws`

        ws = new WebSocket(url)

        ws.onopen = () => {
            connected.value = true
            error.value = null
            if (reconnectTimer) clearInterval(reconnectTimer)
        }

        ws.onclose = () => {
            connected.value = false
            ws = null
            reconnectTimer = setTimeout(() => connect(), 2000)
        }

        ws.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data)
                handleMessage(msg)
            } catch (e) {
                console.error('Failed to parse WS message:', e)
            }
        }
    }

    function handleMessage(msg: any) {
        switch (msg.type) {
            case 'status':
                state.value = msg.data.state
                break
            case 'countdown':
                countdown.value = msg.data
                state.value = 'countdown'
                break
            case 'capturing':
                state.value = 'capturing'
                break
            case 'processing':
                state.value = 'processing'
                break
            case 'photo_ready':
                lastPhoto.value = msg.data
                state.value = 'preview'
                break
            case 'system_info':
                if (msg.data.camera) cameraInfo.value = msg.data.camera
                if (msg.data.disk) diskInfo.value = msg.data.disk
                break
            case 'error':
                error.value = msg.data.message
                state.value = 'error'
                setTimeout(() => { error.value = null }, 5000)
                break
            case 'log':
                addLog(msg.data)
                break
            case 'usb_export_start':
                usbExport.value = { active: true, album: msg.data.album, copiedBytes: 0, totalBytes: 0, copiedFiles: 0, totalFiles: 0, etaSeconds: 0 }
                break
            case 'usb_export_progress':
                usbExport.value = { 
                    active: true, 
                    album: msg.data.album, 
                    copiedBytes: msg.data.copiedBytes,
                    totalBytes: msg.data.totalBytes,
                    copiedFiles: msg.data.copiedFiles,
                    totalFiles: msg.data.totalFiles,
                    etaSeconds: msg.data.etaSeconds,
                }
                break
            case 'usb_export_success':
                usbExport.value.copiedBytes = usbExport.value.totalBytes
                setTimeout(() => {
                    usbExport.value.active = false
                }, 4000)
                break
            case 'usb_export_error':
                usbExport.value.error = msg.data.message
                setTimeout(() => {
                    usbExport.value.active = false
                    usbExport.value.error = undefined
                }, 5000)
                break
        }
    }

    function addLog(entry: LogEntry) {
        logs.value.push(entry)
        // Keep max 200 entries in frontend
        if (logs.value.length > 200) {
            logs.value = logs.value.slice(-200)
        }
    }

    async function fetchLogs() {
        try {
            const res = await fetch('/api/logs?limit=100')
            if (res.ok) {
                logs.value = await res.json()
            }
        } catch (e) {
            console.error('Failed to fetch logs:', e)
        }
    }

    async function fetchStatus() {
        try {
            const res = await fetch('/api/status')
            if (res.ok) {
                const data = await res.json()
                state.value = data.state
                clients.value = data.clients
                uptime.value = data.uptime
                if (data.camera) {
                    cameraInfo.value = data.camera
                }
                if (data.disk) {
                    diskInfo.value = data.disk
                }
                if (data.lastPhoto) {
                    lastPhoto.value = data.lastPhoto
                }
            }
        } catch (e) {
            console.error('Failed to fetch status:', e)
        }
    }

    async function fetchSettings() {
        try {
            const res = await fetch('/api/settings')
            if (res.ok) {
                const data = await res.json()
                if (data.booth) settings.value = data.booth
                if (data.albums) albums.value = data.albums
            }
        } catch (e) {
            console.error('Failed to fetch settings:', e)
        }
    }

    async function saveSettings(s: Partial<BoothSettings & { currentAlbum: string }>) {
        try {
            const res = await fetch('/api/settings', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(s)
            })
            if (res.ok) {
                const data = await res.json()
                if (data.booth) settings.value = data.booth
                if (data.albums) albums.value = data.albums
                return { success: true, data }
            }
            return { success: false, error: 'Request failed' }
        } catch (e) {
            console.error('Failed to save settings:', e)
            return { success: false, error: String(e) }
        }
    }

    function trigger() {
        if (!ws || ws.readyState !== WebSocket.OPEN) return
        ws.send(JSON.stringify({ type: 'trigger' }))
    }

    function register(role: string) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            const interval = setInterval(() => {
                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(JSON.stringify({ type: 'register', data: { role } }))
                    clearInterval(interval)
                }
            }, 500)
            return
        }
        ws.send(JSON.stringify({ type: 'register', data: { role } }))
    }

    async function fetchGalleryCount(album: string) {
        try {
            const res = await fetch(`/api/gallery/count?album=${encodeURIComponent(album)}`)
            const data = await res.json()
            return data.count || 0
        } catch (e) {
            console.error('Failed to fetch gallery count:', e)
            return 0
        }
    }

    async function emptyGallery(album: string) {
        try {
            const res = await fetch(`/api/gallery/empty?album=${encodeURIComponent(album)}`, { method: 'POST' })
            return res.ok
        } catch (e) {
            console.error('Failed to empty gallery:', e)
            return false
        }
    }

    async function deleteGallery(album: string) {
        try {
            const res = await fetch(`/api/gallery/delete?album=${encodeURIComponent(album)}`, { method: 'POST' })
            if (!res.ok) {
                const txt = await res.text()
                throw new Error(txt)
            }
            return { success: true }
        } catch (e) {
            console.error('Failed to delete gallery:', e)
            return { success: false, error: String(e) }
        }
    }

    async function fetchUsbDevices() {
        try {
            const res = await fetch('/api/usb/devices')
            if (res.ok) {
                usbDevices.value = await res.json() || []
            }
        } catch (e) {
            console.error('Failed to fetch USB devices:', e)
        }
    }

    async function fetchCameraFiles() {
        try {
            const res = await fetch('/api/camera/files')
            if (res.ok) {
                cameraFiles.value = await res.json() || []
            }
        } catch (e) {
            console.error('Failed to fetch camera files:', e)
        }
    }

    async function exportToUsb(deviceName: string, albumName: string, copyMode?: string) {
        try {
            const res = await fetch('/api/usb/export', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ deviceName, albumName, copyMode })
            })
            if (!res.ok) {
                const txt = await res.text()
                throw new Error(txt)
            }
            return { success: true }
        } catch (e) {
            console.error('Failed to export to USB:', e)
            return { success: false, error: String(e) }
        }
    }

    return {
        connected,
        state,
        countdown,
        lastPhoto,
        error,
        clients,
        logs,
        uptime,
        cameraInfo,
        diskInfo,
        settings,
        albums,
        usbDevices,
        cameraFiles,
        connect,
        trigger,
        register,
        fetchLogs,
        fetchStatus,
        fetchSettings,
        saveSettings,
        addLog,
        fetchGalleryCount,
        emptyGallery,
        deleteGallery,
        fetchUsbDevices,
        fetchCameraFiles,
        exportToUsb,
        usbExport
    }
})
