<template>
    <div class="h-screen w-screen bg-black text-white overflow-hidden select-none" @click="handleTap">

        <!-- Exit Lock Overlay -->
        <Transition name="fade">
            <div v-if="exitLock.isUnlocked.value"
                class="absolute inset-0 z-50 bg-zinc-950/95 flex flex-col items-center justify-center gap-4 p-6">
                <p class="text-lg font-semibold">Modus wechseln?</p>
                <p class="text-sm text-zinc-400">Aktuell: {{ modeStore.modeDefinition?.label }}</p>
                <button @click.stop="goToModeSelect" class="px-6 py-3 bg-zinc-800 border border-zinc-700 rounded-lg text-sm font-medium
                       hover:bg-zinc-700 transition-colors">
                    Modus wählen
                </button>
                <button @click.stop="goToDashboard" class="px-6 py-3 bg-red-900/30 border border-red-800/50 rounded-lg text-sm font-medium
                       text-red-400 hover:bg-red-900/50 transition-colors">
                    Client Modus verlassen
                </button>
                <button @click.stop="exitLock.lock()"
                    class="px-6 py-3 text-sm text-zinc-500 hover:text-zinc-300 transition-colors">
                    Abbrechen
                </button>
            </div>
        </Transition>

        <!-- Tap counter indicator (subtle) -->
        <div v-if="exitLock.tapCount.value > 0 && exitLock.tapCount.value < exitLock.requiredTaps && !exitLock.isUnlocked.value"
            class="absolute top-2 right-2 z-40">
            <div class="flex gap-0.5">
                <div v-for="i in exitLock.requiredTaps" :key="i"
                    class="w-1 h-1 rounded-full transition-colors duration-200"
                    :class="i <= exitLock.tapCount.value ? 'bg-zinc-500' : 'bg-zinc-800'">
                </div>
            </div>
        </div>

        <!-- GALLERY MODE -->
        <div v-if="mode?.hasGallery" class="h-full overflow-y-auto p-4">
            <div class="grid grid-cols-3 gap-2">
                <div v-for="photo in gallery.photos" :key="photo.filename"
                    class="aspect-square bg-zinc-900 rounded overflow-hidden">
                    <img :src="photo.thumbUrl" class="w-full h-full object-cover" loading="lazy" />
                </div>
            </div>
            <div v-if="gallery.photos.length === 0" class="flex items-center justify-center h-full">
                <p class="text-zinc-600 text-sm">Noch keine Fotos</p>
            </div>
        </div>

        <!-- INTERACTIVE MODES (Buzzer / Countdown / Preview) -->
        <div v-else class="h-full w-full flex items-center justify-center">

            <!-- IDLE STATE -->
            <div v-if="photobooth.state === 'idle'" class="text-center">
                <!-- Buzzer mode: show big touch target -->
                <div v-if="mode?.hasBuzzer" class="cursor-pointer" @click.stop="triggerCapture">
                    <div class="w-48 h-48 rounded-full border-2 border-zinc-700 flex items-center justify-center
                      mx-auto mb-6 transition-all duration-200 hover:border-zinc-500 hover:bg-zinc-900/50
                      active:scale-95 active:border-emerald-500">
                        <svg class="w-16 h-16 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor"
                            stroke-width="1.5">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" />
                            <path stroke-linecap="round" stroke-linejoin="round"
                                d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z" />
                        </svg>
                    </div>
                    <p class="text-sm text-zinc-500 uppercase tracking-[0.2em]">Antippen zum Auslösen</p>
                </div>
                <!-- Non-buzzer modes: "waiting" prompt -->
                <div v-else-if="mode?.hasCountdown">
                    <p class="text-sm text-zinc-600 uppercase tracking-[0.3em]">Warte auf Auslöser</p>
                    <div class="mt-4 w-2 h-2 rounded-full bg-zinc-700 mx-auto animate-pulse"></div>
                </div>
                <!-- Preview only: show last photo -->
                <div v-else-if="mode?.hasPreview" class="w-full h-full">
                    <img v-if="photobooth.lastPhoto" :src="photobooth.lastPhoto.url"
                        class="max-w-full max-h-screen object-contain mx-auto" />
                    <p v-else class="text-zinc-600 text-sm">Warte auf erstes Foto</p>
                </div>
            </div>

            <!-- COUNTDOWN STATE -->
            <div v-else-if="photobooth.state === 'countdown' && mode?.hasCountdown"
                class="relative flex items-center justify-center">
                <svg class="absolute w-80 h-80" viewBox="0 0 200 200">
                    <circle cx="100" cy="100" r="90" stroke="#27272a" stroke-width="4" fill="none" />
                    <circle cx="100" cy="100" r="90" stroke="#f59e0b" stroke-width="4" fill="none"
                        stroke-linecap="round" :stroke-dasharray="circumference" :stroke-dashoffset="dashOffset"
                        class="transition-all duration-1000 ease-linear" transform="rotate(-90 100 100)" />
                </svg>
                <span class="text-[12rem] font-extralight tabular-nums leading-none">
                    {{ photobooth.countdown.remaining }}
                </span>
            </div>

            <!-- COUNTDOWN (mode doesn't have countdown — just show state) -->
            <div v-else-if="photobooth.state === 'countdown' && !mode?.hasCountdown">
                <div class="w-2 h-2 rounded-full bg-amber-500 mx-auto animate-pulse"></div>
            </div>

            <!-- CAPTURING STATE -->
            <div v-else-if="photobooth.state === 'capturing'" class="text-center">
                <div class="w-6 h-6 rounded-full bg-white mx-auto mb-6 animate-ping"></div>
                <p class="text-sm text-zinc-400 uppercase tracking-[0.3em]">Aufnahme</p>
            </div>

            <!-- PROCESSING STATE -->
            <div v-else-if="photobooth.state === 'processing'" class="text-center">
                <div class="w-8 h-8 border-2 border-zinc-600 border-t-white rounded-full mx-auto mb-6 animate-spin">
                </div>
                <p class="text-sm text-zinc-400 uppercase tracking-[0.3em]">Verarbeitung</p>
            </div>

            <!-- PREVIEW STATE -->
            <div v-else-if="photobooth.state === 'preview'" class="w-full h-full flex items-center justify-center">
                <div v-if="mode?.hasPreview && photobooth.lastPhoto"
                    class="w-full h-full flex items-center justify-center">
                    <img :src="photobooth.lastPhoto.url" class="max-w-full max-h-screen object-contain" />
                </div>
                <div v-else class="text-center">
                    <svg class="w-12 h-12 text-emerald-500 mx-auto mb-4" fill="none" viewBox="0 0 24 24"
                        stroke="currentColor" stroke-width="1.5">
                        <path stroke-linecap="round" stroke-linejoin="round"
                            d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
                    </svg>
                    <p class="text-sm text-zinc-400 uppercase tracking-[0.3em]">Foto gespeichert</p>
                </div>
            </div>

            <!-- ERROR STATE -->
            <div v-else-if="photobooth.state === 'error'" class="text-center">
                <div
                    class="w-12 h-12 rounded-full border-2 border-red-500/50 flex items-center justify-center mx-auto mb-4">
                    <span class="text-red-400 text-xl">!</span>
                </div>
                <p class="text-sm text-red-400">{{ photobooth.error || 'Fehler aufgetreten' }}</p>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { usePhotoboothStore } from '../stores/photobooth';
import { useClientModeStore } from '../stores/clientMode';
import { useGalleryStore } from '../stores/gallery';
import { useExitLock } from '../composables/useExitLock';
import { useFullscreen } from '../composables/useFullscreen';

const photobooth = usePhotoboothStore();
const modeStore = useClientModeStore();
const gallery = useGalleryStore();
const exitLock = useExitLock();
const { exitFullscreen } = useFullscreen();
const router = useRouter();

const mode = computed(() => modeStore.modeDefinition);

const circumference = 2 * Math.PI * 90;

const dashOffset = computed(() => {
    const { remaining, total } = photobooth.countdown;
    if (total === 0) return circumference;
    const progress = remaining / total;
    return circumference * (1 - progress);
});

function handleTap() {
    if (exitLock.isUnlocked.value) return;
    exitLock.tap();
}

function triggerCapture() {
    if (photobooth.state === 'idle') {
        photobooth.trigger();
    }
}

function goToModeSelect() {
    exitLock.lock();
    exitFullscreen();
    modeStore.clearMode();
    router.push('/modes');
}

function goToDashboard() {
    exitLock.lock();
    exitFullscreen();
    modeStore.clearMode();
    router.push('/');
}

onMounted(() => {
    // Redirect to mode select if no mode chosen
    if (!modeStore.selectedMode) {
        router.push('/modes');
        return;
    }

    // Load gallery if in gallery mode
    if (mode.value?.hasGallery) {
        gallery.fetchPhotos();
    }
});

// Refresh gallery when new photo arrives in gallery mode
watch(() => photobooth.lastPhoto, () => {
    if (mode.value?.hasGallery) {
        gallery.fetchPhotos();
    }
});
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
