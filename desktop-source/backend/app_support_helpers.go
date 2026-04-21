package backend

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func moveFile(sourcePath, destPath string) error {
	err := os.Rename(sourcePath, destPath)
	if err == nil {
		return nil
	}
	input, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer output.Close()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	input.Close()
	output.Close()
	return os.Remove(sourcePath)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func uniqueNonEmptyStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func normalizeSearchValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func collapseSpaces(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func prettifyAssetLabel(value string) string {
	base := filepath.Base(strings.TrimSpace(value))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, "_", " ")
	return strings.TrimSpace(base)
}

func normalizeAssetKey(value string) string {
	normalized := strings.ToLower(collapseSpaces(prettifyAssetLabel(value)))
	normalized = strings.ReplaceAll(normalized, "-", " ")
	return collapseSpaces(normalized)
}

func extractDateKeyFromRelPath(relPath string, modTime time.Time) string {
	parts := strings.Split(normalizeRelPath(relPath), "/")
	for _, part := range parts {
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, part); matched {
			return part
		}
	}
	if modTime.IsZero() {
		return ""
	}
	return modTime.Format("2006-01-02")
}

func parseDate(value string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func matchesDatePreset(dateKey, preset, start, end string) bool {
	if strings.TrimSpace(dateKey) == "" {
		return false
	}
	dateValue := parseDate(dateKey)
	if dateValue.IsZero() {
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch strings.TrimSpace(strings.ToLower(preset)) {
	case "", "all":
		return true
	case "today":
		return dateValue.Equal(today)
	case "yesterday":
		return dateValue.Equal(today.AddDate(0, 0, -1))
	case "last7":
		startDate := today.AddDate(0, 0, -6)
		return !dateValue.Before(startDate) && !dateValue.After(today)
	case "month":
		return dateValue.Year() == today.Year() && dateValue.Month() == today.Month()
	case "custom":
		startDate := parseDate(start)
		endDate := parseDate(end)
		if !startDate.IsZero() && dateValue.Before(startDate) {
			return false
		}
		if !endDate.IsZero() && dateValue.After(endDate) {
			return false
		}
		return true
	default:
		return true
	}
}

func stripUTF8BOM(data []byte) []byte {
	return bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
}

func repairLegacyMojibake(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	encoded, err := simplifiedchinese.GBK.NewEncoder().Bytes([]byte(trimmed))
	if err != nil || !utf8.Valid(encoded) {
		return trimmed
	}

	repaired := strings.TrimSpace(string(encoded))
	if repaired == "" {
		return trimmed
	}

	return repaired
}

func isSupportedProfileImageExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return true
	default:
		return false
	}
}

func copyFile(sourcePath, destPath string) error {
	input, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return err
	}

	return output.Close()
}
