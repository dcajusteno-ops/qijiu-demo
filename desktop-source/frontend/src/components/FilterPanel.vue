<script setup>
import { Filter } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'
import { useImages } from '@/composables/useImages'
import { Switch } from '@/components/ui/switch'

const { filters, isStackingEnabled, toggleStacking } = useImages()

const resetFilters = () => {
  filters.value.dateRange = { start: null, end: null }
  filters.value.size = { min: null, max: null }
  filters.value.dimensions = { minW: null, minH: null }
}
</script>

<template>
  <Popover>
    <PopoverTrigger as-child>
      <Button variant="ghost" size="icon" title="高级筛选">
        <Filter
          class="h-4 w-4"
          :class="{ 'text-primary fill-primary/20': filters.dateRange.start || filters.dateRange.end || filters.size.min || filters.size.max || filters.dimensions.minW || filters.dimensions.minH }"
        />
      </Button>
    </PopoverTrigger>
    <PopoverContent class="w-80 p-4" align="start" side="bottom" :side-offset="4">
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h4 class="font-medium leading-none">筛选条件</h4>
          <Button
            variant="ghost"
            size="sm"
            class="h-auto p-0 text-xs text-muted-foreground hover:text-foreground"
            @click="resetFilters"
          >
            重置
          </Button>
        </div>

        <Separator />

        <div class="space-y-2">
          <Label class="text-xs font-semibold text-muted-foreground">日期范围</Label>
          <div class="flex items-center gap-2">
            <Input type="date" class="h-8 text-xs px-2" v-model="filters.dateRange.start" />
            <span class="text-muted-foreground">-</span>
            <Input type="date" class="h-8 text-xs px-2" v-model="filters.dateRange.end" />
          </div>
        </div>

        <div class="space-y-2">
          <Label class="text-xs font-semibold text-muted-foreground">文件大小 (MB)</Label>
          <div class="flex items-center gap-2">
            <Input type="number" placeholder="最小" class="h-8 text-xs px-2" v-model.number="filters.size.min" />
            <span class="text-muted-foreground">-</span>
            <Input type="number" placeholder="最大" class="h-8 text-xs px-2" v-model.number="filters.size.max" />
          </div>
        </div>

        <div class="space-y-2">
          <Label class="text-xs font-semibold text-muted-foreground">最小尺寸 (px)</Label>
          <div class="flex items-center gap-2">
            <Input type="number" placeholder="宽" class="h-8 text-xs px-2" v-model.number="filters.dimensions.minW" />
            <span class="text-muted-foreground">x</span>
            <Input type="number" placeholder="高" class="h-8 text-xs px-2" v-model.number="filters.dimensions.minH" />
          </div>
        </div>

        <Separator />

        <div class="flex items-center justify-between space-x-2">
          <Label
            for="stacking-toggle"
            class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            自动折叠相似图片
          </Label>
          <Switch
            id="stacking-toggle"
            :checked="isStackingEnabled"
            @update:checked="toggleStacking"
          />
        </div>
      </div>
    </PopoverContent>
  </Popover>
</template>
