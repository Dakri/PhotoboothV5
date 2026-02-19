<template>
  <div class="min-h-screen bg-black p-4">
    <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold text-white">Gallery</h1>
        <button @click="router.push('/')" class="px-4 py-2 bg-gray-800 text-white rounded hover:bg-gray-700 transition">Back</button>
    </div>

    <!-- Loading -->
    <div v-if="gallery.loading" class="text-white text-center mt-20">
        Loading photos...
    </div>

    <!-- Error -->
    <div v-if="gallery.error" class="text-red-500 text-center mt-20">
        {{ gallery.error }}
    </div>

    <!-- Grid -->
    <div v-else class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        <div v-for="photo in gallery.photos" :key="photo.filename" 
             class="aspect-square relative group overflow-hidden rounded-lg bg-gray-900 cursor-pointer"
             @click="openLightbox(photo)">
            <img :src="photo.thumbUrl" loading="lazy" class="w-full h-full object-cover transition duration-300 group-hover:scale-110" />
        </div>
    </div>

    <!-- Lightbox -->
    <div v-if="lightboxPhoto" class="fixed inset-0 z-50 bg-black/95 flex items-center justify-center p-4" @click="closeLightbox">
        <img :src="lightboxPhoto.url" class="max-w-full max-h-screen shadow-2xl" />
        <button class="absolute top-4 right-4 text-white text-4xl">&times;</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useGalleryStore, type Photo } from '../stores/gallery';

const gallery = useGalleryStore();
const router = useRouter();
const lightboxPhoto = ref<Photo | null>(null);

function openLightbox(photo: Photo) {
    lightboxPhoto.value = photo;
}

function closeLightbox() {
    lightboxPhoto.value = null;
}

onMounted(() => {
    gallery.fetchPhotos();
});
</script>
