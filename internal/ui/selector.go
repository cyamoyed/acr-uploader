package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"acr-uploader/internal/image"
)

func SelectImages(images []*image.ImageInfo) ([]*image.ImageInfo, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("没有可用的镜像")
	}
	
	PrintImageTable(images)
	
	fmt.Println()
	fmt.Print("请选择要上传的镜像序号 (输入序号，可多选用逗号分隔): ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取输入失败: %w", err)
	}
	
	selected := make([]*image.ImageInfo, 0)
	
	if input == "" {
		return selected, nil
	}
	
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		idx, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("无效的序号: %s", part)
		}
		
		if idx < 1 || idx > len(images) {
			return nil, fmt.Errorf("序号超出范围: %d", idx)
		}
		
		selected = append(selected, images[idx-1])
	}
	
	return selected, nil
}

func PrintImageTable(images []*image.ImageInfo) {
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    本地Docker镜像列表                        │")
	fmt.Println("├──────┬──────────────────────────────────────┬──────┬────────┤")
	fmt.Println("│ 序号 │ 镜像名称                               │ 标签    │ 大小    │")
	fmt.Println("├──────┼──────────────────────────────────────┼──────┼────────┤")
	
	for i, img := range images {
		name := img.Name
		if len(name) > 36 {
			name = name[:33] + "..."
		}
		
		fmt.Printf("│ %4d │ %-36s │ %6s │ %8s │\n", 
			i+1, 
			name, 
			img.Tag, 
			img.GetSizeHuman())
	}
	
	fmt.Println("└──────┴──────────────────────────────────────┴──────┴────────┘")
}

func ConfirmAction(message string) bool {
	fmt.Print(message + " (y/n): ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	
	return input == "y" || input == "yes"
}

func GetInput(prompt string) string {
	fmt.Print(prompt + ": ")
	
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
