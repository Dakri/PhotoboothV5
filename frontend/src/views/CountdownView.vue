<template>
  <div class="flex items-center justify-center h-screen bg-black text-white">
      <div v-if="photobooth.state === 'countdown'" class="text-[20rem] font-bold text-yellow-500 animate-ping">
          {{ photobooth.countdown.remaining }}
      </div>
      <div v-else-if="photobooth.state === 'capturing'" class="text-6xl font-bold text-white animate-pulse">
          CHEESE! 🧀
      </div>
      <div v-else-if="photobooth.state === 'processing'" class="text-4xl font-bold text-blue-400 animate-bounce">
          Processing... ⚙️
      </div>
      <div v-else class="text-gray-500 text-xl">
          Waiting...
      </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { usePhotoboothStore } from '../stores/photobooth';

const photobooth = usePhotoboothStore();
const router = useRouter();

onMounted(() => {
    photobooth.register('countdown');
});

watch(() => photobooth.state, (newState) => {
    if (newState === 'preview') {
        router.push('/preview');
    }
});
</script>
