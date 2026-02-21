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
                <div class="flex items-center gap-4">
                    <nav class="flex gap-1">
                        <router-link to="/modes"
                            class="px-3 py-1.5 text-xs font-medium rounded-md bg-zinc-800 text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors">Client
                            Modus</router-link>
                        <router-link to="/gallery"
                            class="px-3 py-1.5 text-xs font-medium rounded-md bg-zinc-800 text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors">Gallery</router-link>
                        <a href="/legacy/"
                            class="px-3 py-1.5 text-xs font-medium rounded-md bg-zinc-800 text-zinc-300 hover:bg-zinc-700 hover:text-white transition-colors">Legacy</a>
                    </nav>
                    <span class="text-xs text-zinc-500 font-mono pl-4 border-l border-zinc-800">v5.0.0</span>
                </div>
            </div>
        </header>

        <main class="max-w-7xl mx-auto px-6 py-6 space-y-6">

            <!-- USB Export Progress Banner (appears at top when active) -->
            <div v-if="photobooth.usbExport.active"
                class="bg-indigo-900/50 border border-indigo-700/50 rounded-lg p-4 flex flex-col gap-2 relative overflow-hidden backdrop-blur-sm shadow-xl shadow-indigo-900/10">
                <div class="flex justify-between items-center text-sm font-medium">
                    <span class="text-indigo-200 flex items-center gap-2">
                        <svg class="w-5 h-5 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                        </svg>
                        Kopiere auf USB... ({{ photobooth.usbExport.album }})
                    </span>
                    <span class="text-indigo-300 font-mono">
                        {{ formatBytes(photobooth.usbExport.copiedBytes) }} / {{
                            formatBytes(photobooth.usbExport.totalBytes) }}
                    </span>
                </div>
                <div class="w-full h-2 bg-indigo-950/50 rounded-full overflow-hidden mt-1">
                    <div class="h-full bg-indigo-500 transition-all duration-300 rounded-full"
                        :style="{ width: (photobooth.usbExport.totalBytes > 0 ? (photobooth.usbExport.copiedBytes / photobooth.usbExport.totalBytes) * 100 : 0) + '%' }">
                    </div>
                </div>
                <div class="flex justify-between text-xs text-indigo-400/70">
                    <span>{{ photobooth.usbExport.copiedFiles }} / {{ photobooth.usbExport.totalFiles }} Dateien</span>
                    <span v-if="photobooth.usbExport.etaSeconds > 0">ETA: {{ formatEta(photobooth.usbExport.etaSeconds)
                    }}</span>
                </div>
                <div v-if="photobooth.usbExport.error"
                    class="text-red-400 text-xs mt-1 font-semibold flex items-center gap-1">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    Fehler: {{ photobooth.usbExport.error }}
                </div>
                <div v-else-if="photobooth.usbExport.copiedBytes > 0 && photobooth.usbExport.copiedBytes === photobooth.usbExport.totalBytes"
                    class="text-emerald-400 text-xs mt-1 font-semibold flex items-center gap-1">
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                    Fertig! Bitte Stick jetzt sicher entfernen.
                </div>
            </div>

            <DashboardOverview :current-album="editSettings.currentAlbum" :gallery-count="galleryCount" />

            <DashboardSettings v-model="editSettings" :gallery-count="galleryCount"
                @gallery-update="updateGalleryCount" />

            <DashboardLogs />

        </main>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue';
import { usePhotoboothStore } from '../stores/photobooth';

import DashboardOverview from '../components/dashboard/DashboardOverview.vue';
import DashboardSettings from '../components/dashboard/DashboardSettings.vue';
import DashboardLogs from '../components/dashboard/DashboardLogs.vue';

const photobooth = usePhotoboothStore();

const galleryCount = ref(0);

const editSettings = reactive({
    countdownSeconds: 3,
    previewDisplaySeconds: 5,
    triggerDelayMs: 0,
    currentAlbum: 'default'
});

// Sync settings when fetched
watch(() => photobooth.settings, (s) => {
    editSettings.countdownSeconds = s.countdownSeconds || 3;
    editSettings.previewDisplaySeconds = s.previewDisplaySeconds || 5;
    editSettings.triggerDelayMs = s.triggerDelayMs || 0;

    // Only update if it's explicitly differing to avoid resetting typing randomly
    if (s.currentAlbum) {
        editSettings.currentAlbum = s.currentAlbum;
    }
}, { deep: true });

onMounted(() => {
    photobooth.register('dashboard');
    photobooth.fetchLogs();
    photobooth.fetchStatus();
    photobooth.fetchSettings();

    // Poll status every 10s
    setInterval(() => {
        photobooth.fetchStatus();
        updateGalleryCount();
    }, 10000);

    updateGalleryCount();
});

async function updateGalleryCount() {
    if (photobooth.settings.currentAlbum || editSettings.currentAlbum) {
        galleryCount.value = await photobooth.fetchGalleryCount(photobooth.settings.currentAlbum || editSettings.currentAlbum);
    }
}

function formatBytes(bytes: number): string {
    if (!bytes) return '0 B';
    if (bytes >= 1024 * 1024 * 1024) return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB';
    if (bytes >= 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return Math.round(bytes / 1024) + ' KB';
}

function formatEta(secs: number): string {
    if (!secs || secs <= 0) return '...';
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    if (m > 0) return `${m}m ${s}s`;
    return `${s}s`;
}
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
