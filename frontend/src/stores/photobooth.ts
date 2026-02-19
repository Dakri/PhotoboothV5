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
    storageTotal: string
    storageFree: string
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
    const uptime = ref('00:00')
    const cameraInfo = ref<CameraInfo>({
        connected: false,
        model: '',
        manufacturer: '',
        serialNumber: '',
        lensName: '',
        batteryLevel: '',
        storageTotal: '',
        storageFree: ''
    })

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
            case 'error':
                error.value = msg.data.message
                state.value = 'error'
                setTimeout(() => { error.value = null }, 5000)
                break
            case 'log':
                addLog(msg.data)
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
            }
        } catch (e) {
            console.error('Failed to fetch status:', e)
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
        connect,
        trigger,
        register,
        fetchLogs,
        fetchStatus,
        addLog
    }
})
