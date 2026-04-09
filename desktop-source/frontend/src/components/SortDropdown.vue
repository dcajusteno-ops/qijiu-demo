<template>
    <DropdownMenu>
        <DropdownMenuTrigger asChild>
            <Button variant="outline" size="sm" class="gap-2">
                <ArrowUpDown class="w-4 h-4" />
                {{ sortLabels[sortBy] }}
                <component :is="sortOrder === 'asc' ? ArrowUp : ArrowDown" class="w-3 h-3 text-muted-foreground" />
            </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-48">
            <DropdownMenuLabel>排序方式</DropdownMenuLabel>
            <DropdownMenuSeparator />
            
            <DropdownMenuItem @click="setSortBy('time')">
                <Clock class="w-4 h-4 mr-2" />
                时间
                <Check v-if="sortBy === 'time'" class="w-4 h-4 ml-auto" />
            </DropdownMenuItem>
            
            <DropdownMenuItem @click="setSortBy('size')">
                <HardDrive class="w-4 h-4 mr-2" />
                大小
                <Check v-if="sortBy === 'size'" class="w-4 h-4 ml-auto" />
            </DropdownMenuItem>
            
            <DropdownMenuItem @click="setSortBy('name')">
                <FileText class="w-4 h-4 mr-2" />
                名称
                <Check v-if="sortBy === 'name'" class="w-4 h-4 ml-auto" />
            </DropdownMenuItem>
            
            <DropdownMenuItem @click="setSortBy('dimensions')">
                <Maximize class="w-4 h-4 mr-2" />
                尺寸
                <Check v-if="sortBy === 'dimensions'" class="w-4 h-4 ml-auto" />
            </DropdownMenuItem>
            
            <DropdownMenuSeparator />
            <DropdownMenuLabel>排序顺序</DropdownMenuLabel>
            
            <DropdownMenuItem @click="setSortOrder('asc')">
                <ArrowUp class="w-4 h-4 mr-2" />
                升序
                <Check v-if="sortOrder === 'asc'" class="w-4 h-4 ml-auto" />
            </DropdownMenuItem>
            
            <DropdownMenuItem @click="setSortOrder('desc')">
                <ArrowDown class="w-4 h-4 mr-2" />
                降序
                <Check v-if="sortOrder === 'desc'" class="w-4 h-4 ml-auto" />
            </DropdownMenuItem>
        </DropdownMenuContent>
    </DropdownMenu>
</template>

<script setup>
import { useImages } from '@/composables/useImages'
import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ArrowUpDown, ArrowUp, ArrowDown, Clock, HardDrive, FileText, Maximize, Check } from 'lucide-vue-next'

const { sortBy, sortOrder, setSortBy, setSortOrder } = useImages()

const sortLabels = {
    time: '时间',
    size: '大小',
    name: '名称',
    dimensions: '尺寸'
}
</script>
