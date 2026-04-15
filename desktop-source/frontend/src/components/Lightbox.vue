<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import {
  X,
  ChevronLeft,
  ChevronRight,
  Download,
  Heart,
  Tags,
  Trash2,
  RotateCcw,
  FileImage,
  FolderOpen,
  Copy,
  FileJson,
  Info,
  Loader2,
  Layers,
  ChevronUp,
  ChevronDown,
  StickyNote,
  Bookmark,
} from 'lucide-vue-next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import FavoriteGroupsDialog from './FavoriteGroupsDialog.vue'
import PromptTemplateDialog from './PromptTemplateDialog.vue'

const props = defineProps({
  image: Object,
  isOpen: Boolean,
  images: { type: Array, default: () => [] },
  currentIndex: Number,
  favoriteGroups: { type: Array, default: () => [] },
  tags: { type: Array, default: () => [] },
  imageTags: { type: Object, default: () => ({}) },
  imageNotes: { type: Object, default: () => ({}) },
  openTagsOnMount: { type: Boolean, default: false },
})

const emit = defineEmits(['close', 'navigate', 'toggle-favorite', 'add-tag', 'remove-tag', 'delete', 'open-location', 'favorite-groups-changed'])

const stackViewOpen = ref(false)
const currentStackImage = ref(null)

const currentDisplayImage = computed(() => currentStackImage.value || props.image)

const handleStackImageClick = (img) => {
  currentStackImage.value = img
}

watch(() => props.image, () => {
  currentStackImage.value = null
})

const scale = ref(1)
const offset = ref({ x: 0, y: 0 })
const isDragging = ref(false)
const lastMousePos = ref({ x: 0, y: 0 })
const tagsPopoverOpen = ref(false)
const metadata = ref(null)
const metadataLoading = ref(false)
const metadataError = ref('')
const favoriteGroupsDialogOpen = ref(false)
let metadataRequestId = 0

const promptTemplateDialogOpen = ref(false)
const promptTemplateInitialContent = ref('')
const promptTemplateInitialType = ref('')

const noteText = ref('')
const noteSaving = ref(false)
const noteExpanded = ref(true)

const currentNote = computed(() => {
  if (!currentDisplayImage.value) return ''
  return props.imageNotes[currentDisplayImage.value.relPath] || ''
})

const hasNote = computed(() => currentNote.value.trim() !== '')

watch(() => [props.isOpen, currentDisplayImage.value?.relPath], () => {
  noteText.value = currentNote.value
})

const saveNote = async () => {
  if (!currentDisplayImage.value) return
  const text = noteText.value
  noteSaving.value = true
  try {
    await App.SetImageNote(currentDisplayImage.value.relPath, text)
    if (text.trim()) {
      props.imageNotes[currentDisplayImage.value.relPath] = text
    } else {
      delete props.imageNotes[currentDisplayImage.value.relPath]
    }
  } catch (e) {
    console.error('Failed to save note:', e)
  } finally {
    noteSaving.value = false
  }
}

const handleNoteBlur = () => {
  if (noteText.value !== currentNote.value) {
    saveNote()
  }
}

const handleNoteKeydown = (e) => {
  if (e.ctrlKey && (e.key === 'Enter' || e.key === 's')) {
    e.preventDefault()
    handleNoteBlur()
  }
}

const totalImages = computed(() => props.images?.length || 0)
const canGoPrev = computed(() => props.currentIndex > 0)
const canGoNext = computed(() => props.currentIndex < totalImages.value - 1)
const imageCounter = computed(() => `${Math.min((props.currentIndex || 0) + 1, Math.max(totalImages.value, 1))} / ${Math.max(totalImages.value, 1)}`)

const imageTags = computed(() => {
  if (!currentDisplayImage.value) return []
  const tagIds = props.imageTags[currentDisplayImage.value.relPath] || []
  return tagIds.map(id => props.tags.find(tag => tag.id === id)).filter(Boolean)
})

const availableTags = computed(() => {
  const assigned = imageTags.value.map(tag => tag.id)
  return props.tags.filter(tag => !assigned.includes(tag.id))
})

const metadataFacts = computed(() => {
  if (!metadata.value) return []

  const facts = [
    { label: '模型', value: metadata.value.model },
    { label: '采样器', value: metadata.value.sampler },
    { label: '调度器', value: metadata.value.scheduler },
    { label: 'Seed', value: metadata.value.seed },
    { label: 'Steps', value: metadata.value.steps },
    { label: 'CFG', value: metadata.value.cfg },
  ]

  if (metadata.value.width && metadata.value.height) {
    facts.push({ label: '尺寸', value: `${metadata.value.width} 脳 ${metadata.value.height}` })
  }

  if (metadata.value.nodeCount) {
    facts.push({ label: 'Workflow 节点', value: `${metadata.value.nodeCount}` })
  }

  return facts.filter(item => item.value)
})

const extraMetadataEntries = computed(() => Object.entries(metadata.value?.extraFields || {}))

const resetZoom = () => {
  scale.value = 1
  offset.value = { x: 0, y: 0 }
  isDragging.value = false
}

const loadImageMetadata = async () => {
  const relPath = currentDisplayImage.value?.relPath
  if (!props.isOpen || !relPath) {
    metadata.value = null
    metadataLoading.value = false
    metadataError.value = ''
    return
  }

  const requestId = ++metadataRequestId
  metadataLoading.value = true
  metadataError.value = ''

  try {
    const result = await App.GetImageMetadata(relPath)
    if (requestId !== metadataRequestId) return
    metadata.value = result
  } catch (error) {
    if (requestId !== metadataRequestId) return
    metadata.value = null
    metadataError.value = error?.message || `${error}`
  } finally {
    if (requestId === metadataRequestId) {
      metadataLoading.value = false
    }
  }
}

const copyMetadataField = async (value, label) => {
  if (!value) {
    toast.error(`${label}为空`)
    return
  }

  try {
    await App.CopyText(value)
    toast.success(`${label}已复制`)
  } catch (error) {
    toast.error(`复制失败：${error?.message || error}`)
  }
}

const goToPrev = () => {
  if (!canGoPrev.value) return
  resetZoom()
  emit('navigate', 'prev')
}

const goToNext = () => {
  if (!canGoNext.value) return
  resetZoom()
  emit('navigate', 'next')
}

const goToPrevStackItem = () => {
  if (!props.image.isStackPrimary || props.image.stackCount <= 1) return
  const stackItems = [props.image, ...props.image.stackChildren]
  const currentIndex = stackItems.findIndex(i => i.relPath === currentDisplayImage.value.relPath)
  if (currentIndex > 0) {
    currentStackImage.value = stackItems[currentIndex - 1]
  } else {
    // Wrap around
    currentStackImage.value = stackItems[stackItems.length - 1]
  }
}

const goToNextStackItem = () => {
  if (!props.image.isStackPrimary || props.image.stackCount <= 1) return
  const stackItems = [props.image, ...props.image.stackChildren]
  const currentIndex = stackItems.findIndex(i => i.relPath === currentDisplayImage.value.relPath)
  if (currentIndex < stackItems.length - 1) {
    currentStackImage.value = stackItems[currentIndex + 1]
  } else {
    // Wrap around
    currentStackImage.value = stackItems[0]
  }
}

const getStackCurrentIndex = () => {
  if (!props.image?.isStackPrimary) return 1
  const stackItems = [props.image, ...(props.image.stackChildren || [])]
  const idx = stackItems.findIndex(i => i.relPath === currentDisplayImage.value?.relPath)
  return idx >= 0 ? idx + 1 : 1
}

const handleWheel = event => {
  if (!props.isOpen) return
  const delta = -event.deltaY
  const zoomFactor = 1.1
  const nextScale = delta > 0 ? scale.value * zoomFactor : scale.value / zoomFactor
  scale.value = Math.min(Math.max(nextScale, 0.5), 10)
}

const handleMouseDown = event => {
  if (scale.value <= 1) return
  isDragging.value = true
  lastMousePos.value = { x: event.clientX, y: event.clientY }
  event.preventDefault()
}

const handleMouseMove = event => {
  if (!isDragging.value) return
  const dx = event.clientX - lastMousePos.value.x
  const dy = event.clientY - lastMousePos.value.y
  offset.value = {
    x: offset.value.x + dx,
    y: offset.value.y + dy,
  }
  lastMousePos.value = { x: event.clientX, y: event.clientY }
}

const handleMouseUp = () => {
  isDragging.value = false
}

const handleKey = event => {
  if (!props.isOpen) return
  if (event.key === 'Escape') emit('close')
  if (event.key === 'ArrowLeft') {
    goToPrev()
  }
  if (event.key === 'ArrowRight') {
    goToNext()
  }
  if (event.key === 'ArrowUp') {
    if (props.image?.isStackPrimary && props.image.stackCount > 1) {
      goToPrevStackItem()
    }
  }
  if (event.key === 'ArrowDown') {
    if (props.image?.isStackPrimary && props.image.stackCount > 1) {
      goToNextStackItem()
    }
  }
  if (event.key === '0' && event.ctrlKey) resetZoom()
}

watch(
  () => props.isOpen,
  isOpen => {
    if (isOpen && props.openTagsOnMount) {
      window.setTimeout(() => {
        tagsPopoverOpen.value = true
      }, 100)
    } else if (!isOpen) {
      tagsPopoverOpen.value = false
    }
  }
)

watch(
  () => [props.isOpen, currentDisplayImage.value?.relPath],
  () => {
    void loadImageMetadata()
  },
  { immediate: true }
)

onMounted(() => {
  window.addEventListener('keydown', handleKey)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKey)
})
</script>

<template>
  <transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="isOpen && image"
      class="fixed inset-0 z-[50] flex items-center justify-center overflow-hidden bg-black"
      @click.self="$emit('close')"
    >
      <div class="absolute inset-0 z-0 overflow-hidden pointer-events-none">
        <transition
          enter-active-class="transition-opacity duration-700 ease-in-out"
          leave-active-class="transition-opacity duration-700 ease-in-out"
          enter-from-class="opacity-0"
          leave-to-class="opacity-0"
        >
          <img
            :key="currentDisplayImage.path"
            :src="currentDisplayImage.path"
            class="h-full w-full scale-110 object-cover opacity-40 blur-[60px]"
            loading="eager"
            decoding="async"
          />
        </transition>
        <div class="absolute inset-0 bg-black/40 backdrop-blur-[2px]" />
      </div>

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
                    @click="$emit('remove-tag', currentDisplayImage, tag.id)"
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
                  @click="$emit('add-tag', currentDisplayImage, tag.id)"
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
          @click="$emit('open-location', currentDisplayImage)"
        >
          <FileImage class="h-6 w-6" />
        </button>

        <button
          class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-pink-400"
          title="收藏分组"
          @click="favoriteGroupsDialogOpen = true"
        >
          <FolderOpen class="h-6 w-6" />
        </button>

        <button
          class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-red-500"
          title="删除图片"
          @click="$emit('delete', currentDisplayImage)"
        >
          <Trash2 class="h-6 w-6" />
        </button>

        <button
          class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-white"
          title="收藏"
          @click="$emit('toggle-favorite', currentDisplayImage)"
        >
          <Heart class="h-6 w-6" :class="{ 'fill-red-500 text-red-500': currentDisplayImage.isFavorite }" />
        </button>

        <a
          :href="currentDisplayImage.path"
          download
          class="rounded-full p-2 text-white/70 transition-opacity hover:bg-white/10 hover:text-white"
          title="涓嬭浇"
        >
          <Download class="h-6 w-6" />
        </a>

        <button
          class="rounded-full p-2 text-white/50 transition-all hover:bg-white/10 hover:text-white"
          title="关闭"
          @click="$emit('close')"
        >
          <X class="h-6 w-6" />
        </button>
      </div>

      <div class="absolute inset-y-0 left-0 right-[408px]">
        <button
          v-if="canGoPrev"
          class="absolute left-8 top-1/2 z-[70] flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-all hover:scale-110 hover:bg-white/20"
          @click="goToPrev"
        >
          <ChevronLeft class="h-8 w-8" />
        </button>

        <button
          v-if="canGoNext"
          class="absolute right-8 top-1/2 z-[70] flex h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full bg-white/10 text-white transition-all hover:scale-110 hover:bg-white/20"
          @click="goToNext"
        >
          <ChevronRight class="h-8 w-8" />
        </button>

        <div
          class="relative flex h-full w-full flex-col items-center justify-center overflow-hidden px-24"
          :class="{ 'cursor-grab': scale > 1 && !isDragging, 'cursor-grabbing': isDragging }"
          @wheel="handleWheel"
          @mousedown="handleMouseDown"
          @mousemove="handleMouseMove"
          @mouseup="handleMouseUp"
          @mouseleave="handleMouseUp"
        >
          <div
            class="relative select-none transition-transform duration-75 ease-out"
            :style="{ transform: `translate(${offset.x}px, ${offset.y}px) scale(${scale})` }"
          >
            <img
              :src="currentDisplayImage.path"
              :alt="currentDisplayImage.name"
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
          <span class="truncate text-base font-semibold tracking-wide text-white/90">{{ currentDisplayImage.name }}</span>
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
            @click="resetZoom"
          >
            <RotateCcw class="h-3 w-3" />
            閲嶇疆瑙嗗浘
          </button>
        </div>
      </div>

        <!-- Stack Navigator (Replaces thumbnails) -->
        <div v-if="image.isStackPrimary && image.stackCount > 1" class="absolute bottom-10 left-[50%] -translate-x-1/2 z-[60] flex items-center gap-4 rounded-xl border border-white/10 bg-black/70 px-4 py-2 shadow-2xl backdrop-blur-xl" @click.stop @wheel.stop>
          <button 
             class="flex h-8 w-8 items-center justify-center rounded-full hover:bg-white/20 transition-all text-white/70 hover:text-white"
             @click="goToPrevStackItem"
          >
             <ChevronLeft class="h-5 w-5" />
          </button>
          
          <div class="flex flex-col items-center">
             <span class="text-[10px] font-semibold tracking-widest text-white/50 uppercase mb-0.5">连拍组</span>
             <span class="text-xs font-mono font-medium text-white/90">
               {{ getStackCurrentIndex() }} / {{ image.stackCount }}
             </span>
          </div>

          <button 
             class="flex h-8 w-8 items-center justify-center rounded-full hover:bg-white/20 transition-all text-white/70 hover:text-white"
             @click="goToNextStackItem"
          >
             <ChevronRight class="h-5 w-5" />
          </button>
        </div>

        <div
          class="absolute bottom-10 top-[90px] right-8 z-[60] flex w-[360px] flex-col overflow-hidden rounded-xl border border-white/10 bg-black/70 text-white shadow-2xl backdrop-blur-xl"
          @click.stop
          @wheel.stop
        >
          <div class="flex items-center gap-3 border-b border-white/10 px-4 py-3">
            <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-white/10">
              <Info class="h-4 w-4 text-blue-200" />
            </div>
            <div class="min-w-0">
              <div class="text-sm font-semibold tracking-wide">PNG 元数据</div>
              <div class="text-[11px] text-white/50">
                {{ metadata?.hasMetadata ? '可查看并复制 prompt / workflow' : '按需读取当前图片信息' }}
              </div>
            </div>
          </div>

          <ScrollArea class="min-h-0 flex-1">
            <div class="space-y-3 p-4">
              <!-- Image Note -->
              <div class="rounded-lg border border-white/10 bg-white/5 overflow-hidden">
                <button
                  class="w-full flex items-center justify-between px-3 py-2 hover:bg-white/5 transition-colors"
                  @click="noteExpanded = !noteExpanded"
                >
                  <div class="flex items-center gap-2">
                    <StickyNote class="h-3.5 w-3.5" :class="hasNote ? 'text-amber-300' : 'text-white/45'" />
                    <span class="text-[11px] font-semibold uppercase tracking-wider" :class="hasNote ? 'text-amber-300' : 'text-white/45'">笔记</span>
                  </div>
                  <ChevronUp v-if="noteExpanded" class="h-3.5 w-3.5 text-white/45" />
                  <ChevronDown v-else class="h-3.5 w-3.5 text-white/45" />
                </button>
                <div v-if="noteExpanded" class="px-3 pb-3">
                  <textarea
                    v-model="noteText"
                    class="w-full rounded-md border border-white/10 bg-black/30 px-3 py-2 text-sm text-white/90 placeholder:text-white/30 focus:outline-none focus:ring-1 focus:ring-white/20 resize-none"
                    rows="3"
                    placeholder="添加笔记... (Ctrl+S / Ctrl+Enter 保存)"
                    @blur="handleNoteBlur"
                    @keydown="handleNoteKeydown"
                  ></textarea>
                  <div v-if="noteSaving" class="text-[10px] text-white/40 mt-1">保存中...</div>
                  <div v-else-if="hasNote" class="text-[10px] text-white/40 mt-1">已保存</div>
                </div>
              </div>

              <div v-if="metadataLoading" class="flex items-center gap-2 text-sm text-white/70">
                <Loader2 class="h-4 w-4 animate-spin" />
                <span>正在读取图片元数据...</span>
              </div>

              <div v-else-if="metadataError" class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-200">
                {{ metadataError }}
              </div>

              <div v-else class="space-y-3">
                <div v-if="metadataFacts.length > 0" class="grid grid-cols-2 gap-2">
                  <div
                    v-for="fact in metadataFacts"
                    :key="fact.label"
                    class="rounded-lg border border-white/10 bg-white/5 px-3 py-2"
                  >
                    <div class="text-[11px] uppercase tracking-wider text-white/45">{{ fact.label }}</div>
                    <div class="mt-1 truncate text-sm text-white/90" :title="fact.value">{{ fact.value }}</div>
                  </div>
                </div>

                  <div v-if="metadata?.loras?.length" class="space-y-2">
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">LoRA</div>
                  <div class="flex flex-wrap gap-2">
                    <Badge
                      v-for="lora in metadata.loras"
                      :key="lora"
                      variant="secondary"
                      class="max-w-full border-white/10 bg-white/10 text-white/85"
                    >
                      <span class="truncate">{{ lora }}</span>
                    </Badge>
                  </div>
                </div>

                <div v-if="metadata?.hasMetadata || extraMetadataEntries.length > 0" class="space-y-4 rounded-lg border border-white/10 bg-white/5 p-3">
                  <div v-if="metadata?.positive" class="space-y-2">
                    <div class="flex items-center justify-between gap-3">
                      <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">正向 Prompt</div>
                      <div class="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                          @click="promptTemplateInitialContent = metadata.positive; promptTemplateInitialType = 'positive'; promptTemplateDialogOpen = true"
                          title="存为模板"
                        >
                          <Bookmark class="h-3.5 w-3.5" />
                          存为模板
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                          @click="copyMetadataField(metadata.positive, '正向 Prompt')"
                        >
                          <Copy class="h-3.5 w-3.5" />
                          复制
                        </Button>
                      </div>
                    </div>
                    <ScrollArea class="h-40 rounded-md border border-white/10 bg-black/20" @wheel.stop>
                      <div class="whitespace-pre-wrap break-words p-3 text-sm leading-6 text-white/88">
                        {{ metadata.positive }}
                      </div>
                    </ScrollArea>
                  </div>

                  <div v-if="metadata?.negative" class="space-y-2">
                    <div class="flex items-center justify-between gap-3">
                      <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">反向 Prompt</div>
                      <div class="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                          @click="promptTemplateInitialContent = metadata.negative; promptTemplateInitialType = 'negative'; promptTemplateDialogOpen = true"
                          title="存为模板"
                        >
                          <Bookmark class="h-3.5 w-3.5" />
                          存为模板
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                          @click="copyMetadataField(metadata.negative, '反向 Prompt')"
                        >
                          <Copy class="h-3.5 w-3.5" />
                          复制
                        </Button>
                      </div>
                    </div>
                    <ScrollArea class="h-40 rounded-md border border-white/10 bg-black/20" @wheel.stop>
                      <div class="whitespace-pre-wrap break-words p-3 text-sm leading-6 text-white/88">
                        {{ metadata.negative }}
                      </div>
                    </ScrollArea>
                  </div>

                  <div v-if="metadata?.prompt" class="space-y-2">
                    <div class="flex items-center justify-between gap-3">
                      <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">ComfyUI 提示词</div>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                        @click="copyMetadataField(metadata.prompt, 'Prompt JSON')"
                      >
                        <FileJson class="h-3.5 w-3.5" />
                        复制 JSON
                      </Button>
                    </div>
                    <div class="text-xs leading-5 text-white/60">
                      已检测到 ComfyUI prompt 节点图，可直接复制 JSON 用于恢复或分析。
                    </div>
                    <ScrollArea
                      v-if="!metadata?.positive && !metadata?.negative"
                      class="h-32 rounded-md border border-white/10 bg-black/20"
                    >
                      <div class="whitespace-pre-wrap break-words p-3 font-mono text-xs leading-5 text-white/75">
                        {{ metadata.prompt }}
                      </div>
                    </ScrollArea>
                  </div>

                  <div v-if="metadata?.workflow" class="space-y-2">
                    <div class="flex items-center justify-between gap-3">
                      <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">工作流</div>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                        @click="copyMetadataField(metadata.workflow, 'Workflow JSON')"
                      >
                        <FileJson class="h-3.5 w-3.5" />
                        复制 JSON
                      </Button>
                    </div>
                    <div class="text-xs leading-5 text-white/60">
                      {{ metadata.nodeCount ? `已检测到 ${metadata.nodeCount} 个 workflow 节点，可直接复制回 ComfyUI。` : '已检测到 workflow JSON，可直接复制回 ComfyUI。' }}
                    </div>
                  </div>

                  <div
                    v-if="!metadata?.positive && !metadata?.negative && metadata?.extraFields?.parameters"
                    class="space-y-2"
                  >
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">原始参数</div>
                    <ScrollArea class="h-32 rounded-md border border-white/10 bg-black/20">
                      <div class="whitespace-pre-wrap break-words p-3 text-xs leading-5 text-white/80">
                        {{ metadata.extraFields.parameters }}
                      </div>
                    </ScrollArea>
                  </div>

                  <div v-if="extraMetadataEntries.length > 0" class="space-y-2">
                    <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">其他字段</div>
                    <div class="space-y-2">
                      <div
                        v-for="[key, value] in extraMetadataEntries"
                        :key="key"
                        class="rounded-lg border border-white/10 bg-black/20 px-3 py-2"
                      >
                        <div class="flex items-center justify-between gap-3">
                          <div class="text-[11px] text-white/55">{{ key }}</div>
                          <Button
                            variant="ghost"
                            size="sm"
                            class="h-6 gap-1 px-2 text-white/65 hover:bg-white/10 hover:text-white"
                            @click="copyMetadataField(value, key)"
                          >
                            <Copy class="h-3 w-3" />
                            复制
                          </Button>
                        </div>
                        <div class="mt-2 max-h-16 overflow-hidden whitespace-pre-wrap break-words text-xs leading-5 text-white/70">
                          {{ value }}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div v-else class="rounded-lg border border-white/10 bg-white/5 px-3 py-4 text-sm leading-6 text-white/60">
                  当前图片没有检测到可读取的 ComfyUI PNG 元数据。
                </div>
              </div>
            </div>
          </ScrollArea>
        </div>

      <FavoriteGroupsDialog
        v-model:open="favoriteGroupsDialogOpen"
        :groups="favoriteGroups"
        :image="currentDisplayImage"
        @change="$emit('favorite-groups-changed')"
      />

      <PromptTemplateDialog
        v-model:open="promptTemplateDialogOpen"
        :initial-content="promptTemplateInitialContent"
        :initial-type="promptTemplateInitialType"
        :initial-source-path="currentDisplayImage?.relPath || ''"
      />
    </div>
  </transition>
</template>


