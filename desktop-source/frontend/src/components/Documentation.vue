<script setup>

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { 
  BookOpen, 
  Lightbulb, 
  Keyboard, 
  HelpCircle, 
  LayoutDashboard, 
  Calendar, 
  Wrench, 
  Trash2, 
  Moon, 
  Sun,
  Tag  // Added
} from 'lucide-vue-next'

const features = [
  {
    icon: LayoutDashboard,
    title: '工作室概览 (Studio Overview)',
    description: '直观展示最新的创作成果，提供快速的统计数据（今日新增、总图片数、存储占用），以及最近更新的文件夹快捷入口。'
  },
  {
    icon: Calendar,
    title: '日期归档 (Date Archives)',
    description: '自动识别并展示按日期归档的文件夹。ComfyUI 输出的图片通常按日期存放，管理器能自动解析这些结构，方便回顾历史创作。'
  },
  {
    icon: Wrench,
    title: '批量管理 (Batch Management)',
    description: '提供高效的批量选择模式。点击侧边栏底部的“批量管理”即可进入多选模式，支持移动、删除等操作，轻松整理大量图片。'
  },
  {
    icon: Trash2,
    title: '回收站保护 (Recycle Bin)',
    description: '误删保护机制。删除的图片会先进入回收站（.trash 文件夹），支持一键清空或还原，确保数据安全。'
  }
]

const shortcuts = [
  { key: 'Ctrl + Click', action: '多选图片 (进入选择模式)' },
  { key: 'Esc', action: '退出选择模式 / 关闭大图预览' },
  { key: '← / →', action: '在大图预览中切换上一张/下一张' },
  { key: 'Delete', action: '删除当前选中图片 (需确认)' }
]

const faqs = [
  {
    q: '图片存储在哪里？',
    a: 'Comfy Manager 直接读取您的 ComfyUI `output` 目录。它不会移动您的文件，除非您执行删除或移动操作。'
  },
  {
    q: '如何切换深色模式？',
    a: '点击侧边栏底部的 <span class="inline-flex items-center justify-center align-middle"><svg class="w-3 h-3 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path></svg></span> 图标即可在深色和浅色主题之间切换。'
  },
  {
    q: '如何清理空文件夹？',
    a: '侧边栏底部提供了“清理空文件夹”工具，点击后系统会扫描并移除 `output` 目录下所有空的子目录，保持目录整洁。'
  }
]
</script>

<template>
  <div class="h-full flex flex-col bg-background text-foreground">
    <!-- Header -->
    <div class="h-16 flex items-center px-6 border-b shrink-0 bg-card/50 backdrop-blur-sm">
      <BookOpen class="w-5 h-5 mr-3 text-primary" />
      <h1 class="text-xl font-bold tracking-tight">使用文档 (Documentation)</h1>
    </div>

    <div class="flex-1 overflow-y-auto">

      <div class="p-6 max-w-4xl mx-auto space-y-10 pb-20">
        
        <!-- Intro Section -->
        <section class="space-y-4">
          <div class="p-6 rounded-2xl bg-gradient-to-br from-primary/5 to-transparent border">
            <h2 class="text-2xl font-bold mb-2 flex items-center">
              <Lightbulb class="w-6 h-6 mr-2 text-amber-500" /> 
              欢迎使用 Comfy Manager
            </h2>
            <p class="text-muted-foreground leading-relaxed text-lg">
              **Comfy Manager** 是一款专为 ComfyUI 输出目录设计的现代化图片管理工具。旨在提供更加直观、高效的浏览与整理体验，让你从繁杂的文件夹中解放出来，专注于创作本身。
            </p>
          </div>
        </section>

        <!-- Features Grid -->
        <section class="space-y-6">
          <h3 class="text-xl font-semibold flex items-center border-l-4 border-primary pl-3">
            核心功能
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Card v-for="(feat, idx) in features" :key="idx" class="border-muted/60 shadow-sm hover:shadow-md transition-all hover:bg-accent/5">
              <CardHeader class="pb-2">
                <CardTitle class="text-base flex items-center gap-2">
                  <div class="p-2 rounded-md bg-primary/10 text-primary">
                    <component :is="feat.icon" class="w-4 h-4" />
                  </div>
                  {{ feat.title }}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription class="text-sm leading-relaxed">
                  {{ feat.description }}
                </CardDescription>
              </CardContent>
            </Card>
          </div>
        </section>

        <Separator />

        <!-- Shortcuts -->
        <section class="space-y-6">
          <h3 class="text-xl font-semibold flex items-center border-l-4 border-primary pl-3">
            <Keyboard class="w-5 h-5 mr-2" />
            快捷键指南
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div v-for="(item, i) in shortcuts" :key="i" class="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors">
              <span class="text-sm font-medium text-muted-foreground">{{ item.action }}</span>
              <Badge variant="outline" class="font-mono bg-muted/50 text-foreground shadow-sm">
                {{ item.key }}
              </Badge>
            </div>
          </div>
        </section>

        <Separator />

        <!-- Tag System -->
        <section class="space-y-6">
          <h3 class="text-xl font-semibold flex items-center border-l-4 border-primary pl-3">
             <Tag class="w-5 h-5 mr-2" />
             标签系统 (Tag System)
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
             <Card class="border-muted/60 shadow-sm">
                <CardHeader class="pb-2">
                   <CardTitle class="text-base">1. 创建与管理</CardTitle>
                </CardHeader>
                <CardContent>
                   <p class="text-sm text-muted-foreground leading-relaxed">
                      点击侧边栏的 <span class="font-bold">“管理标签”</span> 或 <span class="font-bold">“+”</span> 按钮即可打开管理器。支持自定义标签名称、颜色以及分类。
                   </p>
                </CardContent>
             </Card>

             <Card class="border-muted/60 shadow-sm">
                <CardHeader class="pb-2">
                   <CardTitle class="text-base">2. 批量操作</CardTitle>
                </CardHeader>
                <CardContent>
                   <p class="text-sm text-muted-foreground leading-relaxed">
                      在标签管理器中点击 <span class="font-bold">“批量管理”</span>。您可以选中多个标签，然后一次性进行<span class="font-bold">删除</span>或<span class="font-bold">修改分类</span>操作。
                   </p>
                </CardContent>
             </Card>

             <Card class="border-muted/60 shadow-sm">
                <CardHeader class="pb-2">
                   <CardTitle class="text-base">3. 快速筛选</CardTitle>
                </CardHeader>
                <CardContent>
                   <p class="text-sm text-muted-foreground leading-relaxed">
                      点击侧边栏中的任意标签即可过滤当前视图。再次点击或使用底部的 <span class="font-bold">“清除筛选”</span> 按钮即可恢复默认视图。
                   </p>
                </CardContent>
             </Card>
          </div>
        </section>

        <Separator />

        <!-- FAQ -->
        <section class="space-y-6">
          <h3 class="text-xl font-semibold flex items-center border-l-4 border-primary pl-3">
            <HelpCircle class="w-5 h-5 mr-2" />
            常见问题 (FAQ)
          </h3>
          <div class="space-y-4">
            <Card v-for="(item, i) in faqs" :key="i" class=" overflow-hidden">
              <div class="p-4 flex gap-4">
                <div class="font-bold text-primary/40 text-xl italic select-none">Q</div>
                <div class="space-y-2 flex-1">
                  <h4 class="font-medium text-base">{{ item.q }}</h4>
                  <p class="text-sm text-muted-foreground leading-relaxed" v-html="item.a"></p>
                </div>
              </div>
            </Card>
          </div>
        </section>
        
        <div class="text-center pt-10 text-muted-foreground/40 text-sm">
          Comfy Manager v1.0 · Designed for Creators
        </div>

      </div>
    </div>

  </div>
</template>
