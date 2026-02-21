<template>
    <div class="h-screen w-screen bg-zinc-950 text-zinc-50 overflow-hidden select-none font-sans" @click="handleTap">

        <!-- Exit Lock Overlay -->
        <Transition name="fade">
            <div v-if="exitLock.isUnlocked.value"
                class="absolute inset-0 z-50 bg-zinc-900/95 flex flex-col items-center justify-center gap-4 p-6 backdrop-blur-sm">
                <p class="text-2xl font-bold tracking-tight text-white">Modus wechseln?</p>
                <p class="text-zinc-400">Aktuell: {{ modeStore.modeDefinition?.label }}</p>
                <div class="flex flex-col gap-3 w-full max-w-xs">
                    <button @click.stop="goToModeSelect" class="px-6 py-4 bg-zinc-800 rounded-xl text-lg font-medium shadow-sm border border-zinc-700
                       hover:bg-zinc-700 text-white transition-all active:scale-95">
                        Modus wählen
                    </button>
                    <button @click.stop="goToDashboard" class="px-6 py-4 bg-zinc-800 border border-zinc-700 rounded-xl text-lg font-medium shadow-sm
                       text-white hover:bg-zinc-700 transition-all active:scale-95">
                        Dashboard
                    </button>
                    <button @click.stop="exitLock.lock()"
                        class="px-6 py-3 text-zinc-500 hover:text-zinc-300 transition-colors">
                        Abbrechen
                    </button>
                </div>
            </div>
        </Transition>

        <!-- Tap counter indicator (subtle) -->
        <div v-if="exitLock.tapCount.value > 0 && exitLock.tapCount.value < exitLock.requiredTaps && !exitLock.isUnlocked.value"
            class="absolute top-4 right-4 z-60">
            <div class="flex gap-1">
                <div v-for="i in exitLock.requiredTaps" :key="i"
                    class="w-2 h-2 rounded-full transition-colors duration-200"
                    :class="i <= exitLock.tapCount.value ? 'bg-zinc-500' : 'bg-zinc-800'">
                </div>
            </div>
        </div>

        <!-- SYSTEM WARNINGS OVERLAY -->
        <Transition name="fade">
            <div v-if="!photobooth.cameraInfo.connected || photobooth.diskInfo.usedPercent > 95"
                class="absolute inset-0 z-40 bg-zinc-950/95 backdrop-blur-md flex flex-col items-center justify-center p-8 text-center text-white">

                <div v-if="!photobooth.cameraInfo.connected"
                    class="flex flex-col items-center animate-[pulse_2s_ease-in-out_infinite]">
                    <svg class="w-32 h-32 text-red-500 mb-6 drop-shadow-[0_0_15px_rgba(239,68,68,0.4)]" fill="none"
                        viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                            d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4l16 16" />
                    </svg>
                    <h2 class="text-5xl font-black tracking-tight uppercase mb-4">Kamera getrennt</h2>
                    <p class="text-2xl text-zinc-400 font-light">Bitte überprüfe die USB-Verbindung und schalte die
                        Kamera ein.</p>
                </div>

                <div v-else-if="photobooth.diskInfo.usedPercent > 95" class="flex flex-col items-center">
                    <svg class="w-32 h-32 text-orange-500 mb-6 drop-shadow-[0_0_15px_rgba(249,115,22,0.4)]" fill="none"
                        viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                            d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" />
                    </svg>
                    <h2 class="text-5xl font-black tracking-tight uppercase mb-4 text-orange-500">Speicher voll</h2>
                    <p class="text-2xl text-zinc-400 font-light">Es ist fast kein Speicherplatz mehr für neue Fotos
                        vorhanden.</p>
                    <p class="mt-4 text-zinc-500 text-lg">Wird zu {{ Math.round(photobooth.diskInfo.usedPercent) }}%
                        genutzt</p>
                </div>

                <!-- <p class="absolute bottom-12 text-zinc-600 font-medium tracking-wide">Tippe oben rechts um das
                    Admin-Menü zu öffnen</p> -->
            </div>
        </Transition>

        <!-- GALLERY MODE -->
        <div v-if="mode?.hasGallery" class="h-full overflow-y-auto p-6 bg-zinc-950">
            <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                <div v-for="photo in gallery.photos" :key="photo.filename"
                    class="aspect-square bg-zinc-900 rounded-2xl shadow-sm border border-zinc-800 overflow-hidden relative group">
                    <img :src="photo.thumbUrl"
                        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                        loading="lazy" />
                </div>
            </div>
            <div v-if="gallery.photos.length === 0"
                class="flex flex-col items-center justify-center h-full text-zinc-400 gap-4">
                <svg class="w-16 h-16 opacity-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"
                        d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <p class="font-medium">Galerie ist leer</p>
            </div>

            <!-- Back Button for Gallery Mode if needed contextually, 
                  but usually gallery mode is a standalone mode. 
                  Tap-pattern is used to exit. -->
        </div>

        <!-- CAPTURE / INTERACTIVE MODES -->
        <div v-else class="h-full w-full relative flex flex-col">

            <!-- MAIN CONTENT AREA (Photo / Preview) -->
            <div class="flex-1 relative bg-zinc-950 overflow-hidden">
                <!-- Persistent Last Photo OR Active Preview -->
                <Transition name="zoom">
                    <div v-if="photobooth.lastPhoto && (mode?.id === 'preview-only' || (photobooth.state === 'preview' && mode?.hasPreview))"
                        :key="photobooth.lastPhoto.filename" class="absolute inset-0 z-10">
                        <img :src="photobooth.lastPhoto.url" class="w-full h-full object-cover" />
                        <!-- Gradient Overlay for contrast if needed -->
                        <div class="absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-black/50 to-transparent">
                        </div>
                    </div>
                </Transition>

                <!-- Trigger Button Container -->
                <div v-if="photobooth.state === 'idle' || (photobooth.state === 'preview' && !mode?.hasPreview)"
                    class="absolute inset-0 flex items-center justify-center z-20 pointer-events-none">
                    <!-- Buzzer Trigger -->
                    <div v-if="mode?.hasBuzzer"
                        class="pointer-events-auto cursor-pointer flex flex-col items-center transition-opacity duration-300"
                        @click.stop="triggerCapture">
                        <div class="w-80 h-80 rounded-full border-[6px] border-zinc-400 flex items-center justify-center
                          transition-all duration-300 hover:border-zinc-300 hover:bg-zinc-800/30 hover:shadow-[0_0_40px_rgba(255,255,255,0.1)]
                          active:scale-95 active:border-white">
                            <svg class="w-32 h-32 text-zinc-300" fill="none" viewBox="0 0 24 24" stroke="currentColor"
                                stroke-width="1.2">
                                <path stroke-linecap="round" stroke-linejoin="round"
                                    d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" />
                                <path stroke-linecap="round" stroke-linejoin="round"
                                    d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z" />
                            </svg>
                        </div>
                    </div>
                    <!-- Fallback if no photo or not in preview mode -->
                    <div v-else-if="photobooth.state === 'idle' && (!photobooth.lastPhoto || mode?.id !== 'preview-only')"
                        class="text-zinc-600 text-center pointer-events-auto">
                        <svg class="w-24 h-24 mx-auto mb-4 opacity-20" fill="none" viewBox="0 0 24 24"
                            stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1"
                                d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1"
                                d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        <p class="text-xl font-light tracking-widest uppercase">Bereit</p>
                    </div>
                </div>

                <!-- PREVIEW OVERLAY (Shows immediately after capture) -->
                <!-- Handled by the image logic above -->

                <!-- COUNTDOWN OVERLAY -->
                <div v-if="photobooth.state === 'countdown' && mode?.hasCountdown"
                    class="absolute inset-0 flex items-center justify-center bg-zinc-950/80 backdrop-blur-sm z-30">
                    <div class="relative flex items-center justify-center">
                        <svg class="absolute w-96 h-96" viewBox="0 0 200 200">
                            <circle cx="100" cy="100" r="90" stroke="#27272a" stroke-width="3" fill="none" />
                            <circle cx="100" cy="100" r="90" stroke="#f59e0b" stroke-width="3" fill="none"
                                stroke-linecap="round" stroke-dasharray="565.48" stroke-dashoffset="0"
                                class="animate-shrink-ring"
                                :style="{ animationDuration: `${photobooth.countdown.total || photobooth.settings.countdownSeconds}s` }"
                                transform="rotate(-90 100 100)" />
                        </svg>
                        <span
                            class="text-[14rem] font-extralight text-zinc-100 drop-shadow-2xl tabular-nums leading-none tracking-tighter -mt-6"
                            :key="photobooth.countdown.remaining">
                            {{ photobooth.countdown.remaining }}
                        </span>
                    </div>
                </div>

                <!-- CAPTURING / PROCESSING OVERLAY -->
                <div v-if="photobooth.state === 'capturing' || photobooth.state === 'processing'"
                    class="absolute inset-0 flex flex-col items-center justify-center bg-black/50 backdrop-blur-md z-20">
                    <div class="flex flex-col items-center gap-10">
                        <div v-if="!isSmiling"
                            class="w-16 h-16 border-[6px] border-zinc-700 border-t-white rounded-full animate-spin">
                        </div>
                        <p v-if="isSmiling"
                            class="text-6xl font-black text-white uppercase tracking-[0.2em] drop-shadow-2xl animate-tada">
                            Bitte lächeln!
                        </p>
                        <p v-else class="text-3xl font-light text-white uppercase tracking-[0.3em] drop-shadow-2xl"
                            :class="processingAnimClass">
                            Vorschau wird geladen...
                        </p>
                    </div>
                </div>

                <!-- ERROR OVERLAY -->
                <div v-if="photobooth.state === 'error'"
                    class="absolute inset-0 flex flex-col items-center justify-center bg-red-900/90 backdrop-blur-sm z-50 p-10 text-center text-white">
                    <svg class="w-20 h-20 mb-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                    <p class="text-3xl font-bold mb-2">Ups!</p>
                    <p class="text-xl opacity-90">{{ photobooth.error || 'Ein Fehler ist aufgetreten' }}</p>
                    <!-- <p class="mt-8 text-sm opacity-75">Tippe oben rechts 5x um neu zu laden</p> -->
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onMounted, computed, watch, ref } from 'vue';
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
const { exitFullscreen, enterFullscreen } = useFullscreen();
const router = useRouter();

const mode = computed(() => modeStore.modeDefinition);

// Random Processing Animation – 16 unique CSS animations
const processingAnimClass = ref('anim-pulse-glow');
const animOptions = [
    'anim-pulse-glow',
    'anim-float',
    'anim-wiggle',
    'anim-bounce-in',
    'anim-slide-up',
    'anim-slide-down',
    'anim-zoom-pulse',
    'anim-swing',
    'anim-rubber',
    'anim-jello',
    'anim-heartbeat',
    'anim-flash',
    'anim-shake-x',
    'anim-blur-in',
    'anim-tracking-expand',
    'anim-wave',
];

const isSmiling = ref(false);
let smileTimeout: ReturnType<typeof setTimeout> | null = null;

watch(() => photobooth.state, (newVal) => {
    if (newVal === 'capturing') {
        isSmiling.value = true;
        processingAnimClass.value = animOptions[Math.floor(Math.random() * animOptions.length)];
        if (smileTimeout) clearTimeout(smileTimeout);
        smileTimeout = setTimeout(() => {
            isSmiling.value = false;
        }, 1200);
    } else if (newVal !== 'processing') {
        isSmiling.value = false;
        if (smileTimeout) clearTimeout(smileTimeout);
    }
});

function handleTap() {
    // Aggressively re-enter fullscreen on every tap (user may have dismissed it)
    if (!exitLock.isUnlocked.value) {
        enterFullscreen();
    }
    if (exitLock.isUnlocked.value) return;
    exitLock.tap();
}

function triggerCapture() {
    if (photobooth.state === 'idle' || photobooth.state === 'preview') {
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

// Watch for last photo changes to maybe animate? 
// The Transition group handles enter/leave.

onMounted(() => {
    if (!modeStore.selectedMode) {
        router.push('/modes');
        return;
    }
    if (mode.value?.hasGallery) {
        gallery.fetchPhotos();
    }
    // Fetch latest status to ensure lastPhoto is present
    photobooth.fetchStatus();
});

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

.zoom-enter-active,
.zoom-leave-active {
    transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.zoom-enter-from {
    opacity: 0;
    transform: scale(0.95);
}

.zoom-leave-to {
    opacity: 0;
    /* transform: scale(1.05); */
    /* Don't scale out to avoid messy overlap */
}

/* Custom animations */
@keyframes shrink-ring {
    from {
        stroke-dashoffset: 0;
    }

    to {
        stroke-dashoffset: 565.48;
    }
}

.animate-shrink-ring {
    animation: shrink-ring linear forwards;
}

@keyframes tada {
    from {
        transform: scale3d(1, 1, 1);
    }

    10%,
    20% {
        transform: scale3d(0.9, 0.9, 0.9) rotate3d(0, 0, 1, -3deg);
    }

    30%,
    50%,
    70%,
    90% {
        transform: scale3d(1.1, 1.1, 1.1) rotate3d(0, 0, 1, 3deg);
    }

    40%,
    60%,
    80% {
        transform: scale3d(1.1, 1.1, 1.1) rotate3d(0, 0, 1, -3deg);
    }

    to {
        transform: scale3d(1, 1, 1);
    }
}

.animate-tada {
    animation: tada 1s ease-in-out forwards;
}

/* === 16 Random Processing Animations === */
@keyframes pulse-glow {

    0%,
    100% {
        opacity: 1;
        text-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
    }

    50% {
        opacity: 0.6;
        text-shadow: 0 0 30px rgba(255, 255, 255, 0.8);
    }
}

.anim-pulse-glow {
    animation: pulse-glow 2s ease-in-out infinite;
}

@keyframes float {

    0%,
    100% {
        transform: translateY(0);
    }

    50% {
        transform: translateY(-12px);
    }
}

.anim-float {
    animation: float 2s ease-in-out infinite;
}

@keyframes wiggle {

    0%,
    100% {
        transform: rotate(0deg);
    }

    15% {
        transform: rotate(6deg);
    }

    30% {
        transform: rotate(-6deg);
    }

    45% {
        transform: rotate(4deg);
    }

    60% {
        transform: rotate(-4deg);
    }

    75% {
        transform: rotate(2deg);
    }
}

.anim-wiggle {
    animation: wiggle 1.5s ease-in-out infinite;
}

@keyframes bounce-in {
    0% {
        transform: scale(0.3);
        opacity: 0;
    }

    50% {
        transform: scale(1.08);
    }

    70% {
        transform: scale(0.95);
    }

    100% {
        transform: scale(1);
        opacity: 1;
    }
}

.anim-bounce-in {
    animation: bounce-in 0.8s ease-out forwards;
}

@keyframes slide-up {
    from {
        transform: translateY(40px);
        opacity: 0;
    }

    to {
        transform: translateY(0);
        opacity: 1;
    }
}

.anim-slide-up {
    animation: slide-up 0.6s ease-out forwards;
}

@keyframes slide-down {
    from {
        transform: translateY(-40px);
        opacity: 0;
    }

    to {
        transform: translateY(0);
        opacity: 1;
    }
}

.anim-slide-down {
    animation: slide-down 0.6s ease-out forwards;
}

@keyframes zoom-pulse {

    0%,
    100% {
        transform: scale(1);
    }

    50% {
        transform: scale(1.08);
    }
}

.anim-zoom-pulse {
    animation: zoom-pulse 1.8s ease-in-out infinite;
}

@keyframes swing {
    20% {
        transform: rotate(8deg);
    }

    40% {
        transform: rotate(-6deg);
    }

    60% {
        transform: rotate(4deg);
    }

    80% {
        transform: rotate(-2deg);
    }

    100% {
        transform: rotate(0deg);
    }
}

.anim-swing {
    animation: swing 1.5s ease-in-out infinite;
    transform-origin: top center;
}

@keyframes rubber {
    0% {
        transform: scaleX(1) scaleY(1);
    }

    30% {
        transform: scaleX(1.25) scaleY(0.75);
    }

    40% {
        transform: scaleX(0.75) scaleY(1.25);
    }

    50% {
        transform: scaleX(1.15) scaleY(0.85);
    }

    65% {
        transform: scaleX(0.95) scaleY(1.05);
    }

    75% {
        transform: scaleX(1.05) scaleY(0.95);
    }

    100% {
        transform: scaleX(1) scaleY(1);
    }
}

.anim-rubber {
    animation: rubber 1s ease-in-out infinite;
}

@keyframes jello {

    0%,
    100% {
        transform: skewX(0) skewY(0);
    }

    30% {
        transform: skewX(-8deg) skewY(-8deg);
    }

    40% {
        transform: skewX(6deg) skewY(6deg);
    }

    50% {
        transform: skewX(-4deg) skewY(-4deg);
    }

    65% {
        transform: skewX(2deg) skewY(2deg);
    }

    75% {
        transform: skewX(-1deg) skewY(-1deg);
    }
}

.anim-jello {
    animation: jello 1.5s ease-in-out infinite;
}

@keyframes heartbeat {

    0%,
    100% {
        transform: scale(1);
    }

    14% {
        transform: scale(1.15);
    }

    28% {
        transform: scale(1);
    }

    42% {
        transform: scale(1.15);
    }

    70% {
        transform: scale(1);
    }
}

.anim-heartbeat {
    animation: heartbeat 1.5s ease-in-out infinite;
}

@keyframes flash {

    0%,
    50%,
    100% {
        opacity: 1;
    }

    25%,
    75% {
        opacity: 0.3;
    }
}

.anim-flash {
    animation: flash 2s ease-in-out infinite;
}

@keyframes shake-x {

    0%,
    100% {
        transform: translateX(0);
    }

    10%,
    30%,
    50%,
    70%,
    90% {
        transform: translateX(-6px);
    }

    20%,
    40%,
    60%,
    80% {
        transform: translateX(6px);
    }
}

.anim-shake-x {
    animation: shake-x 1.5s ease-in-out infinite;
}

@keyframes blur-in {
    from {
        filter: blur(12px);
        opacity: 0;
    }

    to {
        filter: blur(0);
        opacity: 1;
    }
}

.anim-blur-in {
    animation: blur-in 0.8s ease-out forwards;
}

@keyframes tracking-expand {
    from {
        letter-spacing: -0.1em;
        opacity: 0;
    }

    to {
        letter-spacing: 0.3em;
        opacity: 1;
    }
}

.anim-tracking-expand {
    animation: tracking-expand 0.8s ease-out forwards;
}

@keyframes wave {

    0%,
    100% {
        transform: translateY(0) rotate(0);
    }

    25% {
        transform: translateY(-6px) rotate(2deg);
    }

    75% {
        transform: translateY(6px) rotate(-2deg);
    }
}

.anim-wave {
    animation: wave 2s ease-in-out infinite;
}
</style>
