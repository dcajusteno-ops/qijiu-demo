package backend

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) BatchMove(paths []string, targetFolder string) (int, error) {
	targetFolder = filepath.Clean(targetFolder)

	var targetPath string
	isAbs := filepath.IsAbs(targetFolder)
	hasDrive := len(targetFolder) >= 2 && targetFolder[1] == ':'

	if isAbs || hasDrive {
		if strings.ContainsAny(targetFolder, "<>\"|?*") {
			return 0, fmt.Errorf("folder path contains invalid characters")
		}
		targetPath = targetFolder
	} else {
		targetFolder = normalizeRelPath(targetFolder)
		if strings.ContainsAny(targetFolder, "<>:\"|?*") {
			return 0, fmt.Errorf("folder name contains invalid characters")
		}
		var err error
		targetPath, err = a.resolveRootPath(targetFolder)
		if err != nil {
			return 0, err
		}
	}

	if !isSubPath(a.rootDir, filepath.Clean(targetPath)) {
		return 0, fmt.Errorf("target folder must stay inside root directory")
	}
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return 0, err
	}
	targetRel, err := filepath.Rel(a.rootDir, targetPath)
	if err != nil {
		return 0, err
	}
	normalizedTargetFolder := normalizeRelPath(targetRel)

	successCount := 0
	updated := false
	for _, relPath := range paths {
		if _, moved, err := a.moveManagedImageToFolder(relPath, normalizedTargetFolder); err == nil {
			successCount++
			updated = updated || moved
		}
	}
	if updated {
		a.scheduleImagesChangedEvent()
	}

	return successCount, nil
}

func (a *App) ExportImages(paths []string, targetDir string, move bool) (int, error) {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return 0, err
		}
	}

	successCount := 0
	for _, relPath := range paths {
		sourcePath, err := a.resolveRootPath(relPath)
		if err != nil {
			continue
		}
		fileName := filepath.Base(sourcePath)
		destPath := filepath.Join(targetDir, fileName)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue
		}

		srcFile, err := os.Open(sourcePath)
		if err != nil {
			continue
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			continue
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			continue
		}

		if move {
			os.Remove(sourcePath)
		}
		successCount++
	}

	return successCount, nil
}

type UploadResult struct {
	Count  int      `json:"count"`
	Errors []string `json:"errors"`
}

func (a *App) UploadImages(paths []string, targetFolder string) (*UploadResult, error) {
	var targetPath string
	if targetFolder == "" {
		targetPath = a.imageDir
	} else {
		var err error
		targetPath, err = a.resolveRootPath(targetFolder)
		if err != nil {
			return nil, fmt.Errorf("invalid target folder: %v", err)
		}
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return nil, fmt.Errorf("cannot create target directory: %v", err)
	}

	result := &UploadResult{}
	importedRelPaths := make([]string, 0, len(paths))
	for _, srcPath := range paths {
		info, err := os.Stat(srcPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: file not found", filepath.Base(srcPath)))
			continue
		}

		ext := strings.ToLower(filepath.Ext(srcPath))
		if !isImageExt(ext) {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: not an image file", filepath.Base(srcPath)))
			continue
		}

		if info.IsDir() {
			continue
		}

		fileName := filepath.Base(srcPath)
		destPath := filepath.Join(targetPath, fileName)

		if _, err := os.Stat(destPath); err == nil {
			name := strings.TrimSuffix(fileName, ext)
			timestamp := time.Now().Format("20060102_150405")
			destPath = filepath.Join(targetPath, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
		}

		srcFile, err := os.Open(srcPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: cannot open file", fileName))
			continue
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			result.Errors = append(result.Errors, fmt.Sprintf("%s: cannot create destination", fileName))
			continue
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: copy failed", fileName))
			continue
		}

		result.Count++
		if rel, err := filepath.Rel(a.rootDir, destPath); err == nil {
			importedRelPaths = append(importedRelPaths, normalizeRelPath(rel))
		}
	}
	if len(importedRelPaths) > 0 {
		a.scheduleAutoRulesRun(importedRelPaths)
		a.scheduleImagesChangedEvent()
	}

	return result, nil
}

func (a *App) SelectFolder() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Folder",
	}
	dir, err := runtime.OpenDirectoryDialog(a.ctx, options)
	return dir, err
}

func openDirectoryInExplorer(path string) error {
	normalized, err := normalizeExistingPath(path)
	if err != nil {
		return err
	}

	cmd := exec.Command("explorer", normalized)
	return cmd.Start()
}

func (a *App) OpenCurrentOutputDirectory() error {
	return openDirectoryInExplorer(a.imageDir)
}

func (a *App) OpenCurrentRootDirectory() error {
	return openDirectoryInExplorer(a.rootDir)
}

func (a *App) OpenImageLocation(relPath string) error {
	absPath, err := a.resolveRootPath(relPath)
	if err != nil {
		return fmt.Errorf("invalid path")
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	query := fmt.Sprintf("/select,%s", absPath)
	cmd := exec.Command("explorer", query)
	return cmd.Start()
}

func (a *App) OpenFile(relPath string) error {
	absPath, err := a.resolveRootPath(relPath)
	if err != nil {
		return fmt.Errorf("invalid path")
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found")
	}

	var cmd *exec.Cmd
	switch {
	case os.Getenv("OS") == "Windows_NT":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", absPath)
	default:
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", absPath)
	}

	return cmd.Start()
}

func (a *App) OrganizeFiles(mode string) (int, error) {
	organizedCount := 0
	entries, err := os.ReadDir(a.imageDir)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".gif" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		year := info.ModTime().Format("2006")
		month := info.ModTime().Format("01")

		targetRelDir := filepath.Join(year, month)
		if mode == "year" {
			targetRelDir = year
		}

		targetDir := filepath.Join(a.imageDir, targetRelDir)
		os.MkdirAll(targetDir, 0755)

		sourcePath := filepath.Join(a.imageDir, entry.Name())
		destPath := filepath.Join(targetDir, entry.Name())

		if _, err := os.Stat(destPath); err == nil {
			name := strings.TrimSuffix(entry.Name(), ext)
			timestamp := time.Now().Format("20060102_150405")
			destPath = filepath.Join(targetDir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
		}

		if err := moveFile(sourcePath, destPath); err == nil {
			organizedCount++
		}
	}

	return organizedCount, nil
}
