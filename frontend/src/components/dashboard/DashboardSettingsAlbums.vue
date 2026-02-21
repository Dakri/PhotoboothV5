<template>
    <div v-if="activeTab === 'albums'" class="p-5 space-y-8">
        <!-- Active Album Status -->
        <div class="grid grid-cols-1 md:grid-cols-1 gap-6">
            <div class="bg-zinc-950/50 rounded-lg p-4 border border-zinc-800/50">
                <label class="block text-xs text-zinc-300 uppercase tracking-wider mb-2">Aktives Album</label>
                <div class="flex items-center justify-between">
                    <div>
                        <p class="text-lg font-semibold text-emerald-400">{{ activeAlbumOriginalName }}</p>
                        <p class="text-xs text-zinc-500 font-mono mt-1">Intern: {{ localSettings.currentAlbum }}</p>
                    </div>
                    <div class="text-right">
                        <p class="text-2xl font-mono text-white">{{ galleryCount }}</p>
                        <p class="text-xs text-zinc-500 uppercase tracking-wider">Fotos</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Album List -->
        <div>
            <label class="block text-xs text-zinc-500 uppercase tracking-wider mb-3">Alle Alben ({{
                photobooth.albums.length }})</label>
            <div class="flex flex-col gap-2">
                <!-- Neues Album Button inline in list -->
                <div
                    class="p-4 rounded-lg border border-dashed border-zinc-800 hover:border-emerald-500/50 transition-colors flex flex-col justify-center min-h-[72px]">
                    <div v-if="!showNewAlbumInput" @click="showNewAlbumInput = true"
                        class="flex items-center gap-3 cursor-pointer text-emerald-500 hover:text-emerald-400">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2"
                            stroke="currentColor" class="w-6 h-6">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                        </svg>
                        <span class="font-medium">Neues Album erstellen</span>
                    </div>
                    <div v-else class="flex flex-col gap-3 w-full">
                        <div class="flex flex-col lg:flex-row items-start gap-4 w-full">
                            <!-- Inputs (Name and Dropdown) -->
                            <div class="flex flex-col gap-3 w-full max-w-xl">
                                <input type="text" v-model="pendingAlbumName" placeholder="Z.b. Hochzeit Laura & Max"
                                    class="w-full bg-zinc-800 border border-zinc-700 rounded px-4 py-3 text-sm text-zinc-200 focus:outline-none focus:border-emerald-500 transition-colors shadow-inner" />

                                <!-- Custom Dropdown for Capture Method -->
                                <div class="relative w-full">
                                    <button type="button" @click="showCaptureDropdown = !showCaptureDropdown"
                                        class="w-full bg-zinc-800/80 border border-zinc-700 hover:border-zinc-500 rounded px-4 py-3 text-left focus:outline-none focus:ring-1 focus:ring-emerald-500 flex justify-between items-center transition-all cursor-pointer min-h-[76px]">
                                        <div class="pr-4">
                                            <div class="text-sm font-bold" :class="methodColors[pendingCaptureMethod]">
                                                {{ getMethodTitle(pendingCaptureMethod) }}</div>
                                            <div class="text-xs text-zinc-400 mt-1.5 leading-snug line-clamp-2">{{
                                                getMethodDesc(pendingCaptureMethod) }}</div>
                                        </div>
                                        <svg xmlns="http://www.w3.org/2000/svg"
                                            class="h-6 w-6 text-zinc-500 shrink-0 transition-transform duration-200"
                                            :class="showCaptureDropdown ? 'rotate-180' : ''" fill="none"
                                            viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                                d="M19 9l-7 7-7-7" />
                                        </svg>
                                    </button>

                                    <transition name="fade">
                                        <div v-if="showCaptureDropdown"
                                            class="absolute z-20 w-full mt-2 bg-zinc-800 border border-zinc-600 rounded-lg shadow-2xl overflow-hidden divide-y divide-zinc-700">
                                            <div @click="selectCaptureMethod('A')"
                                                class="p-4 hover:bg-zinc-700 cursor-pointer transition-colors"
                                                :class="{ 'bg-emerald-500/5': pendingCaptureMethod === 'A' }">
                                                <div class="text-sm font-bold text-emerald-400">Schnell (A)</div>
                                                <div class="text-xs text-zinc-300 mt-1.5 leading-snug">Nur JPEG
                                                    Download. RAW bleibt sicher auf der Kamera SD-Karte. </div>
                                            </div>
                                            <div @click="selectCaptureMethod('B')"
                                                class="p-4 hover:bg-zinc-700 cursor-pointer transition-colors"
                                                :class="{ 'bg-blue-500/5': pendingCaptureMethod === 'B' }">
                                                <div class="text-sm font-bold text-blue-400">Sicher (B)</div>
                                                <div class="text-xs text-zinc-300 mt-1.5 leading-snug">JPEG Download,
                                                    danach asynchroner RAW Download. RAW bleibt als Backup auf SD-Karte.
                                                </div>
                                            </div>
                                            <div @click="selectCaptureMethod('C')"
                                                class="p-4 hover:bg-zinc-700 cursor-pointer transition-colors"
                                                :class="{ 'bg-amber-500/5': pendingCaptureMethod === 'C' }">
                                                <div class="text-sm font-bold text-amber-400">Lokal (C)</div>
                                                <div class="text-xs text-zinc-300 mt-1.5 leading-snug">JPEG und RAW
                                                    Download auf den Raspberry Pi. SD-Karte füllt sich nicht. (<span
                                                        class="text-emerald-500 font-semibold">Empfohlen</span>)</div>
                                            </div>
                                        </div>
                                    </transition>
                                </div>
                            </div>

                            <!-- Buttons -->
                            <div class="flex lg:flex-col gap-2 shrink-0 lg:w-32 w-full">
                                <button @click="handleCreateOrSwitchAlbum" :disabled="!pendingAlbumName || savingAlbum"
                                    class="flex-1 lg:flex-none px-4 py-3 bg-emerald-600 hover:bg-emerald-500 disabled:bg-zinc-800 disabled:text-zinc-600 border border-transparent disabled:border-zinc-700 text-white text-sm font-medium rounded transition-colors cursor-pointer text-center h-[52px]">
                                    {{ savingAlbum ? '...' : 'Erstellen' }}
                                </button>
                                <button
                                    @click="showNewAlbumInput = false; pendingAlbumName = ''; showCaptureDropdown = false"
                                    class="flex-none px-4 py-3 text-zinc-400 hover:text-white hover:bg-zinc-800 border border-zinc-700 rounded cursor-pointer transition-colors text-center h-[52px]">
                                    Abbrechen
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Existing Albums -->
                <div v-if="photobooth.albums.length > 0">
                    <div v-for="album in photobooth.albums" :key="album.id" @click="switchToAlbum(album.id)"
                        class="p-4 rounded-lg border transition-all cursor-pointer flex flex-col md:flex-row md:items-center justify-between gap-4"
                        :class="localSettings.currentAlbum === album.id ? 'border-emerald-500/50 bg-emerald-500/10' : 'border-zinc-800 bg-zinc-900 hover:border-zinc-600'">

                        <div class="flex items-center gap-3 flex-1 min-w-0">
                            <div class="w-2 h-2 rounded-full shrink-0"
                                :class="localSettings.currentAlbum === album.id ? 'bg-emerald-400' : 'bg-transparent'">
                            </div>
                            <span class="font-medium text-white text-lg truncate" :title="album.name">{{ album.name
                                }}</span>
                            <span v-if="album.captureMethod"
                                class="text-[10px] font-mono bg-zinc-800 text-zinc-400 px-1.5 py-0.5 rounded border border-zinc-700">Methode
                                {{ album.captureMethod }}</span>
                        </div>

                        <div
                            class="flex flex-wrap md:flex-nowrap items-center gap-4 md:gap-6 justify-between w-full md:w-auto mt-2 md:mt-0">
                            <div class="flex items-center gap-4 text-right">
                                <div class="flex flex-col items-end">
                                    <span class="text-sm font-bold text-white">{{ album.count }}</span>
                                    <span class="text-xs text-zinc-500 uppercase">Fotos</span>
                                </div>
                                <div class="flex flex-col items-end w-16">
                                    <span class="text-sm font-bold text-white">{{ formatAlbumSize(album.size) }}</span>
                                    <span class="text-xs text-zinc-500 uppercase">Größe</span>
                                </div>
                            </div>
                            <div
                                class="flex items-center gap-2 border-t md:border-t-0 border-zinc-800/50 pt-3 md:pt-0 w-full md:w-auto justify-end">
                                <button @click.stop="emptyGallery(album.id, album.count)" title="Galerie leeren"
                                    class="px-3 py-1.5 text-xs bg-zinc-800 hover:bg-amber-900/30 text-amber-500 border border-zinc-700 hover:border-amber-700 rounded transition-colors"
                                    :class="{ 'opacity-50 cursor-not-allowed': album.count === 0 }"
                                    :disabled="album.count === 0">
                                    Leeren
                                </button>
                                <button @click.stop="deleteGallery(album.id)" title="Album löschen"
                                    class="px-3 py-1.5 text-xs bg-zinc-800 hover:bg-red-900/30 text-red-500 border border-zinc-700 hover:border-red-700 rounded transition-colors"
                                    :class="{ 'opacity-50 cursor-not-allowed': localSettings.currentAlbum === album.id }"
                                    :disabled="localSettings.currentAlbum === album.id">
                                    Löschen
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- USB Export Section -->
        <div class="pt-6 border-t border-zinc-800">
            <div class="flex items-center justify-between mb-4">
                <label class="block text-xs text-zinc-500 uppercase tracking-wider">USB Export</label>
                <button @click="refreshUsb"
                    class="text-xs text-emerald-400 hover:text-emerald-300 flex items-center gap-1">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4"
                        :class="{ 'animate-spin': usbLoading }">
                        <path fill-rule="evenodd"
                            d="M15.312 11.424a5.5 5.5 0 01-9.201 2.466l-.312-.311h2.433a.75.75 0 000-1.5H3.989a.75.75 0 00-.75.75v4.242a.75.75 0 001.5 0v-2.43l.31.31a7 7 0 0011.712-3.138.75.75 0 00-1.449-.39zm1.23-3.723a.75.75 0 00.219-.53V2.929a.75.75 0 00-1.5 0V5.36l-.31-.31A7 7 0 003.239 8.188a.75.75 0 101.448.389A5.5 5.5 0 0113.89 6.11l.311.31h-2.432a.75.75 0 000 1.5h4.243a.75.75 0 00.53-.219z"
                            clip-rule="evenodd" />
                    </svg>
                    Aktualisieren
                </button>
            </div>

            <div v-if="photobooth.usbDevices.length === 0"
                class="text-sm text-zinc-500 bg-zinc-950/50 p-4 rounded-lg border border-zinc-800/50 text-center">
                Kein USB-Speicher erkannt. Bitte Stick einstecken und aktualisieren.
            </div>
            <div v-else class="space-y-4">
                <div v-for="dev in photobooth.usbDevices" :key="dev.name"
                    class="flex flex-col p-4 bg-zinc-950 border border-zinc-800 rounded-lg gap-3">

                    <!-- Device Header -->
                    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                        <div class="flex flex-col">
                            <span class="text-sm font-medium text-white">{{ dev.label || 'USB Stick' }} <span
                                    class="text-zinc-500 text-xs ml-2">({{ dev.size }})</span></span>
                            <span class="text-xs text-zinc-500 font-mono">{{ dev.name }}</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="text-xs px-2 py-1 rounded inline-flex self-start sm:self-auto border"
                                :class="dev.free ? 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20' : 'text-zinc-500 bg-zinc-800 border-zinc-700'">
                                Frei: {{ dev.free || 'Unbekannt' }}
                            </div>
                            <!-- Safely Remove Button -->
                            <button @click.stop="safelyRemove(dev.name)" :disabled="photobooth.usbExport.active"
                                title="Sicher entfernen"
                                class="text-xs px-2 py-1 rounded border border-zinc-700 text-zinc-400 hover:text-orange-400 hover:border-orange-600 transition-colors disabled:opacity-40 disabled:cursor-not-allowed">
                                ⏏ Entfernen
                            </button>
                        </div>
                    </div>

                    <!-- Space Warning -->
                    <div v-if="spaceWarning(dev)"
                        class="text-xs text-orange-400 bg-orange-900/20 border border-orange-800/30 rounded px-3 py-2 flex items-center gap-2">
                        <svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                        </svg>
                        Möglicherweise nicht genug freier Speicher für alle {{ formatAlbumSize(albumOriginalSize) }}
                        Originaldateien.
                    </div>

                    <!-- Export / Cancel Button -->
                    <div class="flex items-center gap-2 pt-2 border-t border-zinc-900">
                        <button v-if="!photobooth.usbExport.active" @click="startUsbExport(dev.name)"
                            class="flex-1 px-4 py-2 text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white rounded transition-colors flex items-center justify-center gap-2">
                            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                    d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                            </svg>
                            Originale auf Stick kopieren
                        </button>
                        <template v-else-if="photobooth.usbExport.active">
                            <!-- Progress UI -->
                            <div class="flex-1 flex flex-col gap-1.5">
                                <div class="flex justify-between text-xs text-zinc-400">
                                    <span class="text-blue-400 font-medium">{{
                                        formatBytes(photobooth.usbExport.copiedBytes) }} / {{
                                            formatBytes(photobooth.usbExport.totalBytes) }}</span>
                                    <span>{{ photobooth.usbExport.copiedFiles }} / {{ photobooth.usbExport.totalFiles }}
                                        Dateien · ETA {{ formatEta(photobooth.usbExport.etaSeconds) }}</span>
                                </div>
                                <div class="w-full h-2 bg-zinc-800 rounded-full overflow-hidden">
                                    <div class="h-full bg-blue-500 transition-all duration-300 rounded-full"
                                        :style="{ width: (photobooth.usbExport.totalBytes > 0 ? (photobooth.usbExport.copiedBytes / photobooth.usbExport.totalBytes) * 100 : 0) + '%' }">
                                    </div>
                                </div>
                            </div>
                            <button @click="cancelExport"
                                class="px-3 py-2 text-xs font-medium bg-red-900/30 hover:bg-red-900/60 text-red-400 border border-red-800 rounded transition-colors">
                                Abbrechen
                            </button>
                        </template>
                    </div>

                    <!-- Error/Success message -->
                    <div v-if="photobooth.usbExport.error" class="text-xs text-red-400 flex items-center gap-1">
                        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                        {{ photobooth.usbExport.error }}
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { usePhotoboothStore } from '../../stores/photobooth';

const props = defineProps<{
    modelValue: any;
    activeTab: string;
    galleryCount: number;
}>();

const emit = defineEmits(['update:modelValue', 'switch-album', 'gallery-update']);

const photobooth = usePhotoboothStore();

const localSettings = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
});

const activeAlbumOriginalName = computed(() => {
    const album = photobooth.albums.find(a => a.id === localSettings.value.currentAlbum);
    return album ? album.name : localSettings.value.currentAlbum;
});

const showNewAlbumInput = ref(false);
const showCaptureDropdown = ref(false);
const pendingAlbumName = ref('');
const pendingCaptureMethod = ref('C');
const savingAlbum = ref(false);

const usbLoading = ref(false);

async function refreshUsb() {
    usbLoading.value = true;
    await photobooth.fetchUsbDevices();
    usbLoading.value = false;
}

// Size of the current album's `original` folder (from album.size if method counts it)
const albumOriginalSize = computed(() => {
    const album = photobooth.albums.find(a => a.id === localSettings.value.currentAlbum);
    return album ? album.size : 0;
});

function spaceWarning(dev: any): boolean {
    if (!dev.free || albumOriginalSize.value === 0) return false;
    // Parse free space string like "12G" or "800M"
    const raw = dev.free.toUpperCase();
    let freeBytes = 0;
    if (raw.endsWith('G')) freeBytes = parseFloat(raw) * 1024 * 1024 * 1024;
    else if (raw.endsWith('M')) freeBytes = parseFloat(raw) * 1024 * 1024;
    else if (raw.endsWith('K')) freeBytes = parseFloat(raw) * 1024;
    else freeBytes = parseFloat(raw);
    return albumOriginalSize.value > freeBytes;
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

async function safelyRemove(deviceName: string) {
    if (photobooth.usbExport.active) return;
    if (!confirm(`Gerät "${deviceName}" jetzt sicher entfernen?`)) return;
    const res = await fetch('/api/usb/unmount', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ deviceName })
    });
    if (res.ok) {
        await photobooth.fetchUsbDevices();
    } else {
        alert('Entfernen fehlgeschlagen: ' + await res.text());
    }
}

async function cancelExport() {
    await fetch('/api/usb/export/cancel', { method: 'POST' });
};

function formatAlbumSize(bytes: number) {
    if (!bytes) return '0 MB';
    const mb = bytes / (1024 * 1024);
    if (mb > 9999) {
        return (mb / 1024).toFixed(1) + ' GB';
    }
    return Math.round(mb) + ' MB';
}

const methodColors: Record<string, string> = {
    'A': 'text-emerald-400',
    'B': 'text-blue-400',
    'C': 'text-amber-400'
};

function getMethodTitle(method: string) {
    if (method === 'C') return 'Superschnell (C)';
    if (method === 'B') return 'Schnell (B)';
    if (method === 'A') return 'Langsam (A)';
    return 'Unbekannt';
}

function getMethodDesc(method: string) {
    if (method === 'C') return 'Lädt JPEG & RAW direkt auf den Raspberry Pi. Keine SD-Speicherung. (Empfohlen)';
    if (method === 'B') return 'JPEG Download, danach asynchroner RAW Download. RAW bleibt als Backup auf SD.';
    if (method === 'A') return 'Nur JPEG Download. RAW bleibt sicher auf der Kamera SD-Karte.';
    return '';
}

function selectCaptureMethod(method: string) {
    pendingCaptureMethod.value = method;
    showCaptureDropdown.value = false;
}

async function handleCreateOrSwitchAlbum() {
    if (!pendingAlbumName.value) return;
    localSettings.value.currentAlbum = pendingAlbumName.value;

    savingAlbum.value = true;

    // Explicitly send captureStrategy during album creation so backend registers it for the new album
    await photobooth.saveSettings({
        countdownSeconds: localSettings.value.countdownSeconds,
        previewDisplaySeconds: localSettings.value.previewDisplaySeconds,
        currentAlbum: localSettings.value.currentAlbum,
        captureStrategy: pendingCaptureMethod.value,
        triggerDelayMs: localSettings.value.triggerDelayMs
    });

    savingAlbum.value = false;

    await photobooth.fetchSettings();
    emit('gallery-update');

    pendingAlbumName.value = '';
    showNewAlbumInput.value = false;
    showCaptureDropdown.value = false;
}

function switchToAlbum(albumId: string) {
    emit('switch-album', albumId);
}

async function emptyGallery(albumId: string, count: number) {
    if (!confirm(`Sicher, dass du alle ${count} Fotos aus "${albumId}" löschen willst?`)) return;

    const success = await photobooth.emptyGallery(albumId);
    if (success) {
        if (albumId === localSettings.value.currentAlbum) {
            emit('gallery-update');
        }
        await photobooth.fetchSettings();
        alert('Galerie wurde geleert.');
    } else {
        alert('Fehler beim Leeren der Galerie.');
    }
}

async function deleteGallery(albumId: string) {
    if (albumId === localSettings.value.currentAlbum) {
        alert('Aktives Album kann nicht gelöscht werden.');
        return;
    }

    if (!confirm(`WARNUNG: Album "${albumId}" wirklich KOMPLETT löschen?`)) return;

    const result = await photobooth.deleteGallery(albumId);
    if (result.success) {
        alert('Album gelöscht.');
        await photobooth.fetchSettings();
    } else {
        alert('Fehler: ' + (result.error || 'Unbekannt'));
    }
}

async function startUsbExport(deviceName: string) {
    if (photobooth.usbExport.active) { alert('Ein Export läuft bereits!'); return; }
    if (!confirm(`Originaldateien von "${localSettings.value.currentAlbum}" auf "${deviceName}" kopieren?`)) return;

    const res = await photobooth.exportToUsb(deviceName, localSettings.value.currentAlbum, 'jpeg_only');
    if (!res.success) {
        alert('Export Fehler: ' + res.error);
    }
}
</script>
