import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Photo {
    filename: string
    timestamp: string
    url: string
    thumbUrl: string
}

export const useGalleryStore = defineStore('gallery', () => {
    const photos = ref<Photo[]>([])
    const loading = ref(false)
    const error = ref<string | null>(null)

    async function fetchPhotos() {
        loading.value = true
        try {
            const res = await fetch('/api/photos')
            if (!res.ok) throw new Error('Failed to fetch photos')
            photos.value = await res.json()
        } catch (e: any) {
            error.value = e.message
        } finally {
            loading.value = false
        }
    }

    return {
        photos,
        loading,
        error,
        fetchPhotos
    }
})
