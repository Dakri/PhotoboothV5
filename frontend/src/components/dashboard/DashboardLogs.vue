<template>
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
                <span class="shrink-0 w-12 uppercase font-semibold" :class="levelColor(entry.level)">{{ entry.level
                    }}</span>
                <span class="text-zinc-400 shrink-0 w-20">{{ entry.source }}</span>
                <span class="text-zinc-300">{{ entry.message }}</span>
            </div>
            <div v-if="photobooth.logs.length === 0" class="text-zinc-600 text-center py-8">
                No log entries yet. Trigger a capture to see activity.
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';
import { usePhotoboothStore } from '../../stores/photobooth';

const photobooth = usePhotoboothStore();
const logContainer = ref<HTMLElement | null>(null);
const autoScroll = ref(true);

function toggleAutoScroll() {
    autoScroll.value = !autoScroll.value;
}

function scrollToBottom() {
    if (autoScroll.value && logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
}

watch(() => photobooth.logs.length, () => {
    nextTick(scrollToBottom);
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
</script>
