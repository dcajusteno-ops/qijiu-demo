<template>
    <Dialog :open="open" @update:open="emit('update:open', $event)">
        <DialogContent class="sm:max-w-[425px] flex flex-col gap-0 p-0 overflow-hidden border-0 shadow-2xl">
            <!-- Header Section with Colorful Background -->
            <div class="bg-gradient-to-br from-primary/10 via-primary/5 to-background p-6 flex flex-col items-center justify-center text-center border-b border-border/50">
                <div class="w-16 h-16 rounded-full bg-background shadow-sm flex items-center justify-center mb-4 ring-4 ring-background/50">
                    <FolderInput class="w-8 h-8 text-primary" />
                </div>
                <DialogTitle class="text-xl font-bold tracking-tight">移动图片</DialogTitle>
                <DialogDescription class="mt-2 text-muted-foreground max-w-[280px]">
                    将选中的 <span class="font-semibold text-foreground">{{ count }}</span> 张图片移动到新的位置
                </DialogDescription>
            </div>

            <div class="p-6 space-y-6">
                <div class="space-y-2">
                    <Label class="text-sm font-medium text-foreground/80">目标文件夹路径</Label>
                    <div class="relative group">
                        <Input 
                            v-model="targetFolder" 
                            @input="handleInput"
                            placeholder="例如: 2026-01/精选集"
                            class="pl-10 h-11 transition-all border-input/60 focus:border-primary focus:ring-2 focus:ring-primary/20 bg-muted/30 hover:bg-muted/50 focus:bg-background"
                            :class="{'border-destructive focus:border-destructive focus:ring-destructive/20': error}"
                            @keyup.enter="handleMove"
                        />
                        <FolderOpen class="absolute left-3 top-3 w-5 h-5 text-muted-foreground/60 transition-colors group-hover:text-muted-foreground" />
                    </div>
                    <p v-if="error" class="text-xs text-destructive font-medium animate-pulse">{{ error }}</p>
                    <div class="flex items-start gap-2 mt-2 text-xs text-muted-foreground bg-muted/30 p-2.5 rounded-md border border-border/40">
                        <Info class="w-4 h-4 text-blue-500 shrink-0 mt-0.5" />
                        <p>支持相对路径（如 <span class="font-mono text-primary/80">new_folder</span>）或绝对路径（如 <span class="font-mono text-primary/80">H:\Backup</span>）。<br/>不支持字符: <span class="text-destructive font-mono">&lt; &gt; " | ? *</span></p>
                    </div>
                </div>
            </div>

            <DialogFooter class="p-6 pt-2 bg-muted/10 flex gap-3">
                <Button variant="outline" class="flex-1 h-10 hover:bg-muted/80" @click="emit('update:open', false)">取消</Button>
                <Button 
                    class="flex-1 h-10 shadow-lg shadow-primary/20 hover:shadow-primary/30 transition-all font-medium" 
                    @click="handleMove" 
                    :disabled="!targetFolder.trim() || !!error"
                >
                    <Check class="w-4 h-4 mr-2" />
                    确认移动
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { FolderInput, FolderOpen, Info, Check } from 'lucide-vue-next'

const props = defineProps({
    open: Boolean,
    count: Number
})

const emit = defineEmits(['update:open', 'move'])

const targetFolder = ref('')
const error = ref('')

// Reset input when dialog closes
watch(() => props.open, (newVal) => {
    if (!newVal) {
        targetFolder.value = ''
        error.value = ''
    }
})

const handleInput = (e) => {
    const val = e.target.value
    // Windows invalid chars: < > : " | ? * (Allow : for drive letters)
    if (/[<>"|?*]/.test(val)) {
        error.value = '检测到非法字符，已自动移除'
        // Temporarily show error then clear it
        targetFolder.value = val.replace(/[<>"|?*]/g, '')
        
        setTimeout(() => {
            error.value = ''
        }, 2000)
    } else {
        error.value = ''
    }
}

const handleMove = () => {
    if (!targetFolder.value.trim() || error.value) return
    emit('move', targetFolder.value.trim())
}
</script>
