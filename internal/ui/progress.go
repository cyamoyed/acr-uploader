package ui

import (
	"fmt"
	"strings"
)

type ProgressBar struct {
	Total     int64
	Completed int64
	Width     int
}

func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		Total: total,
		Width: 50,
	}
}

func (pb *ProgressBar) Update(completed int64) {
	pb.Completed = completed
}

func (pb *ProgressBar) Render() string {
	if pb.Total == 0 {
		return "[] 0%"
	}
	
	percentage := float64(pb.Completed) / float64(pb.Total) * 100
	filled := int(float64(pb.Width) * percentage / 100)
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.Width-filled)
	
	return fmt.Sprintf("[%s] %.1f%%", bar, percentage)
}

func (pb *ProgressBar) Print() {
	fmt.Printf("\r%s", pb.Render())
}

func PrintUploadProgress(current, total int, layer string) {
	percentage := float64(current) / float64(total) * 100
	filled := int(percentage / 2)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 50-filled)
	
	fmt.Printf("\r上传进度: [%s] %d/%d (%.1f%%) - %s", bar, current, total, percentage, layer)
}

func PrintStatus(status string) {
	fmt.Printf("\r%s", status)
}

func PrintSuccess(message string) {
	fmt.Printf("\r\033[32m✓ %s\033[0m\n", message)
}

func PrintError(message string) {
	fmt.Printf("\r\033[31m✗ %s\033[0m\n", message)
}

func PrintInfo(message string) {
	fmt.Printf("\r\033[34mℹ %s\033[0m\n", message)
}

func PrintWarning(message string) {
	fmt.Printf("\r\033[33m⚠ %s\033[0m\n", message)
}
