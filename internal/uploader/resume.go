package uploader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type UploadProgress struct {
	ImageID        string   `json:"image_id"`
	Version        string   `json:"version"`
	UploadedLayers []string `json:"uploaded_layers"`
	TotalLayers    int      `json:"total_layers"`
	CreatedAt      int64    `json:"created_at"`
}

func getProgressDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户目录: %w", err)
	}
	return filepath.Join(homeDir, ".acr-uploader", "upload-progress"), nil
}

func getProgressPath(imageID, version string) (string, error) {
	dir, err := getProgressDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s.json", imageID, version)), nil
}

func LoadUploadProgress(imageID, version string) (*UploadProgress, error) {
	path, err := getProgressPath(imageID, version)
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &UploadProgress{
				ImageID:        imageID,
				Version:        version,
				UploadedLayers: []string{},
			}, nil
		}
		return nil, fmt.Errorf("读取上传进度失败: %w", err)
	}
	
	var progress UploadProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("解析上传进度失败: %w", err)
	}
	
	return &progress, nil
}

func SaveUploadProgress(progress *UploadProgress) error {
	dir, err := getProgressDir()
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建进度目录失败: %w", err)
	}
	
	path, err := getProgressPath(progress.ImageID, progress.Version)
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(progress, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化进度失败: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入进度文件失败: %w", err)
	}
	
	return nil
}

func DeleteUploadProgress(imageID, version string) error {
	path, err := getProgressPath(imageID, version)
	if err != nil {
		return err
	}
	
	if err := os.Remove(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("删除进度文件失败: %w", err)
		}
	}
	
	return nil
}

func filterPendingLayers(layers []string, uploadedLayers []string) []string {
	pending := make([]string, 0)
	uploadedSet := make(map[string]bool)
	
	for _, layer := range uploadedLayers {
		uploadedSet[layer] = true
	}
	
	for _, layer := range layers {
		if !uploadedSet[layer] {
			pending = append(pending, layer)
		}
	}
	
	logrus.Info(fmt.Sprintf("已上传 %d 层，剩余 %d 层待上传", len(uploadedLayers), len(pending)))
	return pending
}
