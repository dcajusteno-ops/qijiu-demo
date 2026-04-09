<script setup>
import { computed, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Check,
  FolderOpen,
  FolderSymlink,
  Loader2,
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
    toast.error(normalizeError(error, '所选目录不在当前 output 目录内'))
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

const handleDelete = async (id) => {
  try {
    await App.DeleteCustomRoot(id)
    toast.success('自定义目录已删除')
    if (editingRootId.value === id) {
      resetEditState()
    }
    emit('change')
  } catch (error) {
    console.error(error)
    toast.error(normalizeError(error, '删除失败'))
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[520px] overflow-hidden p-0">
      <div class="max-h-[82vh] overflow-y-auto p-5">
        <DialogHeader class="pr-8">
          <DialogTitle class="flex items-center gap-2">
            <FolderSymlink class="h-5 w-5 text-primary" />
            管理自定义目录
          </DialogTitle>
          <DialogDescription>
            自定义目录会作为侧边栏里的独立入口显示。这里可以新增目录，也可以修改已有目录的显示名称和图标。
          </DialogDescription>
        </DialogHeader>

        <div class="mt-5 space-y-4">
          <div class="space-y-2">
            <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              已配置（{{ customRoots.length }}）
            </div>
            <div class="max-h-[170px] space-y-2 overflow-y-auto rounded-md border bg-muted/10 p-2">
              <div
                v-if="customRoots.length === 0"
                class="flex flex-col items-center justify-center py-8 text-muted-foreground opacity-50"
              >
                <FolderSymlink class="mb-2 h-8 w-8" />
                <span class="text-sm">暂无自定义目录</span>
              </div>

              <div
                v-for="root in customRoots"
                :key="root.id"
                class="flex min-h-[64px] items-center gap-3 rounded-md border bg-background px-3 py-2"
              >
                <component
                  :is="root.icon && availableIcons[root.icon] ? availableIcons[root.icon] : availableIcons.FolderSymlink"
                  class="h-4 w-4 shrink-0 text-primary/70"
                />
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-medium">{{ root.name }}</div>
                  <div class="truncate text-xs text-muted-foreground opacity-70">{{ root.path }}</div>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 text-muted-foreground hover:bg-primary/10 hover:text-primary"
                    title="编辑"
                    @click="startEdit(root)"
                  >
                    <Pencil class="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                    title="删除"
                    @click="handleDelete(root.id)"
                  >
                    <Trash2 class="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <div v-if="editingRoot" class="space-y-3 rounded-md border bg-muted/20 p-3">
            <div class="flex items-center justify-between gap-2">
              <div class="min-w-0">
                <div class="text-sm font-semibold">编辑自定义目录</div>
                <div class="truncate text-xs text-muted-foreground">{{ editingRoot.path }}</div>
              </div>
              <Button variant="ghost" size="icon" class="h-8 w-8 shrink-0" title="取消编辑" @click="resetEditState">
                <X class="h-4 w-4" />
              </Button>
            </div>

            <Input
              v-model="editingName"
              placeholder="显示名称"
              @keydown.enter="handleUpdate"
            />

            <div class="space-y-2">
              <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                更改图标（{{ iconCount }}）
              </div>
              <div class="max-h-[112px] space-y-3 overflow-y-auto rounded-md border bg-background/70 p-3">
                <div v-for="(icons, category) in categorizedIcons" :key="`edit-${category}`" class="space-y-2">
                  <div class="border-l-2 border-primary/30 pl-1 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
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

            <div class="flex justify-end gap-2">
              <Button variant="outline" @click="resetEditState">取消</Button>
              <Button class="gap-2" :disabled="isUpdating" @click="handleUpdate">
                <Loader2 v-if="isUpdating" class="h-4 w-4 animate-spin" />
                <Check v-else class="h-4 w-4" />
                保存更改
              </Button>
            </div>
          </div>

          <div v-else class="space-y-3 rounded-md border bg-muted/20 p-3">
            <div class="text-sm font-semibold">添加新目录</div>

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
              <Button
                variant="outline"
                size="icon"
                class="shrink-0"
                :disabled="isSelecting"
                title="选择文件夹"
                @click="selectFolder"
              >
                <Loader2 v-if="isSelecting" class="h-4 w-4 animate-spin" />
                <FolderOpen v-else class="h-4 w-4" />
              </Button>
            </div>

            <Input
              v-model="newName"
              placeholder="显示名称（留空则默认使用文件夹名）"
              @keydown.enter="handleAdd"
            />

            <div class="space-y-2">
              <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                选择图标（{{ iconCount }}）
              </div>
              <div class="max-h-[128px] space-y-3 overflow-y-auto rounded-md border bg-background/70 p-3">
                <div v-for="(icons, category) in categorizedIcons" :key="category" class="space-y-2">
                  <div class="border-l-2 border-primary/30 pl-1 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
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

            <div class="flex justify-end pt-1">
              <Button class="gap-2 px-5" :disabled="!selectedPath || isAdding" @click="handleAdd">
                <Loader2 v-if="isAdding" class="h-4 w-4 animate-spin" />
                <Plus v-else class="h-4 w-4" />
                添加自定义目录
              </Button>
            </div>
          </div>
        </div>

        <DialogFooter class="mt-5">
          <Button variant="outline" @click="$emit('update:open', false)">关闭</Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>
