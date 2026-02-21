<template>
    <div class="bg-zinc-900 border border-zinc-800 rounded-lg overflow-hidden">
        <div class="flex border-b border-zinc-800">
            <button @click="activeTab = 'setup'"
                class="flex-1 py-3 px-5 text-sm font-medium tracking-wide transition-colors cursor-pointer"
                :class="activeTab === 'setup' ? 'bg-zinc-800 text-white border-b-2 border-emerald-500' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'">
                ⚙️ Technical Setup
            </button>
            <button @click="activeTab = 'albums'"
                class="flex-1 py-3 px-5 text-sm font-medium tracking-wide transition-colors cursor-pointer"
                :class="activeTab === 'albums' ? 'bg-zinc-800 text-white border-b-2 border-emerald-500' : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800/50'">
                📸 Alben Verwaltung
            </button>
        </div>

        <DashboardSettingsSetup v-if="activeTab === 'setup'" v-model="localSettings" :active-tab="activeTab"
            :saving="saving" :save-message="saveMessage" :save-success="saveSuccess" @save="handleSaveSettings" />

        <DashboardSettingsAlbums v-if="activeTab === 'albums'" v-model="localSettings" :active-tab="activeTab"
            :gallery-count="galleryCount" @switch-album="switchToAlbum" @gallery-update="$emit('gallery-update')" />
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import DashboardSettingsSetup from './DashboardSettingsSetup.vue';
import DashboardSettingsAlbums from './DashboardSettingsAlbums.vue';
import { usePhotoboothStore } from '../../stores/photobooth';

const props = defineProps<{
    modelValue: any;
    galleryCount: number;
}>();

const emit = defineEmits(['update:modelValue', 'gallery-update']);

const photobooth = usePhotoboothStore();

const activeTab = ref('setup');
const saving = ref(false);
const saveMessage = ref('');
const saveSuccess = ref(false);

const localSettings = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
});

async function handleSaveSettings() {
    saving.value = true;
    saveMessage.value = '';

    const result = await photobooth.saveSettings({
        countdownSeconds: localSettings.value.countdownSeconds,
        previewDisplaySeconds: localSettings.value.previewDisplaySeconds,
        triggerDelayMs: localSettings.value.triggerDelayMs,
        currentAlbum: localSettings.value.currentAlbum
    });

    saving.value = false;
    if (result.success) {
        saveSuccess.value = true;
        saveMessage.value = '✓ Gespeichert';
    } else {
        saveSuccess.value = false;
        saveMessage.value = '✗ Fehler beim Speichern';
    }
    setTimeout(() => { saveMessage.value = ''; }, 3000);
}

async function switchToAlbum(albumId: string) {
    localSettings.value.currentAlbum = albumId;
    await handleSaveSettings();
    emit('gallery-update');
}
</script>
