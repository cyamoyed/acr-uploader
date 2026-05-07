package image

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	client *client.Client
}

func NewManager() (*Manager, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建Docker客户端失败: %w", err)
	}

	return &Manager{client: cli}, nil
}

func (m *Manager) ListImages(options *FilterOptions) ([]*ImageInfo, error) {
	logrus.Info("正在获取本地Docker镜像列表")

	images, err := m.client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}

	result := make([]*ImageInfo, 0)
	for _, img := range images {
		if len(img.RepoTags) == 0 {
			continue
		}
		result = append(result, NewImageInfo(img))
	}

	if options != nil {
		result = FilterImages(result, options)
	}

	SortImages(result, "created", true)

	logrus.Info(fmt.Sprintf("共找到 %d 个镜像", len(result)))
	return result, nil
}

func (m *Manager) GetImageByID(imageID string) (*ImageInfo, error) {
	images, err := m.client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}

	for _, img := range images {
		if len(img.RepoTags) == 0 {
			continue
		}
		if img.ID == imageID || len(imageID) <= 12 && img.ID[:len(imageID)] == imageID {
			return NewImageInfo(img), nil
		}
	}

	return nil, fmt.Errorf("镜像不存在: %s", imageID)
}

func (m *Manager) GetImageByName(name string) (*ImageInfo, error) {
	images, err := m.client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}

	for _, img := range images {
		if len(img.RepoTags) == 0 {
			continue
		}
		for _, tag := range img.RepoTags {
			if tag == name {
				return NewImageInfo(img), nil
			}
		}
	}

	return nil, fmt.Errorf("镜像不存在: %s", name)
}

func (m *Manager) Close() {
	if m.client != nil {
		m.client.Close()
	}
}

func (m *Manager) GetAPIVersion() string {
	return m.client.ClientVersion()
}
