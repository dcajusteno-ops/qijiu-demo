<template>
  <transition
    enter-active-class="transition-opacity duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-200 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="isOpen && imageA && imageB"
      class="fixed inset-0 z-[100] flex flex-col bg-black overflow-hidden select-none"
      @keydown.esc="close"
      tabindex="-1"
      ref="sliderModalRef"
    >
      <!-- Top Bar: Actions & Info -->
      <div
        class="absolute top-0 left-0 right-0 z-[110] flex justify-between items-center p-4 bg-gradient-to-b from-black/80 to-transparent pointer-events-none"
      >
        <!-- Left side: Image A info -->
        <div
          class="pointer-events-auto bg-black/60 backdrop-blur-md px-3 py-1.5 rounded-lg border border-white/10 flex items-center gap-2"
        >
          <div class="w-3 h-3 rounded-full bg-blue-500"></div>
          <span class="text-white/90 text-sm font-medium">{{ imageA.name }}</span>
        </div>

        <!-- Right side: Close button & Image B info -->
        <div class="flex items-center gap-4 pointer-events-auto">
          <div
            class="bg-black/60 backdrop-blur-md px-3 py-1.5 rounded-lg border border-white/10 flex items-center gap-2"
          >
            <span class="text-white/90 text-sm font-medium">{{ imageB.name }}</span>
            <div class="w-3 h-3 rounded-full bg-pink-500"></div>
          </div>
          <button
            @click="close"
            class="p-2 bg-white/10 hover:bg-white/20 text-white rounded-full transition-colors"
            title="Close Compare"
          >
            <X class="w-6 h-6" />
          </button>
        </div>
      </div>

      <!-- Main Comparison Area -->
      <div class="relative flex-1 w-full h-full flex items-center justify-center p-4 sm:p-12 overflow-hidden">
        <!-- Inner Wrapper to exactly match image dimensions -->
        <div
          ref="containerRef"
          class="relative max-h-[calc(100vh-120px)] max-w-full inline-block"
        >
          <!-- Bottom Image (Image B - Right side) -->
          <img
            :src="imageB.path"
            class="block max-h-[calc(100vh-120px)] max-w-full object-contain pointer-events-none rounded shadow-2xl"
            draggable="false"
            @load="handleImageLoad"
          />

          <!-- Top Image (Image A - Left side) with Clip Path -->
          <img
            v-show="imagesLoaded"
            :src="imageA.path"
            class="absolute top-0 left-0 h-full w-full object-contain pointer-events-none rounded shadow-2xl"
            :style="{ clipPath: `polygon(0 0, ${sliderPosition}% 0, ${sliderPosition}% 100%, 0 100%)` }"
            draggable="false"
          />

          <!-- Draggable Handle Container -->
          <div
            v-show="imagesLoaded"
            class="absolute inset-y-0 z-50 cursor-ew-resize flex items-center justify-center -ml-[20px] w-[40px]"
            :style="{ left: `${sliderPosition}%` }"
            @mousedown.prevent="startDrag"
            @touchstart.passive="startDrag"
          >
            <!-- Vertical Line -->
            <div class="absolute inset-y-0 w-0.5 bg-white shadow-[0_0_5px_rgba(0,0,0,0.5)]"></div>

            <!-- Grabber Button -->
            <div
              class="relative flex items-center justify-center w-8 h-8 bg-white rounded-full shadow-[0_0_10px_rgba(0,0,0,0.3)] text-gray-800 transition-transform"
              :class="{ 'scale-110': isDragging }"
            >
              <ArrowLeftRight class="w-4 h-4" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue';
import { X, ArrowLeftRight } from 'lucide-vue-next';

const props = defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
  imageA: {
    type: Object,
    default: null,
  },
  imageB: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(['close']);

const sliderPosition = ref(50);
const isDragging = ref(false);
const containerRef = ref(null);
const sliderModalRef = ref(null);
const imagesLoaded = ref(false);

const close = () => {
  emit('close');
};

const handleImageLoad = () => {
  imagesLoaded.value = true;
};

const startDrag = (e) => {
  isDragging.value = true;
  window.addEventListener('mousemove', onDrag);
  window.addEventListener('mouseup', stopDrag);
  window.addEventListener('touchmove', onDrag, { passive: false });
  window.addEventListener('touchend', stopDrag);
};

const onDrag = (e) => {
  if (!isDragging.value || !containerRef.value) return;

  // Prevent default scrolling on touch
  if (e.type === 'touchmove') {
    e.preventDefault();
  }

  const clientX = e.touches ? e.touches[0].clientX : e.clientX;

  // Get the bounding rect of the image element itself to calculate exact percentage
  const imgElement = containerRef.value.querySelector('img');
  if (!imgElement) return;

  const rect = imgElement.getBoundingClientRect();

  let x = clientX - rect.left;
  let percent = (x / rect.width) * 100;

  sliderPosition.value = Math.max(0, Math.min(100, percent));
};

const stopDrag = () => {
  isDragging.value = false;
  window.removeEventListener('mousemove', onDrag);
  window.removeEventListener('mouseup', stopDrag);
  window.removeEventListener('touchmove', onDrag);
  window.removeEventListener('touchend', stopDrag);
};

watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal) {
      sliderPosition.value = 50;
      imagesLoaded.value = false;
      nextTick(() => {
        if (sliderModalRef.value) {
          sliderModalRef.value.focus();
        }
      });
    } else {
      stopDrag();
    }
  }
);

onUnmounted(() => {
  stopDrag();
});
</script>