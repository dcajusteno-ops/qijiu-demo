import { extractDateFolder, formatDateKey } from '@/lib/dateWorkbench'

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

const normalizePerformanceSettings = (settings = {}) => ({
  mode: ['auto', 'standard', 'performance'].includes(settings?.mode) ? settings.mode : 'auto',
  initialBatchSize: Math.min(Math.max(Number(settings?.initialBatchSize) || 60, 20), 500),
  pageSize: Math.min(Math.max(Number(settings?.pageSize) || 50, 20), 500),
  thumbPreferred: settings?.thumbPreferred !== false,
  backgroundVariantWarmup: settings?.backgroundVariantWarmup !== false,
  metadataLazy: settings?.metadataLazy !== false,
})

export {
  buildGroupedFilterOptions,
  buildImageDisplayPath,
  dateFolderPattern,
  getDateSegment,
  getImageDateKey,
  normalizeAssetKey,
  normalizeFilterValue,
  normalizeFolderPath,
  normalizePerformanceSettings,
  normalizeSearchText,
  prettifyAssetLabel,
  stripModelExtension,
  stripPathSegments,
  syncGroupedFilterValue,
}
