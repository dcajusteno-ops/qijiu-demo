<script setup>
import { ChevronLeft, ChevronRight, Loader2, RotateCcw } from 'lucide-vue-next'

defineProps({
  image: { type: Object, default: null },
  currentDisplayImage: { type: Object, default: null },
  canGoPrev: { type: Boolean, default: false },
  canGoNext: { type: Boolean, default: false },
  displayImageSrc: { type: String, default: '' },
  fullImageLoading: { type: Boolean, default: false },
  scale: { type: Number, default: 1 },
  offset: {
    type: Object,
    default: () => ({ x: 0, y: 0 }),
  },
  isDragging: { type: Boolean, default: false },
  imageCounter: { type: String, default: '1 / 1' },
  stackCurrentIndex: { type: Number, default: 1 },
})

defineEmits([
  'prev',
  'next',
  'prev-stack',
  'next-stack',
  'reset-zoom',
  'viewer-wheel',
  'viewer-mousedown',
  'viewer-mousemove',
  'viewer-mouseup',
])
</script>

<template>
  <div class="absolute inset-y-0 left-0 right-[408px]">
    <button
      v-if="canGoPrev"
      class="absolute left-8 top-1/2 z-[70] flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-all hover:scale-110 hover:bg-white/20"
      @click="$emit('prev')"
    >
      <ChevronLeft class="h-8 w-8" />
    </button>

    <button
      v-if="canGoNext"
      class="absolute right-8 top-1/2 z-[70] flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-all hover:scale-110 hover:bg-white/20"
      @click="$emit('next')"
    >
      <ChevronRight class="h-8 w-8" />
    </button>

    <div
      class="relative flex h-full w-full flex-col items-center justify-center overflow-hidden px-24"
      :class="{ 'cursor-grab': scale > 1 && !isDragging, 'cursor-grabbing': isDragging }"
      @wheel="$emit('viewer-wheel', $event)"
      @mousedown="$emit('viewer-mousedown', $event)"
      @mousemove="$emit('viewer-mousemove', $event)"
      @mouseup="$emit('viewer-mouseup')"
      @mouseleave="$emit('viewer-mouseup')"
    >
      <div
        v-if="fullImageLoading"
        class="absolute top-8 z-[65] flex items-center gap-2 rounded-full border border-white/10 bg-black/55 px-3 py-1.5 text-sm text-white/80 shadow-lg backdrop-blur-xl"
      >
        <Loader2 class="h-4 w-4 animate-spin" />
        <span>正在载入原图</span>
      </div>
      <div
        class="relative select-none transition-transform duration-75 ease-out"
        :style="{ transform: `translate(${offset.x}px, ${offset.y}px) scale(${scale})` }"
      >
        <img
          :src="displayImageSrc || currentDisplayImage?.path"
          :alt="currentDisplayImage?.name || ''"
          loading="eager"
          decoding="async"
          class="pointer-events-none max-h-[calc(100vh-120px)] max-w-full rounded object-contain shadow-2xl"
        />
      </div>
    </div>
  </div>

  <div
    class="group absolute bottom-10 left-8 z-[60] flex min-w-[220px] max-w-md flex-col gap-3 rounded-xl border border-white/10 bg-black/70 p-4 text-white shadow-2xl backdrop-blur-xl transition-all hover:bg-black/80"
    @click.stop
  >
    <div class="flex items-center">
      <span class="truncate text-base font-semibold tracking-wide text-white/90">{{ currentDisplayImage?.name }}</span>
    </div>

    <div class="h-px w-full bg-white/10" />

    <div class="flex items-center justify-between text-xs">
      <div class="flex items-center gap-3 font-mono text-white/60">
        <span>{{ imageCounter }}</span>
        <span class="h-3 w-px bg-white/20" />
        <span :class="{ 'font-bold text-blue-400': scale !== 1 }">{{ Math.round(scale * 100) }}%</span>
      </div>

      <button
        v-if="scale !== 1 || offset.x !== 0 || offset.y !== 0"
        class="-mr-2 flex items-center gap-1.5 rounded px-2 py-1 text-xs font-medium text-blue-300 transition-all hover:bg-white/20 hover:text-blue-200"
        title="重置缩放"
        @click="$emit('reset-zoom')"
      >
        <RotateCcw class="h-3 w-3" />
        重置视图
      </button>
    </div>
  </div>

  <div
    v-if="image?.isStackPrimary && image.stackCount > 1"
    class="absolute bottom-10 left-[50%] z-[60] flex -translate-x-1/2 items-center gap-4 rounded-xl border border-white/10 bg-black/70 px-4 py-2 shadow-2xl backdrop-blur-xl"
    @click.stop
    @wheel.stop
  >
    <button
      class="flex h-8 w-8 items-center justify-center rounded-full text-white/70 transition-all hover:bg-white/20 hover:text-white"
      @click="$emit('prev-stack')"
    >
      <ChevronLeft class="h-5 w-5" />
    </button>

    <div class="flex flex-col items-center">
      <span class="mb-0.5 text-[10px] font-semibold uppercase tracking-widest text-white/50">连拍组</span>
      <span class="text-xs font-mono font-medium text-white/90">
        {{ stackCurrentIndex }} / {{ image.stackCount }}
      </span>
    </div>

    <button
      class="flex h-8 w-8 items-center justify-center rounded-full text-white/70 transition-all hover:bg-white/20 hover:text-white"
      @click="$emit('next-stack')"
    >
      <ChevronRight class="h-5 w-5" />
    </button>
  </div>
</template>
