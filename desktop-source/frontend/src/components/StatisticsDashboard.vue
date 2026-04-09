<script setup>

import { ref, onMounted, computed, onUnmounted, watch } from 'vue'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ImageIcon, HardDrive, RefreshCw, TrendingUp } from 'lucide-vue-next'
import { useImages } from '@/composables/useImages'
import * as App from '@/api'

const stats = ref(null)
const loading = ref(false)
const viewMode = ref(localStorage.getItem('statsViewMode') || 'month')

watch(viewMode, (newVal) => {
    localStorage.setItem('statsViewMode', newVal)
    loadStats()
})
let pollingInterval = null

const loadStats = async () => {
    loading.value = true
    try {
        const data = await App.GetStatistics(viewMode.value)
        stats.value = data
    } catch (e) {
        console.error('Failed to load statistics:', e)
    } finally {
        loading.value = false
    }
}

onMounted(() => {
    loadStats()
    // Auto-refresh every 5 seconds to stay in sync with new images
    pollingInterval = setInterval(() => {
        loadStats()
    }, 5000)
})

onUnmounted(() => {
    if (pollingInterval) {
        clearInterval(pollingInterval)
    }
})

const formatSize = (bytes) => {
    if (!bytes) return '0 B'
    const gb = bytes / (1024 * 1024 * 1024)
    if (gb >= 1) return `${gb.toFixed(2)} GB`
    const mb = bytes / (1024 * 1024)
    return `${mb.toFixed(2)} MB`
}

const sortedDates = computed(() => {
    if (!stats.value?.byDate) return []
    const rawData = Object.entries(stats.value.byDate)
        .sort((a, b) => a[0].localeCompare(b[0]))
    
    // Even if no data, we want to show range if possible, or at least today?
    // If absolutely no data, just return empty or maybe just today?
    // Let's stick to existing logic but extend END date.

    if (rawData.length === 0) return []
    
    const start = rawData[0][0]
    // Use Today as the end date bounds, or the last image date, whichever is later.
    // Actually, usually we want "Today" to be the right-most point for "Realtime" feel.
    const today = new Date()
    let endStr = ""
    
    if (viewMode.value === 'month') {
        const y = today.getFullYear()
        const m = (today.getMonth() + 1).toString().padStart(2, '0')
        endStr = `${y}-${m}`
    } else if (viewMode.value === 'year') {
        endStr = today.getFullYear().toString()
    } else {
         const y = today.getFullYear()
         const m = (today.getMonth() + 1).toString().padStart(2, '0')
         const d = today.getDate().toString().padStart(2, '0')
         endStr = `${y}-${m}-${d}`
    }

    // Compare rawData last date with today.
    const rawEnd = rawData[rawData.length - 1][0]
    const end = rawEnd > endStr ? rawEnd : endStr
    
    const result = []

    if (viewMode.value === 'month') {
        // Month format: YYYY-MM
        let current = new Date(start + "-01")
        const endMonth = new Date(end + "-01")
        while (current <= endMonth) {
            const y = current.getFullYear()
            const m = (current.getMonth() + 1).toString().padStart(2, '0')
            const key = `${y}-${m}`
            result.push([key, stats.value.byDate[key] || 0])
            current.setMonth(current.getMonth() + 1)
        }
    } else if (viewMode.value === 'year') {
        // Year format: YYYY
        let currentYear = parseInt(start)
        const endYear = parseInt(end)
        while (currentYear <= endYear) {
            const key = currentYear.toString()
            result.push([key, stats.value.byDate[key] || 0])
            currentYear++
        }
    } else {
        // Day format: YYYY-MM-DD
        let current = new Date(start)
        const endDate = new Date(end)
        while (current <= endDate) {
            const key = current.toISOString().split('T')[0]
            result.push([key, stats.value.byDate[key] || 0])
            current.setDate(current.getDate() + 1)
        }
    }
    return result
})

const maxDateCount = computed(() => {
    if (sortedDates.value.length === 0) return 0
    return Math.max(...sortedDates.value.map(([_, count]) => count))
})

// Helper function to create smooth curve points using Catmull-Rom
const createSmoothPath = (points, tension = 0.5) => {
    if (points.length < 2) return ''
    
    let path = `M${points[0].x},${points[0].y}`
    
    for (let i = 0; i < points.length - 1; i++) {
        const p0 = points[Math.max(i - 1, 0)]
        const p1 = points[i]
        const p2 = points[i + 1]
        const p3 = points[Math.min(i + 2, points.length - 1)]
        
        // Calculate control points for cubic bezier
        const cp1x = p1.x + (p2.x - p0.x) / 6 * tension
        const cp1y = p1.y + (p2.y - p0.y) / 6 * tension
        const cp2x = p2.x - (p3.x - p1.x) / 6 * tension
        const cp2y = p2.y - (p3.y - p1.y) / 6 * tension
        
        path += ` C${cp1x},${cp1y} ${cp2x},${cp2y} ${p2.x},${p2.y}`
    }
    
    return path
}

// Path calculations for charts
const trendPath = computed(() => {
    if (sortedDates.value.length < 2) return ''
    const points = sortedDates.value.map(([_, count], index) => {
        const x = (index / (sortedDates.value.length - 1)) * 100
        const y = 100 - (count / yAxisMax.value) * 100
        return { x, y }
    })
    
    const smoothPath = createSmoothPath(points)
    return `${smoothPath} L100,100 L0,100 Z`
})

const linePath = computed(() => {
    if (sortedDates.value.length < 2) return ''
    const points = sortedDates.value.map(([_, count], index) => {
        const x = (index / (sortedDates.value.length - 1)) * 100
        const y = 100 - (count / yAxisMax.value) * 100
        return { x, y }
    })
    return createSmoothPath(points)
})

const yAxisMax = computed(() => {
    const max = maxDateCount.value
    if (max === 0) return 0
    // Add 20% padding to the max value for better visual spacing
    return Math.ceil(max * 1.2)
})

const yTicks = computed(() => {
    const max = yAxisMax.value
    if (max === 0) return [0, 0, 0, 0, 0]
    return [
        Math.round(max),
        Math.round(max * 0.75),
        Math.round(max * 0.5),
        Math.round(max * 0.25),
        0
    ]
})
</script>

<template>
    <div class="flex-1 p-8 space-y-8 overflow-y-auto bg-background/50 select-none relative">
        <!-- Dashboard Animated Background -->
        <div class="absolute inset-0 z-0 pointer-events-none overflow-hidden">
            <div class="absolute -top-[20%] -left-[10%] w-[50%] h-[50%] bg-primary/10 rounded-full blur-[120px] animate-pulse"></div>
            <div class="absolute -bottom-[10%] -right-[10%] w-[40%] h-[40%] bg-blue-500/10 rounded-full blur-[100px] animate-pulse" style="animation-delay: 2s"></div>
            <div class="absolute top-[30%] left-[40%] w-[30%] h-[30%] bg-purple-500/5 rounded-full blur-[80px] animate-pulse" style="animation-delay: 4s"></div>
        </div>

        <!-- Header -->
        <div class="flex items-center justify-between relative z-10">
            <div>
                <h2 class="text-3xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-foreground to-foreground/70">数据视界</h2>
                <p class="text-muted-foreground mt-1">探索您的创意资产数据</p>
            </div>
            <div class="flex items-center space-x-2">
                <Button @click="loadStats" variant="outline" :disabled="loading" class="group backdrop-blur-md">
                    <RefreshCw class="w-4 h-4 mr-2 group-hover:animate-spin" :class="{ 'animate-spin': loading }" />
                    {{ loading ? '刷新中...' : '刷新数据' }}
                </Button>
            </div>
        </div>

        <!-- KPI Cards -->
        <div class="grid gap-6 md:grid-cols-3 relative z-10">
            <!-- Added Today Card -->
            <Card class="relative overflow-hidden border bg-card/60 backdrop-blur-md text-card-foreground shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 group">
                 <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                    <div>
                        <p class="text-sm font-medium text-muted-foreground">今日新增</p>
                        <h3 class="text-2xl font-bold mt-2 tabular-nums">
                            {{ stats?.todayCount?.toLocaleString() || 0 }} 张
                        </h3>
                    </div>
                    <div class="h-12 w-12 rounded-full bg-green-500/10 flex items-center justify-center">
                        <ImageIcon class="h-6 w-6 text-green-500" />
                    </div>
                </div>
                 <div class="px-6 pb-6">
                    <div class="text-xs text-muted-foreground mt-1">
                        今日新加入图片
                    </div>
                </div>
            </Card>

            <!-- Total Images Card -->
            <Card class="relative overflow-hidden border bg-card/60 backdrop-blur-md text-card-foreground shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 group">
                <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                     <div>
                        <p class="text-sm font-medium text-muted-foreground">总图片数</p>
                        <h3 class="text-2xl font-bold mt-2 tabular-nums">{{ stats?.totalCount?.toLocaleString() || 0 }} 张</h3>
                    </div>
                     <div class="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center">
                        <ImageIcon class="h-6 w-6 text-primary" />
                    </div>
                </div>
                 <div class="px-6 pb-6">
                    <div class="text-xs text-muted-foreground mt-1">
                        当前库中所有图片
                    </div>
                </div>
                <!-- Subtle gradient accent -->

            </Card>

            <!-- Total Storage Card -->
            <Card class="relative overflow-hidden border bg-card/60 backdrop-blur-md text-card-foreground shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-300 group">
                 <div class="p-6 flex flex-row items-center justify-between space-y-0 pb-2">
                    <div>
                        <p class="text-sm font-medium text-muted-foreground">占用空间</p>
                         <h3 class="text-2xl font-bold mt-2 tabular-nums">
                            {{ formatSize(stats?.totalSize) }}
                        </h3>
                    </div>
                    <div class="h-12 w-12 rounded-full bg-blue-500/10 flex items-center justify-center">
                        <HardDrive class="h-6 w-6 text-blue-500" />
                    </div>
                </div>
                 <div class="px-6 pb-6">
                    <div class="text-xs text-muted-foreground mt-1">
                        所有文件总大小
                    </div>
                </div>
                 <!-- Subtle gradient accent -->

            </Card>
        </div>

        <!-- Charts Section -->
        <div class="space-y-6 relative z-10">
            <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                    <TrendingUp class="h-5 w-5 text-primary" />
                    <h3 class="text-lg font-semibold">趋势分析</h3>
                </div>
                <div class="flex items-center bg-muted  rounded-lg p-1">
                    <button 
                        @click="viewMode = 'year'"
                        class="px-3 py-1 text-xs font-medium rounded-md transition-all"
                        :class="viewMode === 'year' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                    >
                        按年
                    </button>
                    <button 
                        @click="viewMode = 'month'"
                        class="px-3 py-1 text-xs font-medium rounded-md transition-all"
                        :class="viewMode === 'month' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                    >
                        按月
                    </button>
                    <button 
                        @click="viewMode = 'day'"
                        class="px-3 py-1 text-xs font-medium rounded-md transition-all"
                        :class="viewMode === 'day' ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'"
                    >
                        按日
                    </button>
                </div>
            </div>
            
            <div class="grid grid-cols-1 gap-6">
                <!-- Bar Chart -->
                <Card class="p-6 border bg-card/60 backdrop-blur-md text-card-foreground shadow-sm hover:shadow-lg transition-all duration-500">
                    <div class="mb-6">
                         <h4 class="text-base font-medium">数据分布</h4>
                         <p class="text-xs text-muted-foreground">{{ viewMode === 'month' ? '每月' : (viewMode === 'year' ? '每年' : '每日') }}新增图片数量统计</p>
                    </div>
                    
                     <div class="h-[200px] w-full relative pt-4 pl-8">
                         <!-- Y-Axis -->
                        <div class="absolute left-0 top-4 bottom-0 w-8 flex flex-col justify-between text-[10px] text-muted-foreground py-0 text-right pr-2 pointer-events-none z-10">
                            <span v-for="tick in yTicks" :key="tick">{{ tick }}</span>
                        </div>

                        <!-- Grid Lines -->
                        <div class="absolute inset-0 left-8 flex flex-col justify-between pointer-events-none">
                            <div v-for="i in 5" :key="i" class="border-t border-muted-foreground/20 w-full"></div>
                        </div>

                        <!-- Bar Chart Columns with Full Height Hover Areas -->
                        <div class="absolute inset-0 left-8 top-4 flex">
                            <div 
                                v-for="[date, count] in sortedDates" 
                                :key="date" 
                                class="flex-1 relative group cursor-pointer"
                            >
                                <!-- Vertical Hover Line -->
                                <div class="absolute left-1/2 -translate-x-1/2 top-0 bottom-0 w-px bg-primary/30 opacity-0 group-hover:opacity-100 transition-opacity"></div>

                                <!-- Visible Bar -->
                                <div 
                                    class="absolute bottom-0 inset-x-0 mx-1 bg-[#6BB6FF] rounded-t-md transition-colors duration-200 group-hover:bg-[#5CABFF]"
                                    :style="{ height: `${(count / yAxisMax) * 100}%` }"
                                ></div>

                                <!-- Tooltip -->
                                <div class="opacity-0 group-hover:opacity-100 absolute -top-10 left-1/2 -translate-x-1/2 bg-popover text-popover-foreground text-xs px-2 py-1 rounded shadow-lg whitespace-nowrap z-20 border transition-opacity pointer-events-none">
                                    {{ date }}: {{ count }}
                                </div>
                            </div>
                        </div>
                    </div>
                    <div class="flex justify-between mt-4 ml-8 text-xs text-muted-foreground">
                        <span>{{ sortedDates[0]?.[0] }}</span>
                        <span>{{ sortedDates[sortedDates.length - 1]?.[0] }}</span>
                    </div>
                </Card>

                <!-- Line Chart -->
                <Card class="p-6 border bg-card/60 backdrop-blur-md text-card-foreground shadow-sm hover:shadow-lg transition-all duration-500">
                    <div class="mb-6">
                         <h4 class="text-base font-medium">增长曲线</h4>
                         <p class="text-xs text-muted-foreground">图库容量趋势</p>
                    </div>

                    <div class="h-[200px] w-full relative pl-8">
                         <!-- Y-Axis -->
                        <div class="absolute left-0 top-0 bottom-0 w-8 flex flex-col justify-between text-[10px] text-muted-foreground py-0 text-right pr-2 pointer-events-none z-10">
                            <span v-for="tick in yTicks" :key="tick">{{ tick }}</span>
                        </div>

                        <!-- Grid Lines -->
                        <div class="absolute inset-0 left-8 flex flex-col justify-between pointer-events-none">
                            <div v-for="i in 5" :key="i" class="border-t border-muted-foreground/20 w-full"></div>
                        </div>

                        <!-- SVG Chart -->
                        <svg class="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
                            <defs>
                                <linearGradient id="chartGradient" x1="0" x2="0" y1="0" y2="1">
                                    <stop offset="0%" stop-color="#5CABFF" stop-opacity="0.5" />
                                    <stop offset="100%" stop-color="#5CABFF" stop-opacity="0.1" />
                                </linearGradient>
                            </defs>
                            
                            <path 
                                :d="trendPath" 
                                fill="url(#chartGradient)" 
                                class="transition-all duration-500"
                            />
                            <path 
                                :d="linePath" 
                                fill="none" 
                                stroke="#5CABFF" 
                                stroke-width="0.5" 
                                vector-effect="non-scaling-stroke"
                                stroke-linejoin="round"
                                stroke-linecap="round"
                            />
                        </svg>

                        <!-- Interactive Hover Columns -->
                        <div class="absolute inset-0 flex">
                            <div 
                                v-for="([date, count], i) in sortedDates" 
                                :key="date" 
                                class="flex-1 relative group cursor-pointer"
                            >
                                <!-- Vertical Hover Line -->
                                <div class="absolute left-1/2 -translate-x-1/2 top-0 bottom-0 w-px bg-primary/30 opacity-0 group-hover:opacity-100 transition-opacity"></div>

                                <!-- Data Point Dot -->
                                <div 
                                    class="absolute left-1/2 w-2.5 h-2.5 bg-[#5CABFF] border-2 border-background rounded-full shadow-md opacity-0 group-hover:opacity-100 transition-opacity z-10"
                                    :style="{ bottom: `calc(${(count / yAxisMax) * 100}% - 5px)`, left: '50%', transform: 'translateX(-50%)' }"
                                ></div>

                                <!-- Tooltip (Smart Positioning) -->
                                <div 
                                    class="absolute left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 bg-popover text-popover-foreground text-xs px-2 py-1 rounded shadow-lg whitespace-nowrap z-20 border transition-opacity pointer-events-none"
                                    :style="{ 
                                        bottom: (count / yAxisMax) > 0.7 ? `${(count / yAxisMax) * 100 + 5}%` : 'auto',
                                        top: (count / yAxisMax) <= 0.7 ? `${100 - (count / yAxisMax) * 100 + 5}%` : 'auto'
                                    }"
                                >
                                    {{ date }}: {{ count }}
                                </div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="mt-4 flex justify-between text-xs text-muted-foreground pl-8">
                         <div v-for="[date, _] in sortedDates.slice(-3)" :key="date">
                             {{ date }}
                        </div>
                    </div>
                </Card>
            </div>
        </div>
    </div>
</template>
