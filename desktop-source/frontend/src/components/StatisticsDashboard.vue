<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import Lightbox from './Lightbox.vue'
import { useImages } from '@/composables/useImages'
import * as App from '@/api'
import {
  ArrowUpRight,
  BarChart3,
  CalendarDays,
  ChevronDown,
  Clock3,
  Eye,
  Flame,
  FolderOpen,
  HardDrive,
  History,
  ImageIcon,
  RefreshCw,
  Sparkles,
  TrendingUp,
  Zap,
} from 'lucide-vue-next'

const TEXT = {
  badge: '数据视界',
  title: '生成历史时间线',
  desc: '用更清晰的时间结构回看作品，观察创作频率、活跃时段和阶段变化。',
  refresh: '刷新数据',
  trendTitle: '产出趋势',
  trendDesc: '按年、按月或按日查看作品数量变化。',
  heatmapTitle: '活跃热图',
  heatmapDesc: '点击任意有记录的日期，直接跳到当天时间线。',
  timelineTitle: '历史回溯',
  timelineDesc: '按日或按月整理创作节点，点击节点查看该时间段的全部作品。',
  detailTitle: '节点详情',
  emptyChart: '当前没有足够的数据来绘制图表。',
  emptyTimeline: '还没有可展示的生成历史。',
  emptyDetail: '选择一个时间节点后，这里会显示该时间段的作品。',
  emptyLatest: '还没有可展示的最新作品。',
  emptyImages: '这个时间点暂时没有可展示的作品。',
  preview: '预览',
  locate: '定位',
  delete: '删除',
  loadMoreEntries: '加载更多节点',
  loadMoreImages: '加载更多作品',
}

const showToast = (message, type) => {
  if (type === 'error') return toast.error(message)
  if (type === 'success') return toast.success(message)
  return toast(message)
}

const confirmAction = async (message) => window.confirm(message)

const {
  images,
  favoriteGroups,
  tags,
  imageTags,
  imageNotes,
  toggleFavorite,
  addTagToImage,
  removeTagFromImage,
  openImageLocation,
  fetchImages,
  fetchTags,
  fetchImageTags,
  fetchImageNotes,
  handleDelete,
} = useImages(showToast, confirmAction)

const stats = ref(null)
const loading = ref(false)
const viewMode = ref(localStorage.getItem('statsViewMode') || 'month')
const timelineMode = ref(localStorage.getItem('statsTimelineMode') || 'day')
const activeTimelineKey = ref('')
const visibleTimelineCount = ref(16)
const visibleTimelineImages = ref(8)
const hoveredTrendKey = ref('')
const chartHost = ref(null)
const chartHostWidth = ref(0)

const lightboxOpen = ref(false)
const lightboxImage = ref(null)
const lightboxImages = ref([])
const lightboxIndex = ref(0)

const chartHeight = 280
const chartPadding = { top: 18, right: 18, bottom: 42, left: 52 }
let pollingInterval = null
let chartResizeObserver = null

const dayFormatter = new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  weekday: 'short',
})

const shortDayFormatter = new Intl.DateTimeFormat('zh-CN', {
  month: '2-digit',
  day: '2-digit',
})

const monthFormatter = new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: 'long',
})

const dateTimeFormatter = new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
})

const timeFormatter = new Intl.DateTimeFormat('zh-CN', {
  hour: '2-digit',
  minute: '2-digit',
})

const formatDateKey = (date) => {
  const y = date.getFullYear()
  const m = `${date.getMonth() + 1}`.padStart(2, '0')
  const d = `${date.getDate()}`.padStart(2, '0')
  return `${y}-${m}-${d}`
}

const formatMonthKey = (date) => {
  const y = date.getFullYear()
  const m = `${date.getMonth() + 1}`.padStart(2, '0')
  return `${y}-${m}`
}

const parseTimelineKey = (key, mode) => {
  if (!key) return null
  if (mode === 'month') {
    const [year, month] = key.split('-').map(Number)
    return new Date(year, (month || 1) - 1, 1)
  }
  const [year, month, day] = key.split('-').map(Number)
  return new Date(year, (month || 1) - 1, day || 1)
}

const getTimelineKey = (date, mode) => (mode === 'month' ? formatMonthKey(date) : formatDateKey(date))

const startOfWeek = (date) => {
  const next = new Date(date)
  const weekDay = (next.getDay() + 6) % 7
  next.setDate(next.getDate() - weekDay)
  next.setHours(0, 0, 0, 0)
  return next
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(index === 0 ? 0 : 2)} ${units[index]}`
}

const loadStats = async () => {
  loading.value = true
  try {
    stats.value = await App.GetStatistics(viewMode.value)
  } catch (error) {
    console.error('Failed to load statistics:', error)
    toast.error('统计数据读取失败')
  } finally {
    loading.value = false
  }
}

const refreshAll = async () => {
  if (loading.value) return
  await fetchImages()
  await Promise.all([loadStats(), fetchTags(), fetchImageTags(), fetchImageNotes()])
  toast.success('数据视界已刷新')
}

const syncChartHostWidth = () => {
  chartHostWidth.value = chartHost.value?.clientWidth || 0
}

const setupChartResizeObserver = () => {
  if (chartResizeObserver) {
    chartResizeObserver.disconnect()
    chartResizeObserver = null
  }
  if (!chartHost.value || typeof ResizeObserver === 'undefined') return
  chartResizeObserver = new ResizeObserver(() => syncChartHostWidth())
  chartResizeObserver.observe(chartHost.value)
  syncChartHostWidth()
}

watch(viewMode, async (value) => {
  localStorage.setItem('statsViewMode', value)
  hoveredTrendKey.value = ''
  await loadStats()
})

watch(timelineMode, (value) => {
  localStorage.setItem('statsTimelineMode', value)
  visibleTimelineCount.value = 16
})

watch(activeTimelineKey, () => {
  visibleTimelineImages.value = 8
})

onMounted(async () => {
  await Promise.all([
    images.value.length ? Promise.resolve() : fetchImages(),
    fetchTags(),
    fetchImageTags(),
    fetchImageNotes(),
  ])
  await loadStats()
  await nextTick()
  setupChartResizeObserver()
  pollingInterval = setInterval(loadStats, 8000)
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
  if (chartResizeObserver) chartResizeObserver.disconnect()
})

const sortedImages = computed(() =>
  [...images.value].sort((a, b) => new Date(b.modTime) - new Date(a.modTime)),
)

const latestCreation = computed(() => sortedImages.value[0] || null)

const activityCountByDay = computed(() => {
  const map = new Map()
  sortedImages.value.forEach((image) => {
    const key = formatDateKey(new Date(image.modTime))
    map.set(key, (map.get(key) || 0) + 1)
  })
  return map
})

const activeDaysSorted = computed(() => [...activityCountByDay.value.keys()].sort((a, b) => a.localeCompare(b)))

const timelineSummary = computed(() => {
  const days = activeDaysSorted.value
  if (!days.length) {
    return { activeDays: 0, longestStreak: 0, currentStreak: 0, firstDate: '', lastDate: '' }
  }

  let longestStreak = 1
  let run = 1
  for (let i = 1; i < days.length; i += 1) {
    const prev = parseTimelineKey(days[i - 1], 'day')
    const curr = parseTimelineKey(days[i], 'day')
    const diff = Math.round((curr - prev) / 86400000)
    if (diff === 1) {
      run += 1
      longestStreak = Math.max(longestStreak, run)
    } else {
      run = 1
    }
  }

  let currentStreak = 1
  for (let i = days.length - 1; i > 0; i -= 1) {
    const prev = parseTimelineKey(days[i - 1], 'day')
    const curr = parseTimelineKey(days[i], 'day')
    const diff = Math.round((curr - prev) / 86400000)
    if (diff === 1) currentStreak += 1
    else break
  }

  return {
    activeDays: days.length,
    longestStreak,
    currentStreak,
    firstDate: days[0],
    lastDate: days[days.length - 1],
  }
})

const timelineEntries = computed(() => {
  const groups = new Map()

  sortedImages.value.forEach((image) => {
    const date = new Date(image.modTime)
    const key = getTimelineKey(date, timelineMode.value)

    if (!groups.has(key)) {
      groups.set(key, {
        key,
        mode: timelineMode.value,
        images: [],
        totalSize: 0,
        latestDate: date,
        earliestDate: date,
        activeDays: new Set(),
      })
    }

    const entry = groups.get(key)
    entry.images.push(image)
    entry.totalSize += image.size || 0
    entry.latestDate = entry.latestDate > date ? entry.latestDate : date
    entry.earliestDate = entry.earliestDate < date ? entry.earliestDate : date
    entry.activeDays.add(formatDateKey(date))
  })

  return [...groups.values()]
    .map((entry) => ({
      ...entry,
      count: entry.images.length,
      coverImages: entry.images.slice(0, 4),
      activeDaysCount: entry.activeDays.size,
      label: entry.mode === 'month'
        ? monthFormatter.format(parseTimelineKey(entry.key, 'month'))
        : dayFormatter.format(parseTimelineKey(entry.key, 'day')),
      shortLabel: entry.mode === 'month'
        ? entry.key
        : shortDayFormatter.format(parseTimelineKey(entry.key, 'day')),
      caption: entry.mode === 'month'
        ? `${entry.activeDays.size} 天有产出`
        : `${timeFormatter.format(entry.earliestDate)} - ${timeFormatter.format(entry.latestDate)}`,
    }))
    .sort((a, b) => b.latestDate - a.latestDate)
})

watch(
  timelineEntries,
  (entries) => {
    if (!entries.length) {
      activeTimelineKey.value = ''
      return
    }
    if (!entries.some((entry) => entry.key === activeTimelineKey.value)) {
      activeTimelineKey.value = entries[0].key
    }
  },
  { immediate: true },
)

const activeTimelineEntry = computed(() =>
  timelineEntries.value.find((entry) => entry.key === activeTimelineKey.value) || null,
)

const visibleTimelineEntries = computed(() => timelineEntries.value.slice(0, visibleTimelineCount.value))
const visibleTimelineImagesList = computed(() =>
  activeTimelineEntry.value?.images.slice(0, visibleTimelineImages.value) || [],
)

const timelineYearGroups = computed(() => {
  const groups = new Map()
  visibleTimelineEntries.value.forEach((entry) => {
    const year = String(parseTimelineKey(entry.key, entry.mode)?.getFullYear() || '')
    if (!groups.has(year)) groups.set(year, [])
    groups.get(year).push(entry)
  })
  return [...groups.entries()].map(([year, entries]) => ({ year, entries }))
})

const sortedTrendKeys = computed(() => Object.keys(stats.value?.byDate || {}).sort((a, b) => a.localeCompare(b)))
const trendMaxCount = computed(() =>
  sortedTrendKeys.value.reduce((max, key) => Math.max(max, stats.value?.byDate?.[key] || 0), 0),
)

const chartYAxisMax = computed(() => {
  const max = trendMaxCount.value
  if (max <= 5) return 5
  const magnitude = 10 ** Math.floor(Math.log10(max))
  const steps = [1, 2, 5, 10]
  return steps.map((step) => step * magnitude).find((value) => value >= max)
    || Math.ceil(max / magnitude) * magnitude
})

const chartPointWidth = computed(() => {
  if (viewMode.value === 'year') return 140
  if (viewMode.value === 'month') return 78
  return 34
})

const chartCanvasWidth = computed(() =>
  Math.max(
    720,
    chartHostWidth.value,
    sortedTrendKeys.value.length * chartPointWidth.value + chartPadding.left + chartPadding.right,
  ),
)
const chartInnerHeight = computed(() => chartHeight - chartPadding.top - chartPadding.bottom)
const chartInnerWidth = computed(() => chartCanvasWidth.value - chartPadding.left - chartPadding.right)

const yTicks = computed(() => {
  const step = chartYAxisMax.value / 4
  return Array.from({ length: 5 }, (_, index) => {
    const value = Math.round(step * index)
    const y = chartPadding.top + chartInnerHeight.value - (value / chartYAxisMax.value) * chartInnerHeight.value
    return { value, y }
  })
})

const formatTrendLabel = (key) => {
  if (viewMode.value === 'year') return `${key} 年`
  if (viewMode.value === 'month') return monthFormatter.format(parseTimelineKey(key, 'month'))
  return dayFormatter.format(parseTimelineKey(key, 'day'))
}

const formatTrendAxisLabel = (key) => {
  if (viewMode.value === 'year') return key
  if (viewMode.value === 'month') {
    const date = parseTimelineKey(key, 'month')
    return `${date.getMonth() + 1}月`
  }
  const date = parseTimelineKey(key, 'day')
  return `${date.getMonth() + 1}/${date.getDate()}`
}

const trendPoints = computed(() => {
  if (!sortedTrendKeys.value.length) return []
  return sortedTrendKeys.value.map((key, index) => {
    const count = stats.value?.byDate?.[key] || 0
    const denominator = sortedTrendKeys.value.length === 1 ? 1 : sortedTrendKeys.value.length - 1
    const x = chartPadding.left + (index / denominator) * chartInnerWidth.value
    const y = chartPadding.top + chartInnerHeight.value - (count / chartYAxisMax.value) * chartInnerHeight.value
    return { key, count, label: formatTrendLabel(key), axisLabel: formatTrendAxisLabel(key), x, y }
  })
})

const createSmoothPath = (points) => {
  if (!points.length) return ''
  if (points.length === 1) return `M ${points[0].x} ${points[0].y}`
  let path = `M ${points[0].x} ${points[0].y}`
  for (let i = 0; i < points.length - 1; i += 1) {
    const current = points[i]
    const next = points[i + 1]
    const controlX = (current.x + next.x) / 2
    path += ` C ${controlX} ${current.y}, ${controlX} ${next.y}, ${next.x} ${next.y}`
  }
  return path
}

const trendLinePath = computed(() => createSmoothPath(trendPoints.value))
const trendAreaPath = computed(() => {
  if (!trendPoints.value.length || !trendLinePath.value) return ''
  const baseline = chartPadding.top + chartInnerHeight.value
  const first = trendPoints.value[0]
  const last = trendPoints.value[trendPoints.value.length - 1]
  return `${trendLinePath.value} L ${last.x} ${baseline} L ${first.x} ${baseline} Z`
})

const chartLabelStep = computed(() => Math.max(1, Math.ceil(sortedTrendKeys.value.length / (viewMode.value === 'day' ? 10 : 7))))
const defaultTrendPoint = computed(() => trendPoints.value[trendPoints.value.length - 1] || null)
const activeTrendPoint = computed(() =>
  trendPoints.value.find((point) => point.key === hoveredTrendKey.value) || defaultTrendPoint.value,
)
const trendHoverZones = computed(() =>
  trendPoints.value.map((point, index, points) => {
    const prevX = points[index - 1]?.x ?? chartPadding.left
    const nextX = points[index + 1]?.x ?? chartCanvasWidth.value - chartPadding.right
    const startX = index === 0 ? chartPadding.left : (prevX + point.x) / 2
    const endX = index === points.length - 1 ? chartCanvasWidth.value - chartPadding.right : (point.x + nextX) / 2
    return { ...point, startX, width: Math.max(18, endX - startX) }
  }),
)
const trendTooltipStyle = computed(() => {
  if (!activeTrendPoint.value) return {}
  const tooltipWidth = 148
  const safeLeft = Math.min(
    chartCanvasWidth.value - chartPadding.right - tooltipWidth,
    Math.max(chartPadding.left, activeTrendPoint.value.x - tooltipWidth / 2),
  )
  return {
    left: `${safeLeft}px`,
    top: `${Math.max(12, activeTrendPoint.value.y - 88)}px`,
  }
})

watch(
  trendPoints,
  async () => {
    await nextTick()
    setupChartResizeObserver()
  },
  { flush: 'post' },
)

const heatmapMaxCount = computed(() =>
  [...activityCountByDay.value.values()].reduce((max, count) => Math.max(max, count), 0),
)

const heatmapWeeks = computed(() => {
  const totalWeeks = 24
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const firstWeekStart = startOfWeek(today)
  firstWeekStart.setDate(firstWeekStart.getDate() - (totalWeeks - 1) * 7)

  let previousMonth = -1
  return Array.from({ length: totalWeeks }, (_, weekIndex) => {
    const weekStart = new Date(firstWeekStart)
    weekStart.setDate(firstWeekStart.getDate() + weekIndex * 7)
    const days = Array.from({ length: 7 }, (_, dayIndex) => {
      const date = new Date(weekStart)
      date.setDate(weekStart.getDate() + dayIndex)
      date.setHours(0, 0, 0, 0)
      const key = formatDateKey(date)
      return {
        key,
        count: activityCountByDay.value.get(key) || 0,
        isFuture: date > today,
        label: dayFormatter.format(date),
      }
    })
    const monthLabel = weekStart.getMonth() !== previousMonth ? `${weekStart.getMonth() + 1}月` : ''
    previousMonth = weekStart.getMonth()
    return { key: `week-${weekIndex}-${formatDateKey(weekStart)}`, monthLabel, days }
  })
})

const heatmapLevelClass = (count, isFuture = false) => {
  if (isFuture) return 'border-border/50 bg-transparent'
  if (!count) return 'border-border/60 bg-muted/25'
  const ratio = count / Math.max(heatmapMaxCount.value, 1)
  if (ratio >= 0.85) return 'border-foreground/60 bg-foreground'
  if (ratio >= 0.55) return 'border-foreground/40 bg-foreground/70'
  if (ratio >= 0.25) return 'border-foreground/25 bg-foreground/45'
  return 'border-foreground/15 bg-foreground/20'
}

const jumpToTimelineKey = async (dateKey) => {
  timelineMode.value = 'day'
  await nextTick()
  activeTimelineKey.value = dateKey
  const index = timelineEntries.value.findIndex((entry) => entry.key === dateKey)
  if (index >= 0) visibleTimelineCount.value = Math.max(visibleTimelineCount.value, index + 6)
  await nextTick()
  document.getElementById('stats-timeline-section')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

const openTimelineLightbox = (items, index) => {
  if (!items?.length) return
  lightboxImages.value = items
  lightboxIndex.value = index
  lightboxImage.value = items[index] || null
  lightboxOpen.value = Boolean(lightboxImage.value)
}

const handleLightboxNavigate = (direction) => {
  const nextIndex = direction === 'prev' ? lightboxIndex.value - 1 : lightboxIndex.value + 1
  if (nextIndex < 0 || nextIndex >= lightboxImages.value.length) return
  lightboxIndex.value = nextIndex
  lightboxImage.value = lightboxImages.value[nextIndex]
}

const handleTimelineDelete = async (image) => {
  await handleDelete(image)
  await loadStats()
}

const statCards = computed(() => [
  { key: 'today', title: '今日新增', value: `${stats.value?.todayCount ?? 0}`, caption: '今天进入图库的作品数量', icon: Sparkles },
  { key: 'total', title: '作品总量', value: `${stats.value?.totalCount ?? 0}`, caption: '当前已收录的全部图片', icon: ImageIcon },
  { key: 'days', title: '活跃天数', value: `${timelineSummary.value.activeDays}`, caption: '至少有作品产出的日期数', icon: CalendarDays },
  { key: 'size', title: '存储占用', value: formatSize(stats.value?.totalSize ?? 0), caption: '作品与元数据占用空间', icon: HardDrive },
])

const insightCards = computed(() => [
  { key: 'streak', title: '最长连更', value: `${timelineSummary.value.longestStreak} 天`, caption: '历史上最长连续创作天数', icon: Flame },
  { key: 'current', title: '当前节奏', value: `${timelineSummary.value.currentStreak} 天`, caption: '最近一段连续活跃时长', icon: TrendingUp },
  { key: 'density', title: '活跃日均产出', value: `${(sortedImages.value.length / Math.max(timelineSummary.value.activeDays, 1)).toFixed(1)} 张`, caption: '只统计有作品产出的日期', icon: Zap },
  { key: 'range', title: '统计范围', value: timelineSummary.value.lastDate || '暂无', caption: timelineSummary.value.firstDate ? `${timelineSummary.value.firstDate} 至 ${timelineSummary.value.lastDate}` : '等待更多作品', icon: Clock3 },
])
</script>

<template>
  <div class="h-full overflow-y-auto bg-background">
    <div class="mx-auto flex max-w-[1460px] flex-col gap-6 p-6 pb-10">
      <section class="space-y-6">
        <Card class="rounded-[28px] border-border/70 bg-card p-6 shadow-none">
          <div class="flex flex-col gap-5">
            <div class="flex flex-col gap-4 xl:flex-row xl:items-end xl:justify-between">
              <div class="space-y-3">
                <Badge variant="outline" class="w-fit rounded-full px-3 py-1 text-xs font-medium">
                  <History class="mr-2 h-3.5 w-3.5" />
                  {{ TEXT.badge }}
                </Badge>
                <div class="space-y-2">
                  <h1 class="text-2xl font-semibold tracking-tight text-foreground">{{ TEXT.title }}</h1>
                  <p class="max-w-3xl text-[13px] leading-6 text-muted-foreground">{{ TEXT.desc }}</p>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-3">
                <div class="inline-flex rounded-full border border-border bg-background p-1">
                  <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="timelineMode === 'day' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="timelineMode = 'day'">按日</button>
                  <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="timelineMode === 'month' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="timelineMode = 'month'">按月</button>
                </div>
                <Button variant="outline" class="h-10 rounded-full px-4 shadow-none" @click="refreshAll">
                  <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': loading }" />
                  {{ TEXT.refresh }}
                </Button>
              </div>
            </div>

            <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              <article v-for="card in statCards" :key="card.key" class="rounded-2xl border border-border bg-background p-4">
                <div class="flex items-start justify-between gap-3">
                  <div class="space-y-2">
                    <p class="text-[13px] text-muted-foreground">{{ card.title }}</p>
                    <p class="text-xl font-semibold tracking-tight text-foreground">{{ card.value }}</p>
                  </div>
                  <component :is="card.icon" class="h-5 w-5 text-muted-foreground" />
                </div>
                <p class="mt-4 text-[11px] leading-5 text-muted-foreground">{{ card.caption }}</p>
              </article>
            </div>
          </div>
        </Card>

        <div class="grid gap-6 xl:grid-cols-[minmax(0,0.96fr)_minmax(360px,0.84fr)]">
          <Card class="rounded-[28px] border-border/70 bg-card p-4 shadow-none">
            <div v-if="latestCreation" class="grid gap-4 md:grid-cols-[220px_minmax(0,1fr)]">
              <div class="overflow-hidden rounded-2xl border border-border bg-background">
                <img :src="latestCreation.path" :alt="latestCreation.name" class="h-full min-h-[220px] w-full object-cover" loading="lazy" />
              </div>

              <div class="flex flex-col justify-between gap-4 rounded-2xl border border-border bg-background p-5">
                <div class="space-y-2">
                  <p class="text-[13px] text-muted-foreground">最近作品</p>
                  <h2 class="line-clamp-2 text-lg font-semibold tracking-tight text-foreground">{{ latestCreation.name }}</h2>
                  <p class="text-[13px] leading-5 text-muted-foreground">
                    {{ dateTimeFormatter.format(new Date(latestCreation.modTime)) }}
                  </p>
                </div>

                <div class="flex flex-wrap gap-2">
                  <Button variant="outline" class="h-10 rounded-full px-4 shadow-none" @click="openTimelineLightbox([latestCreation], 0)">
                    <Eye class="mr-2 h-4 w-4" />
                    {{ TEXT.preview }}
                  </Button>
                  <Button variant="outline" class="h-10 rounded-full px-4 shadow-none" @click="jumpToTimelineKey(formatDateKey(new Date(latestCreation.modTime)))">
                    回到当天
                  </Button>
                </div>
              </div>
            </div>

            <div v-else class="flex min-h-[220px] items-center justify-center rounded-2xl border border-dashed border-border text-sm text-muted-foreground">
              {{ TEXT.emptyLatest }}
            </div>
          </Card>

          <div class="grid gap-3 md:grid-cols-2">
            <Card v-for="item in insightCards" :key="item.key" class="rounded-[24px] border-border/70 bg-card p-5 shadow-none">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-[13px] text-muted-foreground">{{ item.title }}</p>
                  <p class="mt-2 text-[1.6rem] font-semibold tracking-tight text-foreground">{{ item.value }}</p>
                </div>
                <component :is="item.icon" class="h-5 w-5 text-muted-foreground" />
              </div>
              <p class="mt-4 text-[11px] leading-5 text-muted-foreground">{{ item.caption }}</p>
            </Card>
          </div>
        </div>
      </section>

      <section class="space-y-6">
        <Card class="rounded-[28px] border-border/70 bg-card p-6 shadow-none">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div class="space-y-2">
              <div class="flex items-center gap-2 text-muted-foreground">
                <BarChart3 class="h-4 w-4" />
                <span class="text-[13px] font-medium">{{ TEXT.trendTitle }}</span>
              </div>
              <h2 class="text-xl font-semibold tracking-tight text-foreground">{{ TEXT.trendTitle }}</h2>
              <p class="max-w-2xl text-[13px] leading-5 text-muted-foreground">{{ TEXT.trendDesc }}</p>
            </div>

            <div class="inline-flex rounded-full border border-border bg-background p-1">
              <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="viewMode === 'year' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="viewMode = 'year'">按年</button>
              <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="viewMode === 'month' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="viewMode = 'month'">按月</button>
              <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="viewMode === 'day' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="viewMode = 'day'">按日</button>
            </div>
          </div>

          <div v-if="!trendPoints.length" class="mt-6 flex h-[280px] items-center justify-center rounded-2xl border border-dashed border-border text-sm text-muted-foreground">
            {{ TEXT.emptyChart }}
          </div>

          <div v-else class="mt-6 grid gap-4">
            <div ref="chartHost" class="relative overflow-x-auto" @mouseleave="hoveredTrendKey = ''">
              <div
                v-if="activeTrendPoint"
                class="pointer-events-none absolute z-10 min-w-[148px] rounded-2xl border border-border/80 bg-background/95 px-4 py-3 shadow-sm backdrop-blur"
                :style="trendTooltipStyle"
              >
                <p class="text-[11px] font-medium text-muted-foreground">{{ activeTrendPoint.label }}</p>
                <div class="mt-2 flex items-end gap-2">
                  <p class="text-xl font-semibold leading-none text-foreground">{{ activeTrendPoint.count }}</p>
                  <span class="pb-0.5 text-[11px] text-muted-foreground">张作品</span>
                </div>
              </div>

              <svg :width="chartCanvasWidth" :height="chartHeight" class="min-w-full overflow-visible">
                <g>
                  <line v-for="tick in yTicks" :key="tick.value" :x1="chartPadding.left" :x2="chartCanvasWidth - chartPadding.right" :y1="tick.y" :y2="tick.y" stroke="currentColor" class="text-border/70" stroke-dasharray="4 6" />
                  <text v-for="tick in yTicks" :key="`label-${tick.value}`" :x="chartPadding.left - 10" :y="tick.y + 4" text-anchor="end" class="fill-muted-foreground text-[11px]">{{ tick.value }}</text>
                </g>
                <g v-if="activeTrendPoint">
                  <line :x1="activeTrendPoint.x" :x2="activeTrendPoint.x" :y1="chartPadding.top" :y2="chartPadding.top + chartInnerHeight" stroke="currentColor" class="text-border" stroke-dasharray="4 6" />
                </g>
                <path :d="trendAreaPath" class="fill-muted/35" />
                <path :d="trendLinePath" fill="none" stroke="currentColor" class="text-foreground" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
                <rect
                  v-for="zone in trendHoverZones"
                  :key="`${zone.key}-hover`"
                  :x="zone.startX"
                  :y="chartPadding.top"
                  :width="zone.width"
                  :height="chartInnerHeight"
                  class="fill-transparent"
                  @mouseenter="hoveredTrendKey = zone.key"
                />
                <circle v-for="point in trendPoints" :key="`${point.key}-dot`" :cx="point.x" :cy="point.y" :r="activeTrendPoint?.key === point.key ? 5 : 3.5" class="fill-background stroke-foreground" stroke-width="2" @mouseenter="hoveredTrendKey = point.key" @mouseleave="hoveredTrendKey = ''" />
                <text v-for="(point, index) in trendPoints" v-show="index % chartLabelStep === 0 || index === trendPoints.length - 1" :key="`${point.key}-label`" :x="point.x" :y="chartHeight - 12" text-anchor="middle" class="fill-muted-foreground text-[11px]">{{ point.axisLabel }}</text>
              </svg>
            </div>

            <div class="grid gap-3 md:grid-cols-3">
              <div class="rounded-2xl border border-border bg-background/70 p-4">
                <div class="mb-2 flex items-center gap-2 text-[13px] font-medium text-muted-foreground">
                  <TrendingUp class="h-4 w-4" />
                  当前焦点
                </div>
                <p class="text-[13px] font-medium text-foreground">{{ activeTrendPoint?.label || '暂无数据' }}</p>
                <p class="mt-2 text-2xl font-semibold tracking-tight text-foreground">{{ activeTrendPoint?.count ?? 0 }}</p>
              </div>

              <div class="rounded-2xl border border-border bg-background/70 p-4">
                <div class="mb-2 flex items-center gap-2 text-[13px] font-medium text-muted-foreground">
                  <Clock3 class="h-4 w-4" />
                  统计范围
                </div>
                <p class="text-[13px] leading-5 text-foreground">{{ timelineSummary.firstDate && timelineSummary.lastDate ? `${timelineSummary.firstDate} 至 ${timelineSummary.lastDate}` : '暂无记录' }}</p>
              </div>

              <div class="rounded-2xl border border-border bg-background/70 p-4">
                <div class="mb-2 flex items-center gap-2 text-[13px] font-medium text-muted-foreground">
                  <BarChart3 class="h-4 w-4" />
                  当前粒度
                </div>
                <p class="text-[13px] leading-5 text-foreground">{{ viewMode === 'year' ? '按年统计' : viewMode === 'month' ? '按月统计' : '按日统计' }}</p>
              </div>
            </div>
          </div>
        </Card>

        <Card class="rounded-[28px] border-border/70 bg-card p-6 shadow-none">
          <div class="space-y-2">
            <div class="flex items-center gap-2 text-muted-foreground">
              <CalendarDays class="h-4 w-4" />
              <span class="text-[13px] font-medium">{{ TEXT.heatmapTitle }}</span>
            </div>
            <h2 class="text-xl font-semibold tracking-tight text-foreground">{{ TEXT.heatmapTitle }}</h2>
            <p class="text-[13px] leading-5 text-muted-foreground">{{ TEXT.heatmapDesc }}</p>
          </div>

          <div class="mt-6 overflow-x-auto pb-2">
            <div class="inline-flex min-w-full gap-2">
              <div v-for="week in heatmapWeeks" :key="week.key" class="flex min-w-[18px] flex-col gap-2">
                <div class="h-4 text-center text-[10px] text-muted-foreground">{{ week.monthLabel }}</div>
                <button
                  v-for="day in week.days"
                  :key="day.key"
                  type="button"
                  class="h-[18px] w-[18px] rounded-md border transition-transform duration-150 hover:scale-110"
                  :class="[heatmapLevelClass(day.count, day.isFuture), day.count && !day.isFuture ? 'shadow-sm' : '']"
                  :title="`${day.label} · ${day.count} 张`"
                  :disabled="!day.count || day.isFuture"
                  @click="jumpToTimelineKey(day.key)"
                />
              </div>
            </div>
          </div>

          <div class="mt-5 grid gap-3 md:grid-cols-3">
            <div class="rounded-2xl border border-border bg-background p-4 text-[13px] text-muted-foreground">颜色越深表示当天越活跃。</div>
            <div class="rounded-2xl border border-border bg-background p-4 text-[13px] text-muted-foreground">点击有记录的格子，可直接跳到当天时间线。</div>
            <div class="rounded-2xl border border-border bg-background p-4 text-[13px] text-muted-foreground">适合快速找出高产日和断档期。</div>
          </div>
        </Card>
      </section>

      <section id="stats-timeline-section" class="grid gap-6 xl:h-[calc(100vh-12rem)] xl:grid-cols-[minmax(0,1.08fr)_380px]">
        <Card class="rounded-[28px] border-border/70 bg-card p-6 shadow-none xl:flex xl:h-full xl:flex-col xl:overflow-hidden">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <div class="space-y-2">
              <div class="flex items-center gap-2 text-muted-foreground">
                <History class="h-4 w-4" />
                <span class="text-[13px] font-medium">{{ TEXT.timelineTitle }}</span>
              </div>
              <h2 class="text-xl font-semibold tracking-tight text-foreground">{{ TEXT.timelineTitle }}</h2>
              <p class="max-w-3xl text-[13px] leading-5 text-muted-foreground">{{ TEXT.timelineDesc }}</p>
            </div>

            <div class="inline-flex rounded-full border border-border bg-background p-1">
              <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="timelineMode === 'day' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="timelineMode = 'day'">按日</button>
              <button type="button" class="rounded-full px-4 py-2 text-[13px] transition-colors" :class="timelineMode === 'month' ? 'bg-foreground text-background' : 'text-muted-foreground hover:text-foreground'" @click="timelineMode = 'month'">按月</button>
            </div>
          </div>

          <div v-if="!timelineEntries.length" class="mt-6 flex h-[340px] items-center justify-center rounded-2xl border border-dashed border-border text-sm text-muted-foreground xl:flex-1">
            {{ TEXT.emptyTimeline }}
          </div>

          <div v-else class="mt-8 space-y-8 xl:min-h-0 xl:flex-1 xl:overflow-y-auto xl:pr-2 custom-scrollbar">
            <div v-for="group in timelineYearGroups" :key="group.year" class="grid gap-4 md:grid-cols-[72px_minmax(0,1fr)]">
              <div class="pt-2 text-[13px] font-medium tracking-wide text-muted-foreground">{{ group.year }}</div>
              <div class="relative space-y-3 pl-6">
                <div class="absolute bottom-0 left-[8px] top-0 w-px bg-border" />

                <button
                  v-for="entry in group.entries"
                  :key="entry.key"
                  type="button"
                  class="relative block w-full rounded-2xl border px-4 py-4 text-left transition-colors"
                  :class="activeTimelineKey === entry.key ? 'border-foreground/20 bg-muted/35' : 'border-border bg-background hover:bg-muted/20'"
                  @click="activeTimelineKey = entry.key"
                >
                  <div class="absolute left-[-22px] top-5 h-4 w-4 rounded-full border-4 border-background bg-foreground" />

                  <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                    <div class="min-w-0 space-y-3">
                      <div class="flex flex-wrap items-center gap-2">
                        <Badge variant="outline" class="rounded-full px-3 py-1 text-[11px]">{{ entry.shortLabel }}</Badge>
                        <Badge variant="outline" class="rounded-full px-3 py-1 text-[11px]">{{ entry.count }} 张</Badge>
                        <Badge variant="outline" class="rounded-full px-3 py-1 text-[11px]">{{ entry.activeDaysCount }} 天</Badge>
                      </div>
                      <div>
                        <h3 class="truncate text-base font-medium text-foreground">{{ entry.label }}</h3>
                        <p class="mt-1 text-[13px] text-muted-foreground">{{ entry.caption }}</p>
                      </div>
                    </div>

                    <ChevronDown class="h-5 w-5 shrink-0 text-muted-foreground transition-transform duration-200" :class="{ 'rotate-180': activeTimelineKey === entry.key }" />
                  </div>

                  <div class="mt-4 flex flex-wrap items-center gap-3">
                    <div class="flex -space-x-2">
                      <div v-for="(cover, index) in entry.coverImages" :key="cover.relPath" class="h-10 w-10 overflow-hidden rounded-xl border border-background" :style="{ zIndex: 10 - index }">
                        <img :src="cover.path" :alt="cover.name" class="h-full w-full object-cover" loading="lazy" />
                      </div>
                    </div>
                    <div class="flex flex-wrap gap-2 text-xs text-muted-foreground">
                      <span class="rounded-full border border-border px-3 py-1">{{ formatSize(entry.totalSize) }}</span>
                    </div>
                  </div>
                </button>
              </div>
            </div>

            <div class="flex justify-center pt-1">
              <Button v-if="visibleTimelineCount < timelineEntries.length" variant="outline" class="h-10 rounded-full px-6 shadow-none" @click="visibleTimelineCount += 12">
                <ChevronDown class="mr-2 h-4 w-4" />
                {{ TEXT.loadMoreEntries }}
              </Button>
            </div>
          </div>
        </Card>

        <Card class="rounded-[28px] border-border/70 bg-card p-6 shadow-none xl:flex xl:h-full xl:flex-col xl:overflow-hidden">
          <div class="space-y-2">
            <div class="flex items-center gap-2 text-muted-foreground">
              <ArrowUpRight class="h-4 w-4" />
              <span class="text-[13px] font-medium">{{ TEXT.detailTitle }}</span>
            </div>
            <h2 class="text-xl font-semibold tracking-tight text-foreground">{{ activeTimelineEntry?.label || '选择一个时间节点' }}</h2>
            <p class="text-[13px] leading-5 text-muted-foreground">
              {{ activeTimelineEntry ? `${dateTimeFormatter.format(activeTimelineEntry.earliestDate)} - ${dateTimeFormatter.format(activeTimelineEntry.latestDate)}` : TEXT.emptyDetail }}
            </p>
          </div>

          <div v-if="!activeTimelineEntry" class="mt-6 flex h-[220px] items-center justify-center rounded-2xl border border-dashed border-border text-sm text-muted-foreground xl:flex-1">
            {{ TEXT.emptyDetail }}
          </div>

          <div v-else class="mt-6 space-y-4 xl:min-h-0 xl:flex xl:flex-1 xl:flex-col">
            <div class="grid gap-3 sm:grid-cols-2">
              <div class="rounded-2xl border border-border bg-background p-4">
                <div class="flex items-center gap-2 text-[13px] text-muted-foreground">
                  <ImageIcon class="h-4 w-4" />
                  作品规模
                </div>
                <p class="mt-3 text-2xl font-semibold tracking-tight text-foreground">{{ activeTimelineEntry.count }}</p>
              </div>

              <div class="rounded-2xl border border-border bg-background p-4">
                <div class="flex items-center gap-2 text-[13px] text-muted-foreground">
                  <HardDrive class="h-4 w-4" />
                  空间占用
                </div>
                <p class="mt-3 text-2xl font-semibold tracking-tight text-foreground">{{ formatSize(activeTimelineEntry.totalSize) }}</p>
              </div>
            </div>

            <div class="xl:min-h-0 xl:flex-1 xl:overflow-y-auto xl:pr-2 custom-scrollbar">
              <div v-if="visibleTimelineImagesList.length" class="grid gap-3">
                <article v-for="(image, index) in visibleTimelineImagesList" :key="image.relPath" class="overflow-hidden rounded-2xl border border-border bg-background">
                  <button type="button" class="block w-full" @click="openTimelineLightbox(activeTimelineEntry.images, index)">
                    <div class="aspect-[16/9] overflow-hidden bg-muted/20">
                      <img :src="image.path" :alt="image.name" class="h-full w-full object-cover" loading="lazy" />
                    </div>
                  </button>

                  <div class="space-y-3 p-4">
                    <div>
                      <p class="truncate text-[13px] font-medium text-foreground">{{ image.name }}</p>
                      <p class="mt-1 text-xs text-muted-foreground">{{ dateTimeFormatter.format(new Date(image.modTime)) }}</p>
                    </div>

                    <div class="flex flex-wrap gap-2">
                      <Button variant="outline" size="sm" class="h-8 rounded-full px-3 text-xs shadow-none" @click="openTimelineLightbox(activeTimelineEntry.images, index)">{{ TEXT.preview }}</Button>
                      <Button variant="outline" size="sm" class="h-8 rounded-full px-3 text-xs shadow-none" @click="openImageLocation(image)">
                        <FolderOpen class="mr-1.5 h-3.5 w-3.5" />
                        {{ TEXT.locate }}
                      </Button>
                      <Button variant="outline" size="sm" class="h-8 rounded-full border-red-500/30 px-3 text-xs text-red-500 shadow-none hover:bg-red-500/8 hover:text-red-400" @click="handleTimelineDelete(image)">{{ TEXT.delete }}</Button>
                    </div>
                  </div>
                </article>

                <div class="flex justify-center">
                  <Button v-if="activeTimelineEntry.images.length > visibleTimelineImages" variant="outline" class="h-10 rounded-full px-6 shadow-none" @click="visibleTimelineImages += 8">
                    {{ TEXT.loadMoreImages }}
                  </Button>
                </div>
              </div>

              <div v-else class="flex h-32 items-center justify-center rounded-2xl border border-dashed border-border text-sm text-muted-foreground">
                {{ TEXT.emptyImages }}
              </div>
            </div>
          </div>
        </Card>
      </section>

      <Lightbox
        :is-open="lightboxOpen"
        :image="lightboxImage"
        :images="lightboxImages"
        :current-index="lightboxIndex"
        :favorite-groups="favoriteGroups"
        :tags="tags"
        :image-tags="imageTags"
        :image-notes="imageNotes"
        @close="lightboxOpen = false"
        @navigate="handleLightboxNavigate"
        @toggle-favorite="toggleFavorite"
        @add-tag="addTagToImage"
        @remove-tag="removeTagFromImage"
        @delete="handleTimelineDelete"
        @open-location="openImageLocation"
        @favorite-groups-changed="fetchImages"
      />
    </div>
  </div>
</template>
