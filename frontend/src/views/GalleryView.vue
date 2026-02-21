<template>
    <div class="min-h-screen bg-black p-4 text-white font-sans select-none">
        <!-- Header -->
        <div
            class="flex justify-between items-center mb-6 sticky top-0 z-10 bg-black/80 backdrop-blur-sm p-2 rounded-lg">
            <h1 class="text-2xl font-bold">Gallery</h1>
            <button @click="router.push('/')"
                class="px-4 py-2 bg-zinc-800 text-white rounded hover:bg-zinc-700 transition">Back</button>
        </div>

        <!-- Loading -->
        <div v-if="gallery.loading" class="text-center mt-20 text-zinc-400">
            Loading photos...
        </div>

        <!-- Error -->
        <div v-if="gallery.error" class="text-red-500 text-center mt-20">
            {{ gallery.error }}
        </div>

        <!-- Grid -->
        <div v-else class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-2">
            <div v-for="(photo, index) in gallery.photos" :key="photo.filename"
                class="aspect-square relative group overflow-hidden rounded bg-zinc-900 cursor-pointer"
                @click="openLightbox(index)">
                <img :src="photo.thumbUrl" loading="lazy"
                    class="w-full h-full object-cover transition duration-300 group-hover:scale-105" />
            </div>
        </div>

        <div v-if="!gallery.loading && gallery.photos.length === 0" class="text-center mt-20 text-zinc-500">
            No photos yet.
        </div>

        <!-- Lightbox -->
        <Transition name="fade">
            <div v-if="lightboxIndex !== null" class="fixed inset-0 z-50 bg-black/95 flex items-center justify-center"
                @click.self="closeLightbox" @touchstart="onTouchStart" @touchend="onTouchEnd">

                <!-- Close Button -->
                <button class="absolute top-4 right-4 text-white/50 hover:text-white z-50 p-4" @click="closeLightbox">
                    <svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>

                <!-- Navigation Buttons (Desktop) -->
                <button
                    class="absolute left-4 top-1/2 -translate-y-1/2 p-4 text-white/30 hover:text-white hidden md:block"
                    @click.stop="prevPhoto">
                    <svg class="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                    </svg>
                </button>
                <button
                    class="absolute right-4 top-1/2 -translate-y-1/2 p-4 text-white/30 hover:text-white hidden md:block"
                    @click.stop="nextPhoto">
                    <svg class="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                    </svg>
                </button>

                <!-- Image -->
                <div
                    class="relative max-w-full max-h-screen p-2 overflow-hidden flex items-center justify-center w-full h-full">
                    <img :src="currentPhoto?.url" class="max-w-full max-h-full object-contain shadow-2xl" />

                    <!-- Counter -->
                    <div
                        class="absolute bottom-6 left-1/2 -translate-x-1/2 text-white/50 text-sm font-mono bg-black/50 px-3 py-1 rounded-full">
                        {{ lightboxIndex + 1 }} / {{ gallery.photos.length }}
                    </div>
                </div>
            </div>
        </Transition>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref, computed, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { useGalleryStore } from '../stores/gallery';

const gallery = useGalleryStore();
const router = useRouter();
const lightboxIndex = ref<number | null>(null);

const currentPhoto = computed(() => {
    if (lightboxIndex.value === null) return null;
    return gallery.photos[lightboxIndex.value];
});

function openLightbox(index: number) {
    lightboxIndex.value = index;
}

function closeLightbox() {
    lightboxIndex.value = null;
}

function nextPhoto() {
    if (lightboxIndex.value === null) return;
    if (lightboxIndex.value < gallery.photos.length - 1) {
        lightboxIndex.value++;
    } else {
        lightboxIndex.value = 0; // Loop
    }
}

function prevPhoto() {
    if (lightboxIndex.value === null) return;
    if (lightboxIndex.value > 0) {
        lightboxIndex.value--;
    } else {
        lightboxIndex.value = gallery.photos.length - 1; // Loop
    }
}

// Swipe Logic
const touchStartX = ref(0);
const touchEndX = ref(0);

function onTouchStart(e: TouchEvent) {
    touchStartX.value = e.changedTouches[0].screenX;
}

function onTouchEnd(e: TouchEvent) {
    touchEndX.value = e.changedTouches[0].screenX;
    handleSwipe();
}

function handleSwipe() {
    const diff = touchEndX.value - touchStartX.value;
    if (Math.abs(diff) < 50) return; // Threshold

    if (diff > 0) {
        prevPhoto(); // Swipe Right -> Prev
    } else {
        nextPhoto(); // Swipe Left -> Next
    }
}

// Keyboard Navigation
function handleKeydown(e: KeyboardEvent) {
    if (lightboxIndex.value === null) return;
    if (e.key === 'ArrowRight') nextPhoto();
    if (e.key === 'ArrowLeft') prevPhoto();
    if (e.key === 'Escape') closeLightbox();
}

onMounted(() => {
    gallery.fetchPhotos();
    window.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown);
});
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
