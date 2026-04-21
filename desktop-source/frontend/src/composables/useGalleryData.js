import { computed, ref } from 'vue'
import * as App from '@/api'
import { normalizeFolderPath, normalizePerformanceSettings } from './useGalleryHelpers'

const images = ref([])
const indexedImages = ref([])
const loading = ref(true)
const currentPage = ref(Number(localStorage.getItem('currentPage')) || 1)
const itemsPerPage = ref(Number(localStorage.getItem('itemsPerPage')) || 50)
const performanceSettings = ref({
  mode: 'auto',
  initialBatchSize: 60,
  pageSize: 50,
  thumbPreferred: true,
  backgroundVariantWarmup: true,
  metadataLazy: true,
})
const gallerySummary = ref({
  totalImages: 0,
  managedRootCount: 0,
  activeMode: 'standard',
  modeReason: '',
  thumbCacheCount: 0,
  thumbCacheBytes: 0,
  previewCacheCount: 0,
  previewCacheBytes: 0,
})
const directoryHealthSummary = ref({
  totalImages: 0,
  emptyFolderCount: 0,
  invalidTagReferenceCount: 0,
  invalidFavoriteReferenceCount: 0,
  thumbCacheCount: 0,
  thumbCacheBytes: 0,
  previewCacheCount: 0,
  previewCacheBytes: 0,
  issues: [],
})
const workbenchAggregate = ref({
  availableModels: [],
  availableLoras: [],
  summary: {
    total: 0,
    datedTotal: 0,
    today: 0,
    yesterday: 0,
    last7: 0,
    month: 0,
    recentDates: [],
  },
  filteredCount: 0,
})
const galleryLoadMode = ref('standard')
const pagedImages = ref([])
const pagedTotal = ref(0)
const pagedTotalPages = ref(1)
const hasMorePagedImages = ref(false)
const isPagedLoading = ref(false)
const isPagedAppending = ref(false)
const modeReason = ref('')
const lastSuccessfulQuery = ref(null)
let latestPagedRequestId = 0

const buildFavoritePathSet = (groups) =>
  new Set(
    (groups || []).flatMap((group) =>
      (group.paths || [])
        .map((path) => normalizeFolderPath(path))
        .filter(Boolean),
    ),
  )

export function useGalleryData() {
  const loadPerformanceSettings = async () => {
    try {
      const next = await App.GetGalleryPerformanceSettings()
      performanceSettings.value = normalizePerformanceSettings(next || {})
    } catch (e) {
      console.error('Failed to load performance settings:', e)
      performanceSettings.value = normalizePerformanceSettings()
    }
    itemsPerPage.value = performanceSettings.value.pageSize
  }

  const savePerformanceSettings = async (nextSettings) => {
    const saved = await App.SaveGalleryPerformanceSettings(normalizePerformanceSettings(nextSettings))
    performanceSettings.value = normalizePerformanceSettings(saved || {})
    itemsPerPage.value = performanceSettings.value.pageSize
    localStorage.setItem('itemsPerPage', itemsPerPage.value)
    return performanceSettings.value
  }

  const fetchGallerySummary = async () => {
    try {
      gallerySummary.value = await App.GetImageGallerySummary()
    } catch (e) {
      console.error('Failed to fetch gallery summary:', e)
    }
    return gallerySummary.value
  }

  const fetchDirectoryHealthSummary = async () => {
    try {
      directoryHealthSummary.value = await App.GetDirectoryHealthSummary()
    } catch (e) {
      console.error('Failed to fetch directory health summary:', e)
    }
    return directoryHealthSummary.value
  }

  const fetchWorkbenchAggregate = async (filters) => {
    try {
      workbenchAggregate.value = await App.GetWorkbenchAggregate({
        activeDatePreset: filters?.activeDatePreset || 'all',
        activeDateStart: filters?.activeDateStart || '',
        activeDateEnd: filters?.activeDateEnd || '',
        activeModelFilter: filters?.activeModelFilter || '',
        activeLoraFilter: filters?.activeLoraFilter || '',
      })
    } catch (e) {
      console.error('Failed to fetch workbench aggregate:', e)
    }
    return workbenchAggregate.value
  }

  const fetchImages = async ({ sortBy, sortOrder, favoriteGroupsRef, favoritesRef, mapLoadedImage }) => {
    try {
      const [imgs, groups] = await Promise.all([
        App.GetImages(sortBy, sortOrder),
        App.GetFavoriteGroups(),
      ])

      favoriteGroupsRef.value = groups || []
      favoritesRef.value = buildFavoritePathSet(groups)
      images.value = (imgs || []).map(mapLoadedImage)
    } catch (err) {
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  const fetchImageIndex = async ({ sortBy, sortOrder, favoriteGroupsRef, favoritesRef, mapLoadedImage }) => {
    try {
      const [imgs, groups] = await Promise.all([
        App.GetImagesIndex(sortBy, sortOrder),
        App.GetFavoriteGroups(),
      ])

      images.value = []
      favoriteGroupsRef.value = groups || []
      favoritesRef.value = buildFavoritePathSet(groups)
      indexedImages.value = (imgs || []).map(mapLoadedImage)
    } catch (err) {
      console.error('Failed to fetch image index:', err)
    } finally {
      loading.value = false
    }
  }

  const ensureStandardImagesReady = async ({ fetchImagesFn }) => {
    latestPagedRequestId += 1
    if (images.value.length > 0) return
    await fetchImagesFn()
  }

  const removeImagesLocally = ({
    relPaths,
    favoritesRef,
    imageTagsRef,
    imageNotesRef,
  }) => {
    const normalizedPaths = (relPaths || [])
      .map((path) => normalizeFolderPath(path))
      .filter(Boolean)
    if (normalizedPaths.length === 0) return

    const pathSet = new Set(normalizedPaths)
    const filterImages = (list) => (list || []).filter((img) => !pathSet.has(normalizeFolderPath(img.relPath)))
    const removedFromPaged = (pagedImages.value || []).filter((img) => pathSet.has(normalizeFolderPath(img.relPath))).length

    images.value = filterImages(images.value)
    indexedImages.value = filterImages(indexedImages.value)
    pagedImages.value = filterImages(pagedImages.value)

    favoritesRef.value = new Set(
      Array.from(favoritesRef.value).filter((path) => !pathSet.has(normalizeFolderPath(path))),
    )

    normalizedPaths.forEach((relPath) => {
      delete imageTagsRef.value[relPath]
      delete imageNotesRef.value[relPath]
    })

    if (removedFromPaged > 0) {
      pagedTotal.value = Math.max(0, pagedTotal.value - removedFromPaged)
      pagedTotalPages.value = pagedTotal.value > 0 ? Math.ceil(pagedTotal.value / itemsPerPage.value) : 0
      if (currentPage.value > Math.max(pagedTotalPages.value, 1)) {
        currentPage.value = Math.max(pagedTotalPages.value, 1)
        localStorage.setItem('currentPage', currentPage.value)
      }
      hasMorePagedImages.value = currentPage.value < Math.max(pagedTotalPages.value, 1)
    }

    if (gallerySummary.value?.totalImages) {
      gallerySummary.value = {
        ...gallerySummary.value,
        totalImages: Math.max(0, gallerySummary.value.totalImages - normalizedPaths.length),
      }
    }
  }

  const fetchImagesPage = async ({
    page,
    append = false,
    buildPagedQuery,
    mapLoadedImage,
  }) => {
    const requestId = ++latestPagedRequestId
    const targetPage = page || currentPage.value
    const query = buildPagedQuery({ page: targetPage, pageSize: itemsPerPage.value })
    if (append) {
      isPagedAppending.value = true
    } else {
      isPagedLoading.value = true
      loading.value = true
    }

    try {
      const result = await App.GetImagesPage(query)
      if (requestId !== latestPagedRequestId) return
      const mappedItems = (result?.items || []).map(mapLoadedImage)
      pagedImages.value = append ? [...pagedImages.value, ...mappedItems] : mappedItems
      pagedTotal.value = Number(result?.total) || 0
      pagedTotalPages.value = Number(result?.totalPages) || (pagedTotal.value > 0 ? 1 : 0)
      hasMorePagedImages.value = !!result?.hasMore
      currentPage.value = Number(result?.page) || targetPage || 1
      localStorage.setItem('currentPage', currentPage.value)
      lastSuccessfulQuery.value = query
      if (result?.modeReason) {
        modeReason.value = result.modeReason
      }
    } catch (err) {
      if (requestId !== latestPagedRequestId) return
      console.error('Failed to fetch paged images:', err)
      if (!append) {
        pagedImages.value = []
        pagedTotal.value = 0
        pagedTotalPages.value = 0
        hasMorePagedImages.value = false
      }
    } finally {
      if (requestId !== latestPagedRequestId) return
      isPagedLoading.value = false
      isPagedAppending.value = false
      if (!append) {
        loading.value = false
      }
    }
  }

  const invalidatePagedRequests = () => {
    latestPagedRequestId += 1
  }

  const useGalleryPagination = ({ stackedImagesRef, galleryLoadModeRef, fetchImagesPageFn }) => {
    const paginatedImages = computed(() => {
      if (galleryLoadModeRef.value === 'performance') {
        return pagedImages.value
      }
      const startIndex = (currentPage.value - 1) * itemsPerPage.value
      const endIndex = startIndex + itemsPerPage.value
      return stackedImagesRef.value.slice(startIndex, endIndex)
    })

    const totalPages = computed(() => {
      if (galleryLoadModeRef.value === 'performance') {
        return pagedTotalPages.value || 1
      }
      return Math.max(1, Math.ceil(stackedImagesRef.value.length / itemsPerPage.value))
    })

    const setPage = (page) => {
      if (galleryLoadModeRef.value === 'performance') {
        if (page < 1) return
        currentPage.value = page
        localStorage.setItem('currentPage', page)
        fetchImagesPageFn({ page })
        return
      }
      if (page < 1 || page > totalPages.value) return
      currentPage.value = page
      localStorage.setItem('currentPage', page)
    }

    const prevPage = () => setPage(currentPage.value - 1)
    const nextPage = () => setPage(currentPage.value + 1)

    const setItemsPerPage = (count) => {
      itemsPerPage.value = count
      localStorage.setItem('itemsPerPage', count)
      if (galleryLoadModeRef.value === 'performance') {
        currentPage.value = 1
        fetchImagesPageFn({ page: 1 })
        return
      }
      setPage(1)
    }

    const resetPage = () => {
      if (galleryLoadModeRef.value === 'performance') {
        currentPage.value = 1
        localStorage.setItem('currentPage', 1)
        return
      }
      setPage(1)
    }

    return {
      paginatedImages,
      totalPages,
      setPage,
      prevPage,
      nextPage,
      setItemsPerPage,
      resetPage,
    }
  }

  return {
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
    loadPerformanceSettings,
    savePerformanceSettings,
    fetchGallerySummary,
    fetchDirectoryHealthSummary,
    fetchWorkbenchAggregate,
    fetchImages,
    fetchImageIndex,
    ensureStandardImagesReady,
    removeImagesLocally,
    fetchImagesPage,
    invalidatePagedRequests,
    useGalleryPagination,
  }
}
