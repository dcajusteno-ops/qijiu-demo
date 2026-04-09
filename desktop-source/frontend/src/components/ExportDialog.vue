<script setup>
import { ref } from 'vue'
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

const props = defineProps({
    open: Boolean,
    count: Number
})

const emit = defineEmits(['update:open', 'confirm'])

const targetDir = ref('')
const moveFiles = ref(false)
const loading = ref(false)

const handleConfirm = async () => {
    if (!targetDir.value.trim()) return
    loading.value = true
    try {
        await emit('confirm', { 
            targetDir: targetDir.value, 
            move: moveFiles.value 
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
            placeholder="C:/Images/Export"
            class="col-span-3"
          />
        </div>
        <!-- 
        <div class="flex items-center space-x-2 ml-[25%] hidden">
           <Checkbox id="move-mode" v-model="moveFiles" />
           <Label for="move-mode" class="cursor-pointer">移动文件 (而不是复制)</Label>
        </div>
        -->
         <p class="text-xs text-muted-foreground ml-[25%]">
            目前仅支持复制模式，不会删除原文件。
        </p>
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
