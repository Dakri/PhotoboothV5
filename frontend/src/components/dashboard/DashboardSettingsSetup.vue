<template>
    <div v-if="activeTab === 'setup'" class="p-5 space-y-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Countdown -->
            <div>
                <label class="block text-xs text-zinc-500 uppercase tracking-wider mb-2">Countdown (Sekunden)</label>
                <input type="number" v-model.number="localSettings.countdownSeconds" min="1" max="10"
                    class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 font-mono focus:outline-none focus:border-zinc-500" />
                <p class="text-xs text-zinc-600 mt-1">1–10 Sekunden</p>
            </div>

            <!-- Preview Duration -->
            <div>
                <label class="block text-xs text-zinc-500 uppercase tracking-wider mb-2">Preview-Dauer
                    (Sekunden)</label>
                <input type="number" v-model.number="localSettings.previewDisplaySeconds" min="1" max="30"
                    class="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-200 font-mono focus:outline-none focus:border-zinc-500" />
                <p class="text-xs text-zinc-600 mt-1">1–30 Sekunden</p>
            </div>

            <!-- Trigger Delay Offset -->
            <div class="md:col-span-2">
                <div class="flex items-center justify-between mb-2">
                    <label class="block text-xs text-zinc-500 uppercase tracking-wider">Auslöse-Verzögerung</label>
                    <span class="text-sm font-mono"
                        :class="localSettings.triggerDelayMs < 0 ? 'text-amber-400' : (localSettings.triggerDelayMs > 0 ? 'text-blue-400' : 'text-zinc-400')">
                        {{ localSettings.triggerDelayMs > 0 ? '+' : '' }}{{ localSettings.triggerDelayMs }} ms
                    </span>
                </div>
                <input type="range" v-model.number="localSettings.triggerDelayMs" min="-3000" max="1000" step="50"
                    class="w-full accent-emerald-500 bg-zinc-800 h-2 rounded-lg appearance-none cursor-pointer" />
                <div class="flex justify-between text-xs text-zinc-600 mt-2 font-mono">
                    <span>-3000ms (Kamera früher)</span>
                    <span>0</span>
                    <span>+1000ms (Später)</span>
                </div>
            </div>
        </div>

        <div class="flex items-center gap-3 pt-2">
            <button @click="$emit('save')" :disabled="saving"
                class="px-5 py-2 rounded text-sm font-semibold tracking-wide transition-all duration-200 cursor-pointer"
                :class="saving ? 'bg-zinc-700 text-zinc-500 cursor-not-allowed' : 'bg-emerald-600 hover:bg-emerald-500 text-white'">
                {{ saving ? 'Speichern...' : 'Setup Speichern' }}
            </button>
            <span v-if="saveMessage" class="text-xs" :class="saveSuccess ? 'text-emerald-400' : 'text-red-400'">
                {{ saveMessage }}
            </span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
    modelValue: any;
    activeTab: string;
    saving: boolean;
    saveMessage: string;
    saveSuccess: boolean;
}>();

const emit = defineEmits(['update:modelValue', 'save']);

const localSettings = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
});
</script>
