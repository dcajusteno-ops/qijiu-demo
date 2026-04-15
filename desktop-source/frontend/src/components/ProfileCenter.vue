<script setup>
import { computed, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { isDark, toggleTheme } from '@/theme'
import * as App from '@/api'
import {
  BarChart3,
  BookOpen,
  Camera,
  ChevronDown,
  Copy,
  FolderOpen,
  Heart,
  ImageIcon,
  LayoutDashboard,
  Link2,
  MapPin,
  Moon,
  Save,
  Sparkles,
  Sun,
  Tag,
  Target,
  Trash2,
  TrendingUp,
  UserRound,
  Zap,
} from 'lucide-vue-next'

const emit = defineEmits(['navigate'])

const profileControlClass = 'w-full rounded-2xl border border-border/80 bg-background/90 px-4 py-3 text-sm text-foreground shadow-none outline-none transition-[border-color,box-shadow,background-color] focus:border-ring focus:ring-1 focus:ring-ring/60'
const helperCardClass = 'rounded-[24px] border border-border bg-background/90 p-5'

const defaultProfile = () => ({
  displayName: '创作者',
  headline: '把灵感整理成稳定、清爽的作品集',
  bio: '这里保存你的出图节奏、偏好设置和常用入口，让工作流保持顺手。',
  location: '',
  website: '',
  dailyGoal: 12,
  preferredStartPage: 'dashboard',
  imagePath: '',
})

const startPageOptions = [
  { value: 'dashboard', label: '工作台总览' },
  { value: 'statistics', label: '数据视界' },
  { value: 'profile', label: '个人中心' },
  { value: 'favorites', label: '收藏夹' },
  { value: 'documentation', label: '使用文档' },
  { value: 'output', label: 'output 图库' },
]

const quickActions = [
  { key: 'dashboard', label: '工作台总览', icon: LayoutDashboard, desc: '回到首页看最新作品' },
  { key: 'statistics', label: '数据视界', icon: BarChart3, desc: '查看趋势和时间线' },
  { key: 'favorites', label: '收藏夹', icon: Heart, desc: '整理你的精选作品' },
  { key: 'output', label: '打开图库', icon: FolderOpen, desc: '继续浏览当前图库' },
  { key: 'documentation', label: '使用文档', icon: BookOpen, desc: '快速查看功能说明' },
]

const loading = ref(true)
const saving = ref(false)
const uploadingImage = ref(false)
const profile = ref(defaultProfile())
const savedProfile = ref(defaultProfile())

const stats = ref({
  todayCount: 0,
  totalCount: 0,
  totalSize: 0,
  currentMonthCount: 0,
})

const activitySummary = ref({
  activeDays: 0,
  bestDayLabel: '暂无记录',
  bestDayCount: 0,
  averagePerActiveDay: 0,
})

const tagCount = ref(0)
const favoriteGroupCount = ref(0)
const enabledShortcutCount = ref(0)

const normalizeProfile = (value = {}) => {
  const fallback = defaultProfile()
  return {
    displayName: (value.displayName || fallback.displayName).trim() || fallback.displayName,
    headline: (value.headline || fallback.headline).trim() || fallback.headline,
    bio: (value.bio || fallback.bio).trim() || fallback.bio,
    location: (value.location || '').trim(),
    website: (value.website || '').trim(),
    dailyGoal: Math.min(999, Math.max(1, Number(value.dailyGoal) || fallback.dailyGoal)),
    preferredStartPage: startPageOptions.some((option) => option.value === value.preferredStartPage)
      ? value.preferredStartPage
      : fallback.preferredStartPage,
    imagePath: (value.imagePath || '').trim(),
  }
}

const initials = computed(() => {
  const plain = profile.value.displayName.replace(/\s+/g, '')
  return plain.slice(0, 2).toUpperCase() || '创'
})

const completionRatio = computed(() => {
  if (!profile.value.dailyGoal) return 0
  return Math.min(1, stats.value.todayCount / profile.value.dailyGoal)
})

const completionLabel = computed(() => {
  if (stats.value.todayCount >= profile.value.dailyGoal) {
    return `今天已经完成目标，多出 ${stats.value.todayCount - profile.value.dailyGoal} 张`
  }
  return `距离今日目标还差 ${profile.value.dailyGoal - stats.value.todayCount} 张`
})

const preferredStartPageLabel = computed(() =>
  startPageOptions.find((option) => option.value === profile.value.preferredStartPage)?.label || '工作台总览',
)

const dirty = computed(() => JSON.stringify(profile.value) !== JSON.stringify(savedProfile.value))

const highlightStats = computed(() => [
  {
    key: 'today',
    title: '今日出图',
    value: stats.value.todayCount,
    caption: completionLabel.value,
    icon: Sparkles,
  },
  {
    key: 'month',
    title: '本月产出',
    value: stats.value.currentMonthCount,
    caption: '当前月度累计的出图数量',
    icon: Target,
  },
])

const statCards = computed(() => [
  {
    key: 'favorites',
    title: '收藏分组',
    value: favoriteGroupCount.value,
    caption: '当前已经建立的精选分组',
    icon: Heart,
  },
  {
    key: 'tags',
    title: '标签数量',
    value: tagCount.value,
    caption: '可用于筛选和管理图片',
    icon: Tag,
  },
])

const insightCards = computed(() => [
  {
    key: 'activeDays',
    title: '活跃天数',
    value: activitySummary.value.activeDays,
    caption: '至少有一张作品产出的日期',
    icon: Zap,
  },
  {
    key: 'bestDay',
    title: '最高峰值',
    value: activitySummary.value.bestDayCount,
    caption: activitySummary.value.bestDayLabel,
    icon: TrendingUp,
  },
  {
    key: 'average',
    title: '日均产出',
    value: activitySummary.value.averagePerActiveDay,
    caption: '按活跃日计算的平均产出',
    icon: BarChart3,
  },
])

const computeActivitySummary = (byDate = {}) => {
  const entries = Object.entries(byDate || {})
    .map(([date, count]) => ({ date, count }))
    .filter((item) => item.count > 0)
    .sort((a, b) => a.date.localeCompare(b.date))

  if (!entries.length) {
    activitySummary.value = {
      activeDays: 0,
      bestDayLabel: '暂无记录',
      bestDayCount: 0,
      averagePerActiveDay: 0,
    }
    return
  }

  const bestDay = [...entries].sort((a, b) => b.count - a.count)[0]
  const total = entries.reduce((sum, item) => sum + item.count, 0)

  activitySummary.value = {
    activeDays: entries.length,
    bestDayLabel: bestDay.date,
    bestDayCount: bestDay.count,
    averagePerActiveDay: Number((total / entries.length).toFixed(1)),
  }
}

const loadProfile = async () => {
  loading.value = true
  try {
    const now = new Date()
    const monthKey = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    const [profileData, dayStats, monthStats, favoriteGroups, tags, shortcutSettings] = await Promise.all([
      App.GetUserProfile(),
      App.GetStatistics('day'),
      App.GetStatistics('month'),
      App.GetFavoriteGroups(),
      App.GetTags(),
      App.GetShortcutSettings(),
    ])

    const nextProfile = normalizeProfile(profileData)
    profile.value = nextProfile
    savedProfile.value = { ...nextProfile }

    stats.value = {
      todayCount: dayStats?.todayCount || 0,
      totalCount: dayStats?.totalCount || 0,
      totalSize: dayStats?.totalSize || 0,
      currentMonthCount: monthStats?.byDate?.[monthKey] || 0,
    }

    computeActivitySummary(dayStats?.byDate || {})
    favoriteGroupCount.value = favoriteGroups?.length || 0
    tagCount.value = tags?.length || 0
    enabledShortcutCount.value = Object.values(shortcutSettings?.bindings || {}).filter(Boolean).length
  } catch (error) {
    console.error('Failed to load profile center:', error)
    toast.error(`个人中心加载失败: ${error}`)
  } finally {
    loading.value = false
  }
}

const saveProfile = async () => {
  saving.value = true
  try {
    const saved = await App.SaveUserProfile(normalizeProfile(profile.value))
    const nextProfile = normalizeProfile(saved)
    profile.value = nextProfile
    savedProfile.value = { ...nextProfile }
    toast.success('个人资料已保存')
  } catch (error) {
    console.error('Failed to save profile:', error)
    toast.error(`保存失败: ${error}`)
  } finally {
    saving.value = false
  }
}

const resetProfile = () => {
  profile.value = { ...savedProfile.value }
}

const copyWebsite = async () => {
  const website = profile.value.website.trim()
  if (!website) return
  try {
    await App.CopyText(website)
    toast.success('链接已复制')
  } catch (error) {
    console.error('Failed to copy website:', error)
    toast.error('复制失败')
  }
}

const uploadProfileImage = async () => {
  uploadingImage.value = true
  try {
    const saved = await App.SelectUserProfileImage()
    const nextProfile = normalizeProfile(saved)
    profile.value = nextProfile
    savedProfile.value = { ...nextProfile }
    toast.success('个人中心图片已更新')
  } catch (error) {
    console.error('Failed to upload profile image:', error)
    toast.error(`上传失败: ${error}`)
  } finally {
    uploadingImage.value = false
  }
}

const clearProfileImage = async () => {
  uploadingImage.value = true
  try {
    const saved = await App.ClearUserProfileImage()
    const nextProfile = normalizeProfile(saved)
    profile.value = nextProfile
    savedProfile.value = { ...nextProfile }
    toast.success('个人中心图片已清除')
  } catch (error) {
    console.error('Failed to clear profile image:', error)
    toast.error(`清除失败: ${error}`)
  } finally {
    uploadingImage.value = false
  }
}

const handleThemeToggle = (event) => {
  toggleTheme(event)
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
}

onMounted(loadProfile)
</script>

<template>
  <div class="h-full overflow-y-auto bg-background">
    <div class="mx-auto flex min-h-full w-full max-w-7xl flex-col gap-6 px-6 py-8">
      <section class="grid items-start gap-6 xl:grid-cols-[minmax(0,1fr)_340px]">
        <div class="grid gap-6">
          <Card class="overflow-hidden rounded-[30px] border-border/70 bg-card/95 shadow-none">
            <div class="profile-hero grid gap-6 p-8 lg:grid-cols-[minmax(0,1fr)_280px]">
              <div class="min-w-0 space-y-5">
                <div class="flex items-start gap-5">
                  <div class="relative h-24 w-24 shrink-0 overflow-hidden rounded-[30px] border border-foreground/10 bg-foreground/95 shadow-sm">
                    <img
                      v-if="profile.imagePath"
                      :src="profile.imagePath"
                      alt="个人中心图片"
                      class="h-full w-full object-cover"
                    />
                    <div v-else class="flex h-full w-full items-center justify-center text-3xl font-semibold text-background">
                      {{ initials }}
                    </div>
                  </div>

                  <div class="min-w-0 space-y-3">
                    <Badge variant="outline" class="rounded-full px-3 py-1 text-[11px]">个人中心</Badge>
                    <div class="space-y-2">
                      <h1 class="truncate text-3xl font-semibold tracking-tight text-foreground">
                        {{ profile.displayName }}
                      </h1>
                      <p class="text-sm text-muted-foreground">{{ profile.headline }}</p>
                    </div>
                  </div>
                </div>

                <p class="max-w-3xl text-sm leading-7 text-muted-foreground">
                  {{ profile.bio }}
                </p>

                <div class="flex flex-wrap gap-2">
                  <span class="inline-flex items-center gap-2 rounded-full border border-border bg-background/80 px-3 py-1.5 text-xs text-muted-foreground">
                    <MapPin class="h-3.5 w-3.5" />
                    {{ profile.location || '未设置所在地' }}
                  </span>
                  <span class="inline-flex items-center gap-2 rounded-full border border-border bg-background/80 px-3 py-1.5 text-xs text-muted-foreground">
                    <Target class="h-3.5 w-3.5" />
                    日目标 {{ profile.dailyGoal }} 张
                  </span>
                  <span class="inline-flex items-center gap-2 rounded-full border border-border bg-background/80 px-3 py-1.5 text-xs text-muted-foreground">
                    <LayoutDashboard class="h-3.5 w-3.5" />
                    默认进入 {{ preferredStartPageLabel }}
                  </span>
                </div>
              </div>

              <div class="grid gap-4 self-start">
                <div class="rounded-[26px] border border-border bg-background/85 p-5">
                  <p class="text-xs uppercase tracking-[0.18em] text-muted-foreground/80">今日完成度</p>
                  <div class="mt-3 flex items-end gap-2">
                    <span class="text-4xl font-semibold tracking-tight text-foreground">
                      {{ Math.round(completionRatio * 100) }}%
                    </span>
                    <span class="pb-1 text-xs text-muted-foreground">{{ stats.todayCount }}/{{ profile.dailyGoal }}</span>
                  </div>
                  <div class="mt-4 h-2 overflow-hidden rounded-full bg-muted">
                    <div class="h-full rounded-full bg-foreground transition-all duration-300" :style="{ width: `${completionRatio * 100}%` }" />
                  </div>
                  <p class="mt-3 text-xs leading-5 text-muted-foreground">{{ completionLabel }}</p>
                </div>

                <div class="rounded-[26px] border border-border bg-background/85 p-5">
                  <p class="text-xs uppercase tracking-[0.18em] text-muted-foreground/80">创作速览</p>
                  <div class="mt-4 grid gap-3">
                    <div class="flex items-center justify-between">
                      <span class="text-sm text-muted-foreground">总作品</span>
                      <span class="text-lg font-semibold text-foreground">{{ stats.totalCount }}</span>
                    </div>
                    <div class="flex items-center justify-between">
                      <span class="text-sm text-muted-foreground">占用空间</span>
                      <span class="text-lg font-semibold text-foreground">{{ formatSize(stats.totalSize) }}</span>
                    </div>
                    <div class="flex items-center justify-between">
                      <span class="text-sm text-muted-foreground">已启用快捷键</span>
                      <span class="text-lg font-semibold text-foreground">{{ enabledShortcutCount }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Card>

          <div class="grid gap-4 sm:grid-cols-2 2xl:grid-cols-4">
            <article
              v-for="item in [...highlightStats, ...statCards]"
              :key="item.key"
              class="rounded-[26px] border border-border bg-card p-5 shadow-none"
            >
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-[13px] text-muted-foreground">{{ item.title }}</p>
                  <p class="mt-2 text-[1.9rem] font-semibold tracking-tight text-foreground">{{ item.value }}</p>
                </div>
                <component :is="item.icon" class="h-5 w-5 text-muted-foreground" />
              </div>
              <p class="mt-4 text-[11px] leading-5 text-muted-foreground">{{ item.caption }}</p>
            </article>
          </div>

          <div class="grid gap-4 md:grid-cols-3">
            <article v-for="item in insightCards" :key="item.key" class="rounded-[26px] border border-border bg-card p-6 shadow-none">
              <div class="flex items-start justify-between gap-3">
                <div>
                  <p class="text-sm text-muted-foreground">{{ item.title }}</p>
                  <p class="mt-2 text-3xl font-semibold tracking-tight text-foreground">{{ item.value }}</p>
                </div>
                <component :is="item.icon" class="h-5 w-5 text-muted-foreground" />
              </div>
              <p class="mt-4 text-xs leading-6 text-muted-foreground">{{ item.caption }}</p>
            </article>
          </div>

          <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
          <CardHeader class="space-y-2">
            <CardTitle class="text-lg">编辑资料</CardTitle>
            <CardDescription>把常用字段集中在一个区块里，样式也和项目现有控件保持一致。</CardDescription>
          </CardHeader>
          <CardContent v-if="!loading" class="grid gap-6">
            <div class="grid gap-6 md:grid-cols-2">
              <div class="grid gap-2">
                <Label for="profile-name">昵称</Label>
                <Input id="profile-name" v-model="profile.displayName" maxlength="24" placeholder="例如：阿七 / Studio Q" class="h-11 rounded-2xl border-border/80 bg-background/90 shadow-none" />
              </div>
              <div class="grid gap-2">
                <Label for="profile-location">所在地</Label>
                <Input id="profile-location" v-model="profile.location" maxlength="24" placeholder="例如：上海 / Remote" class="h-11 rounded-2xl border-border/80 bg-background/90 shadow-none" />
              </div>
            </div>

            <div class="grid gap-2">
              <Label for="profile-headline">一句话简介</Label>
              <Input id="profile-headline" v-model="profile.headline" maxlength="60" placeholder="写一句你希望首页看到的话" class="h-11 rounded-2xl border-border/80 bg-background/90 shadow-none" />
            </div>

            <div class="grid gap-2">
              <Label for="profile-bio">个人说明</Label>
              <textarea
                id="profile-bio"
                v-model="profile.bio"
                rows="5"
                maxlength="180"
                :class="`${profileControlClass} min-h-[132px] resize-none leading-6`"
                placeholder="写一点你的工作流偏好、主题方向，或者你希望这套图库怎么服务你。"
              />
            </div>

            <div class="grid gap-6 md:grid-cols-[minmax(0,1fr)_220px]">
              <div class="grid gap-2">
                <Label for="profile-website">主页 / 链接</Label>
                <div class="flex gap-2">
                  <Input id="profile-website" v-model="profile.website" maxlength="120" placeholder="https://example.com 或社媒主页" class="h-11 rounded-2xl border-border/80 bg-background/90 shadow-none" />
                  <Button variant="outline" class="h-11 shrink-0 rounded-2xl shadow-none" :disabled="!profile.website.trim()" @click="copyWebsite">
                    <Copy class="mr-2 h-4 w-4" />
                    复制
                  </Button>
                </div>
              </div>

              <div class="grid gap-2">
                <Label for="profile-goal">每日目标</Label>
                <input
                  id="profile-goal"
                  v-model.number="profile.dailyGoal"
                  type="number"
                  min="1"
                  max="999"
                  :class="`${profileControlClass} h-11`"
                />
              </div>
            </div>

            <div class="grid gap-2 md:max-w-[300px]">
              <Label for="profile-start-page">默认进入页</Label>
              <div class="relative">
                <select
                  id="profile-start-page"
                  v-model="profile.preferredStartPage"
                  :class="`${profileControlClass} h-11 appearance-none pr-10`"
                >
                  <option v-for="option in startPageOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
                <ChevronDown class="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-3 border-t border-border pt-2">
              <Button class="rounded-2xl px-5" :disabled="!dirty || saving" @click="saveProfile">
                <Save class="mr-2 h-4 w-4" />
                {{ saving ? '保存中...' : '保存资料' }}
              </Button>
              <Button variant="outline" class="rounded-2xl px-5 shadow-none" :disabled="!dirty || saving" @click="resetProfile">
                恢复上次保存
              </Button>
              <span class="text-xs text-muted-foreground">
                {{ dirty ? '你有未保存的修改。' : '资料已与本地配置同步。' }}
              </span>
            </div>
          </CardContent>

          <CardContent v-else class="grid gap-4">
            <div class="h-11 rounded-2xl bg-muted/60" />
            <div class="h-11 rounded-2xl bg-muted/60" />
            <div class="h-32 rounded-[24px] bg-muted/60" />
            <div class="h-11 rounded-2xl bg-muted/60" />
          </CardContent>
          </Card>
        </div>

        <div class="grid gap-6 self-start">
          <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
            <CardHeader class="space-y-2">
              <CardTitle class="text-lg">图片卡片</CardTitle>
              <CardDescription>只能保留 1 张，上传新图会自动覆盖旧图。</CardDescription>
            </CardHeader>
            <CardContent class="grid gap-4">
              <div class="overflow-hidden rounded-[26px] border border-border bg-background/90">
                <div v-if="profile.imagePath" class="aspect-[4/3] overflow-hidden bg-muted/30">
                  <img :src="profile.imagePath" alt="个人中心图片预览" class="h-full w-full object-cover" />
                </div>
                <div v-else class="flex aspect-[4/3] items-center justify-center bg-muted/30">
                  <div class="flex flex-col items-center gap-3 text-muted-foreground">
                    <div class="flex h-14 w-14 items-center justify-center rounded-full border border-dashed border-border">
                      <ImageIcon class="h-6 w-6" />
                    </div>
                    <span class="text-sm">还没有上传个人中心图片</span>
                  </div>
                </div>
              </div>

              <div class="grid gap-2">
                <Button class="h-11 rounded-2xl" :disabled="uploadingImage" @click="uploadProfileImage">
                  <Camera class="mr-2 h-4 w-4" />
                  {{ uploadingImage ? '处理中...' : (profile.imagePath ? '更换图片' : '上传图片') }}
                </Button>
                <Button
                  variant="outline"
                  class="h-11 rounded-2xl shadow-none"
                  :disabled="uploadingImage || !profile.imagePath"
                  @click="clearProfileImage"
                >
                  <Trash2 class="mr-2 h-4 w-4" />
                  清除当前图片
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
            <CardHeader class="space-y-2">
              <CardTitle class="text-lg">外观与偏好</CardTitle>
              <CardDescription>这里的主题切换会直接调用项目自带的切换逻辑。</CardDescription>
            </CardHeader>
            <CardContent class="grid gap-4">
              <button
                type="button"
                class="flex items-center justify-between rounded-[22px] border border-border bg-background px-4 py-3 text-left transition-colors hover:bg-muted/35"
                @click="handleThemeToggle"
              >
                <div>
                  <p class="text-sm font-medium text-foreground">界面主题</p>
                  <p class="mt-1 text-xs text-muted-foreground">
                    当前为 {{ isDark ? '深色模式' : '浅色模式' }}，点击即可切换。
                  </p>
                </div>
                <div class="flex items-center gap-3">
                  <Sun v-if="!isDark" class="h-4 w-4 text-amber-500" />
                  <Moon v-else class="h-4 w-4 text-sky-400" />
                  <Switch :model-value="isDark" class="pointer-events-none" />
                </div>
              </button>

              <div class="rounded-[22px] border border-border bg-background px-4 py-4 text-sm text-muted-foreground">
                <p class="font-medium text-foreground">设置保存位置</p>
                <p class="mt-2 leading-6">
                  个人资料会和回收站、快捷键配置一起保存到本地 `data/settings.json`，
                  个人中心图片会单独存到 `data/profile/`，再次上传时会自动覆盖旧图。
                </p>
              </div>
            </CardContent>
          </Card>

          <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
            <CardHeader class="space-y-2">
              <CardTitle class="text-lg">创作摘要</CardTitle>
              <CardDescription>把个人信息和实际产出放在同一个视图里，更容易调整节奏。</CardDescription>
            </CardHeader>
            <CardContent class="grid gap-4 text-sm text-muted-foreground">
              <div :class="helperCardClass">
                <div class="flex items-center gap-2 text-foreground">
                  <ImageIcon class="h-4 w-4" />
                  <span class="font-medium">图库规模</span>
                </div>
                <p class="mt-3 text-2xl font-semibold tracking-tight text-foreground">{{ stats.totalCount }}</p>
                <p class="mt-2 leading-6">当前图库累计占用 {{ formatSize(stats.totalSize) }}，适合在这里定期回看和整理。</p>
              </div>

              <div :class="helperCardClass">
                <div class="flex items-center gap-2 text-foreground">
                  <Heart class="h-4 w-4" />
                  <span class="font-medium">偏好管理</span>
                </div>
                <p class="mt-3 leading-6">
                  你目前有 {{ favoriteGroupCount }} 个收藏分组、{{ tagCount }} 个标签，
                  默认从“{{ preferredStartPageLabel }}”进入应用。
                </p>
              </div>

              <div :class="helperCardClass">
                <div class="flex items-center gap-2 text-foreground">
                  <Link2 class="h-4 w-4" />
                  <span class="font-medium">个人链接</span>
                </div>
                <p class="mt-3 break-all leading-6">
                  {{ profile.website || '还没有设置主页链接，可以留空，也可以放自己的展示页或社媒主页。' }}
                </p>
              </div>
            </CardContent>
          </Card>

          <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
            <CardHeader class="space-y-2">
              <CardTitle class="text-lg">快捷入口</CardTitle>
              <CardDescription>把常用区域放在一侧，切换时更顺手。</CardDescription>
            </CardHeader>
            <CardContent class="grid gap-3">
              <button
                v-for="item in quickActions"
                :key="item.key"
                type="button"
                class="flex h-14 items-center justify-between rounded-2xl border border-border bg-background px-4 text-left transition-colors hover:bg-muted/35"
                @click="emit('navigate', item.key)"
              >
                <div class="flex items-center gap-3">
                  <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-muted text-muted-foreground">
                    <component :is="item.icon" class="h-4 w-4" />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-foreground">{{ item.label }}</p>
                    <p class="text-xs text-muted-foreground">{{ item.desc }}</p>
                  </div>
                </div>
              </button>
            </CardContent>
          </Card>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profile-hero {
  background:
    radial-gradient(circle at top right, color-mix(in oklab, var(--foreground) 8%, transparent) 0, transparent 34%),
    linear-gradient(180deg, color-mix(in oklab, var(--background) 86%, var(--card) 14%) 0%, var(--card) 100%);
}
</style>
