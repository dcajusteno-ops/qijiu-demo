<script setup>
import { ref, computed, provide, onMounted, onUnmounted } from 'vue'
import AppSidebar from './components/AppSidebar.vue'
import ImageGallery from './components/ImageGallery.vue'
import Home from './components/Home.vue'
import Documentation from './components/Documentation.vue'
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
} = useImages(showToast, confirm)

// Selection State (Global or App level)
const isSelectionMode = ref(false)
const selectedPaths = ref(new Set())

// Smart Album filter state
const smartAlbumFilter = ref(null)
const isSidebarCollapsed = ref(false)
const toggleSidebar = () => {
    isSidebarCollapsed.value = !isSidebarCollapsed.value
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

const handleSmartAlbumSelect = (filter) => {
  if (!filter) {
    smartAlbumFilter.value = null
    return
  }
  if (smartAlbumFilter.value?.field === filter.field && smartAlbumFilter.value?.value === filter.value) {
    smartAlbumFilter.value = null
    return
  }
  smartAlbumFilter.value = filter
  if (activeRoot.value === 'dashboard' || activeRoot.value === 'documentation') {
    activeRoot.value = 'output'
    activeSub.value = ''
    activeChild.value = ''
  }
}

const finalPaginatedImages = computed(() => {
  if (!smartAlbumFilter.value) return paginatedImages.value
  // Smart album paths come from all images across all folders,
  // so filter from the full image list, not just the current page/folder
  const filterPaths = new Set(smartAlbumFilter.value.paths)
  const matched = images.value.filter(img => filterPaths.has(img.relPath))
  const startIndex = (currentPage.value - 1) * itemsPerPage.value
  const endIndex = startIndex + itemsPerPage.value
  return matched.slice(startIndex, endIndex)
})

const finalTotalImages = computed(() => {
  if (!smartAlbumFilter.value) return currentImages.value.length
  const filterPaths = new Set(smartAlbumFilter.value.paths)
  return images.value.filter(img => filterPaths.has(img.relPath)).length
})

import { watch } from 'vue'
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

const handleFavoriteGroupsChanged = async () => {
    await fetchImages()
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
    await fetchImageTags()
}

let unsubscribeImagesChanged = null
onMounted(async () => {
    await fetchCustomRoots()
    await fetchImages()
    fetchTags()
    fetchImageTags()
    fetchImageNotes()
    unsubscribeImagesChanged = EventsOn('images:changed', async () => {
        await handleRefresh()
    })
})
onUnmounted(() => {
    if (typeof unsubscribeImagesChanged === 'function') {
        unsubscribeImagesChanged()
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
        :active-smart-album="smartAlbumFilter"
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
        @custom-root-change="fetchCustomRoots"
        @favorite-group-change="handleFavoriteGroupsChanged"
        @clear-preview-cache="handleClearPreviewCache"
        @smart-album-select="handleSmartAlbumSelect"
    />
    
    <div class="flex-1 h-screen overflow-hidden transition-all duration-300">
        <div v-if="activeRoot === 'dashboard'" class="h-full overflow-hidden">
             <Home />
        </div>
        <div v-else-if="activeRoot === 'documentation'" class="h-full overflow-hidden">
             <Documentation />
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
            :tags="tags"
            :image-tags="imageTags"
            :favorite-groups="favoriteGroups"
            :image-notes="imageNotes"
            :smart-album-filter="smartAlbumFilter"
            :current-page="currentPage"
            :items-per-page="itemsPerPage"
            :total-pages="totalPages"
            :is-sidebar-collapsed="isSidebarCollapsed"
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
            @toggle-sidebar="toggleSidebar"
            @clear-smart-album-filter="smartAlbumFilter = null"
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
  </div>
</template>
