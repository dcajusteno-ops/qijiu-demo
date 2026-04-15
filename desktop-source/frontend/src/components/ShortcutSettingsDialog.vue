<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Switch } from '@/components/ui/switch'
import {
  Keyboard,
  RotateCcw,
  Save,
  X,
  Eraser,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import * as App from '@/api'
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
})

const emit = defineEmits(['update:open'])

const loading = ref(false)
const saving = ref(false)
const enabled = ref(true)
const bindings = ref(buildBindingsFromCatalog())
const capturingAction = ref('')

const duplicateAccelerators = computed(() => findDuplicateAccelerators(bindings.value))

const rows = computed(() => shortcutActionCatalog.map((action) => {
  const binding = bindings.value.find((item) => item.action === action.id) || {
    action: action.id,
    accelerator: normalizeAccelerator(action.defaultAccelerator),
  }
  const normalized = normalizeAccelerator(binding.accelerator)
  return {
    ...action,
    accelerator: normalized,
    isDuplicate: normalized ? duplicateAccelerators.value.has(normalized) : false,
    isCapturing: capturingAction.value === action.id,
  }
}))

const hasDuplicateAccelerators = computed(() => duplicateAccelerators.value.size > 0)

const setBinding = (actionId, accelerator) => {
  bindings.value = bindings.value.map((binding) => (
    binding.action === actionId
      ? { ...binding, accelerator: normalizeAccelerator(accelerator) }
      : binding
  ))
}

const resetAllBindings = () => {
  enabled.value = true
  bindings.value = buildBindingsFromCatalog()
  capturingAction.value = ''
}

const resetSingleBinding = (actionId) => {
  const action = shortcutActionMap[actionId]
  if (!action) return
  setBinding(actionId, action.defaultAccelerator)
}

const clearBinding = (actionId) => {
  setBinding(actionId, '')
}

const startCapture = (actionId) => {
  capturingAction.value = actionId
}

const cancelCapture = () => {
  capturingAction.value = ''
}

const loadSettings = async () => {
  loading.value = true
  try {
    const settings = await App.GetShortcutSettings()
    enabled.value = settings?.enabled !== false
    bindings.value = buildBindingsFromCatalog(settings?.bindings || [])
  } catch {
    bindings.value = buildBindingsFromCatalog()
    enabled.value = true
    toast.error('加载快捷键设置失败')
  } finally {
    loading.value = false
  }
}

const handleCaptureKeydown = (event) => {
  if (!capturingAction.value) return

  if (
    event.key === 'Escape'
    && !event.ctrlKey
    && !event.altKey
    && !event.shiftKey
    && !event.metaKey
  ) {
    event.preventDefault()
    cancelCapture()
    return
  }

  const accelerator = getAcceleratorFromEvent(event)
  if (!accelerator) return

  event.preventDefault()
  event.stopPropagation()
  setBinding(capturingAction.value, accelerator)
  cancelCapture()
}

const saveSettings = async () => {
  if (hasDuplicateAccelerators.value) {
    toast.error('存在重复的快捷键，请先调整后再保存')
    return
  }

  saving.value = true
  try {
    const saved = await App.SaveShortcutSettings({
      enabled: enabled.value,
      bindings: bindings.value.map((binding) => ({
        action: binding.action,
        accelerator: binding.accelerator,
      })),
    })
    enabled.value = saved?.enabled !== false
    bindings.value = buildBindingsFromCatalog(saved?.bindings || [])
    toast.success('快捷键设置已保存')
    emit('update:open', false)
  } catch (error) {
    toast.error(String(error || '保存失败'))
  } finally {
    saving.value = false
  }
}

watch(capturingAction, (next, prev) => {
  if (next && !prev) {
    window.addEventListener('keydown', handleCaptureKeydown, true)
    return
  }
  if (!next && prev) {
    window.removeEventListener('keydown', handleCaptureKeydown, true)
  }
})

watch(() => props.open, async (nextOpen) => {
  if (!nextOpen) {
    cancelCapture()
    return
  }
  await loadSettings()
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleCaptureKeydown, true)
})
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[860px] h-[78vh] max-h-[88vh] flex flex-col overflow-hidden p-6">
      <DialogHeader class="shrink-0 mb-3">
        <DialogTitle class="flex items-center gap-2">
          <Keyboard class="h-5 w-5" />
          <span>全局快捷键设置</span>
        </DialogTitle>
        <DialogDescription>
          为常用视图和操作配置系统级快捷键。应用运行期间，即使窗口不在前台也会生效。
        </DialogDescription>
      </DialogHeader>

      <div class="shrink-0 rounded-lg border bg-muted/20 p-4">
        <div class="flex items-center justify-between gap-4">
          <div class="space-y-1">
            <div class="text-sm font-medium">启用全局快捷键</div>
            <div class="text-xs text-muted-foreground">
              如果某个按键被系统或其他程序占用，保存时会直接提示。
            </div>
          </div>
          <Switch :checked="enabled" @update:checked="enabled = $event" />
        </div>
      </div>

      <div
        v-if="hasDuplicateAccelerators"
        class="shrink-0 mt-3 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-xs text-destructive"
      >
        检测到重复快捷键，请确保每个动作使用唯一组合键。
      </div>

      <div class="mt-4 flex-1 min-h-0 overflow-hidden rounded-lg border bg-card">
        <ScrollArea class="h-full">
          <div class="divide-y">
            <div
              v-for="row in rows"
              :key="row.id"
              class="flex items-center justify-between gap-4 px-4 py-4"
            >
              <div class="min-w-0 flex-1 space-y-1">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium">{{ row.label }}</span>
                  <Badge v-if="row.isDuplicate" variant="destructive" class="text-[10px]">
                    重复
                  </Badge>
                </div>
                <div class="text-xs text-muted-foreground">
                  {{ row.description }}
                </div>
                <div class="text-[11px] text-muted-foreground/80">
                  默认：{{ row.defaultAccelerator }}
                </div>
              </div>

              <div class="flex items-center gap-2 shrink-0">
                <div
                  class="min-w-[140px] rounded-md border bg-background px-3 py-2 text-center text-sm font-medium"
                  :class="row.isDuplicate ? 'border-destructive text-destructive' : 'border-border'"
                >
                  {{ row.isCapturing ? '请按下组合键...' : formatShortcutLabel(row.accelerator) }}
                </div>

                <Button
                  size="sm"
                  :variant="row.isCapturing ? 'default' : 'outline'"
                  @click="row.isCapturing ? cancelCapture() : startCapture(row.id)"
                >
                  {{ row.isCapturing ? '取消录制' : '录制快捷键' }}
                </Button>
                <Button size="icon" variant="ghost" title="恢复默认" @click="resetSingleBinding(row.id)">
                  <RotateCcw class="h-4 w-4" />
                </Button>
                <Button size="icon" variant="ghost" title="清空快捷键" @click="clearBinding(row.id)">
                  <Eraser class="h-4 w-4" />
                </Button>
              </div>
            </div>

            <div v-if="loading" class="px-4 py-10 text-center text-sm text-muted-foreground">
              正在加载快捷键设置...
            </div>
          </div>
        </ScrollArea>
      </div>

      <DialogFooter class="shrink-0 mt-4 flex items-center justify-between sm:justify-between border-t pt-4">
        <Button variant="ghost" size="sm" @click="resetAllBindings">
          <RotateCcw class="mr-2 h-4 w-4" />
          恢复全部默认
        </Button>

        <div class="flex items-center gap-2">
          <Button variant="ghost" size="sm" @click="$emit('update:open', false)">
            <X class="mr-2 h-4 w-4" />
            关闭
          </Button>
          <Button size="sm" :disabled="loading || saving || hasDuplicateAccelerators" @click="saveSettings">
            <Save class="mr-2 h-4 w-4" />
            {{ saving ? '保存中...' : '保存设置' }}
          </Button>
        </div>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
