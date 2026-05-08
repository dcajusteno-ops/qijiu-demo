<script setup>
import { computed, ref, watch } from 'vue'
import {
  ArrowLeft,
  Check,
  CheckSquare,
  ChevronLeft,
  ChevronRight,
  Copy,
  Eraser,
  ExternalLink,
  FileImage,
  FileText,
  FolderTree,
  FolderOpen,
  Grid2X2,
  Heart,
  ImageIcon,
  Info,
  List,
  Loader2,
  Maximize2,
  Pin,
  PinOff,
  RefreshCw,
  Search,
  SlidersHorizontal,
  Sparkles,
  Square,
  StickyNote,
  Tag,
  Trash2,
  Wand2,
  X,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import * as App from '@/api'

const props = defineProps({
  latestImages: { type: Array, default: () => [] },
  gallerySummary: { type: Object, default: () => ({}) },
  directoryHealthSummary: { type: Object, default: () => ({}) },
  fileTree: { type: Array, default: () => [] },
  tags: { type: Array, default: () => [] },
  imageTags: { type: Object, default: () => ({}) },
  imageNotes: { type: Object, default: () => ({}) },
  alwaysOnTop: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
})

const emit = defineEmits(['restore', 'toggle-always-on-top', 'refresh', 'open-prompt-assistant'])

const query = ref('')
const viewMode = ref('recent')
const displayMode = ref(localStorage.getItem('compactDisplayMode') || 'list')
const dateFilter = ref('all')
const sortMode = ref('newest')
const selectionMode = ref(false)
const controlsOpen = ref(false)
const folderPanelOpen = ref(false)
const folderSearch = ref('')
const currentImagePage = ref(1)
const pageSize = ref(120)
const selectedFolderId = ref('all')
const selectedPaths = ref(new Set())
const selectedImage = ref(null)
const metadata = ref(null)
const metadataLoading = ref(false)
const metadataError = ref('')
const actionBusy = ref('')
const deleteArmedPath = ref('')
const confirmDialog = ref({
  open: false,
  title: '',
  message: '',
  confirmText: '确定',
  danger: false,
  resolve: null,
})
const favoriteOverrides = ref({})
const tagOverrides = ref({})
const noteOverrides = ref({})
const deletedPaths = ref(new Set())
const noteDraft = ref('')
const tagAddId = ref('')
let metadataRequestId = 0
let deleteTimer = 0

const normalizedQuery = computed(() => query.value.trim().toLowerCase())

const tagById = computed(() => new Map((props.tags || []).map((tag) => [tag.id, tag])))

const getImagePath = (img) => img?.relPath || img?.path || ''

const formatFolderName = (name) => (/^\d{2}$/.test(name || '') ? `${name}月` : name)

const getFolderLabel = (node, fallback = '') => formatFolderName(node?.displayName || node?.name || fallback)

const getNodeChildren = (node) => node?.children || node?.subs || []

const dedupeImages = (images = []) => {
  const map = new Map()
  images.forEach((img) => {
    const key = getImagePath(img)
    if (key && !map.has(key)) map.set(key, img)
  })
  return Array.from(map.values())
}

const collectNodeImages = (node) => {
  const images = []
  const visit = (item) => {
    if (!item) return
    if (Array.isArray(item.images)) images.push(...item.images)
    getNodeChildren(item).forEach(visit)
  }
  visit(node)
  return dedupeImages(images)
}

const libraryImages = computed(() => {
  if (!props.fileTree?.length) return dedupeImages(props.latestImages || [])
  return dedupeImages((props.fileTree || []).flatMap((node) => collectNodeImages(node)))
})

const allFolderRows = computed(() => {
  const rows = [
    {
      id: 'all',
      label: '全部图库',
      depth: 0,
      lineage: '全部图库',
      count: libraryImages.value.length,
      images: libraryImages.value,
    },
  ]

  const visit = (node, depth = 0, lineage = []) => {
    if (!node?.id) return
    const label = getFolderLabel(node, node.id)
    const nextLineage = [...lineage, label]
    const images = collectNodeImages(node)
    if (images.length > 0 || getNodeChildren(node).length > 0) {
      rows.push({
        id: node.id,
        label,
        depth,
        lineage: nextLineage.join(' / '),
        count: images.length,
        images,
      })
    }
    getNodeChildren(node).forEach((child) => visit(child, depth + 1, nextLineage))
  }

  ;(props.fileTree || []).forEach((node) => visit(node))
  return rows
})

const folderRows = computed(() => {
  const term = folderSearch.value.trim().toLowerCase()
  if (!term) return allFolderRows.value
  return allFolderRows.value.filter((row) => row.lineage.toLowerCase().includes(term))
})

const activeFolder = computed(() =>
  allFolderRows.value.find((row) => row.id === selectedFolderId.value) || allFolderRows.value[0],
)

const scopedRawImages = computed(() => activeFolder.value?.images || libraryImages.value)

const normalizeDate = (value) => {
  const date = new Date(value || 0)
  return Number.isNaN(date.getTime()) ? null : date
}

const getImageSrc = (img) => img?.previewPath || img?.cardPath || img?.thumbPath || img?.path || ''

const getIsFavorite = (img) => {
  const path = getImagePath(img)
  if (path && Object.prototype.hasOwnProperty.call(favoriteOverrides.value, path)) {
    return favoriteOverrides.value[path]
  }
  return !!img?.isFavorite
}

const getImageTagIds = (img) => {
  const path = getImagePath(img)
  if (!path) return []
  if (Object.prototype.hasOwnProperty.call(tagOverrides.value, path)) {
    return tagOverrides.value[path]
  }
  return Array.isArray(props.imageTags?.[path]) ? props.imageTags[path] : []
}

const getImageNote = (img) => {
  const path = getImagePath(img)
  if (!path) return ''
  if (Object.prototype.hasOwnProperty.call(noteOverrides.value, path)) {
    return noteOverrides.value[path]
  }
  return props.imageNotes?.[path] || ''
}

const setFavoriteOverride = (path, value) => {
  favoriteOverrides.value = { ...favoriteOverrides.value, [path]: value }
}

const setTagOverride = (path, value) => {
  tagOverrides.value = { ...tagOverrides.value, [path]: value }
}

const setNoteOverride = (path, value) => {
  noteOverrides.value = { ...noteOverrides.value, [path]: value }
}

const allImages = computed(() =>
  scopedRawImages.value
    .filter((img) => img && !deletedPaths.value.has(getImagePath(img)))
    .map((img) => ({
      ...img,
      isFavorite: getIsFavorite(img),
      tagIds: getImageTagIds(img),
      note: getImageNote(img),
    })),
)

const favoriteImages = computed(() => allImages.value.filter((img) => img.isFavorite))
const notedImages = computed(() => allImages.value.filter((img) => !!img.note?.trim()))
const todayCount = computed(() => filterByDate(allImages.value, 'today').length)
const weekCount = computed(() => filterByDate(allImages.value, 'week').length)

const activeSourceImages = computed(() => {
  if (viewMode.value === 'favorites') return favoriteImages.value
  if (viewMode.value === 'notes') return notedImages.value
  return allImages.value
})

const filteredImages = computed(() => {
  let imgs = filterByDate(activeSourceImages.value, dateFilter.value)
  const term = normalizedQuery.value

  if (term) {
    imgs = imgs.filter((img) => {
      const tagNames = getImageTagIds(img).map((tagId) => tagById.value.get(tagId)?.name || '')
      const haystack = [
        img.name,
        img.relPath,
        img.model,
        img.prompt,
        img.note,
        ...tagNames,
        ...(Array.isArray(img.loras) ? img.loras : []),
      ].filter(Boolean).join(' ').toLowerCase()
      return haystack.includes(term)
    })
  }

  const sorted = [...imgs]
  sorted.sort((left, right) => {
    if (sortMode.value === 'oldest') {
      return (normalizeDate(left.modTime)?.getTime() || 0) - (normalizeDate(right.modTime)?.getTime() || 0)
    }
    if (sortMode.value === 'size-desc') {
      return (right.size || 0) - (left.size || 0)
    }
    if (sortMode.value === 'size-asc') {
      return (left.size || 0) - (right.size || 0)
    }
    if (sortMode.value === 'name') {
      return (left.name || '').localeCompare(right.name || '')
    }
    return (normalizeDate(right.modTime)?.getTime() || 0) - (normalizeDate(left.modTime)?.getTime() || 0)
  })

  return sorted
})

const totalImagePages = computed(() => Math.max(1, Math.ceil(filteredImages.value.length / pageSize.value)))

const pageStart = computed(() => (currentImagePage.value - 1) * pageSize.value)

const pageEnd = computed(() => Math.min(filteredImages.value.length, pageStart.value + pageSize.value))

const visibleImages = computed(() => filteredImages.value.slice(pageStart.value, pageEnd.value))

const canPrevImagePage = computed(() => currentImagePage.value > 1)

const canNextImagePage = computed(() => currentImagePage.value < totalImagePages.value)

const prevImagePage = () => {
  if (canPrevImagePage.value) currentImagePage.value -= 1
}

const nextImagePage = () => {
  if (canNextImagePage.value) currentImagePage.value += 1
}

const listTitle = computed(() => {
  if (viewMode.value === 'favorites') return '收藏作品'
  if (viewMode.value === 'notes') return '备注作品'
  return '最近作品'
})

const selectedPathList = computed(() => Array.from(selectedPaths.value))

const selectedImages = computed(() => {
  const selected = selectedPaths.value
  return allImages.value.filter((img) => selected.has(img.relPath))
})

const selectedVisibleCount = computed(() =>
  visibleImages.value.filter((img) => selectedPaths.value.has(img.relPath)).length,
)

const allVisibleSelected = computed(() =>
  visibleImages.value.length > 0 && selectedVisibleCount.value === visibleImages.value.length,
)

const selectedIsFavorite = computed(() => getIsFavorite(selectedImage.value))
const selectedTagIds = computed(() => getImageTagIds(selectedImage.value))
const selectedTags = computed(() => selectedTagIds.value.map((tagId) => tagById.value.get(tagId)).filter(Boolean))
const availableTagsToAdd = computed(() =>
  (props.tags || []).filter((tag) => tag?.id && !selectedTagIds.value.includes(tag.id)),
)
const selectedNote = computed(() => getImageNote(selectedImage.value))
const noteDirty = computed(() => noteDraft.value !== selectedNote.value)
const detailImageSrc = computed(() => getImageSrc(selectedImage.value))
const selectedDetailIndex = computed(() => visibleImages.value.findIndex((img) => img.relPath === selectedImage.value?.relPath))
const canOpenPromptAssistant = computed(() =>
  !!selectedImage.value && !!(metadata.value?.positive || metadata.value?.negative),
)

const detailFacts = computed(() => {
  const img = selectedImage.value
  if (!img) return []

  const facts = [
    { label: '文件大小', value: formatBytes(img.size || 0) },
    { label: '修改时间', value: formatTime(img.modTime) },
    { label: '模型', value: metadata.value?.model || img.model },
    { label: '采样器', value: metadata.value?.sampler },
    { label: 'Seed', value: metadata.value?.seed },
    { label: 'Steps', value: metadata.value?.steps },
    { label: 'CFG', value: metadata.value?.cfg },
  ]

  const width = metadata.value?.width || img.width
  const height = metadata.value?.height || img.height
  if (width && height) {
    facts.splice(2, 0, { label: '尺寸', value: `${width} × ${height}` })
  }

  return facts.filter((fact) => fact.value)
})

const loras = computed(() => {
  if (metadata.value?.loras?.length) return metadata.value.loras
  return Array.isArray(selectedImage.value?.loras) ? selectedImage.value.loras : []
})

function filterByDate(images, mode) {
  if (mode === 'all') return images
  const now = new Date()
  const start = new Date(now)
  start.setHours(0, 0, 0, 0)
  if (mode === 'week') {
    start.setDate(start.getDate() - 6)
  }
  if (mode === 'month') {
    start.setDate(1)
  }
  return images.filter((img) => {
    const date = normalizeDate(img.modTime)
    return date && date >= start
  })
}

const runAction = async (key, action) => {
  if (actionBusy.value) return
  actionBusy.value = key
  try {
    await action()
  } catch (error) {
    console.error(error)
    toast.error(error?.message || `${error}`)
  } finally {
    actionBusy.value = ''
  }
}

const requestConfirm = ({ title = '确认操作', message, confirmText = '确定', danger = false }) =>
  new Promise((resolve) => {
    confirmDialog.value = {
      open: true,
      title,
      message,
      confirmText,
      danger,
      resolve,
    }
  })

const closeConfirm = (result) => {
  const resolver = confirmDialog.value.resolve
  confirmDialog.value = {
    open: false,
    title: '',
    message: '',
    confirmText: '确定',
    danger: false,
    resolve: null,
  }
  if (resolver) resolver(result)
}

const replaceSelectedPaths = (next) => {
  selectedPaths.value = new Set(next)
}

const isPathSelected = (path) => selectedPaths.value.has(path)

const toggleSelection = (img) => {
  if (!img?.relPath) return
  const next = new Set(selectedPaths.value)
  if (next.has(img.relPath)) next.delete(img.relPath)
  else next.add(img.relPath)
  replaceSelectedPaths(next)
}

const toggleSelectionMode = () => {
  selectionMode.value = !selectionMode.value
  if (!selectionMode.value) replaceSelectedPaths([])
}

const toggleSelectVisible = () => {
  const next = new Set(selectedPaths.value)
  if (allVisibleSelected.value) {
    visibleImages.value.forEach((img) => next.delete(img.relPath))
  } else {
    visibleImages.value.forEach((img) => next.add(img.relPath))
  }
  replaceSelectedPaths(next)
}

const clearSelection = () => {
  replaceSelectedPaths([])
}

const selectFolder = (row) => {
  if (!row?.id) return
  selectedFolderId.value = row.id
  folderPanelOpen.value = false
  clearSelection()
  if (selectedImage.value && !row.images.some((img) => img.relPath === selectedImage.value.relPath)) {
    closeDetail()
  }
}

const openDetail = (img) => {
  if (selectionMode.value) {
    toggleSelection(img)
    return
  }
  selectedImage.value = img
  noteDraft.value = getImageNote(img)
  deleteArmedPath.value = ''
}

const closeDetail = () => {
  selectedImage.value = null
  metadata.value = null
  metadataError.value = ''
  deleteArmedPath.value = ''
  noteDraft.value = ''
  tagAddId.value = ''
}

const openAdjacent = (offset) => {
  if (selectedDetailIndex.value < 0 || visibleImages.value.length === 0) return
  const nextIndex = (selectedDetailIndex.value + offset + visibleImages.value.length) % visibleImages.value.length
  openDetail(visibleImages.value[nextIndex])
}

const copyText = async (value, label) => {
  if (!value) return
  await runAction(`copy-${label}`, async () => {
    await App.CopyText(value)
    toast.success(`已复制${label}`)
  })
}

const copySelectedPaths = async () => {
  if (selectedPathList.value.length === 0) return
  await copyText(selectedPathList.value.join('\n'), `${selectedPathList.value.length} 条路径`)
}

const openFile = async (img = selectedImage.value) => {
  if (!img?.relPath) return
  await runAction('open-file', async () => {
    await App.OpenFile(img.relPath)
  })
}

const openLocation = async (img = selectedImage.value) => {
  if (!img?.relPath) return
  await runAction('open-location', async () => {
    await App.OpenImageLocation(img.relPath)
  })
}

const toggleFavorite = async (img = selectedImage.value) => {
  if (!img?.relPath) return
  const nextFavorite = !getIsFavorite(img)
  await runAction('favorite', async () => {
    if (nextFavorite) await App.AddFavorite(img.relPath)
    else await App.RemoveFavorite(img.relPath)
    setFavoriteOverride(img.relPath, nextFavorite)
    if (selectedImage.value?.relPath === img.relPath) {
      selectedImage.value = { ...selectedImage.value, isFavorite: nextFavorite }
    }
    toast.success(nextFavorite ? '已加入收藏' : '已取消收藏')
    emit('refresh')
  })
}

const batchFavorite = async (shouldFavorite) => {
  if (selectedImages.value.length === 0) return
  await runAction(shouldFavorite ? 'batch-favorite' : 'batch-unfavorite', async () => {
    const jobs = selectedImages.value.map((img) =>
      shouldFavorite ? App.AddFavorite(img.relPath) : App.RemoveFavorite(img.relPath),
    )
    await Promise.allSettled(jobs)
    selectedImages.value.forEach((img) => setFavoriteOverride(img.relPath, shouldFavorite))
    toast.success(shouldFavorite ? '已批量收藏' : '已批量取消收藏')
    emit('refresh')
  })
}

const deleteImage = async (img = selectedImage.value) => {
  if (!img?.relPath) return
  if (deleteArmedPath.value !== img.relPath) {
    deleteArmedPath.value = img.relPath
    window.clearTimeout(deleteTimer)
    deleteTimer = window.setTimeout(() => {
      if (deleteArmedPath.value === img.relPath) deleteArmedPath.value = ''
    }, 3000)
    toast.info('再次点击删除可移至回收站')
    return
  }

  await runAction('delete', async () => {
    await App.DeleteImage(img.relPath)
    deletedPaths.value = new Set([...deletedPaths.value, img.relPath])
    closeDetail()
    toast.success('已移至回收站')
    emit('refresh')
  })
}

const batchDelete = async () => {
  if (selectedImages.value.length === 0) return
  const ok = await requestConfirm({
    title: '移至回收站',
    message: `确定将选中的 ${selectedImages.value.length} 张图片移至回收站吗？`,
    confirmText: '移至回收站',
    danger: true,
  })
  if (!ok) return
  await runAction('batch-delete', async () => {
    const targets = selectedImages.value.map((img) => img.relPath)
    await Promise.allSettled(targets.map((path) => App.DeleteImage(path)))
    deletedPaths.value = new Set([...deletedPaths.value, ...targets])
    clearSelection()
    toast.success('已批量移至回收站')
    emit('refresh')
  })
}

const clearPreviewCache = async () => {
  const ok = await requestConfirm({
    title: '清空预览缓存',
    message: '确定清空预览缓存吗？下次查看图片时会重新生成。',
    confirmText: '清空缓存',
  })
  if (!ok) return
  await runAction('clear-cache', async () => {
    const result = await App.ClearPreviewCache()
    toast.success(`已释放 ${formatBytes(result?.bytesFreed || 0)}`)
    emit('refresh')
  })
}

const openOutputFolder = async () => {
  await runAction('open-output', async () => {
    await App.OpenCurrentOutputDirectory()
  })
}

const saveNote = async () => {
  if (!selectedImage.value?.relPath) return
  const path = selectedImage.value.relPath
  await runAction('save-note', async () => {
    await App.SetImageNote(path, noteDraft.value)
    setNoteOverride(path, noteDraft.value.trim())
    selectedImage.value = { ...selectedImage.value, note: noteDraft.value.trim() }
    toast.success('备注已保存')
    emit('refresh')
  })
}

const clearNote = async () => {
  if (!selectedImage.value?.relPath) return
  const path = selectedImage.value.relPath
  await runAction('clear-note', async () => {
    await App.DeleteImageNote(path)
    setNoteOverride(path, '')
    noteDraft.value = ''
    selectedImage.value = { ...selectedImage.value, note: '' }
    toast.success('备注已清空')
    emit('refresh')
  })
}

const addSelectedTag = async () => {
  if (!selectedImage.value?.relPath || !tagAddId.value) return
  const path = selectedImage.value.relPath
  const targetTagId = tagAddId.value
  await runAction('add-tag', async () => {
    const result = await App.AddTagToImage(path, targetTagId)
    const nextTagIds = Array.isArray(result) ? result : [...new Set([...selectedTagIds.value, targetTagId])]
    setTagOverride(path, nextTagIds)
    tagAddId.value = ''
    toast.success('标签已添加')
    emit('refresh')
  })
}

const removeSelectedTag = async (tagId) => {
  if (!selectedImage.value?.relPath || !tagId) return
  const path = selectedImage.value.relPath
  await runAction(`remove-tag-${tagId}`, async () => {
    await App.RemoveTagFromImage(path, tagId)
    setTagOverride(path, selectedTagIds.value.filter((id) => id !== tagId))
    toast.success('标签已移除')
    emit('refresh')
  })
}

const openPromptAssistant = () => {
  if (!selectedImage.value) return
  emit('open-prompt-assistant', {
    initialPositive: metadata.value?.positive || '',
    initialNegative: metadata.value?.negative || '',
    sourcePath: selectedImage.value.relPath || '',
    contextLabel: selectedImage.value.name || '',
  })
}

const loadMetadata = async () => {
  const relPath = selectedImage.value?.relPath
  if (!relPath) {
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
    metadata.value = result || null
  } catch (error) {
    if (requestId !== metadataRequestId) return
    metadata.value = null
    metadataError.value = error?.message || `${error}`
  } finally {
    if (requestId === metadataRequestId) metadataLoading.value = false
  }
}

watch(displayMode, (value) => {
  localStorage.setItem('compactDisplayMode', value)
})

watch(allFolderRows, (rows) => {
  if (!rows.some((row) => row.id === selectedFolderId.value)) {
    selectedFolderId.value = 'all'
  }
})

watch([filteredImages, pageSize], () => {
  if (currentImagePage.value > totalImagePages.value) {
    currentImagePage.value = totalImagePages.value
  }
  if (currentImagePage.value < 1) {
    currentImagePage.value = 1
  }
})

watch([normalizedQuery, viewMode, dateFilter, sortMode, selectedFolderId], () => {
  currentImagePage.value = 1
  clearSelection()
})

watch(() => selectedImage.value?.relPath, () => {
  noteDraft.value = getImageNote(selectedImage.value)
  tagAddId.value = ''
  void loadMetadata()
})

watch(() => props.latestImages, (images) => {
  const relPath = selectedImage.value?.relPath
  if (!relPath) return
  const updated = (images || []).find((img) => img?.relPath === relPath)
  if (updated) {
    selectedImage.value = {
      ...selectedImage.value,
      ...updated,
      isFavorite: getIsFavorite(updated),
      tagIds: getImageTagIds(updated),
      note: getImageNote(updated),
    }
  }
})

const formatBytes = (value) => {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

const formatTime = (value) => {
  if (!value) return ''
  const date = normalizeDate(value)
  if (!date) return ''
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

const folderRowStyle = (row) => ({
  paddingLeft: `${8 + Math.min(row.depth || 0, 5) * 12}px`,
})
</script>

<template>
  <section class="flex h-screen min-h-0 flex-col overflow-hidden bg-background text-foreground">
    <header
      class="grid h-12 shrink-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-2 border-b bg-muted/30 px-3"
      style="--wails-draggable: drag"
    >
      <div class="flex min-w-0 items-center gap-2">
        <Sparkles class="h-4 w-4 shrink-0 text-primary" />
        <div class="min-w-0">
          <div class="truncate text-sm font-semibold leading-none">灵动图库</div>
          <div class="mt-1 truncate text-[11px] leading-none text-muted-foreground">
            今日 {{ todayCount }} · 7天 {{ weekCount }}
          </div>
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-1" style="--wails-draggable: no-drag">
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8 rounded-md"
          :title="alwaysOnTop ? '取消置顶' : '置顶窗口'"
          @click="emit('toggle-always-on-top')"
        >
          <Pin v-if="alwaysOnTop" class="h-4 w-4 text-primary" />
          <PinOff v-else class="h-4 w-4 text-muted-foreground" />
        </Button>
        <Button variant="ghost" size="icon" class="h-8 w-8 rounded-md" title="刷新" @click="emit('refresh')">
          <RefreshCw class="h-4 w-4 text-muted-foreground" :class="{ 'animate-spin': loading }" />
        </Button>
        <Button variant="ghost" size="icon" class="h-8 w-8 rounded-md" title="恢复主窗口" @click="emit('restore')">
          <Maximize2 class="h-4 w-4 text-muted-foreground" />
        </Button>
      </div>
    </header>

    <main v-if="selectedImage" class="flex min-h-0 flex-1 flex-col overflow-hidden">
      <div class="flex h-10 shrink-0 items-center gap-1 border-b px-3">
        <Button variant="ghost" size="icon" class="h-8 w-8 rounded-md" title="返回列表" @click="closeDetail">
          <ArrowLeft class="h-4 w-4" />
        </Button>
        <div class="min-w-0 flex-1 truncate text-xs font-medium">{{ selectedImage.name }}</div>
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8 rounded-md"
          title="上一张"
          :disabled="visibleImages.length < 2"
          @click="openAdjacent(-1)"
        >
          <ChevronLeft class="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="h-8 w-8 rounded-md"
          title="下一张"
          :disabled="visibleImages.length < 2"
          @click="openAdjacent(1)"
        >
          <ChevronRight class="h-4 w-4" />
        </Button>
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto p-3">
        <div class="overflow-hidden rounded-md border bg-card">
          <div class="grid aspect-[4/3] place-items-center bg-muted">
            <img v-if="detailImageSrc" :src="detailImageSrc" :alt="selectedImage.name" class="h-full w-full object-contain" />
            <ImageIcon v-else class="h-8 w-8 text-muted-foreground" />
          </div>
        </div>

        <div class="mt-2 grid grid-cols-4 gap-1.5">
          <Button variant="outline" class="h-8 gap-1 px-2 text-xs" title="打开图片" :disabled="actionBusy === 'open-file'" @click="openFile()">
            <ExternalLink class="h-3.5 w-3.5" />
            打开
          </Button>
          <Button variant="outline" class="h-8 gap-1 px-2 text-xs" title="打开所在位置" :disabled="actionBusy === 'open-location'" @click="openLocation()">
            <FolderOpen class="h-3.5 w-3.5" />
            位置
          </Button>
          <Button
            variant="outline"
            class="h-8 gap-1 px-2 text-xs"
            :class="selectedIsFavorite ? 'border-rose-500/40 text-rose-500' : ''"
            :title="selectedIsFavorite ? '取消收藏' : '加入收藏'"
            :disabled="actionBusy === 'favorite'"
            @click="toggleFavorite()"
          >
            <Heart class="h-3.5 w-3.5" :class="selectedIsFavorite ? 'fill-current' : ''" />
            收藏
          </Button>
          <Button
            variant="outline"
            class="h-8 gap-1 px-2 text-xs text-destructive hover:text-destructive"
            :title="deleteArmedPath === selectedImage.relPath ? '确认删除' : '删除图片'"
            :disabled="actionBusy === 'delete'"
            @click="deleteImage()"
          >
            <Trash2 class="h-3.5 w-3.5" />
            {{ deleteArmedPath === selectedImage.relPath ? '确认' : '删除' }}
          </Button>
        </div>

        <div class="mt-2 grid grid-cols-2 gap-1.5">
          <Button variant="outline" class="h-8 gap-1 px-2 text-xs" @click="copyText(selectedImage.name, '文件名')">
            <Copy class="h-3.5 w-3.5" />
            文件名
          </Button>
          <Button variant="outline" class="h-8 gap-1 px-2 text-xs" @click="copyText(selectedImage.relPath, '路径')">
            <FileImage class="h-3.5 w-3.5" />
            路径
          </Button>
        </div>

        <Button v-if="canOpenPromptAssistant" variant="secondary" class="mt-2 h-8 w-full gap-2 text-xs" @click="openPromptAssistant">
          <Wand2 class="h-3.5 w-3.5" />
          用提示词助手打开
        </Button>

        <div class="mt-3 grid grid-cols-2 gap-2">
          <div v-for="fact in detailFacts" :key="fact.label" class="min-w-0 rounded-md border bg-card px-3 py-2">
            <div class="text-[11px] text-muted-foreground">{{ fact.label }}</div>
            <div class="mt-1 truncate text-xs font-medium" :title="String(fact.value)">{{ fact.value }}</div>
          </div>
        </div>

        <div class="mt-3 rounded-md border bg-card px-3 py-2">
          <div class="flex items-center gap-2 text-xs font-medium">
            <Tag class="h-3.5 w-3.5 text-muted-foreground" />
            标签
          </div>
          <div v-if="selectedTags.length" class="mt-2 flex flex-wrap gap-1.5">
            <button
              v-for="tagItem in selectedTags"
              :key="tagItem.id"
              type="button"
              class="inline-flex max-w-full cursor-pointer items-center gap-1 rounded border bg-muted px-2 py-1 text-[11px] transition-colors hover:bg-muted/70"
              :title="`移除 ${tagItem.name}`"
              @click="removeSelectedTag(tagItem.id)"
            >
              <span class="h-2 w-2 rounded-full" :style="{ backgroundColor: tagItem.color || '#64748b' }"></span>
              <span class="truncate">{{ tagItem.name }}</span>
              <X class="h-3 w-3 text-muted-foreground" />
            </button>
          </div>
          <div v-else class="mt-2 text-[11px] text-muted-foreground">暂无标签</div>
          <div v-if="availableTagsToAdd.length" class="mt-2 grid grid-cols-[minmax(0,1fr)_auto] gap-2">
            <select v-model="tagAddId" class="h-8 min-w-0 rounded-md border bg-background px-2 text-xs">
              <option value="">选择标签</option>
              <option v-for="tagItem in availableTagsToAdd" :key="tagItem.id" :value="tagItem.id">
                {{ tagItem.name }}
              </option>
            </select>
            <Button variant="outline" class="h-8 px-3 text-xs" :disabled="!tagAddId || actionBusy === 'add-tag'" @click="addSelectedTag">
              添加
            </Button>
          </div>
        </div>

        <div class="mt-3 rounded-md border bg-card px-3 py-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2 text-xs font-medium">
              <StickyNote class="h-3.5 w-3.5 text-muted-foreground" />
              备注
            </div>
            <div class="flex items-center gap-1">
              <Button variant="ghost" size="sm" class="h-7 px-2 text-xs" :disabled="!noteDirty || actionBusy === 'save-note'" @click="saveNote">
                保存
              </Button>
              <Button variant="ghost" size="sm" class="h-7 px-2 text-xs text-destructive hover:text-destructive" :disabled="!selectedNote || actionBusy === 'clear-note'" @click="clearNote">
                清空
              </Button>
            </div>
          </div>
          <textarea
            v-model="noteDraft"
            class="mt-2 h-20 w-full resize-none rounded-md border bg-background px-2 py-2 text-xs leading-5 outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
            placeholder="给这张图写一点备注"
          ></textarea>
        </div>

        <div v-if="loras.length" class="mt-3 rounded-md border bg-card px-3 py-2">
          <div class="text-[11px] text-muted-foreground">LoRA</div>
          <div class="mt-2 flex flex-wrap gap-1.5">
            <span v-for="lora in loras" :key="lora" class="max-w-full truncate rounded border bg-muted px-2 py-1 text-[11px]" :title="lora">
              {{ lora }}
            </span>
          </div>
        </div>

        <div v-if="metadataLoading" class="mt-3 flex items-center gap-2 rounded-md border bg-card px-3 py-3 text-xs text-muted-foreground">
          <Loader2 class="h-4 w-4 animate-spin" />
          正在读取元数据
        </div>
        <div v-else-if="metadataError" class="mt-3 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-3 text-xs text-destructive">
          {{ metadataError }}
        </div>
        <div v-else-if="metadata?.positive || metadata?.negative || metadata?.prompt || metadata?.workflow" class="mt-3 space-y-2">
          <div v-if="metadata?.positive" class="rounded-md border bg-card">
            <div class="flex items-center justify-between border-b px-3 py-2">
              <div class="flex items-center gap-2 text-xs font-medium">
                <FileText class="h-3.5 w-3.5 text-muted-foreground" />
                正向 Prompt
              </div>
              <Button variant="ghost" size="sm" class="h-7 px-2 text-xs" @click="copyText(metadata.positive, '正向 Prompt')">
                <Copy class="mr-1 h-3.5 w-3.5" />
                复制
              </Button>
            </div>
            <div class="max-h-32 overflow-y-auto whitespace-pre-wrap break-words px-3 py-2 text-xs leading-5 text-muted-foreground">
              {{ metadata.positive }}
            </div>
          </div>

          <div v-if="metadata?.negative" class="rounded-md border bg-card">
            <div class="flex items-center justify-between border-b px-3 py-2">
              <div class="flex items-center gap-2 text-xs font-medium">
                <FileText class="h-3.5 w-3.5 text-muted-foreground" />
                反向 Prompt
              </div>
              <Button variant="ghost" size="sm" class="h-7 px-2 text-xs" @click="copyText(metadata.negative, '反向 Prompt')">
                <Copy class="mr-1 h-3.5 w-3.5" />
                复制
              </Button>
            </div>
            <div class="max-h-28 overflow-y-auto whitespace-pre-wrap break-words px-3 py-2 text-xs leading-5 text-muted-foreground">
              {{ metadata.negative }}
            </div>
          </div>

          <div v-if="metadata?.prompt || metadata?.workflow" class="grid grid-cols-2 gap-2">
            <Button v-if="metadata?.prompt" variant="outline" class="h-9 gap-2 text-xs" @click="copyText(metadata.prompt, 'Prompt JSON')">
              <Copy class="h-3.5 w-3.5" />
              Prompt JSON
            </Button>
            <Button v-if="metadata?.workflow" variant="outline" class="h-9 gap-2 text-xs" @click="copyText(metadata.workflow, 'Workflow')">
              <Copy class="h-3.5 w-3.5" />
              Workflow
            </Button>
          </div>
        </div>

        <div v-else class="mt-3 flex items-start gap-2 rounded-md border bg-card px-3 py-3 text-xs leading-5 text-muted-foreground">
          <Info class="mt-0.5 h-4 w-4 shrink-0" />
          当前图片没有检测到可读取的 PNG 生成元数据。
        </div>
      </div>
    </main>

    <main v-else class="flex min-h-0 flex-1 flex-col gap-1.5 overflow-hidden p-2">
      <div class="grid shrink-0 grid-cols-[minmax(0,0.95fr)_minmax(0,1.15fr)_auto] gap-1.5">
        <div class="min-w-0 overflow-hidden rounded-md border bg-card">
          <button
            type="button"
            class="flex h-8 w-full cursor-pointer items-center gap-2 px-2.5 text-left text-xs transition-colors hover:bg-muted/50"
            @click="folderPanelOpen = !folderPanelOpen"
          >
            <FolderTree class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <span class="min-w-0 flex-1 truncate">{{ activeFolder?.lineage || '全部图库' }}</span>
            <span class="shrink-0 text-[11px] text-muted-foreground">{{ activeFolder?.count || 0 }}</span>
          </button>
        </div>

        <div class="relative min-w-0">
          <Search class="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input v-model="query" class="h-8 pl-8 pr-3 text-xs" placeholder="搜索文件、模型、标签" />
        </div>

        <Button
          variant="outline"
          size="icon"
          class="h-8 w-8 rounded-md"
          title="筛选和工具"
          :class="controlsOpen ? 'border-primary text-primary' : ''"
          @click="controlsOpen = !controlsOpen"
        >
          <SlidersHorizontal class="h-3.5 w-3.5" />
        </Button>
      </div>

      <div v-if="folderPanelOpen" class="shrink-0 overflow-hidden rounded-md border bg-card p-2">
        <div class="relative">
          <Search class="pointer-events-none absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input v-model="folderSearch" class="h-8 pl-8 pr-3 text-xs" placeholder="搜索目录" />
        </div>
        <div class="mt-2 max-h-32 overflow-y-auto">
          <button
            v-for="row in folderRows"
            :key="row.id"
            type="button"
            class="flex h-8 w-full cursor-pointer items-center gap-2 rounded-md pr-2 text-left text-xs transition-colors hover:bg-muted/50"
            :class="selectedFolderId === row.id ? 'bg-secondary text-primary' : 'text-foreground/80'"
            :style="folderRowStyle(row)"
            :title="row.lineage"
            @click="selectFolder(row)"
          >
            <FolderOpen class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <span class="min-w-0 flex-1 truncate">{{ row.label }}</span>
            <span class="shrink-0 text-[11px] text-muted-foreground">{{ row.count }}</span>
          </button>
        </div>
      </div>

      <div v-if="controlsOpen" class="shrink-0 space-y-1.5 rounded-md border bg-card p-1.5">
        <div class="grid grid-cols-3 gap-1 rounded-md bg-muted/40 p-1">
          <button
            type="button"
            class="h-7 cursor-pointer rounded px-2 text-xs transition-colors"
            :class="viewMode === 'recent' ? 'bg-background font-medium shadow-sm' : 'text-muted-foreground hover:text-foreground'"
            @click="viewMode = 'recent'"
          >
            最近 {{ allImages.length }}
          </button>
          <button
            type="button"
            class="h-7 cursor-pointer rounded px-2 text-xs transition-colors"
            :class="viewMode === 'favorites' ? 'bg-background font-medium shadow-sm' : 'text-muted-foreground hover:text-foreground'"
            @click="viewMode = 'favorites'"
          >
            收藏 {{ favoriteImages.length }}
          </button>
          <button
            type="button"
            class="h-7 cursor-pointer rounded px-2 text-xs transition-colors"
            :class="viewMode === 'notes' ? 'bg-background font-medium shadow-sm' : 'text-muted-foreground hover:text-foreground'"
            @click="viewMode = 'notes'"
          >
            备注 {{ notedImages.length }}
          </button>
        </div>

        <div class="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_92px_auto_auto] gap-1.5">
          <select v-model="dateFilter" class="h-8 min-w-0 rounded-md border bg-background px-2 text-xs">
            <option value="all">全部时间</option>
            <option value="today">今天</option>
            <option value="week">最近 7 天</option>
            <option value="month">本月</option>
          </select>
          <select v-model="sortMode" class="h-8 min-w-0 rounded-md border bg-background px-2 text-xs">
            <option value="newest">最新优先</option>
            <option value="oldest">最旧优先</option>
            <option value="size-desc">体积从大到小</option>
            <option value="size-asc">体积从小到大</option>
            <option value="name">文件名</option>
          </select>
          <select v-model.number="pageSize" class="h-8 min-w-0 rounded-md border bg-background px-2 text-xs" title="每页数量">
            <option :value="60">60 / 页</option>
            <option :value="120">120 / 页</option>
            <option :value="240">240 / 页</option>
          </select>
          <Button variant="outline" size="icon" class="h-8 w-8 rounded-md" title="Output" :disabled="actionBusy === 'open-output'" @click="openOutputFolder">
            <FolderOpen class="h-3.5 w-3.5" />
          </Button>
          <Button variant="outline" size="icon" class="h-8 w-8 rounded-md" title="清缓存" :disabled="actionBusy === 'clear-cache'" @click="clearPreviewCache">
            <Eraser class="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      <div v-if="selectionMode" class="grid shrink-0 grid-cols-4 gap-1.5 rounded-md border bg-card p-2">
        <div class="col-span-4 flex items-center justify-between text-[11px] text-muted-foreground">
          <span>已选 {{ selectedPathList.length }} 张</span>
          <button type="button" class="cursor-pointer hover:text-foreground" @click="clearSelection">清空</button>
        </div>
        <Button variant="outline" class="h-8 px-2 text-xs" :disabled="selectedPathList.length === 0" @click="copySelectedPaths">
          路径
        </Button>
        <Button variant="outline" class="h-8 px-2 text-xs" :disabled="selectedPathList.length === 0" @click="batchFavorite(true)">
          收藏
        </Button>
        <Button variant="outline" class="h-8 px-2 text-xs" :disabled="selectedPathList.length === 0" @click="batchFavorite(false)">
          取消
        </Button>
        <Button variant="outline" class="h-8 px-2 text-xs text-destructive hover:text-destructive" :disabled="selectedPathList.length === 0" @click="batchDelete">
          删除
        </Button>
      </div>

      <div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border bg-card">
        <div class="grid h-9 shrink-0 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-1.5 border-b px-2">
          <div class="min-w-0 truncate text-xs font-medium">{{ listTitle }}</div>
          <div class="flex items-center gap-1 text-[11px] text-muted-foreground">
            <Button variant="ghost" size="icon" class="h-7 w-7 rounded-md" title="上一页" :disabled="!canPrevImagePage" @click="prevImagePage">
              <ChevronLeft class="h-3.5 w-3.5" />
            </Button>
            <span class="w-16 text-center tabular-nums">{{ currentImagePage }} / {{ totalImagePages }}</span>
            <Button variant="ghost" size="icon" class="h-7 w-7 rounded-md" title="下一页" :disabled="!canNextImagePage" @click="nextImagePage">
              <ChevronRight class="h-3.5 w-3.5" />
            </Button>
            <span class="hidden min-w-[76px] text-right sm:inline">
              {{ filteredImages.length }} 张<span v-if="filteredImages.length"> · {{ pageStart + 1 }}-{{ pageEnd }}</span>
            </span>
          </div>
          <div class="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              class="h-7 w-7 rounded-md"
              title="选择"
              :class="selectionMode ? 'text-primary' : ''"
              @click="toggleSelectionMode"
            >
              <CheckSquare class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 rounded-md" title="列表视图" :class="displayMode === 'list' ? 'text-primary' : ''" @click="displayMode = 'list'">
              <List class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-7 w-7 rounded-md" title="网格视图" :class="displayMode === 'grid' ? 'text-primary' : ''" @click="displayMode = 'grid'">
              <Grid2X2 class="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>

        <div v-if="loading && visibleImages.length === 0" class="grid flex-1 place-items-center text-sm text-muted-foreground">
          加载中
        </div>
        <div v-else-if="visibleImages.length === 0" class="grid flex-1 place-items-center px-4 text-center text-sm text-muted-foreground">
          {{ normalizedQuery ? '没有匹配的图片' : '暂无图片' }}
        </div>

        <div v-else-if="displayMode === 'grid'" class="min-h-0 flex-1 overflow-y-auto p-2">
          <div class="grid grid-cols-[repeat(auto-fill,minmax(104px,1fr))] gap-2">
            <button
              v-for="img in visibleImages"
              :key="img.relPath"
              type="button"
              class="group relative aspect-square cursor-pointer overflow-hidden rounded-md border bg-muted text-left transition-colors hover:border-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/60"
              :title="img.name"
              @click="openDetail(img)"
            >
              <img v-if="getImageSrc(img)" :src="getImageSrc(img)" :alt="img.name" class="h-full w-full object-cover" />
              <ImageIcon v-else class="absolute left-1/2 top-1/2 h-5 w-5 -translate-x-1/2 -translate-y-1/2 text-muted-foreground" />
              <div class="absolute inset-x-0 bottom-0 bg-black/55 px-1.5 py-1 text-[10px] text-white opacity-90">
                <div class="truncate">{{ img.name }}</div>
              </div>
              <div class="absolute left-1.5 top-1.5 flex gap-1">
                <span v-if="selectionMode" class="grid h-5 w-5 place-items-center rounded-full bg-black/60 text-white">
                  <Check v-if="isPathSelected(img.relPath)" class="h-3.5 w-3.5" />
                  <Square v-else class="h-3.5 w-3.5" />
                </span>
                <span v-if="img.note" class="grid h-5 w-5 place-items-center rounded-full bg-black/60 text-white">
                  <StickyNote class="h-3 w-3" />
                </span>
              </div>
              <Heart
                class="absolute right-1.5 top-1.5 h-4 w-4 text-white drop-shadow"
                :class="img.isFavorite ? 'fill-rose-500 text-rose-500' : 'opacity-0 group-hover:opacity-100'"
              />
            </button>
          </div>
        </div>

        <div v-else class="min-h-0 flex-1 overflow-y-auto p-1.5">
          <button
            v-for="img in visibleImages"
            :key="img.relPath"
            type="button"
            class="grid w-full cursor-pointer grid-cols-[64px_minmax(0,1fr)_auto] gap-2 rounded-md p-1.5 text-left transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/60"
            :title="`查看 ${img.name} 详情`"
            @click="openDetail(img)"
          >
            <div class="grid aspect-square place-items-center overflow-hidden rounded border bg-muted">
              <img v-if="getImageSrc(img)" :src="getImageSrc(img)" :alt="img.name" class="h-full w-full object-cover" />
              <ImageIcon v-else class="h-5 w-5 text-muted-foreground" />
            </div>
            <div class="min-w-0 self-center">
              <div class="flex min-w-0 items-center gap-1.5">
                <div class="truncate text-xs font-medium">{{ img.name }}</div>
                <StickyNote v-if="img.note" class="h-3 w-3 shrink-0 text-amber-500" />
              </div>
              <div class="mt-1 truncate text-[11px] text-muted-foreground">
                {{ formatTime(img.modTime) }}<span v-if="img.model"> · {{ img.model }}</span>
              </div>
            </div>
            <div class="flex flex-col items-end gap-1">
              <CheckSquare v-if="selectionMode && isPathSelected(img.relPath)" class="h-4 w-4 text-primary" />
              <Square v-else-if="selectionMode" class="h-4 w-4 text-muted-foreground" />
              <Heart
                v-else
                class="h-3.5 w-3.5 text-muted-foreground"
                :class="img.isFavorite ? 'fill-rose-500 text-rose-500' : ''"
              />
            </div>
          </button>
        </div>
      </div>
    </main>

    <div
      v-if="confirmDialog.open"
      class="fixed inset-0 z-50 grid place-items-center bg-black/55 p-4"
      role="dialog"
      aria-modal="true"
      @click.self="closeConfirm(false)"
    >
      <div class="w-full max-w-sm overflow-hidden rounded-lg border bg-popover text-popover-foreground shadow-2xl">
        <div class="border-b px-4 py-3">
          <div class="text-sm font-semibold">{{ confirmDialog.title }}</div>
        </div>
        <div class="px-4 py-4 text-sm leading-6 text-muted-foreground">
          {{ confirmDialog.message }}
        </div>
        <div class="flex items-center justify-end gap-2 border-t bg-muted/20 px-4 py-3">
          <Button variant="outline" class="h-8 px-3 text-xs" @click="closeConfirm(false)">
            取消
          </Button>
          <Button
            class="h-8 px-3 text-xs"
            :variant="confirmDialog.danger ? 'destructive' : 'default'"
            @click="closeConfirm(true)"
          >
            {{ confirmDialog.confirmText }}
          </Button>
        </div>
      </div>
    </div>
  </section>
</template>
