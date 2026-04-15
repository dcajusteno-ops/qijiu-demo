<script setup>
import { ref, onMounted, watch } from 'vue'
import { Link, ExternalLink, Plus, Trash2, Edit2, Save, X, Globe } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
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
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import * as App from '@/api'
import { toast } from 'vue-sonner'

const props = defineProps({
  variant: { type: String, default: 'ghost' },
  size: { type: String, default: 'icon' },
  showText: { type: Boolean, default: false },
  text: { type: String, default: '提示词助手' },
})

const builtinLinks = [
  { name: 'QPipi 提示词', url: 'https://prompt.qpipi.com/', icon: 'QP' },
  { name: 'NovelAI 标签库', url: 'https://tags.novelai.dev/', icon: 'NV' },
  { name: 'OPS 提示词', url: 'https://prompt.newzone.top/app/zh', icon: 'OP' },
  { name: 'Gwliang 提示词', url: 'https://prompt.gwliang.com/', icon: 'GW' },
  { name: 'AiWind 提示词', url: 'https://www.aiwind.org/?ref=producthunt', icon: 'AW' },
]

const customLinks = ref([])
const manageOpen = ref(false)
const deleteDialogOpen = ref(false)
const pendingDeleteLink = ref(null)
const isEditing = ref(false)
const editId = ref('')
const formData = ref({ name: '', url: '', icon: '' })

const openLink = (url) => {
  BrowserOpenURL(url)
}

const loadCustomLinks = async () => {
  try {
    const links = await App.GetPromptToolLinks()
    customLinks.value = links || []
  } catch {
    customLinks.value = []
  }
}

const resetForm = () => {
  formData.value = { name: '', url: '', icon: '' }
  isEditing.value = false
  editId.value = ''
}

const saveLink = async () => {
  if (!formData.value.name || !formData.value.url) {
    toast.error('名称和网址不能为空')
    return
  }
  try {
    if (isEditing.value) {
      await App.UpdatePromptToolLink(editId.value, formData.value)
      toast.success('链接已更新')
    } else {
      await App.AddPromptToolLink(formData.value)
      toast.success('链接已添加')
    }
    await loadCustomLinks()
    resetForm()
  } catch {
    toast.error('保存失败')
  }
}

const editLink = (link) => {
  formData.value = { name: link.name, url: link.url, icon: link.icon || '' }
  editId.value = link.id
  isEditing.value = true
}

const requestDeleteLink = (link) => {
  pendingDeleteLink.value = link
  deleteDialogOpen.value = true
}

const confirmDeleteLink = async () => {
  const link = pendingDeleteLink.value
  if (!link?.id) return
  try {
    await App.DeletePromptToolLink(link.id)
    toast.success('链接已删除')
    await loadCustomLinks()
  } catch {
    toast.error('删除失败')
  } finally {
    deleteDialogOpen.value = false
    pendingDeleteLink.value = null
  }
}

const getIconLetters = (name) => {
  if (!name) return '??'
  const chars = name.replace(/[^\p{L}\p{N}]/gu, '')
  if (chars.length >= 2) return chars.slice(0, 2).toUpperCase()
  return name.slice(0, 2).toUpperCase()
}

watch(manageOpen, (val) => {
  if (val) {
    loadCustomLinks()
    resetForm()
  }
})

onMounted(loadCustomLinks)
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger asChild>
      <Button :variant="variant" :size="size" :class="showText ? 'gap-2' : ''" :title="text">
        <Link class="w-4 h-4" />
        <span v-if="showText">{{ text }}</span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end" class="w-56">
      <DropdownMenuLabel>提示词辅助工具</DropdownMenuLabel>
      <DropdownMenuSeparator />

      <DropdownMenuItem
        v-for="link in builtinLinks"
        :key="link.url"
        class="flex items-center gap-2 cursor-pointer w-full"
        @click="openLink(link.url)"
      >
        <div class="w-5 h-5 rounded text-[9px] font-bold bg-primary/10 flex items-center justify-center shrink-0 border border-primary/20 text-primary">
          {{ link.icon }}
        </div>
        <span class="flex-1 truncate">{{ link.name }}</span>
        <ExternalLink class="w-3 h-3 text-muted-foreground opacity-50" />
      </DropdownMenuItem>

      <template v-if="customLinks.length > 0">
        <DropdownMenuSeparator />
        <DropdownMenuLabel class="text-xs text-muted-foreground">自定义链接</DropdownMenuLabel>
        <DropdownMenuItem
          v-for="link in customLinks"
          :key="link.id"
          class="flex items-center gap-2 cursor-pointer w-full"
          @click="openLink(link.url)"
        >
          <div class="w-5 h-5 rounded text-[9px] font-bold bg-blue-500/10 flex items-center justify-center shrink-0 border border-blue-500/20 text-blue-500">
            {{ link.icon || getIconLetters(link.name) }}
          </div>
          <span class="flex-1 truncate">{{ link.name }}</span>
          <ExternalLink class="w-3 h-3 text-muted-foreground opacity-50" />
        </DropdownMenuItem>
      </template>

      <DropdownMenuSeparator />
      <DropdownMenuItem class="flex items-center gap-2 cursor-pointer" @click="manageOpen = true">
        <Plus class="w-4 h-4 text-muted-foreground" />
        <span>管理自定义链接</span>
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>

  <Dialog :open="manageOpen" @update:open="manageOpen = $event">
    <DialogContent class="sm:max-w-[520px] max-h-[80vh] flex flex-col overflow-hidden p-6 select-none">
      <DialogHeader>
        <DialogTitle class="flex items-center gap-2">
          <Globe class="w-5 h-5" />
          管理自定义提示词链接
        </DialogTitle>
        <DialogDescription>
          添加、编辑或删除自定义提示词辅助网站链接
        </DialogDescription>
      </DialogHeader>

      <div class="flex-1 min-h-0 overflow-y-auto space-y-3 pr-1">
        <div v-if="customLinks.length > 0" class="space-y-2">
          <div
            v-for="link in customLinks"
            :key="link.id"
            class="flex items-center justify-between p-3 rounded-md border bg-background hover:bg-accent/50 group transition-all"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div class="w-7 h-7 rounded text-[9px] font-bold bg-blue-500/10 flex items-center justify-center shrink-0 border border-blue-500/20 text-blue-500">
                {{ link.icon || getIconLetters(link.name) }}
              </div>
              <div class="min-w-0">
                <div class="font-medium text-sm truncate">{{ link.name }}</div>
                <div class="text-xs text-muted-foreground truncate opacity-70" :title="link.url">{{ link.url }}</div>
              </div>
            </div>
            <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
              <Button size="icon" variant="ghost" class="h-7 w-7" title="编辑" @click="editLink(link)">
                <Edit2 class="w-3 h-3" />
              </Button>
              <Button
                size="icon"
                variant="ghost"
                class="h-7 w-7 text-destructive hover:text-destructive hover:bg-destructive/10"
                title="删除"
                @click="requestDeleteLink(link)"
              >
                <Trash2 class="w-3 h-3" />
              </Button>
            </div>
          </div>
        </div>
        <div v-else class="flex flex-col items-center justify-center py-6 text-muted-foreground opacity-50">
          <Globe class="w-8 h-8 mb-2" />
          <span class="text-sm">暂无自定义链接</span>
        </div>

        <div class="border rounded-md p-3 space-y-3 bg-muted/20">
          <div class="flex items-center justify-between">
            <h4 class="text-sm font-semibold">{{ isEditing ? '编辑链接' : '添加新链接' }}</h4>
            <Button v-if="isEditing" size="icon" variant="ghost" class="h-6 w-6" @click="resetForm">
              <X class="w-3 h-3" />
            </Button>
          </div>
          <div class="space-y-2">
            <div class="space-y-1">
              <Label class="text-xs">名称</Label>
              <Input v-model="formData.name" placeholder="例如：Civitai" class="h-8 select-text" />
            </div>
            <div class="space-y-1">
              <Label class="text-xs">网址</Label>
              <Input v-model="formData.url" placeholder="https://example.com" class="h-8 font-mono text-xs select-text" />
            </div>
            <div class="space-y-1">
              <Label class="text-xs">图标文字（可选，最多 2 个字符）</Label>
              <Input v-model="formData.icon" placeholder="例如：CV" class="h-8 select-text" maxlength="2" />
            </div>
          </div>
          <div class="flex justify-end gap-2">
            <Button v-if="isEditing" variant="ghost" size="sm" @click="resetForm">取消</Button>
            <Button size="sm" @click="saveLink">
              <Save class="w-3 h-3 mr-2" />
              {{ isEditing ? '更新' : '添加' }}
            </Button>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>

  <AlertDialog v-model:open="deleteDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>确认删除链接</AlertDialogTitle>
        <AlertDialogDescription>
          确定要删除“{{ pendingDeleteLink?.name || '该链接' }}”吗？
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="pendingDeleteLink = null">取消</AlertDialogCancel>
        <AlertDialogAction @click="confirmDeleteLink" class="bg-destructive hover:bg-destructive/90">
          删除
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
