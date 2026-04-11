# Implementation Plan for Favorite Groups Features

## Overview
This plan outlines the steps required to resolve two issues with the Favorite Groups functionality:
1. Empty favorite groups not rendering in the sidebar.
2. Providing a dialog to select favorite groups when clicking the heart icon on an ImageCard, instead of automatically adding to the default group.

## File Modifications

### 1. `frontend/src/composables/useImages.js`
**Goal:** Allow empty favorite groups to render in the sidebar.
*   **Locate:** The `pruneManagedNodes` function inside the `fileTree` computed property.
*   **Change:** Add a check `node.isFavoriteGroup` to bypass the pruning logic that hides empty nodes.

```javascript
// BEFORE
if (!hasImages && children.length === 0 && node.type !== 'root') {
  return null
}

// AFTER
if (!hasImages && children.length === 0 && node.type !== 'root' && !node.isFavoriteGroup) {
  return null
}
```

### 2. `frontend/src/components/ImageGallery.vue`
**Goal:** Integrate the `FavoriteGroupsDialog` and handle the new `manage-favorites` event from `ImageCard`.
*   **Import:** `FavoriteGroupsDialog` from `./FavoriteGroupsDialog.vue`.
*   **State:** Add `showFavoriteDialog` (boolean ref, default false) and `currentFavoriteImage` (object ref, default null).
*   **Template (Dialog):** Add the `<FavoriteGroupsDialog>` component to the template. Bind `open`, `groups`, and `image` props. Listen for `change` (to refresh favorites) and `update:open`.
*   **Template (ImageCard):** Change `@toggle-favorite="emit('toggle-favorite', img)"` to `@manage-favorites="openFavoriteDialog(img)"`. Keep `@toggle-favorite` if we still want direct toggling from other places (like lightbox), but update ImageCard to use `manage-favorites`.
*   **Methods:** Add `openFavoriteDialog(img)` method to set `currentFavoriteImage` and open the dialog.

### 3. `frontend/src/components/ImageCard.vue`
**Goal:** Change the Heart button behavior to open the dialog instead of instantly toggling.
*   **Emits:** Add `manage-favorites` to `defineEmits`.
*   **Template (Heart Button):** Change `@click.stop="$emit('toggle-favorite', image)"` to `@click.stop="$emit('manage-favorites', image)"`.
*   **Tooltip:** Update the tooltip text from `{{ image.isFavorite ? '取消收藏' : '添加到收藏' }}` to `'管理收藏'`. Keep the heart icon red if `image.isFavorite` is true.
