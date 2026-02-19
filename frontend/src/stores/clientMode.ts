import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ClientMode =
    | 'buzzer-countdown-preview'
    | 'buzzer-countdown'
    | 'countdown-preview'
    | 'preview-only'
    | 'countdown-only'
    | 'gallery'

export interface ModeDefinition {
    id: ClientMode
    label: string
    description: string
    hasBuzzer: boolean
    hasCountdown: boolean
    hasPreview: boolean
    hasGallery: boolean
}

export const MODES: ModeDefinition[] = [
    {
        id: 'buzzer-countdown-preview',
        label: 'Vollständig',
        description: 'Auslöser, Countdown & Vorschau',
        hasBuzzer: true, hasCountdown: true, hasPreview: true, hasGallery: false
    },
    {
        id: 'buzzer-countdown',
        label: 'Auslöser + Countdown',
        description: 'Ohne Bildvorschau',
        hasBuzzer: true, hasCountdown: true, hasPreview: false, hasGallery: false
    },
    {
        id: 'countdown-preview',
        label: 'Monitor + Vorschau',
        description: 'Zeigt Countdown & letztes Bild',
        hasBuzzer: false, hasCountdown: true, hasPreview: true, hasGallery: false
    },
    {
        id: 'preview-only',
        label: 'Nur Vorschau',
        description: 'Zeigt nur das letzte Bild',
        hasBuzzer: false, hasCountdown: false, hasPreview: true, hasGallery: false
    },
    {
        id: 'countdown-only',
        label: 'Nur Countdown',
        description: 'Zeigt nur den Countdown',
        hasBuzzer: false, hasCountdown: true, hasPreview: false, hasGallery: false
    },
    {
        id: 'gallery',
        label: 'Galerie',
        description: 'Alle Bilder durchstöbern',
        hasBuzzer: false, hasCountdown: false, hasPreview: false, hasGallery: true
    }
]

export const useClientModeStore = defineStore('clientMode', () => {
    const selectedMode = ref<ClientMode | null>(null)
    const modeDefinition = ref<ModeDefinition | null>(null)

    function selectMode(modeId: ClientMode) {
        selectedMode.value = modeId
        modeDefinition.value = MODES.find(m => m.id === modeId) || null
    }

    function clearMode() {
        selectedMode.value = null
        modeDefinition.value = null
    }

    return {
        selectedMode,
        modeDefinition,
        selectMode,
        clearMode
    }
})
