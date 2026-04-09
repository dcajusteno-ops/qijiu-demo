import { computed } from 'vue'

export function useImageStacks(images, isStackingEnabled) {
  const stackedImages = computed(() => {
    if (!images.value || images.value.length === 0) return []
    
    if (!isStackingEnabled.value) {
      return images.value
    }
    
    // Group images by Prompt and Model or filename heuristic
    const groups = {}
    
    images.value.forEach(img => {
      // 1. Try to group by exact Prompt + Model (if available from backend)
      // We added Prompt and Model to ImageFile in app.go
      let groupKey = ''
      
      if (img.prompt && img.prompt.trim() !== '') {
        // Has metadata, use Prompt + Model as key
        // hash or string concat
        groupKey = `meta_${img.model || 'unknown'}_${img.prompt.substring(0, 100)}`
      } else {
        // 2. Fallback to filename heuristic if no metadata
        let baseName = img.name
        const match = img.name.match(/^(.*?)_?(\d+)?_?(\.\w+)$/)
        if (match && match[1]) {
          baseName = match[1] + match[3] 
        }
        groupKey = `name_${baseName}`
      }
      
      if (!groups[groupKey]) {
        groups[groupKey] = []
      }
      groups[groupKey].push(img)
    })
    
    const result = []
    Object.values(groups).forEach(group => {
      if (group.length > 1) {
        // Sort group by modTime desc
        group.sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
        
        // Add the primary image, with children attached
        const primary = { ...group[0], stackChildren: group.slice(1), isStackPrimary: true, stackCount: group.length }
        result.push(primary)
      } else {
        result.push(group[0])
      }
    })
    
    // Sort overall result by modTime desc
    result.sort((a, b) => new Date(b.modTime) - new Date(a.modTime))
    
    return result
  })
  
  return {
    stackedImages
  }
}
