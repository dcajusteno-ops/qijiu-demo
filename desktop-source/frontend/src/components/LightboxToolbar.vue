<script setup>
import { computed, ref, watch } from 'vue'
import { Download, FileImage, FolderOpen, Heart, Sparkles, Tags, Trash2, X } from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

const props = defineProps({
  isOpen: { type: Boolean, default: false },
  openTagsOnMount: { type: Boolean, default: false },
  currentDisplayImage: { type: Object, default: null },
  imageTags: { type: Array, default: () => [] },
  availableTags: { type: Array, default: () => [] },
  metadata: { type: Object, default: null },
})

const emit = defineEmits([
  'close',
  'remove-tag',
  'add-tag',
  'open-location',
  'open-favorite-groups',
  'delete',
  'toggle-favorite',
  'open-prompt-assistant',
])

const tagsPopoverOpen = ref(false)

watch(
  () => props.isOpen,
  (isOpen) => {
    if (isOpen && props.openTagsOnMount) {
      window.setTimeout(() => {
        tagsPopoverOpen.value = true
      }, 100)
    } else if (!isOpen) {
      tagsPopoverOpen.value = false
    }
  },
)

const promptAssistantPayload = computed(() => ({
  initialPositive: props.metadata?.positive || '',
  initialNegative: props.metadata?.negative || '',
  sourcePath: props.currentDisplayImage?.relPath || '',
  contextLabel: props.currentDisplayImage?.name || '',
}))
</script>

<template>
  <div class="absolute right-8 top-6 z-[60] flex items-center gap-3">
    <Popover v-model:open="tagsPopoverOpen">
      <PopoverTrigger as-child>
        <button
          class="relative rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-white"
          title="标签管理"
        >
          <Tags class="h-6 w-6" />
          <span
            v-if="imageTags.length > 0"
            class="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-blue-500 text-[10px]"
          >
            {{ imageTags.length }}
          </span>
        </button>
      </PopoverTrigger>
      <PopoverContent class="w-72 p-3" align="end">
        <div class="space-y-3">
          <div v-if="imageTags.length > 0">
            <div class="mb-2 text-xs font-medium text-muted-foreground">当前标签</div>
            <div class="flex flex-wrap gap-2">
              <Badge
                v-for="tag in imageTags"
                :key="tag.id"
                :style="{ backgroundColor: tag.color }"
                class="flex max-w-[200px] items-center gap-1 py-1 pl-2 pr-1 text-white"
              >
                <span class="truncate">{{ tag.name }}</span>
                <Button
                  variant="ghost"
                  size="icon"
                  class="ml-1 h-4 w-4 rounded-full p-0 hover:bg-white/20"
                  @click="emit('remove-tag', currentDisplayImage, tag.id)"
                >
                  <X class="h-3 w-3" />
                </Button>
              </Badge>
            </div>
          </div>

          <div v-if="availableTags.length > 0">
            <div class="mb-2 text-xs font-medium text-muted-foreground">添加标签</div>
            <div class="flex flex-wrap gap-2">
              <Badge
                v-for="tag in availableTags"
                :key="tag.id"
                :style="{ backgroundColor: tag.color }"
                class="cursor-pointer text-white transition-transform hover:scale-105"
                @click="emit('add-tag', currentDisplayImage, tag.id)"
              >
                <span class="block truncate">{{ tag.name }}</span>
              </Badge>
            </div>
          </div>

          <div v-if="imageTags.length === 0 && availableTags.length === 0" class="py-2 text-center text-sm text-muted-foreground">
            暂无可用标签，请先在侧边栏创建标签
          </div>
        </div>
      </PopoverContent>
    </Popover>

    <button
      class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-amber-500"
      title="打开文件位置"
      @click="emit('open-location', currentDisplayImage)"
    >
      <FileImage class="h-6 w-6" />
    </button>

    <button
      class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-pink-400"
      title="收藏分组"
      @click="emit('open-favorite-groups')"
    >
      <FolderOpen class="h-6 w-6" />
    </button>

    <button
      class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-red-500"
      title="删除图片"
      @click="emit('delete', currentDisplayImage)"
    >
      <Trash2 class="h-6 w-6" />
    </button>

    <button
      class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-white"
      title="收藏"
      @click="emit('toggle-favorite', currentDisplayImage)"
    >
      <Heart class="h-6 w-6" :class="{ 'fill-red-500 text-red-500': currentDisplayImage?.isFavorite }" />
    </button>

    <a
      :href="currentDisplayImage?.path"
      download
      class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-white"
      title="下载"
    >
      <Download class="h-6 w-6" />
    </a>

    <button
      class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-emerald-300"
      title="提示词提示器"
      @click="emit('open-prompt-assistant', promptAssistantPayload)"
    >
      <Sparkles class="h-6 w-6" />
    </button>

    <button
      class="rounded-full p-2 text-white/50 transition-all hover:bg-white/10 hover:text-white"
      title="关闭"
      @click="emit('close')"
    >
      <X class="h-6 w-6" />
    </button>
  </div>
</template>
