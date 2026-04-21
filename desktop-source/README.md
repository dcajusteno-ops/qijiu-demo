# Desktop Source Layout

`desktop-source` is the Wails desktop application source for Comfy Manager.

## Current Go file grouping

The backend is no longer organized in the root directory.
It now lives under `desktop-source/backend/` and is grouped by filename prefix so related code stays easy to find.

- `backend/app.go`
  Minimal application shell. Keeps the `App` struct and shared runtime state only.
- `backend/app_core_*.go`
  Core runtime and lifecycle logic, such as constants, app startup, shutdown, and common runtime wiring.
- `backend/app_feature_*.go`
  Business features exposed by the desktop app, such as gallery, prompt library, favorites, tags, trash, profile, and settings.
- `backend/app_support_*.go`
  Internal helpers and infrastructure, such as paths, watchers, metadata cache, and compatibility helpers.
- `backend/app_types_*.go`
  Shared data structures split by domain to keep model definitions readable.
- `backend/exports.go`
  Small exported wrappers that let the root `main.go` wire Wails hooks without exposing internal layout details.

## Quick navigation

- `backend/app_core_constants.go`
  Shared constants used across multiple backend modules.
- `backend/app_core_lifecycle.go`
  `NewApp`, startup, and shutdown flow.
- `backend/app_core_runtime.go`
  Lightweight runtime helpers.
- `backend/app_feature_gallery.go`
  Gallery loading and image browsing related backend methods.
- `backend/app_feature_prompt.go`
  Prompt library and prompt assistant related backend methods.
- `backend/app_feature_settings.go`
  Settings persistence entry points.
- `backend/app_feature_profile.go`
  User profile loading and avatar management.
- `backend/app_feature_directory_binding.go`
  Output directory binding and directory configuration.
- `backend/app_feature_trash.go`
  Delete and trash behavior.
- `backend/app_support_metadata_cache.go`
  Image metadata cache helpers.
- `backend/app_support_watcher.go`
  File watching and refresh triggers.
- `backend/app_support_paths.go`
  Common path resolution helpers.

## Intent of this organization

- Keep the Wails entry layer small in the root directory.
- Move backend implementation into one dedicated folder without doing another deep feature split.
- Make related backend code easier to scan by directory plus filename.
- Leave room for future maintenance without re-introducing a giant single file in the root.

For the broader backend map, see [`../docs/BACKEND_FILE_MAP.md`](../docs/BACKEND_FILE_MAP.md).
