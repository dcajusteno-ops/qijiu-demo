import { computed, ref } from 'vue'
import { buildDateCountMap, getDatePresetLabel, matchesDatePreset } from '@/lib/dateWorkbench'
import { buildGroupedFilterOptions, getImageDateKey, normalizeAssetKey } from './useGalleryHelpers'

const activeDatePreset = ref('all')
const activeDateStart = ref('')
const activeDateEnd = ref('')
const activeModelFilter = ref('')
const activeLoraFilter = ref('')

export function useWorkbenchFilters(images, workbenchAggregate) {
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

  return {
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
  }
}
