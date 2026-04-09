<script setup>
import { computed, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import { Badge } from '@/components/ui/badge'
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
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Check, Heart, Pencil, Plus, Trash2, X } from 'lucide-vue-next'

const props = defineProps({
  open: { type: Boolean, default: false },
  groups: { type: Array, default: () => [] },
  image: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'change'])

const newGroupName = ref('')
const editingGroupId = ref('')
const editingGroupName = ref('')
const selectedGroupIds = ref([])
const deleteGroupId = ref('')

const isSavingAssignments = ref(false)
const isCreating = ref(false)
const isUpdating = ref(false)
const isDeleting = ref(false)

const isImageMode = computed(() => Boolean(props.image?.relPath))
const currentImageName = computed(() => props.image?.name || '')

const resetState = () => {
  newGroupName.value = ''
  editingGroupId.value = ''
  editingGroupName.value = ''
  deleteGroupId.value = ''
  if (props.image?.relPath) {
    selectedGroupIds.value = (props.groups || [])
      .filter((group) => (group.paths || []).includes(props.image.relPath))
      .map((group) => group.id)
  } else {
    selectedGroupIds.value = []
  }
}

watch(
  () => [props.open, props.groups, props.image?.relPath],
  () => {
    if (props.open) {
      resetState()
    }
  },
  { deep: true },
)

const normalizeError = (error, fallback) => {
  const message = String(error ?? '').trim()
  if (!message || message.includes('锟')) {
    return fallback
  }
  return message
}

const toggleGroupSelection = (groupId, checked) => {
  const next = new Set(selectedGroupIds.value)
  if (checked) next.add(groupId)
  else next.delete(groupId)
  selectedGroupIds.value = Array.from(next)
}

const saveAssignments = async () => {
  if (!props.image?.relPath) {
    emit('update:open', false)
    return
  }

  isSavingAssignments.value = true
  try {
    await App.SetImageFavoriteGroups(props.image.relPath, selectedGroupIds.value)
    toast.success('收藏分组已更新')
    emit('change')
    emit('update:open', false)
  } catch (error) {
    toast.error(normalizeError(error, '更新收藏分组失败'))
  } finally {
    isSavingAssignments.value = false
  }
}

const createGroup = async () => {
  const name = newGroupName.value.trim()
  if (!name) {
    toast.error('请输入分组名称')
    return
  }

  isCreating.value = true
  try {
    const group = await App.CreateFavoriteGroup(name)
    if (props.image?.relPath) {
      const next = Array.from(new Set([...selectedGroupIds.value, group.id]))
      await App.SetImageFavoriteGroups(props.image.relPath, next)
      selectedGroupIds.value = next
    }
    newGroupName.value = ''
    toast.success('收藏分组已创建')
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '创建收藏分组失败'))
  } finally {
    isCreating.value = false
  }
}

const startRename = (group) => {
  editingGroupId.value = group.id
  editingGroupName.value = group.name || ''
}

const cancelRename = () => {
  editingGroupId.value = ''
  editingGroupName.value = ''
}

const saveRename = async () => {
  const name = editingGroupName.value.trim()
  if (!editingGroupId.value || !name) {
    toast.error('请输入分组名称')
    return
  }

  isUpdating.value = true
  try {
    await App.UpdateFavoriteGroup(editingGroupId.value, name)
    toast.success('分组名称已更新')
    cancelRename()
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '更新分组失败'))
  } finally {
    isUpdating.value = false
  }
}

const confirmDelete = async () => {
  if (!deleteGroupId.value) return

  isDeleting.value = true
  try {
    await App.DeleteFavoriteGroup(deleteGroupId.value)
    selectedGroupIds.value = selectedGroupIds.value.filter((id) => id !== deleteGroupId.value)
    toast.success('收藏分组已删除')
    deleteGroupId.value = ''
    emit('change')
  } catch (error) {
    toast.error(normalizeError(error, '删除分组失败'))
  } finally {
    isDeleting.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[560px] overflow-hidden p-0">
      <div class="max-h-[82vh] overflow-y-auto p-5">
        <DialogHeader class="pr-8">
          <DialogTitle class="flex items-center gap-2">
            <Heart class="h-5 w-5 text-red-500" />
            收藏夹分组
          </DialogTitle>
          <DialogDescription>
            创建收藏分组，或把图片放进你常用的收藏夹分类里。
          </DialogDescription>
        </DialogHeader>

        <div class="mt-5 space-y-4">
          <div v-if="isImageMode" class="space-y-3 rounded-md border bg-muted/20 p-3">
            <div class="space-y-1">
              <div class="text-sm font-semibold">当前图片</div>
              <div class="truncate text-xs text-muted-foreground">{{ currentImageName }}</div>
            </div>

            <div class="space-y-2">
              <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                选择分组
              </div>
              <div class="max-h-[180px] space-y-2 overflow-y-auto rounded-md border bg-background p-2">
                <label
                  v-for="group in groups"
                  :key="group.id"
                  class="flex cursor-pointer items-center justify-between gap-3 rounded-md border px-3 py-2 transition-colors hover:bg-muted/50"
                >
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-medium">{{ group.name }}</div>
                    <div class="text-xs text-muted-foreground">{{ (group.paths || []).length }} 张</div>
                  </div>
                  <input
                    type="checkbox"
                    class="h-4 w-4 accent-primary"
                    :checked="selectedGroupIds.includes(group.id)"
                    @change="toggleGroupSelection(group.id, $event.target.checked)"
                  />
                </label>

                <div
                  v-if="groups.length === 0"
                  class="rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground"
                >
                  还没有收藏分组
                </div>
              </div>
            </div>
          </div>

          <div class="space-y-3 rounded-md border bg-muted/20 p-3">
            <div class="text-sm font-semibold">新建分组</div>
            <div class="flex gap-2">
              <Input
                v-model="newGroupName"
                placeholder="例如：待精修、成品、参考构图"
                @keydown.enter.prevent="createGroup"
              />
              <Button class="gap-2 shrink-0" :disabled="isCreating" @click="createGroup">
                <Plus class="h-4 w-4" />
                新建
              </Button>
            </div>
          </div>

          <div class="space-y-2">
            <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              已有分组（{{ groups.length }}）
            </div>
            <div class="max-h-[220px] space-y-2 overflow-y-auto rounded-md border bg-muted/10 p-2">
              <div
                v-for="group in groups"
                :key="group.id"
                class="rounded-md border bg-background px-3 py-3"
              >
                <div v-if="editingGroupId === group.id" class="space-y-2">
                  <Input
                    v-model="editingGroupName"
                    @keydown.enter.prevent="saveRename"
                  />
                  <div class="flex justify-end gap-2">
                    <Button variant="outline" size="sm" @click="cancelRename">
                      <X class="mr-1 h-3.5 w-3.5" />
                      取消
                    </Button>
                    <Button size="sm" :disabled="isUpdating" @click="saveRename">
                      <Check class="mr-1 h-3.5 w-3.5" />
                      保存
                    </Button>
                  </div>
                </div>

                <div v-else class="flex items-center gap-3">
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <span class="truncate text-sm font-medium">{{ group.name }}</span>
                      <Badge variant="secondary">{{ (group.paths || []).length }}</Badge>
                      <Badge v-if="group.id === 'default'" variant="outline">默认</Badge>
                    </div>
                  </div>
                  <div class="flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      class="h-8 w-8 text-muted-foreground hover:bg-primary/10 hover:text-primary"
                      @click="startRename(group)"
                    >
                      <Pencil class="h-3.5 w-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      class="h-8 w-8 text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                      :disabled="group.id === 'default'"
                      @click="deleteGroupId = group.id"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                    </Button>
                  </div>
                </div>
              </div>

              <div
                v-if="groups.length === 0"
                class="rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground"
              >
                还没有创建收藏分组
              </div>
            </div>
          </div>
        </div>
      </div>

      <DialogFooter class="border-t px-5 py-4">
        <div class="flex w-full justify-end gap-2">
          <Button variant="outline" @click="$emit('update:open', false)">关闭</Button>
          <Button v-if="isImageMode" :disabled="isSavingAssignments" @click="saveAssignments">
            保存当前图片分组
          </Button>
        </div>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <AlertDialog :open="Boolean(deleteGroupId)">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>删除收藏分组</AlertDialogTitle>
        <AlertDialogDescription>
          删除后不会删除图片，只会移除这个分组本身。
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="deleteGroupId = ''">取消</AlertDialogCancel>
        <AlertDialogAction :disabled="isDeleting" @click="confirmDelete">确认删除</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
