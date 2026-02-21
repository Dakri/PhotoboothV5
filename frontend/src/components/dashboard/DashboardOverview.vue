<template>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <!-- System Control (State + Trigger + Pi Storage) -->
        <div
            class="bg-zinc-900 border border-zinc-800 rounded-lg p-5 flex flex-col justify-between relative overflow-hidden">
            <div>
                <div class="flex items-center justify-between mb-4">
                    <p class="text-xs text-zinc-500 uppercase tracking-wider">System Control</p>
                    <div class="flex items-center gap-2">
                        <span class="w-2 h-2 rounded-full"
                            :class="photobooth.connected ? 'bg-emerald-500' : 'bg-red-500'"></span>
                    </div>
                </div>

                <div class="flex items-center gap-3 mb-6">
                    <span class="text-3xl font-mono font-semibold tracking-tight" :class="stateColor">{{
                        photobooth.state.toUpperCase() }}</span>
                </div>

                <!-- Gallery Stats -->
                <div class="mb-4">
                    <div class="flex justify-between text-xs mb-1">
                        <span class="text-zinc-500">Aktuelle Galerie</span>
                        <span class="font-mono text-emerald-400 font-medium">{{ currentAlbum }}</span>
                    </div>
                    <div class="flex justify-between text-xs text-zinc-500">
                        <span>Fotos</span>
                        <span class="font-mono text-zinc-300">{{ galleryCount }}</span>
                    </div>
                </div>

                <!-- Pi Storage -->
                <div class="mb-4">
                    <div class="flex justify-between text-xs text-zinc-500 mb-1">
                        <span>Photobooth Speicher</span>
                        <span>{{ formatBytes(photobooth.diskInfo.free) }} frei</span>
                    </div>
                    <div class="w-full h-1.5 bg-zinc-800 rounded-full overflow-hidden">
                        <div class="h-full rounded-full transition-all duration-500" :class="diskUsageColor"
                            :style="{ width: photobooth.diskInfo.usedPercent + '%' }">
                        </div>
                    </div>
                </div>

                <div class="flex items-center justify-between text-xs text-zinc-500 font-mono pr-14">
                    <span>UPTIME {{ photobooth.uptime }}</span>
                    <span>{{ photobooth.clients }} CLIENT{{ photobooth.clients !== 1 ? 'S' : '' }}</span>
                </div>
            </div>

            <button @click="photobooth.trigger()" :disabled="photobooth.state !== 'idle'"
                class="absolute bottom-5 right-5 w-12 h-12 rounded-full flex items-center justify-center shadow-lg transition-all duration-200"
                :class="photobooth.state === 'idle' ? 'bg-emerald-600 hover:bg-emerald-500 text-white cursor-pointer hover:scale-105' : 'bg-zinc-800 text-zinc-600 cursor-not-allowed'">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                    stroke="currentColor" class="w-6 h-6">
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z" />
                    <path stroke-linecap="round" stroke-linejoin="round"
                        d="M16.5 12.75a4.5 4.5 0 11-9 0 4.5 4.5 0 019 0zM18.75 10.5h.008v.008h-.008V10.5z" />
                </svg>
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
        <div class="bg-zinc-900 border border-zinc-800 rounded-lg p-5 relative">
            <p class="text-xs text-zinc-500 uppercase tracking-wider mb-2">Camera</p>
            <div v-if="photobooth.cameraInfo.connected">
                <!-- Top Right Action Button -->
                <button @click="openSdBrowser" title="SD Karte durchsuchen"
                    class="absolute top-4 right-4 p-1.5 text-zinc-500 hover:text-emerald-400 bg-zinc-800 hover:bg-zinc-700 rounded transition-colors cursor-pointer flex items-center gap-2">
                    <span class="text-xs font-medium px-1">SD Inhalte prüfen</span>
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                        stroke="currentColor" class="w-4 h-4">
                        <path stroke-linecap="round" stroke-linejoin="round"
                            d="M3.75 9.776c.112-.017.227-.026.344-.026h15.812c.117 0 .232.009.344.026m-16.5 0a2.25 2.25 0 00-1.883 2.542l.857 6a2.25 2.25 0 002.227 1.932H19.05a2.25 2.25 0 002.227-1.932l.857-6a2.25 2.25 0 00-1.883-2.542m-16.5 0V6A2.25 2.25 0 016 3.75h3.879a1.5 1.5 0 011.06.44l2.122 2.12a1.5 1.5 0 001.06.44H18A2.25 2.25 0 0120.25 9v.776" />
                    </svg>
                </button>
                <div class="flex items-center gap-2 mb-3">
                    <div class="w-2 h-2 rounded-full bg-emerald-400"></div>
                    <span class="text-sm font-semibold text-zinc-200">{{ photobooth.cameraInfo.model }}</span>
                </div>
                <div class="space-y-1.5 text-xs text-zinc-400">
                    <div v-if="photobooth.cameraInfo.manufacturer" class="flex justify-between">
                        <span class="text-zinc-500">Hersteller</span>
                        <span>{{ photobooth.cameraInfo.manufacturer }}</span>
                    </div>
                    <div v-if="photobooth.cameraInfo.serialNumber" class="flex justify-between">
                        <span class="text-zinc-500">Seriennr.</span>
                        <span class="font-mono">{{ photobooth.cameraInfo.serialNumber }}</span>
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
                                <div class="h-full rounded-full transition-all duration-500" :class="batteryBarColor"
                                    :style="{ width: photobooth.cameraInfo.batteryPercent + '%' }"></div>
                            </div>
                            <span class="font-mono">{{ photobooth.cameraInfo.batteryPercent }}%</span>
                        </div>
                    </div>
                    <div v-if="photobooth.cameraInfo.storageFree" class="flex justify-between items-center">
                        <span class="text-zinc-500">Speicher</span>
                        <div class="flex items-center gap-2">
                            <div class="w-16 h-1.5 bg-zinc-700 rounded-full overflow-hidden">
                                <div class="h-full rounded-full transition-all duration-500 bg-sky-400"
                                    :style="{ width: storageUsedPercent + '%' }"></div>
                            </div>
                            <span class="font-mono">{{ photobooth.cameraInfo.storageFree }}</span>
                        </div>
                    </div>
                </div>
            </div>
            <div v-else class="flex flex-col items-center justify-center py-4 gap-2">
                <div class="w-2 h-2 rounded-full bg-red-400 animate-pulse"></div>
                <span class="text-zinc-600 text-sm">Keine Kamera erkannt</span>
            </div>
        </div>

        <!-- SD Card Browser Modal -->
        <transition name="fade">
            <div v-if="showSdBrowser" class="fixed inset-0 z-50 flex items-center justify-center p-4">
                <div class="absolute inset-0 bg-black/80 backdrop-blur-sm" @click="showSdBrowser = false"></div>
                <div
                    class="relative bg-zinc-900 border border-zinc-700 rounded-xl shadow-2xl w-full max-w-3xl flex flex-col max-h-[85vh]">
                    <div class="flex items-center justify-between p-5 border-b border-zinc-800">
                        <div>
                            <h3 class="text-lg font-semibold text-white">Camera SD-Karte</h3>
                            <p class="text-xs text-zinc-500 mt-0.5">Liest Dateien direkt vom Kameraspeicher via USB</p>
                        </div>
                        <button @click="showSdBrowser = false"
                            class="text-zinc-500 hover:text-white p-2 cursor-pointer">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                                stroke="currentColor" class="w-6 h-6">
                                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    </div>

                    <div v-if="isLoadingSd" class="p-12 flex flex-col items-center justify-center text-zinc-500 gap-4">
                        <svg class="animate-spin h-8 w-8 text-emerald-500" xmlns="http://www.w3.org/2000/svg"
                            fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4">
                            </circle>
                            <path class="opacity-75" fill="currentColor"
                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                            </path>
                        </svg>
                        <span>Dateiliste wird von der Kamera geladen... (Kann einige Sekunden dauern)</span>
                    </div>

                    <div v-else-if="photobooth.cameraFiles.length === 0"
                        class="p-12 flex flex-col items-center justify-center text-zinc-500 gap-3">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5"
                            stroke="currentColor" class="w-12 h-12">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
                        </svg>
                        <span>Keine Dateien auf der SD-Karte gefunden.</span>
                    </div>

                    <div v-else class="overflow-y-auto p-2">
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                            <div v-for="file in photobooth.cameraFiles" :key="file.name"
                                class="flex items-center justify-between p-3 bg-zinc-950/50 border border-zinc-800 rounded-lg hover:border-zinc-700 transition-colors">
                                <div class="flex items-center gap-3 overflow-hidden">
                                    <svg v-if="file.name.toLowerCase().endsWith('.jpg') || file.name.toLowerCase().endsWith('.jpeg')"
                                        xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                        stroke-width="1.5" stroke="currentColor" class="w-5 h-5 text-zinc-500 shrink-0">
                                        <path stroke-linecap="round" stroke-linejoin="round"
                                            d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 001.5-1.5V6a1.5 1.5 0 00-1.5-1.5H3.75A1.5 1.5 0 002.25 6v12a1.5 1.5 0 001.5 1.5zm10.5-11.25h.008v.008h-.008V8.25zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" />
                                    </svg>
                                    <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
                                        stroke-width="1.5" stroke="currentColor"
                                        class="w-5 h-5 text-amber-500 shrink-0">
                                        <path stroke-linecap="round" stroke-linejoin="round"
                                            d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m3.75 9v6m3-3H9m1.5-12H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                                    </svg>
                                    <span class="text-sm font-mono text-zinc-300 truncate" :title="file.name">{{
                                        file.name }}</span>
                                </div>
                                <span class="text-xs text-zinc-500 font-mono whitespace-nowrap">{{ Math.round(file.size
                                    / 1024) }} MB</span>
                            </div>
                        </div>
                    </div>

                    <div
                        class="p-4 border-t border-zinc-800 bg-zinc-950/50 flex justify-between items-center mt-auto rounded-b-xl">
                        <span class="text-xs text-zinc-500">Gesamt: {{ photobooth.cameraFiles.length }} Dateien</span>
                        <button @click="openSdBrowser"
                            class="px-3 py-1.5 text-xs border border-zinc-700 hover:border-zinc-500 hover:bg-zinc-800 rounded bg-zinc-900 text-zinc-300 transition-colors cursor-pointer flex items-center gap-2">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                                stroke="currentColor" class="w-3.5 h-3.5">
                                <path stroke-linecap="round" stroke-linejoin="round"
                                    d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99" />
                            </svg>
                            Liste Aktualisieren
                        </button>
                    </div>
                </div>
            </div>
        </transition>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { usePhotoboothStore } from '../../stores/photobooth';

defineProps<{
    currentAlbum: string;
    galleryCount: number;
}>();

const photobooth = usePhotoboothStore();

const showSdBrowser = ref(false);
const isLoadingSd = ref(false);

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

const batteryBarColor = computed(() => {
    const level = photobooth.cameraInfo.batteryPercent;
    if (level === undefined) return 'bg-zinc-700';
    if (level > 50) return 'bg-emerald-400';
    if (level > 20) return 'bg-amber-400';
    return 'bg-red-400';
});

const diskUsageColor = computed(() => {
    const p = photobooth.diskInfo.usedPercent;
    if (p > 90) return 'bg-red-500';
    if (p > 75) return 'bg-amber-500';
    return 'bg-blue-500';
});

const storageUsedPercent = computed(() => {
    if (photobooth.cameraInfo.storagePercent !== undefined && photobooth.cameraInfo.storagePercent > 0) {
        return 100 - photobooth.cameraInfo.storagePercent;
    }
    const total = parseStorageValue(photobooth.cameraInfo.storageTotal);
    const free = parseStorageValue(photobooth.cameraInfo.storageFree);
    if (total <= 0) return 0;
    return Math.round(((total - free) / total) * 100);
});

function formatBytes(bytes: number, decimals = 2) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

function parseStorageValue(val: string): number {
    if (!val) return 0;
    const match = val.match(/([\d.]+)\s*(GB|MB|KB|TB)/i);
    if (!match) return 0;
    const num = parseFloat(match[1]);
    switch (match[2].toUpperCase()) {
        case 'TB': return num * 1024;
        case 'GB': return num;
        case 'MB': return num / 1024;
        case 'KB': return num / (1024 * 1024);
        default: return num;
    }
}

async function openSdBrowser() {
    showSdBrowser.value = true;
    isLoadingSd.value = true;
    try {
        await photobooth.fetchCameraFiles();
    } finally {
        isLoadingSd.value = false;
    }
}
</script>
