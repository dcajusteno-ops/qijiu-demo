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
import { ScrollArea } from '@/components/ui/scroll-area'
import { 
  Plus, 
  Trash2, 
  Edit2, 
  Play, 
  TerminalSquare, 
  X, 
  Save 
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import { availableIcons, categorizedIcons, iconCount } from '@/lib/icons'

const props = defineProps({
  open: { type: Boolean, default: false }
})

const emit = defineEmits(['update:open'])
const defaultBuiltinIcon = 'Terminal'

const tools = ref([])
const loading = ref(false)
const showForm = ref(false)
const isEditing = ref(false)
const formId = ref('') // ID of tool being edited
const deleteDialogOpen = ref(false)
const pendingDeleteTool = ref(null)

const formData = ref({
  name: '',
  path: '',
  args: '',
  icon: ''
})

const isBuiltinIcon = (icon) => Boolean(icon && availableIcons[icon])
const isExecutablePath = computed(() => formData.value.path?.trim().toLowerCase().endsWith('.exe'))
const currentFormIconComponent = computed(() => {
    if (isBuiltinIcon(formData.value.icon)) {
        return availableIcons[formData.value.icon]
    }
    return availableIcons[defaultBuiltinIcon] || TerminalSquare
})
const hasImageIcon = computed(() => Boolean(formData.value.icon && !isBuiltinIcon(formData.value.icon)))

const resetForm = () => {
  formData.value = { name: '', path: '', args: '', icon: '' }
  isEditing.value = false
  formId.value = ''
  showForm.value = false
}

const loadTools = async () => {
    loading.value = true
    try {
        const list = await App.GetLauncherTools()
        tools.value = list || []
    } catch (e) {
        toast.error('请求出错')
    } finally {
        loading.value = false
    }
}

const saveTool = async () => {
    if (!formData.value.name || !formData.value.path) {
        toast.error('名称和路径不能为空')
        return
    }

    try {
        if (isEditing.value) {
            await App.UpdateLauncherTool(formId.value, formData.value)
            toast.success('工具已更新')
        } else {
            await App.AddLauncherTool(formData.value)
            toast.success('工具已添加')
        }
        await loadTools()
        resetForm()
    } catch (e) {
        toast.error('保存失败')
    }
}

const editTool = (tool) => {
    formData.value = {
        name: tool.name,
        path: tool.path,
        args: tool.args || '',
        icon: tool.icon || ''
    }
    formId.value = tool.id
    isEditing.value = true
    showForm.value = true
}

const requestDeleteTool = (tool) => {
    pendingDeleteTool.value = tool
    deleteDialogOpen.value = true
}

const confirmDeleteTool = async () => {
    const tool = pendingDeleteTool.value
    if (!tool?.id) return

    try {
        await App.DeleteLauncherTool(tool.id)
        toast.success('工具已删除')
        await loadTools()
    } catch (e) {
        toast.error('删除失败')
    } finally {
        deleteDialogOpen.value = false
        pendingDeleteTool.value = null
    }
}

const runTool = async (id) => {
    try {
        await App.RunLauncherTool(id)
        toast.success('已启动')
        emit('update:open', false) // Optionally close window
    } catch (e) {
        toast.error('启动失败')
    }
}

watch(() => props.open, (newVal) => {
    if (newVal) {
        loadTools()
        resetForm()
    }
})

const extractIcon = async (force = false) => {
    const path = formData.value.path
    if (!path || !path.toLowerCase().endsWith('.exe')) return
    if (!force && isBuiltinIcon(formData.value.icon)) return
    // Only fetch if icon is empty to avoid overwriting user custom icon (if we supported that fully)
    // But here we might want to update it if path changed. Let's just do it if icon is empty or looks like auto-generated
    
    try {
        const data = await App.ExtractIcon(path)
        if (data) {
            formData.value.icon = data
        }
    } catch (e) {
        console.error(e)
    }
}

const handlePathBlur = () => {
    void extractIcon(false)
}

const useBuiltInIcon = (iconName) => {
    formData.value.icon = iconName
}
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[920px] h-[74vh] max-h-[88vh] flex flex-col overflow-hidden p-6 select-none">
      <DialogHeader class="shrink-0 mb-4">
        <DialogTitle class="flex items-center gap-2">
            <TerminalSquare class="w-5 h-5" />
            <span>外部工具启动器</span>
        </DialogTitle>
        <DialogDescription>
          配置并启动本地程序、脚本或网页文件（如 .exe、.bat、.cmd、.html）
        </DialogDescription>
      </DialogHeader>

      <div class="flex-1 min-h-0 flex gap-4 overflow-hidden">
          <!-- Left: Tool List -->
          <div class="min-w-0 flex-1 flex flex-col border rounded-md overflow-hidden bg-card">
              <div class="p-2 border-b bg-muted/30 flex justify-between items-center">
                  <span class="text-xs font-semibold px-2">已配置工具</span>
                  <Button size="sm" variant="ghost" class="h-8 w-8 p-0" @click="resetForm(); showForm = true" title="添加工具">
                      <Plus class="w-4 h-4" />
                  </Button>
              </div>
              
              <ScrollArea class="flex-1">
                  <div class="p-2 space-y-2">
                      <div v-if="tools.length === 0" class="flex flex-col items-center justify-center py-10 text-muted-foreground opacity-50">
                          <TerminalSquare class="w-10 h-10 mb-2" />
                          <span class="text-sm">暂无工具</span>
                      </div>
                      
                      <div 
                        v-for="tool in tools" 
                        :key="tool.id" 
                        class="flex items-center justify-between p-3 rounded-md border bg-background hover:bg-accent/50 group transition-all cursor-pointer"
                        @dblclick="runTool(tool.id)"
                        title="双击启动"
                      >
                          <div class="flex items-center gap-3 min-w-0">
                              <div class="w-8 h-8 rounded bg-primary/10 flex items-center justify-center shrink-0">
                                  <component v-if="tool.icon && isBuiltinIcon(tool.icon)" :is="availableIcons[tool.icon]" class="w-4 h-4 text-primary" />
                                  <img v-else-if="tool.icon" :src="tool.icon" class="w-5 h-5 object-contain" />
                                  <TerminalSquare v-else class="w-4 h-4 text-primary" />
                              </div>
                              <div class="min-w-0">
                                  <div class="font-medium text-sm truncate">{{ tool.name }}</div>
                                  <div class="text-xs text-muted-foreground truncate opacity-70" :title="tool.path">{{ tool.path }}</div>
                              </div>
                          </div>
                          
                          <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                              <Button size="icon" variant="ghost" class="h-8 w-8 text-primary hover:text-primary hover:bg-primary/10" @click.stop="runTool(tool.id)" title="启动">
                                  <Play class="w-4 h-4" />
                              </Button>
                              <Button size="icon" variant="ghost" class="h-8 w-8 text-muted-foreground" @click.stop="editTool(tool)" title="编辑">
                                  <Edit2 class="w-3 h-3" />
                              </Button>
                              <Button size="icon" variant="ghost" class="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10" @click.stop="requestDeleteTool(tool)" title="删除">
                                  <Trash2 class="w-3 h-3" />
                              </Button>
                          </div>
                      </div>
                  </div>
              </ScrollArea>
          </div>

          <!-- Right: Form (Conditional) -->
          <div v-if="showForm" class="w-[380px] min-h-0 border-l pl-4 flex flex-col overflow-hidden animate-in slide-in-from-right-5 fade-in duration-300">
              <div class="flex shrink-0 items-center justify-between mb-4">
                  <h3 class="font-semibold text-sm">{{ isEditing ? '编辑工具' : '添加新工具' }}</h3>
                  <Button size="icon" variant="ghost" class="h-6 w-6" @click="showForm = false">
                      <X class="w-4 h-4" />
                  </Button>
              </div>
              
              <div class="space-y-3 flex-1 min-h-0 overflow-y-auto pr-1">
                  <div class="space-y-1">
                      <Label class="text-xs">名称</Label>
                      <Input v-model="formData.name" placeholder="例如: Checkpoint Merger" class="h-8 select-text" />
                  </div>
                  
                  <div class="space-y-1">
                      <Label class="text-xs">目标路径</Label>
                      <Input 
                        v-model="formData.path" 
                        placeholder="D:\Tools\Tool.exe 或 D:\Tools\脚本.bat" 
                        class="h-8 font-mono text-xs select-text" 
                        @blur="handlePathBlur"
                      />
                      <p class="text-[10px] text-muted-foreground">支持 .exe、.bat、.cmd、.html、.htm 等本地文件</p>
                  </div>
                  
                  <div class="space-y-1">
                      <Label class="text-xs">启动参数 (可选)</Label>
                      <Input v-model="formData.args" placeholder="--gpu-id 0" class="h-8 font-mono text-xs select-text" />
                  </div>
                  
                  <div class="space-y-2">
                      <div class="flex items-center justify-between gap-2">
                          <Label class="text-xs">图标</Label>
                          <div class="flex items-center gap-1">
                              <Button
                                v-if="isExecutablePath"
                                type="button"
                                variant="outline"
                                size="sm"
                                class="h-7 px-2 text-[11px]"
                                @click="extractIcon(true)"
                              >
                                提取程序图标
                              </Button>
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                class="h-7 px-2 text-[11px]"
                                @click="formData.icon = ''"
                              >
                                使用默认
                              </Button>
                          </div>
                      </div>

                      <div class="flex items-center gap-3 rounded-md border bg-muted/20 p-3">
                          <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-md border bg-background">
                              <component v-if="!hasImageIcon" :is="currentFormIconComponent" class="h-5 w-5 text-primary" />
                              <img v-else :src="formData.icon" class="h-6 w-6 object-contain" />
                          </div>
                          <div class="min-w-0">
                              <div class="text-xs font-medium">
                                {{ hasImageIcon ? '当前使用程序图标' : (formData.icon || '当前使用默认图标') }}
                              </div>
                              <div class="text-[10px] text-muted-foreground">
                                也可以在下方直接选择项目内置图标
                              </div>
                          </div>
                      </div>

                      <div class="space-y-1">
                          <div class="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                            内置图标（{{ iconCount }}）
                          </div>
                          <div class="max-h-[150px] space-y-3 overflow-y-auto rounded-md border bg-background/70 p-3">
                            <div v-for="(icons, category) in categorizedIcons" :key="`launcher-${category}`" class="space-y-2">
                              <div class="border-l-2 border-primary/30 pl-1 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/60">
                                {{ category }}
                              </div>
                              <div class="grid grid-cols-7 gap-1.5">
                                <button
                                  v-for="iconName in icons"
                                  :key="`launcher-${iconName}`"
                                  type="button"
                                  class="flex items-center justify-center rounded-md p-2 transition-all hover:scale-105"
                                  :class="formData.icon === iconName ? 'bg-primary text-primary-foreground shadow-sm' : 'text-muted-foreground hover:bg-muted'"
                                  :title="iconName"
                                  @click="useBuiltInIcon(iconName)"
                                >
                                  <component :is="availableIcons[iconName]" class="h-4 w-4" />
                                </button>
                              </div>
                            </div>
                          </div>
                      </div>
                  </div>
              </div>
              
              <div class="mt-4 shrink-0 border-t pt-4 flex justify-end gap-2 bg-background">
                  <Button variant="ghost" size="sm" @click="showForm = false">取消</Button>
                  <Button size="sm" @click="saveTool">
                      <Save class="w-3 h-3 mr-2" />
                      保存
                  </Button>
              </div>
          </div>
      </div>
    </DialogContent>
  </Dialog>

  <AlertDialog v-model:open="deleteDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>确认删除工具</AlertDialogTitle>
        <AlertDialogDescription>
          确定要删除“{{ pendingDeleteTool?.name || '该工具' }}”吗？这只会移除启动器里的配置，不会删除原始文件。
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel @click="pendingDeleteTool = null">取消</AlertDialogCancel>
        <AlertDialogAction @click="confirmDeleteTool" class="bg-destructive hover:bg-destructive/90">
          删除
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>
