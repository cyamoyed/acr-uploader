package image

import (
	"fmt"

	"github.com/docker/docker/api/types/image"
)

type ImageInfo struct {
	ID       string
	Name     string
	Tag      string
	Size     int64
	Created  int64
	FullName string
}

func NewImageInfo(img image.Summary) *ImageInfo {
	var name, tag string
	if len(img.RepoTags) > 0 {
		parts := splitTag(img.RepoTags[0])
		name = parts[0]
		tag = parts[1]
	} else if len(img.RepoDigests) > 0 {
		name = img.RepoDigests[0]
		tag = "<none>"
	} else {
		name = "<none>"
		tag = "<none>"
	}

	return &ImageInfo{
		ID:       img.ID[:12],
		Name:     name,
		Tag:      tag,
		Size:     img.Size,
		Created:  img.Created,
		FullName: img.RepoTags[0],
	}
}

func splitTag(fullTag string) []string {
	for i := len(fullTag) - 1; i >= 0; i-- {
		if fullTag[i] == ':' {
			return []string{fullTag[:i], fullTag[i+1:]}
		}
	}
	return []string{fullTag, "latest"}
}

func (i *ImageInfo) GetSizeHuman() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case i.Size >= GB:
		return fmt.Sprintf("%.2fGB", float64(i.Size)/GB)
	case i.Size >= MB:
		return fmt.Sprintf("%.2fMB", float64(i.Size)/MB)
	case i.Size >= KB:
		return fmt.Sprintf("%.2fKB", float64(i.Size)/KB)
	default:
		return fmt.Sprintf("%dB", i.Size)
	}
}

func (i *ImageInfo) String() string {
	return fmt.Sprintf("%s:%s (%s)", i.Name, i.Tag, i.GetSizeHuman())
}
