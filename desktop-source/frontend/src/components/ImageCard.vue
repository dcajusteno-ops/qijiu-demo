<script setup>
import { ref } from 'vue'
import { Check, Trash2, Heart, Tags, FileImage, Layers, StickyNote } from 'lucide-vue-next'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

const props = defineProps({
    image: Object,
    selectable: Boolean,
    selected: Boolean,
    hasNote: Boolean,
})

const emit = defineEmits(['view', 'delete', 'toggle', 'toggle-favorite', 'manage-tags', 'open-location', 'manage-favorites'])

// 3D Tilt State
const cardRef = ref(null)
const tiltStyle = ref({ transform: 'none' })
const glareStyle = ref({ opacity: 0, transform: 'translate(-50%, -50%)' })

const handleMouseMove = (e) => {
    if (!cardRef.value) return
    
    const rect = cardRef.value.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    
    const centerX = rect.width / 2
    const centerY = rect.height / 2
    
    const rotateX = ((y - centerY) / centerY) * -10
    const rotateY = ((x - centerX) / centerX) * 10
    
    tiltStyle.value = {
        transform: `perspective(1000px) rotateX(${rotateX}deg) rotateY(${rotateY}deg)`
    }
    
    glareStyle.value = {
        opacity: 0.4,
        left: `${x}px`,
        top: `${y}px`,
        transform: 'translate(-50%, -50%)'
    }
}

const handleMouseLeave = () => {
    tiltStyle.value = { transform: 'perspective(1000px) rotateX(0deg) rotateY(0deg)' }
    glareStyle.value = { opacity: 0, transform: 'translate(-50%, -50%)' }
}
</script>

<template>
    <div 
      ref="cardRef"
      class="group relative rounded-xl border bg-card text-card-foreground shadow-sm overflow-hidden cursor-pointer transition-all duration-300 ease-out select-none"
      :class="{ 'ring-2 ring-primary bg-accent': selected }"
      :style="tiltStyle"
      @click="selectable ? $emit('toggle') : $emit('view')"
      @mousemove="handleMouseMove"
      @mouseleave="handleMouseLeave"
    >
        <div 
            class="absolute pointer-events-none z-20 w-64 h-64 rounded-full bg-white/20 blur-3xl transition-opacity duration-300"
            :style="glareStyle"
        ></div>

        <div class="relative aspect-[3/4] bg-muted overflow-hidden">
            <img 
               :src="image.path" 
               :alt="image.name" 
               loading="lazy" 
               decoding="async"
               class="h-full w-full object-cover transition-transform duration-700 ease-in-out group-hover:scale-110"
            />
            
            <!-- Selection Overlay -->
            <div v-if="selectable" class="absolute top-3 left-3 z-10">
                <div 
                  class="h-6 w-6 rounded-full border-2 flex items-center justify-center transition-all duration-300"
                  :class="selected ? 'bg-primary border-primary text-primary-foreground scale-110 shadow-lg' : 'bg-black/40 border-white/80'"
                >
                    <Check v-if="selected" class="h-3.5 w-3.5" />
                </div>
            </div>

            <!-- Stack Indicator Overlay -->
            <div v-if="image.isStackPrimary && image.stackCount > 1" class="absolute bottom-3 right-3 z-10">
                <div class="h-6 px-2 rounded-full border-2 bg-black/60 backdrop-blur-md border-white/40 text-white flex items-center justify-center shadow-lg gap-1">
                    <Layers class="h-3.5 w-3.5" />
                    <span class="text-xs font-medium">{{ image.stackCount }}</span>
                </div>
            </div>

            <!-- Hover Actions -->
            <div v-if="!selectable" class="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-all duration-300 translate-x-2 group-hover:translate-x-0 flex flex-col gap-2 z-30">
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger as-child>
                        <button 
                           class="h-9 w-9 rounded-full bg-black/60 backdrop-blur-md text-white flex items-center justify-center shadow-lg hover:bg-white hover:text-red-500 transition-all hover:scale-110 active:scale-90"
                           @click.stop="emit('manage-favorites', image)" 
                        >
                            <Heart class="h-4 w-4" :class="{ 'fill-red-500 text-red-500': image.isFavorite }" />
                        </button>
                    </TooltipTrigger>
                    <TooltipContent side="left">
                      <p>管理收藏</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                <!-- ... other buttons similar style ... -->
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger as-child>
                        <button 
                           class="h-9 w-9 rounded-full bg-black/60 backdrop-blur-md text-white flex items-center justify-center shadow-lg hover:bg-white hover:text-amber-500 transition-all hover:scale-110 active:scale-90"
                           @click.stop="$emit('open-location', image)" 
                        >
                            <FileImage class="h-4 w-4" />
                        </button>
                    </TooltipTrigger>
                    <TooltipContent side="left">
                      <p>打开文件</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger as-child>
                        <button 
                           class="h-9 w-9 rounded-full bg-black/60 backdrop-blur-md text-white flex items-center justify-center shadow-lg hover:bg-white hover:text-blue-500 transition-all hover:scale-110 active:scale-90"
                           @click.stop="$emit('manage-tags', image)" 
                        >
                            <Tags class="h-4 w-4" />
                        </button>
                    </TooltipTrigger>
                    <TooltipContent side="left">
                      <p>管理标签</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>

                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger as-child>
                        <button 
                           class="h-9 w-9 rounded-full bg-white/90 text-destructive dark:bg-black/60 dark:text-white backdrop-blur-md flex items-center justify-center shadow-lg hover:bg-destructive hover:text-white transition-all hover:scale-110 active:scale-90"
                           @click.stop="$emit('delete')" 
                        >
                            <Trash2 class="h-4 w-4" />
                        </button>
                    </TooltipTrigger>
                    <TooltipContent side="left">
                      <p>删除图片</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
            </div>
        </div>
        
        <div class="p-4 flex items-center justify-between gap-2 bg-gradient-to-b from-card to-card/50">
            <span class="text-sm font-semibold truncate flex-1 min-w-0 text-foreground/90" :title="image.name">{{ image.name }}</span>
            <div class="flex items-center gap-1 shrink-0">
                <StickyNote v-if="hasNote" class="h-3 w-3 text-amber-500" />
                <span class="text-[10px] font-mono bg-muted px-1.5 py-0.5 rounded text-muted-foreground">{{ (image.size / 1024 / 1024).toFixed(2) }}MB</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.group:hover {
    z-index: 50;
    box-shadow: 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1);
}
</style>
