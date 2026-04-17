<script setup>
import { ref, computed } from 'vue'
import ImageCard from './ImageCard.vue'
import Lightbox from './Lightbox.vue'
import ExportDialog from './ExportDialog.vue'
import FilterPanel from './FilterPanel.vue'
import SortDropdown from './SortDropdown.vue'
import PromptToolsDropdown from './PromptToolsDropdown.vue'
import BatchActionsPanel from './BatchActionsPanel.vue'
import MoveToFolderDialog from './MoveToFolderDialog.vue'
import PaginationControls from './PaginationControls.vue'
import ABCompareSlider from './ABCompareSlider.vue'
import FavoriteGroupsDialog from './FavoriteGroupsDialog.vue'

// import StatisticsPanel from './StatisticsPanel.vue' - replaced by dashboard
import StatisticsDashboard from './StatisticsDashboard.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Slider } from '@/components/ui/slider'
import { Heart, Grid, Download, BarChart3, Upload, Layers, Search, Sparkles, X } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import { useImages } from '@/composables/useImages'

const { isStackingEnabled, setSortBy, setSortOrder } = useImages()

const toggleStacking = () => {
  isStackingEnabled.value = !isStackingEnabled.value
  localStorage.setItem('isStackingEnabled', isStackingEnabled.value ? 'true' : 'false')
}

const props = defineProps({
  images: { type: Array, default: () => [] },
  totalImages: { type: Number, default: 0 },
  loading: { type: Boolean, default: false },
  rootName: { type: String, default: '' },
  subName: { type: String, default: '' },
  childName: { type: String, default: '' },
  rootLabel: { type: String, default: '' },
  subLabel: { type: String, default: '' },
  childLabel: { type: String, default: '' },
  targetFolderPath: { type: String, default: '' },
  isSelectionMode: { type: Boolean, default: false },
  selectedPaths: { type: Set, default: () => new Set() },
  scopeImageCount: { type: Number, default: 0 },
  tags: { type: Array, default: () => [] },
  imageTags: { type: Object, default: () => ({}) },
  favoriteGroups: { type: Array, default: () => [] },
  imageNotes: { type: Object, default: () => ({}) },
  searchQuery: { type: String, default: '' },
  availableModels: { type: Array, default: () => [] },
  availableLoras: { type: Array, default: () => [] },
  activeDateLabel: { type: String, default: '全部时间' },
  activeModelFilter: { type: String, default: '' },
  activeLoraFilter: { type: String, default: '' },
  hasActiveWorkbenchFilters: { type: Boolean, default: false },
  currentPage: { type: Number, default: 1 },
  itemsPerPage: { type: Number, default: 50 },
  totalPages: { type: Number, default: 1 }
})

const emit = defineEmits([
  'delete',
  'toggle-selection',
  'select-all',
  'clear-selection',
  'delete-selected',
  'toggle-favorite',
  'add-tag',
  'remove-tag',
  'view-favorites',
  'refresh-images',
  'favorite-groups-changed',
  'page-change',
  'items-per-page-change',
  'open-location',
  'update:search-query',
  'update:model-filter',
  'update:lora-filter',
  'clear-workbench-filters',
  'clear-all-filters',
  'open-prompt-assistant',
])

const lightboxOpen = ref(false)
const currentImage = ref(null)
const currentImageIndex = ref(0) // Local index in the filtered list
const thumbnailSize = ref([300]) // Default size
const exportDialogOpen = ref(false)
const openTagsOnLightboxOpen = ref(false) // Track if tags should auto-open

// Drag & Drop state
const isDragging = ref(false)
const isUploading = ref(false)

// Batch Operations state
const moveDialogOpen = ref(false)
const compareSliderOpen = ref(false)
const compareImageA = ref(null)
const compareImageB = ref(null)
const favoriteGroupsDialogOpen = ref(false)
const favoriteDialogImage = ref(null)

const hasTopFilters = computed(() =>
  !!props.searchQuery || props.hasActiveWorkbenchFilters,
)

const emptyStateMessage = computed(() => {
  if (props.searchQuery || props.hasActiveWorkbenchFilters) {
    if (props.scopeImageCount > 0) {
      return `当前目录原本有 ${props.scopeImageCount} 张图片，但被搜索或工作台筛选条件全部过滤掉了`
    }
    return '没有找到匹配当前搜索条件的图片'
  }
  return '该文件夹为空'
})

const handleCompare = () => {
    const paths = Array.from(props.selectedPaths)
    if (paths.length !== 2) return

    // Find the full image objects from the currently loaded images
    compareImageA.value = props.images.find(img => img.relPath === paths[0]) || { relPath: paths[0], name: paths[0].split(/[/\\]/).pop() }
    compareImageB.value = props.images.find(img => img.relPath === paths[1]) || { relPath: paths[1], name: paths[1].split(/[/\\]/).pop() }

    compareSliderOpen.value = true
}

const openFavoriteDialog = (img) => {
    favoriteDialogImage.value = img
    favoriteGroupsDialogOpen.value = true
}

const handleFavoriteDialogOpenChange = (open) => {
    favoriteGroupsDialogOpen.value = open
    if (!open) {
        favoriteDialogImage.value = null
    }
}

const handleFavoriteGroupsChanged = () => {
    emit('favorite-groups-changed')
}

const handleExport = async ({ targetDir, move }) => {
    const paths = Array.from(props.selectedPaths)
    if (paths.length === 0) return

    try {
        const result = await App.ExportImages(paths, targetDir, move)
        
        if (result.errors && result.errors.length > 0) {
            console.error('Export errors:', result.errors)
            toast.warning(`导出完成: ${result.success} 成功, ${result.failed} 失败`, {
                description: '部分文件导出失败，请检查控制台详情'
            })
        } else {
             toast.success(`成功导出 ${result.success} 张图片`)
             emit('clear-selection')
        }
    } catch (e) {
        console.error(e)
        // Wails passes back the error message as string
        toast.error(`导出失败: ${e}`)
    }
}

// Drag & Drop Handlers
const targetFolder = computed(() => {
    if (props.rootName === 'favorites' || props.rootName === 'statistics') {
        return '根目录'
    }
    if (props.targetFolderPath) {
        return props.targetFolderPath
    }
    return '根目录'
})

let dragCounter = 0

const handleDragEnter = (e) => {
    e.preventDefault()
    e.stopPropagation()
    dragCounter++
    isDragging.value = true
}

const handleDragLeave = (e) => {
    e.preventDefault()
    e.stopPropagation()
    dragCounter--
    if (dragCounter === 0) {
        isDragging.value = false
    }
}

const handleDragOver = (e) => {
    e.preventDefault()
    e.stopPropagation()
}

const handleDrop = async (e) => {
    e.preventDefault()
    e.stopPropagation()
    dragCounter = 0
    isDragging.value = false

    const files = Array.from(e.dataTransfer.files)
    
    const imageFiles = files.filter(file => {
        const ext = file.name.toLowerCase()
        return ext.endsWith('.png') || ext.endsWith('.jpg') || 
               ext.endsWith('.jpeg') || ext.endsWith('.webp') || 
               ext.endsWith('.gif')
    })

    if (imageFiles.length === 0) {
        toast.error('没有有效的图片文件')
        return
    }

    if (imageFiles.length !== files.length) {
        toast.warning(`已忽略 ${files.length - imageFiles.length} 个非图片文件`)
    }

    await uploadFiles(imageFiles)
}

const uploadFiles = async (files) => {
    // With Wails, we shouldn't pass raw browser File objects directly via bindings if they are large,
    // because that means reading them entirely into memory in JS and converting to base64.
    // Instead, since it's a local app, we should provide an API to open a folder dialog or file dialog.
    // However, for drag-and-drop, we can get the actual file path from the dropped File object in Electron/Wails
    // In Wails v2 webview, file.path contains the absolute path on the local system.
    
    const targetFolderPath = props.rootName === 'favorites' || props.rootName === 'statistics'
        ? ''
        : (props.targetFolderPath || '')

    isUploading.value = true
    try {
        const paths = []
        for (const file of files) {
            if (file.path) {
                paths.push(file.path)
            }
        }
        
        if (paths.length === 0) {
           throw new Error("无法获取文件的本地路径，请使用导入按钮")
        }

        const result = await App.UploadImages(paths, targetFolderPath)

        if (result.count > 0) {
            toast.success(`成功导入 ${result.count} 张图片`)
        }

        if (result.errors && result.errors.length > 0) {
            result.errors.forEach(err => toast.error(err))
        }

        emit('refresh-images')

    } catch (error) {
        console.error('Upload error:', error)
        toast.error(`导入失败: ${error.message || error}`)
    } finally {
        isUploading.value = false
    }
}

// Batch Operations Handlers
const handleBatchAddTag = async (tagId) => {
    const paths = Array.from(props.selectedPaths)
    try {
        const count = await App.BatchAddTag(paths, tagId)
        
        toast.success(`已为 ${count} 张图片添加标签`)
        emit('refresh-images')
    } catch (error) {
        console.error('Batch add tag error:', error)
        toast.error(`批量添加标签失败: ${error}`)
    }
}

const handleBatchRemoveTag = async (tagId) => {
    const paths = Array.from(props.selectedPaths)
    try {
        const count = await App.BatchRemoveTag(paths, tagId)
        
        toast.success(`已为 ${count} 张图片移除标签`)
        emit('refresh-images')
    } catch (error) {
        console.error('Batch remove tag error:', error)
        toast.error(`批量移除标签失败: ${error}`)
    }
}

const handleBatchMove = () => {
    moveDialogOpen.value = true
}

const handleMoveConfirm = async (targetFolder) => {
    const paths = Array.from(props.selectedPaths)
    try {
        const count = await App.BatchMove(paths, targetFolder)
        
        toast.success(`成功移动 ${count} 张图片`)
        
        emit('clear-selection')
        emit('refresh-images')
    } catch (error) {
        console.error('Batch move error:', error)
        toast.error(`移动失败: ${error}`)
    } finally {
        moveDialogOpen.value = false
    }
}

const handleBatchFavorite = async () => {
    const paths = Array.from(props.selectedPaths)
    try {
        const count = await App.BatchFavorites(paths, 'add')
        
        toast.success(`已收藏 ${count} 张图片`)
        emit('refresh-images')
    } catch (error) {
        console.error('Batch favorite error:', error)
        toast.error(`批量收藏失败: ${error}`)
    }
}


const openLightbox = (img, shouldOpenTags = false) => {
    if (props.isSelectionMode) return
    currentImageIndex.value = props.images.findIndex(i => i.relPath === img.relPath)
    currentImage.value = img
    openTagsOnLightboxOpen.value = shouldOpenTags
    lightboxOpen.value = true
}

const handleNavigate = (direction) => {
    if (direction === 'prev' && currentImageIndex.value > 0) {
        currentImageIndex.value--
    } else if (direction === 'next' && currentImageIndex.value < props.images.length - 1) {
        currentImageIndex.value++
    }
    currentImage.value = props.images[currentImageIndex.value]
}

// Scroll to top when page changes
const galleryContainer = ref(null)
// Watch for page changes
import { watch } from 'vue'
watch(() => props.currentPage, () => {
    if (galleryContainer.value) {
        galleryContainer.value.scrollTop = 0
    }
})
</script>

<template>
  <div 
    class="h-full flex flex-col bg-background w-full overflow-hidden relative select-none"
    @dragenter="handleDragEnter"
    @dragleave="handleDragLeave"
    @dragover="handleDragOver"
    @drop="handleDrop"
  >
      <!-- Header -->
      <header class="flex-none border-b bg-background/80 px-6 py-4 backdrop-blur-md z-10 select-none">
          <div class="flex flex-col gap-4">
            <div class="min-w-0">
              <div class="flex items-start gap-4">
                  <div v-if="rootName === 'statistics'" class="flex flex-col justify-center">
                     <h2 class="text-xl font-semibold tracking-tight">数据视界</h2>
                     <p class="text-xs text-muted-foreground">生成历史时间线与趋势分析</p>
                  </div>
                  <div v-else class="min-w-0 flex-1">
                     <h2 class="break-words text-[2rem] font-semibold leading-tight tracking-tight">
                        {{ rootLabel || rootName || '图片库' }}
                        <span v-if="subLabel && subLabel !== '默认'" class="text-muted-foreground font-normal">
                            / {{ subLabel }}
                        </span>
                        <span v-if="childLabel" class="text-muted-foreground font-normal">
                            / {{ childLabel }}
                        </span>
                     </h2>
                     <p class="mt-0.5 text-xs text-muted-foreground">
                       {{ totalImages > 0 ? `当前结果 ${totalImages} 张` : (hasTopFilters ? '当前目录有图片，但被搜索或筛选条件过滤了' : '当前目录暂无图片') }}
                     </p>
                  </div>
              </div>
            </div>
              <div v-if="rootName !== 'statistics'" class="flex flex-col gap-3">
                  <div class="flex flex-wrap items-center gap-3">
                      <div class="relative w-full min-w-[260px] md:w-72">
                          <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                          <Input
                            :model-value="searchQuery"
                            placeholder="搜索文件名、Prompt、模型、LoRA..."
                            class="h-10 rounded-2xl border-border/80 bg-background/90 pl-9 pr-10 shadow-none"
                            @update:model-value="emit('update:search-query', $event)"
                          />
                          <button
                            v-if="searchQuery"
                            class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                            type="button"
                            title="清空搜索"
                            @click="emit('update:search-query', '')"
                          >
                            <X class="h-4 w-4" />
                          </button>
                      </div>
                      <div v-if="isSelectionMode" class="flex items-center gap-2 text-sm font-medium transition-colors" :class="selectedPaths.size > 0 ? 'text-primary bg-primary/10 px-3 py-1 rounded-full' : 'text-muted-foreground bg-muted/50 border border-dashed border-muted-foreground/30 px-3 py-1 rounded-full'">
                          <span>{{ selectedPaths.size === 0 ? '批量模式：请点击选择图片' : `已选 ${selectedPaths.size} 张` }}</span>
                          <template v-if="selectedPaths.size > 0">
                              <Separator orientation="vertical" class="h-4 bg-primary/20" />
                              <button 
                                class="hover:text-foreground transition-colors flex items-center gap-1"
                                title="导出选中图片"
                                @click="exportDialogOpen = true"
                              >
                                  <Download class="h-4 w-4" />
                                  导出
                              </button>
                          </template>
                      </div>
                      <div class="flex items-center gap-1">
                          <Button 
                            variant="ghost" 
                            size="icon" 
                            @click="emit('view-favorites')"
                            :class="rootName === 'favorites' ? 'text-red-500 bg-red-500/10' : 'text-muted-foreground hover:text-red-500'"
                            title="收藏夹"
                          >
                              <Heart class="h-5 w-5" :class="{ 'fill-current': rootName === 'favorites' }" />
                          </Button>

                          <Button
                            variant="ghost"
                            size="icon"
                            @click="toggleStacking"
                            :class="isStackingEnabled ? 'text-primary bg-primary/10' : 'text-muted-foreground'"
                            title="连拍叠图"
                          >
                            <Layers class="h-5 w-5" />
                          </Button>

                          <SortDropdown />

                          <div class="flex items-center gap-2 px-2">
                            <Grid class="h-4 w-4 text-muted-foreground shrink-0" />
                            <Slider
                              v-model="thumbnailSize"
                              :min="120"
                              :max="500"
                              :step="10"
                              class="w-20"
                            />
                          </div>

                          <FilterPanel />

                          <PromptToolsDropdown />

                          <Button
                            variant="ghost"
                            size="icon"
                            title="提示词提示器"
                            @click="emit('open-prompt-assistant', { contextLabel: rootLabel || rootName || '当前图库', sourcePath: targetFolderPath || '' })"
                          >
                            <Sparkles class="h-5 w-5 text-muted-foreground" />
                          </Button>
                      </div>
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="rounded-full border border-border bg-background px-3 py-1.5 text-xs text-muted-foreground">
                      日期：{{ activeDateLabel || '全部时间' }}
                    </span>
                    <select
                      :value="activeModelFilter"
                      class="h-9 min-w-[180px] rounded-full border border-border bg-background px-3 text-sm text-foreground outline-none transition focus:border-primary"
                      @change="emit('update:model-filter', $event.target.value)"
                    >
                      <option value="">全部模型</option>
                      <option v-for="item in availableModels" :key="item.value || item.name" :value="item.value || item.name">
                        {{ item.label || item.name }} ({{ item.count }})
                      </option>
                    </select>
                    <select
                      :value="activeLoraFilter"
                      class="h-9 min-w-[180px] rounded-full border border-border bg-background px-3 text-sm text-foreground outline-none transition focus:border-primary"
                      @change="emit('update:lora-filter', $event.target.value)"
                    >
                      <option value="">全部 LoRA</option>
                      <option v-for="item in availableLoras" :key="item.value || item.name" :value="item.value || item.name">
                        {{ item.label || item.name }} ({{ item.count }})
                      </option>
                    </select>
                    <Button
                      v-if="hasActiveWorkbenchFilters"
                      variant="outline"
                      class="h-9 rounded-full px-4"
                      @click="emit('clear-workbench-filters')"
                    >
                      清空工作台筛选
                    </Button>
                    <Button
                      v-if="hasTopFilters"
                      variant="outline"
                      class="h-9 rounded-full px-4"
                      @click="emit('clear-all-filters')"
                    >
                      清空全部筛选
                    </Button>
                  </div>
              </div>
          </div>
      </header>

      <!-- Content -->
      <StatisticsDashboard v-if="rootName === 'statistics'" class="flex-1" />

      <div v-else-if="!rootName" class="flex-1 flex flex-col items-center justify-center text-muted-foreground p-10">
          <p class="text-lg font-medium opacity-50">请在左侧选择标签以查看图片</p>
      </div>

      <div v-else-if="loading && images.length === 0" class="flex-1 flex items-center justify-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>

      <div v-else-if="images.length > 0" ref="galleryContainer" class="flex-1 p-6 overflow-y-auto custom-scrollbar flex flex-col gap-4">
          <!-- Pagination Controls (Top) -->
          <PaginationControls 
            v-if="totalImages > 0"
            :current-page="currentPage"
            :total-items="totalImages"
            :items-per-page="itemsPerPage"
            @page-change="emit('page-change', $event)"
            @items-per-page-change="emit('items-per-page-change', $event)"
          />

          <div class="grid gap-6" :style="{ gridTemplateColumns: `repeat(auto-fill, minmax(${thumbnailSize[0]}px, 1fr))` }">
              <ImageCard
                v-for="img in images"
                :key="img.relPath"
                :image="img"
                :selectable="isSelectionMode"
                :selected="selectedPaths.has(img.relPath)"
                :has-note="!!imageNotes[img.relPath]"
                @view="openLightbox(img)"
                @delete="emit('delete', img)"
                @toggle="emit('toggle-selection', img)"
                @manage-favorites="openFavoriteDialog(img)"
                @manage-tags="openLightbox(img, true)"
                @open-location="emit('open-location', img)"
              />
          </div>

          <!-- Pagination Controls (Bottom) -->
          <PaginationControls 
            v-if="totalImages > 0"
            :current-page="currentPage"
            :total-items="totalImages"
            :items-per-page="itemsPerPage"
            @page-change="emit('page-change', $event)"
            @items-per-page-change="emit('items-per-page-change', $event)"
          />
      </div>

      <div v-else class="flex-1 flex flex-col items-center justify-center gap-4 px-6 pt-20 text-center text-muted-foreground">
          <p>{{ emptyStateMessage }}</p>
          <Button
            v-if="hasTopFilters"
            variant="outline"
            class="rounded-full px-4"
            @click="emit('clear-all-filters')"
          >
            清空搜索和筛选
          </Button>
      </div>

      <Lightbox
        :image="currentImage"
        :images="images"
        :current-index="currentImageIndex"
        :isOpen="lightboxOpen"
        :favorite-groups="favoriteGroups"
        :tags="tags"
        :image-tags="imageTags"
        :image-notes="imageNotes"
        :open-tags-on-mount="openTagsOnLightboxOpen"
        @close="lightboxOpen = false; openTagsOnLightboxOpen = false"
        @navigate="handleNavigate"
        @toggle-favorite="(img) => emit('toggle-favorite', img)"
        @add-tag="(img, tagId) => emit('add-tag', img, tagId)"
        @remove-tag="(img, tagId) => emit('remove-tag', img, tagId)"
        @delete="(img) => { emit('delete', img); lightboxOpen = false }"
        @open-location="(img) => emit('open-location', img)"
        @favorite-groups-changed="emit('favorite-groups-changed')"
        @open-prompt-assistant="emit('open-prompt-assistant', $event)"
      />
      <FavoriteGroupsDialog
        :open="favoriteGroupsDialogOpen"
        :groups="favoriteGroups"
        :image="favoriteDialogImage"
        @update:open="handleFavoriteDialogOpenChange"
        @change="handleFavoriteGroupsChanged"
      />
      <ExportDialog 
        v-model:open="exportDialogOpen"
        :count="selectedPaths.size"
        @confirm="handleExport"
      />

      <!-- Drag & Drop Overlay -->
      <Transition
        enter-active-class="transition-opacity duration-200"
        leave-active-class="transition-opacity duration-200"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div 
          v-if="isDragging"
          class="absolute inset-0 bg-background/95 backdrop-blur-sm z-50 flex items-center justify-center border-4 border-dashed border-primary/50"
        >
          <div class="text-center">
            <Upload class="w-16 h-16 mx-auto mb-4 text-primary animate-bounce" />
            <p class="text-xl font-semibold text-foreground">拖放图片到这里上传</p>
            <p class="text-sm text-muted-foreground mt-2">
              上传到: {{ targetFolder }}
            </p>
          </div>
        </div>
      </Transition>

      <!-- Upload Progress Overlay -->
      <Transition
        enter-active-class="transition-opacity duration-200"
        leave-active-class="transition-opacity duration-200"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div 
          v-if="isUploading"
          class="absolute inset-0 bg-background/95 backdrop-blur-sm z-50 flex items-center justify-center"
        >
          <div class="text-center">
            <div class="w-16 h-16 mx-auto mb-4 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
            <p class="text-xl font-semibold text-foreground">上传中...</p>
          </div>
        </div>
      </Transition>

      <!-- Batch Actions Panel -->
      <BatchActionsPanel
        :show="isSelectionMode"
        :count="selectedPaths.size"
        :tags="tags"
        @batch-add-tag="handleBatchAddTag"
        @batch-remove-tag="handleBatchRemoveTag"
        @batch-move="handleBatchMove"
        @batch-favorite="handleBatchFavorite"
        @batch-delete="emit('delete-selected')"
        @select-all="emit('select-all')"
        @clear-selection="emit('clear-selection')"
        @compare="handleCompare"
      />

      <!-- Move To Folder Dialog -->
      <MoveToFolderDialog
        v-model:open="moveDialogOpen"
        :count="selectedPaths.size"
        @move="handleMoveConfirm"
      />

      <!-- A/B Compare Slider Modal -->
      <ABCompareSlider
        :is-open="compareSliderOpen"
        :image-a="compareImageA"
        :image-b="compareImageB"
        @close="compareSliderOpen = false"
      />

  </div>
</template>

