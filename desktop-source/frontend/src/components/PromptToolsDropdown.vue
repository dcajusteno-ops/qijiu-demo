<script setup>
import { Link, ExternalLink } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

const props = defineProps({
    variant: { type: String, default: 'ghost' },
    size: { type: String, default: 'icon' },
    showText: { type: Boolean, default: false },
    text: { type: String, default: '提示词辅助' }
})

const promptToolLinks = [
    { name: 'QPipi 提示词', url: 'https://prompt.qpipi.com/', icon: 'QP' },
    { name: 'NovelAI 标签库', url: 'https://tags.novelai.dev/', icon: 'NV' },
    { name: 'OPS 提示词', url: 'https://prompt.newzone.top/app/zh', icon: 'OP' },
    { name: 'Gwliang 提示词', url: 'https://prompt.gwliang.com/', icon: 'GW' },
    { name: 'AiWind 提示词', url: 'https://www.aiwind.org/?ref=producthunt', icon: 'AW' },
]
</script>

<template>
    <DropdownMenu>
        <DropdownMenuTrigger asChild>
            <Button :variant="variant" :size="size" :class="showText ? 'gap-2' : ''" :title="text">
                <Link class="w-4 h-4" />
                <span v-if="showText">{{ text }}</span>
            </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-56">
            <DropdownMenuLabel>提示词辅助工具</DropdownMenuLabel>
            <DropdownMenuSeparator />
            
            <DropdownMenuItem 
                v-for="link in promptToolLinks" 
                :key="link.url"
                asChild
            >
                <a :href="link.url" target="_blank" class="flex items-center gap-2 cursor-pointer w-full">
                    <div class="w-5 h-5 rounded text-[9px] font-bold bg-primary/10 flex items-center justify-center shrink-0 border border-primary/20 text-primary">
                        {{ link.icon }}
                    </div>
                    <span class="flex-1 truncate">{{ link.name }}</span>
                    <ExternalLink class="w-3 h-3 text-muted-foreground opacity-50" />
                </a>
            </DropdownMenuItem>
        </DropdownMenuContent>
    </DropdownMenu>
</template>
