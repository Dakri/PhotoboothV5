<template>
  <div class="min-h-screen bg-gray-900 text-white p-6">
    <h1 class="text-3xl font-bold mb-8">Photobooth Dashboard</h1>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <!-- Status Card -->
        <div class="bg-gray-800 p-6 rounded-xl shadow-lg border border-gray-700">
            <h2 class="text-xl font-semibold mb-2 text-gray-400">System Status</h2>
            <div class="text-4xl font-mono mb-2" :class="statusColor">
                {{ photobooth.state.toUpperCase() }}
            </div>
            <div class="text-sm text-gray-500">
                Connected: {{ photobooth.connected ? 'Yes' : 'No' }} <br>
                Clients: {{ photobooth.clients }}
            </div>
        </div>

        <!-- Controls Card -->
        <div class="bg-gray-800 p-6 rounded-xl shadow-lg border border-gray-700">
            <h2 class="text-xl font-semibold mb-2 text-gray-400">Controls</h2>
            <button @click="photobooth.trigger()" 
                    :disabled="photobooth.state !== 'idle'"
                    class="w-full py-4 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed rounded font-bold transition">
                Trigger Photo
            </button>
        </div>

        <!-- Latest Photo Card -->
        <div class="bg-gray-800 p-6 rounded-xl shadow-lg border border-gray-700">
             <h2 class="text-xl font-semibold mb-2 text-gray-400">Latest Photo</h2>
             <div v-if="photobooth.lastPhoto" class="aspect-video bg-gray-800 rounded overflow-hidden">
                 <img :src="photobooth.lastPhoto.thumbUrl" class="w-full h-full object-cover" />
             </div>
             <div v-else class="aspect-video bg-gray-800 rounded flex items-center justify-center text-gray-600">
                 No photos yet
             </div>
        </div>

        <!-- Links -->
        <div class="bg-gray-800 p-6 rounded-xl shadow-lg border border-gray-700 col-span-full">
            <h2 class="text-xl font-semibold mb-4 text-gray-400">Navigate to Views</h2>
            <div class="flex flex-wrap gap-4">
                <router-link to="/buzzer" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-center transition">Scanner / Buzzer</router-link>
                <router-link to="/countdown" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-center transition">Countdown Monitor</router-link>
                <router-link to="/gallery" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-center transition">Gallery</router-link>
                <a href="/legacy/" class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded text-center transition">Legacy Client (HTML)</a>
            </div>
        </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { usePhotoboothStore } from '../stores/photobooth';

const photobooth = usePhotoboothStore();

const statusColor = computed(() => {
    switch (photobooth.state) {
        case 'idle': return 'text-green-500';
        case 'error': return 'text-red-500';
        default: return 'text-yellow-500';
    }
});

onMounted(() => {
    console.log('registering dashboard')
    photobooth.register('dashboard');
});
</script>

<style scoped>
/* inline classes used instead of @apply */
</style>
