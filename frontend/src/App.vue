<template>
  <div class="min-h-screen bg-gray-900 text-white font-sans overflow-hidden">
    <!-- Status Bar / Header (Optional) -->
    <header class="fixed top-0 left-0 right-0 p-2 z-50 flex justify-between items-center bg-black/50 backdrop-blur-sm"
      v-if="showStatus">
      <div class="text-xs text-gray-400">
        State: {{ photobooth.state }} | Clients: {{ photobooth.clients }}
      </div>
      <div class="text-xs text-gray-400 ">
        {{ photobooth.connected ? '🟢 Online' : '🔴 Offline' }}
      </div>
    </header>
    <div :class="{ 'pt-8 max-h-screen overflow-auto': showStatus }">
      <!-- Main View -->
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
    <!-- Global Notifications / Overlays could go here -->
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { usePhotoboothStore } from './stores/photobooth';

const photobooth = usePhotoboothStore();
const route = useRoute();

const showStatus = computed(() => route.path === '/' || route.path === '/dashboard');

onMounted(() => {
  photobooth.connect();
});
</script>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
