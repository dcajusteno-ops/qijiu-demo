import { ref, computed, watch } from 'vue'
import * as App from '@/api'
import { useImageStacks } from './useImageStacks'
import {
  buildDateCountMap,
  extractDateFolder,
  formatDateKey,
  getDatePresetLabel,
  matchesDatePreset,
} from '@/lib/dateWorkbench'

const images = ref([])
const indexedImages = ref([])
const loading = ref(true)
const activeRoot = ref('dashboard')
const activeSub = ref('')
const activeChild = ref('')
const isInitialized = ref(false)

const tags = ref([])
const imageTags = ref({})
const activeTagFilter = ref(null)

const customRoots = ref([])

const imageNotes = ref({})

const filters = ref({
  dateRange: { start: null, end: null },
  size: { min: null, max: null },
  dimensions: { minW: null, minH: null },
})
const searchQuery = ref(localStorage.getItem('gallerySearchQuery') || '')
const activeDatePreset = ref('all')
const activeDateStart = ref('')
const activeDateEnd = ref('')
const activeModelFilter = ref('')
const activeLoraFilter = ref('')

const sortBy = ref(localStorage.getItem('sortBy') || 'time')
const sortOrder = ref(localStorage.getItem('sortOrder') || 'desc')
const isStackingEnabled = ref(localStorage.getItem('isStackingEnabled') !== 'false') // default true

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

const favorites = ref(new Set())
const favoriteGroups = ref([])

const normalizeFolderPath = (path) => (path || '')
  .replace(/\\/g, '/')
  .replace(/^\/+|\/+$/g, '')

const dateFolderPattern = /^\d{4}-\d{2}-\d{2}$/
const getDateSegment = (path) => {
  const parts = normalizeFolderPath(path).split('/').filter(Boolean)
  return parts.find((part) => dateFolderPattern.test(part)) || ''
}

const normalizeSearchText = (value) => String(value ?? '').trim().toLowerCase()
const normalizeFilterValue = (value) => normalizeSearchText(value).replace(/\s+/g, ' ')
const stripPathSegments = (value) => String(value ?? '').split(/[\\/]/).pop() || ''
const buildImageDisplayPath = (pathValue, modTime, size) => {
  const normalizedPath = String(pathValue || '').trim()
  if (!normalizedPath) return ''

  const version = [String(modTime || '').trim(), String(size ?? '').trim()]
    .filter(Boolean)
    .join('-')

  if (!version) return normalizedPath

  const separator = normalizedPath.includes('?') ? '&' : '?'
  return `${normalizedPath}${separator}v=${encodeURIComponent(version)}`
}
const stripModelExtension = (value) =>
  String(value ?? '').replace(/\.(safetensors|ckpt|pt|pth|bin)$/i, '')
const prettifyAssetLabel = (value) =>
  stripModelExtension(stripPathSegments(value)).replace(/[_]+/g, ' ').trim()
const normalizeAssetKey = (value) =>
  normalizeFilterValue(prettifyAssetLabel(value)).replace(/[-]+/g, ' ')

const buildGroupedFilterOptions = (values = []) => {
  const grouped = new Map()

  values.forEach((rawValue) => {
    const raw = String(rawValue || '').trim()
    if (!raw) return

    const key = normalizeAssetKey(raw)
    if (!key) return

    const label = prettifyAssetLabel(raw) || raw
    if (!grouped.has(key)) {
      grouped.set(key, {
        value: key,
        label,
        count: 0,
        aliases: new Set(),
      })
    }

    const entry = grouped.get(key)
    entry.count += 1
    entry.aliases.add(raw)

    if (label.length < entry.label.length) {
      entry.label = label
    }
  })

  return Array.from(grouped.values())
    .sort((a, b) => {
      const diff = b.count - a.count
      if (diff !== 0) return diff
      return a.label.localeCompare(b.label)
    })
    .map((item) => ({
      value: item.value,
      label: item.label,
      count: item.count,
      aliases: Array.from(item.aliases),
    }))
}

const syncGroupedFilterValue = (currentValue, options = []) => {
  const normalizedCurrent = normalizeAssetKey(currentValue)
  if (!normalizedCurrent) return ''

  const matched = (options || []).find((option) => option.value === normalizedCurrent)
  return matched ? matched.value : ''
}

const getImageDateKey = (img) => {
  const folderDate = extractDateFolder(img?.relPath)
  if (folderDate) return folderDate

  const modTime = img?.modTime ? new Date(img.modTime) : null
  if (modTime && !Number.isNaN(modTime.getTime())) {
    return formatDateKey(modTime)
  }

  return ''
}

const imageMatchesWorkbenchFilters = (img, datePreset, customRange, modelFilter, loraFilter) => {
  if (datePreset && datePreset !== 'all') {
    const dateKey = getImageDateKey(img)
    if (!dateKey || !matchesDatePreset(dateKey, datePreset, customRange)) {
      return false
    }
  }

  if (modelFilter) {
    const imageModelKey = normalizeAssetKey(img?.model)
    const selectedModelKey = normalizeAssetKey(modelFilter)
    if (!imageModelKey || !selectedModelKey || imageModelKey !== selectedModelKey) {
      return false
    }
  }

  if (loraFilter) {
    const loras = Array.isArray(img?.loras) ? img.loras : []
    const target = normalizeAssetKey(loraFilter)
    if (!target || !loras.some((item) => normalizeAssetKey(item) === target)) {
      return false
    }
  }

  return true
}

const getFavoritePathSet = (groups) => {
  const set = new Set()
  ;(groups || []).forEach((group) => {
    ;(group.paths || []).forEach((path) => {
      const normalized = normalizeFolderPath(path)
      if (normalized) set.add(normalized)
    })
  })
  return set
}

const normalizePerformanceSettings = (settings = {}) => ({
  mode: ['auto', 'standard', 'performance'].includes(settings?.mode) ? settings.mode : 'auto',
  initialBatchSize: Math.min(Math.max(Number(settings?.initialBatchSize) || 60, 20), 500),
  pageSize: Math.min(Math.max(Number(settings?.pageSize) || 50, 20), 500),
  thumbPreferred: settings?.thumbPreferred !== false,
  backgroundVariantWarmup: settings?.backgroundVariantWarmup !== false,
  metadataLazy: settings?.metadataLazy !== false,
})

export function useImages(showToast = () => {}, confirm = async () => false) {
  const fetchCustomRoots = async () => {
    try {
      const roots = await App.GetCustomRoots()
      customRoots.value = roots || []
    } catch (e) {
      console.error('Failed to fetch custom roots:', e)
    }
  }

  const fetchFavorites = async () => {
    try {
      const groups = await App.GetFavoriteGroups()
      favoriteGroups.value = groups || []
      favorites.value = getFavoritePathSet(groups)
    } catch (e) {
      console.error(e)
    }
  }

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

  const fetchWorkbenchAggregate = async () => {
    try {
      workbenchAggregate.value = await App.GetWorkbenchAggregate({
        activeDatePreset: activeDatePreset.value || 'all',
        activeDateStart: activeDateStart.value || '',
        activeDateEnd: activeDateEnd.value || '',
        activeModelFilter: activeModelFilter.value || '',
        activeLoraFilter: activeLoraFilter.value || '',
      })
    } catch (e) {
      console.error('Failed to fetch workbench aggregate:', e)
    }
    return workbenchAggregate.value
  }

  const toggleFavorite = async (img) => {
    if (!img) return

    const path = normalizeFolderPath(img.relPath)
    const isFav = favorites.value.has(path)
    const activeFavoriteGroupId =
      activeRoot.value === 'favorites' && activeSub.value.startsWith('favorite-group:')
        ? activeSub.value.replace('favorite-group:', '')
        : ''

    try {
      if (isFav) {
        if (activeFavoriteGroupId) {
          await App.RemoveImageFromFavoriteGroup(img.relPath, activeFavoriteGroupId)
        } else {
          await App.RemoveFavorite(img.relPath)
        }
      } else {
        await App.AddImageToFavoriteGroup(img.relPath, activeFavoriteGroupId || 'default')
      }
      await fetchFavorites()
      img.isFavorite = favorites.value.has(normalizeFolderPath(img.relPath))
    } catch (e) {
      console.error(e)
      showToast('操作失败', 'error')
    }
  }

  const fetchImages = async () => {
    try {
      const [imgs, groups] = await Promise.all([
        App.GetImages(sortBy.value, sortOrder.value),
        App.GetFavoriteGroups(),
      ])

      favoriteGroups.value = groups || []
      favorites.value = getFavoritePathSet(groups)
      images.value = (imgs || []).map(mapLoadedImage)
    } catch (err) {
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  const fetchImageIndex = async () => {
    try {
      const [imgs, groups] = await Promise.all([
        App.GetImagesIndex(sortBy.value, sortOrder.value),
        App.GetFavoriteGroups(),
      ])

      images.value = []
      favoriteGroups.value = groups || []
      favorites.value = getFavoritePathSet(groups)
      indexedImages.value = (imgs || []).map(mapLoadedImage)
    } catch (err) {
      console.error('Failed to fetch image index:', err)
    } finally {
      loading.value = false
    }
  }

  const sourceImages = computed(() => (images.value.length > 0 ? images.value : indexedImages.value))

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
    latestPagedRequestId += 1
    if (images.value.length > 0) return
    await fetchImages()
  }

  const removeImagesLocally = (relPaths) => {
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

    favorites.value = new Set(
      Array.from(favorites.value).filter((path) => !pathSet.has(normalizeFolderPath(path))),
    )

    normalizedPaths.forEach((relPath) => {
      delete imageTags.value[relPath]
      delete imageNotes.value[relPath]
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

  const fetchImagesPage = async ({ page = currentPage.value, append = false } = {}) => {
    const requestId = ++latestPagedRequestId
    const query = buildPagedQuery({ page, pageSize: itemsPerPage.value })
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
      currentPage.value = Number(result?.page) || page || 1
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
    latestPagedRequestId += 1
    pagedImages.value = []
    pagedTotal.value = 0
    pagedTotalPages.value = 1
    hasMorePagedImages.value = false
    if (!syncSourceImages) {
      await fetchImages()
    }
    loading.value = false
  }

  const availableModels = computed(() => {
    if ((workbenchAggregate.value?.availableModels || []).length > 0) {
      return workbenchAggregate.value.availableModels
    }
    return buildGroupedFilterOptions(images.value.map((img) => img?.model || ''))
  })

  const availableLoras = computed(() => {
    if ((workbenchAggregate.value?.availableLoras || []).length > 0) {
      return workbenchAggregate.value.availableLoras
    }
    const loraValues = []
    images.value.forEach((img) => {
      ;(img?.loras || []).forEach((lora) => {
        loraValues.push(lora)
      })
    })
    return buildGroupedFilterOptions(loraValues)
  })

  const activeDateRange = computed(() => ({
    start: activeDateStart.value || '',
    end: activeDateEnd.value || '',
  }))

  const activeDateLabel = computed(() =>
    getDatePresetLabel(activeDatePreset.value, activeDateRange.value),
  )

  const hasActiveWorkbenchFilters = computed(() =>
    activeDatePreset.value !== 'all' || !!activeModelFilter.value || !!activeLoraFilter.value,
  )

  const workbenchFilteredImages = computed(() =>
    images.value.filter((img) =>
      imageMatchesWorkbenchFilters(
        img,
        activeDatePreset.value,
        activeDateRange.value,
        activeModelFilter.value,
        activeLoraFilter.value,
      ),
    ),
  )

  const fallbackDateWorkbenchSummary = computed(() => {
    const dateCountMap = buildDateCountMap(images.value)
    const datedImages = images.value.filter((img) => getImageDateKey(img))
    const countWithPreset = (preset) =>
      datedImages.filter((img) =>
        imageMatchesWorkbenchFilters(
          img,
          preset,
          null,
          activeModelFilter.value,
          activeLoraFilter.value,
        ),
      ).length

    const recentDates = Array.from(dateCountMap.entries())
      .sort((a, b) => b[0].localeCompare(a[0]))
      .map(([date, count]) => ({ date, count }))

    return {
      total: workbenchFilteredImages.value.length,
      datedTotal: datedImages.length,
      today: countWithPreset('today'),
      yesterday: countWithPreset('yesterday'),
      last7: countWithPreset('last7'),
      month: countWithPreset('month'),
      recentDates,
    }
  })
  const dateWorkbenchSummary = computed(() => {
    const summary = workbenchAggregate.value?.summary
    if (summary && (summary.recentDates?.length || workbenchAggregate.value?.filteredCount || summary.datedTotal)) {
      return summary
    }
    return fallbackDateWorkbenchSummary.value
  })
  const workbenchFilteredCount = computed(() =>
    Number.isFinite(Number(workbenchAggregate.value?.filteredCount))
      ? Number(workbenchAggregate.value.filteredCount)
      : workbenchFilteredImages.value.length,
  )

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

  const paginatedImages = computed(() => {
    if (galleryLoadMode.value === 'performance') {
      return pagedImages.value
    }
    const startIndex = (currentPage.value - 1) * itemsPerPage.value
    const endIndex = startIndex + itemsPerPage.value
    return stackedImages.value.slice(startIndex, endIndex)
  })

  const totalPages = computed(() => {
    if (galleryLoadMode.value === 'performance') {
      return pagedTotalPages.value || 1
    }
    return Math.max(1, Math.ceil(stackedImages.value.length / itemsPerPage.value))
  })

  const setPage = (page) => {
    if (galleryLoadMode.value === 'performance') {
      if (page < 1) return
      currentPage.value = page
      localStorage.setItem('currentPage', page)
      fetchImagesPage({ page })
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
    if (galleryLoadMode.value === 'performance') {
      currentPage.value = 1
      fetchImagesPage({ page: 1 })
      return
    }
    setPage(1)
  }

  const resetPage = () => {
    if (galleryLoadMode.value === 'performance') {
      currentPage.value = 1
      localStorage.setItem('currentPage', 1)
      return
    }
    setPage(1)
  }

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

  const setActiveDatePreset = (preset) => {
    activeDatePreset.value = preset || 'all'
    if (activeDatePreset.value !== 'custom') {
      activeDateStart.value = ''
      activeDateEnd.value = ''
    }
  }

  const setActiveDateRange = ({ start = '', end = '' } = {}) => {
    activeDateStart.value = start || ''
    activeDateEnd.value = end || ''
    activeDatePreset.value = activeDateStart.value || activeDateEnd.value ? 'custom' : 'all'
  }

  const clearDateFilter = () => {
    activeDatePreset.value = 'all'
    activeDateStart.value = ''
    activeDateEnd.value = ''
  }

  const setActiveModel = (value) => {
    activeModelFilter.value = value || ''
  }

  const setActiveLora = (value) => {
    activeLoraFilter.value = value || ''
  }

  const clearWorkbenchFilters = () => {
    clearDateFilter()
    activeModelFilter.value = ''
    activeLoraFilter.value = ''
  }

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

  const fetchTags = async () => {
    try {
      const tgs = await App.GetTags()
      tags.value = tgs || []
    } catch (e) {
      console.error('Failed to fetch tags:', e)
    }
  }

  const fetchImageTags = async () => {
    try {
      const imgsTags = await App.GetImageTags()
      imageTags.value = imgsTags || {}
    } catch (e) {
      console.error('Failed to fetch image tags:', e)
    }
  }

  const fetchImageNotes = async () => {
    try {
      const notes = await App.GetImageNotes()
      imageNotes.value = notes || {}
    } catch (e) {
      console.error('Failed to fetch image notes:', e)
    }
  }

  const createTag = async (name, color, category = '') => {
    try {
      const newTag = await App.CreateTag(name, color, category)
      tags.value.push(newTag)
      showToast('标签已创建', 'success')
      return newTag
    } catch (e) {
      console.error(e)
      showToast('创建失败', 'error')
      return null
    }
  }

  const deleteTag = async (tagId) => {
    const ok = await confirm('确定要删除该标签吗？此操作将同时移除所有图片的该标签。')
    if (!ok) return

    try {
      await App.DeleteTag(tagId)
      tags.value = tags.value.filter((t) => t.id !== tagId)

      for (const relPath in imageTags.value) {
        imageTags.value[relPath] = imageTags.value[relPath].filter((id) => id !== tagId)
      }

      if (activeTagFilter.value === tagId) {
        activeTagFilter.value = null
      }

      showToast('标签已删除', 'success')
    } catch (e) {
      console.error(e)
      showToast('删除失败', 'error')
    }
  }

  const batchDeleteTags = async (tagIds) => {
    if (!tagIds || tagIds.length === 0) return

    const ok = await confirm(`确定要删除选中的 ${tagIds.length} 个标签吗？此操作将同时移除所有图片的这些标签。`)
    if (!ok) return

    try {
      await App.BatchDeleteTags(tagIds)

      tags.value = tags.value.filter((t) => !tagIds.includes(t.id))
      for (const relPath in imageTags.value) {
        imageTags.value[relPath] = imageTags.value[relPath].filter((id) => !tagIds.includes(id))
      }
      if (tagIds.includes(activeTagFilter.value)) {
        activeTagFilter.value = null
      }

      showToast(`成功删除 ${tagIds.length} 个标签`, 'success')
    } catch (e) {
      console.error(e)
      showToast('删除失败', 'error')
    }
  }

  const batchUpdateTags = async (tagIds, data) => {
    if (!tagIds || tagIds.length === 0) return

    let successCount = 0
    let failCount = 0

    const promises = tagIds.map(async (tagId) => {
      try {
        await App.UpdateTag(tagId, data.name || null, data.color || null, data.category || null)
        const tag = tags.value.find((t) => t.id === tagId)
        if (tag) {
          if (data.name !== undefined) tag.name = data.name
          if (data.color !== undefined) tag.color = data.color
          if (data.category !== undefined) tag.category = data.category
        }
        successCount++
      } catch (e) {
        failCount++
      }
    })

    await Promise.all(promises)

    if (failCount === 0) {
      showToast(`成功更新 ${successCount} 个标签`, 'success')
    } else {
      showToast(`更新完成：${successCount} 成功，${failCount} 失败`, 'error')
    }
  }

  const updateTag = async (tagId, data) => {
    try {
      await App.UpdateTag(tagId, data.name || null, data.color || null, data.category || null)

      const tag = tags.value.find((t) => t.id === tagId)
      if (tag) {
        if (data.name !== undefined) tag.name = data.name
        if (data.color !== undefined) tag.color = data.color
        if (data.category !== undefined) tag.category = data.category
      }

      showToast('标签已更新', 'success')
      return true
    } catch (e) {
      console.error(e)
      showToast('更新失败', 'error')
      return false
    }
  }

  const addTagToImage = async (img, tagId) => {
    const relPath = img.relPath
    if (!imageTags.value[relPath]) {
      imageTags.value[relPath] = []
    }
    if (!imageTags.value[relPath].includes(tagId)) {
      imageTags.value[relPath].push(tagId)
    }

    try {
      await App.AddTagToImage(relPath, tagId)
    } catch (e) {
      console.error(e)
      imageTags.value[relPath] = imageTags.value[relPath].filter((id) => id !== tagId)
      showToast('添加标签失败', 'error')
    }
  }

  const removeTagFromImage = async (img, tagId) => {
    const relPath = img.relPath
    const originalTags = imageTags.value[relPath] || []
    if (imageTags.value[relPath]) {
      imageTags.value[relPath] = imageTags.value[relPath].filter((id) => id !== tagId)
    }

    try {
      await App.RemoveTagFromImage(relPath, tagId)
    } catch (e) {
      console.error(e)
      imageTags.value[relPath] = originalTags
      showToast('移除标签失败', 'error')
    }
  }

  const toggleTagFilter = (tagId) => {
    if (activeTagFilter.value === tagId) {
      activeTagFilter.value = null
    } else {
      activeTagFilter.value = tagId
    }
  }

  const getTagCount = (tagId) => {
    let count = 0
    for (const relPath in imageTags.value) {
      if (imageTags.value[relPath]?.includes(tagId)) {
        count++
      }
    }
    return count
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


