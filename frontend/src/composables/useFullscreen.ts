import { ref } from 'vue'

export function useFullscreen() {
    const isFullscreen = ref(false)

    function enterFullscreen(el?: HTMLElement) {
        const target = el || document.documentElement
        const rfs = target.requestFullscreen
            || (target as any).webkitRequestFullscreen
            || (target as any).mozRequestFullScreen
            || (target as any).msRequestFullscreen

        if (rfs) {
            rfs.call(target).catch(() => {
                // Fullscreen denied — silently continue
            })
        }
        isFullscreen.value = true
    }

    function exitFullscreen() {
        const eFS = document.exitFullscreen
            || (document as any).webkitExitFullscreen
            || (document as any).mozCancelFullScreen
            || (document as any).msExitFullscreen

        if (eFS && document.fullscreenElement) {
            eFS.call(document)
        }
        isFullscreen.value = false
    }

    // Listen for fullscreen changes
    document.addEventListener('fullscreenchange', () => {
        isFullscreen.value = !!document.fullscreenElement
    })
    document.addEventListener('webkitfullscreenchange', () => {
        isFullscreen.value = !!(document as any).webkitFullscreenElement
    })

    return { isFullscreen, enterFullscreen, exitFullscreen }
}
