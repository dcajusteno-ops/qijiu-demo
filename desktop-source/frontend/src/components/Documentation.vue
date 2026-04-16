<script setup>
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import {
  BookOpen,
  CalendarDays,
  Clock3,
  Filter,
  Heart,
  HelpCircle,
  Keyboard,
  LayoutDashboard,
  Search,
  Tag,
  Trash2,
  Wand2,
} from 'lucide-vue-next'

const featureCards = [
  {
    icon: LayoutDashboard,
    title: '工作台总览',
    description: '打开应用后的总入口，适合先看近期产出、快速返回常用区域。',
  },
  {
    icon: CalendarDays,
    title: '日期产出工作台',
    description: '围绕 ComfyUI 的日期目录浏览图片，支持今天、昨天、最近7天、本月和指定日期。',
  },
  {
    icon: Filter,
    title: '模型 / LoRA 筛选',
    description: '从图片 metadata 中提取模型和 LoRA，并和图库结果联动，方便回看某一套出图条件。',
  },
  {
    icon: Search,
    title: '搜索与整理',
    description: '支持按文件名、路径、Prompt、模型、LoRA、标签、笔记等内容搜索和整理。',
  },
  {
    icon: Heart,
    title: '收藏与分组',
    description: '把喜欢的图片放进收藏夹，还可以继续按分组做精选整理。',
  },
  {
    icon: Tag,
    title: '标签系统',
    description: '支持自定义标签、分类和颜色，并且可以作为常用筛选入口。',
  },
  {
    icon: Wand2,
    title: '自动规则',
    description: '根据模型、LoRA、Prompt、文件名等条件自动打标签、加入收藏组或移动目录。',
  },
  {
    icon: Trash2,
    title: '回收站保护',
    description: '删除图片会先进入回收站，支持恢复、清空和保留期清理。',
  },
]

const quickGuides = [
  {
    title: '按日期回看近期产出',
    steps: [
      '点击左侧“日期产出”进入工作台。',
      '先选“今天”“昨天”“最近7天”或“本月”，也可以直接指定某一天。',
      '需要继续挑图时，点击“在图库中查看”。',
    ],
  },
  {
    title: '按模型或 LoRA 回看图片',
    steps: [
      '进入日期工作台或普通图库。',
      '在顶部筛选栏选择模型或 LoRA。',
      '如果结果为空，可以使用“清空工作台筛选”或“清空全部筛选”快速恢复。',
    ],
  },
  {
    title: '批量整理图片',
    steps: [
      '点击左下角“批量模式”。',
      '选中多张图片后，可以导出、移动、加标签、收藏或删除。',
      '删除的图片会进入回收站，不会立即永久消失。',
    ],
  },
]

const shortcuts = [
  { key: 'Esc', action: '关闭大图预览，或退出当前的选择状态' },
  { key: 'Delete', action: '删除当前选中的图片，删除前会先确认' },
  { key: '方向键', action: '在大图预览中切换上一张 / 下一张' },
  { key: 'Ctrl + 点击', action: '配合批量模式快速选择多张图片' },
]

const faqs = [
  {
    q: '为什么侧边栏有数量，但主区域没有图片？',
    a: '通常不是图片丢了，而是顶部还保留着搜索词、模型、LoRA 或日期筛选。可以直接使用“清空全部筛选”。',
  },
  {
    q: '模型筛选为什么有时看起来像重复项？',
    a: 'v1.8.0 已经开始对模型名做归一化处理，但不同工作流仍可能写出不同的原始 metadata。现在会尽量合并路径、扩展名和大小写差异。',
  },
  {
    q: '图片文件会不会被程序自动改动？',
    a: '普通浏览、搜索、筛选不会改动图片。只有你主动执行删除、移动、导出或自动规则时，文件才会被处理。',
  },
  {
    q: '日期产出工作台和日期归档有什么区别？',
    a: '日期产出工作台更像快速入口和筛选台，适合看“最近产出”；日期归档是目录树视角，适合按年、月、日层级慢慢翻。',
  },
]
</script>

<template>
  <div class="flex h-full flex-col bg-background text-foreground">
    <div class="flex h-16 shrink-0 items-center border-b bg-card/50 px-6 backdrop-blur-sm">
      <BookOpen class="mr-3 h-5 w-5 text-primary" />
      <div class="flex items-center gap-3">
        <h1 class="text-xl font-bold tracking-tight">使用文档</h1>
        <Badge variant="outline" class="rounded-full px-3 py-1 text-xs">v1.8.0</Badge>
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
            <h2 class="text-3xl font-semibold tracking-tight">v1.8.0 使用说明</h2>
            <p class="max-w-3xl text-sm leading-7 text-muted-foreground">
              这一版的重点是“日期产出工作台”和“模型 / LoRA 筛选”。
              它不要求你改现有的 ComfyUI 输出目录结构，而是尽量直接围绕你现在的日期文件夹工作流来做整理。
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
          <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
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
          Comfy Manager v1.8.0 · 面向 ComfyUI 出图整理工作流
        </div>
      </div>
    </div>
  </div>
</template>
