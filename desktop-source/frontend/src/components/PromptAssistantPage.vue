<script setup>
import { computed, onMounted, ref, watch } from 'vue'
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
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import PromptFilterSelect from '@/components/PromptFilterSelect.vue'
import {
  Bookmark,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Copy,
  Heart,
  Loader2,
  Plus,
  Search,
  Sparkles,
  Trash2,
  Wand2,
  X,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import * as App from '@/api'

const props = defineProps({
  initialPositive: { type: String, default: '' },
  initialNegative: { type: String, default: '' },
  sourcePath: { type: String, default: '' },
  contextLabel: { type: String, default: '' },
  contextRevision: { type: Number, default: 0 },
})

const customPromptSource = '\u6211\u7684\u8bcd\u5e93'
const emptyCategoryValue = '__EMPTY_CATEGORY__'
const emptySubcategoryValue = '__EMPTY_SUBCATEGORY__'

const positivePresetPacks = [
  { id: 'portrait', name: '人物基础', terms: ['masterpiece', 'best quality', '1girl', 'detailed face', 'soft lighting'] },
  { id: 'cinematic', name: '电影感', terms: ['cinematic lighting', 'dramatic shadows', 'depth of field', 'film grain', 'high contrast'] },
  { id: 'camera', name: '镜头语言', terms: ['close-up', '85mm lens', 'bokeh', 'dynamic composition', 'sharp focus'] },
  { id: 'illustration', name: '插画细节', terms: ['highly detailed', 'clean lineart', 'delicate texture', 'rich colors', 'beautiful composition'] },
]

const negativePresetPacks = [
  { id: 'common', name: '通用负面', terms: ['low quality', 'worst quality', 'blurry', 'bad anatomy', 'text', 'watermark'] },
  { id: 'handfix', name: '手部修正', terms: ['bad hands', 'extra fingers', 'missing fingers', 'mutated hands', 'poorly drawn hands'] },
  { id: 'facefix', name: '面部修正', terms: ['deformed face', 'bad eyes', 'cross-eyed', 'extra eyes', 'poorly drawn face'] },
  { id: 'artifact', name: '杂项瑕疵', terms: ['jpeg artifacts', 'cropped', 'duplicate', 'out of frame', 'extra limbs'] },
]

const libraryLoading = ref(false)
const libraryError = ref('')
const systemEntries = ref([])
const customEntries = ref([])
const defaultAssistantState = () => ({
  favoriteIds: [],
  recentIds: [],
  activeSource: '',
  activeCategory: '',
  activeSubcategory: '',
  activeScope: '',
  viewMode: 'all',
  activeEditor: 'positive',
  itemsPerPage: 12,
  currentPage: 1,
})

const assistantState = ref(defaultAssistantState())
const stateLoaded = ref(false)
const isApplyingAssistantState = ref(false)

const searchQuery = ref('')
const activeSource = ref('')
const activeCategory = ref('')
const activeSubcategory = ref('')
const activeScope = ref('')
const viewMode = ref('all')
const activeEditor = ref('positive')
const currentPage = ref(1)
const itemsPerPage = ref(12)
const pageSizeOptions = [
  { value: '8', label: '8 / \u9875' },
  { value: '12', label: '12 / \u9875' },
  { value: '24', label: '24 / \u9875' },
  { value: '48', label: '48 / \u9875' },
]

const positiveBase = ref('')
const negativeBase = ref('')
const positiveParts = ref([])
const negativeParts = ref([])
const positiveQuickInput = ref('')
const negativeQuickInput = ref('')

const customPromptForm = ref({
  text_zh: '',
  text_en: '',
  category: '',
  subcategory: '',
  scope: 'default',
})
const customPromptSaving = ref(false)
const customPromptDeleteOpen = ref(false)
const pendingDeleteCustomPrompt = ref(null)

const templateName = ref('')
const templateCategory = ref('')
const templateSaving = ref(false)

const normalizeState = (state = {}) => {
  const nextItemsPerPage = Number(state.itemsPerPage)
  const nextCurrentPage = Number(state.currentPage)

  return {
    ...defaultAssistantState(),
    favoriteIds: Array.isArray(state.favoriteIds) ? [...new Set(state.favoriteIds.filter(Boolean))] : [],
    recentIds: Array.isArray(state.recentIds) ? [...new Set(state.recentIds.filter(Boolean))].slice(0, 120) : [],
    activeSource: String(state.activeSource || '').trim(),
    activeCategory: String(state.activeCategory || '').trim(),
    activeSubcategory: String(state.activeSubcategory || '').trim(),
    activeScope: String(state.activeScope || '').trim(),
    viewMode: ['all', 'favorites', 'recent'].includes(state.viewMode) ? state.viewMode : 'all',
    activeEditor: state.activeEditor === 'negative' ? 'negative' : 'positive',
    itemsPerPage: [8, 12, 24, 48].includes(nextItemsPerPage) ? nextItemsPerPage : 12,
    currentPage: Number.isFinite(nextCurrentPage) && nextCurrentPage > 0 ? Math.trunc(nextCurrentPage) : 1,
  }
}

const compareText = (left, right) => String(left || '').localeCompare(String(right || ''), 'zh-CN')

const normalizeTextKey = (value) => String(value || '').trim().toLowerCase().replace(/\s+/g, ' ')

const createPart = (entry) => ({
  key: `${entry.id || 'manual'}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
  entryId: entry.id || '',
  text: String(entry.text_en || '').trim(),
  textZh: String(entry.text_zh || '').trim(),
  source: entry.source || '',
  category: entry.category || '',
})

const createManualPart = (text) => ({
  key: `manual-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
  entryId: '',
  text: String(text || '').trim(),
  textZh: '',
  source: 'manual',
  category: '',
})

const parseManualTerms = (text) => {
  return String(text || '')
    .split(/[,\n闁挎稑鐭夌槐?]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

const loadPromptLibrary = async () => {
  libraryLoading.value = true
  libraryError.value = ''
  try {
    const [system, custom] = await Promise.all([
      App.GetPromptLibraryEntries(),
      App.GetCustomPromptEntries(),
    ])
    systemEntries.value = Array.isArray(system) ? system : []
    customEntries.value = Array.isArray(custom) ? custom : []
  } catch (error) {
    libraryError.value = String(error?.message || error || '加载提示词词库失败')
    systemEntries.value = []
    customEntries.value = []
  } finally {
    libraryLoading.value = false
  }
}

const loadAssistantState = async () => {
  try {
    const state = normalizeState(await App.GetPromptAssistantState())
    assistantState.value = state
    isApplyingAssistantState.value = true
    activeSource.value = state.activeSource
    activeCategory.value = state.activeCategory
    activeSubcategory.value = state.activeSubcategory
    activeScope.value = state.activeScope
    viewMode.value = state.viewMode
    activeEditor.value = state.activeEditor
    itemsPerPage.value = state.itemsPerPage
    currentPage.value = state.currentPage
  } catch {
    assistantState.value = normalizeState()
  } finally {
    isApplyingAssistantState.value = false
    stateLoaded.value = true
  }
}

const persistAssistantState = async () => {
  if (!stateLoaded.value) return
  try {
    const nextState = normalizeState({
      ...assistantState.value,
      activeSource: activeSource.value,
      activeCategory: activeCategory.value,
      activeSubcategory: activeSubcategory.value,
      activeScope: activeScope.value,
      viewMode: viewMode.value,
      activeEditor: activeEditor.value,
      itemsPerPage: itemsPerPage.value,
      currentPage: currentPage.value,
    })
    const saved = await App.SavePromptAssistantState(nextState)
    assistantState.value = normalizeState(saved)
  } catch (error) {
    toast.error(`保存提示词状态失败：${error?.message || error}`)
  }
}

const rememberEntry = async (entryId) => {
  if (!entryId) return
  const ids = assistantState.value.recentIds.filter((id) => id !== entryId)
  assistantState.value = {
    ...assistantState.value,
    recentIds: [entryId, ...ids].slice(0, 120),
  }
  await persistAssistantState()
}

const addToBuilder = async (target, entry) => {
  const part = createPart(entry)
  if (!part.text) {
    toast.error('词条内容为空，无法加入当前编辑区')
    return
  }

  if (target === 'positive') {
    positiveParts.value = [...positiveParts.value, part]
  } else {
    negativeParts.value = [...negativeParts.value, part]
  }

  await rememberEntry(entry.id)
}

const addManualTerms = (target) => {
  const source = target === 'positive' ? positiveQuickInput.value : negativeQuickInput.value
  const terms = parseManualTerms(source)
  if (terms.length === 0) {
    toast.error('没有可添加的手动词条')
    return
  }

  const nextParts = terms.map(createManualPart)
  if (target === 'positive') {
    positiveParts.value = [...positiveParts.value, ...nextParts]
    positiveQuickInput.value = ''
  } else {
    negativeParts.value = [...negativeParts.value, ...nextParts]
    negativeQuickInput.value = ''
  }
}

const applyPresetPack = (target, pack) => {
  const sourceParts = target === 'positive' ? positiveParts.value : negativeParts.value
  const seen = new Set(sourceParts.map((item) => normalizeTextKey(item.text)))
  const nextParts = []

  for (const term of pack.terms) {
    const key = normalizeTextKey(term)
    if (!key || seen.has(key)) continue
    seen.add(key)
    nextParts.push(createManualPart(term))
  }

  if (nextParts.length === 0) {
    toast.info(pack.name + ' 词包没有新增词条')
    return
  }

  if (target === 'positive') {
    positiveParts.value = [...positiveParts.value, ...nextParts]
  } else {
    negativeParts.value = [...negativeParts.value, ...nextParts]
  }
}

const removePart = (target, key) => {
  if (target === 'positive') {
    positiveParts.value = positiveParts.value.filter((item) => item.key !== key)
    return
  }
  negativeParts.value = negativeParts.value.filter((item) => item.key !== key)
}

const movePart = (target, index, direction) => {
  const list = target === 'positive' ? [...positiveParts.value] : [...negativeParts.value]
  const nextIndex = direction === 'up' ? index - 1 : index + 1
  if (nextIndex < 0 || nextIndex >= list.length) return
  ;[list[index], list[nextIndex]] = [list[nextIndex], list[index]]
  if (target === 'positive') {
    positiveParts.value = list
    return
  }
  negativeParts.value = list
}

const dedupeParts = (target) => {
  const source = target === 'positive' ? positiveParts.value : negativeParts.value
  const seen = new Set()
  const next = source.filter((item) => {
    const key = normalizeTextKey(item.text)
    if (!key || seen.has(key)) return false
    seen.add(key)
    return true
  })
  if (target === 'positive') {
    positiveParts.value = next
    return
  }
  negativeParts.value = next
}

const clearBuilder = (target) => {
  if (target === 'positive') {
    positiveBase.value = ''
    positiveParts.value = []
    positiveQuickInput.value = ''
    return
  }
  negativeBase.value = ''
  negativeParts.value = []
  negativeQuickInput.value = ''
}

const resetFilters = () => {
  searchQuery.value = ''
  activeSource.value = ''
  activeCategory.value = ''
  activeSubcategory.value = ''
  activeScope.value = ''
  viewMode.value = 'all'
}

const resetCustomPromptForm = () => {
  customPromptForm.value = {
    text_zh: '',
    text_en: '',
    category: activeCategory.value && activeCategory.value !== emptyCategoryValue ? activeCategory.value : '',
    subcategory: '',
    scope: activeScope.value || 'default',
  }
}

const applyIncomingContext = () => {
  positiveBase.value = String(props.initialPositive || '').trim()
  negativeBase.value = String(props.initialNegative || '').trim()
  positiveParts.value = []
  negativeParts.value = []
  positiveQuickInput.value = ''
  negativeQuickInput.value = ''
  templateName.value = ''
  templateCategory.value = ''
  resetCustomPromptForm()
}

const combinePrompt = (base, parts) => {
  const segments = [base, ...parts.map((item) => item.text)]
    .map((item) => String(item || '').trim())
    .filter(Boolean)
  return segments.join(', ')
}

const favoriteIdSet = computed(() => new Set(assistantState.value.favoriteIds))
const recentIdSet = computed(() => new Set(assistantState.value.recentIds))
const recentIdRank = computed(() => new Map(assistantState.value.recentIds.map((id, index) => [id, index])))

const libraryEntries = computed(() => [...customEntries.value, ...systemEntries.value])
const libraryEntryIdSet = computed(() => new Set(libraryEntries.value.map((entry) => entry.id).filter(Boolean)))
const finalPositivePrompt = computed(() => combinePrompt(positiveBase.value, positiveParts.value))
const finalNegativePrompt = computed(() => combinePrompt(negativeBase.value, negativeParts.value))
const combinedPrompt = computed(() => {
  return [finalPositivePrompt.value, finalNegativePrompt.value]
    .map((item) => String(item || '').trim())
    .filter(Boolean)
    .join('\n\n')
})

const totalEntryCount = computed(() => libraryEntries.value.length)
const favoriteEntryCount = computed(() => assistantState.value.favoriteIds.filter((id) => libraryEntryIdSet.value.has(id)).length)
const recentEntryCount = computed(() => assistantState.value.recentIds.filter((id) => libraryEntryIdSet.value.has(id)).length)
const customEntryCount = computed(() => customEntries.value.length)

const syncAssistantStateWithLibrary = async () => {
  if (!stateLoaded.value) return

  const nextFavoriteIds = assistantState.value.favoriteIds.filter((id) => libraryEntryIdSet.value.has(id))
  const nextRecentIds = assistantState.value.recentIds.filter((id) => libraryEntryIdSet.value.has(id))
  const favoriteChanged = nextFavoriteIds.length !== assistantState.value.favoriteIds.length
  const recentChanged = nextRecentIds.length !== assistantState.value.recentIds.length

  if (!favoriteChanged && !recentChanged) return

  assistantState.value = normalizeState({
    ...assistantState.value,
    favoriteIds: nextFavoriteIds,
    recentIds: nextRecentIds,
  })
  await persistAssistantState()
}

const filteredForCategoryOptions = computed(() => {
  let list = [...libraryEntries.value]
  if (viewMode.value === 'favorites') {
    list = list.filter((entry) => favoriteIdSet.value.has(entry.id))
  } else if (viewMode.value === 'recent') {
    list = list.filter((entry) => recentIdSet.value.has(entry.id))
  }
  if (activeSource.value) {
    list = list.filter((entry) => entry.source === activeSource.value)
  }
  return list
})

const filteredForSubcategoryOptions = computed(() => {
  let list = [...filteredForCategoryOptions.value]
  if (activeCategory.value) {
    list = list.filter((entry) => (
      activeCategory.value === emptyCategoryValue
        ? !entry.category
        : entry.category === activeCategory.value
    ))
  }
  return list
})

const sourceOptions = computed(() => {
  const counter = new Map()
  libraryEntries.value.forEach((entry) => {
    const key = entry.source || '未分类来源'
    counter.set(key, (counter.get(key) || 0) + 1)
  })
  return [...counter.entries()]
    .map(([value, count]) => ({ value, count }))
    .sort((left, right) => compareText(left.value, right.value))
})

const categoryOptions = computed(() => {
  const counter = new Map()
  filteredForCategoryOptions.value.forEach((entry) => {
    const key = entry.category || ''
    counter.set(key, (counter.get(key) || 0) + 1)
  })
  return [...counter.entries()]
    .map(([value, count]) => ({
      value: value || emptyCategoryValue,
      label: value || '未分类',
      count,
    }))
    .sort((left, right) => compareText(left.label, right.label))
})

const subcategoryOptions = computed(() => {
  const counter = new Map()
  filteredForSubcategoryOptions.value.forEach((entry) => {
    const key = entry.subcategory || ''
    counter.set(key, (counter.get(key) || 0) + 1)
  })
  return [...counter.entries()]
    .map(([value, count]) => ({
      value: value || emptySubcategoryValue,
      label: value || '未设子分类',
      count,
    }))
    .sort((left, right) => compareText(left.label, right.label))
})

const scopeOptions = computed(() => {
  const counter = new Map()
  libraryEntries.value.forEach((entry) => {
    const key = entry.scope || 'default'
    counter.set(key, (counter.get(key) || 0) + 1)
  })
  return [...counter.entries()]
    .map(([value, count]) => ({ value, count }))
    .sort((left, right) => compareText(left.value, right.value))
})

const sourceSelectOptions = computed(() => ([
  { value: '', label: '\u5168\u90E8\u6765\u6E90' },
  ...sourceOptions.value.map((item) => ({
    value: item.value,
    label: item.value + ' (' + item.count + ')',
  })),
]))

const categorySelectOptions = computed(() => ([
  { value: '', label: '\u5168\u90E8\u5206\u7C7B' },
  ...categoryOptions.value.map((item) => ({
    value: item.value,
    label: item.label + ' (' + item.count + ')',
  })),
]))

const subcategorySelectOptions = computed(() => ([
  { value: '', label: '\u5168\u90E8\u5B50\u5206\u7C7B' },
  ...subcategoryOptions.value.map((item) => ({
    value: item.value,
    label: item.label + ' (' + item.count + ')',
  })),
]))

const scopeSelectOptions = computed(() => ([
  { value: '', label: '\u5168\u90E8\u4F5C\u7528\u57DF' },
  ...scopeOptions.value.map((item) => ({
    value: item.value,
    label: item.value + ' (' + item.count + ')',
  })),
]))

const quickCategoryOptions = computed(() => {
  return [...categoryOptions.value]
    .sort((left, right) => right.count - left.count)
    .slice(0, 10)
})

const filteredEntries = computed(() => {
  const keywords = searchQuery.value
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)

  let list = [...libraryEntries.value]

  if (viewMode.value === 'favorites') {
    list = list.filter((entry) => favoriteIdSet.value.has(entry.id))
  } else if (viewMode.value === 'recent') {
    list = list.filter((entry) => recentIdSet.value.has(entry.id))
  }

  if (activeSource.value) {
    list = list.filter((entry) => entry.source === activeSource.value)
  }
  if (activeCategory.value) {
    list = list.filter((entry) => (
      activeCategory.value === emptyCategoryValue
        ? !entry.category
        : entry.category === activeCategory.value
    ))
  }
  if (activeSubcategory.value) {
    list = list.filter((entry) => (
      activeSubcategory.value === emptySubcategoryValue
        ? !entry.subcategory
        : entry.subcategory === activeSubcategory.value
    ))
  }
  if (activeScope.value) {
    list = list.filter((entry) => entry.scope === activeScope.value)
  }
  if (keywords.length > 0) {
    list = list.filter((entry) => {
      const haystack = String(entry.search_text || '').toLowerCase()
      return keywords.every((keyword) => haystack.includes(keyword))
    })
  }

  if (viewMode.value === 'recent') {
    list.sort((left, right) => (recentIdRank.value.get(left.id) ?? 999999) - (recentIdRank.value.get(right.id) ?? 999999))
  }

  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredEntries.value.length / itemsPerPage.value)))
const paginatedEntries = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredEntries.value.slice(start, start + itemsPerPage.value)
})
const pageStart = computed(() => (filteredEntries.value.length === 0 ? 0 : (currentPage.value - 1) * itemsPerPage.value + 1))
const pageEnd = computed(() => Math.min(currentPage.value * itemsPerPage.value, filteredEntries.value.length))

const visiblePages = computed(() => {
  const pages = []
  const total = totalPages.value
  const current = currentPage.value

  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
    return pages
  }

  pages.push(1)
  if (current > 4) pages.push('...')

  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)

  if (current < total - 3) pages.push('...')
  pages.push(total)
  return pages
})

const resultSummary = computed(() => {
  if (filteredEntries.value.length === 0) return '没有匹配结果'
  return '当前第 ' + currentPage.value + ' / ' + totalPages.value + ' 页，显示 ' + pageStart.value + '-' + pageEnd.value + ' / ' + filteredEntries.value.length
})

const setQuickCategory = (value) => {
  activeCategory.value = value
}

const isCustomEntry = (entry) => entry.source === customPromptSource

const toggleFavorite = async (entryId) => {
  const favorites = favoriteIdSet.value.has(entryId)
    ? assistantState.value.favoriteIds.filter((id) => id !== entryId)
    : [entryId, ...assistantState.value.favoriteIds.filter((id) => id !== entryId)]

  assistantState.value = {
    ...assistantState.value,
    favoriteIds: favorites.slice(0, 512),
  }
  await persistAssistantState()
}

const copyText = async (text, label) => {
  if (!String(text || '').trim()) {
    toast.error(`${label}为空，无法复制`)
    return
  }
  try {
    await App.CopyText(text)
    toast.success(`${label}已复制到剪贴板`)
  } catch (error) {
    toast.error(`复制失败：${error?.message || error}`)
  }
}

const makeTemplateName = (type) => {
  const customName = templateName.value.trim()
  if (customName) return customName
  const labelMap = {
    positive: '提示词组合-正向',
    negative: '提示词组合-反向',
    other: '提示词组合-完整',
  }
  const timestamp = new Date().toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).replace(/[/: ]/g, '-')
  return `${labelMap[type] || '提示词组合'}-${timestamp}`
}

const saveAsTemplate = async (type) => {
  const contentMap = {
    positive: finalPositivePrompt.value,
    negative: finalNegativePrompt.value,
    other: combinedPrompt.value,
  }

  const content = String(contentMap[type] || '').trim()
  if (!content) {
    toast.error('当前没有可保存的提示词内容')
    return
  }

  templateSaving.value = true
  try {
    await App.AddPromptTemplate({
      name: makeTemplateName(type),
      content,
      type,
      category: templateCategory.value.trim(),
      sourcePath: props.sourcePath || '',
    })
    toast.success('提示词模板已保存')
  } catch (error) {
    toast.error(`保存模板失败：${error?.message || error}`)
  } finally {
    templateSaving.value = false
  }
}

const saveCustomPrompt = async () => {
  const payload = {
    text_zh: customPromptForm.value.text_zh.trim(),
    text_en: customPromptForm.value.text_en.trim(),
    category: customPromptForm.value.category.trim(),
    subcategory: customPromptForm.value.subcategory.trim(),
    scope: customPromptForm.value.scope.trim() || 'default',
  }

  if (!payload.text_zh && !payload.text_en) {
    toast.error('请至少填写中文或英文提示词')
    return
  }

  customPromptSaving.value = true
  try {
    const saved = await App.AddCustomPromptEntry(payload)
    customEntries.value = [saved, ...customEntries.value]
    activeSource.value = customPromptSource
    currentPage.value = 1
    resetCustomPromptForm()
    await persistAssistantState()
    toast.success('自定义提示词已保存')
  } catch (error) {
    toast.error(String(error?.message || error || '保存自定义提示词失败'))
  } finally {
    customPromptSaving.value = false
  }
}

const requestDeleteCustomPrompt = (entry) => {
  if (!entry?.id) return
  pendingDeleteCustomPrompt.value = entry
  customPromptDeleteOpen.value = true
}

const deleteCustomPrompt = async () => {
  const entry = pendingDeleteCustomPrompt.value
  if (!entry?.id) return
  try {
    await App.DeleteCustomPromptEntry(entry.id)
    customEntries.value = customEntries.value.filter((item) => item.id !== entry.id)
    assistantState.value = normalizeState({
      ...assistantState.value,
      favoriteIds: assistantState.value.favoriteIds.filter((id) => id !== entry.id),
      recentIds: assistantState.value.recentIds.filter((id) => id !== entry.id),
    })
    await persistAssistantState()
    await loadAssistantState()
    toast.success('\u81ea\u5b9a\u4e49\u63d0\u793a\u8bcd\u5df2\u5220\u9664')
  } catch (error) {
    toast.error(String(error?.message || error || '\u5220\u9664\u81ea\u5b9a\u4e49\u63d0\u793a\u8bcd\u5931\u8d25'))
  } finally {
    customPromptDeleteOpen.value = false
    pendingDeleteCustomPrompt.value = null
  }
}

watch(activeSource, () => {
  if (activeCategory.value && !categoryOptions.value.some((item) => item.value === activeCategory.value)) {
    activeCategory.value = ''
  }
  if (activeSubcategory.value && !subcategoryOptions.value.some((item) => item.value === activeSubcategory.value)) {
    activeSubcategory.value = ''
  }
})

watch(activeCategory, () => {
  if (activeSubcategory.value && !subcategoryOptions.value.some((item) => item.value === activeSubcategory.value)) {
    activeSubcategory.value = ''
  }
})

watch([searchQuery, activeSource, activeCategory, activeSubcategory, activeScope, viewMode, itemsPerPage], () => {
  if (isApplyingAssistantState.value) return
  currentPage.value = 1
})

watch(totalPages, (value) => {
  if (currentPage.value > value) {
    currentPage.value = value
  }
})

watch([activeSource, activeCategory, activeSubcategory, activeScope, viewMode, activeEditor, itemsPerPage], () => {
  if (isApplyingAssistantState.value || !stateLoaded.value) return
  void persistAssistantState()
})

watch(currentPage, () => {
  if (isApplyingAssistantState.value || !stateLoaded.value) return
  void persistAssistantState()
})

watch([libraryEntries, stateLoaded], ([, loaded]) => {
  if (!loaded) return
  void syncAssistantStateWithLibrary()
})

watch(() => props.contextRevision, () => {
  applyIncomingContext()
}, { immediate: true })

onMounted(async () => {
  await Promise.all([loadPromptLibrary(), loadAssistantState()])
  await syncAssistantStateWithLibrary()
})
</script>

<template>
  <div class="h-full overflow-y-auto bg-background text-foreground">
    <div class="mx-auto flex w-full max-w-[1580px] flex-col gap-6 px-6 py-6">
      <section class="flex flex-col gap-4 rounded-3xl border border-border/70 bg-card/80 p-5 shadow-sm xl:flex-row xl:items-start xl:justify-between">
        <div class="min-w-0 space-y-2">
          <div class="flex items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary">
              <Sparkles class="h-5 w-5" />
            </div>
            <div class="min-w-0">
              <h1 class="truncate text-2xl font-semibold">提示词编辑器</h1>
              <p class="text-sm text-muted-foreground">Stable Diffusion 提示词优化工具</p>
            </div>
          </div>
          <div class="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span>词库 {{ totalEntryCount }} 条</span>
            <span v-if="contextLabel">上下文：{{ contextLabel }}</span>
            <span v-if="sourcePath" class="truncate">来源：{{ sourcePath }}</span>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <Button variant="ghost" size="sm" @click="copyText(combinedPrompt, '完整组合')">
            <Copy class="mr-1.5 h-3.5 w-3.5" />
            复制全部
          </Button>
          <Button size="sm" @click="saveAsTemplate('other')" :disabled="templateSaving || !combinedPrompt">
            <Bookmark class="mr-1.5 h-3.5 w-3.5" />
            存为组合
          </Button>
        </div>
      </section>

      <Card class="rounded-3xl border-border/70">
        <CardHeader class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-1">
            <CardTitle class="text-lg text-primary">正向编辑区</CardTitle>
            <CardDescription>支持基础 Prompt、预设词包、手动补充和词条排序。</CardDescription>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button variant="ghost" size="sm" @click="dedupeParts('positive')" :disabled="positiveParts.length < 2">去重</Button>
            <Button variant="ghost" size="sm" @click="clearBuilder('positive')" :disabled="!positiveBase && positiveParts.length === 0">清空</Button>
            <Button variant="outline" size="sm" @click="copyText(finalPositivePrompt, '正向 Prompt')" :disabled="!finalPositivePrompt">复制正向</Button>
            <Button size="sm" @click="saveAsTemplate('positive')" :disabled="templateSaving || !finalPositivePrompt">存为正向模板</Button>
          </div>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>基础正向 Prompt</Label>
            <textarea
              v-model="positiveBase"
              class="min-h-[120px] w-full rounded-2xl border border-input bg-background px-4 py-3 text-sm leading-6 outline-none ring-0"
              placeholder="可直接粘贴当前项目的正向 Prompt，或者在下方继续拼装。"
            />
          </div>

          <div class="space-y-3">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium">常用预设词包</span>
              <Button
                v-for="pack in positivePresetPacks"
                :key="pack.id"
                variant="outline"
                size="sm"
                @click="applyPresetPack('positive', pack)"
              >
                {{ pack.name }}
              </Button>
            </div>
            <div class="flex flex-col gap-3 md:flex-row">
              <Input
                v-model="positiveQuickInput"
                class="rounded-xl"
                placeholder="手动补充正向词条，支持逗号、换行分隔。"
              />
              <Button class="rounded-xl md:w-auto" @click="addManualTerms('positive')">
                <Plus class="mr-1.5 h-4 w-4" />
                添加到正向
              </Button>
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-primary">已选正向词条</span>
              <Badge variant="secondary">{{ positiveParts.length }} 条</Badge>
            </div>
            <div v-if="positiveParts.length === 0" class="rounded-2xl border border-dashed p-4 text-sm text-muted-foreground">
              还没有加入正向词条。
            </div>
            <div v-else class="rounded-2xl border border-border/70 bg-muted/10 p-3">
              <div class="flex flex-wrap gap-2">
                <div
                  v-for="(item, index) in positiveParts"
                  :key="item.key"
                  class="flex min-w-[148px] max-w-[220px] items-center gap-2 rounded-xl border border-primary/15 bg-background px-3 py-2 shadow-sm"
                >
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-semibold text-primary/90">{{ item.text || item.textZh }}</div>
                    <div v-if="item.textZh && item.textZh !== item.text" class="mt-0.5 truncate text-xs text-muted-foreground">{{ item.textZh }}</div>
                  </div>
                  <div class="flex shrink-0 items-center gap-1 rounded-lg border border-border/60 bg-muted/40 p-0.5">
                    <button class="flex h-5 w-5 items-center justify-center rounded-md text-[11px] text-muted-foreground transition-colors hover:bg-background hover:text-foreground disabled:opacity-30" :disabled="index === 0" @click="movePart('positive', index, 'up')">↑</button>
                    <button class="flex h-5 w-5 items-center justify-center rounded-md text-[11px] text-muted-foreground transition-colors hover:bg-background hover:text-foreground disabled:opacity-30" :disabled="index === positiveParts.length - 1" @click="movePart('positive', index, 'down')">↓</button>
                    <button class="flex h-5 w-5 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive" @click="removePart('positive', item.key)">
                      <X class="h-3 w-3" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <Label>正向结果</Label>
            <textarea
              :value="finalPositivePrompt"
              readonly
              class="min-h-[120px] w-full rounded-2xl border border-input bg-background px-4 py-3 text-sm leading-6 outline-none ring-0"
              placeholder="正向结果会显示在这里"
            />
          </div>
        </CardContent>
      </Card>

      <Card class="rounded-3xl border-border/70">
        <CardHeader class="space-y-4">
          <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div class="space-y-1">
              <CardTitle class="text-lg">词库浏览</CardTitle>
              <CardDescription>{{ resultSummary }}</CardDescription>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <Button :variant="activeEditor === 'positive' ? 'default' : 'outline'" size="sm" @click="activeEditor = 'positive'">当前编辑：正向</Button>
              <Button :variant="activeEditor === 'negative' ? 'default' : 'outline'" size="sm" @click="activeEditor = 'negative'">当前编辑：反向</Button>
              <Badge variant="outline" class="rounded-full px-3 py-1">{{ customEntryCount }} 我的词库</Badge>
              <Badge variant="outline" class="rounded-full px-3 py-1">{{ favoriteEntryCount }} 收藏</Badge>
              <Badge variant="outline" class="rounded-full px-3 py-1">{{ recentEntryCount }} 最近</Badge>
            </div>
          </div>

          <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px] 2xl:grid-cols-[minmax(0,1fr)_420px]">
            <div class="min-w-0 space-y-4">
              <div class="relative">
                <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  v-model="searchQuery"
                  class="h-11 rounded-2xl pl-10"
                  placeholder="搜索关键词（中文或英文）..."
                />
              </div>

              <div class="flex flex-wrap gap-2">
                <Button :variant="viewMode === 'all' ? 'default' : 'outline'" size="sm" @click="viewMode = 'all'">全部</Button>
                <Button :variant="viewMode === 'favorites' ? 'default' : 'outline'" size="sm" @click="viewMode = 'favorites'">收藏</Button>
                <Button :variant="viewMode === 'recent' ? 'default' : 'outline'" size="sm" @click="viewMode = 'recent'">最近</Button>
                <Button
                  v-for="item in quickCategoryOptions"
                  :key="item.value"
                  :variant="activeCategory === item.value ? 'default' : 'outline'"
                  size="sm"
                  @click="setQuickCategory(item.value)"
                >
                  {{ item.label }}
                </Button>
                <Button v-if="activeCategory" variant="ghost" size="sm" @click="activeCategory = ''">清除分类</Button>
              </div>

              <div class="grid gap-3 sm:grid-cols-2 2xl:grid-cols-5">
                <PromptFilterSelect
                  v-model="activeSource"
                  :options="sourceSelectOptions"
                  :placeholder="'全部来源'"
                  trigger-class="h-10 w-full"
                />
                <PromptFilterSelect
                  v-model="activeCategory"
                  :options="categorySelectOptions"
                  :placeholder="'全部分类'"
                  trigger-class="h-10 w-full"
                />
                <PromptFilterSelect
                  v-model="activeSubcategory"
                  :options="subcategorySelectOptions"
                  :placeholder="'全部子分类'"
                  trigger-class="h-10 w-full"
                />
                <PromptFilterSelect
                  v-model="activeScope"
                  :options="scopeSelectOptions"
                  :placeholder="'全部作用域'"
                  trigger-class="h-10 w-full"
                />
                <div class="flex min-w-0 items-center gap-2 sm:col-span-2 2xl:col-span-1">
                  <PromptFilterSelect
                    :model-value="String(itemsPerPage)"
                    :options="pageSizeOptions"
                    :placeholder="'12 / 页'"
                    trigger-class="h-10 w-full flex-1"
                    @update:model-value="itemsPerPage = Number($event)"
                  />
                  <Button variant="outline" class="h-10 shrink-0 rounded-xl" @click="resetFilters">重置</Button>
                </div>
              </div>
            </div>

            <div class="min-w-0 rounded-2xl border border-dashed bg-muted/20 p-4">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <div class="text-sm font-semibold">添加到我的词库</div>
                  <div class="mt-1 text-xs leading-5 text-muted-foreground">保存时会自动查重，和系统词库或我的词库重复都会被拦截。</div>
                </div>
                <Badge variant="secondary">{{ customEntryCount }} 条</Badge>
              </div>

              <div class="mt-4 space-y-3">
                <Input v-model="customPromptForm.text_zh" placeholder="中文提示词，例如：手放后面" class="rounded-xl" />
                <Input v-model="customPromptForm.text_en" placeholder="英文提示词，例如：arms behind back" class="rounded-xl" />
                <div class="grid gap-3 md:grid-cols-3">
                  <Input v-model="customPromptForm.category" placeholder="分类，例如：人物" class="min-w-0 rounded-xl" />
                  <Input v-model="customPromptForm.subcategory" placeholder="子分类，例如：动作" class="min-w-0 rounded-xl" />
                  <Input v-model="customPromptForm.scope" placeholder="作用域，例如：default" class="min-w-0 rounded-xl" />
                </div>
                <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div class="text-xs text-muted-foreground">来源会自动标记为“我的词库”。</div>
                  <Button class="w-full rounded-xl md:w-auto md:shrink-0" @click="saveCustomPrompt" :disabled="customPromptSaving">
                    <Plus class="mr-1.5 h-4 w-4" />
                    保存自定义提示词
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </CardHeader>

        <CardContent class="space-y-4">
          <div class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-border/70 px-4 py-3 text-sm">
            <div>
              当前显示 <span class="font-medium text-foreground">{{ pageStart }}-{{ pageEnd }}</span> / <span class="font-medium text-foreground">{{ filteredEntries.length }}</span>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <Button variant="outline" size="icon-sm" :disabled="currentPage === 1" @click="currentPage = 1">
                <ChevronsLeft class="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon-sm" :disabled="currentPage === 1" @click="currentPage = Math.max(1, currentPage - 1)">
                <ChevronLeft class="h-4 w-4" />
              </Button>
              <Button
                v-for="page in visiblePages"
                :key="`page-${page}`"
                :variant="page === currentPage ? 'default' : 'ghost'"
                size="sm"
                :disabled="page === '...'"
                @click="typeof page === 'number' && (currentPage = page)"
              >
                {{ page }}
              </Button>
              <Button variant="outline" size="icon-sm" :disabled="currentPage === totalPages" @click="currentPage = Math.min(totalPages, currentPage + 1)">
                <ChevronRight class="h-4 w-4" />
              </Button>
              <Button variant="outline" size="icon-sm" :disabled="currentPage === totalPages" @click="currentPage = totalPages">
                <ChevronsRight class="h-4 w-4" />
              </Button>
            </div>
          </div>

          <div v-if="paginatedEntries.length === 0" class="rounded-2xl border border-dashed p-10 text-center text-muted-foreground">
            没有符合条件的提示词。
          </div>

          <div v-else class="grid gap-3 md:grid-cols-2 2xl:grid-cols-3">
            <Card
              v-for="entry in paginatedEntries"
              :key="entry.id"
              class="rounded-2xl border-border/70 bg-card/60 shadow-sm transition-colors hover:bg-card"
            >
              <CardContent class="p-4">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-lg font-semibold leading-tight">{{ entry.text_zh || entry.text_en || '未命名词条' }}</div>
                    <div class="mt-1 break-all text-sm text-muted-foreground">{{ entry.text_en || '暂无英文内容' }}</div>
                  </div>
                  <div class="flex items-center gap-2">
                    <button class="text-muted-foreground transition-colors hover:text-rose-500" @click="toggleFavorite(entry.id)">
                      <Heart class="h-4 w-4" :class="favoriteIdSet.has(entry.id) ? 'fill-rose-500 text-rose-500' : ''" />
                    </button>
                    <button
                      v-if="isCustomEntry(entry)"
                      class="text-muted-foreground transition-colors hover:text-destructive"
                      @click="requestDeleteCustomPrompt(entry)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <div class="mt-3 flex flex-wrap gap-1.5 text-xs">
                  <Badge :variant="isCustomEntry(entry) ? 'default' : 'secondary'">{{ entry.source || '未分类来源' }}</Badge>
                  <Badge variant="outline">{{ entry.category || '未分类' }}</Badge>
                  <Badge variant="outline">{{ entry.subcategory || '未设子分类' }}</Badge>
                  <Badge variant="outline">{{ entry.scope || 'default' }}</Badge>
                </div>

                <div class="mt-3 flex flex-wrap gap-2">
                  <Button size="sm" class="h-8 rounded-lg px-3" @click="addToBuilder(activeEditor, entry)">
                    <Plus class="mr-1.5 h-4 w-4" />
                    加入当前区
                  </Button>
                  <Button size="sm" variant="outline" class="h-8 rounded-lg px-3" @click="addToBuilder('positive', entry)">加入正向</Button>
                  <Button size="sm" variant="outline" class="h-8 rounded-lg px-3" @click="addToBuilder('negative', entry)">加入反向</Button>
                  <Button size="sm" variant="ghost" class="h-8 rounded-lg px-2.5" @click="copyText(entry.text_en || entry.text_zh, '词条')">
                    <Copy class="mr-1.5 h-4 w-4" />
                    复制
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        </CardContent>
      </Card>

      <Card class="rounded-3xl border-border/70">
        <CardHeader class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-1">
            <CardTitle class="text-lg text-rose-500">反向编辑区</CardTitle>
            <CardDescription>用于整理负向 Prompt，支持预设词包和手动词条补充。</CardDescription>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button variant="ghost" size="sm" @click="dedupeParts('negative')" :disabled="negativeParts.length < 2">去重</Button>
            <Button variant="ghost" size="sm" @click="clearBuilder('negative')" :disabled="!negativeBase && negativeParts.length === 0">清空</Button>
            <Button variant="outline" size="sm" @click="copyText(finalNegativePrompt, '反向 Prompt')" :disabled="!finalNegativePrompt">复制反向</Button>
            <Button size="sm" @click="saveAsTemplate('negative')" :disabled="templateSaving || !finalNegativePrompt">存为反向模板</Button>
          </div>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>基础反向 Prompt</Label>
            <textarea
              v-model="negativeBase"
              class="min-h-[120px] w-full rounded-2xl border border-input bg-background px-4 py-3 text-sm leading-6 outline-none ring-0"
              placeholder="可直接粘贴当前项目的反向 Prompt，或者在下方继续拼装。"
            />
          </div>

          <div class="space-y-3">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium">常用预设词包</span>
              <Button
                v-for="pack in negativePresetPacks"
                :key="pack.id"
                variant="outline"
                size="sm"
                @click="applyPresetPack('negative', pack)"
              >
                {{ pack.name }}
              </Button>
            </div>
            <div class="flex flex-col gap-3 md:flex-row">
              <Input
                v-model="negativeQuickInput"
                class="rounded-xl"
                placeholder="手动补充反向词条，支持逗号、换行分隔。"
              />
              <Button class="rounded-xl md:w-auto" @click="addManualTerms('negative')">
                <Plus class="mr-1.5 h-4 w-4" />
                添加到反向
              </Button>
            </div>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-rose-500">已选反向词条</span>
              <Badge variant="secondary">{{ negativeParts.length }} 条</Badge>
            </div>
            <div v-if="negativeParts.length === 0" class="rounded-2xl border border-dashed p-4 text-sm text-muted-foreground">
              还没有加入反向词条。
            </div>
            <div v-else class="rounded-2xl border border-border/70 bg-muted/10 p-3">
              <div class="flex flex-wrap gap-2">
                <div
                  v-for="(item, index) in negativeParts"
                  :key="item.key"
                  class="flex min-w-[148px] max-w-[220px] items-center gap-2 rounded-xl border border-rose-500/15 bg-background px-3 py-2 shadow-sm"
                >
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-semibold text-rose-500/90">{{ item.text || item.textZh }}</div>
                    <div v-if="item.textZh && item.textZh !== item.text" class="mt-0.5 truncate text-xs text-muted-foreground">{{ item.textZh }}</div>
                  </div>
                  <div class="flex shrink-0 items-center gap-1 rounded-lg border border-border/60 bg-muted/40 p-0.5">
                    <button class="flex h-5 w-5 items-center justify-center rounded-md text-[11px] text-muted-foreground transition-colors hover:bg-background hover:text-foreground disabled:opacity-30" :disabled="index === 0" @click="movePart('negative', index, 'up')">↑</button>
                    <button class="flex h-5 w-5 items-center justify-center rounded-md text-[11px] text-muted-foreground transition-colors hover:bg-background hover:text-foreground disabled:opacity-30" :disabled="index === negativeParts.length - 1" @click="movePart('negative', index, 'down')">↓</button>
                    <button class="flex h-5 w-5 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive" @click="removePart('negative', item.key)">
                      <X class="h-3 w-3" />
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <Label>反向结果</Label>
            <textarea
              :value="finalNegativePrompt"
              readonly
              class="min-h-[120px] w-full rounded-2xl border border-input bg-background px-4 py-3 text-sm leading-6 outline-none ring-0"
              placeholder="反向结果会显示在这里"
            />
          </div>
        </CardContent>
      </Card>

      <Card class="rounded-3xl border-border/70">
        <CardHeader>
          <CardTitle class="text-lg">模板保存设置</CardTitle>
          <CardDescription>保存前可自定义模板名称和模板分类，组合模板保存后可直接复制使用。</CardDescription>
        </CardHeader>
        <CardContent class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
          <Input v-model="templateName" placeholder="模板名称，可留空自动生成" class="h-10 rounded-xl" />
          <Input v-model="templateCategory" placeholder="模板分类，例如：人物、画风、质量词" class="h-10 rounded-xl" />
          <Button class="h-10 rounded-xl" @click="saveAsTemplate('other')" :disabled="templateSaving || !combinedPrompt">
            <Bookmark class="mr-1.5 h-4 w-4" />
            保存组合模板
          </Button>
        </CardContent>
      </Card>
    </div>

    <AlertDialog v-model:open="customPromptDeleteOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除自定义提示词</AlertDialogTitle>
          <AlertDialogDescription>
            确定要删除“{{ pendingDeleteCustomPrompt?.text_zh || pendingDeleteCustomPrompt?.text_en || '该提示词' }}”吗？此操作不可撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            class="bg-destructive hover:bg-destructive/90"
            @click="deleteCustomPrompt"
          >
            删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
