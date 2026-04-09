<script setup>
import { useProvideCarousel } from './useCarousel'

const props = defineProps({
  opts: {
    type: Object,
    default: () => ({}),
  },
  plugins: {
    type: Array,
    default: () => [],
  },
  orientation: {
    type: String,
    default: 'horizontal',
  },
  class: {
    type: String,
    default: '',
  },
})

const emits = defineEmits(['init-api'])

const { carouselRef, carouselApi, canScrollPrev, canScrollNext, scrollPrev, scrollNext } = useProvideCarousel(
  props.opts,
  props.plugins,
  props.orientation,
  emits
)

defineExpose({
  carouselApi,
  canScrollPrev,
  canScrollNext,
  scrollPrev,
  scrollNext,
})
</script>

<template>
  <div
    :class="['relative', props.class]"
    role="region"
    aria-roledescription="carousel"
  >
    <slot />
  </div>
</template>
