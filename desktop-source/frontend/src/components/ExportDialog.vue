<script setup>
import { ref, inject } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'

defineProps({
  open: Boolean,
  count: Number,
})

const emit = defineEmits(['update:open', 'confirm'])

const confirm = inject('confirm')
const targetDir = ref('')
const moveFiles = ref(false)
const loading = ref(false)

const handleConfirm = async () => {
  if (!targetDir.value.trim()) return
  if (moveFiles.value) {
    const ok = await confirm('移动模式会从原目录删除文件，此操作不可撤销。确定继续吗？')
    if (!ok) return
  }
  loading.value = true
  try {
    await emit('confirm', {
      targetDir: targetDir.value,
      move: moveFiles.value,
    })
    emit('update:open', false)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="$emit('update:open', $event)">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <DialogTitle>导出图片</DialogTitle>
        <DialogDescription>
          将选中的 {{ count }} 张图片导出到指定文件夹。
        </DialogDescription>
      </DialogHeader>

      <div class="grid gap-4 py-4">
        <div class="grid grid-cols-4 items-center gap-4">
          <Label for="target-dir" class="text-right">
            目标路径
          </Label>
          <Input
            id="target-dir"
            v-model="targetDir"
            placeholder="例如：D:/图片/导出"
            class="col-span-3"
          />
        </div>
        <div class="flex items-center space-x-2 ml-[25%]">
          <Checkbox id="move-mode" v-model:checked="moveFiles" />
          <Label for="move-mode" class="cursor-pointer">移动文件而不是复制</Label>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" @click="$emit('update:open', false)">取消</Button>
        <Button @click="handleConfirm" :disabled="loading || !targetDir">
          {{ loading ? '导出中...' : '开始导出' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
