const getApp = () => {
  const app = window.go?.main?.App
  if (!app) {
    throw new Error('Comfy Manager 仅支持桌面端运行。')
  }
  return app
}

const callApp = (method, ...args) => {
  const app = getApp()
  const fn = app[method]
  if (typeof fn !== 'function') {
    throw new Error(`桌面端接口缺失: ${method}`)
  }
  return fn(...args)
}

export const GetCustomRoots = async () => callApp('GetCustomRoots')
export const AddCustomRoot = async (name, path, icon = '') => callApp('AddCustomRoot', name, path, icon)
export const UpdateCustomRoot = async (id, name, icon = '') => callApp('UpdateCustomRoot', id, name, icon)
export const DeleteCustomRoot = async (id) => callApp('DeleteCustomRoot', id)
export const SelectFolder = async () => callApp('SelectFolder')
export const GetRelativePath = async (path) => callApp('GetRelativePath', path)

export const GetFavorites = async () => callApp('GetFavorites')
export const RemoveFavorite = async (path) => callApp('RemoveFavorite', path)
export const AddFavorite = async (path) => callApp('AddFavorite', path)
export const GetFavoriteGroups = async () => callApp('GetFavoriteGroups')
export const CreateFavoriteGroup = async (name) => callApp('CreateFavoriteGroup', name)
export const UpdateFavoriteGroup = async (id, name) => callApp('UpdateFavoriteGroup', id, name)
export const DeleteFavoriteGroup = async (id) => callApp('DeleteFavoriteGroup', id)
export const SetImageFavoriteGroups = async (path, groupIDs) => callApp('SetImageFavoriteGroups', path, groupIDs)
export const AddImageToFavoriteGroup = async (path, groupID) => callApp('AddImageToFavoriteGroup', path, groupID)
export const RemoveImageFromFavoriteGroup = async (path, groupID) => callApp('RemoveImageFromFavoriteGroup', path, groupID)

export const GetImages = async (sortBy, sortOrder) => callApp('GetImages', sortBy, sortOrder)
export const GetImageMetadata = async (relPath) => callApp('GetImageMetadata', relPath)
export const DeleteImage = async (path) => callApp('DeleteImage', path)
export const CopyText = async (text) => callApp('CopyText', text)

export const GetTags = async () => callApp('GetTags')
export const GetImageTags = async () => callApp('GetImageTags')
export const CreateTag = async (name, color, category) => callApp('CreateTag', name, color, category)
export const DeleteTag = async (tagId) => callApp('DeleteTag', tagId)
export const BatchDeleteTags = async (tagIds) => callApp('BatchDeleteTags', tagIds)
export const UpdateTag = async (tagId, name, color, category) => callApp('UpdateTag', tagId, name, color, category)
export const AddTagToImage = async (relPath, tagId) => callApp('AddTagToImage', relPath, tagId)
export const RemoveTagFromImage = async (relPath, tagId) => callApp('RemoveTagFromImage', relPath, tagId)

export const OpenImageLocation = async (path) => callApp('OpenImageLocation', path)
export const OpenFile = async (path) => callApp('OpenFile', path)

export const GetTrashList = async () => callApp('GetTrashList')
export const GetTrashSettings = async () => callApp('GetTrashSettings')
export const SaveTrashSettings = async (settings) => callApp('SaveTrashSettings', settings)
export const GetUserProfile = async () => callApp('GetUserProfile')
export const SaveUserProfile = async (profile) => callApp('SaveUserProfile', profile)
export const SelectUserProfileImage = async () => callApp('SelectUserProfileImage')
export const ClearUserProfileImage = async () => callApp('ClearUserProfileImage')
export const RestoreTrash = async (filename) => callApp('RestoreTrash', filename)
export const BatchRestoreTrash = async (filenames) => callApp('BatchRestoreTrash', filenames)
export const BatchDeleteTrash = async (filenames) => callApp('BatchDeleteTrash', filenames)
export const EmptyTrash = async () => callApp('EmptyTrash')
export const GetShortcutSettings = async () => callApp('GetShortcutSettings')
export const SaveShortcutSettings = async (settings) => callApp('SaveShortcutSettings', settings)
export const GetShortcutActions = async () => callApp('GetShortcutActions')

export const CleanupTags = async () => callApp('CleanupTags')
export const GetStatistics = async (mode) => callApp('GetStatistics', mode)

export const GetLauncherTools = async () => callApp('GetLauncherTools')
export const UpdateLauncherTool = async (id, data) => callApp('UpdateLauncherTool', id, data)
export const AddLauncherTool = async (data) => callApp('AddLauncherTool', data)
export const DeleteLauncherTool = async (id) => callApp('DeleteLauncherTool', id)
export const RunLauncherTool = async (id) => callApp('RunLauncherTool', id)
export const ExtractIcon = async (path) => callApp('ExtractIcon', path)

export const GetPromptToolLinks = async () => callApp('GetPromptToolLinks')
export const AddPromptToolLink = async (data) => callApp('AddPromptToolLink', data)
export const UpdatePromptToolLink = async (id, data) => callApp('UpdatePromptToolLink', id, data)
export const DeletePromptToolLink = async (id) => callApp('DeletePromptToolLink', id)

export const ExportImages = async (paths, targetDir, move) => callApp('ExportImages', paths, targetDir, move)
export const UploadImages = async (paths, targetFolder) => callApp('UploadImages', paths, targetFolder)
export const BatchAddTag = async (paths, tagId) => callApp('BatchAddTag', paths, tagId)
export const BatchRemoveTag = async (paths, tagId) => callApp('BatchRemoveTag', paths, tagId)
export const BatchMove = async (paths, targetFolder) => callApp('BatchMove', paths, targetFolder)
export const BatchFavorites = async (paths, action) => callApp('BatchFavorites', paths, action)
export const CleanEmptyFolders = async () => callApp('CleanEmptyFolders')
export const ClearPreviewCache = async () => callApp('ClearPreviewCache')
export const OrganizeFiles = async (mode) => callApp('OrganizeFiles', mode)

// Image Notes
export const GetImageNotes = async () => callApp('GetImageNotes')
export const SetImageNote = async (relPath, note) => callApp('SetImageNote', relPath, note)
export const DeleteImageNote = async (relPath) => callApp('DeleteImageNote', relPath)

// Smart Albums
export const GetSmartAlbumFields = async () => callApp('GetSmartAlbumFields')
export const GetSmartAlbums = async (field) => callApp('GetSmartAlbums', field)

// Prompt Templates
export const GetPromptTemplates = async () => callApp('GetPromptTemplates')
export const AddPromptTemplate = async (data) => callApp('AddPromptTemplate', data)
export const UpdatePromptTemplate = async (id, data) => callApp('UpdatePromptTemplate', id, data)
export const DeletePromptTemplate = async (id) => callApp('DeletePromptTemplate', id)
