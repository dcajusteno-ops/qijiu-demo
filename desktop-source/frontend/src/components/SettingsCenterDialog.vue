<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import { isDark, toggleTheme } from '@/theme'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import {
  BrushCleaning,
  ChevronDown,
  ChevronUp,
  Eye,
  EyeOff,
  FolderTree,
  Heart,
  Keyboard,
  Loader2,
  ListOrdered,
  Moon,
  Pencil,
  RotateCcw,
  Save,
  Sparkles,
  Sun,
  Trash2,
  Wrench,
  X,
} from 'lucide-vue-next'
import {
  buildBindingsFromCatalog,
  findDuplicateAccelerators,
  formatShortcutLabel,
  getAcceleratorFromEvent,
  normalizeAccelerator,
  shortcutActionCatalog,
  shortcutActionMap,
} from '@/lib/shortcuts'

const props = defineProps({
  open: { type: Boolean, default: false },
  favoriteGroups: { type: Array, default: () => [] },
})

const emit = defineEmits([
  'update:open',
  'favorite-group-change',
  'refresh-images',
  'organize-files',
  'utility-menu-change',
])

const sections = [
  { id: 'toolmenu', label: '工具菜单', icon: ListOrdered, description: '控制菜单顺序与显示' },
  { id: 'appearance', label: '外观模式', icon: Sparkles, description: '主题与视觉偏好' },
  { id: 'favorites', label: '收藏分组', icon: Heart, description: '整理收藏夹结构' },
  { id: 'shortcuts', label: '快捷键设置', icon: Keyboard, description: '调整全局快捷键' },
  { id: 'cache', label: '缓存清理', icon: BrushCleaning, description: '清空预览缓存' },
  { id: 'folders', label: '文件夹维护', icon: FolderTree, description: '清理空文件夹' },
]

const activeSection = ref('appearance')
const confirmState = ref({ open: false, type: '', title: '', description: '' })

const normalizeError = (error, fallback) => {
  const message = String(error ?? '').trim()
  if (!message || message.includes('�')) return fallback
  return message
}

watch(() => props.open, (open) => {
  if (!open) return
  activeSection.value = 'toolmenu'
  loadShortcutSettings()
  loadUtilityMenuSettings()
  resetFavoriteEditing()
})

const activeSectionMeta = computed(() => sections.find((item) => item.id === activeSection.value) || sections[0])

const utilityMenuCatalog = [
  { id: 'settings', label: '设置', description: '打开设置中心', locked: true },
  { id: 'trash', label: '回收站管理', description: '查看已删除图片' },
  { id: 'documentation', label: '使用文档', description: '查看内置说明文档' },
  { id: 'statistics', label: '数据视界', description: '查看图像统计面板' },
  { id: 'launcher', label: '外部工具', description: '打开外部工具入口' },
  { id: 'prompt-templates', label: '提示词模板', description: '管理常用提示词模板' },
  { id: 'prompt-assistant', label: '提示词提示器', description: '搜索词库并拼装 Prompt' },
  { id: 'auto-rules', label: '自动规则引擎', description: '执行自动规则处理' },
  { id: 'open-output', label: '打开当前 output', description: '打开当前绑定的输出目录' },
  { id: 'switch-output', label: '切换 output 位置', description: '重新绑定 ComfyUI output' },
  { id: 'custom-roots', label: '管理自定义目录', description: '维护自定义目录与显示顺序' },
]

const utilityMenuItems = ref([])
const utilityMenuLoading = ref(false)
const utilityMenuSaving = ref(false)

const utilityMenuRows = computed(() => {
  const settingsMap = new Map((utilityMenuItems.value || []).map((item) => [item.id, item]))
  return utilityMenuCatalog
    .map((item, index) => {
      const saved = settingsMap.get(item.id)
      return {
        ...item,
        visible: item.locked ? true : saved?.visible !== false,
        order: saved?.order ?? index + 1,
      }
    })
    .sort((a, b) => {
      if (a.id === 'settings') return -1
      if (b.id === 'settings') return 1
      return a.order - b.order
    })
})

const loadUtilityMenuSettings = async () => {
  utilityMenuLoading.value = true
  try {
    const state = await App.GetUtilityMenuSettings()
    utilityMenuItems.value = state?.items || []
  } catch (error) {
    utilityMenuItems.value = utilityMenuCatalog.map((item, index) => ({
      id: item.id,
      visible: true,
      order: index + 1,
    }))
    toast.error(normalizeError(error, '加载工具菜单设置失败'))
  } finally {
    utilityMenuLoading.value = false
  }
}

const buildUtilityMenuPayload = () => utilityMenuRows.value.map((item, index) => ({
  id: item.id,
  visible: item.locked ? true : item.visible,
  order: index + 1,
}))

const saveUtilityMenuSettings = async () => {
  utilityMenuSaving.value = true
  try {
    const state = await App.SaveUtilityMenuSettings({ items: buildUtilityMenuPayload() })
    utilityMenuItems.value = state?.items || []
    emit('utility-menu-change', state)
    toast.success('工具菜单设置已保存')
  } catch (error) {
    toast.error(normalizeError(error, '保存工具菜单设置失败'))
  } finally {
    utilityMenuSaving.value = false
  }
}

const reorderUtilityMenu = (id, direction) => {
  const rows = [...utilityMenuRows.value]
  const index = rows.findIndex((item) => item.id === id)
  if (index < 0) return

  const targetIndex = direction === 'up' ? index - 1 : index + 1
  if (targetIndex < 0 || targetIndex >= rows.length) return

  const [current] = rows.splice(index, 1)
  rows.splice(targetIndex, 0, current)
  utilityMenuItems.value = rows.map((item, idx) => ({
    id: item.id,
    visible: item.visible,
    order: idx + 1,
  }))
}

const toggleUtilityMenuVisibility = (id, visible) => {
  utilityMenuItems.value = utilityMenuRows.value.map((item, index) => ({
    id: item.id,
    visible: item.id === id ? visible : item.visible,
    order: index + 1,
  }))
}

const openConfirm = (type, title, description) => {
  confirmState.value = { open: true, type, title, description }
}

const closeConfirm = () => {
  confirmState.value = { open: false, type: '', title: '', description: '' }
}

const confirmAction = async () => {
  const type = confirmState.value.type
  closeConfirm()
  if (type === 'clear-cache') await clearPreviewCache()
  if (type === 'clean-empty-folders') await cleanEmptyFolders()
  if (type === 'delete-group') await confirmDeleteFavoriteGroup()
}

const handleThemeToggle = (event) => {
  toggleTheme(event)
  toast.success(isDark.value ? '已切换为暗色模式' : '已切换为亮色模式')
}

const clearingCache = ref(false)
const cleanEmptyLoading = ref(false)

const formatBytes = (value) => {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

const clearPreviewCache = async () => {
  clearingCache.value = true
  try {
    const result = await App.ClearPreviewCache()
    toast.success(`已清空 ${result.deletedFiles || 0} 个缓存文件，释放 ${formatBytes(result.bytesFreed || 0)}`)
  } catch (error) {
    toast.error(normalizeError(error, '清空预览缓存失败'))
  } finally {
    clearingCache.value = false
  }
}

const cleanEmptyFolders = async () => {
  cleanEmptyLoading.value = true
  try {
    const count = await App.CleanEmptyFolders()
    toast.success(`已清理 ${count} 个空文件夹`)
    emit('refresh-images')
  } catch (error) {
    toast.error(normalizeError(error, '清理空文件夹失败'))
  } finally {
    cleanEmptyLoading.value = false
  }
}

const newFavoriteGroupName = ref('')
const editingFavoriteGroupId = ref('')
const editingFavoriteGroupName = ref('')
const creatingFavoriteGroup = ref(false)
const updatingFavoriteGroup = ref(false)
const deletingFavoriteGroup = ref(false)
const deletingFavoriteGroupId = ref('')

const resetFavoriteEditing = () => {
  newFavoriteGroupName.value = ''
  editingFavoriteGroupId.value = ''
  editingFavoriteGroupName.value = ''
  deletingFavoriteGroupId.value = ''
}

const startFavoriteRename = (group) => {
  editingFavoriteGroupId.value = group.id
  editingFavoriteGroupName.value = group.name || ''
}

const cancelFavoriteRename = () => {
  editingFavoriteGroupId.value = ''
  editingFavoriteGroupName.value = ''
}

const createFavoriteGroup = async () => {
  const name = newFavoriteGroupName.value.trim()
  if (!name) {
    toast.error('请输入收藏分组名称')
    return
  }
  creatingFavoriteGroup.value = true
  try {
    await App.CreateFavoriteGroup(name)
    newFavoriteGroupName.value = ''
    toast.success('收藏分组已创建')
    emit('favorite-group-change')
  } catch (error) {
    toast.error(normalizeError(error, '创建收藏分组失败'))
  } finally {
    creatingFavoriteGroup.value = false
  }
}

const saveFavoriteRename = async () => {
  const name = editingFavoriteGroupName.value.trim()
  if (!editingFavoriteGroupId.value || !name) {
    toast.error('请输入分组名称')
    return
  }
  updatingFavoriteGroup.value = true
  try {
    await App.UpdateFavoriteGroup(editingFavoriteGroupId.value, name)
    toast.success('分组名称已更新')
    cancelFavoriteRename()
    emit('favorite-group-change')
  } catch (error) {
    toast.error(normalizeError(error, '更新收藏分组失败'))
  } finally {
    updatingFavoriteGroup.value = false
  }
}

const requestDeleteFavoriteGroup = (groupId) => {
  deletingFavoriteGroupId.value = groupId
  openConfirm('delete-group', '删除收藏分组', '删除后不会删除图片，只会移除这个分组本身。')
}

const confirmDeleteFavoriteGroup = async () => {
  if (!deletingFavoriteGroupId.value) return
  deletingFavoriteGroup.value = true
  try {
    await App.DeleteFavoriteGroup(deletingFavoriteGroupId.value)
    deletingFavoriteGroupId.value = ''
    toast.success('收藏分组已删除')
    emit('favorite-group-change')
  } catch (error) {
    toast.error(normalizeError(error, '删除收藏分组失败'))
  } finally {
    deletingFavoriteGroup.value = false
  }
}

const shortcutLoading = ref(false)
const shortcutSaving = ref(false)
const shortcutEnabled = ref(true)
const shortcutBindings = ref(buildBindingsFromCatalog())
const capturingShortcutAction = ref('')

const duplicateAccelerators = computed(() => findDuplicateAccelerators(shortcutBindings.value))
const hasDuplicateAccelerators = computed(() => duplicateAccelerators.value.size > 0)

const shortcutRows = computed(() => shortcutActionCatalog.map((action) => {
  const binding = shortcutBindings.value.find((item) => item.action === action.id) || {
    action: action.id,
    accelerator: normalizeAccelerator(action.defaultAccelerator),
  }
  const normalized = normalizeAccelerator(binding.accelerator)
  return {
    ...action,
    accelerator: normalized,
    isDuplicate: normalized ? duplicateAccelerators.value.has(normalized) : false,
    isCapturing: capturingShortcutAction.value === action.id,
  }
}))

const setShortcutBinding = (actionId, accelerator) => {
  shortcutBindings.value = shortcutBindings.value.map((binding) => (
    binding.action === actionId
      ? { ...binding, accelerator: normalizeAccelerator(accelerator) }
      : binding
  ))
}

const resetShortcutDefaults = () => {
  shortcutEnabled.value = true
  shortcutBindings.value = buildBindingsFromCatalog()
  capturingShortcutAction.value = ''
}

const resetSingleShortcut = (actionId) => {
  const action = shortcutActionMap[actionId]
  if (!action) return
  setShortcutBinding(actionId, action.defaultAccelerator)
}

const clearSingleShortcut = (actionId) => {
  setShortcutBinding(actionId, '')
}

const startShortcutCapture = (actionId) => {
  capturingShortcutAction.value = actionId
}

const cancelShortcutCapture = () => {
  capturingShortcutAction.value = ''
}

const handleShortcutCaptureKeydown = (event) => {
  if (!capturingShortcutAction.value) return

  if (event.key === 'Escape' && !event.ctrlKey && !event.altKey && !event.shiftKey && !event.metaKey) {
    event.preventDefault()
    cancelShortcutCapture()
    return
  }

  const accelerator = getAcceleratorFromEvent(event)
  if (!accelerator) return

  event.preventDefault()
  event.stopPropagation()
  setShortcutBinding(capturingShortcutAction.value, accelerator)
  cancelShortcutCapture()
}

const loadShortcutSettings = async () => {
  shortcutLoading.value = true
  try {
    const settings = await App.GetShortcutSettings()
    shortcutEnabled.value = settings?.enabled !== false
    shortcutBindings.value = buildBindingsFromCatalog(settings?.bindings || [])
  } catch (error) {
    shortcutEnabled.value = true
    shortcutBindings.value = buildBindingsFromCatalog()
    toast.error(normalizeError(error, '加载快捷键设置失败'))
  } finally {
    shortcutLoading.value = false
  }
}

const saveShortcutSettings = async () => {
  if (hasDuplicateAccelerators.value) {
    toast.error('存在重复快捷键，请先调整后再保存')
    return
  }
  shortcutSaving.value = true
  try {
    const saved = await App.SaveShortcutSettings({
      enabled: shortcutEnabled.value,
      bindings: shortcutBindings.value.map((binding) => ({ action: binding.action, accelerator: binding.accelerator })),
    })
    shortcutEnabled.value = saved?.enabled !== false
    shortcutBindings.value = buildBindingsFromCatalog(saved?.bindings || [])
    toast.success('快捷键设置已保存')
  } catch (error) {
    toast.error(normalizeError(error, '保存快捷键设置失败'))
  } finally {
    shortcutSaving.value = false
  }
}

watch(capturingShortcutAction, (next, prev) => {
  if (next && !prev) {
    window.addEventListener('keydown', handleShortcutCaptureKeydown, true)
    return
  }
  if (!next && prev) {
    window.removeEventListener('keydown', handleShortcutCaptureKeydown, true)
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleShortcutCaptureKeydown, true)
})
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[1100px] h-[82vh] max-h-[90vh] overflow-hidden p-0">
      <div class="flex h-full min-h-0">
        <div class="w-[260px] shrink-0 border-r bg-muted/20">
          <DialogHeader class="px-5 pb-3 pt-5">
            <DialogTitle class="flex items-center gap-2 text-lg">
              <Wrench class="h-5 w-5 text-primary" />
              设置中心
            </DialogTitle>
            <DialogDescription>
              把常用的系统功能集中到一个地方，右侧按模块查看与操作。
            </DialogDescription>
          </DialogHeader>

          <div class="px-3 pb-4">
            <div class="space-y-1">
              <button
                v-for="section in sections"
                :key="section.id"
                class="flex w-full items-start gap-3 rounded-xl px-3 py-3 text-left transition-colors"
                :class="activeSection === section.id ? 'bg-background text-foreground shadow-sm border' : 'text-muted-foreground hover:bg-background/70 hover:text-foreground'"
                @click="activeSection = section.id"
              >
                <component :is="section.icon" class="mt-0.5 h-4 w-4 shrink-0" />
                <div class="min-w-0">
                  <div class="text-sm font-medium">{{ section.label }}</div>
                  <div class="mt-1 text-xs opacity-75">{{ section.description }}</div>
                </div>
              </button>
            </div>
          </div>
        </div>

        <div class="flex min-h-0 flex-1 flex-col bg-background">
          <div class="border-b px-6 py-5">
            <div class="flex items-center gap-3">
              <component :is="activeSectionMeta.icon" class="h-5 w-5 text-primary" />
              <div>
                <div class="text-xl font-semibold">{{ activeSectionMeta.label }}</div>
                <div class="text-sm text-muted-foreground">{{ activeSectionMeta.description }}</div>
              </div>
            </div>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto px-6 py-6">
            <div v-if="activeSection === 'toolmenu'" class="space-y-4">
              <div class="rounded-2xl border bg-muted/20 p-5 flex items-center justify-between gap-4">
                <div>
                  <div class="text-base font-semibold">工具菜单顺序与显示</div>
                  <div class="text-sm text-muted-foreground">设置按钮固定保留，其余项可调整显示状态和排列顺序。</div>
                </div>
                <Button :disabled="utilityMenuLoading || utilityMenuSaving" @click="saveUtilityMenuSettings">
                  <Save class="mr-2 h-4 w-4" />
                  {{ utilityMenuSaving ? '保存中...' : '保存工具菜单设置' }}
                </Button>
              </div>

              <div class="rounded-2xl border overflow-hidden">
                <div v-if="utilityMenuLoading" class="px-5 py-10 text-center text-sm text-muted-foreground">正在加载工具菜单设置...</div>
                <template v-else>
                  <div
                    v-for="(row, index) in utilityMenuRows"
                    :key="row.id"
                    class="flex items-center justify-between gap-4 border-b px-5 py-4 last:border-b-0"
                  >
                    <div class="min-w-0">
                      <div class="flex items-center gap-2">
                        <div class="font-medium">{{ row.label }}</div>
                        <Badge v-if="row.locked" variant="secondary">固定</Badge>
                        <Badge v-if="row.visible" variant="outline">显示中</Badge>
                        <Badge v-else variant="outline" class="text-muted-foreground">已隐藏</Badge>
                      </div>
                      <div class="mt-1 text-sm text-muted-foreground">{{ row.description }}</div>
                    </div>
                    <div class="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="icon"
                        :disabled="row.locked || index === 0"
                        title="上移"
                        @click="reorderUtilityMenu(row.id, 'up')"
                      >
                        <ChevronUp class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        :disabled="row.locked || index === utilityMenuRows.length - 1"
                        title="下移"
                        @click="reorderUtilityMenu(row.id, 'down')"
                      >
                        <ChevronDown class="h-4 w-4" />
                      </Button>
                      <Button
                        variant="outline"
                        class="gap-2"
                        :disabled="row.locked"
                        @click="toggleUtilityMenuVisibility(row.id, !row.visible)"
                      >
                        <Eye v-if="row.visible" class="h-4 w-4" />
                        <EyeOff v-else class="h-4 w-4" />
                        {{ row.visible ? '隐藏' : '显示' }}
                      </Button>
                    </div>
                  </div>
                </template>
              </div>
            </div>

            <div v-else-if="activeSection === 'appearance'" class="space-y-4">
              <div class="rounded-2xl border bg-muted/20 p-5">
                <div class="flex items-center justify-between gap-4">
                  <div class="space-y-2">
                    <div class="text-base font-semibold">界面模式</div>
                    <div class="text-sm text-muted-foreground">当前支持亮色与暗色两种模式，切换后立即生效。</div>
                    <Badge variant="secondary">{{ isDark ? '当前为暗色模式' : '当前为亮色模式' }}</Badge>
                  </div>
                  <Button class="gap-2" @click="handleThemeToggle($event)">
                    <Moon v-if="isDark" class="h-4 w-4" />
                    <Sun v-else class="h-4 w-4" />
                    {{ isDark ? '切换为亮色模式' : '切换为暗色模式' }}
                  </Button>
                </div>
              </div>
            </div>

            <div v-else-if="activeSection === 'favorites'" class="space-y-4">
              <div class="rounded-2xl border bg-muted/20 p-5 space-y-4">
                <div class="text-base font-semibold">新建收藏分组</div>
                <div class="flex gap-2">
                  <Input v-model="newFavoriteGroupName" placeholder="输入新的收藏分组名称" @keydown.enter="createFavoriteGroup" />
                  <Button :disabled="creatingFavoriteGroup" @click="createFavoriteGroup">
                    <Loader2 v-if="creatingFavoriteGroup" class="mr-2 h-4 w-4 animate-spin" />
                    新建
                  </Button>
                </div>
              </div>

              <div class="rounded-2xl border bg-muted/20 p-5 space-y-3">
                <div class="text-base font-semibold">现有分组</div>
                <div v-for="group in favoriteGroups" :key="group.id" class="rounded-xl border bg-background p-4">
                  <div v-if="editingFavoriteGroupId === group.id" class="flex gap-2">
                    <Input v-model="editingFavoriteGroupName" @keydown.enter="saveFavoriteRename" />
                    <Button :disabled="updatingFavoriteGroup" @click="saveFavoriteRename">保存</Button>
                    <Button variant="outline" @click="cancelFavoriteRename">取消</Button>
                  </div>
                  <div v-else class="flex items-center justify-between gap-3">
                    <div>
                      <div class="font-medium">{{ group.name }}</div>
                      <div class="text-xs text-muted-foreground">{{ group.paths?.length || 0 }} 张图片</div>
                    </div>
                    <div class="flex gap-2">
                      <Button variant="outline" size="icon" @click="startFavoriteRename(group)">
                        <Pencil class="h-4 w-4" />
                      </Button>
                      <Button variant="outline" size="icon" :disabled="group.id === 'default' || deletingFavoriteGroup" @click="requestDeleteFavoriteGroup(group.id)">
                        <Trash2 class="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-else-if="activeSection === 'shortcuts'" class="space-y-4">
              <div class="rounded-2xl border bg-muted/20 p-5 flex items-center justify-between gap-4">
                <div>
                  <div class="text-base font-semibold">启用全局快捷键</div>
                  <div class="text-sm text-muted-foreground">程序运行期间，快捷键可用于切换视图或执行快速操作。</div>
                </div>
                <Switch v-model:checked="shortcutEnabled" />
              </div>

              <div v-if="hasDuplicateAccelerators" class="rounded-2xl border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
                存在重复快捷键，请先处理后再保存。
              </div>

              <div class="rounded-2xl border overflow-hidden">
                <div v-if="shortcutLoading" class="px-5 py-10 text-center text-sm text-muted-foreground">正在加载快捷键设置...</div>
                <template v-else>
                  <div v-for="row in shortcutRows" :key="row.id" class="flex items-center justify-between gap-4 border-b px-5 py-4 last:border-b-0">
                    <div class="min-w-0">
                      <div class="font-medium">{{ row.label }}</div>
                      <div class="mt-1 text-sm text-muted-foreground">{{ row.description }}</div>
                      <div class="mt-2 text-xs text-muted-foreground">默认：{{ formatShortcutLabel(row.defaultAccelerator) }}</div>
                    </div>
                    <div class="flex items-center gap-2">
                      <div class="min-w-[160px] rounded-xl border px-4 py-3 text-center text-sm" :class="row.isDuplicate ? 'border-destructive text-destructive' : 'border-border'">
                        {{ formatShortcutLabel(row.accelerator) }}
                      </div>
                      <Button :variant="row.isCapturing ? 'default' : 'outline'" @click="row.isCapturing ? cancelShortcutCapture() : startShortcutCapture(row.id)">
                        {{ row.isCapturing ? '取消录制' : '录制' }}
                      </Button>
                      <Button size="icon" variant="ghost" title="恢复默认" @click="resetSingleShortcut(row.id)">
                        <RotateCcw class="h-4 w-4" />
                      </Button>
                      <Button size="icon" variant="ghost" title="清空快捷键" @click="clearSingleShortcut(row.id)">
                        <X class="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </template>
              </div>

              <div class="flex items-center justify-between gap-3">
                <Button variant="outline" @click="resetShortcutDefaults">
                  <RotateCcw class="mr-2 h-4 w-4" />
                  恢复全部默认
                </Button>
                <Button :disabled="shortcutLoading || shortcutSaving || hasDuplicateAccelerators" @click="saveShortcutSettings">
                  <Save class="mr-2 h-4 w-4" />
                  {{ shortcutSaving ? '保存中...' : '保存快捷键设置' }}
                </Button>
              </div>
            </div>

            <div v-else-if="activeSection === 'cache'" class="space-y-4">
              <div class="rounded-2xl border bg-muted/20 p-5 flex items-center justify-between gap-4">
                <div>
                  <div class="text-base font-semibold">清空预览缓存</div>
                  <div class="text-sm text-muted-foreground">下次浏览图片时会重新生成缩略图与缓存信息。</div>
                </div>
                <Button :disabled="clearingCache" @click="openConfirm('clear-cache', '清空预览缓存', '确定要清空预览缓存吗？')">
                  <Loader2 v-if="clearingCache" class="mr-2 h-4 w-4 animate-spin" />
                  清空缓存
                </Button>
              </div>
            </div>

            <div v-else-if="activeSection === 'folders'" class="space-y-4">
              <div class="rounded-2xl border bg-muted/20 p-5 flex items-center justify-between gap-4">
                <div>
                  <div class="text-base font-semibold">清理空文件夹</div>
                  <div class="text-sm text-muted-foreground">扫描当前输出目录与自定义目录，移除无内容的空文件夹。</div>
                </div>
                <Button :disabled="cleanEmptyLoading" @click="openConfirm('clean-empty-folders', '清理空文件夹', '确定要清理所有空文件夹吗？')">
                  <Loader2 v-if="cleanEmptyLoading" class="mr-2 h-4 w-4 animate-spin" />
                  清理空文件夹
                </Button>
              </div>

              <div class="rounded-2xl border bg-muted/20 p-5 flex items-center justify-between gap-4">
                <div>
                  <div class="text-base font-semibold">按照日期整理文件</div>
                  <div class="text-sm text-muted-foreground">把根目录下的散落图片整理到日期目录，便于归档查看。</div>
                </div>
                <Button @click="$emit('organize-files')">
                  按日期整理文件
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </DialogContent>

    <AlertDialog :open="confirmState.open">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{{ confirmState.title }}</AlertDialogTitle>
          <AlertDialogDescription>{{ confirmState.description }}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel @click="closeConfirm">取消</AlertDialogCancel>
          <AlertDialogAction @click="confirmAction">确定</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </Dialog>
</template>
