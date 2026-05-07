package image

import (
	"strings"
	"time"
)

type FilterOptions struct {
	Name   string
	Tag    string
	SizeMin int64
	SizeMax int64
	Since   time.Time
}

func FilterImages(images []*ImageInfo, options *FilterOptions) []*ImageInfo {
	if options == nil {
		return images
	}

	result := make([]*ImageInfo, 0)

	for _, img := range images {
		if options.Name != "" && !strings.Contains(img.Name, options.Name) {
			continue
		}

		if options.Tag != "" && img.Tag != options.Tag {
			continue
		}

		if options.SizeMin > 0 && img.Size < options.SizeMin {
			continue
		}

		if options.SizeMax > 0 && img.Size > options.SizeMax {
			continue
		}

		if !options.Since.IsZero() && img.Created < options.Since.Unix() {
			continue
		}

		result = append(result, img)
	}

	return result
}

func SortImages(images []*ImageInfo, by string, desc bool) {
	switch by {
	case "size":
		sortBySize(images, desc)
	case "created":
		sortByCreated(images, desc)
	case "name":
		sortByName(images, desc)
	default:
		sortByName(images, desc)
	}
}

func sortBySize(images []*ImageInfo, desc bool) {
	for i := 0; i < len(images)-1; i++ {
		for j := i + 1; j < len(images); j++ {
			shouldSwap := images[i].Size < images[j].Size
			if !desc {
				shouldSwap = !shouldSwap
			}
			if shouldSwap {
				images[i], images[j] = images[j], images[i]
			}
		}
	}
}

func sortByCreated(images []*ImageInfo, desc bool) {
	for i := 0; i < len(images)-1; i++ {
		for j := i + 1; j < len(images); j++ {
			shouldSwap := images[i].Created < images[j].Created
			if !desc {
				shouldSwap = !shouldSwap
			}
			if shouldSwap {
				images[i], images[j] = images[j], images[i]
			}
		}
	}
}

func sortByName(images []*ImageInfo, desc bool) {
	for i := 0; i < len(images)-1; i++ {
		for j := i + 1; j < len(images); j++ {
			shouldSwap := strings.Compare(images[i].Name, images[j].Name) < 0
			if !desc {
				shouldSwap = !shouldSwap
			}
			if shouldSwap {
				images[i], images[j] = images[j], images[i]
			}
		}
	}
}
