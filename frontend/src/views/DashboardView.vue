<template>
    <div class="min-h-screen bg-zinc-950 text-zinc-100 font-sans">
        <!-- Header -->
        <header class="border-b border-zinc-800 bg-zinc-900/80 backdrop-blur-sm sticky top-0 z-10">
            <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
                <div class="flex items-center gap-3">
                    <div class="w-2 h-2 rounded-full" :class="photobooth.connected ? 'bg-emerald-400' : 'bg-red-400'">
                    </div>
                    <h1 class="text-lg font-semibold tracking-tight">Photobooth V5</h1>
                    <span class="text-xs text-zinc-500 font-mono">Dashboard</span>
                </div>
                <div class="flex items-center gap-4 text-xs text-zinc-500">
                    <span class="font-mono">Uptime {{ photobooth.uptime }}</span>
                    <span>{{ photobooth.clients }} client{{ photobooth.clients !== 1 ? 's' : '' }}</span>
                </div>
            </div>
        </header>

        <main class="max-w-7xl mx-auto px-6 py-6 space-y-6">

            <!-- Top Row: Status + Controls -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">

                <!-- State -->
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wider mb-2">System State</p>
                    <div class="flex items-center gap-3">
                        <div class="w-3 h-3 rounded-full" :class="stateIndicator"></div>
                        <span class="text-2xl font-mono font-semibold" :class="stateColor">{{
                            photobooth.state.toUpperCase() }}</span>
                    </div>
                </div>

                <!-- Trigger -->
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-5 flex items-center">
                    <button @click="photobooth.trigger()" :disabled="photobooth.state !== 'idle'"
                        class="w-full py-3 rounded-md text-sm font-semibold tracking-wide transition-all duration-200"
                        :class="photobooth.state === 'idle'
                            ? 'bg-emerald-600 hover:bg-emerald-500 text-white cursor-pointer'
                            : 'bg-zinc-800 text-zinc-600 cursor-not-allowed'">
                        {{ photobooth.state === 'idle' ? 'TRIGGER CAPTURE' : photobooth.state.toUpperCase() }}
                    </button>
                </div>

                <!-- Latest Photo -->
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wider mb-2">Latest Photo</p>
                    <div v-if="photobooth.lastPhoto" class="aspect-video bg-zinc-800 rounded overflow-hidden">
                        <img :src="photobooth.lastPhoto.thumbUrl" class="w-full h-full object-cover" />
                    </div>
                    <div v-else class="aspect-video bg-zinc-800 rounded flex items-center justify-center">
                        <span class="text-zinc-600 text-sm">No photos yet</span>
                    </div>
                </div>

                <!-- Camera Info -->
                <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
                    <p class="text-xs text-zinc-500 uppercase tracking-wider mb-2">Camera</p>
                    <div v-if="photobooth.cameraInfo.connected">
                        <div class="flex items-center gap-2 mb-3">
                            <div class="w-2 h-2 rounded-full bg-emerald-400"></div>
                            <span class="text-sm font-semibold text-zinc-200">{{ photobooth.cameraInfo.model }}</span>
                        </div>
                        <div class="space-y-1.5 text-xs text-zinc-400">
                            <div v-if="photobooth.cameraInfo.manufacturer" class="flex justify-between">
                                <span class="text-zinc-500">Hersteller</span>
                                <span>{{ photobooth.cameraInfo.manufacturer }}</span>
                            </div>
                            <div v-if="photobooth.cameraInfo.lensName" class="flex justify-between">
                                <span class="text-zinc-500">Objektiv</span>
                                <span class="text-right max-w-[60%] truncate" :title="photobooth.cameraInfo.lensName">{{
                                    photobooth.cameraInfo.lensName }}</span>
                            </div>
                            <div v-if="photobooth.cameraInfo.batteryLevel" class="flex justify-between items-center">
                                <span class="text-zinc-500">Akku</span>
                                <div class="flex items-center gap-2">
                                    <div class="w-16 h-1.5 bg-zinc-700 rounded-full overflow-hidden">
                                        <div class="h-full rounded-full transition-all duration-500"
                                            :class="batteryBarColor"
                                            :style="{ width: photobooth.cameraInfo.batteryLevel }">
                                        </div>
                                    </div>
                                    <span class="font-mono">{{ photobooth.cameraInfo.batteryLevel }}</span>
                                </div>
                            </div>
                            <div v-if="photobooth.cameraInfo.storageFree" class="flex justify-between">
                                <span class="text-zinc-500">Speicher frei</span>
                                <span class="font-mono">{{ photobooth.cameraInfo.storageFree }}</span>
                            </div>
                        </div>
                    </div>
                    <div v-else class="flex flex-col items-center justify-center py-4 gap-2">
                        <div class="w-2 h-2 rounded-full bg-red-400"></div>
                        <span class="text-zinc-600 text-sm">Keine Kamera erkannt</span>
                    </div>
                </div>
            </div>

            <!-- Navigation -->
            <div class="flex flex-wrap gap-2">
                <router-link to="/modes" class="nav-link">Client Modus</router-link>
                <router-link to="/gallery" class="nav-link">Gallery</router-link>
                <a href="/legacy/" class="nav-link">Legacy Client</a>
            </div>

            <!-- Log Viewer -->
            <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
                <div class="flex items-center justify-between px-5 py-3 border-b border-zinc-800">
                    <p class="text-xs text-zinc-500 uppercase tracking-wider">Server Log</p>
                    <div class="flex items-center gap-2">
                        <button @click="toggleAutoScroll" class="text-xs px-2 py-1 rounded transition-colors"
                            :class="autoScroll ? 'bg-emerald-600/20 text-emerald-400' : 'bg-zinc-800 text-zinc-500'">
                            Auto-scroll {{ autoScroll ? 'ON' : 'OFF' }}
                        </button>
                        <span class="text-xs text-zinc-600">{{ photobooth.logs.length }} entries</span>
                    </div>
                </div>
                <div ref="logContainer" class="h-80 overflow-y-auto font-mono text-xs p-4 space-y-0.5 bg-zinc-950">
                    <div v-for="(entry, i) in photobooth.logs" :key="i"
                        class="flex gap-3 py-0.5 hover:bg-zinc-900/50 px-2 rounded">
                        <span class="text-zinc-600 shrink-0 w-20">{{ formatTime(entry.timestamp) }}</span>
                        <span class="shrink-0 w-12 uppercase font-semibold" :class="levelColor(entry.level)">{{
                            entry.level }}</span>
                        <span class="text-zinc-400 shrink-0 w-20">{{ entry.source }}</span>
                        <span class="text-zinc-300">{{ entry.message }}</span>
                    </div>
                    <div v-if="photobooth.logs.length === 0" class="text-zinc-600 text-center py-8">
                        No log entries yet. Trigger a capture to see activity.
                    </div>
                </div>
            </div>

        </main>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { usePhotoboothStore } from '../stores/photobooth';

const photobooth = usePhotoboothStore();
const logContainer = ref<HTMLElement | null>(null);
const autoScroll = ref(true);

const stateColor = computed(() => {
    switch (photobooth.state) {
        case 'idle': return 'text-emerald-400';
        case 'error': return 'text-red-400';
        case 'countdown': return 'text-amber-400';
        case 'capturing': return 'text-blue-400';
        case 'processing': return 'text-violet-400';
        case 'preview': return 'text-cyan-400';
        default: return 'text-zinc-400';
    }
});

const stateIndicator = computed(() => {
    switch (photobooth.state) {
        case 'idle': return 'bg-emerald-400';
        case 'error': return 'bg-red-400 animate-pulse';
        case 'countdown': return 'bg-amber-400 animate-pulse';
        default: return 'bg-blue-400 animate-pulse';
    }
});

const batteryBarColor = computed(() => {
    const level = parseInt(photobooth.cameraInfo.batteryLevel) || 0;
    if (level > 50) return 'bg-emerald-400';
    if (level > 20) return 'bg-amber-400';
    return 'bg-red-400';
});

function formatTime(ts: number) {
    const d = new Date(ts);
    return d.toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function levelColor(level: string) {
    switch (level) {
        case 'info': return 'text-blue-400';
        case 'warn': return 'text-amber-400';
        case 'error': return 'text-red-400';
        case 'debug': return 'text-zinc-500';
        default: return 'text-zinc-400';
    }
}

function toggleAutoScroll() {
    autoScroll.value = !autoScroll.value;
}

function scrollToBottom() {
    if (autoScroll.value && logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
}

// Watch for new logs and auto-scroll
watch(() => photobooth.logs.length, () => {
    nextTick(scrollToBottom);
});

onMounted(() => {
    photobooth.register('dashboard');
    photobooth.fetchLogs();
    photobooth.fetchStatus();

    // Poll status every 5s
    setInterval(() => photobooth.fetchStatus(), 5000);
});
</script>

<style scoped>
.nav-link {
    padding: 0.375rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 500;
    border-radius: 0.375rem;
    border: 1px solid rgb(39 39 42);
    background-color: rgb(24 24 27);
    color: rgb(161 161 170);
    transition: all 0.15s;
    text-decoration: none;
}

.nav-link:hover {
    background-color: rgb(39 39 42);
    color: rgb(228 228 231);
}
</style>
