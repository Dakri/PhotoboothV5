import { ref } from 'vue'

const REQUIRED_TAPS = 10
const TAP_TIMEOUT = 3000 // Reset counter after 3s of inactivity

export function useExitLock() {
    const tapCount = ref(0)
    const isUnlocked = ref(false)
    let resetTimer: any = null

    function tap() {
        tapCount.value++

        // Reset timer on each tap
        if (resetTimer) clearTimeout(resetTimer)
        resetTimer = setTimeout(() => {
            tapCount.value = 0
        }, TAP_TIMEOUT)

        if (tapCount.value >= REQUIRED_TAPS) {
            isUnlocked.value = true
            tapCount.value = 0
        }
    }

    function lock() {
        isUnlocked.value = false
        tapCount.value = 0
    }

    return {
        tapCount,
        isUnlocked,
        requiredTaps: REQUIRED_TAPS,
        tap,
        lock
    }
}
