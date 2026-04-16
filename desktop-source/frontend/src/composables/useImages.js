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
const activeDatePreset = ref(localStorage.getItem('activeDatePreset') || 'all')
const activeDateValue = ref(localStorage.getItem('activeDateValue') || '')
const activeModelFilter = ref(localStorage.getItem('activeModelFilter') || '')
const activeLoraFilter = ref(localStorage.getItem('activeLoraFilter') || '')

const sortBy = ref(localStorage.getItem('sortBy') || 'time')
const sortOrder = ref(localStorage.getItem('sortOrder') || 'desc')
const isStackingEnabled = ref(localStorage.getItem('isStackingEnabled') !== 'false') // default true

const currentPage = ref(Number(localStorage.getItem('currentPage')) || 1)
const itemsPerPage = ref(Number(localStorage.getItem('itemsPerPage')) || 50)

const favorites = ref(new Set())
const favoriteGroups = ref([])

const normalizeFolderPath = (path) => (path || '')
  .replace(/\\/g, '/')
  .replace(/^\/+|\/+$/g, '')

const normalizeSearchText = (value) => String(value ?? '').trim().toLowerCase()
const normalizeFilterValue = (value) => normalizeSearchText(value).replace(/\s+/g, ' ')
const stripPathSegments = (value) => String(value ?? '').split(/[\\/]/).pop() || ''
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

const imageMatchesWorkbenchFilters = (img, datePreset, customDate, modelFilter, loraFilter) => {
  if (datePreset && datePreset !== 'all') {
    const dateKey = getImageDateKey(img)
    if (!dateKey || !matchesDatePreset(dateKey, datePreset, customDate)) {
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
      if (path) set.add(path)
    })
  })
  return set
}

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

  const toggleFavorite = async (img) => {
    if (!img) return

    const path = img.relPath
    const isFav = favorites.value.has(path)
    const activeFavoriteGroupId =
      activeRoot.value === 'favorites' && activeSub.value.startsWith('favorite-group:')
        ? activeSub.value.replace('favorite-group:', '')
        : ''

    try {
      if (isFav) {
        if (activeFavoriteGroupId) {
          await App.RemoveImageFromFavoriteGroup(path, activeFavoriteGroupId)
        } else {
          await App.RemoveFavorite(path)
        }
      } else {
        await App.AddImageToFavoriteGroup(path, activeFavoriteGroupId || 'default')
      }
      await fetchFavorites()
      img.isFavorite = favorites.value.has(path)
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
      images.value = (imgs || []).map((img) => ({
        ...img,
        loras: Array.isArray(img.loras) ? img.loras : [],
      }))
    } catch (err) {
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  const fileTree = computed(() => {
    const getOrCreate = (nodes, name, parentId, relPath = '') => {
      let node = nodes.find((n) => n.name === name)
      if (!node) {
        node = {
          name,
          id: parentId ? `${parentId}/${name}` : name,
          relPath,
          children: [],
          images: [],
        }
        nodes.push(node)
      } else if (relPath && !node.relPath) {
        node.relPath = relPath
      }
      return node
    }

    const addLeaf = (parentChildren, leafNode) => {
      const existing = parentChildren.find((n) => n.id === leafNode.id)
      if (existing) {
        existing.images = [...existing.images, ...leafNode.images]
        existing.lastMod = Math.max(existing.lastMod || 0, leafNode.lastMod || 0)
      } else {
        parentChildren.push(leafNode)
      }
    }

    const matchDate = (str) => {
      let m = str.match(/(\d{4})-(\d{2})-(\d{2})/)
      if (m) return { year: m[1], month: m[2], day: m[3], full: m[0] }

      m = str.match(/(?:^|\D)(\d{4})(\d{2})(\d{2})(?:\D|$)/)
      if (m) {
        const y = parseInt(m[1], 10)
        const mo = parseInt(m[2], 10)
        const d = parseInt(m[3], 10)
        if (y > 2000 && y < 2100 && mo >= 1 && mo <= 12 && d >= 1 && d <= 31) {
          return {
            year: m[1],
            month: m[2],
            day: m[3],
            full: `${y}-${mo.toString().padStart(2, '0')}-${d.toString().padStart(2, '0')}`,
          }
        }
      }
      return null
    }

    const rootNodes = [
      { name: '收藏夹', id: 'favorites', displayName: '收藏夹', relPath: '', children: [], images: [], type: 'root', icon: 'Heart', count: favorites.value.size },
      { name: 'output', id: 'output', displayName: 'output', relPath: '', children: [], images: [], type: 'root', icon: 'FolderOpen' },
      { name: '日期归档', id: '日期归档', displayName: '日期归档', relPath: '', children: [], images: [], type: 'root', icon: 'Calendar' },
      { name: 'OXYZ测试图片', id: 'OXYZ测试图片', displayName: 'OXYZ测试图片', relPath: '', children: [], images: [], type: 'root', icon: 'FlaskConical' },
      {
        name: '修复',
        id: '修复',
        displayName: '修复',
        relPath: '',
        children: [
          { name: '手动', id: '修复/手动', relPath: '修复/手动', children: [], images: [] },
          { name: '自动', id: '修复/自动', relPath: '修复/自动', children: [], images: [] },
        ],
        images: [],
        type: 'root',
        icon: 'Wrench',
      },
    ]

    const otherNodes = []

    let imagesToUse = images.value
    if (activeTagFilter.value) {
      imagesToUse = images.value.filter((img) =>
        imageTags.value[img.relPath]?.includes(activeTagFilter.value),
      )
    }

    const customRootPaths = customRoots.value
      .map((root) => normalizeFolderPath(root.path))
      .filter(Boolean)

    const isManagedByCustomRoot = (folderName) => {
      if (!folderName || folderName === 'root') return false

      const normalizedName = normalizeFolderPath(folderName)
      return customRootPaths.some((rootPath) =>
        normalizedName === rootPath || normalizedName.startsWith(`${rootPath}/`),
      )
    }

    const folderMap = {}
    imagesToUse.forEach((img) => {
      const normalizedRelPath = normalizeFolderPath(img.relPath)
      img.isFavorite = favorites.value.has(img.relPath)

      const parts = normalizedRelPath ? normalizedRelPath.split('/') : []
      let folderName = 'root'
      if (parts.length > 1) {
        folderName = parts.slice(0, parts.length - 1).join('/')
      }

      if (!folderMap[folderName]) {
        folderMap[folderName] = { name: folderName, images: [] }
      }
      folderMap[folderName].images.push(img)
    })

    Object.values(folderMap).forEach((folder) => {
      if (folder.images.length === 0) return

      const name = normalizeFolderPath(folder.name)
      const dateInfo = matchDate(name)

      if (isManagedByCustomRoot(name)) return

      if (name === 'root') {
        const outputRoot = rootNodes.find((r) => r.name === 'output')
        if (outputRoot) {
          outputRoot.images = [...(outputRoot.images || []), ...folder.images]
        }
        return
      }

      if (name === '日期归档' || name.startsWith('日期归档/')) {
        const archiveRoot = rootNodes.find((r) => r.name === '日期归档')
        if (name === '日期归档') {
          archiveRoot.images = [...(archiveRoot.images || []), ...folder.images]
        } else if (dateInfo) {
          const yearNode = getOrCreate(archiveRoot.children, dateInfo.year, archiveRoot.id)
          const monthNode = getOrCreate(yearNode.children, dateInfo.month, yearNode.id)
          const leafName = name.split('/').pop()
          addLeaf(monthNode.children, {
            name,
            id: `${monthNode.id}/${leafName}`,
            displayName: dateInfo.full,
            relPath: name,
            images: folder.images,
            lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
          })
        } else {
          const parts = name.split('/')
          const subName = parts[parts.length - 1]
          if (/^\d{4}$/.test(subName)) {
            const yearNode = getOrCreate(archiveRoot.children, subName, archiveRoot.id)
            yearNode.images = [...(yearNode.images || []), ...folder.images]
            yearNode.lastMod = Math.max(
              yearNode.lastMod || 0,
              ...folder.images.map((i) => new Date(i.modTime).getTime()),
            )
          } else {
            addLeaf(archiveRoot.children, {
              name,
              id: `${archiveRoot.id}/${subName}`,
              displayName: subName,
              relPath: name,
              images: folder.images,
              lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
            })
          }
        }
        return
      }

      if (!name.includes('XYZ') && !name.includes('OXYZ') && !name.includes('修复') && dateInfo) {
        const archiveRoot = rootNodes.find((r) => r.name === '日期归档')
        const yearNode = getOrCreate(archiveRoot.children, dateInfo.year, archiveRoot.id)
        const monthNode = getOrCreate(yearNode.children, dateInfo.month, yearNode.id)
        const leafName = name.split('/').pop()

        addLeaf(monthNode.children, {
          name,
          id: `${monthNode.id}/${leafName}`,
          displayName: name.includes('日期归档') ? name.replace('日期归档', '') : name,
          relPath: name,
          images: folder.images,
          lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
        })
        return
      }

      if (name.includes('XYZ') || name.startsWith('0-')) {
        const oxyzRoot = rootNodes.find((r) => r.name === 'OXYZ测试图片')
        const info = matchDate(name)

        if (info) {
          const yearNode = getOrCreate(oxyzRoot.children, info.year, oxyzRoot.id)
          const monthNode = getOrCreate(yearNode.children, info.month, yearNode.id)
          addLeaf(monthNode.children, {
            name,
            id: `${monthNode.id}/${name.split('/').pop()}`,
            displayName: info.full,
            relPath: name,
            images: folder.images,
            lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
          })
        } else if (name === '0-XYZ测试图片' || name === 'OXYZ测试图片') {
          oxyzRoot.images = [...(oxyzRoot.images || []), ...folder.images]
        } else {
          const baseName = name.split('/').pop()
          addLeaf(oxyzRoot.children, {
            name,
            id: `${oxyzRoot.id}/${baseName}`,
            displayName: baseName,
            relPath: name,
            images: folder.images,
            lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
          })
        }
        return
      }

      if (name.includes('修复')) {
        const repairRoot = rootNodes.find((r) => r.name === '修复')
        const info = matchDate(name)

        let typeNode = null
        if (name.includes('手动')) typeNode = repairRoot.children.find((c) => c.name === '手动')
        if (name.includes('自动')) typeNode = repairRoot.children.find((c) => c.name === '自动')

        if (typeNode && info) {
          const yearNode = getOrCreate(typeNode.children, info.year, typeNode.id)
          const monthNode = getOrCreate(yearNode.children, info.month, yearNode.id)
          addLeaf(monthNode.children, {
            name,
            id: `${monthNode.id}/${name.split('/').pop()}`,
            displayName: info.full,
            relPath: name,
            images: folder.images,
            lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
          })
          return
        }

        const baseName = name.split('/').pop()
        const targetChildren = typeNode ? typeNode.children : repairRoot.children
        addLeaf(targetChildren, {
          name,
          id: `${typeNode ? typeNode.id : repairRoot.id}/${baseName}`,
          displayName: baseName,
          relPath: name,
          images: folder.images,
          lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
        })
        return
      }

      const outputRoot = rootNodes.find((r) => r.name === 'output')
      if (outputRoot) {
        addLeaf(outputRoot.children, {
          name,
          id: `output/${name}`,
          displayName: name,
          relPath: name,
          children: [],
          images: folder.images,
          lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
        })
      } else {
        otherNodes.push({
          name,
          id: name,
          displayName: name,
          relPath: name,
          children: [],
          images: folder.images,
          lastMod: Math.max(...folder.images.map((i) => new Date(i.modTime).getTime())),
        })
      }
    })

    const sortNodes = (nodes) => {
      nodes.sort((a, b) => {
        const timeDiff = (b.lastMod || 0) - (a.lastMod || 0)
        if (timeDiff !== 0) return timeDiff
        return (b.name || '').localeCompare(a.name || '')
      })
      nodes.forEach((node) => {
        if (node.children && node.children.length > 0) {
          sortNodes(node.children)
        }
      })
    }

    rootNodes.forEach((root) => {
      if (root.children.length > 0) sortNodes(root.children)
    })

    const favoritesRoot = rootNodes.find((node) => node.id === 'favorites')
    if (favoritesRoot) {
      favoritesRoot.children = (favoriteGroups.value || [])
        .map((group) => {
          const groupImages = imagesToUse
            .filter((img) => (group.paths || []).includes(img.relPath))
            .sort((a, b) => new Date(b.modTime) - new Date(a.modTime))

          const lastMod = groupImages.length > 0
            ? Math.max(...groupImages.map((img) => new Date(img.modTime).getTime()))
            : 0

          return {
            name: group.name,
            id: `favorite-group:${group.id}`,
            displayName: group.name,
            relPath: '',
            groupId: group.id,
            children: [],
            subs: [],
            images: groupImages,
            lastMod,
            count: groupImages.length,
            isFavoriteGroup: true,
          }
        })
        .sort((a, b) => {
          const timeDiff = (b.lastMod || 0) - (a.lastMod || 0)
          if (timeDiff !== 0) return timeDiff
          return (a.name || '').localeCompare(b.name || '')
        })
      favoritesRoot.subs = favoritesRoot.children
    }

    const buildCustomNode = (rootPath, name, id, imgs, icon) => {
      const normalizedRootPath = normalizeFolderPath(rootPath)
      const prefix = normalizedRootPath ? `${normalizedRootPath}/` : ''

      const nodeImgs = imgs.filter((img) => {
        const relPath = normalizeFolderPath(img.relPath)
        if (!normalizedRootPath) return !relPath.includes('/')
        return relPath === normalizedRootPath || relPath.startsWith(prefix)
      })

      const childrenMap = {}
      nodeImgs.forEach((img) => {
        const relPath = normalizeFolderPath(img.relPath)
        const rest = normalizedRootPath
          ? relPath.slice(prefix.length)
          : relPath
        const parts = rest.split('/')
        if (parts.length > 1) {
          const childName = parts[0]
          if (!childrenMap[childName]) childrenMap[childName] = []
          childrenMap[childName].push(img)
        }
      })

      const directImgs = nodeImgs.filter((img) => {
        const relPath = normalizeFolderPath(img.relPath)
        const rest = normalizedRootPath
          ? relPath.slice(prefix.length)
          : relPath
        return !rest.includes('/')
      })

      const children = Object.keys(childrenMap)
        .map((childFolderName) => {
          const childPath = normalizedRootPath
            ? `${normalizedRootPath}/${childFolderName}`
            : childFolderName
          const childId = `${id}/${childFolderName}`
          return buildCustomNode(childPath, childFolderName, childId, childrenMap[childFolderName], icon)
        })
        .filter(Boolean)

      if (directImgs.length === 0 && children.length === 0) {
        return {
          name,
          id,
          displayName: name,
          relPath: normalizedRootPath,
          icon,
          isCustomRoot: true,
          images: [],
          children: [],
          subs: [],
          lastMod: 0,
        }
      }

      children.sort((a, b) => {
        const timeDiff = (b.lastMod || 0) - (a.lastMod || 0)
        if (timeDiff !== 0) return timeDiff
        return (b.name || '').localeCompare(a.name || '')
      })

      const directLastMod = directImgs.length > 0
        ? Math.max(...directImgs.map((img) => new Date(img.modTime).getTime()))
        : 0
      const childLastMod = children.length > 0
        ? Math.max(...children.map((child) => child.lastMod || 0))
        : 0

      return {
        name,
        id,
        displayName: name,
        relPath: normalizedRootPath,
        icon,
        isCustomRoot: true,
        images: directImgs,
        children,
        subs: children,
        lastMod: Math.max(directLastMod, childLastMod),
      }
    }

    const customRootNodes = customRoots.value.map((root) =>
      buildCustomNode(
        normalizeFolderPath(root.path),
        root.name,
        `custom:${root.id}`,
        imagesToUse,
        root.icon,
      ),
    )

    const pruneManagedNodes = (nodes) =>
      (nodes || [])
        .map((node) => {
          const children = pruneManagedNodes(node.children || node.subs || [])
          const candidatePath = normalizeFolderPath(node.relPath || node.name)
          const hasImages = Array.isArray(node.images) && node.images.length > 0

          if (candidatePath && isManagedByCustomRoot(candidatePath)) {
            return null
          }

          if (!hasImages && children.length === 0 && node.type !== 'root' && !node.isFavoriteGroup) {
            return null
          }

          return {
            ...node,
            children,
            subs: children,
          }
        })
        .filter(Boolean)

    rootNodes.forEach((root) => {
      root.children = pruneManagedNodes(root.children || [])
      root.subs = root.children
    })

    const visibleOtherNodes = pruneManagedNodes(otherNodes)
    const visibleRootNodes = rootNodes.filter((node) => {
      if (node.name !== 'output') return true
      return (node.images && node.images.length > 0) || (node.children && node.children.length > 0)
    })

    return [...visibleRootNodes, ...visibleOtherNodes, ...customRootNodes].map((node) => ({
      ...node,
      subs: node.children || [],
    }))
  })

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

    if (activeRoot.value === 'favorites') {
      const imgs = images.value.filter((img) => favorites.value.has(img.relPath))
      imgs.sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
      if (!activeSub.value) return imgs

      const groupId = activeSub.value.startsWith('favorite-group:')
        ? activeSub.value.replace('favorite-group:', '')
        : ''
      if (!groupId) return imgs

      const group = favoriteGroups.value.find((item) => item.id === groupId)
      if (!group) return imgs

      return imgs.filter((img) => (group.paths || []).includes(img.relPath))
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
    return targetNode ? collectImages(targetNode) : []
  })

  const availableModels = computed(() => {
    return buildGroupedFilterOptions(images.value.map((img) => img?.model || ''))
  })

  const availableLoras = computed(() => {
    const loraValues = []
    images.value.forEach((img) => {
      ;(img?.loras || []).forEach((lora) => {
        loraValues.push(lora)
      })
    })
    return buildGroupedFilterOptions(loraValues)
  })

  const activeDateLabel = computed(() =>
    getDatePresetLabel(activeDatePreset.value, activeDateValue.value),
  )

  const hasActiveWorkbenchFilters = computed(() =>
    activeDatePreset.value !== 'all' || !!activeModelFilter.value || !!activeLoraFilter.value,
  )

  const workbenchFilteredImages = computed(() =>
    images.value.filter((img) =>
      imageMatchesWorkbenchFilters(
        img,
        activeDatePreset.value,
        activeDateValue.value,
        activeModelFilter.value,
        activeLoraFilter.value,
      ),
    ),
  )

  const dateWorkbenchSummary = computed(() => {
    const dateCountMap = buildDateCountMap(images.value)
    const datedImages = images.value.filter((img) => getImageDateKey(img))
    const countWithPreset = (preset) =>
      datedImages.filter((img) =>
        imageMatchesWorkbenchFilters(
          img,
          preset,
          '',
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

  const finalImages = computed(() => {
    let imgs = currentImages.value

    if (imgs.length > 0 && hasActiveWorkbenchFilters.value) {
      imgs = imgs.filter((img) =>
        imageMatchesWorkbenchFilters(
          img,
          activeDatePreset.value,
          activeDateValue.value,
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
    const startIndex = (currentPage.value - 1) * itemsPerPage.value
    const endIndex = startIndex + itemsPerPage.value
    return stackedImages.value.slice(startIndex, endIndex)
  })

  const totalPages = computed(() => Math.ceil(stackedImages.value.length / itemsPerPage.value))

  const setPage = (page) => {
    if (page < 1 || page > totalPages.value) return
    currentPage.value = page
    localStorage.setItem('currentPage', page)
  }

  const prevPage = () => setPage(currentPage.value - 1)
  const nextPage = () => setPage(currentPage.value + 1)

  const setItemsPerPage = (count) => {
    itemsPerPage.value = count
    localStorage.setItem('itemsPerPage', count)
    setPage(1)
  }

  const resetPage = () => {
    setPage(1)
  }

  watch([activeRoot, activeSub, activeChild], () => {
    resetPage()
  })

  watch(searchQuery, (value) => {
    localStorage.setItem('gallerySearchQuery', value)
    resetPage()
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
    [activeDatePreset, activeDateValue, activeModelFilter, activeLoraFilter],
    ([datePreset, dateValue, modelFilter, loraFilter]) => {
      localStorage.setItem('activeDatePreset', datePreset)
      localStorage.setItem('activeDateValue', dateValue)
      localStorage.setItem('activeModelFilter', modelFilter)
      localStorage.setItem('activeLoraFilter', loraFilter)
      resetPage()
    },
  )

  const setActiveDatePreset = (preset) => {
    activeDatePreset.value = preset || 'all'
    if (activeDatePreset.value !== 'custom') {
      activeDateValue.value = ''
    }
  }

  const setActiveDateValue = (value) => {
    activeDateValue.value = value || ''
    activeDatePreset.value = value ? 'custom' : 'all'
  }

  const clearDateFilter = () => {
    activeDatePreset.value = 'all'
    activeDateValue.value = ''
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
        if (children.length > 0) {
          const subExists = children.find((s) => s.id === activeSub.value)
          if (!activeSub.value || !subExists) {
            activeSub.value = children[0].id
          }
        }
      }

      isInitialized.value = true
    }
  }

  const startPolling = () => {
    fetchImages().then(initAutoSelect)
    return null
  }

  const handleDelete = async (img) => {
    const ok = await confirm(`确定将 ${img.name} 移至回收站吗？`)
    if (!ok) return

    const original = images.value
    images.value = images.value.filter((i) => i.relPath !== img.relPath)
    try {
      await App.DeleteImage(img.relPath)
      showToast('删除成功', 'success')
      if (favorites.value.has(img.relPath)) {
        favorites.value.delete(img.relPath)
        App.RemoveFavorite(img.relPath).catch(console.error)
      }
      if (imageTags.value[img.relPath]) {
        delete imageTags.value[img.relPath]
      }
    } catch (err) {
      showToast('删除失败', 'error')
      images.value = original
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
    fetchImages()
  }

  const setSortOrder = (order) => {
    sortOrder.value = order
    localStorage.setItem('sortOrder', order)
    fetchImages()
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

  return {
    images,
    favorites,
    favoriteGroups,
    loading,
    activeRoot,
    activeSub,
    activeChild,
    fileTree,
    scopeImageCount: computed(() => currentImages.value.length),
    currentImages: finalImages,
    fetchImages,
    fetchFavorites,
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
    activeDatePreset,
    activeDateValue,
    activeModelFilter,
    activeLoraFilter,
    activeDateLabel,
    hasActiveWorkbenchFilters,
    setActiveDatePreset,
    setActiveDateValue,
    clearDateFilter,
    setActiveModel,
    setActiveLora,
    clearWorkbenchFilters,
    clearSearchQuery,
    toggleStacking: () => {
      isStackingEnabled.value = !isStackingEnabled.value
      localStorage.setItem('isStackingEnabled', isStackingEnabled.value)
    }
  }
}

