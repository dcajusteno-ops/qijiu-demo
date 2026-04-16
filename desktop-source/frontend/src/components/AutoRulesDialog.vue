<script setup>
import { computed, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ChevronDown, Plus, Trash2 } from 'lucide-vue-next'

const props = defineProps({
  open: { type: Boolean, default: false },
  rule: { type: Object, default: null },
  saving: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'save', 'delete'])

const fieldOptions = [
  { value: 'model', label: '模型' },
  { value: 'sampler', label: '采样器' },
  { value: 'lora', label: 'LoRA' },
  { value: 'dimensions', label: '尺寸' },
  { value: 'filename', label: '文件名' },
  { value: 'prompt', label: '正向 Prompt' },
  { value: 'negative', label: '反向 Prompt' },
]

const operatorOptions = [
  { value: 'contains', label: '包含' },
  { value: 'equals', label: '等于' },
  { value: 'starts_with', label: '开头是' },
  { value: 'ends_with', label: '结尾是' },
]

const actionOptions = [
  { value: 'add_tag', label: '添加标签', placeholder: '例如：Pony' },
  { value: 'add_favorite_group', label: '加入收藏分组', placeholder: '例如：精选人像' },
  { value: 'move_to_folder', label: '移动到目录', placeholder: '例如：日期归档/竖图' },
]

const createCondition = () => ({
  field: 'model',
  operator: 'contains',
  value: '',
})

const createAction = () => ({
  type: 'add_tag',
  value: '',
})

const createRuleDraft = () => ({
  id: '',
  name: '',
  enabled: true,
  matchMode: 'all',
  conditions: [createCondition()],
  actions: [createAction()],
})

const draft = ref(createRuleDraft())

const isEditing = computed(() => Boolean(props.rule?.id))

const cloneRule = (rule) => ({
  id: rule?.id || '',
  name: rule?.name || '',
  enabled: rule?.enabled ?? true,
  matchMode: rule?.matchMode || 'all',
  conditions: (rule?.conditions || [createCondition()]).map((condition) => ({
    field: condition.field || 'model',
    operator: condition.operator || 'contains',
    value: condition.value || '',
  })),
  actions: (rule?.actions || [createAction()]).map((action) => ({
    type: action.type || 'add_tag',
    value: action.value || '',
  })),
})

const actionPlaceholderMap = computed(() =>
  new Map(actionOptions.map((option) => [option.value, option.placeholder])),
)

watch(
  () => [props.open, props.rule],
  () => {
    if (!props.open) return
    draft.value = props.rule ? cloneRule(props.rule) : createRuleDraft()
  },
  { deep: true },
)

const addCondition = () => {
  draft.value.conditions.push(createCondition())
}

const removeCondition = (index) => {
  if (draft.value.conditions.length === 1) return
  draft.value.conditions.splice(index, 1)
}

const addAction = () => {
  draft.value.actions.push(createAction())
}

const removeAction = (index) => {
  if (draft.value.actions.length === 1) return
  draft.value.actions.splice(index, 1)
}

const saveRule = () => {
  const normalizedName = String(draft.value.name || '').trim()
  if (!normalizedName) {
    toast.error('请输入规则名称')
    return
  }

  if (draft.value.conditions.some((condition) => !String(condition.value || '').trim())) {
    toast.error('请补全所有条件值')
    return
  }

  if (draft.value.actions.some((action) => !String(action.value || '').trim())) {
    toast.error('请补全所有动作目标')
    return
  }

  emit('save', {
    ...draft.value,
    name: normalizedName,
    conditions: draft.value.conditions.map((condition) => ({
      field: condition.field,
      operator: condition.operator,
      value: String(condition.value || '').trim(),
    })),
    actions: draft.value.actions.map((action) => ({
      type: action.type,
      value: String(action.value || '').trim(),
    })),
  })
}

const requestDelete = () => {
  if (!draft.value.id) return
  emit('delete', draft.value.id)
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent class="overflow-hidden p-0 sm:max-w-[760px]">
      <div class="max-h-[82vh] overflow-y-auto px-6 py-5">
        <DialogHeader class="space-y-2 pr-8">
          <DialogTitle>{{ isEditing ? '编辑规则' : '新建规则' }}</DialogTitle>
          <DialogDescription>
            规则会在新图片进入图库时自动执行，你也可以在个人中心手动触发一次全库运行。
          </DialogDescription>
        </DialogHeader>

        <div class="mt-6 grid gap-6">
          <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_140px_180px]">
            <div class="grid gap-2">
              <Label for="rule-name">规则名称</Label>
              <Input
                id="rule-name"
                v-model="draft.name"
                placeholder="例如：Pony 自动打标"
                class="h-11 rounded-2xl shadow-none"
              />
            </div>

            <div class="grid gap-2">
              <Label for="rule-match-mode">匹配模式</Label>
              <div class="relative">
                <select
                  id="rule-match-mode"
                  v-model="draft.matchMode"
                  class="h-11 w-full appearance-none rounded-2xl border border-border/80 bg-background px-4 pr-10 text-sm shadow-none outline-none transition-[border-color,box-shadow] focus:border-ring focus:ring-1 focus:ring-ring/60"
                >
                  <option value="all">全部满足</option>
                  <option value="any">任一满足</option>
                </select>
                <ChevronDown class="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              </div>
            </div>

            <div class="flex items-end justify-between rounded-[22px] border border-border bg-muted/20 px-4 py-3">
              <div class="space-y-1">
                <p class="text-sm font-medium text-foreground">启用规则</p>
                <p class="text-xs text-muted-foreground">关闭后会保留配置，但不参与执行</p>
              </div>
              <Switch :model-value="draft.enabled" @update:model-value="draft.enabled = $event" />
            </div>
          </div>

          <section class="grid gap-3 rounded-[24px] border border-border bg-card p-4">
            <div class="flex items-center justify-between gap-4">
              <div>
                <h3 class="text-sm font-medium text-foreground">条件</h3>
                <p class="mt-1 text-xs text-muted-foreground">支持模型、Prompt、尺寸、文件名等字段。</p>
              </div>
              <Button variant="outline" class="h-9 rounded-2xl shadow-none" @click="addCondition">
                <Plus class="mr-2 h-4 w-4" />
                新增条件
              </Button>
            </div>

            <div class="grid gap-3">
              <div
                v-for="(condition, index) in draft.conditions"
                :key="`condition-${index}`"
                class="grid gap-3 rounded-[20px] border border-border/80 bg-background p-3 md:grid-cols-[150px_150px_minmax(0,1fr)_44px]"
              >
                <div class="relative">
                  <select
                    v-model="condition.field"
                    class="h-11 w-full appearance-none rounded-2xl border border-border/80 bg-background px-4 pr-10 text-sm shadow-none outline-none transition-[border-color,box-shadow] focus:border-ring focus:ring-1 focus:ring-ring/60"
                  >
                    <option v-for="option in fieldOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                  <ChevronDown class="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                </div>

                <div class="relative">
                  <select
                    v-model="condition.operator"
                    class="h-11 w-full appearance-none rounded-2xl border border-border/80 bg-background px-4 pr-10 text-sm shadow-none outline-none transition-[border-color,box-shadow] focus:border-ring focus:ring-1 focus:ring-ring/60"
                  >
                    <option v-for="option in operatorOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                  <ChevronDown class="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                </div>

                <Input
                  v-model="condition.value"
                  placeholder="请输入条件值"
                  class="h-11 rounded-2xl shadow-none"
                />

                <Button
                  variant="ghost"
                  size="icon"
                  class="h-11 w-11 rounded-2xl text-muted-foreground hover:bg-destructive/8 hover:text-destructive"
                  :disabled="draft.conditions.length === 1"
                  @click="removeCondition(index)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </section>

          <section class="grid gap-3 rounded-[24px] border border-border bg-card p-4">
            <div class="flex items-center justify-between gap-4">
              <div>
                <h3 class="text-sm font-medium text-foreground">动作</h3>
                <p class="mt-1 text-xs text-muted-foreground">当前支持自动打标、加入收藏分组、移动到目录。</p>
              </div>
              <Button variant="outline" class="h-9 rounded-2xl shadow-none" @click="addAction">
                <Plus class="mr-2 h-4 w-4" />
                新增动作
              </Button>
            </div>

            <div class="grid gap-3">
              <div
                v-for="(action, index) in draft.actions"
                :key="`action-${index}`"
                class="grid gap-3 rounded-[20px] border border-border/80 bg-background p-3 md:grid-cols-[220px_minmax(0,1fr)_44px]"
              >
                <div class="relative">
                  <select
                    v-model="action.type"
                    class="h-11 w-full appearance-none rounded-2xl border border-border/80 bg-background px-4 pr-10 text-sm shadow-none outline-none transition-[border-color,box-shadow] focus:border-ring focus:ring-1 focus:ring-ring/60"
                  >
                    <option v-for="option in actionOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                  <ChevronDown class="pointer-events-none absolute right-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                </div>

                <Input
                  v-model="action.value"
                  :placeholder="actionPlaceholderMap.get(action.type) || '请输入动作目标'"
                  class="h-11 rounded-2xl shadow-none"
                />

                <Button
                  variant="ghost"
                  size="icon"
                  class="h-11 w-11 rounded-2xl text-muted-foreground hover:bg-destructive/8 hover:text-destructive"
                  :disabled="draft.actions.length === 1"
                  @click="removeAction(index)"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </div>
            </div>
          </section>
        </div>
      </div>

      <DialogFooter class="border-t border-border bg-background/95 px-6 py-4">
        <div class="flex w-full items-center justify-between gap-3">
          <Button
            v-if="isEditing"
            variant="ghost"
            class="rounded-2xl text-destructive hover:bg-destructive/8 hover:text-destructive"
            @click="requestDelete"
          >
            删除规则
          </Button>
          <div v-else />

          <div class="flex items-center gap-3">
            <Button variant="outline" class="rounded-2xl shadow-none" @click="emit('update:open', false)">
              取消
            </Button>
            <Button class="rounded-2xl" :disabled="saving" @click="saveRule">
              {{ saving ? '保存中...' : '保存规则' }}
            </Button>
          </div>
        </div>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
