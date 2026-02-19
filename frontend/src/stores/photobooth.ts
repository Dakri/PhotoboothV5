import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const usePhotoboothStore = defineStore('photobooth', () => {
    // State
    const connected = ref(false)
    const state = ref('idle') // idle, countdown, capturing, processing, preview, error
    const countdown = ref({ remaining: 0, total: 0 })
    const lastPhoto = ref<any>(null)
    const error = ref<string | null>(null)
    const clients = ref(0)
    
    // WebSocket
    let ws: WebSocket | null = null
    let reconnectTimer: any = null

    function connect() {
        if (ws) return

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const host = window.location.host
        const url = `${protocol}//${host}/ws`

        console.log('🔌 Connecting to WebSocket:', url)
        ws = new WebSocket(url)

        ws.onopen = () => {
            console.log('✅ WebSocket connected')
            connected.value = true
            error.value = null
            if (reconnectTimer) clearInterval(reconnectTimer)
            
            // Identify based on URL logic or just default
            // Views will register roles on mount
        }

        ws.onclose = () => {
            console.log('❌ WebSocket disconnected')
            connected.value = false
            ws = null
            // Reconnect logic
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
            case 'clients_update':
                clients.value = msg.data.count // If we implemented this
                break
        }
    }

    function trigger() {
        if (!ws || ws.readyState !== WebSocket.OPEN) return
        ws.send(JSON.stringify({ type: 'trigger' }))
    }

    function register(role: string) {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
             // Retry if not connected yet
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
        connect,
        trigger,
        register
    }
})
