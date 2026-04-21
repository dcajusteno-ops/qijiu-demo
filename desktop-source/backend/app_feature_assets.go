package backend

import (
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

func scaledImageDimensions(width, height, maxDimension int) (int, int) {
	if width <= 0 || height <= 0 || maxDimension <= 0 {
		return 0, 0
	}
	if width <= maxDimension && height <= maxDimension {
		return width, height
	}
	if width >= height {
		targetWidth := maxDimension
		targetHeight := int(float64(height) * float64(maxDimension) / float64(width))
		if targetHeight < 1 {
			targetHeight = 1
		}
		return targetWidth, targetHeight
	}
	targetHeight := maxDimension
	targetWidth := int(float64(width) * float64(maxDimension) / float64(height))
	if targetWidth < 1 {
		targetWidth = 1
	}
	return targetWidth, targetHeight
}

func (a *App) ensureImageVariant(kind, relPath string) (string, error) {
	sourceRelPath := normalizeRelPath(relPath)
	if sourceRelPath == "" {
		return "", fmt.Errorf("variant source path is empty")
	}

	variantDir, maxDimension, err := a.variantSpec(kind)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(variantDir, 0755); err != nil {
		return "", err
	}

	sourcePath, err := a.resolveRootPath(sourceRelPath)
	if err != nil {
		return "", err
	}
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", err
	}

	variantPath := filepath.Join(variantDir, imageVariantFilename(kind, sourceRelPath))
	if variantInfo, statErr := os.Stat(variantPath); statErr == nil && !sourceInfo.ModTime().After(variantInfo.ModTime()) {
		return variantPath, nil
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	sourceImage, _, err := image.Decode(sourceFile)
	if err != nil {
		return "", err
	}

	bounds := sourceImage.Bounds()
	targetWidth, targetHeight := scaledImageDimensions(bounds.Dx(), bounds.Dy(), maxDimension)
	if targetWidth <= 0 || targetHeight <= 0 {
		return "", fmt.Errorf("invalid image dimensions")
	}

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), sourceImage, bounds, draw.Over, nil)

	tempPath := variantPath + "." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".tmp"
	outputFile, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	encodeErr := png.Encode(outputFile, dst)
	closeErr := outputFile.Close()
	if encodeErr != nil {
		_ = os.Remove(tempPath)
		return "", encodeErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return "", closeErr
	}

	_ = os.Remove(variantPath)
	if err := os.Chtimes(tempPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Rename(tempPath, variantPath); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}

	return variantPath, nil
}

func (a *App) warmImageVariantsAsync(items []ImageFile, settings GalleryPerformanceSettings) {
	if !settings.BackgroundVariantWarmup || len(items) == 0 {
		return
	}

	targets := make([]ImageFile, len(items))
	copy(targets, items)

	go func() {
		previewLimit := 6
		if settings.InitialBatchSize > 0 && settings.InitialBatchSize < previewLimit {
			previewLimit = settings.InitialBatchSize
		}
		if settings.PageSize > 0 && settings.PageSize < previewLimit {
			previewLimit = settings.PageSize
		}
		if previewLimit < 1 {
			previewLimit = 1
		}

		for index, item := range targets {
			if _, err := a.ensureImageVariant("thumb", item.RelPath); err != nil {
				continue
			}
			if index < previewLimit {
				_, _ = a.ensureImageVariant("preview", item.RelPath)
			}
		}
	}()
}

func isLegacyProfileAssetPath(relPath string) bool {
	base := strings.ToLower(filepath.Base(filepath.FromSlash(relPath)))
	if !strings.HasPrefix(base, "profile-image.") {
		return false
	}
	return isSupportedProfileImageExt(filepath.Ext(base))
}

func (a *App) serveImage(w http.ResponseWriter, r *http.Request) {
	path := normalizeRelPath(strings.TrimPrefix(r.URL.Path, "/"))

	var (
		absPath string
		err     error
	)

	switch {
	case strings.HasPrefix(path, profileAssetPrefix):
		absPath, err = a.resolveProfileAssetPath(path)
	case strings.HasPrefix(path, variantAssetPrefix):
		kind, sourceRelPath, resolveErr := a.resolveVariantSource(path)
		if resolveErr != nil {
			err = resolveErr
			break
		}
		absPath, err = a.ensureImageVariant(kind, sourceRelPath)
	case strings.HasPrefix(path, trashAssetPrefix):
		absPath, err = a.resolveTrashAssetPath(path)
	default:
		absPath, err = a.resolveRootPath(path)
		if err != nil && isLegacyProfileAssetPath(path) {
			absPath, err = a.resolveProfileAssetPath(profileAssetPrefix + filepath.Base(filepath.FromSlash(path)))
		}
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid Path"))
		return
	}

	http.ServeFile(w, r, absPath)
}
