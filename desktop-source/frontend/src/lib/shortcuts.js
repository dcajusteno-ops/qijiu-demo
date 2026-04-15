export const shortcutActionCatalog = [
  {
    id: 'switch_dashboard',
    label: '切换到总览',
    description: '快速进入工作台总览页面。',
    defaultAccelerator: 'Ctrl+Alt+1',
  },
  {
    id: 'switch_gallery',
    label: '切换到图库',
    description: '快速回到主图库视图。',
    defaultAccelerator: 'Ctrl+Alt+2',
  },
  {
    id: 'switch_favorites',
    label: '切换到收藏',
    description: '快速打开收藏图片视图。',
    defaultAccelerator: 'Ctrl+Alt+3',
  },
  {
    id: 'switch_documentation',
    label: '切换到文档',
    description: '快速打开内置使用文档。',
    defaultAccelerator: 'Ctrl+Alt+4',
  },
  {
    id: 'refresh_images',
    label: '刷新图库',
    description: '立即重新加载图片和元数据。',
    defaultAccelerator: 'Ctrl+Alt+R',
  },
  {
    id: 'toggle_sidebar',
    label: '切换侧边栏',
    description: '折叠或展开左侧导航栏。',
    defaultAccelerator: 'Ctrl+Alt+B',
  },
  {
    id: 'toggle_selection_mode',
    label: '切换批量模式',
    description: '进入或退出批量选择模式。',
    defaultAccelerator: 'Ctrl+Alt+M',
  },
]

export const shortcutActionMap = shortcutActionCatalog.reduce((acc, item) => {
  acc[item.id] = item
  return acc
}, {})

export const modifierKeys = new Set(['Control', 'Shift', 'Alt', 'Meta'])

const specialKeyMap = {
  ' ': 'Space',
  Spacebar: 'Space',
  Enter: 'Enter',
  Escape: 'Escape',
  Esc: 'Escape',
  Tab: 'Tab',
  Delete: 'Delete',
  Backspace: 'Backspace',
  Insert: 'Insert',
  Home: 'Home',
  End: 'End',
  PageUp: 'PageUp',
  PageDown: 'PageDown',
  ArrowUp: 'Up',
  ArrowDown: 'Down',
  ArrowLeft: 'Left',
  ArrowRight: 'Right',
}

export const normalizeAccelerator = (accelerator) => {
  const tokens = String(accelerator || '')
    .split('+')
    .map((token) => token.trim())
    .filter(Boolean)

  let hasCtrl = false
  let hasAlt = false
  let hasShift = false
  let hasWin = false
  let key = ''

  tokens.forEach((token) => {
    const lower = token.toLowerCase()
    if (lower === 'ctrl' || lower === 'control') {
      hasCtrl = true
      return
    }
    if (lower === 'alt') {
      hasAlt = true
      return
    }
    if (lower === 'shift') {
      hasShift = true
      return
    }
    if (lower === 'win' || lower === 'meta' || lower === 'cmd' || lower === 'super') {
      hasWin = true
      return
    }
    key = token.toUpperCase()
  })

  const parts = []
  if (hasCtrl) parts.push('Ctrl')
  if (hasAlt) parts.push('Alt')
  if (hasShift) parts.push('Shift')
  if (hasWin) parts.push('Win')
  if (key) parts.push(key)
  return parts.join('+')
}

export const formatShortcutLabel = (accelerator) => {
  if (!accelerator) return '未设置'
  return normalizeAccelerator(accelerator)
}

export const buildBindingsFromCatalog = (bindings = []) => {
  const savedMap = new Map(
    (bindings || []).map((binding) => [binding.action, normalizeAccelerator(binding.accelerator)])
  )

  return shortcutActionCatalog.map((action) => ({
    action: action.id,
    accelerator: savedMap.has(action.id)
      ? savedMap.get(action.id)
      : normalizeAccelerator(action.defaultAccelerator),
  }))
}

export const getAcceleratorFromEvent = (event) => {
  const rawKey = event.key
  if (!rawKey || modifierKeys.has(rawKey)) {
    return ''
  }

  let key = ''
  if (specialKeyMap[rawKey]) {
    key = specialKeyMap[rawKey]
  } else if (/^F\d{1,2}$/i.test(rawKey)) {
    key = rawKey.toUpperCase()
  } else if (/^[a-z0-9]$/i.test(rawKey)) {
    key = rawKey.toUpperCase()
  } else {
    return ''
  }

  const parts = []
  if (event.ctrlKey) parts.push('Ctrl')
  if (event.altKey) parts.push('Alt')
  if (event.shiftKey) parts.push('Shift')
  if (event.metaKey) parts.push('Win')
  parts.push(key)
  return normalizeAccelerator(parts.join('+'))
}

export const findDuplicateAccelerators = (bindings = []) => {
  const owners = new Map()
  const duplicates = new Set()

  bindings.forEach((binding) => {
    const accelerator = normalizeAccelerator(binding.accelerator)
    if (!accelerator) return
    if (owners.has(accelerator)) {
      duplicates.add(accelerator)
      return
    }
    owners.set(accelerator, binding.action)
  })

  return duplicates
}
