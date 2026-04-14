<script setup>
import { computed, ref, onMounted, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import TagManager from './TagManager.vue'
import {
  Calendar,
  FlaskConical,
  Wrench,
  BookOpen,
  Folder,
  Moon,
  Sun,
  Trash2,
  CheckSquare,
  Heart,
  Tags,
  Settings2,
  X,
  Plus,
  PanelLeftClose,
  PanelLeftOpen,
  Eraser,
  ChevronDown,
  ChevronRight,
  LayoutDashboard,
  Link,
  FolderSymlink,
  Cpu,
  Puzzle,
  Maximize,
  Bookmark,
  Sparkles,
  Search,
  FolderTree,
} from 'lucide-vue-next'
import { isDark, toggleTheme } from '@/theme'
import TrashDialog from './TrashDialog.vue'
import LauncherDialog from './LauncherDialog.vue'
import CustomRootDialog from './CustomRootDialog.vue'
import FavoriteGroupsDialog from './FavoriteGroupsDialog.vue'
import PromptTemplateDialog from './PromptTemplateDialog.vue'
import { TerminalSquare } from 'lucide-vue-next'
import { availableIcons } from '@/lib/icons'
import * as App from '@/api'

const props = defineProps({
  fileTree: { type: Array, required: true },
  activeRoot: { type: String, required: true },
  activeSub: { type: String, required: true },
  activeChild: { type: String, default: '' },
  isSelectionMode: { type: Boolean, default: false },
  tags: { type: Array, default: () => [] },
  activeTagFilter: { type: String, default: null },
  getTagCount: { type: Function, required: true },
  collapsed: { type: Boolean, default: false },
  customRoots: { type: Array, default: () => [] },
  favoriteGroups: { type: Array, default: () => [] },
  activeSmartAlbum: { type: Object, default: null },
})

const emit = defineEmits([
  'update:activeRoot',
  'update:activeSub',
  'update:activeChild',
  'toggle-selection-mode',
  'clean-empty-folders',
  'create-tag',
  'delete-tag',
  'batch-delete-tags',
  'update-tag',
  'batch-update-tags',
  'toggle-tag-filter',
  'refresh-images',
  'toggle-collapse',
  'custom-root-change',
  'favorite-group-change',
  'clear-preview-cache',
  'organize-files',
  'smart-album-select',
])

const getIcon = (node) => {
    // 1. Check if node specifically has an icon configured (from CustomRoot or rootNodes)
    if (node.icon && availableIcons[node.icon]) return availableIcons[node.icon]
    
    // 2. Fallback to name-based logic for built-in folders
    const name = node.name || ''
    if (name.includes('日期')) return Calendar
    if (name.includes('XYZ')) return FlaskConical
    if (name.includes('修复')) return Wrench
    if (name.includes('收藏')) return Heart
    
    return Folder
}

const handleRootClick = (rootName) => {
    emit('update:activeRoot', rootName)
}

const isExpanded = (id) => {
    if (!props.activeSub) return false
    return props.activeSub === id || props.activeSub.startsWith(id + '/')
}

const handleSubClick = (id) => {
    if (isExpanded(id)) {
        // Already expanded (exactly or ancestor) -> Collapse this folder level
        const lastSlash = id.lastIndexOf('/')
        if (lastSlash > 0) {
            const parentId = id.substring(0, lastSlash)
            // If parent is the root itself, clear activeSub
            if (parentId === props.activeRoot) {
                emit('update:activeSub', '')
            } else {
                emit('update:activeSub', parentId)
            }
        } else {
            emit('update:activeSub', '')
        }
    } else {
        // Not expanded -> Select/Expand this folder
        emit('update:activeSub', id)
    }
    emit('update:activeChild', '')
}

// Group tags by user-defined categories
const tagsByCategory = computed(() => {
    const groups = {}
    props.tags.forEach(tag => {
        const categoryName = tag.category || '未分组'
        if (!groups[categoryName]) {
            groups[categoryName] = []
        }
        groups[categoryName].push(tag)
    })
    return groups
})

// Get category Names sorted ("未分组" always last)
const categoryNames = computed(() => {
    const names = Object.keys(tagsByCategory.value)
    return names.sort((a, b) => {
        if (a === '未分组') return 1
        if (b === '未分组') return -1
        return a.localeCompare(b)
    })
})

const getRecursiveCount = (node) => {
    if (typeof node.count === 'number') return node.count
    let c = node.images ? node.images.length : 0
    // Try both subs and children properties as we aliased them but best to be safe
    const children = node.subs || node.children || []
    if (children.length > 0) {
        c += children.reduce((acc, n) => acc + getRecursiveCount(n), 0)
    }
    return c
}

const formatFolderName = (name) => {
    // Only add "月" if name matches exactly 2 digits (e.g. 01, 12)
    if (/^\d{2}$/.test(name)) return name + '月'
    return name
}

const showTrashDialog = ref(false)
const showLauncherDialog = ref(false)
const showCustomRootDialog = ref(false)
const showFavoriteGroupsDialog = ref(false)
const showPromptTemplateDialog = ref(false)
const isTagsCollapsed = ref(false)
const showUtilityMenu = ref(false)

// Smart Albums state
const isSmartAlbumsCollapsed = ref(false)
const smartAlbumFields = ref([])
const smartAlbumsData = ref({}) // field -> albums[]
const smartAlbumsLoading = ref({})
const smartAlbumSearch = ref({}) // field -> search text

const smartAlbumFieldIcons = {
  model: Cpu,
  sampler: FlaskConical,
  lora: Puzzle,
  dimensions: Maximize,
}

const openSmartAlbumPopover = async (fieldKey) => {
  if (!smartAlbumsData.value[fieldKey]) {
    smartAlbumsLoading.value[fieldKey] = true
    try {
      const albums = await App.GetSmartAlbums(fieldKey)
      smartAlbumsData.value[fieldKey] = albums || []
      smartAlbumSearch.value[fieldKey] = ''
    } catch (e) {
      console.error('Failed to fetch smart albums:', e)
    } finally {
      smartAlbumsLoading.value[fieldKey] = false
    }
  }
}

const filteredSmartAlbums = (fieldKey) => {
  const albums = smartAlbumsData.value[fieldKey] || []
  const search = (smartAlbumSearch.value[fieldKey] || '').trim().toLowerCase()
  if (!search) return albums
  return albums.filter(a => a.value.toLowerCase().includes(search))
}

const handleSmartAlbumClick = (album) => {
  emit('smart-album-select', { field: album.field, value: album.value, paths: album.paths })
}

const clearSmartAlbumFilter = () => {
  emit('smart-album-select', null)
}

// Refresh smart album data (clear cache so next open re-fetches)
const refreshSmartAlbums = () => {
  smartAlbumsData.value = {}
}

// Auto-refresh smart albums when files change
watch(() => props.fileTree, () => {
  refreshSmartAlbums()
}, { deep: true })

onMounted(async () => {
  try {
    const fields = await App.GetSmartAlbumFields()
    smartAlbumFields.value = fields || []
  } catch (e) {
    console.error('Failed to fetch smart album fields:', e)
  }
})

const closeUtilityMenu = () => {
    showUtilityMenu.value = false
}

const openTrashManager = () => {
    showTrashDialog.value = true
    closeUtilityMenu()
}

const openDocumentation = () => {
    emit('update:activeRoot', 'documentation')
    closeUtilityMenu()
}

const cleanEmptyFolders = () => {
    emit('clean-empty-folders')
    closeUtilityMenu()
}

const clearPreviewCache = () => {
    emit('clear-preview-cache')
    closeUtilityMenu()
}

const handleThemeToggle = (event) => {
    toggleTheme(event)
    closeUtilityMenu()
}

const openFavoriteGroups = () => {
    showFavoriteGroupsDialog.value = true
    closeUtilityMenu()
}

const openLauncher = () => {
    showLauncherDialog.value = true
    closeUtilityMenu()
}

const openPromptTemplates = () => {
    showPromptTemplateDialog.value = true
    closeUtilityMenu()
}

const openCustomRootManager = () => {
    showCustomRootDialog.value = true
    closeUtilityMenu()
}

// Hover Drawer State
const hoveredRoot = ref(null)
const hoverTop = ref(0)
let hoverLeaveTimeout = null

const handleHoverEnter = (event, root) => {
    if (hoverLeaveTimeout) {
        clearTimeout(hoverLeaveTimeout)
        hoverLeaveTimeout = null
    }
    hoveredRoot.value = root
    const rect = event.currentTarget.getBoundingClientRect()
    hoverTop.value = rect.top
}

const handleHoverLeave = () => {
    hoverLeaveTimeout = setTimeout(() => {
        hoveredRoot.value = null
    }, 300)
}

const cancelHoverLeave = () => {
    if (hoverLeaveTimeout) {
        clearTimeout(hoverLeaveTimeout)
        hoverLeaveTimeout = null
    }
}

const handleDrawerClick = (subId) => {
    if (hoveredRoot.value) {
        handleRootClick(hoveredRoot.value.id)
    }
    handleSubClick(subId)
    hoveredRoot.value = null
}

</script>

<template>
  <aside class="h-full bg-muted/30 border-r flex flex-col transition-all duration-300" :class="collapsed ? 'w-[60px]' : 'w-64'">
    
    <!-- Header / Title -->
    <div class="h-16 shrink-0 flex items-center px-4 border-b bg-background/50">
      <div v-if="!collapsed" class="flex items-center gap-2 text-primary font-semibold truncate tracking-tight">
          <svg class="h-5 w-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <span class="truncate">灵动图库</span>
      </div>
      <div v-else class="w-full flex justify-center text-primary">
          <svg class="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
      </div>
    </div>

    <!-- Main Scrollable Area -->
    <div class="flex-1 overflow-hidden flex flex-col min-h-0">
      <ScrollArea class="flex-1 overflow-y-auto">
        <div class="p-3 space-y-4">
          
          <!-- File Tree -->
          <div class="space-y-1">
            <!-- Home Dashboard Button -->
            <div class="space-y-1">
                <button 
                  class="w-full flex items-center justify-between px-3 py-2 text-sm font-medium rounded-md transition-colors hover:bg-secondary text-left relative group"
                  :class="activeRoot === 'dashboard' ? 'bg-secondary text-primary' : 'text-foreground/80'"
                  @click="$emit('update:activeRoot', 'dashboard')"
                  :title="collapsed ? '工作室概览' : ''"
                >
                  <div class="flex items-center gap-2">
                    <LayoutDashboard class="h-4 w-4" :class="activeRoot === 'dashboard' ? 'text-primary' : 'text-muted-foreground'" />
                    <span v-if="!collapsed" class="truncate">工作室概览</span>
                  </div>
                </button>
            </div>
            <div v-for="root in fileTree" :key="root.id" class="space-y-1">
                <!-- Root Level -->
                <button 
                  class="w-full flex items-center justify-between px-3 py-2 text-sm font-medium rounded-md transition-colors hover:bg-secondary text-left relative group"
                  :class="activeRoot === root.id ? 'bg-secondary text-primary' : 'text-foreground/80'"
                  @click="handleRootClick(root.id)"
                  @mouseenter="collapsed ? handleHoverEnter($event, root) : null"
                  @mouseleave="collapsed ? handleHoverLeave($event) : null"
                  :title="collapsed ? (root.displayName || root.name) : ''"
                >
                  <div class="flex items-center gap-2">
                    <component :is="getIcon(root)" class="h-4 w-4" :class="activeRoot === root.id ? 'text-primary' : 'text-muted-foreground'" />
                    <span v-if="!collapsed" class="truncate">{{ root.displayName || root.name }}</span>
                  </div>
                  <span v-if="!collapsed && (root.children && root.children.length > 0 || getRecursiveCount(root) > 0)" class="text-xs font-normal opacity-60">
                    {{ getRecursiveCount(root) }}
                  </span>
                </button>

                <!-- Children -->
                <div v-if="!collapsed && activeRoot === root.id && (root.subs || root.children)?.length > 0" class="pl-4 space-y-1 mt-1">
                  <!-- Level 1 -->
                  <div v-for="l1 in (root.subs || root.children)" :key="l1.id" class="space-y-1">
                      <button
                        @click="handleSubClick(l1.id)"
                        class="w-full flex items-center justify-between px-3 py-1.5 text-sm rounded-md transition-colors hover:bg-secondary/50 text-left"
                        :class="activeSub === l1.id ? 'bg-secondary/80 text-primary font-medium' : 'text-foreground/80'"
                      >
                        <span class="truncate">{{ l1.displayName || l1.name }}</span>
                        <span class="text-xs opacity-50">{{ getRecursiveCount(l1) }}</span>
                      </button>

                      <!-- Level 2 -->
                      <div v-if="isExpanded(l1.id) && l1.children && l1.children.length > 0" class="pl-4 border-l ml-3 space-y-1 my-1">
                        <div v-for="l2 in l1.children" :key="l2.id" class="space-y-1">
                            <button
                                @click="handleSubClick(l2.id)"
                                class="w-full flex items-center justify-between px-3 py-1 text-xs font-semibold rounded-md transition-colors uppercase tracking-wider hover:bg-secondary/50 text-left"
                                :class="activeSub === l2.id ? 'bg-secondary/80 text-primary' : 'text-muted-foreground'"
                            >
                                <span class="truncate">{{ l2.displayName || formatFolderName(l2.name) }}</span>
                                <span class="text-xs opacity-50 font-normal">{{ getRecursiveCount(l2) }}</span>
                            </button>

                            <!-- Level 3 -->
                            <div v-if="isExpanded(l2.id) && l2.children && l2.children.length > 0" class="space-y-0.5 pl-2 my-0.5">
                                <div v-for="l3 in l2.children" :key="l3.id" class="space-y-0.5">
                                    <button
                                        @click="handleSubClick(l3.id)"
                                        class="w-full flex items-center justify-between px-3 py-1 text-xs rounded-md transition-colors hover:bg-secondary/50 text-left"
                                        :class="activeSub === l3.id ? 'bg-secondary/80 text-primary font-medium' : 'text-muted-foreground'"
                                    >
                                        <span class="truncate">{{ l3.displayName || formatFolderName(l3.name) }}</span>
                                        <span class="text-xs opacity-70">{{ getRecursiveCount(l3) }}</span>
                                    </button>

                                    <!-- Level 4: Leaf Folders -->
                                    <div v-if="isExpanded(l3.id) && l3.children && l3.children.length > 0" class="space-y-0.5 pl-2 my-0.5">
                                        <button
                                            v-for="leaf in l3.children"
                                            :key="leaf.id"
                                            @click="handleSubClick(leaf.id)"
                                            class="w-full flex items-center justify-between px-3 py-1 text-xs rounded-md transition-colors hover:bg-secondary/50 text-left"
                                            :class="activeSub === leaf.id ? 'bg-secondary/80 text-primary font-medium' : 'text-muted-foreground'"
                                        >
                                            <span class="truncate">{{ leaf.displayName || leaf.name }}</span>
                                            <span class="text-xs opacity-70">{{ leaf.images?.length || 0 }}</span>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                      </div>
                  </div>
                </div>
            </div>
          </div>
          <Separator v-if="!collapsed" class="my-2" />

          <!-- Tag Manager area -->
          <div class="space-y-1 relative">
             <TagManager
               :tags="tags"
               :collapsed="collapsed"
               @create-tag="$emit('create-tag', $event)"
               @update-tag="$emit('update-tag', $event.id, $event.data)"
               @delete-tag="$emit('delete-tag', $event)"
               @batch-delete-tags="$emit('batch-delete-tags', $event)"
               @batch-update-tags="$emit('batch-update-tags', $event)"
               @toggle-tag-filter="$emit('toggle-tag-filter', $event)"
               :active-filter="activeTagFilter"
             />

             <!-- Tags List by Category (Expanded) -->
             <div v-if="!collapsed && tags.length > 0" class="mt-2 space-y-3">
                 <div v-for="category in categoryNames" :key="category" class="space-y-1">
                     <div class="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1">
                         {{ category }}
                     </div>
                     <div class="grid grid-cols-1 gap-0.5">
                         <button
                            v-for="tag in tagsByCategory[category]"
                            :key="tag.id"
                            @click="$emit('toggle-tag-filter', tag.id)"
                            class="flex items-center justify-between px-3 py-1.5 text-xs rounded-md transition-colors hover:bg-secondary/50 text-left border border-transparent"
                            :class="activeTagFilter === tag.id ? 'bg-secondary/80 text-primary border-primary/20 bg-primary/10' : 'text-foreground/80'"
                         >
                            <div class="flex items-center gap-2 overflow-hidden">
                                <div class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ backgroundColor: tag.color || '#ccc' }"></div>
                                <span class="truncate">{{ tag.name }}</span>
                            </div>
                            <span class="text-[10px] opacity-60 ml-2 font-normal">{{ getTagCount(tag.id) }}</span>
                         </button>
                     </div>
                 </div>
             </div>

             <!-- Tags Filter (Collapsed) -->
             <div v-if="collapsed && tags.length > 0" class="mt-2 flex justify-center">
                 <Popover>
                     <PopoverTrigger as-child>
                         <Button variant="ghost" class="w-full justify-center px-0 h-10" :class="activeTagFilter ? 'text-primary bg-primary/10' : ''" title="标签筛选">
                             <Tags class="h-5 w-5" />
                         </Button>
                     </PopoverTrigger>
                     <PopoverContent side="right" align="start" :side-offset="12" class="w-56 p-2 max-h-[60vh] overflow-y-auto">
                        <div class="text-xs font-semibold mb-2 px-2 text-foreground/80">标签筛选</div>
                        <div v-for="category in categoryNames" :key="category" class="space-y-1 mt-2 first:mt-0">
                            <div class="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider px-2 py-1">
                                {{ category }}
                            </div>
                            <div class="grid grid-cols-1 gap-0.5">
                                <button
                                   v-for="tag in tagsByCategory[category]"
                                   :key="tag.id"
                                   @click="$emit('toggle-tag-filter', tag.id)"
                                   class="flex items-center justify-between px-3 py-1.5 text-xs rounded-md transition-colors hover:bg-secondary/50 text-left border border-transparent"
                                   :class="activeTagFilter === tag.id ? 'bg-secondary/80 text-primary border-primary/20 bg-primary/10' : 'text-foreground/80'"
                                >
                                   <div class="flex items-center gap-2 overflow-hidden">
                                       <div class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ backgroundColor: tag.color || '#ccc' }"></div>
                                       <span class="truncate">{{ tag.name }}</span>
                                   </div>
                                   <span class="text-[10px] opacity-60 ml-2 font-normal">{{ getTagCount(tag.id) }}</span>
                                </button>
                            </div>
                        </div>
                     </PopoverContent>
                 </Popover>
             </div>

             <!-- Clear Tag Filter Button -->
             <div v-if="activeTagFilter && !collapsed" class="pt-2">
                <Button 
                  variant="ghost" 
                  size="sm"
                  class="w-full justify-start h-8 text-xs font-medium gap-2 shadow-sm border bg-background hover:bg-destructive/10 hover:text-destructive hover:border-destructive/20 transition-all rounded-md"
                  @click="$emit('toggle-tag-filter', activeTagFilter)"
                >
                  <X class="h-3.5 w-3.5 font-bold" />
                  清除标签筛选
                </Button>
            </div>
             <!-- Clear Filter Icon Only -->
            <div v-if="activeTagFilter && collapsed" class="pt-1 sticky bottom-0 bg-background/95 backdrop-blur-sm py-1 rounded-md w-full flex justify-center">
                <Button 
                  variant="ghost" 
                  size="icon"
                  class="h-8 w-8 text-destructive hover:bg-destructive/10 hover:text-destructive transition-all rounded-full border border-destructive/20 shadow-sm"
                  @click="$emit('toggle-tag-filter', activeTagFilter)"
                  title="清除标签筛选"
                >
                  <X class="h-4 w-4" />
                </Button>
            </div>
          </div>

          <Separator v-if="!collapsed" class="my-2" />

          <!-- Smart Albums Section -->
          <div v-if="!collapsed && smartAlbumFields.length > 0" class="space-y-1">
            <button
              class="w-full flex items-center justify-between px-3 py-2 text-sm font-medium rounded-md transition-colors hover:bg-secondary text-left"
              @click="isSmartAlbumsCollapsed = !isSmartAlbumsCollapsed"
            >
              <div class="flex items-center gap-2">
                <Sparkles class="h-4 w-4 text-purple-500" />
                <span>智能相册</span>
              </div>
              <ChevronDown v-if="!isSmartAlbumsCollapsed" class="h-4 w-4 text-muted-foreground" />
              <ChevronRight v-else class="h-4 w-4 text-muted-foreground" />
            </button>

            <div v-if="!isSmartAlbumsCollapsed" class="space-y-0.5 pl-1">
              <Popover
                v-for="field in smartAlbumFields"
                :key="field.key"
              >
                <PopoverTrigger as-child>
                  <button
                    class="w-full flex items-center justify-between px-3 py-1.5 text-xs rounded-md transition-colors hover:bg-secondary/50 text-left"
                    :class="activeSmartAlbum?.field === field.key ? 'bg-primary/10 text-primary font-medium' : 'text-foreground/80'"
                    @click="openSmartAlbumPopover(field.key)"
                  >
                    <div class="flex items-center gap-2">
                      <component :is="smartAlbumFieldIcons[field.key] || Folder" class="h-3.5 w-3.5" />
                      <span>{{ field.label }}</span>
                    </div>
                    <ChevronRight class="h-3 w-3 text-muted-foreground" />
                  </button>
                </PopoverTrigger>
                <PopoverContent side="right" align="center" :side-offset="12" class="w-72 p-0 max-h-[70vh] flex flex-col overflow-hidden ml-2 mb-2">
                  <!-- Header with search -->
                  <div class="shrink-0 bg-background border-b px-3 py-2 z-10">
                    <div class="flex items-center gap-2 mb-2">
                      <component :is="smartAlbumFieldIcons[field.key] || Folder" class="h-4 w-4 text-purple-500" />
                      <span class="text-sm font-medium">{{ field.label }}</span>
                      <span v-if="smartAlbumsData[field.key]" class="text-[10px] text-muted-foreground ml-auto">{{ smartAlbumsData[field.key].length }}项</span>
                    </div>
                    <div class="relative">
                      <input
                        v-model="smartAlbumSearch[field.key]"
                        type="text"
                        :placeholder="`搜索${field.label}...`"
                        class="w-full h-7 px-2 pl-7 text-xs rounded-md border bg-background focus:outline-none focus:ring-1 focus:ring-primary"
                      />
                      <Search class="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
                    </div>
                  </div>
                  <!-- Loading -->
                  <div v-if="smartAlbumsLoading[field.key]" class="flex items-center justify-center py-8">
                    <div class="flex items-center gap-2 text-xs text-muted-foreground">
                      <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
                      <span>加载中...</span>
                    </div>
                  </div>
                  <!-- Album list -->
                  <div v-else class="flex-1 overflow-y-auto min-h-0">
                    <div class="p-1">
                      <button
                        v-for="album in filteredSmartAlbums(field.key)"
                        :key="album.value"
                        class="w-full flex items-center justify-between px-3 py-1.5 text-xs rounded-md transition-colors hover:bg-secondary/50 text-left"
                        :class="activeSmartAlbum?.field === album.field && activeSmartAlbum?.value === album.value ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground'"
                        @click="handleSmartAlbumClick(album)"
                      >
                        <span class="truncate mr-2">{{ album.value }}</span>
                        <span class="text-[10px] opacity-60 shrink-0">{{ album.count }}</span>
                      </button>
                      <div v-if="filteredSmartAlbums(field.key).length === 0" class="px-3 py-4 text-xs text-muted-foreground text-center">
                        无匹配结果
                      </div>
                    </div>
                  </div>
                </PopoverContent>
              </Popover>

              <!-- Clear Smart Album Filter -->
              <div v-if="activeSmartAlbum" class="pt-1">
                <Button
                  variant="ghost"
                  size="sm"
                  class="w-full justify-start h-7 text-xs font-medium gap-2 hover:bg-destructive/10 hover:text-destructive transition-all rounded-md"
                  @click="clearSmartAlbumFilter"
                >
                  <X class="h-3 w-3" />
                  清除智能筛选
                </Button>
              </div>
            </div>
          </div>
        </div>
      </ScrollArea>
    </div>

    <!-- Footer - Always visible -->
    <div class="shrink-0 border-t bg-background" :class="collapsed ? 'p-2 space-y-2' : 'p-3'">
       <Popover v-model:open="showUtilityMenu">
          <PopoverTrigger as-child>
             <Button
                variant="outline"
                :class="collapsed ? 'w-full justify-center px-0 h-10' : 'w-full justify-start gap-2 h-9 px-3 text-sm font-medium'"
                :title="collapsed ? '工具菜单' : ''"
             >
                <Settings2 class="h-4 w-4 text-muted-foreground" :class="{ 'h-5 w-5': collapsed }" />
                <span v-if="!collapsed" class="flex-1 text-left">工具菜单</span>
                <ChevronRight v-if="!collapsed" class="h-4 w-4 text-muted-foreground" />
             </Button>
          </PopoverTrigger>
          <PopoverContent side="right" align="end" :side-offset="12" class="w-56 p-2 ml-2 mb-2">
             <div class="space-y-1">
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="openTrashManager">
                   <Trash2 class="h-4 w-4 text-muted-foreground" />
                   <span>回收站管理</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="openDocumentation">
                   <BookOpen class="h-4 w-4 text-muted-foreground" />
                   <span>使用文档</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="cleanEmptyFolders">
                   <Eraser class="h-4 w-4 text-muted-foreground" />
                   <span>清理空文件夹</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="$emit('organize-files')">
                   <FolderTree class="h-4 w-4 text-muted-foreground" />
                   <span>按日期整理文件</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="clearPreviewCache">
                   <Trash2 class="h-4 w-4 text-muted-foreground" />
                   <span>清空预览缓存</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="handleThemeToggle">
                   <Moon v-if="isDark" class="h-4 w-4 text-yellow-500" />
                   <Sun v-else class="h-4 w-4 text-orange-500" />
                   <span>{{ isDark ? '切换亮色模式' : '切换暗色模式' }}</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="openFavoriteGroups">
                   <Heart class="h-4 w-4 text-red-500" />
                   <span>收藏分组</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="openLauncher">
                   <TerminalSquare class="h-4 w-4 text-muted-foreground" />
                   <span>外部工具</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="openPromptTemplates">
                   <Bookmark class="h-4 w-4 text-amber-500" />
                   <span>提示词模板</span>
                </Button>
                <Button variant="ghost" class="w-full justify-start gap-2 h-9 px-3 text-sm" @click="openCustomRootManager">
                   <FolderSymlink class="h-4 w-4 text-muted-foreground" />
                   <span>管理目录</span>
                </Button>
             </div>
          </PopoverContent>
       </Popover>

       <Button 
          :variant="isSelectionMode ? 'default' : 'secondary'"
          :class="collapsed ? 'w-full justify-center px-0 h-10' : 'mt-2 w-full justify-start gap-2 h-9 px-3 text-sm font-medium'"
          @click="$emit('toggle-selection-mode')"
          :title="collapsed ? (isSelectionMode ? '退出批量' : '批量管理') : ''"
       >
          <CheckSquare class="h-4 w-4" :class="{'h-5 w-5': collapsed}" />
          <span v-if="!collapsed">{{ isSelectionMode ? '退出批量' : '批量管理' }}</span>
       </Button>
    </div>
    
    <TrashDialog 
      v-model:open="showTrashDialog" 
      @refresh="$emit('refresh-images')" 
    />
    <LauncherDialog 
      v-model:open="showLauncherDialog" 
    />
    <FavoriteGroupsDialog
      v-model:open="showFavoriteGroupsDialog"
      :groups="favoriteGroups"
      @change="$emit('favorite-group-change')"
    />
    <CustomRootDialog
      v-model:open="showCustomRootDialog"
      :custom-roots="customRoots"
      @change="$emit('custom-root-change')"
    />
    <PromptTemplateDialog
      v-model:open="showPromptTemplateDialog"
    />
  </aside>

  <!-- Hover Drawer for Collapsed Sidebar -->
  <Teleport to="body">
    <div 
        v-if="collapsed && hoveredRoot"
        class="fixed z-50 bg-popover text-popover-foreground border shadow-md rounded-r-md w-[200px] flex flex-col transition-all duration-200"
        :style="{ top: hoverTop + 'px', left: '60px', height: 'auto', maxHeight: 'calc(100vh - ' + hoverTop + 'px)' }"
        @mouseenter="cancelHoverLeave"
        @mouseleave="handleHoverLeave"
    >
        <div class="px-3 py-2 border-b bg-muted/20 font-medium text-sm flex items-center gap-2">
            <component :is="getIcon(hoveredRoot)" class="h-4 w-4 opacity-70" />
            <span>{{ hoveredRoot.name.replace(/[\u{1F300}-\u{1F9FF}]/gu, '').replace(/[^\w\u4e00-\u9fa5\s]/g, '').trim() }}</span>
        </div>
        
        <ScrollArea class="flex-1 min-h-0 py-1">
             <div class="px-1 space-y-1">
                 <!-- Level 1: Years/Categories -->
                 <div v-for="l1 in (hoveredRoot.subs || hoveredRoot.children)" :key="l1.id" class="space-y-1">
                     <button
                        @click.stop="handleDrawerClick(l1.id)"
                        :class="cn(
                          'w-full flex items-center justify-between px-3 py-1.5 text-sm rounded-md transition-colors hover:bg-secondary/50 hover:text-foreground text-left',
                          activeSub === l1.id ? 'bg-secondary/80 text-primary font-medium' : 'text-foreground/80'
                        )"
                      >
                        <span>{{ l1.name }}</span>
                        <span class="text-xs opacity-50">{{ getRecursiveCount(l1) }}</span>
                     </button>

                     <!-- Level 2 -->
                     <div v-if="l1.children && l1.children.length > 0" class="pl-4 border-l ml-3 space-y-1 my-1">
                        <div v-for="l2 in l1.children" :key="l2.id" class="space-y-1">
                             <button
                                @click.stop="handleDrawerClick(l2.id)"
                                :class="cn(
                                  'w-full flex items-center justify-between px-3 py-1 text-xs font-semibold rounded-md transition-colors uppercase tracking-wider hover:bg-secondary/50 hover:text-foreground text-left',
                                  activeSub === l2.id ? 'bg-secondary/80 text-primary' : 'text-muted-foreground'
                                )"
                             >
                                <span>{{ l2.displayName || formatFolderName(l2.name) }}</span>
                                <span class="text-xs opacity-50 font-normal">{{ getRecursiveCount(l2) }}</span>
                             </button>

                             <!-- Level 3 -->
                             <div v-if="isExpanded(l2.id) && l2.children && l2.children.length > 0" class="space-y-0.5 pl-2 my-0.5">
                                 <div v-for="l3 in l2.children" :key="l3.id" class="space-y-0.5">
                                     <button
                                        @click.stop="handleDrawerClick(l3.id)"
                                        :class="cn(
                                          'w-full flex items-center justify-between px-3 py-1 text-xs rounded-md transition-colors hover:bg-secondary/50 hover:text-foreground text-left',
                                          activeSub === l3.id ? 'bg-secondary/80 text-primary font-medium' : 'text-muted-foreground'
                                        )"
                                      >
                                          <span>{{ l3.displayName || formatFolderName(l3.name) }}</span>
                                          <span class="text-xs opacity-70">{{ getRecursiveCount(l3) }}</span>
                                      </button>
                                       
                                      <!-- Level 4: Leaf Folders -->
                                      <div v-if="isExpanded(l3.id) && l3.children && l3.children.length > 0" class="space-y-0.5 pl-2 my-0.5">
                                          <button
                                             v-for="leaf in l3.children"
                                             :key="leaf.id"
                                             @click.stop="handleDrawerClick(leaf.id)"
                                             :class="cn(
                                               'w-full flex items-center justify-between px-3 py-1 text-xs rounded-md transition-colors hover:bg-secondary/50 hover:text-foreground text-left',
                                               activeSub === leaf.id ? 'bg-secondary/80 text-primary font-medium' : 'text-muted-foreground'
                                             )"
                                           >
                                               <span class="truncate">{{ leaf.displayName || leaf.name }}</span>
                                               <span class="text-xs opacity-70">{{ leaf.images?.length || 0 }}</span>
                                           </button>
                                      </div>
                                 </div>
                             </div>
                        </div>
                     </div>
                 </div>
            </div>
        </ScrollArea>
    </div>
  </Teleport>
</template>
