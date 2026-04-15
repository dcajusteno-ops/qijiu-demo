<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@/components/ui/carousel'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import Autoplay from 'embla-carousel-autoplay'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import Lightbox from '@/components/Lightbox.vue'
import { useImages } from '@/composables/useImages'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ImageIcon, HardDrive, Heart, Tag, ArrowRight, Sparkles, Maximize2, BarChart3 } from 'lucide-vue-next'
import { toast } from 'vue-sonner' // Import toast
import PromptToolsDropdown from '@/components/PromptToolsDropdown.vue'
import * as App from '@/api'

// Custom Confirm Dialog Logic
const deleteDialogOpen = ref(false)
const deleteDialogMessage = ref('')
let deleteDialogResolve = null

const confirmDialog = (msg) => {
    deleteDialogMessage.value = msg
    deleteDialogOpen.value = true
    return new Promise((resolve) => {
        deleteDialogResolve = resolve
    })
}

const handleConfirm = () => {
    deleteDialogOpen.value = false
    if (deleteDialogResolve) {
        deleteDialogResolve(true)
        deleteDialogResolve = null
    }
}

const handleCancel = () => {
    deleteDialogOpen.value = false
    if (deleteDialogResolve) {
        deleteDialogResolve(false)
        deleteDialogResolve = null
    }
}

// Watch for dialog closing (e.g. via Escape or clicking outside)
watch(deleteDialogOpen, (newVal) => {
    if (!newVal && deleteDialogResolve) {
        deleteDialogResolve(false)
        deleteDialogResolve = null
    }
})

// Pass dependencies to useImages
const { 
  images, 
  loading: imagesLoading, 
  favorites, 
  favoriteGroups,
  tags, 
  imageTags, 
  toggleFavorite,
  toggleRoot,
  activeRoot,
  addTagToImage,
  removeTagFromImage,
  handleDelete,
  openImageLocation,
  fetchImages 
} = useImages(
    (msg, type) => {
        if (type === 'error') toast.error(msg)
        else if (type === 'success') toast.success(msg)
        else toast.info(msg)
    },
    confirmDialog
)

// Wrapper to close lightbox after delete
const handleLightboxDelete = async (img) => {
    await handleDelete(img)
    // Check if image is gone from the current list context
    const stillExists = images.value.find(i => i.path === img.path)
    if (!stillExists) {
        lightboxOpen.value = false
    }
}

// --- Data Computation ---

// 1. Recent Images (Latest 4 by modTime)
const recentImages = computed(() => {
    if (!images.value || images.value.length === 0) return []
    // Create a shallow copy to sort
    return [...images.value]
        .sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
        .slice(0, 15)
})

// 2. Statistics
const stats = ref({
    todayCount: 0,
    totalCount: 0,
    totalSize: 0
})
const statsLoading = ref(true)
const greetingMessages = [
  '欢迎回来，你的创作空间已经准备好了。',
  '今天也适合开工，灵感和作品都在等你。',
  '新的画面，新的惊喜，先挑一张开始吧。',
  '愿你今天出图顺利，每一张都更接近想要的感觉。',
  '工作台已经就绪，去看看最近的新作品吧。',
  '灵感补给完毕，现在就开始新的创作。',
  '先看几眼成果，再继续把灵感变成图像。',
  '今天的创作舞台已经亮灯，随时可以开始。',
]
const greetingMessage = ref(greetingMessages[0])

const pickGreetingMessage = () => {
  const randomIndex = Math.floor(Math.random() * greetingMessages.length)
  greetingMessage.value = greetingMessages[randomIndex]
}

const loadStats = async () => {
    statsLoading.value = true
    try {
        const data = await App.GetStatistics('day')
        stats.value = {
            todayCount: data.todayCount || 0,
            totalCount: data.totalCount || 0,
            totalSize: data.totalSize || 0
        }
    } catch (e) {
        console.error('Failed to load stats', e)
    } finally {
        statsLoading.value = false
    }
}

const formatSize = (bytes) => {
    if (!bytes) return '0 B'
    const gb = bytes / (1024 * 1024 * 1024)
    if (gb >= 1) return `${gb.toFixed(2)} GB`
    const mb = bytes / (1024 * 1024)
    return `${mb.toFixed(2)} MB`
}

// 3. Recent Favorites (Top 10 by modTime)
const recentFavorites = computed(() => {
    return images.value
        .filter(img => favorites.value.has(img.relPath))
        .sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
        .slice(0, 10)
})

// --- Lightbox & Interaction ---
const loading = computed(() => imagesLoading.value || statsLoading.value)

let pollInterval = null

// Lightbox State
const lightboxOpen = ref(false)
const currentImage = ref(null)
const currentImageIndex = ref(0)
// We need a context for the lightbox navigation. 
// Since we are clicking from "recentImages" or "recentFavorites", 
// we should probably navigate within THAT list.
const lightboxContext = ref([]) // 'recent' or 'favorites' list

const openLightbox = (img, list) => {
  lightboxContext.value = list
  currentImageIndex.value = list.findIndex(i => i.path === img.path)
  currentImage.value = img
  // Pause autoplay when lightbox is open
  if (plugin && plugin.stop) plugin.stop()
  lightboxOpen.value = true
}

const handleLightboxClose = () => {
  lightboxOpen.value = false
  // Resume autoplay
  if (plugin && plugin.reset) plugin.reset()
}

const handleNavigate = (direction) => {
    const list = lightboxContext.value
    if (direction === 'prev' && currentImageIndex.value > 0) {
        currentImageIndex.value--
    } else if (direction === 'next' && currentImageIndex.value < list.length - 1) {
        currentImageIndex.value++
    }
    currentImage.value = list[currentImageIndex.value]
}

onMounted(() => {
  pickGreetingMessage()
  loadStats()
  // Ensure we have images. If empty (e.g. reload on dashboard), trigger fetch.
  // App.vue calls startPolling, but we might check here to be safe or show loading state.
  if (images.value.length === 0) {
      // Just manually call fetch once to speed up initial paint if polling hasn't hit yet
      fetchImages() 
  }
  pollInterval = setInterval(loadStats, 30000) // Poll stats every 30s
})

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
})

const plugin = Autoplay({
  delay: 4000,
  stopOnMouseEnter: true,
  stopOnInteraction: false,
})

const navigateToFavorites = () => {
    toggleRoot('favorites')
}

</script>

<template>
  <div class="h-full flex flex-col p-6 overflow-hidden select-none">
    <!-- Header -->
    <div class="flex-none mb-6">
       <div class="flex items-center justify-between">
           <div class="space-y-1">
             <h2 class="text-3xl font-bold tracking-tight">工作台总览</h2>
             <p class="text-muted-foreground">{{ greetingMessage }}</p>
           </div>
           
           <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="toggleRoot('favorites')" title="查看收藏夹" class="bg-card/40 backdrop-blur-md border border-border/50 hover:bg-card/60">
                    <Heart class="w-4 h-4 mr-2 text-red-500 fill-red-500/10" />
                    收藏夹
                </Button>
                <Button variant="outline" size="sm" @click="toggleRoot('statistics')" title="查看数据视界" class="bg-card/40 backdrop-blur-md border border-border/50 hover:bg-card/60">
                    <BarChart3 class="w-4 h-4 mr-2 text-blue-500" />
                    数据视界
                </Button>
                
                <PromptToolsDropdown variant="outline" size="sm" :show-text="true" text="提示词助手" class="bg-card/40 backdrop-blur-md border border-border/50 hover:bg-card/60" />
           </div>
       </div>
    </div>

    <!-- Main Content Area (Flex Grow to fill available space) -->
    <div class="flex-1 min-h-0 flex flex-col xl:flex-row gap-6">
        
        <!-- Left: Hero Spotlight (65% width) -->
        <div class="flex-[6.5] min-w-0 flex flex-col gap-4">
             <!-- Spotlight Card -->
             <div 
                v-if="recentImages.length > 0"
                class="flex-1 relative rounded-3xl overflow-hidden border bg-background/50 shadow-sm group cursor-pointer"
                @click="openLightbox(recentImages[0], recentImages)"
             >
                <!-- Cinematic Background -->
                <div class="absolute inset-0 z-0">
                    <img 
                        :src="recentImages[0].path" 
                        class="w-full h-full object-cover blur-3xl opacity-40 dark:opacity-20 scale-110 transition-transform duration-[2s] group-hover:scale-100" 
                        loading="eager"
                        decoding="async"
                    />
                     <div class="absolute inset-0 bg-gradient-to-t from-background via-background/10 to-transparent"></div>
                </div>

                <!-- Main Hero Image -->
                <div class="relative z-10 w-full h-full p-8 flex items-center justify-center">
                     <img 
                        :src="recentImages[0].path" 
                        class="max-w-full max-h-full object-contain shadow-2xl rounded-lg transition-transform duration-500 group-hover:scale-[1.02]" 
                        loading="eager"
                        decoding="async"
                     />
                </div>

                <!-- Hero Overlay -->
                <div class="absolute bottom-0 left-0 right-0 p-8 z-20 bg-gradient-to-t from-background to-transparent pt-24 translate-y-4 group-hover:translate-y-0 transition-transform duration-500">
                    <div class="flex items-end justify-between">
                        <div>
                             <Badge variant="secondary" class="mb-2 bg-primary/10 text-primary hover:bg-primary/20 border-0">
                                <Sparkles class="w-3 h-3 mr-1" /> 最新作品
                             </Badge>
                             <h1 class="text-3xl font-bold tracking-tight text-foreground">{{ recentImages[0].name }}</h1>
                             <p class="text-muted-foreground mt-1 text-sm font-medium">
                                {{ new Date(recentImages[0].modTime).toLocaleString() }} 路 {{ (recentImages[0].size / 1024 / 1024).toFixed(2) }} MB
                             </p>
                        </div>
                        
                        <Button size="lg" class="shadow-lg rounded-full px-6" @click.stop="openLightbox(recentImages[0], recentImages)">
                            查看详情 <ArrowRight class="w-4 h-4 ml-2" />
                        </Button>
                    </div>
                </div>
             </div>
             
             <div v-else class="flex-1 flex items-center justify-center border-2 border-dashed rounded-3xl bg-muted/5">
                 <div class="text-center space-y-2">
                     <ImageIcon class="w-12 h-12 text-muted-foreground mx-auto opacity-50" />
                     <p class="text-muted-foreground font-medium">暂无图片，快去生成吧！</p>
                 </div>
             </div>
        </div>

        <!-- Right: Side Rail (35% width, wider now) - Stats + Random -->
        <div class="flex-[3.5] min-w-[320px] flex flex-col gap-4 overflow-y-auto pr-1">
            <!-- Stats Grid (Compact 2x2) -->
            <div class="grid grid-cols-2 gap-3 shrink-0">
                 <Card class="group hover:border-amber-500/50 transition-colors">
                    <CardHeader class="p-3 pb-1 flex flex-row items-center justify-between space-y-0">
                        <CardTitle class="text-sm font-bold text-muted-foreground uppercase tracking-wider">今日新增</CardTitle>
                        <Sparkles class="w-4 h-4 text-amber-500" />
                    </CardHeader>
                    <CardContent class="p-3 pt-0">
                        <div class="text-2xl font-bold text-amber-600 dark:text-amber-400">{{ stats.todayCount.toLocaleString() }}</div>
                    </CardContent>
                </Card>
                 <Card>
                    <CardHeader class="p-3 pb-1">
                        <CardTitle class="text-sm font-bold text-muted-foreground uppercase tracking-wider">图片总数</CardTitle>
                    </CardHeader>
                    <CardContent class="p-3 pt-0">
                        <div class="text-2xl font-bold">{{ stats.totalCount.toLocaleString() }}</div>
                    </CardContent>
                </Card>
                 <Card class="col-span-2">
                     <CardHeader class="p-3 pb-1 flex flex-row items-center justify-between space-y-0">
                        <CardTitle class="text-sm font-bold text-muted-foreground uppercase tracking-wider">存储占用</CardTitle>
                        <HardDrive class="w-4 h-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent class="p-3 pt-0">
                         <div class="text-2xl font-bold">{{ formatSize(stats.totalSize) }}</div>
                         <div class="w-full bg-secondary h-1.5 rounded-full mt-2 overflow-hidden">
                             <div class="bg-primary h-full rounded-full" style="width: 45%"></div> <!-- Mock progress -->
                         </div>
                    </CardContent>
                 </Card>
            </div>

        </div>
    </div>

    <!-- Bottom: Flow Strip (Recent Flow) -->
    <div class="flex-none mt-6 space-y-3">
         <div class="flex items-center justify-between px-1">
             <h3 class="text-lg font-semibold tracking-tight">最近更新</h3>
             <Button variant="ghost" size="sm" class="text-muted-foreground hover:text-foreground text-xs" @click="toggleRoot('日期归档')">
                 查看全部 <ArrowRight class="w-3 h-3 ml-1" />
             </Button>
         </div>

         <div class="relative group px-1">
             <!-- Carousel for Recent Images -->
             <Carousel
                class="w-full"
                :opts="{ align: 'start', dragFree: true }"
             >
                <CarouselContent class="-ml-4 pb-4"> <!-- Negative margin for spacing -->
                    <!-- First item is skipped (it's the hero) -->
                    <CarouselItem 
                        v-for="img in recentImages.slice(1, 16)" 
                        :key="img.path"
                        class="pl-4 basis-auto" 
                    >
                         <div 
                            class="w-[140px] aspect-[4/5] rounded-xl overflow-hidden cursor-pointer relative group/item shadow-sm border bg-muted/20"
                            @click="openLightbox(img, recentImages)"
                         >
                            <img 
                                :src="img.path" 
                                class="w-full h-full object-cover transition-transform duration-500 group-hover/item:scale-110"
                                loading="lazy"
                                decoding="async"
                            />
                            <!-- Hover Overlay -->
                            <div class="absolute inset-0 bg-black/40 opacity-0 group-hover/item:opacity-100 transition-opacity duration-200 flex items-center justify-center">
                                <Maximize2 class="w-6 h-6 text-white opacity-80" />
                            </div>
                         </div>
                    </CarouselItem>

                    <!-- View More Card -->
                    <CarouselItem class="pl-4 basis-auto">
                        <div 
                             class="w-[140px] aspect-[4/5] rounded-xl border-2 border-dashed border-muted flex flex-col items-center justify-center gap-2 cursor-pointer hover:bg-muted/50 transition-colors text-muted-foreground hover:text-foreground"
                             @click="toggleRoot('日期归档')"
                        >
                             <div class="p-2 rounded-full bg-muted">
                                 <ArrowRight class="w-4 h-4" />
                             </div>
                             <span class="text-xs font-medium">查看更多</span>
                        </div>
                    </CarouselItem>
                </CarouselContent>
                
                <!-- Navigation Buttons -->
                <CarouselPrevious class="-left-4 bg-card/60 backdrop-blur-md border shadow-lg hidden group-hover:flex hover:bg-card/80 hover:scale-110 transition-all duration-300" />
                <CarouselNext class="-right-4 bg-card/60 backdrop-blur-md border shadow-lg hidden group-hover:flex hover:bg-card/80 hover:scale-110 transition-all duration-300" />
             </Carousel>
         </div>
    </div>

    <Lightbox 
        :image="currentImage" 
        :images="lightboxContext"
        :current-index="currentImageIndex"
        :isOpen="lightboxOpen"
        :favorite-groups="favoriteGroups"
        :tags="tags"
        :image-tags="imageTags"
        @close="handleLightboxClose"
        @navigate="handleNavigate"
        @toggle-favorite="toggleFavorite"
        @add-tag="addTagToImage"
        @remove-tag="removeTagFromImage"
        @delete="handleLightboxDelete"
        @open-location="openImageLocation"
        @favorite-groups-changed="fetchImages"
    />

    <AlertDialog v-model:open="deleteDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>
            {{ deleteDialogMessage }}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel @click="handleCancel">取消</AlertDialogCancel>
          <AlertDialogAction @click="handleConfirm" class="bg-red-500 hover:bg-red-600">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>

