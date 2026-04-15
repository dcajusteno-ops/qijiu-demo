<script setup>
import { ref, inject } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Plus, Trash2, Eraser, Pencil, Check, X, FolderTree } from 'lucide-vue-next'

const showToast = inject('showToast')
import * as App from '@/api'
import { useImages } from '@/composables/useImages'

const { fetchTags } = useImages()

const props = defineProps({
    tags: { type: Array, default: () => [] },
    collapsed: { type: Boolean, default: false }
})

const emit = defineEmits(['create-tag', 'delete-tag', 'batch-delete-tags', 'batch-update-tags', 'update-tag', 'refresh-images'])

const isOpen = ref(false)
const newTagName = ref('')
const newTagColor = ref('#3b82f6')
const newTagCategory = ref('')
const isEditing = ref(false)
const editingId = ref(null)
const isBatchMode = ref(false)
const selectedTags = ref(new Set())
const isBatchCategorizing = ref(false)
const batchCategoryName = ref('')

// Predefined color palette
const colorPalette = [
    '#3b82f6', // blue
    '#10b981', // green
    '#f59e0b', // amber
    '#ef4444', // red
    '#8b5cf6', // violet
    '#ec4899', // pink
    '#06b6d4', // cyan
    '#84cc16'  // lime
]

const resetForm = () => {
    isEditing.value = false
    editingId.value = null
    newTagName.value = ''
    newTagColor.value = '#3b82f6'
    newTagCategory.value = ''
    isBatchMode.value = false
    selectedTags.value.clear()
    isBatchCategorizing.value = false
    batchCategoryName.value = ''
}

const handleEdit = (tag) => {
    isEditing.value = true
    editingId.value = tag.id
    newTagName.value = tag.name
    newTagColor.value = tag.color
    newTagCategory.value = tag.category === 'default' || tag.category === '未分组' ? '' : tag.category
}

const handleCancelEdit = () => {
    resetForm()
}

const handleSubmit = () => {
    if (!newTagName.value.trim()) return
    
    if (isEditing.value && editingId.value) {
        emit('update-tag', editingId.value, {
            name: newTagName.value.trim(),
            color: newTagColor.value,
            category: newTagCategory.value.trim() || '未分组'
        })
        isEditing.value = false // Exit edit mode but keep dialog open
        editingId.value = null
        resetForm()
    } else {
        emit('create-tag', newTagName.value.trim(), newTagColor.value, newTagCategory.value.trim())
        resetForm()
    }
}

const handleDelete = (tagId) => {
    if (isEditing.value && editingId.value === tagId) {
        resetForm()
    }
    emit('delete-tag', tagId)
}

const cleanupLoading = ref(false)
const handleCleanup = async () => {
    cleanupLoading.value = true
    try {
        const count = await App.CleanupTags()
        if (count > 0) {
            showToast(`清理完成：移除了 ${count} 个无效标签关联`, 'success')
            emit('refresh-images') // Refresh to update counts
        } else {
            showToast('标签库很干净，没有发现无效关联', 'success')
        }
    } catch (e) {
        showToast('清理失败', 'error')
        console.error(e)
    } finally {
        cleanupLoading.value = false
    }
}
const toggleBatchMode = () => {
    isBatchMode.value = !isBatchMode.value
    selectedTags.value.clear()
    isEditing.value = false
    isBatchCategorizing.value = false
}

const toggleSelection = (tagId) => {
    if (selectedTags.value.has(tagId)) {
        selectedTags.value.delete(tagId)
    } else {
        selectedTags.value.add(tagId)
    }
}

const handleBatchDelete = async () => {
    if (selectedTags.value.size === 0) return
    
    // Use new batch delete API to prevent race conditions
    try {
        await App.BatchDeleteTags(Array.from(selectedTags.value))
        
        showToast(`成功删除 ${selectedTags.value.size} 个标签`, 'success')
        selectedTags.value.clear()
        isBatchMode.value = false
        // Refresh tags
        fetchTags() 
    } catch (e) {
        console.error(e)
        showToast('批量删除失败', 'error')
    }
}

const startBatchCategorize = () => {
    if (selectedTags.value.size === 0) return
    isBatchCategorizing.value = true
    batchCategoryName.value = ''
}

const cancelBatchCategorize = () => {
    isBatchCategorizing.value = false
    batchCategoryName.value = ''
}

const submitBatchCategorize = () => {
    if (selectedTags.value.size === 0) return
    emit('batch-update-tags', Array.from(selectedTags.value), { category: batchCategoryName.value || '未分组' })
    isBatchCategorizing.value = false
    batchCategoryName.value = ''
    selectedTags.value.clear()
    isBatchMode.value = false
}
</script>

<template>
    <Dialog v-model:open="isOpen">
        <DialogTrigger as-child>
            <slot>
                <Button variant="ghost" size="sm" :class="collapsed ? 'w-full justify-center px-0 h-10' : 'w-full justify-start gap-2'" :title="collapsed ? '管理标签' : ''">
                    <Plus class="h-4 w-4" :class="{ 'h-5 w-5': collapsed }" />
                    <span v-if="!collapsed">管理标签</span>
                </Button>
            </slot>
        </DialogTrigger>
        <DialogContent class="sm:max-w-[500px]">
            <DialogHeader>
                <DialogTitle>{{ isEditing ? '编辑标签' : (isBatchMode ? '批量管理' : '标签管理') }}</DialogTitle>
            </DialogHeader>
            
            <div class="space-y-4 py-4">
                <!-- Existing Tags -->
                <div>
                    <div class="flex items-center justify-between mb-2">
                        <Label class="text-sm font-medium">现有标签</Label>
                         <Button 
                            v-if="!isEditing"
                            variant="ghost" 
                            size="sm" 
                            @click="toggleBatchMode"
                            :class="isBatchMode ? 'text-primary bg-secondary' : 'text-muted-foreground'"
                            class="h-6 text-xs px-2"
                        >
                            {{ isBatchMode ? '退出批量' : '批量管理' }}
                        </Button>
                    </div>
                    <div class="flex flex-wrap gap-2 min-h-[60px] p-3 border rounded-md bg-muted/20">
                        <Badge 
                            v-for="tag in tags" 
                            :key="tag.id"
                            :style="{ backgroundColor: isBatchMode && !selectedTags.has(tag.id) ? tag.color + '80' : tag.color }"
                            class="pl-3 pr-1 py-1 text-white font-medium flex items-center gap-1 transition-all cursor-pointer select-none"
                            :class="{'ring-2 ring-primary ring-offset-2': isBatchMode && selectedTags.has(tag.id), 'opacity-60': isBatchMode && !selectedTags.has(tag.id) }"
                            @click="isBatchMode ? toggleSelection(tag.id) : null"
                        >
                            {{ tag.name }}
                            <div v-if="!isBatchMode" class="flex items-center ml-1">
                                <Button
                                    @click.stop="handleEdit(tag)"
                                    variant="ghost"
                                    size="icon"
                                    class="h-5 w-5 hover:bg-white/20 rounded-full"
                                    :class="{ 'ring-2 ring-white': editingId === tag.id }"
                                >
                                    <Pencil class="h-3 w-3" />
                                </Button>
                                <Button 
                                    @click.stop="handleDelete(tag.id)" 
                                    variant="ghost" 
                                    size="icon"
                                    class="h-5 w-5 hover:bg-white/20 rounded-full"
                                >
                                    <Trash2 class="h-3 w-3" />
                                </Button>
                            </div>
                           <div v-else class="flex items-center ml-1 h-5 w-5 justify-center">
                                <Check v-if="selectedTags.has(tag.id)" class="h-4 w-4" />
                           </div>
                        </Badge>
                        <div v-if="tags.length === 0" class="text-sm text-muted-foreground italic">
                            暂无标签
                        </div>
                    </div>
                </div>

                <!-- Create/Edit/Batch Logic Area -->
                <div class="space-y-3 pt-4 border-t">
                    <!-- Batch Categorize Form -->
                    <div v-if="isBatchCategorizing" class="space-y-3 animate-in fade-in slide-in-from-bottom-2">
                         <div class="flex items-center justify-between">
                            <Label class="font-medium">批量设置分类 (选中 {{ selectedTags.size }} 个)</Label>
                            <Button variant="ghost" size="sm" @click="cancelBatchCategorize" class="h-6 text-xs text-muted-foreground">
                                <X class="h-3 w-3 mr-1" />
                                取消
                            </Button>
                        </div>
                        <div class="flex gap-2">
                             <Input 
                                v-model="batchCategoryName" 
                                placeholder="输入新分类名称 (留空为未分组)"
                                @keyup.enter="submitBatchCategorize"
                            />
                            <Button @click="submitBatchCategorize">确定</Button>
                        </div>
                    </div>

                    <!-- Normal Create/Edit Form -->
                    <div v-else class="space-y-3">
                        <div class="flex items-center justify-between">
                            <Label class="font-medium">
                                {{ isEditing ? '编辑标签信息' : '创建新标签' }}
                            </Label>
                        <Button 
                            v-if="isEditing" 
                            variant="ghost" 
                            size="sm" 
                            @click="handleCancelEdit" 
                            class="h-6 text-xs text-muted-foreground"
                        >
                            <X class="h-3 w-3 mr-1" />
                            取消编辑
                        </Button>
                    </div>

                    <div>
                        <Label for="tag-name">标签名称</Label>
                        <Input 
                            id="tag-name" 
                            v-model="newTagName" 
                            placeholder="输入标签名..."
                            class="mt-1"
                            @keyup.enter="handleSubmit"
                        />
                    </div>
                    
                    <div>
                        <Label for="tag-category">分类</Label>
                        <Input 
                            id="tag-category" 
                            v-model="newTagCategory" 
                            placeholder="留空归入未分组"
                            class="mt-1"
                            @keyup.enter="handleSubmit"
                        />
                    </div>
                    
                    <div>
                        <Label class="mb-2 block">颜色</Label>
                        <div class="flex gap-2">
                            <button
                                v-for="color in colorPalette"
                                :key="color"
                                @click="newTagColor = color"
                                :style="{ backgroundColor: color }"
                                :class="[
                                    'h-8 w-8 rounded-full transition-all',
                                    newTagColor === color ? 'ring-2 ring-offset-2 ring-primary scale-110' : 'hover:scale-105'
                                ]"
                            />
                        </div>
                    </div>
                </div>
            </div>

            </div>


            <DialogFooter class="flex justify-between items-center sm:justify-between">
                <div class="flex gap-2">
                    <Button 
                        v-if="!isBatchMode"
                        variant="outline" 
                        size="sm" 
                        @click="handleCleanup" 
                        :disabled="cleanupLoading"
                        class="text-muted-foreground hover:text-foreground"
                    >
                        <Eraser class="h-4 w-4 mr-2" :class="{ 'animate-pulse': cleanupLoading }" />
                        {{ cleanupLoading ? '清理中...' : '清理无效关联' }}
                    </Button>
                     <div v-else class="flex gap-2">
                        <Button 
                            variant="primary"
                            size="sm"
                            class="bg-blue-600 hover:bg-blue-700 text-white"
                            @click="startBatchCategorize" 
                            :disabled="selectedTags.size === 0 || isBatchCategorizing"
                        >
                            <FolderTree class="h-4 w-4 mr-2" />
                            批量分组
                        </Button>
                        <Button 
                            variant="destructive" 
                            size="sm" 
                            @click="handleBatchDelete" 
                            :disabled="selectedTags.size === 0 || isBatchCategorizing"
                        >
                            <Trash2 class="h-4 w-4 mr-2" />
                            删除选中 ({{ selectedTags.size }})
                        </Button>
                     </div>
                </div>

                <Button v-if="!isBatchMode" @click="handleSubmit" :disabled="!newTagName.trim()">
                    <component :is="isEditing ? Check : Plus" class="h-4 w-4 mr-2" />
                    {{ isEditing ? '更新标签' : '创建标签' }}
                </Button>
            </DialogFooter>
        </DialogContent>
    </Dialog>
</template>

