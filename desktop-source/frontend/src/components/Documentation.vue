<script setup>
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import {
  BookOpen,
  CalendarDays,
  Clock3,
  Filter,
  FolderTree,
  Heart,
  HelpCircle,
  Keyboard,
  LayoutDashboard,
  Search,
  Settings2,
  Tag,
  Trash2,
  Wand2,
} from 'lucide-vue-next'

const featureCards = [
  {
    icon: LayoutDashboard,
    title: '工作台总览',
    description: '打开应用后的主入口，适合先看最新作品、最近更新和整体出图概览。',
  },
  {
    icon: CalendarDays,
    title: '日期产出工作台',
    description: '围绕日期目录快速回看近期产出，支持今天、昨天、最近 7 天、本月和自定义日期范围。',
  },
  {
    icon: Filter,
    title: '模型 / LoRA 筛选',
    description: '自动从 PNG 元数据中提取模型与 LoRA，用于图库和日期工作台的联合筛选。',
  },
  {
    icon: Search,
    title: '搜索与细查',
    description: '支持按文件名、路径、Prompt、模型、LoRA、标签和笔记进行搜索。',
  },
  {
    icon: FolderTree,
    title: '目录管理',
    description: '支持默认目录、日期归档目录、自定义目录并行浏览，也支持重新绑定任意 output 目录。',
  },
  {
    icon: Heart,
    title: '收藏与分组',
    description: '可以把喜欢的图片加入收藏夹，并继续按分组做精细整理。',
  },
  {
    icon: Tag,
    title: '标签系统',
    description: '支持自定义标签、颜色和分组，可作为常用筛选入口，也可配合自动规则使用。',
  },
  {
    icon: Wand2,
    title: '自动规则引擎',
    description: '按模型、LoRA、Prompt 或文件名自动执行打标、归类、收藏等规则。',
  },
  {
    icon: Settings2,
    title: '设置中心',
    description: '统一管理主题、快捷键、缓存、文件夹维护和工具菜单顺序与显示。',
  },
  {
    icon: Trash2,
    title: '回收站保护',
    description: '删除图片会先进入回收站，支持恢复、清空和定期清理。',
  },
]

const quickGuides = [
  {
    title: '按日期范围回看近期产出',
    steps: [
      '点击左侧“日期产出”进入日期工作台。',
      '先选择“今天”“昨天”“最近 7 天”“本月”，或者直接指定开始日期和结束日期。',
      '需要继续大范围挑图时，点击“在图库中查看”跳回主图库。',
    ],
  },
  {
    title: '按模型或 LoRA 回看作品',
    steps: [
      '进入日期工作台或任意图库页。',
      '在顶部筛选栏选择模型或 LoRA。',
      '如果结果为空，可以使用“清空工作台筛选”或“清空全部筛选”快速恢复。',
    ],
  },
  {
    title: '整理目录与常用工具',
    steps: [
      '打开“工具菜单 -> 设置”，进入设置中心。',
      '在“工具菜单”中调整工具顺序与显示，在“文件夹维护”中执行清理或按日期整理。',
      '如果 output 位置变了，可在工具菜单里重新绑定目录。',
    ],
  },
  {
    title: '批量整理图片',
    steps: [
      '点击左下角“批量模式”。',
      '选中多张图片后，可以导出、移动、加标签、收藏或删除。',
      '删除的图片会先进入回收站，不会立即永久丢失。',
    ],
  },
]

const shortcuts = [
  { key: 'Esc', action: '关闭大图预览，或退出当前选择状态' },
  { key: 'Delete', action: '删除当前选中的图片，删除前会先确认' },
  { key: '方向键', action: '在大图预览中切换上一张 / 下一张' },
  { key: 'Ctrl + 0', action: '在大图预览中重置缩放' },
  { key: '批量模式 + 点击', action: '快速多选图片' },
]

const faqs = [
  {
    q: '为什么侧边栏有数量，但主区域没有图片？',
    a: '通常不是图片没了，而是顶部还保留着搜索词、模型、LoRA、日期或标签筛选。可以直接使用“清空全部筛选”恢复。',
  },
  {
    q: '为什么软件第一次进入会要求我选择 output 目录？',
    a: '从 v2.0 开始，程序不再默认猜测 exe 上一级目录，而是要求绑定真实的 ComfyUI output 目录，这样才能适配任意安装位置。',
  },
  {
    q: '日期产出工作台和日期归档目录有什么区别？',
    a: '日期产出工作台更像快速筛选台，适合看最近产出；日期归档目录是目录树视角，适合按年和具体日期慢慢翻。',
  },
  {
    q: '工具菜单太长怎么办？',
    a: '可以打开“设置中心 -> 工具菜单”，调整顺序，并隐藏不常用入口。设置按钮会固定保留在最上方。',
  },
]
</script>

<template>
  <div class="flex h-full flex-col bg-background text-foreground">
    <div class="flex h-16 shrink-0 items-center border-b bg-card/50 px-6 backdrop-blur-sm">
      <BookOpen class="mr-3 h-5 w-5 text-primary" />
      <div class="flex items-center gap-3">
        <h1 class="text-xl font-bold tracking-tight">使用文档</h1>
        <Badge variant="outline" class="rounded-full px-3 py-1 text-xs">v2.0.1</Badge>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto">
      <div class="mx-auto flex max-w-5xl flex-col gap-10 px-6 py-6 pb-20">
        <section class="rounded-[28px] border border-border/70 bg-gradient-to-br from-primary/8 via-transparent to-transparent p-7">
          <div class="space-y-3">
            <div class="flex items-center gap-2 text-sm text-muted-foreground">
              <Clock3 class="h-4 w-4" />
              <span>当前版本说明</span>
            </div>
            <h2 class="text-3xl font-semibold tracking-tight">v2.0.1 使用说明</h2>
            <p class="max-w-3xl text-sm leading-7 text-muted-foreground">
              这一版重点补齐了软件内的中文使用文档，并和外部 README、发布文档保持一致。
              如果你主要通过内置“使用文档”了解功能，现在看到的内容已经对齐当前版本。
            </p>
          </div>
        </section>

        <section class="space-y-5">
          <div class="flex items-center gap-3">
            <div class="h-6 w-1 rounded-full bg-primary" />
            <h3 class="text-xl font-semibold">核心功能</h3>
          </div>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <Card
              v-for="feature in featureCards"
              :key="feature.title"
              class="border-border/70 bg-card/70 shadow-sm transition hover:border-primary/30 hover:bg-accent/20"
            >
              <CardHeader class="pb-3">
                <CardTitle class="flex items-center gap-3 text-base">
                  <span class="rounded-xl bg-primary/10 p-2 text-primary">
                    <component :is="feature.icon" class="h-4 w-4" />
                  </span>
                  <span>{{ feature.title }}</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription class="text-sm leading-7 text-muted-foreground">
                  {{ feature.description }}
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </section>

        <Separator />

        <section class="space-y-5">
          <div class="flex items-center gap-3">
            <div class="h-6 w-1 rounded-full bg-primary" />
            <h3 class="text-xl font-semibold">快速上手</h3>
          </div>
          <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
            <Card
              v-for="guide in quickGuides"
              :key="guide.title"
              class="border-border/70 bg-card/70 shadow-sm"
            >
              <CardHeader class="pb-3">
                <CardTitle class="text-base">{{ guide.title }}</CardTitle>
              </CardHeader>
              <CardContent>
                <ol class="space-y-2 text-sm leading-7 text-muted-foreground">
                  <li v-for="(step, index) in guide.steps" :key="step">
                    {{ index + 1 }}. {{ step }}
                  </li>
                </ol>
              </CardContent>
            </Card>
          </div>
        </section>

        <Separator />

        <section class="space-y-5">
          <div class="flex items-center gap-3">
            <div class="h-6 w-1 rounded-full bg-primary" />
            <h3 class="flex items-center text-xl font-semibold">
              <Keyboard class="mr-2 h-5 w-5" />
              常用快捷操作
            </h3>
          </div>
          <div class="grid grid-cols-1 gap-3 md:grid-cols-2">
            <div
              v-for="shortcut in shortcuts"
              :key="shortcut.key"
              class="flex items-center justify-between rounded-2xl border border-border/70 bg-card/70 px-4 py-3"
            >
              <span class="pr-4 text-sm text-muted-foreground">{{ shortcut.action }}</span>
              <Badge variant="outline" class="rounded-full px-3 py-1 font-mono text-xs">
                {{ shortcut.key }}
              </Badge>
            </div>
          </div>
        </section>

        <Separator />

        <section class="space-y-5">
          <div class="flex items-center gap-3">
            <div class="h-6 w-1 rounded-full bg-primary" />
            <h3 class="flex items-center text-xl font-semibold">
              <HelpCircle class="mr-2 h-5 w-5" />
              常见问题
            </h3>
          </div>
          <div class="space-y-4">
            <Card
              v-for="item in faqs"
              :key="item.q"
              class="border-border/70 bg-card/70 shadow-sm"
            >
              <CardContent class="space-y-3 p-5">
                <div class="text-base font-semibold">{{ item.q }}</div>
                <p class="text-sm leading-7 text-muted-foreground">{{ item.a }}</p>
              </CardContent>
            </Card>
          </div>
        </section>

        <div class="pt-4 text-center text-sm text-muted-foreground">
          Comfy Manager v2.0.1 · 面向 ComfyUI 出图整理工作流
        </div>
      </div>
    </div>
  </div>
</template>
