export const utilityMenuCatalog = [
  { id: 'settings', label: '设置', description: '打开设置中心' },
  { id: 'trash', label: '回收站管理', description: '查看和恢复回收站内容' },
  { id: 'documentation', label: '使用文档', description: '查看项目内置帮助' },
  { id: 'statistics', label: '数据视界', description: '进入统计分析页面' },
  { id: 'launcher', label: '外部工具', description: '启动外部工具入口' },
  { id: 'prompt-templates', label: '提示词模板', description: '管理常用提示词模板' },
  { id: 'auto-rules', label: '自动规则引擎', description: '查看和管理自动规则' },
  { id: 'open-output', label: '打开当前 output', description: '直接打开当前输出目录' },
  { id: 'switch-output', label: '切换 output 位置', description: '重新绑定 ComfyUI output 目录' },
  { id: 'custom-roots', label: '管理自定义目录', description: '整理侧边栏目录入口' },
]

export const buildUtilityMenuState = (items = []) => {
  const map = new Map((items || []).map((item) => [item.id, item]))
  return utilityMenuCatalog.map((entry, index) => {
    const current = map.get(entry.id)
    return {
      id: entry.id,
      label: entry.label,
      description: entry.description,
      visible: current?.visible !== false,
      order: Number(current?.order || index + 1),
    }
  }).sort((a, b) => a.order - b.order)
}
