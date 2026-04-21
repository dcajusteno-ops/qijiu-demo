<script setup>
import { computed, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  ArrowDown,
  ArrowUp,
  Check,
  ChevronsUp,
  FolderOpen,
  FolderSymlink,
  Loader2,
  Lock,
  Pencil,
  Plus,
  Trash2,
  X,
} from 'lucide-vue-next'
import { availableIcons, categorizedIcons, iconCount } from '@/lib/icons'

const props = defineProps({
  open: { type: Boolean, default: false },
  customRoots: { type: Array, default: () => [] },
})

const emit = defineEmits(['update:open', 'change'])

const defaultIcon = 'FolderSymlink'

const newName = ref('')
const selectedPath = ref('')
const selectedIcon = ref(defaultIcon)
const isSelecting = ref(false)
const isAdding = ref(false)
const busyId = ref('')

const editingRootId = ref('')
const editingName = ref('')
const editingIcon = ref(defaultIcon)
const isUpdating = ref(false)

const editingRoot = computed(() => props.customRoots.find((root) => root.id === editingRootId.value) ?? null)

const normalizeError = (error, fallback) => {
  const message = String(error ?? '').trim()
  if (!message || message.includes('�')) {
    return fallback
  }
  return message
}

const resetAddState = () => {
  newName.value = ''
  selectedPath.value = ''
  selectedIcon.value = defaultIcon
}

const resetEditState = () => {
  editingRootId.value = ''
  editingName.value = ''
  editingIcon.value = defaultIcon
}

watch(() => props.open, (open) => {
  if (!open) return
  resetAddState()
  resetEditState()
})

const selectFolder = async () => {
  isSelecting.value = true
  try {
    const absPath = await App.SelectFolder()
    if (!absPath) return

    const relPath = await App.GetRelativePath(absPath)
    selectedPath.value = relPath

    if (!newName.value) {
      const parts = relPath.split('/')
      newName.value = parts[parts.length - 1]
    }
  } catch (error) {
    toast.error(normalizeError(error, '所选文件夹必须位于当前 ComfyUI 工作根目录内'))
  } finally {
    isSelecting.value = false
  }
}

const handleAdd = async () => {
  if (!selectedPath.value) {
    toast.error('请先选择一个文件夹')
    return
  }

  const displayName = newName.value.trim() || selectedPath.value.split('/').pop()
  isAdding.value = true
  try {
    await App.AddCustomRoot(displayName, selectedPath.value, selectedIcon.value)
    toast.success('自定义目录已添加')
    resetAddState()
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '添加自定义目录失败'))
  } finally {
    isAdding.value = false
  }
}

const startEdit = (root) => {
  if (root.locked || root.isBuiltin) return
  editingRootId.value = root.id
  editingName.value = root.name || ''
  editingIcon.value = root.icon || defaultIcon
}

const handleUpdate = async () => {
  if (!editingRoot.value) return

  const fallbackName = editingRoot.value.path.split('/').pop() || editingRoot.value.name
  const displayName = editingName.value.trim() || fallbackName
  isUpdating.value = true
  try {
    await App.UpdateCustomRoot(editingRoot.value.id, displayName, editingIcon.value)
    toast.success('目录设置已更新')
    resetEditState()
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '保存失败'))
  } finally {
    isUpdating.value = false
  }
}

const handleDelete = async (root) => {
  try {
    await App.DeleteCustomRoot(root.id)
    toast.success('自定义目录已删除')
    if (editingRootId.value === root.id) {
      resetEditState()
    }
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '删除失败'))
  }
}

const toggleEnabled = async (root, enabled) => {
  busyId.value = root.id
  try {
    await App.UpdateCustomRootEnabled(root.id, enabled)
    toast.success(enabled ? '目录已启用' : '目录已隐藏')
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '更新目录状态失败'))
  } finally {
    busyId.value = ''
  }
}

const moveRoot = async (root, direction) => {
  busyId.value = root.id
  try {
    await App.MoveCustomRoot(root.id, direction)
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '移动目录失败'))
  } finally {
    busyId.value = ''
  }
}

const pinRoot = async (root) => {
  busyId.value = root.id
  try {
    await App.PinCustomRoot(root.id)
    toast.success(root.pinned ? '已取消置顶' : '目录已置顶')
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '置顶目录失败'))
  } finally {
    busyId.value = ''
  }
}

const isFirstRoot = (root) => props.customRoots.findIndex((item) => item.id === root.id) === 0
const isLastRoot = (root) => props.customRoots.findIndex((item) => item.id === root.id) === props.customRoots.length - 1
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="flex max-h-[88vh] overflow-hidden p-0 sm:max-w-[760px]">
      <div class="flex min-h-0 w-full flex-col p-6">
        <DialogHeader class="pr-8">
          <DialogTitle class="flex items-center gap-2">
            <FolderSymlink class="h-5 w-5 text-primary" />
            管理自定义目录
          </DialogTitle>
          <DialogDescription>
            侧边栏现在只保留默认目录，其余入口都从这里管理。内置的“日期归档目录”可以开关显示，但不可删除。
          </DialogDescription>
        </DialogHeader>

        <div class="mt-5 min-h-0 flex-1 grid items-stretch gap-5 lg:grid-cols-[1.1fr_0.9fr]">
          <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
            <div class="flex min-h-0 flex-1 flex-col space-y-2 overflow-hidden">
              <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                当前目录（{{ customRoots.length }}）
              </div>
              <div class="custom-root-list-scroll min-h-[280px] flex-1 space-y-2 overflow-y-auto rounded-xl border bg-muted/10 p-2 lg:min-h-0">
                <div
                  v-for="root in customRoots"
                  :key="root.id"
                  class="rounded-xl border bg-background px-3 py-3"
                >
                  <div class="flex items-start gap-3">
                    <component
                      :is="root.icon && availableIcons[root.icon] ? availableIcons[root.icon] : availableIcons.FolderSymlink"
                      class="mt-0.5 h-4 w-4 shrink-0 text-primary/70"
                    />
                    <div class="min-w-0 flex-1 space-y-2">
                      <div class="flex flex-wrap items-center gap-2">
                        <div class="truncate text-sm font-medium">{{ root.name }}</div>
                        <span v-if="root.locked || root.isBuiltin" class="rounded-full border px-2 py-0.5 text-[11px] text-muted-foreground">
                          默认目录
                        </span>
                        <span v-if="root.pinned" class="rounded-full border border-primary/30 bg-primary/10 px-2 py-0.5 text-[11px] text-primary">
                          已置顶
                        </span>
                        <span v-if="root.enabled === false" class="rounded-full border px-2 py-0.5 text-[11px] text-muted-foreground">
                          已隐藏
                        </span>
                      </div>
                      <div class="truncate text-xs text-muted-foreground">{{ root.path }}</div>

                      <div class="flex flex-wrap items-center gap-2 pt-1">
                        <div class="flex items-center gap-2 rounded-full border px-3 py-1 text-xs" @click.stop>
                          <span class="text-muted-foreground">侧边栏显示</span>
                          <Switch
                            :model-value="root.enabled !== false"
                            :disabled="busyId === root.id"
                            @update:model-value="toggleEnabled(root, $event)"
                          />
                        </div>

                        <Button
                          variant="outline"
                          size="icon"
                          class="h-8 w-8"
                          :disabled="busyId === root.id || isFirstRoot(root)"
                          :title="isFirstRoot(root) ? '已经在最上方' : '上移一位'"
                          @click.stop="moveRoot(root, 'up')"
                        >
                          <ArrowUp class="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="outline"
                          size="icon"
                          class="h-8 w-8"
                          :disabled="busyId === root.id || isLastRoot(root)"
                          :title="isLastRoot(root) ? '已经在最下方' : '下移一位'"
                          @click.stop="moveRoot(root, 'down')"
                        >
                          <ArrowDown class="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="outline"
                          size="icon"
                          class="h-8 w-8"
                          :class="root.pinned ? 'border-primary/40 bg-primary/10 text-primary hover:bg-primary/15' : ''"
                          :disabled="busyId === root.id"
                          :title="root.pinned ? '取消置顶' : '置顶目录'"
                          @click.stop="pinRoot(root)"
                        >
                          <ChevronsUp class="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="outline"
                          size="icon"
                          class="h-8 w-8"
                          :disabled="busyId === root.id || root.locked || root.isBuiltin"
                          :title="root.locked || root.isBuiltin ? '内置目录不可编辑' : '编辑目录'"
                          @click.stop="startEdit(root)"
                        >
                          <Pencil class="h-3.5 w-3.5" />
                        </Button>
                        <Button
                          variant="outline"
                          size="icon"
                          class="h-8 w-8"
                          :disabled="busyId === root.id || root.locked || root.isBuiltin"
                          :title="root.locked || root.isBuiltin ? '内置目录不可删除' : '删除目录'"
                          @click.stop="handleDelete(root)"
                        >
                          <Trash2 class="h-3.5 w-3.5" />
                        </Button>
                        <Loader2 v-if="busyId === root.id" class="h-4 w-4 animate-spin text-muted-foreground" />
                        <Lock v-if="root.locked || root.isBuiltin" class="h-4 w-4 text-muted-foreground" />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="editingRoot" class="flex max-h-[280px] min-h-[280px] shrink-0 flex-col overflow-hidden rounded-xl border bg-muted/20">
              <div class="flex items-center justify-between gap-2 border-b px-4 py-3">
                <div>
                  <div class="text-sm font-semibold">编辑目录</div>
                  <div class="text-xs text-muted-foreground">{{ editingRoot.path }}</div>
                </div>
                <Button variant="ghost" size="icon" class="h-8 w-8" @click="resetEditState">
                  <X class="h-4 w-4" />
                </Button>
              </div>

              <div class="custom-root-edit-scroll min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
                <Input v-model="editingName" placeholder="显示名称" @keydown.enter="handleUpdate" />

                <div class="space-y-2">
                  <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    图标（{{ iconCount }}）
                  </div>
                  <div class="custom-root-icon-scroll max-h-[180px] space-y-3 overflow-y-auto rounded-xl border bg-background/70 p-3">
                    <div v-for="(icons, category) in categorizedIcons" :key="`edit-${category}`" class="space-y-2">
                      <div class="border-l-2 border-primary/30 pl-2 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
                        {{ category }}
                      </div>
                      <div class="grid grid-cols-7 gap-1.5">
                        <button
                          v-for="iconName in icons"
                          :key="`edit-${iconName}`"
                          type="button"
                          class="flex items-center justify-center rounded-md p-2 transition-all hover:scale-105"
                          :class="editingIcon === iconName ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:bg-muted'"
                          :title="iconName"
                          @click="editingIcon = iconName"
                        >
                          <component :is="availableIcons[iconName]" class="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="flex justify-end gap-2 border-t px-4 py-3">
                <Button variant="outline" @click="resetEditState">取消</Button>
                <Button class="gap-2" :disabled="isUpdating" @click="handleUpdate">
                  <Loader2 v-if="isUpdating" class="h-4 w-4 animate-spin" />
                  <Check v-else class="h-4 w-4" />
                  保存修改
                </Button>
              </div>
            </div>
          </div>

          <div class="flex h-full min-h-0 flex-col space-y-3 overflow-hidden rounded-xl border bg-muted/20 p-4">
            <div class="text-sm font-semibold">新增目录</div>

            <div class="flex gap-2">
              <div
                class="flex min-w-0 flex-1 cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-2 transition-colors hover:bg-muted/30"
                @click="selectFolder"
              >
                <FolderOpen class="h-4 w-4 shrink-0 text-muted-foreground" />
                <span class="truncate text-sm" :class="selectedPath ? 'text-foreground' : 'text-muted-foreground'">
                  {{ selectedPath || '点击选择文件夹' }}
                </span>
              </div>
              <Button variant="outline" size="icon" class="shrink-0" :disabled="isSelecting" @click="selectFolder">
                <Loader2 v-if="isSelecting" class="h-4 w-4 animate-spin" />
                <FolderOpen v-else class="h-4 w-4" />
              </Button>
            </div>

            <Input
              v-model="newName"
              placeholder="显示名称（留空则使用文件夹名）"
              @keydown.enter="handleAdd"
            />

            <div class="space-y-2">
              <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                图标（{{ iconCount }}）
              </div>
              <div class="custom-root-icon-scroll max-h-[220px] flex-1 space-y-3 overflow-y-auto rounded-xl border bg-background/70 p-3">
                <div v-for="(icons, category) in categorizedIcons" :key="category" class="space-y-2">
                  <div class="border-l-2 border-primary/30 pl-2 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
                    {{ category }}
                  </div>
                  <div class="grid grid-cols-7 gap-1.5">
                    <button
                      v-for="iconName in icons"
                      :key="iconName"
                      type="button"
                      class="flex items-center justify-center rounded-md p-2 transition-all hover:scale-105"
                      :class="selectedIcon === iconName ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:bg-muted'"
                      :title="iconName"
                      @click="selectedIcon = iconName"
                    >
                      <component :is="availableIcons[iconName]" class="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div class="shrink-0 rounded-lg border border-dashed bg-background/60 px-3 py-2 text-xs leading-5 text-muted-foreground">
              新增成功后会以多层折叠结构显示在侧边栏中，你可以随时调整显示顺序，或临时关闭显示。
            </div>

            <div class="flex shrink-0 justify-end pt-1">
              <Button class="gap-2 px-5" :disabled="!selectedPath || isAdding" @click="handleAdd">
                <Loader2 v-if="isAdding" class="h-4 w-4 animate-spin" />
                <Plus v-else class="h-4 w-4" />
                添加自定义目录
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter class="mt-5 shrink-0">
          <Button variant="outline" @click="$emit('update:open', false)">关闭</Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.custom-root-list-scroll,
.custom-root-icon-scroll,
.custom-root-edit-scroll {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.custom-root-list-scroll::-webkit-scrollbar,
.custom-root-icon-scroll::-webkit-scrollbar,
.custom-root-edit-scroll::-webkit-scrollbar {
  width: 0;
  height: 0;
  display: none;
}
</style>
