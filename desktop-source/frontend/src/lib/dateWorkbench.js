const DATE_FOLDER_PATTERN = /^\d{4}-\d{2}-\d{2}$/

export const formatDateKey = (date) => {
  if (!(date instanceof Date) || Number.isNaN(date.getTime())) return ''
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

export const parseDateKey = (dateKey) => {
  const value = String(dateKey || '').trim()
  if (!DATE_FOLDER_PATTERN.test(value)) return null
  const [year, month, day] = value.split('-').map(Number)
  const date = new Date(year, month - 1, day)
  if (Number.isNaN(date.getTime())) return null
  return date
}

export const extractDateFolder = (relPath) => {
  const normalized = String(relPath || '')
    .replace(/\\/g, '/')
    .replace(/^\/+|\/+$/g, '')

  if (!normalized) return ''

  const parts = normalized.split('/').filter(Boolean)
  if (parts.length <= 1) return ''

  const folderParts = parts.slice(0, -1)
  for (let index = folderParts.length - 1; index >= 0; index -= 1) {
    const segment = folderParts[index]
    if (DATE_FOLDER_PATTERN.test(segment)) return segment
  }
  return ''
}

export const startOfDay = (date) => {
  const next = new Date(date)
  next.setHours(0, 0, 0, 0)
  return next
}

export const endOfDay = (date) => {
  const next = new Date(date)
  next.setHours(23, 59, 59, 999)
  return next
}

export const getDatePresetLabel = (preset, customDate = '') => {
  switch (preset) {
  case 'today':
    return '今天'
  case 'yesterday':
    return '昨天'
  case 'last7':
    return '最近7天'
  case 'month':
    return '本月'
  case 'custom':
    return customDate || '指定日期'
  default:
    return '全部日期'
  }
}

export const getDatePresetRange = (preset, customDate = '', now = new Date()) => {
  const today = startOfDay(now)

  switch (preset) {
  case 'today':
    return { start: today, end: endOfDay(today) }
  case 'yesterday': {
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)
    return { start: yesterday, end: endOfDay(yesterday) }
  }
  case 'last7': {
    const start = new Date(today)
    start.setDate(start.getDate() - 6)
    return { start, end: endOfDay(now) }
  }
  case 'month': {
    const start = new Date(today.getFullYear(), today.getMonth(), 1)
    const end = endOfDay(now)
    return { start, end }
  }
  case 'custom': {
    const parsed = parseDateKey(customDate)
    if (!parsed) return null
    return { start: startOfDay(parsed), end: endOfDay(parsed) }
  }
  default:
    return null
  }
}

export const matchesDatePreset = (dateKey, preset, customDate = '', now = new Date()) => {
  if (!preset || preset === 'all') return true
  const parsed = parseDateKey(dateKey)
  if (!parsed) return false

  const range = getDatePresetRange(preset, customDate, now)
  if (!range) return true
  return parsed >= range.start && parsed <= range.end
}

export const buildDateCountMap = (images = []) => {
  const counts = new Map()
  ;(images || []).forEach((image) => {
    const dateKey = extractDateFolder(image.relPath)
    if (!dateKey) return
    counts.set(dateKey, (counts.get(dateKey) || 0) + 1)
  })
  return counts
}
