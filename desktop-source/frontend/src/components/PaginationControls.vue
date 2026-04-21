<template>
  <div
    :class="
      compact
        ? 'flex items-center gap-3 rounded-2xl border border-border/70 bg-background/95 px-3 py-2 shadow-xl backdrop-blur supports-[backdrop-filter]:bg-background/80 select-none'
        : 'flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-border/70 bg-background/95 px-4 py-3 shadow-sm backdrop-blur supports-[backdrop-filter]:bg-background/80 select-none'
    "
  >
    <div
      :class="
        compact
          ? 'rounded-full bg-muted/60 px-2.5 py-1 text-xs text-muted-foreground whitespace-nowrap'
          : 'text-sm text-muted-foreground'
      "
    >
      <template v-if="compact">
        {{ startIndex + 1 }}-{{ endIndex }} / {{ totalItems }}
      </template>
      <template v-else>
        显示 <span class="font-medium text-foreground">{{ startIndex + 1 }}-{{ endIndex }}</span>
        共 <span class="font-medium text-foreground">{{ totalItems }}</span> 项
      </template>
    </div>

    <div class="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === 1"
        title="首页"
        @click="$emit('page-change', 1)"
      >
        <ChevronsLeft class="h-4 w-4" />
      </Button>

      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === 1"
        title="上一页"
        @click="$emit('page-change', currentPage - 1)"
      >
        <ChevronLeft class="h-4 w-4" />
      </Button>

      <div v-if="!compact" class="flex items-center gap-1">
        <Button
          v-for="page in visiblePages"
          :key="page"
          :variant="page === currentPage ? 'default' : 'ghost'"
          size="sm"
          class="min-w-[2.5rem]"
          :disabled="page === '...'"
          @click="page !== '...' && $emit('page-change', page)"
        >
          {{ page }}
        </Button>
      </div>

      <div
        v-else
        class="flex items-center gap-2 rounded-full bg-primary/10 px-2 py-1 text-sm font-medium text-foreground"
      >
        <input
          v-model="pageInput"
          type="text"
          inputmode="numeric"
          class="h-7 w-12 rounded-full bg-background/80 px-2 text-center text-sm text-foreground outline-none ring-0"
          @keydown.enter.prevent="submitPageInput"
          @blur="submitPageInput"
        />
        <span class="text-muted-foreground">/</span>
        <span>{{ totalPages }}</span>
      </div>

      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === totalPages"
        title="下一页"
        @click="$emit('page-change', currentPage + 1)"
      >
        <ChevronRight class="h-4 w-4" />
      </Button>

      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === totalPages"
        title="末页"
        @click="$emit('page-change', totalPages)"
      >
        <ChevronsRight class="h-4 w-4" />
      </Button>
    </div>

    <div class="flex items-center gap-2">
      <span class="text-sm text-muted-foreground whitespace-nowrap">每页</span>
      <select
        :value="itemsPerPage"
        class="rounded-md border bg-background px-2 py-1 text-sm"
        @change="$emit('items-per-page-change', Number($event.target.value))"
      >
        <option :value="25">25</option>
        <option :value="50">50</option>
        <option :value="100">100</option>
        <option :value="200">200</option>
      </select>
      <span class="text-sm text-muted-foreground">项</span>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-vue-next'

const props = defineProps({
  currentPage: { type: Number, required: true },
  totalItems: { type: Number, required: true },
  itemsPerPage: { type: Number, required: true },
  compact: { type: Boolean, default: false },
})

const emit = defineEmits(['page-change', 'items-per-page-change'])

const totalPages = computed(() => Math.ceil(props.totalItems / props.itemsPerPage))
const startIndex = computed(() => (props.currentPage - 1) * props.itemsPerPage)
const endIndex = computed(() => Math.min(startIndex.value + props.itemsPerPage, props.totalItems))
const pageInput = ref(String(props.currentPage))

watch(
  () => props.currentPage,
  (value) => {
    pageInput.value = String(value)
  },
  { immediate: true },
)

const submitPageInput = () => {
  const parsed = Number.parseInt(String(pageInput.value || '').replace(/[^\d]/g, ''), 10)
  const nextPage = Number.isFinite(parsed) ? Math.min(Math.max(parsed, 1), Math.max(totalPages.value, 1)) : props.currentPage
  pageInput.value = String(nextPage)
  if (nextPage !== props.currentPage) {
    emit('page-change', nextPage)
  }
}

const visiblePages = computed(() => {
  const pages = []
  const total = totalPages.value
  const current = props.currentPage

  if (total <= 7) {
    for (let i = 1; i <= total; i += 1) {
      pages.push(i)
    }
  } else {
    pages.push(1)

    if (current > 3) {
      pages.push('...')
    }

    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)

    for (let i = start; i <= end; i += 1) {
      pages.push(i)
    }

    if (current < total - 2) {
      pages.push('...')
    }

    pages.push(total)
  }

  return pages
})
</script>
