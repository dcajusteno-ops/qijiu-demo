<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import * as App from '@/api'
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
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Trash2, RotateCcw, Settings, CheckSquare, Square, Inbox } from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const props = defineProps({
  open: { type: Boolean, default: false }
})

const emit = defineEmits(['update:open', 'refresh'])

const trashItems = ref([])
const selectedFilenames = ref(new Set())
const loading = ref(false)
const showSettings = ref(false)
const retentionDays = ref(30)
const showEmptyConfirm = ref(false)
const batchDeleting = ref(false)
const batchRestoring = ref(false)

// Load trash items
const loadTrash = async () => {
  loading.value = true
  try {
    const list = await App.GetTrashList()
    trashItems.value = list || []
    selectedFilenames.value = new Set()
  } catch (e) {
    toast.error('加载回收站失败')
  } finally {
    loading.value = false
  }
}

// Load settings
const loadSettings = async () => {
  try {
    const data = await App.GetTrashSettings()
    retentionDays.value = data.trashRetentionDays
  } catch (e) {
    console.error(e)
  }
}

// Save settings
const saveSettings = async () => {
  try {
    await App.SaveTrashSettings({ trashRetentionDays: retentionDays.value })
    toast.success('设置已保存')
    showSettings.value = false
  } catch (e) {
    toast.error('请求出错')
  }
}

// Restore single item
const restoreItem = async (filename) => {
  try {
    await App.RestoreTrash(filename)
    toast.success('已还原')
    await loadTrash()
    emit('refresh')
  } catch (e) {
    toast.error('还原失败')
  }
}

// Batch Restore
const restoreSelected = async () => {
  if (selectedFilenames.value.size === 0) return
  batchRestoring.value = true
  try {
    const successCount = await App.BatchRestoreTrash(Array.from(selectedFilenames.value))
    toast.success(`成功还原 ${successCount} 个文件`)
    await loadTrash()
    emit('refresh')
  } catch (e) {
    toast.error('批量还原过程中出现错误')
  } finally {
    batchRestoring.value = false
  }
}

// Batch Delete
const deleteSelected = async () => {
  if (selectedFilenames.value.size === 0) return
  batchDeleting.value = true
  try {
    const count = await App.BatchDeleteTrash(Array.from(selectedFilenames.value))
    toast.success(`永久删除 ${count} 个文件`)
    await loadTrash()
  } catch (e) {
    toast.error('批量删除失败')
  } finally {
    batchDeleting.value = false
  }
}

// Empty trash
const confirmEmptyTrash = async () => {
  showEmptyConfirm.value = false
  try {
    const count = await App.EmptyTrash()
    toast.success(`已清空 ${count} 个文件`)
    await loadTrash()
    emit('refresh')
  } catch (e) {
    toast.error('清空失败')
  }
}

// Selection handling
const toggleSelect = (filename) => {
  const newSet = new Set(selectedFilenames.value)
  if (newSet.has(filename)) {
    newSet.delete(filename)
  } else {
    newSet.add(filename)
  }
  selectedFilenames.value = newSet
}

const toggleSelectAll = () => {
  if (selectedFilenames.value.size === trashItems.value.length) {
    selectedFilenames.value = new Set()
  } else {
    selectedFilenames.value = new Set(trashItems.value.map(i => i.filename))
  }
}

// Format deleted time
const formatDeletedTime = (deletedAt) => {
  const now = new Date()
  const deleted = new Date(deletedAt)
  
  // Calculate day difference based on calendar days
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const thatDay = new Date(deleted.getFullYear(), deleted.getMonth(), deleted.getDate())
  const diffDays = Math.floor((today - thatDay) / (1000 * 60 * 60 * 24))
  
  const timeStr = deleted.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
  
  if (diffDays === 0) return `今天 ${timeStr}`
  if (diffDays === 1) return `昨天 ${timeStr}`
  if (diffDays < 7) return `${diffDays} 天前`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} 周前`
  return `${Math.floor(diffDays / 30)} 个月前`
}

onMounted(() => {
  if (props.open) {
    loadTrash()
    loadSettings()
  }
})

watch(() => props.open, (newVal) => {
  if (newVal) {
    loadTrash()
    loadSettings()
  }
})

// Watch for dialog open
const handleOpenChange = (isOpen) => {
  emit('update:open', isOpen)
  if (isOpen) {
    loadTrash()
    loadSettings()
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[1000px] h-[85vh] flex flex-col p-6">
      <DialogHeader class="shrink-0 mb-2">
        <DialogTitle class="flex items-center justify-between">
          <span>回收站管理</span>
        </DialogTitle>
        <DialogDescription>
          查看和批量还原已删除的图片
        </DialogDescription>
      </DialogHeader>

      <!-- Batch Action Bar -->
      <div v-if="trashItems.length > 0" class="flex items-center justify-between mb-4 px-1">
        <div class="flex items-center gap-2">
          <Button variant="ghost" size="sm" @click="toggleSelectAll" class="gap-2">
            <template v-if="selectedFilenames.size === trashItems.length && trashItems.length > 0">
               <CheckSquare class="h-4 w-4" /> 取消全选
            </template>
            <template v-else>
               <Square class="h-4 w-4" /> 全选
            </template>
          </Button>
          <span class="text-sm text-muted-foreground">已选择 {{ selectedFilenames.size }} 项</span>
        </div>
        
        <div v-if="selectedFilenames.size > 0" class="flex items-center gap-2 animate-in fade-in slide-in-from-right-2">
           <Button size="sm" variant="outline" @click="restoreSelected" :disabled="batchRestoring">
             <RotateCcw class="h-3 w-3 mr-1" />
             批量还原
           </Button>
           <Button size="sm" variant="destructive" @click="deleteSelected" :disabled="batchDeleting">
             <Trash2 class="h-3 w-3 mr-1" />
             批量删除
           </Button>
        </div>
      </div>

      <!-- Settings Panel (Collapsible) -->
      <div v-if="showSettings" class="shrink-0 border rounded-md p-4 mb-4 space-y-3 bg-muted/30">
        <div class="flex items-center gap-3">
          <Label class="whitespace-nowrap">自动清理设置:</Label>
          <div class="flex items-center gap-2">
            <span class="text-sm">删除超过</span>
            <Input type="number" v-model="retentionDays" class="w-20 text-center" min="1" />
            <span class="text-sm">天的文件</span>
          </div>
          <Button size="sm" @click="saveSettings">保存</Button>
        </div>
        <p class="text-xs text-muted-foreground">
          * 每次启动服务时，会自动删除超过设定天数的回收站文件
        </p>
      </div>

      <!-- Trash Items Grid Area -->
      <div class="flex-1 min-h-0 border rounded-md overflow-hidden bg-muted/5 relative">
        <div v-if="loading && trashItems.length === 0" class="absolute inset-0 flex items-center justify-center">
          <p class="text-muted-foreground">加载中...</p>
        </div>

        <div v-else-if="trashItems.length === 0" class="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground">
          <Inbox class="h-16 w-16 mb-4 opacity-10" />
          <p class="text-lg font-medium">回收站为空</p>
          <p class="text-sm opacity-60">已删除的图片将显示在这里</p>
        </div>

        <ScrollArea v-else class="h-full">
          <div class="grid grid-cols-4 lg:grid-cols-5 gap-4 p-4">
            <div 
              v-for="item in trashItems" 
              :key="item.filename"
              class="group relative border rounded-md overflow-hidden bg-card transition-all cursor-pointer select-none"
              :class="{ 
                'ring-2 ring-primary bg-primary/5 shadow-md': selectedFilenames.has(item.filename),
                'hover:border-primary/50': !selectedFilenames.has(item.filename)
              }"
              @click="toggleSelect(item.filename)"
            >
                <!-- Image Container -->
              <div class="aspect-square bg-muted relative overflow-hidden">
                <img 
                  :src="item.path" 
                  :alt="item.filename"
                  class="w-full h-full object-cover transition-transform group-hover:scale-105"
                  :class="{ 'opacity-80': selectedFilenames.has(item.filename) }"
                  loading="lazy"
                  decoding="async"
                />
                
                <!-- Selection Indicator (Simple Dot or Glow) -->
                <div v-if="selectedFilenames.has(item.filename)" class="absolute top-2 left-2 h-4 w-4 rounded-full bg-primary border-2 border-background shadow-sm z-10"></div>

                <!-- Hover Quick Restore (Only shown when not selecting many) -->
                <div v-if="selectedFilenames.size === 0" class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                  <Button size="sm" variant="secondary" class="h-8" @click.stop="restoreItem(item.filename)">
                    <RotateCcw class="h-4 w-4 mr-1" />
                    还原
                  </Button>
                </div>
              </div>

              <!-- Info Area -->
              <div class="p-2 space-y-1">
                <p class="text-[11px] font-medium truncate" :title="item.filename">{{ item.filename }}</p>
                <div class="flex items-center gap-1 text-[10px] text-muted-foreground">
                  <span class="shrink-0 truncate max-w-[150px]" :title="item.originalPath">
                    来自: {{ item.originalPath.split('/').length <= 1 ? `自动分类/${item.originalPath.split('/')[0]}` : item.originalPath }}
                  </span>
                </div>
                <p class="text-[10px] text-muted-foreground/60">
                  {{ formatDeletedTime(item.deletedAt) }}
                </p>
              </div>
            </div>
          </div>
        </ScrollArea>
      </div>

      <!-- Footer Area -->
      <DialogFooter class="shrink-0 mt-6 flex justify-between items-center sm:justify-between border-t pts-4 pt-4">
        <div class="flex items-center gap-2">
          <Button 
            variant="destructive" 
            size="sm"
            @click="showEmptyConfirm = true"
            :disabled="trashItems.length === 0"
          >
            <Trash2 class="h-4 w-4 mr-2" />
            清空回收站
          </Button>
          <Button 
            variant="ghost" 
            size="sm" 
            @click="showSettings = !showSettings" 
            :class="{ 'bg-accent': showSettings }"
          >
            <Settings class="h-4 w-4 mr-2" />
            自动清理设置
          </Button>
        </div>
        <Button variant="ghost" size="sm" @click="$emit('update:open', false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>

  <!-- Empty Trash Confirmation -->
  <AlertDialog :open="showEmptyConfirm" @update:open="showEmptyConfirm = $event">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>确认清空回收站</AlertDialogTitle>
        <AlertDialogDescription>
          此操作将永久删除回收站中的所有文件（共 {{ trashItems.length }} 个），无法撤销。您确定要继续吗？
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>取消</AlertDialogCancel>
        <AlertDialogAction @click="confirmEmptyTrash" class="bg-destructive hover:bg-destructive/90">
          确认删除
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
