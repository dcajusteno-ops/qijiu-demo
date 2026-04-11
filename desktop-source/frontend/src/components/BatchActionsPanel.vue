<template>
    <Transition
        enter-active-class="transition-all duration-200"
        leave-active-class="transition-all duration-200"
        enter-from-class="translate-y-full opacity-0"
        leave-to-class="translate-y-full opacity-0"
    >
        <div 
            v-if="show"
            class="fixed bottom-8 left-1/2 -translate-x-1/2 z-[999] bg-card/90 backdrop-blur-xl border-2 border-primary/30 shadow-2xl rounded-2xl p-4 flex items-center gap-4 select-none animate-in fade-in slide-in-from-bottom-4 duration-300"
        >
            <!-- Selection Controls -->
            <Button variant="ghost" size="sm" @click="emit('select-all')">全选本页</Button>
            <Button variant="ghost" size="sm" @click="emit('clear-selection')" :disabled="count === 0">清空</Button>
            
            <Separator orientation="vertical" class="h-6" />
            
            <!-- Batch Add Tag -->
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" :disabled="count === 0">
                        <Tags class="w-4 h-4 mr-2" />
                        添加标签
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent side="top" :side-offset="12" class="max-h-64 overflow-y-auto">
                    <DropdownMenuLabel>选择标签</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem 
                        v-for="tag in tags" 
                        :key="tag.id"
                        @click="emit('batch-add-tag', tag.id)"
                        @select="emit('batch-add-tag', tag.id)"
                    >
                        <div 
                            class="w-3 h-3 rounded-full mr-2" 
                            :style="{ backgroundColor: tag.color }"
                        ></div>
                        {{ tag.name }}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator v-if="tags.length === 0" />
                    <DropdownMenuItem v-if="tags.length === 0" disabled>
                        暂无标签
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <!-- Batch Remove Tag -->
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="outline" size="sm" :disabled="count === 0">
                        <Tags class="w-4 h-4 mr-2" />
                        移除标签
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent side="top" :side-offset="12" class="max-h-64 overflow-y-auto">
                    <DropdownMenuLabel>选择标签</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem 
                        v-for="tag in tags" 
                        :key="tag.id"
                        @click="emit('batch-remove-tag', tag.id)"
                        @select="emit('batch-remove-tag', tag.id)"
                    >
                        <div 
                            class="w-3 h-3 rounded-full mr-2" 
                            :style="{ backgroundColor: tag.color }"
                        ></div>
                        {{ tag.name }}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator v-if="tags.length === 0" />
                    <DropdownMenuItem v-if="tags.length === 0" disabled>
                        暂无标签
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            <!-- Batch Move -->
            <Button variant="outline" size="sm" @click="emit('batch-move')" :disabled="count === 0">
                <FolderSymlink class="w-4 h-4 mr-2" />
                移动到文件夹
            </Button>

            <!-- Batch Favorite -->
            <Button variant="outline" size="sm" @click="emit('batch-favorite')" :disabled="count === 0">
                <Heart class="w-4 h-4 mr-2" />
                批量收藏
            </Button>

            <!-- A/B Compare -->
            <Button v-if="count === 2" variant="outline" size="sm" @click="emit('compare')" class="bg-blue-500/10 text-blue-500 hover:bg-blue-500/20 hover:text-blue-600 border-blue-500/30">
                <ArrowLeftRight class="w-4 h-4 mr-2" />
                A/B 对比
            </Button>

            <Separator orientation="vertical" class="h-6" />

            <!-- Batch Delete -->
            <Button variant="destructive" size="sm" @click="emit('batch-delete')" :disabled="count === 0">
                删除选中 ({{ count }})
            </Button>
        </div>
    </Transition>
</template>

<script setup>
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tags, FolderSymlink, Heart, ArrowLeftRight } from 'lucide-vue-next'

defineProps({
    show: Boolean,
    count: Number,
    tags: Array
})

const emit = defineEmits([
    'batch-add-tag',
    'batch-remove-tag',
    'batch-move',
    'batch-favorite',
    'batch-delete',
    'select-all',
    'clear-selection',
    'compare'
])
</script>
