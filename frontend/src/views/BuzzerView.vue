<template>
  <div class="flex flex-col items-center justify-center h-screen bg-gradient-to-br from-indigo-900 to-purple-900" 
       @click="trigger" 
       :class="{'cursor-pointer': photobooth.state === 'idle', 'opacity-50': photobooth.state !== 'idle'}">
    
    <div class="text-center p-10 bg-white/10 rounded-full backdrop-blur-lg shadow-2xl border border-white/20 transition-transform active:scale-95 duration-200">
      <div v-if="photobooth.state === 'idle'">
          <span class="text-8xl">📸</span>
          <h1 class="text-4xl font-bold mt-4 text-white uppercase tracking-wider">Touch to Start</h1>
      </div>
       <div v-else>
          <span class="text-8xl animate-pulse">⏳</span>
          <h1 class="text-4xl font-bold mt-4 text-white uppercase tracking-wider">{{ photobooth.state }}</h1>
      </div>
    </div>

    <div class="absolute bottom-10 text-gray-400 text-sm">
        Photobooth V5 • {{ photobooth.connected ? 'Connected' : 'Connecting...' }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { usePhotoboothStore } from '../stores/photobooth';

const photobooth = usePhotoboothStore();

function trigger() {
    if (photobooth.state === 'idle') {
        photobooth.trigger();
    }
}

onMounted(() => {
    photobooth.register('buzzer');
});
</script>
