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
  FolderOpen,
  Loader2,
  MapPinned,
  RefreshCw,
  Save,
} from 'lucide-vue-next'

const props = defineProps({
  open: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'change'])

const binding = ref({
  rootDir: '',
  outputDir: '',
  outputRelPath: '',
  configured: false,
})
const selectedOutputDir = ref('')
const loading = ref(false)
const selecting = ref(false)
const saving = ref(false)
const opening = ref(false)

const hasSelectionChanged = computed(() => {
  const selected = selectedOutputDir.value.trim()
  if (!selected) return false
  if (!binding.value.configured) return true
  return selected !== binding.value.outputDir
})

const showEmptyState = computed(() => !binding.value.configured)

const normalizeError = (error, fallback) => {
  const message = String(error ?? '').trim()
  if (!message || message.includes('�')) {
    return fallback
  }
  return message
}

const handleDialogOpenChange = (nextOpen) => {
  if (!nextOpen && props.required && !binding.value.configured) {
    emit('update:open', true)
    return
  }
  emit('update:open', nextOpen)
}

const loadBinding = async () => {
  loading.value = true
  try {
    const result = await App.GetDirectoryBinding()
    binding.value = result || { rootDir: '', outputDir: '', outputRelPath: '', configured: false }
    selectedOutputDir.value = result?.configured ? (result.outputDir || '') : ''
  } catch (error) {
    toast.error(normalizeError(error, '读取当前输出目录失败'))
  } finally {
    loading.value = false
  }
}

watch(() => props.open, async (open) => {
  if (!open) return
  await loadBinding()
})

const selectOutputFolder = async () => {
  selecting.value = true
  try {
    const dir = await App.SelectFolder()
    if (!dir) return
    selectedOutputDir.value = dir
  } catch (error) {
    toast.error(normalizeError(error, '选择 output 文件夹失败'))
  } finally {
    selecting.value = false
  }
}

const openCurrentOutput = async () => {
  if (!binding.value.configured) return
  opening.value = true
  try {
    await App.OpenCurrentOutputDirectory()
  } catch (error) {
    toast.error(normalizeError(error, '打开当前 output 文件夹失败'))
  } finally {
    opening.value = false
  }
}

const saveBinding = async () => {
  if (!selectedOutputDir.value.trim()) {
    toast.error('请先选择 ComfyUI 的 output 文件夹')
    return
  }

  saving.value = true
  try {
    const result = await App.SaveOutputDirectory(selectedOutputDir.value.trim())
    binding.value = result || binding.value
    selectedOutputDir.value = result?.outputDir || selectedOutputDir.value.trim()
    toast.success('已绑定新的 output 文件夹')
    emit('change', result)
    emit('update:open', false)
  } catch (error) {
    toast.error(normalizeError(error, '保存 output 目录失败'))
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="handleDialogOpenChange">
    <DialogContent class="sm:max-w-[660px] overflow-hidden p-0">
      <div class="max-h-[85vh] overflow-y-auto p-6">
        <DialogHeader class="pr-8">
          <DialogTitle class="flex items-center gap-2 text-xl">
            <MapPinned class="h-5 w-5 text-primary" />
            绑定 ComfyUI 输出目录
          </DialogTitle>
          <DialogDescription>
            程序不会再默认猜测 exe 上一级目录。这里请直接选择真正的 ComfyUI `output` 文件夹，保存后会自动切换图片来源并刷新图库。
          </DialogDescription>
        </DialogHeader>

        <div class="mt-5 space-y-4">
          <div class="rounded-2xl border bg-muted/20 p-4">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0 flex-1 space-y-3">
                <div>
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                    当前状态
                  </div>
                  <div class="mt-1 text-sm text-foreground">
                    {{ showEmptyState ? '尚未绑定 output 目录' : '已绑定 output 目录' }}
                  </div>
                </div>

                <div>
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                    当前 output
                  </div>
                  <div class="mt-1 break-all text-sm text-foreground">
                    {{ binding.outputDir || '首次进入请手动选择 output 文件夹' }}
                  </div>
                </div>

                <div>
                  <div class="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
                    当前根目录
                  </div>
                  <div class="mt-1 break-all text-sm text-muted-foreground">
                    {{ binding.rootDir || '绑定后自动识别为 output 的上一级目录' }}
                  </div>
                </div>

                <div v-if="binding.configured && binding.outputRelPath" class="text-xs text-muted-foreground">
                  当前 output 相对路径：{{ binding.outputRelPath }}
                </div>
              </div>

              <Button
                variant="outline"
                class="shrink-0 gap-2"
                :disabled="loading || opening || !binding.configured"
                @click="openCurrentOutput"
              >
                <Loader2 v-if="opening" class="h-4 w-4 animate-spin" />
                <FolderOpen v-else class="h-4 w-4" />
                打开当前 output
              </Button>
            </div>
          </div>

          <div class="rounded-2xl border bg-background p-4 space-y-3">
            <div class="text-sm font-semibold">选择新的 output 文件夹</div>

            <div v-if="showEmptyState" class="rounded-xl border border-dashed bg-muted/20 px-4 py-3 text-sm leading-6 text-muted-foreground">
              这是首次进入，请直接选择 ComfyUI 的 output 文件夹本身，而不是程序 exe 所在目录。
            </div>

            <div class="flex gap-2">
              <Input
                v-model="selectedOutputDir"
                placeholder="请选择 ComfyUI 的 output 文件夹完整路径"
                class="flex-1"
              />
              <Button
                variant="outline"
                class="shrink-0 gap-2"
                :disabled="selecting"
                @click="selectOutputFolder"
              >
                <Loader2 v-if="selecting" class="h-4 w-4 animate-spin" />
                <FolderOpen v-else class="h-4 w-4" />
                选择文件夹
              </Button>
            </div>

            <div class="rounded-lg border border-dashed bg-muted/20 px-3 py-2 text-xs leading-5 text-muted-foreground">
              建议直接选中 ComfyUI 的 output 文件夹本身。保存后会立即切换图片来源，并重新加载默认目录和自定义目录内容。
            </div>
          </div>
        </div>

        <DialogFooter class="mt-6 gap-2">
          <Button variant="outline" class="gap-2" :disabled="loading" @click="loadBinding">
            <RefreshCw class="h-4 w-4" />
            重新读取
          </Button>
          <Button v-if="!required || binding.configured" variant="outline" @click="$emit('update:open', false)">
            取消
          </Button>
          <Button class="gap-2" :disabled="saving || !hasSelectionChanged" @click="saveBinding">
            <Loader2 v-if="saving" class="h-4 w-4 animate-spin" />
            <Save v-else class="h-4 w-4" />
            保存并切换
          </Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>
