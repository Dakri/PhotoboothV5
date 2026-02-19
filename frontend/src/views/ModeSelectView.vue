<template>
    <div class="min-h-screen bg-zinc-950 text-zinc-100 flex flex-col items-center justify-center p-6"
        @click="handleFirstClick">

        <div class="max-w-md w-full space-y-6">
            <div class="text-center mb-8">
                <h1 class="text-2xl font-semibold tracking-tight">Photobooth</h1>
                <p class="text-sm text-zinc-500 mt-1">Modus wählen</p>
            </div>

            <button v-for="mode in MODES" :key="mode.id" @click.stop="selectAndStart(mode.id)" class="w-full text-left p-4 rounded-lg border transition-all duration-150
                     bg-zinc-900 border-zinc-800 hover:border-zinc-600 hover:bg-zinc-800/80
                     active:scale-[0.98]">
                <div class="flex items-center justify-between">
                    <div>
                        <p class="font-medium text-sm">{{ mode.label }}</p>
                        <p class="text-xs text-zinc-500 mt-0.5">{{ mode.description }}</p>
                    </div>
                    <div class="flex gap-1">
                        <span v-if="mode.hasBuzzer" class="badge bg-emerald-900/40 text-emerald-400">Buzzer</span>
                        <span v-if="mode.hasCountdown" class="badge bg-amber-900/40 text-amber-400">Timer</span>
                        <span v-if="mode.hasPreview" class="badge bg-blue-900/40 text-blue-400">Bild</span>
                        <span v-if="mode.hasGallery" class="badge bg-violet-900/40 text-violet-400">Galerie</span>
                    </div>
                </div>
            </button>

            <p class="text-center text-xs text-zinc-600 mt-4">
                Nach Auswahl wird Vollbild aktiviert.<br />
                10× tippen um Modus zu wechseln.
            </p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useClientModeStore, MODES, type ClientMode } from '../stores/clientMode';
import { useFullscreen } from '../composables/useFullscreen';
import { usePhotoboothStore } from '../stores/photobooth';

const router = useRouter();
const modeStore = useClientModeStore();
const photobooth = usePhotoboothStore();
const { enterFullscreen } = useFullscreen();
const hasInteracted = ref(false);

function handleFirstClick() {
    // Need user gesture for fullscreen API
    hasInteracted.value = true;
}

function selectAndStart(modeId: ClientMode) {
    hasInteracted.value = true;
    modeStore.selectMode(modeId);
    photobooth.register(modeId);
    enterFullscreen();
    router.push('/client');
}
</script>

<style scoped>
.badge {
    font-size: 0.625rem;
    padding: 0.125rem 0.375rem;
    border-radius: 0.25rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}
</style>
