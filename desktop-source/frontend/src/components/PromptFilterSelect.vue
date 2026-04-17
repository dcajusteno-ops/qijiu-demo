<script setup>
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ChevronDown } from 'lucide-vue-next'

const props = defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: {
    type: Array,
    default: () => [],
  },
  placeholder: { type: String, default: '请选择' },
  triggerClass: { type: String, default: '' },
  contentClass: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue'])

const normalizedValue = computed(() => String(props.modelValue ?? ''))

const selectedLabel = computed(() => {
  const selected = props.options.find((item) => String(item.value ?? '') === normalizedValue.value)
  return selected?.label || props.placeholder
})

const updateValue = (value) => {
  emit('update:modelValue', value)
}
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger asChild>
      <Button
        variant="outline"
        class="min-w-0 justify-between rounded-xl px-3 font-normal"
        :class="triggerClass"
      >
        <span class="truncate text-left">{{ selectedLabel }}</span>
        <ChevronDown class="size-4 text-muted-foreground" />
      </Button>
    </DropdownMenuTrigger>

    <DropdownMenuContent
      align="start"
      :side-offset="6"
      :collision-padding="12"
      class="w-[min(24rem,calc(100vw-2rem))] rounded-xl p-1"
      :class="contentClass"
    >
      <ScrollArea class="max-h-72">
        <DropdownMenuRadioGroup
          :model-value="normalizedValue"
          @update:model-value="updateValue"
        >
          <DropdownMenuRadioItem
            v-for="item in options"
            :key="`${item.value}`"
            :value="String(item.value ?? '')"
            class="rounded-lg pr-3"
          >
            <span class="truncate">{{ item.label }}</span>
          </DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </ScrollArea>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
