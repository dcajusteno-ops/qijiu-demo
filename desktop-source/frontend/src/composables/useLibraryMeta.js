import { ref } from 'vue'
import * as App from '@/api'
import { normalizeFolderPath } from './useGalleryHelpers'

const tags = ref([])
const imageTags = ref({})
const imageNotes = ref({})
const favorites = ref(new Set())
const favoriteGroups = ref([])
const activeTagFilter = ref(null)

const getFavoritePathSet = (groups) => {
  const set = new Set()
  ;(groups || []).forEach((group) => {
    ;(group.paths || []).forEach((path) => {
      const normalized = normalizeFolderPath(path)
      if (normalized) set.add(normalized)
    })
  })
  return set
}

export function useLibraryMeta({ showToast = () => {}, confirm = async () => false, activeRoot, activeSub }) {
  const fetchFavorites = async () => {
    try {
      const groups = await App.GetFavoriteGroups()
      favoriteGroups.value = groups || []
      favorites.value = getFavoritePathSet(groups)
    } catch (e) {
      console.error(e)
    }
  }

  const fetchTags = async () => {
    try {
      const tgs = await App.GetTags()
      tags.value = tgs || []
    } catch (e) {
      console.error('Failed to fetch tags:', e)
    }
  }

  const fetchImageTags = async () => {
    try {
      const imgsTags = await App.GetImageTags()
      imageTags.value = imgsTags || {}
    } catch (e) {
      console.error('Failed to fetch image tags:', e)
    }
  }

  const fetchImageNotes = async () => {
    try {
      const notes = await App.GetImageNotes()
      imageNotes.value = notes || {}
    } catch (e) {
      console.error('Failed to fetch image notes:', e)
    }
  }

  const createTag = async (name, color, category = '') => {
    try {
      const newTag = await App.CreateTag(name, color, category)
      tags.value.push(newTag)
      showToast('标签已创建', 'success')
      return newTag
    } catch (e) {
      console.error(e)
      showToast('创建失败', 'error')
      return null
    }
  }

  const deleteTag = async (tagId) => {
    const ok = await confirm('确定要删除该标签吗？此操作将同时移除所有图片上的这个标签。')
    if (!ok) return

    try {
      await App.DeleteTag(tagId)
      tags.value = tags.value.filter((t) => t.id !== tagId)

      for (const relPath in imageTags.value) {
        imageTags.value[relPath] = imageTags.value[relPath].filter((id) => id !== tagId)
      }

      if (activeTagFilter.value === tagId) {
        activeTagFilter.value = null
      }

      showToast('标签已删除', 'success')
    } catch (e) {
      console.error(e)
      showToast('删除失败', 'error')
    }
  }

  const batchDeleteTags = async (tagIds) => {
    if (!tagIds || tagIds.length === 0) return

    const ok = await confirm(`确定要删除选中的 ${tagIds.length} 个标签吗？此操作将同时移除所有图片上的这些标签。`)
    if (!ok) return

    try {
      await App.BatchDeleteTags(tagIds)

      tags.value = tags.value.filter((t) => !tagIds.includes(t.id))
      for (const relPath in imageTags.value) {
        imageTags.value[relPath] = imageTags.value[relPath].filter((id) => !tagIds.includes(id))
      }
      if (tagIds.includes(activeTagFilter.value)) {
        activeTagFilter.value = null
      }

      showToast(`成功删除 ${tagIds.length} 个标签`, 'success')
    } catch (e) {
      console.error(e)
      showToast('删除失败', 'error')
    }
  }

  const batchUpdateTags = async (tagIds, data) => {
    if (!tagIds || tagIds.length === 0) return

    let successCount = 0
    let failCount = 0

    const promises = tagIds.map(async (tagId) => {
      try {
        await App.UpdateTag(tagId, data.name || null, data.color || null, data.category || null)
        const tag = tags.value.find((t) => t.id === tagId)
        if (tag) {
          if (data.name !== undefined) tag.name = data.name
          if (data.color !== undefined) tag.color = data.color
          if (data.category !== undefined) tag.category = data.category
        }
        successCount++
      } catch (e) {
        failCount++
      }
    })

    await Promise.all(promises)

    if (failCount === 0) {
      showToast(`成功更新 ${successCount} 个标签`, 'success')
    } else {
      showToast(`更新完成：${successCount} 成功，${failCount} 失败`, 'error')
    }
  }

  const updateTag = async (tagId, data) => {
    try {
      await App.UpdateTag(tagId, data.name || null, data.color || null, data.category || null)

      const tag = tags.value.find((t) => t.id === tagId)
      if (tag) {
        if (data.name !== undefined) tag.name = data.name
        if (data.color !== undefined) tag.color = data.color
        if (data.category !== undefined) tag.category = data.category
      }

      showToast('标签已更新', 'success')
      return true
    } catch (e) {
      console.error(e)
      showToast('更新失败', 'error')
      return false
    }
  }

  const addTagToImage = async (img, tagId) => {
    const relPath = img.relPath
    if (!imageTags.value[relPath]) {
      imageTags.value[relPath] = []
    }
    if (!imageTags.value[relPath].includes(tagId)) {
      imageTags.value[relPath].push(tagId)
    }

    try {
      await App.AddTagToImage(relPath, tagId)
    } catch (e) {
      console.error(e)
      imageTags.value[relPath] = imageTags.value[relPath].filter((id) => id !== tagId)
      showToast('添加标签失败', 'error')
    }
  }

  const removeTagFromImage = async (img, tagId) => {
    const relPath = img.relPath
    const originalTags = imageTags.value[relPath] || []
    if (imageTags.value[relPath]) {
      imageTags.value[relPath] = imageTags.value[relPath].filter((id) => id !== tagId)
    }

    try {
      await App.RemoveTagFromImage(relPath, tagId)
    } catch (e) {
      console.error(e)
      imageTags.value[relPath] = originalTags
      showToast('移除标签失败', 'error')
    }
  }

  const toggleTagFilter = (tagId) => {
    if (activeTagFilter.value === tagId) {
      activeTagFilter.value = null
    } else {
      activeTagFilter.value = tagId
    }
  }

  const getTagCount = (tagId) => {
    let count = 0
    for (const relPath in imageTags.value) {
      if (imageTags.value[relPath]?.includes(tagId)) {
        count++
      }
    }
    return count
  }

  const toggleFavorite = async (img) => {
    if (!img) return

    const path = normalizeFolderPath(img.relPath)
    const isFav = favorites.value.has(path)
    const activeFavoriteGroupId =
      activeRoot.value === 'favorites' && activeSub.value.startsWith('favorite-group:')
        ? activeSub.value.replace('favorite-group:', '')
        : ''

    try {
      if (isFav) {
        if (activeFavoriteGroupId) {
          await App.RemoveImageFromFavoriteGroup(img.relPath, activeFavoriteGroupId)
        } else {
          await App.RemoveFavorite(img.relPath)
        }
      } else {
        await App.AddImageToFavoriteGroup(img.relPath, activeFavoriteGroupId || 'default')
      }
      await fetchFavorites()
      img.isFavorite = favorites.value.has(normalizeFolderPath(img.relPath))
    } catch (e) {
      console.error(e)
      showToast('操作失败', 'error')
    }
  }

  return {
    tags,
    imageTags,
    imageNotes,
    favorites,
    favoriteGroups,
    activeTagFilter,
    fetchFavorites,
    fetchTags,
    fetchImageTags,
    fetchImageNotes,
    createTag,
    deleteTag,
    batchDeleteTags,
    batchUpdateTags,
    updateTag,
    addTagToImage,
    removeTagFromImage,
    toggleTagFilter,
    getTagCount,
    toggleFavorite,
  }
}
