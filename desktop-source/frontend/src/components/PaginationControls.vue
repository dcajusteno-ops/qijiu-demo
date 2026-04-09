<template>
  <div class="flex items-center justify-between gap-4 px-4 py-3 bg-muted/30 rounded-lg border select-none">
    <!-- Left: Info -->
    <div class="text-sm text-muted-foreground">
      显示 <span class="font-medium text-foreground">{{ startIndex + 1 }}-{{ endIndex }}</span> 共 <span class="font-medium text-foreground">{{ totalItems }}</span> 项
    </div>

    <!-- Center: Page Navigation -->
    <div class="flex items-center gap-2">
      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === 1"
        @click="$emit('page-change', 1)"
        title="首页"
      >
        <ChevronsLeft class="h-4 w-4" />
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === 1"
        @click="$emit('page-change', currentPage - 1)"
        title="上一页"
      >
        <ChevronLeft class="h-4 w-4" />
      </Button>

      <!-- Page Numbers -->
      <div class="flex items-center gap-1">
        <Button
          v-for="page in visiblePages"
          :key="page"
          :variant="page === currentPage ? 'default' : 'ghost'"
          size="sm"
          class="min-w-[2.5rem]"
          @click="page !== '...' && $emit('page-change', page)"
          :disabled="page === '...'"
        >
          {{ page }}
        </Button>
      </div>

      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === totalPages"
        @click="$emit('page-change', currentPage + 1)"
        title="下一页"
      >
        <ChevronRight class="h-4 w-4" />
      </Button>

      <Button
        variant="outline"
        size="sm"
        :disabled="currentPage === totalPages"
        @click="$emit('page-change', totalPages)"
        title="末页"
      >
        <ChevronsRight class="h-4 w-4" />
      </Button>
    </div>

    <!-- Right: Items Per Page Selector -->
    <div class="flex items-center gap-2">
      <span class="text-sm text-muted-foreground whitespace-nowrap">每页</span>
      <select
        :value="itemsPerPage"
        @change="$emit('items-per-page-change', Number($event.target.value))"
        class="px-2 py-1 text-sm border rounded-md bg-background"
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
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-vue-next'

const props = defineProps({
  currentPage: { type: Number, required: true },
  totalItems: { type: Number, required: true },
  itemsPerPage: { type: Number, required: true }
})

defineEmits(['page-change', 'items-per-page-change'])

const totalPages = computed(() => Math.ceil(props.totalItems / props.itemsPerPage))
const startIndex = computed(() => (props.currentPage - 1) * props.itemsPerPage)
const endIndex = computed(() => Math.min(startIndex.value + props.itemsPerPage, props.totalItems))

// Calculate visible page numbers with ellipsis
const visiblePages = computed(() => {
  const pages = []
  const total = totalPages.value
  const current = props.currentPage

  if (total <= 7) {
    // Show all pages if 7 or fewer
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    // Always show first page
    pages.push(1)

    if (current > 3) {
      pages.push('...')
    }

    // Show pages around current
    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)

    for (let i = start; i <= end; i++) {
      pages.push(i)
    }

    if (current < total - 2) {
      pages.push('...')
    }

    // Always show last page
    pages.push(total)
  }

  return pages
})
</script>
