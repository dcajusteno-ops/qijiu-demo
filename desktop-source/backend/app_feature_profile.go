package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetUserProfile() (UserProfile, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultUserProfile(), err
	}
	return settings.UserProfile, nil
}

func (a *App) SaveUserProfile(profile UserProfile) (UserProfile, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultUserProfile(), err
	}

	settings.UserProfile = normalizeUserProfile(profile)
	if err := a.saveSettings(settings); err != nil {
		return defaultUserProfile(), err
	}

	return settings.UserProfile, nil
}

func (a *App) saveUserProfileImage(sourcePath string) (UserProfile, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultUserProfile(), err
	}

	ext := strings.ToLower(filepath.Ext(sourcePath))
	if !isSupportedProfileImageExt(ext) {
		return settings.UserProfile, fmt.Errorf("unsupported image format")
	}

	if err := os.MkdirAll(a.profileImageDir(), 0755); err != nil {
		return settings.UserProfile, err
	}

	existing, _ := filepath.Glob(filepath.Join(a.profileImageDir(), "profile-image.*"))
	for _, item := range existing {
		if err := os.Remove(item); err != nil && !os.IsNotExist(err) {
			return settings.UserProfile, err
		}
	}

	targetName := "profile-image" + ext
	targetPath := filepath.Join(a.profileImageDir(), targetName)
	if err := copyFile(sourcePath, targetPath); err != nil {
		return settings.UserProfile, err
	}

	settings.UserProfile.ImagePath = normalizeRelPath(profileAssetPrefix + targetName)
	settings.UserProfile = normalizeUserProfile(settings.UserProfile)
	if err := a.saveSettings(settings); err != nil {
		return settings.UserProfile, err
	}

	return settings.UserProfile, nil
}

func (a *App) ClearUserProfileImage() (UserProfile, error) {
	settings, err := a.loadSettings()
	if err != nil {
		return defaultUserProfile(), err
	}

	existing, _ := filepath.Glob(filepath.Join(a.profileImageDir(), "profile-image.*"))
	for _, item := range existing {
		if err := os.Remove(item); err != nil && !os.IsNotExist(err) {
			return settings.UserProfile, err
		}
	}

	settings.UserProfile.ImagePath = ""
	settings.UserProfile = normalizeUserProfile(settings.UserProfile)
	if err := a.saveSettings(settings); err != nil {
		return settings.UserProfile, err
	}

	return settings.UserProfile, nil
}

func (a *App) SelectUserProfileImage() (UserProfile, error) {
	options := runtime.OpenDialogOptions{
		Title: "选择一张新的头像图片",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Image Files (*.png;*.jpg;*.jpeg;*.webp;*.gif)",
				Pattern:     "*.png;*.jpg;*.jpeg;*.webp;*.gif",
			},
		},
	}

	filePath, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return defaultUserProfile(), err
	}

	if strings.TrimSpace(filePath) == "" {
		return a.GetUserProfile()
	}

	return a.saveUserProfileImage(filePath)
}
