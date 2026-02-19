<template>
    <div class="h-screen w-screen bg-black flex items-center justify-center overflow-hidden">

        <!-- Countdown Number -->
        <div v-if="photobooth.state === 'countdown'" class="relative flex items-center justify-center">
            <!-- Circular progress ring -->
            <svg class="absolute w-80 h-80" viewBox="0 0 200 200">
                <circle cx="100" cy="100" r="90" stroke="#27272a" stroke-width="4" fill="none" />
                <circle cx="100" cy="100" r="90" stroke="#f59e0b" stroke-width="4" fill="none" stroke-linecap="round"
                    :stroke-dasharray="circumference" :stroke-dashoffset="dashOffset"
                    class="transition-all duration-1000 ease-linear" transform="rotate(-90 100 100)" />
            </svg>
            <span class="text-[12rem] font-extralight text-white tabular-nums leading-none select-none"
                :key="photobooth.countdown.remaining">
                {{ photobooth.countdown.remaining }}
            </span>
        </div>

        <!-- Capturing -->
        <div v-else-if="photobooth.state === 'capturing'" class="text-center">
            <div class="w-6 h-6 rounded-full bg-white mx-auto mb-6 animate-ping"></div>
            <p class="text-xl font-light text-zinc-400 uppercase tracking-[0.3em]">Capturing</p>
        </div>

        <!-- Processing -->
        <div v-else-if="photobooth.state === 'processing'" class="text-center">
            <div class="w-8 h-8 border-2 border-zinc-600 border-t-white rounded-full mx-auto mb-6 animate-spin"></div>
            <p class="text-xl font-light text-zinc-400 uppercase tracking-[0.3em]">Processing</p>
        </div>

        <!-- Waiting / Idle -->
        <div v-else class="text-center">
            <p class="text-sm font-light text-zinc-600 uppercase tracking-[0.3em]">Waiting for trigger</p>
        </div>

    </div>
</template>

<script setup lang="ts">
import { onMounted, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { usePhotoboothStore } from '../stores/photobooth';

const photobooth = usePhotoboothStore();
const router = useRouter();

const circumference = 2 * Math.PI * 90; // r=90

const dashOffset = computed(() => {
    const { remaining, total } = photobooth.countdown;
    if (total === 0) return circumference;
    const progress = remaining / total;
    return circumference * (1 - progress);
});

onMounted(() => {
    photobooth.register('countdown');
});

watch(() => photobooth.state, (newState) => {
    if (newState === 'preview') {
        router.push('/preview');
    }
});
</script>
