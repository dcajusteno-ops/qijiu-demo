<script setup>
import { computed, ref, watch } from 'vue'
import { Bookmark, ChevronDown, ChevronUp, Copy, FileJson, Info, Loader2, StickyNote } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import * as App from '@/api'
import PromptTemplateDialog from './PromptTemplateDialog.vue'

const props = defineProps({
  currentDisplayImage: { type: Object, default: null },
  imageNotes: { type: Object, default: () => ({}) },
  metadata: { type: Object, default: null },
  metadataLoading: { type: Boolean, default: false },
  metadataError: { type: String, default: '' },
})

const promptTemplateDialogOpen = ref(false)
const promptTemplateInitialContent = ref('')
const promptTemplateInitialType = ref('')
const noteText = ref('')
const noteSaving = ref(false)
const noteExpanded = ref(true)
const promptDebugExpanded = ref(false)

const currentNote = computed(() => {
  if (!props.currentDisplayImage) return ''
  return props.imageNotes[props.currentDisplayImage.relPath] || ''
})

const hasNote = computed(() => currentNote.value.trim() !== '')

const metadataFacts = computed(() => {
  if (!props.metadata) return []

  const facts = [
    { label: '模型', value: props.metadata.model },
    { label: '采样器', value: props.metadata.sampler },
    { label: '调度器', value: props.metadata.scheduler },
    { label: 'Seed', value: props.metadata.seed },
    { label: 'Steps', value: props.metadata.steps },
    { label: 'CFG', value: props.metadata.cfg },
  ]

  if (props.metadata.width && props.metadata.height) {
    facts.push({ label: '尺寸', value: `${props.metadata.width} × ${props.metadata.height}` })
  }

  if (props.metadata.nodeCount) {
    facts.push({ label: 'Workflow 节点', value: `${props.metadata.nodeCount}` })
  }

  return facts.filter((item) => item.value)
})

const extraMetadataEntries = computed(() => Object.entries(props.metadata?.extraFields || {}))

const promptDebugSections = computed(() => {
  const promptDebug = props.metadata?.promptDebug
  if (!promptDebug) return []

  return [
    { key: 'positive', label: '正向解析', data: promptDebug.positive },
    { key: 'negative', label: '反向解析', data: promptDebug.negative },
  ].filter((section) => section.data?.selectedText || section.data?.candidates?.length)
})

watch(
  () => [props.currentDisplayImage?.relPath, props.imageNotes],
  () => {
    noteText.value = currentNote.value
  },
  { immediate: true, deep: true },
)

const saveNote = async () => {
  if (!props.currentDisplayImage) return
  const text = noteText.value
  noteSaving.value = true
  try {
    await App.SetImageNote(props.currentDisplayImage.relPath, text)
    if (text.trim()) {
      props.imageNotes[props.currentDisplayImage.relPath] = text
    } else {
      delete props.imageNotes[props.currentDisplayImage.relPath]
    }
  } catch (e) {
    console.error('Failed to save note:', e)
  } finally {
    noteSaving.value = false
  }
}

const handleNoteBlur = () => {
  if (noteText.value !== currentNote.value) {
    saveNote()
  }
}

const handleNoteKeydown = (e) => {
  if (e.ctrlKey && (e.key === 'Enter' || e.key === 's')) {
    e.preventDefault()
    handleNoteBlur()
  }
}

const copyMetadataField = async (value, label) => {
  if (!value) {
    toast.error(`${label}为空`)
    return
  }

  try {
    await App.CopyText(value)
    toast.success(`${label}已复制`)
  } catch (error) {
    toast.error(`复制失败：${error?.message || error}`)
  }
}

const openPromptTemplateDialog = (content, type) => {
  promptTemplateInitialContent.value = content || ''
  promptTemplateInitialType.value = type || ''
  promptTemplateDialogOpen.value = true
}

const promptDebugStrategyLabel = (strategy) => {
  const labels = {
    'preferred-key': '优先字段',
    'semantic-key': '语义字段',
    'mce-config': '多角色编辑器',
    'direct-value': '直接值',
    'fallback/fallback-positive-node': '全图兜底',
    'fallback/fallback-negative-node': '全图兜底',
  }

  if (!strategy) return '未标记'
  if (labels[strategy]) return labels[strategy]
  if (strategy.startsWith('sdxl-tuple/')) {
    const inner = strategy.slice('sdxl-tuple/'.length)
    return `SDXL 回退 / ${labels[inner] || inner}`
  }
  if (strategy.startsWith('fallback/')) {
    const inner = strategy.slice('fallback/'.length)
    return `全图兜底 / ${labels[inner] || inner}`
  }
  return labels[strategy] || strategy
}

const promptDebugSourceLabel = (candidate) => {
  if (!candidate) return ''
  const parts = [candidate.sourceTitle, candidate.sourceClass].filter(Boolean)
  const nodeLabel = candidate.sourceNodeId ? `#${candidate.sourceNodeId}` : ''
  return [nodeLabel, ...parts].filter(Boolean).join(' ')
}
</script>

<template>
  <div
    class="absolute bottom-10 top-[90px] right-8 z-[60] flex w-[360px] flex-col overflow-hidden rounded-xl border border-white/10 bg-black/70 text-white shadow-2xl backdrop-blur-xl"
    @click.stop
    @wheel.stop
  >
    <div class="flex items-center gap-3 border-b border-white/10 px-4 py-3">
      <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-white/10">
        <Info class="h-4 w-4 text-blue-200" />
      </div>
      <div class="min-w-0">
        <div class="text-sm font-semibold tracking-wide">PNG 元数据</div>
        <div class="text-[11px] text-white/50">
          {{ metadata?.hasMetadata ? '可查看并复制 prompt / workflow' : '按需读取当前图片信息' }}
        </div>
      </div>
    </div>

    <ScrollArea class="min-h-0 flex-1">
      <div class="space-y-3 p-4">
        <div class="overflow-hidden rounded-lg border border-white/10 bg-white/5">
          <button
            class="flex w-full items-center justify-between px-3 py-2 transition-colors hover:bg-white/5"
            @click="noteExpanded = !noteExpanded"
          >
            <div class="flex items-center gap-2">
              <StickyNote class="h-3.5 w-3.5" :class="hasNote ? 'text-amber-300' : 'text-white/45'" />
              <span class="text-[11px] font-semibold uppercase tracking-wider" :class="hasNote ? 'text-amber-300' : 'text-white/45'">笔记</span>
            </div>
            <ChevronUp v-if="noteExpanded" class="h-3.5 w-3.5 text-white/45" />
            <ChevronDown v-else class="h-3.5 w-3.5 text-white/45" />
          </button>
          <div v-if="noteExpanded" class="px-3 pb-3">
            <textarea
              v-model="noteText"
              class="w-full resize-none rounded-md border border-white/10 bg-black/30 px-3 py-2 text-sm text-white/90 placeholder:text-white/30 focus:outline-none focus:ring-1 focus:ring-white/20"
              rows="3"
              placeholder="添加笔记... (Ctrl+S / Ctrl+Enter 保存)"
              @blur="handleNoteBlur"
              @keydown="handleNoteKeydown"
            />
            <div v-if="noteSaving" class="mt-1 text-[10px] text-white/40">保存中...</div>
            <div v-else-if="hasNote" class="mt-1 text-[10px] text-white/40">已保存</div>
          </div>
        </div>

        <div v-if="metadataLoading" class="flex items-center gap-2 text-sm text-white/70">
          <Loader2 class="h-4 w-4 animate-spin" />
          <span>正在读取图片元数据...</span>
        </div>

        <div v-else-if="metadataError" class="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-200">
          {{ metadataError }}
        </div>

        <div v-else class="space-y-3">
          <div v-if="metadataFacts.length > 0" class="grid grid-cols-2 gap-2">
            <div
              v-for="fact in metadataFacts"
              :key="fact.label"
              class="rounded-lg border border-white/10 bg-white/5 px-3 py-2"
            >
              <div class="text-[11px] uppercase tracking-wider text-white/45">{{ fact.label }}</div>
              <div class="mt-1 truncate text-sm text-white/90" :title="fact.value">{{ fact.value }}</div>
            </div>
          </div>

          <div v-if="metadata?.loras?.length" class="space-y-2">
            <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">LoRA</div>
            <div class="flex flex-wrap gap-2">
              <Badge
                v-for="lora in metadata.loras"
                :key="lora"
                variant="secondary"
                class="max-w-full border-white/10 bg-white/10 text-white/85"
              >
                <span class="truncate">{{ lora }}</span>
              </Badge>
            </div>
          </div>

          <div v-if="metadata?.hasMetadata || extraMetadataEntries.length > 0" class="space-y-4 rounded-lg border border-white/10 bg-white/5 p-3">
            <div v-if="metadata?.positive" class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">正向 Prompt</div>
                <div class="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                    title="存为模板"
                    @click="openPromptTemplateDialog(metadata.positive, 'positive')"
                  >
                    <Bookmark class="h-3.5 w-3.5" />
                    存为模板
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                    @click="copyMetadataField(metadata.positive, '正向 Prompt')"
                  >
                    <Copy class="h-3.5 w-3.5" />
                    复制
                  </Button>
                </div>
              </div>
              <ScrollArea class="h-40 rounded-md border border-white/10 bg-black/20" @wheel.stop>
                <div class="whitespace-pre-wrap break-words p-3 text-sm leading-6 text-white/88">
                  {{ metadata.positive }}
                </div>
              </ScrollArea>
            </div>

            <div v-if="metadata?.negative" class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">反向 Prompt</div>
                <div class="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                    title="存为模板"
                    @click="openPromptTemplateDialog(metadata.negative, 'negative')"
                  >
                    <Bookmark class="h-3.5 w-3.5" />
                    存为模板
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                    @click="copyMetadataField(metadata.negative, '反向 Prompt')"
                  >
                    <Copy class="h-3.5 w-3.5" />
                    复制
                  </Button>
                </div>
              </div>
              <ScrollArea class="h-40 rounded-md border border-white/10 bg-black/20" @wheel.stop>
                <div class="whitespace-pre-wrap break-words p-3 text-sm leading-6 text-white/88">
                  {{ metadata.negative }}
                </div>
              </ScrollArea>
            </div>

            <div v-if="metadata?.prompt" class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">ComfyUI 提示词</div>
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                  @click="copyMetadataField(metadata.prompt, 'Prompt JSON')"
                >
                  <FileJson class="h-3.5 w-3.5" />
                  复制 JSON
                </Button>
              </div>
              <div class="text-xs leading-5 text-white/60">
                已检测到 ComfyUI prompt 节点图，可直接复制 JSON 用于恢复或分析。
              </div>
              <ScrollArea
                v-if="!metadata?.positive && !metadata?.negative"
                class="h-32 rounded-md border border-white/10 bg-black/20"
              >
                <div class="whitespace-pre-wrap break-words p-3 font-mono text-xs leading-5 text-white/75">
                  {{ metadata.prompt }}
                </div>
              </ScrollArea>
            </div>

            <div v-if="metadata?.workflow" class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">工作流</div>
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-7 gap-1.5 px-2 text-white/75 hover:bg-white/10 hover:text-white"
                  @click="copyMetadataField(metadata.workflow, 'Workflow JSON')"
                >
                  <FileJson class="h-3.5 w-3.5" />
                  复制 JSON
                </Button>
              </div>
              <div class="text-xs leading-5 text-white/60">
                {{ metadata.nodeCount ? `已检测到 ${metadata.nodeCount} 个 workflow 节点，可直接复制回 ComfyUI。` : '已检测到 workflow JSON，可直接复制回 ComfyUI。' }}
              </div>
            </div>

            <div v-if="promptDebugSections.length > 0" class="space-y-3 rounded-lg border border-white/10 bg-black/15 p-3">
              <button
                type="button"
                class="flex w-full items-center justify-between gap-3 text-left"
                @click="promptDebugExpanded = !promptDebugExpanded"
              >
                <div>
                  <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">提示词解析调试</div>
                  <div class="mt-1 text-xs leading-5 text-white/55">查看本次命中的节点来源、策略和候选排序。</div>
                </div>
                <ChevronUp v-if="promptDebugExpanded" class="h-4 w-4 text-white/45" />
                <ChevronDown v-else class="h-4 w-4 text-white/45" />
              </button>

              <div v-if="promptDebugExpanded" class="space-y-3">
                <div
                  v-for="section in promptDebugSections"
                  :key="section.key"
                  class="min-w-0 space-y-2 overflow-hidden rounded-lg border border-white/10 bg-white/5 p-3"
                >
                  <div class="flex min-w-0 items-center justify-between gap-3">
                    <div class="min-w-0 text-xs font-semibold tracking-wide text-white/80">{{ section.label }}</div>
                    <div class="text-[11px] text-white/45">{{ promptDebugStrategyLabel(section.data?.strategy) }}</div>
                  </div>

                  <div v-if="section.data?.selectedText" class="min-w-0 space-y-1 overflow-hidden rounded-md border border-emerald-400/20 bg-emerald-400/5 p-2">
                    <div class="text-[11px] uppercase tracking-wider text-emerald-100/70">当前命中</div>
                    <div class="whitespace-pre-wrap break-all text-xs leading-5 text-white/90">{{ section.data.selectedText }}</div>
                    <div v-if="promptDebugSourceLabel(section.data)" class="whitespace-pre-wrap break-all text-[11px] leading-5 text-white/50">
                      来源：{{ promptDebugSourceLabel(section.data) }}
                      <span v-if="section.data?.sourceKey"> / {{ section.data.sourceKey }}</span>
                    </div>
                  </div>

                  <div v-if="section.data?.candidates?.length" class="space-y-2">
                    <div class="text-[11px] uppercase tracking-wider text-white/45">候选排序</div>
                    <div class="space-y-2">
                      <div
                        v-for="candidate in section.data.candidates.slice(0, 5)"
                        :key="`${section.key}-${candidate.text}-${candidate.score}`"
                        class="min-w-0 overflow-hidden rounded-md border border-white/10 bg-black/20 p-2"
                      >
                        <div class="flex min-w-0 items-center justify-between gap-3">
                          <div class="min-w-0 text-[11px] text-white/55">{{ promptDebugStrategyLabel(candidate.strategy) }}</div>
                          <div class="text-[11px] text-white/40">分数 {{ candidate.score }}</div>
                        </div>
                        <div class="mt-1 whitespace-pre-wrap break-all text-xs leading-5 text-white/85">{{ candidate.text }}</div>
                        <div v-if="promptDebugSourceLabel(candidate)" class="mt-1 whitespace-pre-wrap break-all text-[11px] leading-5 text-white/45">
                          来源：{{ promptDebugSourceLabel(candidate) }}
                          <span v-if="candidate.sourceKey"> / {{ candidate.sourceKey }}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div
              v-if="!metadata?.positive && !metadata?.negative && metadata?.extraFields?.parameters"
              class="space-y-2"
            >
              <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">原始参数</div>
              <ScrollArea class="h-32 rounded-md border border-white/10 bg-black/20">
                <div class="whitespace-pre-wrap break-words p-3 text-xs leading-5 text-white/80">
                  {{ metadata.extraFields.parameters }}
                </div>
              </ScrollArea>
            </div>

            <div v-if="extraMetadataEntries.length > 0" class="space-y-2">
              <div class="text-[11px] font-semibold uppercase tracking-wider text-white/45">其他字段</div>
              <div class="space-y-2">
                <div
                  v-for="[key, value] in extraMetadataEntries"
                  :key="key"
                  class="rounded-lg border border-white/10 bg-black/20 px-3 py-2"
                >
                  <div class="flex items-center justify-between gap-3">
                    <div class="text-[11px] text-white/55">{{ key }}</div>
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-6 gap-1 px-2 text-white/65 hover:bg-white/10 hover:text-white"
                      @click="copyMetadataField(value, key)"
                    >
                      <Copy class="h-3 w-3" />
                      复制
                    </Button>
                  </div>
                  <div class="mt-2 max-h-16 overflow-hidden whitespace-pre-wrap break-words text-xs leading-5 text-white/70">
                    {{ value }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="rounded-lg border border-white/10 bg-white/5 px-3 py-4 text-sm leading-6 text-white/60">
            当前图片没有检测到可读取的 ComfyUI PNG 元数据。
          </div>
        </div>
      </div>
    </ScrollArea>
  </div>

  <PromptTemplateDialog
    v-model:open="promptTemplateDialogOpen"
    :initial-content="promptTemplateInitialContent"
    :initial-type="promptTemplateInitialType"
    :initial-source-path="currentDisplayImage?.relPath || ''"
  />
</template>
