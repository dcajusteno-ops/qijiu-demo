<script setup>
import { ref, computed, provide, onMounted, onUnmounted } from 'vue'
import { nextTick } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import ImageGallery from './components/ImageGallery.vue'
import Home from './components/Home.vue'
import Documentation from './components/Documentation.vue'
import ProfileCenter from './components/ProfileCenter.vue'
import DateWorkbench from './components/DateWorkbench.vue'
import AutoRulesPanel from './components/AutoRulesPanel.vue'
import StatisticsDashboard from './components/StatisticsDashboard.vue'
import DirectoryBindingDialog from './components/DirectoryBindingDialog.vue'
import { Toaster } from '@/components/ui/sonner'
import { toast } from 'vue-sonner'
import 'vue-sonner/style.css' // Import sonner styles
import { useImages } from './composables/useImages'
import * as App from '@/api'
import { EventsOn } from '../wailsjs/runtime/runtime'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'

const confirmState = ref({
    isOpen: false,
    title: '',
    message: '',
    resolve: null
})

const confirm = (message) => {
    return new Promise((resolve) => {
        confirmState.value = {
            isOpen: true,
            title: '确认操作',
            message: message,
            resolve
        }
    })
}

const handleConfirm = () => {
    if (confirmState.value.resolve) confirmState.value.resolve(true)
    confirmState.value.isOpen = false
}

const handleCancel = () => {
    if (confirmState.value.resolve) confirmState.value.resolve(false)
    confirmState.value.isOpen = false
}

const showToast = (msg, type) => {
    if (type === 'error') toast.error(msg)
    else toast.success(msg)
}

provide('showToast', showToast)
provide('confirm', confirm)

const {
    images,
    favorites,
    favoriteGroups,
    loading,
    activeRoot,
    activeSub,
    activeChild,
    fileTree,
    scopeImageCount,
    currentImages,
    fetchImages,
    fetchFavorites,
    toggleFavorite,
    toggleRoot,
    handleDelete: deleteImage,
    tags,
    imageTags,
    activeTagFilter,
    fetchTags,
    fetchImageTags,
    createTag,
    deleteTag,
    batchDeleteTags,
    updateTag,
    batchUpdateTags,
    toggleTagFilter,
    getTagCount,
    addTagToImage,
    removeTagFromImage,
    sortBy,
    sortOrder,
    searchQuery,
    setSortBy,
    setSortOrder,
    currentPage,
    itemsPerPage,
    totalPages,
    prevPage,
    nextPage,
    setPage,
    setItemsPerPage,
    paginatedImages,
    openImageLocation,
    customRoots,
    fetchCustomRoots,
    imageNotes,
    fetchImageNotes,
    availableModels,
    availableLoras,
    workbenchFilteredImages,
    dateWorkbenchSummary,
    activeDatePreset,
    activeDateStart,
    activeDateEnd,
    activeModelFilter,
    activeLoraFilter,
    activeDateLabel,
    hasActiveWorkbenchFilters,
    setActiveDatePreset,
    setActiveDateRange,
    setActiveModel,
    setActiveLora,
    clearWorkbenchFilters,
    clearSearchQuery,
} = useImages(showToast, confirm)

// Selection State (Global or App level)
const isSelectionMode = ref(false)
const selectedPaths = ref(new Set())

const isSidebarCollapsed = ref(false)
const showInitialDirectoryBinding = ref(false)
const toggleSidebar = () => {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const setActiveView = (rootId) => {
    activeRoot.value = rootId
    activeSub.value = ''
    activeChild.value = ''
}

const getPreferredGalleryRoot = () => {
    if (fileTree.value.some(node => node.id === 'output')) {
        return 'output'
    }
    const firstGalleryRoot = fileTree.value.find(node => !['favorites', 'statistics', 'date-workbench', 'dashboard', 'profile', 'documentation'].includes(node.id))
    return firstGalleryRoot?.id || 'output'
}

const getNodeLabel = (node, fallback = '') => node?.displayName || node?.name || fallback

const findNodeLineage = (nodes, targetId, lineage = []) => {
    for (const node of nodes || []) {
        const nextLineage = [...lineage, node]
        if (node.id === targetId) return nextLineage

        const children = node.children || node.subs || []
        const found = findNodeLineage(children, targetId, nextLineage)
        if (found) return found
    }
    return null
}

const activeLocation = computed(() => {
    const root = fileTree.value.find(node => node.id === activeRoot.value)
    if (!root) {
        return {
            rootLabel: activeRoot.value || '',
            subLabel: activeSub.value || '',
            childLabel: activeChild.value || '',
            targetFolderPath: ''
        }
    }

    if (!activeSub.value) {
        return {
            rootLabel: getNodeLabel(root, activeRoot.value),
            subLabel: '',
            childLabel: '',
            targetFolderPath: root.relPath || ''
        }
    }

    const lineage = findNodeLineage(root.children || root.subs || [], activeSub.value) || []
    const labels = lineage.map(node => getNodeLabel(node)).filter(Boolean)
    const targetNode = lineage[lineage.length - 1]

    return {
        rootLabel: getNodeLabel(root, activeRoot.value),
        subLabel: labels[0] || '',
        childLabel: labels.slice(1).join(' / '),
        targetFolderPath: targetNode?.relPath || root.relPath || ''
    }
})

const updateSearchQuery = (value) => {
  searchQuery.value = value
}

const clearAllGalleryFilters = () => {
    clearSearchQuery()
    clearWorkbenchFilters()
}

const openWorkbenchGallery = () => {
    setActiveView(getPreferredGalleryRoot())
}

const getDateArchiveRootId = () => {
    const builtinArchive = (customRoots.value || []).find(
        (root) => root?.id === 'builtin-date-archive' && root.enabled !== false,
    )
    if (builtinArchive) {
        return `custom:${builtinArchive.id}`
    }
    return getPreferredGalleryRoot()
}

const finalPaginatedImages = computed(() => paginatedImages.value)

const finalTotalImages = computed(() => currentImages.value.length)

import { watch } from 'vue'
watch(activeRoot, (next, prev) => {
    if (next === prev) return
    activeSub.value = ''
    activeChild.value = ''
})

watch([activeRoot, activeSub, activeChild], () => {
    isSelectionMode.value = false
    selectedPaths.value.clear()
})

const toggleSelectionMode = () => {
    isSelectionMode.value = !isSelectionMode.value
    selectedPaths.value.clear()
}

const toggleSelection = (img) => {
    if (selectedPaths.value.has(img.relPath)) {
        selectedPaths.value.delete(img.relPath)
    } else {
        selectedPaths.value.add(img.relPath)
    }
}

const selectAllCurrent = () => {
    // Only select items on the current page to prevent accidental operations on unseen images
    paginatedImages.value.forEach(img => selectedPaths.value.add(img.relPath))
}

const clearSelection = () => {
    selectedPaths.value.clear()
}

const handleCleanEmpty = async () => {
    const ok = await confirm('确定要清理所有空文件夹吗？')
    if (!ok) return
    try {
        const count = await App.CleanEmptyFolders()
        showToast(`已清理 ${count} 个空文件夹`, 'success')
        fetchImages()
    } catch (e) {
        showToast(`清理失败: ${e}`, 'error')
    }
}

const formatBytes = (value) => {
    if (!value) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB']
    let size = value
    let index = 0
    while (size >= 1024 && index < units.length - 1) {
        size /= 1024
        index++
    }
    return `${size.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

const handleClearPreviewCache = async () => {
    const ok = await confirm('确定要清空预览缓存吗？下次查看时会重新生成。')
    if (!ok) return
    try {
        const result = await App.ClearPreviewCache()
        showToast(`已清空 ${result.deletedFiles || 0} 个缓存文件，释放 ${formatBytes(result.bytesFreed || 0)}`, 'success')
    } catch (e) {
        showToast(`清理失败: ${e}`, 'error')
    }
}

const handleOrganizeFiles = async () => {
    const ok = await confirm('确定要按日期自动整理文件吗？这将把散落在根目录的图片移动到年/月子文件夹中。')
    if (!ok) return
    try {
        const count = await App.OrganizeFiles('month')
        showToast(`已整理 ${count} 张图片`, 'success')
        handleRefresh()
    } catch (e) {
        showToast(`整理失败: ${e}`, 'error')
    }
}

const handleFavoriteGroupsChanged = async () => {
    await fetchImages()
}

const handleCustomRootChanged = async () => {
    await fetchCustomRoots()
    await nextTick()
    if (activeRoot.value.startsWith('custom:') && !fileTree.value.some((node) => node.id === activeRoot.value)) {
        setActiveView(getPreferredGalleryRoot())
    }
}

const handleShortcutAction = async (actionId) => {
    switch (actionId) {
    case 'switch_dashboard':
        setActiveView('dashboard')
        break
    case 'switch_gallery':
        setActiveView(getPreferredGalleryRoot())
        break
    case 'switch_favorites':
        setActiveView('favorites')
        break
    case 'switch_documentation':
        setActiveView('documentation')
        break
    case 'switch_auto_rules':
        setActiveView('auto-rules')
        break
    case 'switch_date_workbench':
        setActiveView('date-workbench')
        break
    case 'refresh_images':
        await handleRefresh()
        break
    case 'toggle_sidebar':
        toggleSidebar()
        break
    case 'toggle_selection_mode':
        toggleSelectionMode()
        break
    default:
        console.warn('Unknown shortcut action:', actionId)
    }
}

const deleteSelected = async () => {
    if (selectedPaths.value.size === 0) return
    const count = selectedPaths.value.size
    const ok = await confirm(`确定要将选中的 ${count} 张图片移至回收站吗？`)
    if (!ok) return

    const targets = Array.from(selectedPaths.value)
    
    // Optimistic UI Update
    images.value = images.value.filter(img => !selectedPaths.value.has(img.relPath))
    
    // Update tag counts by removing tags for deleted images
    selectedPaths.value.forEach(path => {
        if (imageTags.value[path]) {
            delete imageTags.value[path]
        }
    })

    selectedPaths.value.clear()
    isSelectionMode.value = false 
    
    // BatchDeleteImages doesn't exist in bindings, iterate and delete one by one
    let successCount = 0
    let failCount = 0
    try {
        const results = await Promise.allSettled(targets.map(p => App.DeleteImage(p)))
        results.forEach(r => { if (r.status === 'fulfilled') successCount++; else failCount++ })
        if (failCount > 0) {
            showToast(`${failCount} 张图片删除失败`, 'error')
            fetchImages()
        } else {
            showToast(`成功删除 ${successCount} 张图片`, 'success')
        }
    } catch (e) {
        showToast(`删除请求失败: ${e}`, 'error')
        fetchImages()
    }
}

const handleRefresh = async () => {
    await fetchImages()
    await fetchTags()
    await fetchImageTags()
    await fetchImageNotes()
}

const handleDirectoryBindingChanged = async () => {
    isSelectionMode.value = false
    selectedPaths.value.clear()
    await fetchCustomRoots()
    await handleRefresh()
    setActiveView(getPreferredGalleryRoot())
}

const handleOpenCurrentOutput = async () => {
    try {
        await App.OpenCurrentOutputDirectory()
    } catch (e) {
        showToast(`打开 output 文件夹失败: ${e}`, 'error')
    }
}

let unsubscribeImagesChanged = null
let unsubscribeShortcutTriggered = null
onMounted(async () => {
    await fetchCustomRoots()
    try {
        const binding = await App.GetDirectoryBinding()
        if (!binding?.configured) {
            showInitialDirectoryBinding.value = true
        }
    } catch (e) {
        console.error('Failed to load directory binding:', e)
        showInitialDirectoryBinding.value = true
    }
    try {
        const profile = await App.GetUserProfile()
        if (profile?.preferredStartPage) {
            setActiveView(profile.preferredStartPage)
        }
    } catch (e) {
        console.error('Failed to load preferred start page:', e)
    }
    await fetchImages()
    fetchTags()
    fetchImageTags()
    fetchImageNotes()
    unsubscribeImagesChanged = EventsOn('images:changed', async () => {
        await handleRefresh()
    })
    unsubscribeShortcutTriggered = EventsOn('shortcut:triggered', async (actionId) => {
        await handleShortcutAction(actionId)
    })
})
onUnmounted(() => {
    if (typeof unsubscribeImagesChanged === 'function') {
        unsubscribeImagesChanged()
    }
    if (typeof unsubscribeShortcutTriggered === 'function') {
        unsubscribeShortcutTriggered()
    }
})
</script>

<template>
  <div class="flex h-screen bg-background text-foreground font-sans antialiased overflow-hidden">
    <AppSidebar
        :file-tree="fileTree"
        :active-root="activeRoot"
        :active-sub="activeSub"
        :active-child="activeChild"
        :is-selection-mode="isSelectionMode"
        :tags="tags"
        :active-tag-filter="activeTagFilter"
        :get-tag-count="getTagCount"
        :collapsed="isSidebarCollapsed"
        :custom-roots="customRoots"
        :favorite-groups="favoriteGroups"
        @update:activeRoot="toggleRoot"
        @update:activeSub="(val) => activeSub = val"
        @update:activeChild="(val) => activeChild = val"
        @toggle-selection-mode="toggleSelectionMode"
        @clean-empty-folders="handleCleanEmpty"
        @create-tag="createTag"
        @delete-tag="deleteTag"
        @batch-delete-tags="batchDeleteTags"
        @update-tag="updateTag"
        @batch-update-tags="batchUpdateTags"
        @toggle-tag-filter="toggleTagFilter"
        @refresh-images="handleRefresh"
        @toggle-collapse="toggleSidebar"
        @custom-root-change="handleCustomRootChanged"
        @directory-binding-change="handleDirectoryBindingChanged"
        @favorite-group-change="handleFavoriteGroupsChanged"
        @clear-preview-cache="handleClearPreviewCache"
        @organize-files="handleOrganizeFiles"
        @open-current-output="handleOpenCurrentOutput"
    />
    
    <div class="flex-1 h-screen overflow-hidden transition-all duration-300">
        <div v-if="activeRoot === 'dashboard'" class="h-full overflow-hidden">
             <Home
                :archive-root-id="getDateArchiveRootId()"
                @navigate-root="setActiveView"
                @clear-filters="clearAllGalleryFilters"
             />
        </div>
        <div v-else-if="activeRoot === 'profile'" class="h-full overflow-hidden">
             <ProfileCenter @navigate="setActiveView" />
        </div>
        <div v-else-if="activeRoot === 'documentation'" class="h-full overflow-hidden">
             <Documentation />
        </div>
        <div v-else-if="activeRoot === 'statistics'" class="h-full overflow-hidden">
             <StatisticsDashboard />
        </div>
        <div v-else-if="activeRoot === 'auto-rules'" class="h-full overflow-hidden">
             <AutoRulesPanel />
        </div>
        <div v-else-if="activeRoot === 'date-workbench'" class="h-full overflow-hidden">
             <DateWorkbench
                :summary="dateWorkbenchSummary"
                :available-models="availableModels"
                :available-loras="availableLoras"
                :active-date-preset="activeDatePreset"
                :active-date-start="activeDateStart"
                :active-date-end="activeDateEnd"
                :active-model-filter="activeModelFilter"
                :active-lora-filter="activeLoraFilter"
                :active-date-label="activeDateLabel"
                :filtered-count="workbenchFilteredImages.length"
                @update:date-preset="setActiveDatePreset"
                @update:date-range="setActiveDateRange"
                @update:model-filter="setActiveModel"
                @update:lora-filter="setActiveLora"
                @clear-filters="clearWorkbenchFilters"
                @open-gallery="openWorkbenchGallery"
             />
        </div>
        <ImageGallery
            v-else
            :images="finalPaginatedImages"
            :total-images="finalTotalImages"
            :loading="loading"
            :root-name="activeRoot"
            :sub-name="activeSub"
            :child-name="activeChild"
            :root-label="activeLocation.rootLabel"
            :sub-label="activeLocation.subLabel"
            :child-label="activeLocation.childLabel"
            :target-folder-path="activeLocation.targetFolderPath"
            :is-selection-mode="isSelectionMode"
            :selected-paths="selectedPaths"
            :scope-image-count="scopeImageCount"
            :tags="tags"
            :image-tags="imageTags"
            :favorite-groups="favoriteGroups"
            :image-notes="imageNotes"
            :search-query="searchQuery"
            :available-models="availableModels"
            :available-loras="availableLoras"
            :active-date-label="activeDateLabel"
            :active-model-filter="activeModelFilter"
            :active-lora-filter="activeLoraFilter"
            :has-active-workbench-filters="hasActiveWorkbenchFilters"
            :current-page="currentPage"
            :items-per-page="itemsPerPage"
            :total-pages="totalPages"
            @delete="deleteImage"
            @toggle-selection="toggleSelection"
            @select-all="selectAllCurrent"
            @clear-selection="clearSelection"
            @delete-selected="deleteSelected"
            @toggle-favorite="toggleFavorite"
            @add-tag="addTagToImage"
            @remove-tag="removeTagFromImage"
            @view-favorites="toggleRoot('favorites')"
            @view-statistics="toggleRoot('statistics')"
            @refresh-images="handleRefresh"
            @favorite-groups-changed="handleFavoriteGroupsChanged"
            @page-change="setPage"
            @items-per-page-change="setItemsPerPage"
            @open-location="openImageLocation"
            @update:search-query="updateSearchQuery"
            @update:model-filter="setActiveModel"
            @update:lora-filter="setActiveLora"
            @clear-workbench-filters="clearWorkbenchFilters"
            @clear-all-filters="clearAllGalleryFilters"
        />
    </div>

    <Toaster position="top-center" richColors />
    
    <AlertDialog :open="confirmState.isOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ confirmState.title }}</AlertDialogTitle>
          <AlertDialogDescription>
            {{ confirmState.message }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel @click="handleCancel">取消</AlertDialogCancel>
          <AlertDialogAction @click="handleConfirm">确定</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <DirectoryBindingDialog
      v-model:open="showInitialDirectoryBinding"
      :required="true"
      @change="handleDirectoryBindingChanged"
    />
  </div>
</template>

