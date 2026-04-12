<script setup>
import { computed, ref, watch } from 'vue'
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
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Switch } from '@/components/ui/switch'
import {
  Plus,
  Trash2,
  Edit2,
  Copy,
  X,
  Save,
  Bookmark,
  Search,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import * as App from '@/api'

const props = defineProps({
  open: { type: Boolean, default: false },
  initialContent: { type: String, default: '' },
  initialType: { type: String, default: '' },
  initialSourcePath: { type: String, default: '' },
})

const emit = defineEmits(['update:open'])

const templates = ref([])
const searchQuery = ref('')
const activeTypeFilter = ref('all')
const activeCategoryFilter = ref('all')
const showForm = ref(false)
const isEditing = ref(false)
const formId = ref('')
const deleteDialogOpen = ref(false)
const pendingDeleteTemplate = ref(null)

const formData = ref({
  name: '',
  content: '',
  type: 'positive',
  category: '',
})

const typeFilters = [
  { key: 'all', label: '全部' },
  { key: 'positive', label: '正向' },
  { key: 'negative', label: '反向' },
  { key: 'other', label: '其他' },
]

const typeLabelMap = {
  positive: '正向',
  negative: '反向',
  other: '其他',
}

const allCategories = computed(() => {
  const cats = new Set()
  templates.value.forEach(t => {
    if (t.category) cats.add(t.category)
  })
  return Array.from(cats).sort()
})

const filteredTemplates = computed(() => {
  let list = templates.value
  if (activeTypeFilter.value !== 'all') {
    list = list.filter(t => t.type === activeTypeFilter.value)
  }
  if (activeCategoryFilter.value !== 'all') {
    list = list.filter(t => t.category === activeCategoryFilter.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter(t =>
      t.name.toLowerCase().includes(q) ||
      t.content.toLowerCase().includes(q)
    )
  }
  return list
})

const resetForm = () => {
  formData.value = { name: '', content: '', type: 'positive', category: '' }
  isEditing.value = false
  formId.value = ''
  showForm.value = false
}

const loadTemplates = async () => {
  try {
    const list = await App.GetPromptTemplates()
    templates.value = list || []
  } catch (e) {
    templates.value = []
  }
}

const saveTemplate = async () => {
  if (!formData.value.name.trim()) {
    toast.error('模板名称不能为空')
    return
  }
  if (!formData.value.content.trim()) {
    toast.error('提示词内容不能为空')
    return
  }

  try {
    if (isEditing.value) {
      await App.UpdatePromptTemplate(formId.value, formData.value)
      toast.success('模板已更新')
    } else {
      await App.AddPromptTemplate(formData.value)
      toast.success('模板已添加')
    }
    await loadTemplates()
    resetForm()
  } catch (e) {
    toast.error('保存失败')
  }
}

const editTemplate = (template) => {
  formData.value = {
    name: template.name,
    content: template.content,
    type: template.type || 'positive',
    category: template.category || '',
  }
  formId.value = template.id
  isEditing.value = true
  showForm.value = true
}

const requestDeleteTemplate = (template) => {
  pendingDeleteTemplate.value = template
  deleteDialogOpen.value = true
}

const confirmDeleteTemplate = async () => {
  const template = pendingDeleteTemplate.value
  if (!template?.id) return

  try {
    await App.DeletePromptTemplate(template.id)
    toast.success('模板已删除')
    await loadTemplates()
  } catch (e) {
    toast.error('删除失败')
  } finally {
    deleteDialogOpen.value = false
    pendingDeleteTemplate.value = null
  }
}

const copyContent = async (template) => {
  try {
    await App.CopyText(template.content)
    toast.success('已复制到剪贴板')
  } catch (e) {
    toast.error('复制失败')
  }
}

const truncate = (text, max = 80) => {
  if (!text) return ''
  return text.length > max ? text.slice(0, max) + '...' : text
}

watch(() => props.open, (newVal) => {
  if (newVal) {
    loadTemplates()
    resetForm()
    // If initial data provided (from Lightbox "save as template"), open form
    if (props.initialContent) {
      formData.value = {
        name: '',
        content: props.initialContent,
        type: props.initialType || 'positive',
        category: '',
      }
      showForm.value = true
    }
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[920px] h-[74vh] max-h-[88vh] flex flex-col overflow-hidden p-6 select-none">
      <DialogHeader class="shrink-0 mb-4">
        <DialogTitle class="flex items-center gap-2">
          <Bookmark class="w-5 h-5" />
          <span>提示词模板库</span>
        </DialogTitle>
        <DialogDescription>
          保存、管理和复用你的常用提示词模板
        </DialogDescription>
      </DialogHeader>

      <div class="flex-1 min-h-0 flex gap-4 overflow-hidden">
        <!-- Left: Template List -->
        <div class="min-w-0 flex-1 flex flex-col border rounded-md overflow-hidden bg-card">
          <!-- Search & Filters -->
          <div class="p-2 border-b bg-muted/30 space-y-2">
            <div class="relative">
              <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                placeholder="搜索模板名称或内容..."
                class="h-8 pl-8 text-xs select-text"
              />
            </div>
            <div class="flex gap-1 flex-wrap">
              <Button
                v-for="f in typeFilters"
                :key="f.key"
                :variant="activeTypeFilter === f.key ? 'secondary' : 'ghost'"
                size="sm"
                class="h-6 px-2 text-[11px]"
                @click="activeTypeFilter = f.key"
              >
                {{ f.label }}
              </Button>
              <template v-if="allCategories.length > 0">
                <span class="text-muted-foreground/40 mx-1">|</span>
                <Button
                  :variant="activeCategoryFilter === 'all' ? 'secondary' : 'ghost'"
                  size="sm"
                  class="h-6 px-2 text-[11px]"
                  @click="activeCategoryFilter = 'all'"
                >
                  全部分类
                </Button>
                <Button
                  v-for="cat in allCategories"
                  :key="cat"
                  :variant="activeCategoryFilter === cat ? 'secondary' : 'ghost'"
                  size="sm"
                  class="h-6 px-2 text-[11px]"
                  @click="activeCategoryFilter = cat"
                >
                  {{ cat }}
                </Button>
              </template>
            </div>
          </div>

          <!-- List -->
          <ScrollArea class="flex-1">
            <div class="p-2 space-y-2">
              <div v-if="filteredTemplates.length === 0" class="flex flex-col items-center justify-center py-10 text-muted-foreground opacity-50">
                <Bookmark class="w-10 h-10 mb-2" />
                <span class="text-sm">{{ templates.length === 0 ? '暂无模板' : '没有匹配的模板' }}</span>
              </div>

              <div
                v-for="template in filteredTemplates"
                :key="template.id"
                class="flex items-start justify-between p-3 rounded-md border bg-background hover:bg-accent/50 group transition-all"
              >
                <div class="flex-1 min-w-0 space-y-1">
                  <div class="flex items-center gap-2">
                    <span class="font-medium text-sm truncate">{{ template.name }}</span>
                    <Badge variant="secondary" class="shrink-0 text-[10px] px-1.5 py-0">
                      {{ typeLabelMap[template.type] || template.type }}
                    </Badge>
                    <Badge v-if="template.category" variant="outline" class="shrink-0 text-[10px] px-1.5 py-0">
                      {{ template.category }}
                    </Badge>
                  </div>
                  <div class="text-xs text-muted-foreground leading-relaxed whitespace-pre-wrap break-words line-clamp-2">
                    {{ truncate(template.content, 120) }}
                  </div>
                </div>

                <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0 ml-2">
                  <Button size="icon" variant="ghost" class="h-7 w-7 text-primary hover:text-primary hover:bg-primary/10" @click="copyContent(template)" title="复制内容">
                    <Copy class="w-3 h-3" />
                  </Button>
                  <Button size="icon" variant="ghost" class="h-7 w-7" @click="editTemplate(template)" title="编辑">
                    <Edit2 class="w-3 h-3" />
                  </Button>
                  <Button size="icon" variant="ghost" class="h-7 w-7 text-destructive hover:text-destructive hover:bg-destructive/10" @click="requestDeleteTemplate(template)" title="删除">
                    <Trash2 class="w-3 h-3" />
                  </Button>
                </div>
              </div>
            </div>
          </ScrollArea>

          <!-- Add button -->
          <div class="p-2 border-t bg-muted/30">
            <Button size="sm" variant="ghost" class="w-full gap-2 text-muted-foreground" @click="resetForm(); showForm = true">
              <Plus class="w-4 h-4" />
              添加模板
            </Button>
          </div>
        </div>

        <!-- Right: Form -->
        <div v-if="showForm" class="w-[380px] min-h-0 border-l pl-4 flex flex-col overflow-hidden animate-in slide-in-from-right-5 fade-in duration-300">
          <div class="flex shrink-0 items-center justify-between mb-4">
            <h3 class="font-semibold text-sm">{{ isEditing ? '编辑模板' : '添加新模板' }}</h3>
            <Button size="icon" variant="ghost" class="h-6 w-6" @click="showForm = false">
              <X class="w-4 h-4" />
            </Button>
          </div>

          <div class="space-y-3 flex-1 min-h-0 overflow-y-auto pr-1">
            <div class="space-y-1">
              <Label class="text-xs">模板名称</Label>
              <Input v-model="formData.name" placeholder="例如: 高质量人物" class="h-8 select-text" />
            </div>

            <div class="space-y-1">
              <Label class="text-xs">提示词内容</Label>
              <textarea
                v-model="formData.content"
                placeholder="输入提示词内容..."
                class="flex min-h-[160px] w-full rounded-md border border-input bg-background px-3 py-2 text-xs leading-6 ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 select-text resize-y"
              />
            </div>

            <div class="space-y-1">
              <Label class="text-xs">类型</Label>
              <div class="flex gap-2">
                <Button
                  v-for="f in typeFilters.slice(1)"
                  :key="f.key"
                  :variant="formData.type === f.key ? 'secondary' : 'outline'"
                  size="sm"
                  class="h-7 text-xs"
                  @click="formData.type = f.key"
                >
                  {{ f.label }}
                </Button>
              </div>
            </div>

            <div class="space-y-1">
              <Label class="text-xs">分类 (可选)</Label>
              <Input v-model="formData.category" placeholder="例如: 人物、画风、质量词" class="h-8 select-text" />
            </div>
          </div>

          <div class="mt-4 shrink-0 border-t pt-4 flex justify-end gap-2 bg-background">
            <Button variant="ghost" size="sm" @click="showForm = false">取消</Button>
            <Button size="sm" @click="saveTemplate">
              <Save class="w-3 h-3 mr-2" />
              {{ isEditing ? '更新' : '保存' }}
            </Button>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>

  <AlertDialog v-model:open="deleteDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>确认删除模板</AlertDialogTitle>
        <AlertDialogDescription>
          确定要删除"{{ pendingDeleteTemplate?.name || '该模板' }}"吗？此操作不可撤销。
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="pendingDeleteTemplate = null">取消</AlertDialogCancel>
        <AlertDialogAction @click="confirmDeleteTemplate" class="bg-destructive hover:bg-destructive/90">
          删除
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
