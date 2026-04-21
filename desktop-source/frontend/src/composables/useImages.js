import { ref, computed, watch } from 'vue'
import * as App from '@/api'
import { useImageStacks } from './useImageStacks'
import { useGalleryData } from './useGalleryData'
import { useLibraryMeta } from './useLibraryMeta'
import { useWorkbenchFilters } from './useWorkbenchFilters'
import {
  buildImageDisplayPath,
  getDateSegment,
  normalizeFolderPath,
  syncGroupedFilterValue,
} from './useGalleryHelpers'

const activeRoot = ref('dashboard')
const activeSub = ref('')
const activeChild = ref('')
const isInitialized = ref(false)

const customRoots = ref([])

const filters = ref({
  dateRange: { start: null, end: null },
  size: { min: null, max: null },
  dimensions: { minW: null, minH: null },
})
const searchQuery = ref(localStorage.getItem('gallerySearchQuery') || '')

const sortBy = ref(localStorage.getItem('sortBy') || 'time')
const sortOrder = ref(localStorage.getItem('sortOrder') || 'desc')
const isStackingEnabled = ref(localStorage.getItem('isStackingEnabled') !== 'false') // default true

export function useImages(showToast = () => {}, confirm = async () => false) {
  const {
    images,
    indexedImages,
    loading,
    currentPage,
    itemsPerPage,
    performanceSettings,
    gallerySummary,
    directoryHealthSummary,
    workbenchAggregate,
    galleryLoadMode,
    pagedImages,
    pagedTotal,
    pagedTotalPages,
    hasMorePagedImages,
    isPagedLoading,
    isPagedAppending,
    modeReason,
    lastSuccessfulQuery,
    loadPerformanceSettings: loadGalleryPerformanceSettings,
    savePerformanceSettings: saveGalleryPerformanceSettings,
    fetchGallerySummary,
    fetchDirectoryHealthSummary,
    fetchWorkbenchAggregate: fetchGalleryWorkbenchAggregate,
    fetchImages: fetchGalleryImages,
    fetchImageIndex: fetchGalleryImageIndex,
    ensureStandardImagesReady: ensureGalleryStandardImagesReady,
    removeImagesLocally: removeGalleryImagesLocally,
    fetchImagesPage: fetchGalleryImagesPage,
    invalidatePagedRequests,
    useGalleryPagination,
  } = useGalleryData()

  const sourceImages = computed(() => (images.value.length > 0 ? images.value : indexedImages.value))

  const {
    tags,
    imageTags,
    imageNotes,
    favorites,
    favoriteGroups,
    activeTagFilter,
    fetchFavorites,
    fetchTags,
    fetchImageTags,
    fetchImageNotes,
    createTag,
    deleteTag,
    batchDeleteTags,
    batchUpdateTags,
    updateTag,
    addTagToImage,
    removeTagFromImage,
    toggleTagFilter,
    getTagCount,
    toggleFavorite,
  } = useLibraryMeta({
    showToast,
    confirm,
    activeRoot,
    activeSub,
  })

  const {
    activeDatePreset,
    activeDateStart,
    activeDateEnd,
    activeModelFilter,
    activeLoraFilter,
    activeDateRange,
    activeDateLabel,
    hasActiveWorkbenchFilters,
    workbenchFilteredImages,
    availableModels,
    availableLoras,
    dateWorkbenchSummary,
    workbenchFilteredCount,
    imageMatchesWorkbenchFilters,
    setActiveDatePreset,
    setActiveDateRange,
    clearDateFilter,
    setActiveModel,
    setActiveLora,
    clearWorkbenchFilters,
  } = useWorkbenchFilters(sourceImages, workbenchAggregate)

  const fetchCustomRoots = async () => {
    try {
      const roots = await App.GetCustomRoots()
      customRoots.value = roots || []
    } catch (e) {
      console.error('Failed to fetch custom roots:', e)
    }
  }

  const loadPerformanceSettings = async () => {
    await loadGalleryPerformanceSettings()
  }

  const savePerformanceSettings = async (nextSettings) => {
    return saveGalleryPerformanceSettings(nextSettings)
  }

  const fetchWorkbenchAggregate = async () => {
    return fetchGalleryWorkbenchAggregate({
      activeDatePreset: activeDatePreset.value || 'all',
      activeDateStart: activeDateStart.value || '',
      activeDateEnd: activeDateEnd.value || '',
      activeModelFilter: activeModelFilter.value || '',
      activeLoraFilter: activeLoraFilter.value || '',
    })
  }

  const fetchImages = async () => {
    await fetchGalleryImages({
      sortBy: sortBy.value,
      sortOrder: sortOrder.value,
      favoriteGroupsRef: favoriteGroups,
      favoritesRef: favorites,
      mapLoadedImage,
    })
  }

  const fetchImageIndex = async () => {
    await fetchGalleryImageIndex({
      sortBy: sortBy.value,
      sortOrder: sortOrder.value,
      favoriteGroupsRef: favoriteGroups,
      favoritesRef: favorites,
      mapLoadedImage,
    })
  }

  const fileTree = computed(() => {
    let imagesToUse = sourceImages.value
    if (activeTagFilter.value) {
      imagesToUse = sourceImages.value.filter((img) =>
        imageTags.value[img.relPath]?.includes(activeTagFilter.value),
      )
    }

    const enabledCustomRoots = (customRoots.value || [])
      .filter((root) => root && root.enabled !== false)
      .sort((a, b) => (a.order || 0) - (b.order || 0))

    const folderRoots = enabledCustomRoots.filter((root) => root.id !== 'builtin-date-archive')
    const customRootPaths = folderRoots
      .map((root) => normalizeFolderPath(root.path))
      .filter(Boolean)

    const isManagedByFolderRoot = (relPath) => {
      const normalized = normalizeFolderPath(relPath)
      if (!normalized) return false
      return customRootPaths.some((rootPath) =>
        normalized === rootPath || normalized.startsWith(`${rootPath}/`),
      )
    }

    const buildLeafNode = (name, id, relPath, imgs) => ({
      name,
      id,
      displayName: name,
      relPath,
      children: [],
      subs: [],
      images: imgs,
      lastMod: imgs.length > 0 ? Math.max(...imgs.map((img) => new Date(img.modTime).getTime())) : 0,
    })

    const sortNodes = (nodes = []) => {
      nodes.sort((a, b) => {
        const timeDiff = (b.lastMod || 0) - (a.lastMod || 0)
        if (timeDiff !== 0) return timeDiff
        return (a.name || '').localeCompare(b.name || '')
      })
      nodes.forEach((node) => {
        if (node.children?.length) sortNodes(node.children)
        node.subs = node.children || []
      })
      return nodes
    }

    const buildStandardTree = (rootPath, name, id, imgs, icon) => {
      const normalizedRootPath = normalizeFolderPath(rootPath)
      const prefix = normalizedRootPath ? `${normalizedRootPath}/` : ''
      const rootNode = {
        name,
        id,
        displayName: name,
        relPath: normalizedRootPath,
        icon,
        children: [],
        subs: [],
        images: [],
        lastMod: 0,
      }

      const childMap = new Map()
      imgs.forEach((img) => {
        const relPath = normalizeFolderPath(img.relPath)
        const rest = normalizedRootPath && relPath.startsWith(prefix) ? relPath.slice(prefix.length) : relPath
        if (!rest || !rest.includes('/')) {
          rootNode.images.push(img)
          return
        }

        const childName = rest.split('/')[0]
        if (!childMap.has(childName)) childMap.set(childName, [])
        childMap.get(childName).push(img)
      })

      rootNode.children = Array.from(childMap.entries()).map(([childName, childImages]) =>
        buildStandardTree(
          normalizedRootPath ? `${normalizedRootPath}/${childName}` : childName,
          childName,
          `${id}/${childName}`,
          childImages,
          icon,
        ),
      )

      rootNode.lastMod = Math.max(
        rootNode.images.length ? Math.max(...rootNode.images.map((img) => new Date(img.modTime).getTime())) : 0,
        rootNode.children.length ? Math.max(...rootNode.children.map((child) => child.lastMod || 0)) : 0,
      )
      sortNodes(rootNode.children)
      rootNode.subs = rootNode.children
      return rootNode
    }

    const buildYearGroupedRoot = (root, imgs) => {
      const normalizedRootPath = normalizeFolderPath(root.path)
      const prefix = normalizedRootPath ? `${normalizedRootPath}/` : ''
      const rootNode = {
        name: root.name,
        id: `custom:${root.id}`,
        displayName: root.name,
        relPath: normalizedRootPath,
        icon: root.icon || 'FolderSymlink',
        type: 'root',
        order: root.order || 0,
        enabled: root.enabled !== false,
        locked: !!root.locked,
        isBuiltin: !!root.isBuiltin,
        pinned: !!root.pinned,
        images: [],
        children: [],
        subs: [],
        lastMod: 0,
      }

      const yearMap = new Map()
      const regularImages = []

      imgs.forEach((img) => {
        const relPath = normalizeFolderPath(img.relPath)
        if (!(relPath === normalizedRootPath || relPath.startsWith(prefix))) return

        const rest = relPath === normalizedRootPath ? '' : relPath.slice(prefix.length)
        const folderRel = normalizeFolderPath(rest.split('/').slice(0, -1).join('/'))
        if (!folderRel) {
          regularImages.push(img)
          return
        }

        const dateSegment = getDateSegment(folderRel)
        if (!dateSegment) {
          regularImages.push(img)
          return
        }

        const leafPath = folderRel.split('/').slice(0, folderRel.split('/').indexOf(dateSegment) + 1).join('/')
        const fullLeafPath = normalizeFolderPath(`${normalizedRootPath}/${leafPath}`)
        const year = dateSegment.slice(0, 4)
        if (!yearMap.has(year)) yearMap.set(year, new Map())
        const leafMap = yearMap.get(year)
        if (!leafMap.has(fullLeafPath)) leafMap.set(fullLeafPath, [])
        leafMap.get(fullLeafPath).push(img)
      })

      if (regularImages.length > 0) {
        rootNode.images = regularImages.filter((img) => {
          const relPath = normalizeFolderPath(img.relPath)
          const rest = relPath === normalizedRootPath ? '' : relPath.slice(prefix.length)
          return !rest.includes('/')
        })

        const nonDateImages = regularImages.filter((img) => !rootNode.images.includes(img))
        if (nonDateImages.length > 0) {
          const regularTree = buildStandardTree(normalizedRootPath, root.name, rootNode.id, nonDateImages, rootNode.icon)
          rootNode.children.push(...(regularTree.children || []))
        }
      }

      yearMap.forEach((leafMap, year) => {
        const yearNode = {
          name: year,
          id: `${rootNode.id}/${year}`,
          displayName: year,
          relPath: '',
          images: [],
          children: [],
          subs: [],
          lastMod: 0,
        }

        leafMap.forEach((leafImages, fullLeafPath) => {
          const leafName = fullLeafPath.split('/').pop()
          yearNode.children.push(buildLeafNode(leafName, `${yearNode.id}/${fullLeafPath}`, fullLeafPath, leafImages))
        })

        sortNodes(yearNode.children)
        yearNode.lastMod = yearNode.children.length ? Math.max(...yearNode.children.map((child) => child.lastMod || 0)) : 0
        yearNode.subs = yearNode.children
        rootNode.children.push(yearNode)
      })

      sortNodes(rootNode.children)
      rootNode.subs = rootNode.children
      rootNode.lastMod = Math.max(
        rootNode.images.length ? Math.max(...rootNode.images.map((img) => new Date(img.modTime).getTime())) : 0,
        rootNode.children.length ? Math.max(...rootNode.children.map((child) => child.lastMod || 0)) : 0,
      )
      return rootNode
    }

    const buildDateArchiveNode = (root, imgs) => {
      const archiveNode = {
        name: root.name,
        id: `custom:${root.id}`,
        displayName: root.name,
        relPath: '',
        icon: root.icon || 'Calendar',
        type: 'root',
        order: root.order || 0,
        enabled: root.enabled !== false,
        locked: !!root.locked,
        isBuiltin: true,
        pinned: !!root.pinned,
        images: [],
        children: [],
        subs: [],
        lastMod: 0,
      }

      const yearMap = new Map()
      imgs.forEach((img) => {
        if (isManagedByFolderRoot(img.relPath)) return

        const relPath = normalizeFolderPath(img.relPath)
        const folderRel = normalizeFolderPath(relPath.split('/').slice(0, -1).join('/'))
        const dateSegment = getDateSegment(folderRel)
        if (!dateSegment) return

        const parts = folderRel.split('/')
        const leafPath = parts.slice(0, parts.indexOf(dateSegment) + 1).join('/')
        const year = dateSegment.slice(0, 4)
        if (!yearMap.has(year)) yearMap.set(year, new Map())
        const leafMap = yearMap.get(year)
        if (!leafMap.has(leafPath)) leafMap.set(leafPath, [])
        leafMap.get(leafPath).push(img)
      })

      yearMap.forEach((leafMap, year) => {
        const yearNode = {
          name: year,
          id: `${archiveNode.id}/${year}`,
          displayName: year,
          relPath: '',
          images: [],
          children: [],
          subs: [],
          lastMod: 0,
        }
        leafMap.forEach((leafImages, leafPath) => {
          const leafName = leafPath.split('/').pop()
          yearNode.children.push(buildLeafNode(leafName, `${yearNode.id}/${leafPath}`, leafPath, leafImages))
        })
        sortNodes(yearNode.children)
        yearNode.lastMod = yearNode.children.length ? Math.max(...yearNode.children.map((child) => child.lastMod || 0)) : 0
        yearNode.subs = yearNode.children
        archiveNode.children.push(yearNode)
      })

      sortNodes(archiveNode.children)
      archiveNode.subs = archiveNode.children
      archiveNode.lastMod = archiveNode.children.length ? Math.max(...archiveNode.children.map((child) => child.lastMod || 0)) : 0
      return archiveNode
    }

    const favoriteGroupsNodes = (favoriteGroups.value || [])
      .map((group) => {
        const normalizedGroupPaths = new Set((group.paths || []).map((path) => normalizeFolderPath(path)))
        const groupImages = imagesToUse
          .filter((img) => normalizedGroupPaths.has(normalizeFolderPath(img.relPath)))
          .sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
        return {
          name: group.name,
          id: `favorite-group:${group.id}`,
          displayName: group.name,
          relPath: '',
          groupId: group.id,
          children: [],
          subs: [],
          images: groupImages,
          lastMod: groupImages.length ? Math.max(...groupImages.map((img) => new Date(img.modTime).getTime())) : 0,
          count: groupImages.length,
          isFavoriteGroup: true,
        }
      })
      .sort((a, b) => (b.lastMod || 0) - (a.lastMod || 0))

    const favoritesRoot = {
      name: '收藏夹',
      id: 'favorites',
      displayName: '收藏夹',
      relPath: '',
      children: favoriteGroupsNodes,
      subs: favoriteGroupsNodes,
      images: [],
      type: 'root',
      icon: 'Heart',
      count: favoriteGroupsNodes.reduce((sum, item) => sum + (item.count || 0), 0),
    }

    const defaultImages = imagesToUse
    const defaultRoot = buildStandardTree('', '默认目录', 'output', defaultImages, 'FolderOpen')
    defaultRoot.type = 'root'
    defaultRoot.icon = 'FolderOpen'
    defaultRoot.displayName = '默认目录'

    const nodes = [favoritesRoot, defaultRoot]
    enabledCustomRoots.forEach((root) => {
      if (root.id === 'builtin-date-archive') {
        nodes.push(buildDateArchiveNode(root, imagesToUse))
        return
      }
      nodes.push(buildYearGroupedRoot(root, imagesToUse))
    })

    return nodes.map((node) => ({
      ...node,
      subs: node.children || [],
    }))
  })

  const findNodeById = (nodes, id) => {
    for (const node of nodes || []) {
      if (node.id === id) return node
      const found = findNodeById(node.children || node.subs || [], id)
      if (found) return found
    }
    return null
  }

  const selectedGalleryNode = computed(() => {
    const root = fileTree.value.find((node) => node.id === activeRoot.value)
    if (!root) return null
    if (!activeSub.value) return root
    return findNodeById(root.children || root.subs || [], activeSub.value) || root
  })

  const isSpecialGalleryRoot = computed(() =>
    ['dashboard', 'profile', 'documentation', 'statistics', 'date-workbench', 'prompt-assistant', 'auto-rules']
      .includes(activeRoot.value),
  )

  const supportsPagedGalleryView = computed(() => {
    if (isSpecialGalleryRoot.value) return false
    if (activeRoot.value === 'favorites') return true
    if (activeRoot.value === 'output') return true
    if (activeRoot.value.startsWith('custom:')) {
      return !!selectedGalleryNode.value?.relPath
    }
    return false
  })

  const hasAdvancedLocalFilters = computed(() => {
    const { dateRange, size, dimensions } = filters.value
    return !!(
      dateRange?.start ||
      dateRange?.end ||
      size?.min !== null ||
      size?.max !== null ||
      dimensions?.minW !== null ||
      dimensions?.minH !== null
    )
  })

  const effectivePreferredMode = computed(() => {
    const selectedMode = performanceSettings.value?.mode || 'auto'
    if (selectedMode === 'performance') return 'performance'
    if (selectedMode === 'standard') return 'standard'
    return (gallerySummary.value?.totalImages || 0) >= 3000 ? 'performance' : 'standard'
  })

  const refreshGalleryMode = () => {
    const preferred = effectivePreferredMode.value
    if (preferred === 'performance' && supportsPagedGalleryView.value && !hasAdvancedLocalFilters.value) {
      galleryLoadMode.value = 'performance'
      modeReason.value = gallerySummary.value?.modeReason || '已启用性能优先模式'
      return
    }
    galleryLoadMode.value = 'standard'
    if (preferred === 'performance' && hasAdvancedLocalFilters.value) {
      modeReason.value = '当前筛选依赖完整结果，已切回标准模式'
      return
    }
    if (preferred === 'performance' && !supportsPagedGalleryView.value) {
      modeReason.value = '当前视图继续使用标准模式'
      return
    }
    modeReason.value = gallerySummary.value?.modeReason || '当前使用标准模式'
  }

const toggleRoot = (name) => {
    // 如果已经在该根目录，则切换到dashboard
    if (activeRoot.value === name) {
      activeRoot.value = 'dashboard'
      activeSub.value = ''
      activeChild.value = ''
    } else {
      activeRoot.value = name
      activeSub.value = ''
      activeChild.value = ''
    }
  }

  const currentImages = computed(() => {
    if (!activeRoot.value) return []
    const activeSourceImages = sourceImages.value

    if (activeRoot.value === 'favorites') {
      const imgs = activeSourceImages.filter((img) => favorites.value.has(normalizeFolderPath(img.relPath)))
      imgs.sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
      if (!activeSub.value) return imgs

      const groupId = activeSub.value.startsWith('favorite-group:')
        ? activeSub.value.replace('favorite-group:', '')
        : ''
      if (!groupId) return imgs

      const group = favoriteGroups.value.find((item) => item.id === groupId)
      if (!group) return imgs

      const groupPathSet = new Set((group.paths || []).map((path) => normalizeFolderPath(path)))
      return imgs.filter((img) => groupPathSet.has(normalizeFolderPath(img.relPath)))
    }

    const root = fileTree.value.find((r) => r.id === activeRoot.value)
    if (!root) return []

    const collectImages = (node) => {
      let acc = node.images ? [...node.images] : []
      const children = node.children || node.subs || []
      children.forEach((child) => {
        acc = acc.concat(collectImages(child))
      })
      return acc
    }

    if (!activeSub.value) {
      return collectImages(root)
    }

    const findNode = (nodes) => {
      for (const node of nodes) {
        if (node.id === activeSub.value) return node
        const children = node.children || node.subs || []
        if (children.length > 0) {
          const found = findNode(children)
          if (found) return found
        }
      }
      return null
    }

    const targetNode = findNode(root.subs || [])
    return targetNode ? collectImages(targetNode) : collectImages(root)
  })

  const buildPagedQuery = ({ page = currentPage.value, pageSize = itemsPerPage.value } = {}) => {
    const query = {
      sortBy: sortBy.value,
      sortOrder: sortOrder.value,
      page,
      pageSize,
      scopeRelPath: '',
      favoritesOnly: false,
      favoriteGroupId: '',
      searchQuery: searchQuery.value || '',
      activeTagId: activeTagFilter.value || '',
      activeModelFilter: activeModelFilter.value || '',
      activeLoraFilter: activeLoraFilter.value || '',
      activeDatePreset: activeDatePreset.value || 'all',
      activeDateStart: activeDateStart.value || '',
      activeDateEnd: activeDateEnd.value || '',
    }

    if (activeRoot.value === 'favorites') {
      query.favoritesOnly = true
      query.favoriteGroupId = activeSub.value.startsWith('favorite-group:')
        ? activeSub.value.replace('favorite-group:', '')
        : ''
      return query
    }

    query.scopeRelPath = selectedGalleryNode.value?.relPath || ''
    return query
  }

  const mapLoadedImage = (img) => ({
    ...img,
    path: buildImageDisplayPath(img.path, img.modTime, img.size),
    thumbPath: buildImageDisplayPath(img.thumbPath, img.modTime, img.size),
    previewPath: buildImageDisplayPath(img.previewPath, img.modTime, img.size),
    cardPath:
      galleryLoadMode.value === 'performance' &&
      performanceSettings.value.thumbPreferred &&
      buildImageDisplayPath(img.thumbPath, img.modTime, img.size)
        ? buildImageDisplayPath(img.thumbPath, img.modTime, img.size)
        : buildImageDisplayPath(img.path, img.modTime, img.size),
    loras: Array.isArray(img.loras) ? img.loras : [],
    isFavorite: favorites.value.has(normalizeFolderPath(img.relPath)),
  })

  const ensureStandardImagesReady = async () => {
    await ensureGalleryStandardImagesReady({ fetchImagesFn: fetchImages })
  }

  const removeImagesLocally = (relPaths) => {
    removeGalleryImagesLocally({
      relPaths,
      favoritesRef: favorites,
      imageTagsRef: imageTags,
      imageNotesRef: imageNotes,
    })
  }

  const fetchImagesPage = async ({ page = currentPage.value, append = false } = {}) => {
    await fetchGalleryImagesPage({
      page,
      append,
      buildPagedQuery,
      mapLoadedImage,
    })
  }

  const refreshCurrentGalleryView = async ({ syncSourceImages = false } = {}) => {
    loading.value = true
    await fetchGallerySummary()
    const preferIndexedSource =
      effectivePreferredMode.value === 'performance' && !hasAdvancedLocalFilters.value
    if (syncSourceImages) {
      if (preferIndexedSource) {
        await fetchImageIndex()
      } else {
        await fetchImages()
      }
    }
    refreshGalleryMode()
    if (galleryLoadMode.value === 'performance') {
      if (!syncSourceImages && images.value.length === 0 && indexedImages.value.length === 0) {
        await fetchImageIndex()
      }
      await fetchImagesPage({ page: currentPage.value || 1 })
      loading.value = false
      return
    }
    invalidatePagedRequests()
    pagedImages.value = []
    pagedTotal.value = 0
    pagedTotalPages.value = 1
    hasMorePagedImages.value = false
    if (!syncSourceImages) {
      await fetchImages()
    }
    loading.value = false
  }

  const finalImages = computed(() => {
    let imgs = currentImages.value

    if (imgs.length > 0 && hasActiveWorkbenchFilters.value) {
      imgs = imgs.filter((img) =>
        imageMatchesWorkbenchFilters(
          img,
          activeDatePreset.value,
          activeDateRange.value,
          activeModelFilter.value,
          activeLoraFilter.value,
        ),
      )
    }

    if (activeTagFilter.value && imgs.length > 0) {
      imgs = imgs.filter((img) =>
        imageTags.value[img.relPath]?.includes(activeTagFilter.value),
      )
    }

    const { dateRange, size, dimensions } = filters.value

    if (imgs.length > 0) {
      if (dateRange.start || dateRange.end) {
        imgs = imgs.filter((img) => {
          const imgDate = new Date(img.modTime)
          if (dateRange.start && imgDate < new Date(dateRange.start)) return false
          if (dateRange.end) {
            const endDate = new Date(dateRange.end)
            endDate.setHours(23, 59, 59, 999)
            if (imgDate > endDate) return false
          }
          return true
        })
      }

      if (size.min !== null || size.max !== null) {
        imgs = imgs.filter((img) => {
          const sizeMB = img.size / (1024 * 1024)
          if (size.min !== null && sizeMB < size.min) return false
          if (size.max !== null && sizeMB > size.max) return false
          return true
        })
      }

      if (dimensions.minW !== null || dimensions.minH !== null) {
        imgs = imgs.filter((img) => {
          if (!img.width && !img.height) return true
          if (dimensions.minW !== null && (img.width || 0) < dimensions.minW) return false
          if (dimensions.minH !== null && (img.height || 0) < dimensions.minH) return false
          return true
        })
      }
    }

    const normalizedQuery = normalizeSearchText(searchQuery.value)
    if (normalizedQuery) {
      const tagNameMap = new Map((tags.value || []).map((tag) => [tag.id, tag.name || '']))
      imgs = imgs.filter((img) => {
        const noteText = imageNotes.value?.[img.relPath] || ''
        const tagTexts = (imageTags.value?.[img.relPath] || [])
          .map((tagId) => tagNameMap.get(tagId) || '')
          .filter(Boolean)

        const searchParts = [
          img.name,
          img.relPath,
          img.prompt,
          img.model,
          ...(img.loras || []),
          img.searchText,
          noteText,
          ...tagTexts,
        ]

        return searchParts.some((part) => normalizeSearchText(part).includes(normalizedQuery))
      })
    }

    return imgs
  })

  const { stackedImages } = useImageStacks(finalImages, isStackingEnabled)
  const {
    paginatedImages,
    totalPages,
    setPage,
    prevPage,
    nextPage,
    setItemsPerPage,
    resetPage,
  } = useGalleryPagination({
    stackedImagesRef: stackedImages,
    galleryLoadModeRef: galleryLoadMode,
    fetchImagesPageFn: fetchImagesPage,
  })

  watch([activeRoot, activeSub, activeChild], () => {
    resetPage()
    refreshGalleryMode()
    if (galleryLoadMode.value === 'performance') {
      fetchImagesPage({ page: 1 })
    }
  })

  watch(searchQuery, (value) => {
    localStorage.setItem('gallerySearchQuery', value)
    resetPage()
    if (galleryLoadMode.value === 'performance') {
      fetchImagesPage({ page: 1 })
    }
  })

  watch(availableModels, (options) => {
    const syncedValue = syncGroupedFilterValue(activeModelFilter.value, options)
    if (activeModelFilter.value && !syncedValue) {
      activeModelFilter.value = ''
      return
    }
    if (syncedValue && activeModelFilter.value !== syncedValue) {
      activeModelFilter.value = syncedValue
    }
  }, { immediate: true })

  watch(availableLoras, (options) => {
    const syncedValue = syncGroupedFilterValue(activeLoraFilter.value, options)
    if (activeLoraFilter.value && !syncedValue) {
      activeLoraFilter.value = ''
      return
    }
    if (syncedValue && activeLoraFilter.value !== syncedValue) {
      activeLoraFilter.value = syncedValue
    }
  }, { immediate: true })

  watch(
    [activeDatePreset, activeDateStart, activeDateEnd, activeModelFilter, activeLoraFilter],
    () => {
      resetPage()
      fetchWorkbenchAggregate()
      if (galleryLoadMode.value === 'performance') {
        fetchImagesPage({ page: 1 })
      }
    },
  )

  watch(activeTagFilter, () => {
    resetPage()
    if (galleryLoadMode.value === 'performance') {
      fetchImagesPage({ page: 1 })
    }
  })

  watch(filters, async () => {
    refreshGalleryMode()
    if (galleryLoadMode.value === 'performance') {
      await fetchImagesPage({ page: 1 })
      return
    }
    await ensureStandardImagesReady()
  }, { deep: true })

  watch([effectivePreferredMode, supportsPagedGalleryView], async () => {
    refreshGalleryMode()
    if (galleryLoadMode.value === 'standard') {
      await ensureStandardImagesReady()
    }
  })

  const clearSearchQuery = () => {
    searchQuery.value = ''
  }

  const initAutoSelect = () => {
    if (fileTree.value.length === 0) return
    if (!isInitialized.value) {
      const isSpecialRoot = ['dashboard', 'profile', 'documentation', 'statistics', 'date-workbench']
        .includes(activeRoot.value)
      const rootExists = fileTree.value.find((r) => r.id === activeRoot.value)
      if (!isSpecialRoot && (!activeRoot.value || !rootExists)) {
        activeRoot.value = fileTree.value[0].id
      }

      const currentRoot = fileTree.value.find((r) => r.id === activeRoot.value)
      if (currentRoot && (currentRoot.subs || currentRoot.children)) {
        const children = currentRoot.subs || currentRoot.children
        if (children.length > 0 && activeSub.value) {
          const subExists = children.find((s) => s.id === activeSub.value)
          if (!subExists) {
            activeSub.value = ''
          }
        }
      }

      isInitialized.value = true
    }
  }

  const startPolling = () => {
    refreshCurrentGalleryView({ syncSourceImages: true }).then(initAutoSelect)
    return null
  }

  const handleDelete = async (img) => {
    const ok = await confirm(`确定将 ${img.name} 移至回收站吗？`)
    if (!ok) return

    const original = images.value
    const originalIndexed = indexedImages.value
    const originalPaged = pagedImages.value
    const originalPagedTotal = pagedTotal.value
    const originalPagedTotalPages = pagedTotalPages.value
    const originalHasMorePaged = hasMorePagedImages.value
    const originalFavorites = new Set(favorites.value)
    const originalTags = imageTags.value[img.relPath] ? [...imageTags.value[img.relPath]] : null
    const originalNote = imageNotes.value[img.relPath]
    const originalGallerySummary = { ...gallerySummary.value }
    removeImagesLocally([img.relPath])
    try {
      await App.DeleteImage(img.relPath)
      showToast('删除成功', 'success')
      if (favorites.value.has(img.relPath)) {
        favorites.value.delete(img.relPath)
        App.RemoveFavorite(img.relPath).catch(console.error)
      }
    } catch (err) {
      showToast('删除失败', 'error')
      images.value = original
      indexedImages.value = originalIndexed
      pagedImages.value = originalPaged
      pagedTotal.value = originalPagedTotal
      pagedTotalPages.value = originalPagedTotalPages
      hasMorePagedImages.value = originalHasMorePaged
      favorites.value = originalFavorites
      gallerySummary.value = originalGallerySummary
      if (originalTags) {
        imageTags.value[img.relPath] = originalTags
      }
      if (originalNote !== undefined) {
        imageNotes.value[img.relPath] = originalNote
      }
    }
  }

  const setSortBy = (newSortBy) => {
    if (sortBy.value === newSortBy) {
      sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
      localStorage.setItem('sortOrder', sortOrder.value)
    } else {
      sortBy.value = newSortBy
      sortOrder.value = 'desc'
      localStorage.setItem('sortBy', newSortBy)
      localStorage.setItem('sortOrder', 'desc')
    }
    refreshCurrentGalleryView({ syncSourceImages: galleryLoadMode.value !== 'performance' })
  }

  const setSortOrder = (order) => {
    sortOrder.value = order
    localStorage.setItem('sortOrder', order)
    refreshCurrentGalleryView({ syncSourceImages: galleryLoadMode.value !== 'performance' })
  }

  const openImageLocation = async (img) => {
    if (!img) return
    try {
      await App.OpenImageLocation(img.relPath)
    } catch (e) {
      console.error(e)
      showToast('无法打开文件位置', 'error')
    }
  }

  const totalVisibleImages = computed(() =>
    galleryLoadMode.value === 'performance' ? pagedTotal.value : finalImages.value.length,
  )

  const isPerformanceModeActive = computed(() => galleryLoadMode.value === 'performance')

  return {
    images,
    sourceImages,
    favorites,
    favoriteGroups,
    loading,
    activeRoot,
    activeSub,
    activeChild,
    fileTree,
    scopeImageCount: computed(() => currentImages.value.length),
    currentImages: finalImages,
    totalVisibleImages,
    fetchImages,
    fetchImageIndex,
    fetchImagesPage,
    fetchFavorites,
    loadPerformanceSettings,
    savePerformanceSettings,
    fetchGallerySummary,
    fetchDirectoryHealthSummary,
    fetchWorkbenchAggregate,
    refreshCurrentGalleryView,
    toggleFavorite,
    startPolling,
    toggleRoot,
    handleDelete,
    openImageLocation,
    tags,
    imageTags,
    activeTagFilter,
    filters,
    fetchTags,
    fetchImageTags,
    createTag,
    deleteTag,
    batchDeleteTags,
    updateTag,
    batchUpdateTags,
    addTagToImage,
    removeTagFromImage,
    toggleTagFilter,
    getTagCount,
    sortBy,
    sortOrder,
    searchQuery,
    setSortBy,
    setSortOrder,
    currentPage,
    itemsPerPage,
    paginatedImages,
    totalPages,
    setPage,
    prevPage,
    nextPage,
    setItemsPerPage,
    resetPage,
    customRoots,
    fetchCustomRoots,
    imageNotes,
    fetchImageNotes,
    isStackingEnabled,
    availableModels,
    availableLoras,
    workbenchFilteredImages,
    dateWorkbenchSummary,
    workbenchFilteredCount,
    activeDatePreset,
    activeDateStart,
    activeDateEnd,
    activeModelFilter,
    activeLoraFilter,
    activeDateLabel,
    hasActiveWorkbenchFilters,
    setActiveDatePreset,
    setActiveDateRange,
    clearDateFilter,
    setActiveModel,
    setActiveLora,
    clearWorkbenchFilters,
    clearSearchQuery,
    performanceSettings,
    gallerySummary,
    directoryHealthSummary,
    workbenchAggregate,
    galleryLoadMode,
    pagedImages,
    pagedTotal,
    pagedTotalPages,
    hasMorePagedImages,
    isPagedLoading,
    isPagedAppending,
    isPerformanceModeActive,
    modeReason,
    lastSuccessfulQuery,
    removeImagesLocally,
    toggleStacking: () => {
      isStackingEnabled.value = !isStackingEnabled.value
      localStorage.setItem('isStackingEnabled', isStackingEnabled.value)
    }
  }
}


