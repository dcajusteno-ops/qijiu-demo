<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import {
  AlertCircle,
  CheckCircle2,
  Clock3,
  Pencil,
  Play,
  Plus,
  Search,
  Trash2,
  Workflow,
} from 'lucide-vue-next'
import AutoRulesDialog from './AutoRulesDialog.vue'
import * as App from '@/api'

const conditionLabels = {
  model: '模型',
  sampler: '采样器',
  lora: 'LoRA',
  dimensions: '尺寸',
  filename: '文件名',
  prompt: '正向 Prompt',
  negative: '反向 Prompt',
}

const operatorLabels = {
  contains: '包含',
  equals: '等于',
  starts_with: '开头是',
  ends_with: '结尾是',
}

const actionLabels = {
  add_tag: '添加标签',
  add_favorite_group: '加入收藏分组',
  move_to_folder: '移动到目录',
}

const loading = ref(true)
const saving = ref(false)
const running = ref(false)
const searchQuery = ref('')
const statusFilter = ref('all')
const store = ref({
  enabled: true,
  rules: [],
})
const dialogOpen = ref(false)
const editingRule = ref(null)
let disposeProgressListener = null

const createEmptyRunProgress = () => ({
  source: '',
  stage: 'idle',
  running: false,
  totalCount: 0,
  processedCount: 0,
  matchedCount: 0,
  updatedCount: 0,
  errorCount: 0,
  currentRelPath: '',
  currentRuleName: '',
  ranAt: '',
  message: '',
})

const runProgress = ref(createEmptyRunProgress())

const normalizeSearchText = (value) => String(value ?? '').trim().toLowerCase()

const enabledRuleCount = computed(() => store.value.rules.filter((rule) => rule.enabled).length)

const latestRuleRun = computed(() => {
  return [...store.value.rules]
    .filter((rule) => rule.lastRunAt)
    .sort((left, right) => new Date(right.lastRunAt).getTime() - new Date(left.lastRunAt).getTime())[0] || null
})

const failedRuleCount = computed(() =>
  store.value.rules.filter((rule) => rule.lastStatus === 'error').length,
)

const runProgressPercent = computed(() => {
  if (!runProgress.value.totalCount) return 0
  return Math.min(100, Math.round((runProgress.value.processedCount / runProgress.value.totalCount) * 100))
})

const showRunProgress = computed(() =>
  running.value || ['completed', 'failed'].includes(runProgress.value.stage),
)

const runProgressTitle = computed(() => {
  if (running.value || runProgress.value.stage === 'running' || runProgress.value.stage === 'started') {
    return '规则执行中'
  }
  if (runProgress.value.stage === 'failed') {
    return '本次执行异常结束'
  }
  if (runProgress.value.stage === 'completed') {
    return '本次执行已完成'
  }
  return '规则执行进度'
})

const runProgressDescription = computed(() => {
  if (runProgress.value.currentRelPath) {
    return `当前图片：${runProgress.value.currentRelPath}`
  }
  if (runProgress.value.message) {
    return runProgress.value.message
  }
  if (runProgress.value.totalCount > 0) {
    return `准备处理 ${runProgress.value.totalCount} 张图片`
  }
  return '等待开始'
})

const filteredRules = computed(() => {
  const normalizedQuery = normalizeSearchText(searchQuery.value)

  return store.value.rules.filter((rule) => {
    const matchesStatus =
      statusFilter.value === 'all'
      || (statusFilter.value === 'enabled' && rule.enabled)
      || (statusFilter.value === 'disabled' && !rule.enabled)
      || (statusFilter.value === 'failed' && rule.lastStatus === 'error')

    if (!matchesStatus) return false
    if (!normalizedQuery) return true

    const conditionText = (rule.conditions || [])
      .map((condition) => `${condition.field} ${condition.operator} ${condition.value}`)
      .join(' ')
    const actionText = (rule.actions || [])
      .map((action) => `${action.type} ${action.value}`)
      .join(' ')

    return [rule.name, conditionText, actionText]
      .some((item) => normalizeSearchText(item).includes(normalizedQuery))
  })
})

const latestErrorText = computed(() => {
  const failedRule = [...store.value.rules]
    .filter((rule) => rule.lastStatus === 'error' && rule.lastError)
    .sort((left, right) => new Date(right.lastRunAt || 0).getTime() - new Date(left.lastRunAt || 0).getTime())[0]

  return failedRule?.lastError || '最近一次执行没有发现异常。'
})

const formatDateTime = (value) => {
  if (!value) return '暂无记录'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString('zh-CN', {
    hour12: false,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatConditionSummary = (rule) =>
  (rule.conditions || [])
    .map((condition) => `${conditionLabels[condition.field] || condition.field} ${operatorLabels[condition.operator] || condition.operator} ${condition.value}`)
    .join('，')

const formatActionSummary = (rule) =>
  (rule.actions || [])
    .map((action) => `${actionLabels[action.type] || action.type} ${action.value}`)
    .join('，')

const statusBadgeClass = (rule) => {
  if (rule.lastStatus === 'error') return 'border-destructive/30 text-destructive'
  if (rule.lastStatus === 'success') return 'border-emerald-500/20 text-emerald-600'
  return 'text-muted-foreground'
}

const statusLabel = (rule) => {
  if (rule.lastStatus === 'error') return '最近失败'
  if (rule.lastStatus === 'success') return '最近成功'
  return '未执行'
}

const loadRules = async () => {
  loading.value = true
  try {
    const data = await App.GetAutoRules()
    store.value = {
      enabled: data?.enabled ?? true,
      rules: Array.isArray(data?.rules) ? data.rules : [],
    }
  } catch (error) {
    console.error('Failed to load auto rules:', error)
    toast.error(`自动规则加载失败: ${error}`)
  } finally {
    loading.value = false
  }
}

const setGlobalEnabled = async (value) => {
  const previous = store.value.enabled
  store.value.enabled = value
  try {
    const data = await App.SetAutoRulesEnabled(value)
    store.value = {
      enabled: data?.enabled ?? value,
      rules: Array.isArray(data?.rules) ? data.rules : store.value.rules,
    }
    toast.success(value ? '自动规则已开启' : '自动规则已关闭')
  } catch (error) {
    store.value.enabled = previous
    console.error('Failed to update auto rules status:', error)
    toast.error(`更新失败: ${error}`)
  }
}

const openCreateDialog = () => {
  editingRule.value = null
  dialogOpen.value = true
}

const openEditDialog = (rule) => {
  editingRule.value = JSON.parse(JSON.stringify(rule))
  dialogOpen.value = true
}

const closeDialog = () => {
  dialogOpen.value = false
  editingRule.value = null
}

const saveRule = async (rule) => {
  saving.value = true
  try {
    if (rule.id) {
      await App.UpdateAutoRule(rule)
      toast.success('规则已更新')
    } else {
      await App.CreateAutoRule(rule)
      toast.success('规则已创建')
    }
    closeDialog()
    await loadRules()
  } catch (error) {
    console.error('Failed to save auto rule:', error)
    toast.error(`保存失败: ${error}`)
  } finally {
    saving.value = false
  }
}

const deleteRule = async (ruleId) => {
  const target = store.value.rules.find((rule) => rule.id === ruleId)
  if (!target) return
  if (!window.confirm(`确定删除“${target.name}”吗？`)) return

  saving.value = true
  try {
    await App.DeleteAutoRule(ruleId)
    toast.success('规则已删除')
    closeDialog()
    await loadRules()
  } catch (error) {
    console.error('Failed to delete auto rule:', error)
    toast.error(`删除失败: ${error}`)
  } finally {
    saving.value = false
  }
}

const toggleRuleEnabled = async (rule, value) => {
  const previous = rule.enabled
  rule.enabled = value
  try {
    await App.UpdateAutoRule({
      ...rule,
      enabled: value,
    })
    toast.success(value ? '规则已启用' : '规则已停用')
    await loadRules()
  } catch (error) {
    rule.enabled = previous
    console.error('Failed to toggle rule:', error)
    toast.error(`更新失败: ${error}`)
  }
}

const runRulesNow = async () => {
  running.value = true
  runProgress.value = {
    ...createEmptyRunProgress(),
    source: 'manual',
    stage: 'started',
    running: true,
    message: '正在准备执行自动规则',
  }
  try {
    const summary = await App.RunAutoRulesNow()
    runProgress.value = {
      ...runProgress.value,
      source: 'manual',
      stage: 'completed',
      running: false,
      totalCount: summary?.totalCount || runProgress.value.totalCount || 0,
      processedCount: summary?.processedCount || 0,
      matchedCount: summary?.matchedCount || 0,
      updatedCount: summary?.updatedCount || 0,
      errorCount: summary?.errorCount || 0,
      message: '',
    }
    const summaryText = `已扫描 ${summary?.processedCount || 0} 张，命中 ${summary?.matchedCount || 0} 次，实际更新 ${summary?.updatedCount || 0} 项`
    if ((summary?.errorCount || 0) > 0) {
      toast.warning(`${summaryText}，其中 ${summary.errorCount} 次执行失败`)
    } else {
      toast.success(summaryText)
    }
    await loadRules()
  } catch (error) {
    console.error('Failed to run auto rules:', error)
    toast.error(`执行失败: ${error}`)
  } finally {
    running.value = false
  }
}

onMounted(() => {
  loadRules()
  disposeProgressListener = EventsOn('auto-rules:progress', (payload) => {
    const nextProgress = Array.isArray(payload) ? payload[0] : payload
    if (!nextProgress || nextProgress.source !== 'manual') return

    runProgress.value = {
      ...runProgress.value,
      ...nextProgress,
    }

    if (!nextProgress.running || ['completed', 'failed'].includes(nextProgress.stage)) {
      running.value = false
    }
  })
})

onBeforeUnmount(() => {
  if (typeof disposeProgressListener === 'function') {
    disposeProgressListener()
  }
})
</script>

<template>
  <div class="grid gap-4">
    <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
      <CardHeader class="space-y-2">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <Workflow class="h-4 w-4 text-muted-foreground" />
              <CardTitle class="text-lg">自动规则</CardTitle>
            </div>
            <CardDescription>把重复整理动作交给系统自动完成，控制开关统一放在个人中心。</CardDescription>
          </div>

          <div class="flex items-center gap-3 rounded-2xl border border-border bg-background px-4 py-3">
            <div class="space-y-1 text-right">
              <p class="text-sm font-medium text-foreground">自动处理开关</p>
              <p class="text-xs text-muted-foreground">关闭后不会自动处理新图片</p>
            </div>
            <Switch :model-value="store.enabled" @update:model-value="setGlobalEnabled" />
          </div>
        </div>
      </CardHeader>

      <CardContent class="grid gap-4">
        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-[22px] border border-border bg-background px-4 py-4">
            <p class="text-xs text-muted-foreground">规则总数</p>
            <p class="mt-2 text-2xl font-semibold tracking-tight text-foreground">{{ store.rules.length }}</p>
          </div>
          <div class="rounded-[22px] border border-border bg-background px-4 py-4">
            <p class="text-xs text-muted-foreground">已启用</p>
            <p class="mt-2 text-2xl font-semibold tracking-tight text-foreground">{{ enabledRuleCount }}</p>
          </div>
          <div class="rounded-[22px] border border-border bg-background px-4 py-4">
            <p class="text-xs text-muted-foreground">最近执行</p>
            <p class="mt-2 text-sm font-medium text-foreground">{{ formatDateTime(latestRuleRun?.lastRunAt) }}</p>
          </div>
          <div class="rounded-[22px] border border-border bg-background px-4 py-4">
            <p class="text-xs text-muted-foreground">最近失败</p>
            <p class="mt-2 text-2xl font-semibold tracking-tight text-foreground">{{ failedRuleCount }}</p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <Button class="rounded-2xl px-5" @click="openCreateDialog">
            <Plus class="mr-2 h-4 w-4" />
            新建规则
          </Button>
          <Button variant="outline" class="rounded-2xl px-5 shadow-none" :disabled="running" @click="runRulesNow">
            <Play class="mr-2 h-4 w-4" />
            {{ running ? '执行中...' : '立即执行一次' }}
          </Button>
          <Button
            variant="outline"
            class="rounded-2xl px-5 shadow-none"
            @click="searchQuery = ''; statusFilter = 'all'"
          >
            重置筛选
          </Button>
        </div>

        <div
          v-if="showRunProgress"
          class="rounded-[22px] border border-border bg-background px-4 py-4"
        >
          <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div class="space-y-1">
              <p class="text-sm font-medium text-foreground">{{ runProgressTitle }}</p>
              <p class="text-xs text-muted-foreground">{{ runProgressDescription }}</p>
            </div>
            <div class="text-left text-xs text-muted-foreground lg:text-right">
              <p>{{ runProgress.processedCount }} / {{ runProgress.totalCount || 0 }}</p>
              <p>{{ runProgressPercent }}%</p>
            </div>
          </div>

          <div class="mt-3 h-2 overflow-hidden rounded-full bg-muted">
            <div
              class="h-full rounded-full bg-foreground transition-all duration-200"
              :style="{ width: `${runProgressPercent}%` }"
            />
          </div>

          <div class="mt-3 flex flex-wrap gap-4 text-xs text-muted-foreground">
            <span>命中 {{ runProgress.matchedCount }}</span>
            <span>更新 {{ runProgress.updatedCount }}</span>
            <span>失败 {{ runProgress.errorCount }}</span>
            <span v-if="runProgress.currentRuleName">当前规则 {{ runProgress.currentRuleName }}</span>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
      <CardHeader class="space-y-2">
        <CardTitle class="text-lg">规则列表</CardTitle>
        <CardDescription>支持搜索、启停、编辑、删除，也支持手动跑一遍规则验证效果。</CardDescription>
      </CardHeader>

      <CardContent class="grid gap-4">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-center">
          <div class="relative flex-1">
            <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索规则名、条件或动作"
              class="h-11 rounded-2xl border-border/80 bg-background/90 pl-9 shadow-none"
            />
          </div>

          <div class="relative lg:w-[160px]">
            <select
              v-model="statusFilter"
              class="h-11 w-full appearance-none rounded-2xl border border-border/80 bg-background px-4 pr-10 text-sm shadow-none outline-none transition-[border-color,box-shadow] focus:border-ring focus:ring-1 focus:ring-ring/60"
            >
              <option value="all">全部规则</option>
              <option value="enabled">已启用</option>
              <option value="disabled">已停用</option>
              <option value="failed">最近失败</option>
            </select>
          </div>
        </div>

        <div v-if="loading" class="grid gap-3">
          <div v-for="index in 3" :key="index" class="h-28 rounded-[24px] bg-muted/60" />
        </div>

        <div v-else class="grid gap-3">
          <div
            v-for="rule in filteredRules"
            :key="rule.id"
            class="rounded-[24px] border border-border bg-background px-4 py-4"
          >
            <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0 flex-1 space-y-3">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="text-sm font-medium text-foreground">{{ rule.name }}</p>
                  <Badge variant="outline" class="rounded-full px-2.5 py-0.5 text-[11px]">
                    {{ rule.enabled ? '已启用' : '已停用' }}
                  </Badge>
                  <Badge variant="outline" class="rounded-full px-2.5 py-0.5 text-[11px]" :class="statusBadgeClass(rule)">
                    {{ statusLabel(rule) }}
                  </Badge>
                </div>

                <div class="space-y-2 text-sm">
                  <p class="text-muted-foreground">
                    <span class="font-medium text-foreground">条件：</span>
                    {{ formatConditionSummary(rule) }}
                  </p>
                  <p class="text-muted-foreground">
                    <span class="font-medium text-foreground">动作：</span>
                    {{ formatActionSummary(rule) }}
                  </p>
                </div>

                <p class="text-xs text-muted-foreground">
                  最近执行：{{ formatDateTime(rule.lastRunAt) }}，命中 {{ rule.lastMatchCount || 0 }} 次
                </p>
              </div>

              <div class="flex items-center gap-2 self-start">
                <Switch :model-value="rule.enabled" @update:model-value="toggleRuleEnabled(rule, $event)" />
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-10 w-10 rounded-2xl text-muted-foreground hover:bg-muted"
                  @click="openEditDialog(rule)"
                >
                  <Pencil class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-10 w-10 rounded-2xl text-muted-foreground hover:bg-destructive/8 hover:text-destructive"
                  @click="deleteRule(rule.id)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>

          <div
            v-if="filteredRules.length === 0"
            class="rounded-[24px] border border-dashed border-border bg-background/70 px-4 py-10 text-center text-sm text-muted-foreground"
          >
            当前筛选条件下没有规则。
          </div>
        </div>
      </CardContent>
    </Card>

    <Card class="rounded-[30px] border-border/70 bg-card shadow-none">
      <CardHeader class="space-y-2">
        <CardTitle class="text-lg">执行状态</CardTitle>
        <CardDescription>快速查看最近运行是否正常，方便确认规则是否生效。</CardDescription>
      </CardHeader>

      <CardContent class="grid gap-3 text-sm">
        <div class="flex items-center gap-3 rounded-[22px] border border-border bg-background px-4 py-3">
          <Clock3 class="h-4 w-4 text-muted-foreground" />
          <div class="min-w-0">
            <p class="font-medium text-foreground">最近执行时间</p>
            <p class="mt-1 text-muted-foreground">{{ formatDateTime(latestRuleRun?.lastRunAt) }}</p>
          </div>
        </div>

        <div class="flex items-center gap-3 rounded-[22px] border border-border bg-background px-4 py-3">
          <CheckCircle2 class="h-4 w-4 text-muted-foreground" />
          <div class="min-w-0">
            <p class="font-medium text-foreground">最近命中规则</p>
            <p class="mt-1 text-muted-foreground">{{ latestRuleRun?.name || '规则会在新图片进入图库后自动运行' }}</p>
          </div>
        </div>

        <div class="flex items-center gap-3 rounded-[22px] border border-border bg-background px-4 py-3">
          <AlertCircle class="h-4 w-4 text-muted-foreground" />
          <div class="min-w-0">
            <p class="font-medium text-foreground">最近失败信息</p>
            <p class="mt-1 text-muted-foreground">{{ latestErrorText }}</p>
          </div>
        </div>
      </CardContent>
    </Card>

    <AutoRulesDialog
      :open="dialogOpen"
      :rule="editingRule"
      :saving="saving"
      @update:open="dialogOpen = $event"
      @save="saveRule"
      @delete="deleteRule"
    />
  </div>
</template>
