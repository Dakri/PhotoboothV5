<template>
  <div class="flex items-center justify-center h-screen bg-black" @click="returnToCountdown">
      <img v-if="photobooth.lastPhoto" :src="photobooth.lastPhoto.url" class="max-w-full max-h-screen shadow-2xl border-4 border-white" />
      <div v-else class="text-white">Waiting for photo...</div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { usePhotoboothStore } from '../stores/photobooth';

const photobooth = usePhotoboothStore();
const router = useRouter();

function returnToCountdown() {
    // Optional: Allow user to dismiss early?
    // Usually timer based on server decides
}

onMounted(() => {
    photobooth.register('preview');
});

watch(() => photobooth.state, (newState) => {
    if (newState === 'idle') {
        router.push('/countdown');
    }
});
</script>
