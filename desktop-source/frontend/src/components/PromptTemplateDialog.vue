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

defineEmits(['update:open'])

const templates = ref([])
const searchQuery = ref('')
const activeTypeFilter = ref('all')
const activeCategoryFilter = ref('all')
const showForm = ref(false)
const isEditing = ref(false)
const formId = ref('')
const deleteDialogOpen = ref(false)
const pendingDeleteTemplate = ref(null)

const createEmptyFormData = (overrides = {}) => ({
  name: '',
  content: '',
  type: 'positive',
  category: '',
  ...overrides,
})

const formData = ref(createEmptyFormData())

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
  const categories = new Set()
  templates.value.forEach((template) => {
    if (template.category) categories.add(template.category)
  })
  return Array.from(categories).sort()
})

const filteredTemplates = computed(() => {
  let list = templates.value

  if (activeTypeFilter.value !== 'all') {
    list = list.filter((template) => template.type === activeTypeFilter.value)
  }

  if (activeCategoryFilter.value !== 'all') {
    list = list.filter((template) => template.category === activeCategoryFilter.value)
  }

  if (searchQuery.value.trim()) {
    const query = searchQuery.value.trim().toLowerCase()
    list = list.filter((template) => (
      template.name.toLowerCase().includes(query)
      || template.content.toLowerCase().includes(query)
    ))
  }

  return list
})

const resetForm = (overrides = {}) => {
  formData.value = createEmptyFormData(overrides)
  isEditing.value = false
  formId.value = ''
}

const openCreateForm = (overrides = {}) => {
  resetForm(overrides)
  showForm.value = true
}

const closeForm = () => {
  resetForm()
  showForm.value = false
}

const loadTemplates = async () => {
  try {
    const list = await App.GetPromptTemplates()
    templates.value = list || []
  } catch {
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
      await loadTemplates()
      closeForm()
      return
    }

    const nextType = formData.value.type || 'positive'
    const nextCategory = formData.value.category || ''

    await App.AddPromptTemplate(formData.value)
    toast.success('模板已添加')
    await loadTemplates()

    openCreateForm({
      type: nextType,
      category: nextCategory,
    })
  } catch {
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
  } catch {
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
  } catch {
    toast.error('复制失败')
  }
}

const truncate = (text, max = 80) => {
  if (!text) return ''
  return text.length > max ? `${text.slice(0, max)}...` : text
}

watch(() => props.open, async (newVal) => {
  if (!newVal) return

  await loadTemplates()
  closeForm()

  if (props.initialContent) {
    formData.value = createEmptyFormData({
      content: props.initialContent,
      type: props.initialType || 'positive',
    })
    showForm.value = true
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="w-[95vw] sm:max-w-[1100px] h-[74vh] max-h-[88vh] flex flex-col overflow-hidden p-6 select-none">
      <DialogHeader class="shrink-0 mb-4">
        <DialogTitle class="flex items-center gap-2">
          <Bookmark class="w-5 h-5" />
          <span>提示词模板库</span>
        </DialogTitle>
        <DialogDescription>
          保存、管理和复用你的常用提示词模板
        </DialogDescription>
      </DialogHeader>

      <div class="flex-1 min-h-0 flex gap-4 overflow-hidden isolate">
        <div class="relative z-0 min-w-0 min-h-0 flex-1 flex flex-col border rounded-md overflow-hidden bg-card">
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
                v-for="filter in typeFilters"
                :key="filter.key"
                :variant="activeTypeFilter === filter.key ? 'secondary' : 'ghost'"
                size="sm"
                class="h-6 px-2 text-[11px]"
                @click="activeTypeFilter = filter.key"
              >
                {{ filter.label }}
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
                  v-for="category in allCategories"
                  :key="category"
                  :variant="activeCategoryFilter === category ? 'secondary' : 'ghost'"
                  size="sm"
                  class="h-6 px-2 text-[11px]"
                  @click="activeCategoryFilter = category"
                >
                  {{ category }}
                </Button>
              </template>
            </div>
          </div>

          <ScrollArea class="min-h-0 flex-1">
            <div class="p-2 space-y-2">
              <div
                v-if="filteredTemplates.length === 0"
                class="flex flex-col items-center justify-center py-10 text-muted-foreground opacity-50"
              >
                <Bookmark class="w-10 h-10 mb-2" />
                <span class="text-sm">
                  {{ templates.length === 0 ? '暂无模板' : '没有匹配的模板' }}
                </span>
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
                    <Badge
                      v-if="template.category"
                      variant="outline"
                      class="shrink-0 text-[10px] px-1.5 py-0"
                    >
                      {{ template.category }}
                    </Badge>
                  </div>

                  <div class="text-xs text-muted-foreground leading-relaxed whitespace-pre-wrap break-words line-clamp-2">
                    {{ truncate(template.content, 120) }}
                  </div>
                </div>

                <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity shrink-0 ml-2">
                  <Button
                    size="icon"
                    variant="ghost"
                    class="h-7 w-7 text-primary hover:text-primary hover:bg-primary/10"
                    title="复制内容"
                    @click="copyContent(template)"
                  >
                    <Copy class="w-3 h-3" />
                  </Button>

                  <Button size="icon" variant="ghost" class="h-7 w-7" title="编辑" @click="editTemplate(template)">
                    <Edit2 class="w-3 h-3" />
                  </Button>

                  <Button
                    size="icon"
                    variant="ghost"
                    class="h-7 w-7 text-destructive hover:text-destructive hover:bg-destructive/10"
                    title="删除"
                    @click="requestDeleteTemplate(template)"
                  >
                    <Trash2 class="w-3 h-3" />
                  </Button>
                </div>
              </div>
            </div>
          </ScrollArea>

          <div class="shrink-0 p-2 border-t bg-muted/30">
            <Button
              size="sm"
              variant="outline"
              class="w-full justify-center gap-2 border-dashed bg-background/80 text-foreground hover:bg-accent"
              @click="openCreateForm()"
            >
              <Plus class="w-4 h-4" />
              添加模板
            </Button>
          </div>
        </div>

        <div
          v-if="showForm"
          class="relative z-10 ml-2 w-[480px] max-w-[50%] min-w-[400px] min-h-0 shrink-0 border-l bg-background pl-6 flex flex-col overflow-hidden animate-in slide-in-from-right-5 fade-in duration-300"
        >
          <div class="flex shrink-0 items-center justify-between mb-4">
            <h3 class="font-semibold text-sm">{{ isEditing ? '编辑模板' : '添加新模板' }}</h3>
            <Button size="icon" variant="ghost" class="h-6 w-6" @click="closeForm">
              <X class="w-4 h-4" />
            </Button>
          </div>

          <div class="flex-1 min-h-0 flex flex-col overflow-hidden pr-1">
            <div class="space-y-3 flex-1 min-h-0 overflow-y-auto scrollbar-hide pr-2">
              <div class="space-y-1">
                <Label class="text-xs">模板名称</Label>
                <Input v-model="formData.name" placeholder="例如：高质量人物" class="h-8 select-text" />
              </div>

              <div class="space-y-1 flex-1 min-h-0 flex flex-col">
                <Label class="text-xs">提示词内容</Label>
                <textarea
                  v-model="formData.content"
                  placeholder="输入提示词内容..."
                  class="w-full flex-1 rounded-md border border-input bg-background px-3 py-2 text-xs leading-6 placeholder:text-muted-foreground transition-[color,box-shadow,border-color] focus-visible:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-ring/60 select-text resize-none min-h-[200px]"
                />
              </div>

              <div class="space-y-1 shrink-0">
                <Label class="text-xs">类型</Label>
                <div class="flex gap-2">
                  <Button
                    v-for="filter in typeFilters.slice(1)"
                    :key="filter.key"
                    :variant="formData.type === filter.key ? 'secondary' : 'outline'"
                    size="sm"
                    class="h-7 text-xs"
                    @click="formData.type = filter.key"
                  >
                    {{ filter.label }}
                  </Button>
                </div>
              </div>

              <div class="space-y-1 shrink-0">
                <Label class="text-xs">分类（可选）</Label>
                <Input v-model="formData.category" placeholder="例如：人物、画风、质量词" class="h-8 select-text" />
              </div>
            </div>
          </div>

          <div class="mt-4 shrink-0 border-t pt-4 flex justify-end gap-2 bg-background">
            <Button variant="ghost" size="sm" @click="closeForm">取消</Button>
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
          确定要删除“{{ pendingDeleteTemplate?.name || '该模板' }}”吗？此操作不可撤销。
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="pendingDeleteTemplate = null">取消</AlertDialogCancel>
        <AlertDialogAction
          class="bg-destructive hover:bg-destructive/90"
          @click="confirmDeleteTemplate"
        >
          删除
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
