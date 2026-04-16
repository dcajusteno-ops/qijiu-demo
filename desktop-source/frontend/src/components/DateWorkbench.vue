<script setup>
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { CalendarDays, Layers3, Sparkles, Wand2 } from 'lucide-vue-next'

const props = defineProps({
  summary: { type: Object, default: () => ({}) },
  availableModels: { type: Array, default: () => [] },
  availableLoras: { type: Array, default: () => [] },
  activeDatePreset: { type: String, default: 'all' },
  activeDateValue: { type: String, default: '' },
  activeModelFilter: { type: String, default: '' },
  activeLoraFilter: { type: String, default: '' },
  activeDateLabel: { type: String, default: '全部日期' },
  filteredCount: { type: Number, default: 0 },
})

const emit = defineEmits([
  'update:date-preset',
  'update:date-value',
  'update:model-filter',
  'update:lora-filter',
  'clear-filters',
  'open-gallery',
])

const presetOptions = [
  { value: 'all', label: '全部' },
  { value: 'today', label: '今天' },
  { value: 'yesterday', label: '昨天' },
  { value: 'last7', label: '最近7天' },
  { value: 'month', label: '本月' },
]

const summaryCards = computed(() => [
  { key: 'today', label: '今日出图', value: props.summary?.today || 0, icon: Sparkles, preset: 'today' },
  { key: 'yesterday', label: '昨日出图', value: props.summary?.yesterday || 0, icon: CalendarDays, preset: 'yesterday' },
  { key: 'last7', label: '最近7天', value: props.summary?.last7 || 0, icon: Layers3, preset: 'last7' },
  { key: 'month', label: '本月累计', value: props.summary?.month || 0, icon: Wand2, preset: 'month' },
])

const recentDates = computed(() => (props.summary?.recentDates || []).slice(0, 12))

const hasActiveFilters = computed(() =>
  props.activeDatePreset !== 'all' || !!props.activeModelFilter || !!props.activeLoraFilter,
)
</script>

<template>
  <div class="h-full overflow-y-auto bg-background">
    <div class="mx-auto flex w-full max-w-6xl flex-col gap-6 px-6 py-6">
      <section class="rounded-3xl border border-border/70 bg-card/70 p-6 shadow-sm">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div class="space-y-2">
            <p class="text-xs font-medium uppercase tracking-[0.24em] text-muted-foreground">日期工作台</p>
            <div>
              <h1 class="text-3xl font-semibold tracking-tight text-foreground">日期产出工作台</h1>
              <p class="mt-2 text-sm text-muted-foreground">
                围绕 ComfyUI 的日期目录看图，再叠加模型和 LoRA 筛选，尽量不改变你现在的 output 使用习惯。
              </p>
            </div>
          </div>
          <div class="flex flex-wrap items-center gap-3">
            <div class="rounded-2xl border border-border/70 bg-background/80 px-4 py-2 text-sm text-muted-foreground">
              当前范围：<span class="font-medium text-foreground">{{ activeDateLabel }}</span>
            </div>
            <div class="rounded-2xl border border-border/70 bg-background/80 px-4 py-2 text-sm text-muted-foreground">
              当前命中：<span class="font-medium text-foreground">{{ filteredCount }}</span> 张
            </div>
            <Button class="rounded-2xl" @click="emit('open-gallery')">在图库中查看</Button>
          </div>
        </div>
      </section>

      <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <button
          v-for="card in summaryCards"
          :key="card.key"
          type="button"
          class="rounded-3xl border border-border/70 bg-card px-5 py-5 text-left transition hover:border-primary/40 hover:bg-accent/30"
          @click="emit('update:date-preset', card.preset)"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">{{ card.label }}</span>
            <component :is="card.icon" class="h-4 w-4 text-muted-foreground" />
          </div>
          <div class="mt-4 text-3xl font-semibold tracking-tight text-foreground">{{ card.value }}</div>
        </button>
      </section>

      <section class="grid gap-6 xl:grid-cols-[1.3fr_0.9fr]">
        <div class="rounded-3xl border border-border/70 bg-card p-6 shadow-sm">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold tracking-tight">快捷筛选</h2>
              <p class="mt-1 text-sm text-muted-foreground">
                先定时间，再补模型和 LoRA 条件，最后一键回到图库继续看图。
              </p>
            </div>
            <Button
              v-if="hasActiveFilters"
              variant="outline"
              class="rounded-2xl"
              @click="emit('clear-filters')"
            >
              清空条件
            </Button>
          </div>

          <div class="mt-6 space-y-6">
            <div>
              <label class="mb-3 block text-sm font-medium text-foreground">时间范围</label>
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="preset in presetOptions"
                  :key="preset.value"
                  type="button"
                  class="rounded-full border px-4 py-2 text-sm transition"
                  :class="activeDatePreset === preset.value ? 'border-primary bg-primary/10 text-primary' : 'border-border bg-background text-muted-foreground hover:text-foreground'"
                  @click="emit('update:date-preset', preset.value)"
                >
                  {{ preset.label }}
                </button>
              </div>
            </div>

            <div>
              <label class="mb-3 block text-sm font-medium text-foreground">指定日期</label>
              <input
                :value="activeDateValue"
                type="date"
                class="h-11 w-full rounded-2xl border border-border bg-background px-4 text-sm outline-none transition focus:border-primary"
                @input="emit('update:date-value', $event.target.value)"
              >
            </div>

            <div class="grid gap-4 lg:grid-cols-2">
              <div>
                <label class="mb-3 block text-sm font-medium text-foreground">模型</label>
                <select
                  :value="activeModelFilter"
                  class="h-11 w-full rounded-2xl border border-border bg-background px-4 text-sm outline-none transition focus:border-primary"
                  @change="emit('update:model-filter', $event.target.value)"
                >
                  <option value="">全部模型</option>
                  <option v-for="item in availableModels" :key="item.value || item.name" :value="item.value || item.name">
                    {{ item.label || item.name }} ({{ item.count }})
                  </option>
                </select>
              </div>
              <div>
                <label class="mb-3 block text-sm font-medium text-foreground">LoRA</label>
                <select
                  :value="activeLoraFilter"
                  class="h-11 w-full rounded-2xl border border-border bg-background px-4 text-sm outline-none transition focus:border-primary"
                  @change="emit('update:lora-filter', $event.target.value)"
                >
                  <option value="">全部 LoRA</option>
                  <option v-for="item in availableLoras" :key="item.value || item.name" :value="item.value || item.name">
                    {{ item.label || item.name }} ({{ item.count }})
                  </option>
                </select>
              </div>
            </div>
          </div>
        </div>

        <div class="rounded-3xl border border-border/70 bg-card p-6 shadow-sm">
          <div class="flex items-center justify-between gap-4">
            <div>
              <h2 class="text-lg font-semibold tracking-tight">最近活跃日期</h2>
              <p class="mt-1 text-sm text-muted-foreground">点一下就能切到那一天的产出。</p>
            </div>
            <div class="text-sm text-muted-foreground">
              日期图片：{{ summary?.datedTotal || 0 }}
            </div>
          </div>

          <div class="mt-5 flex flex-wrap gap-2">
            <button
              v-for="item in recentDates"
              :key="item.date"
              type="button"
              class="inline-flex items-center gap-2 rounded-full border border-border bg-background px-3 py-2 text-sm text-foreground transition hover:border-primary/40 hover:bg-accent/40"
              @click="emit('update:date-value', item.date)"
            >
              <span>{{ item.date }}</span>
              <span class="text-muted-foreground">{{ item.count }}</span>
            </button>
            <div
              v-if="recentDates.length === 0"
              class="rounded-2xl border border-dashed border-border px-4 py-6 text-sm text-muted-foreground"
            >
              还没有识别到日期目录图片。
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
