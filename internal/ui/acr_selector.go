package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"acr-uploader/internal/acr"
)

const (
	ActionCreateNewRepo = -1
)

func SelectNamespace(namespaces []acr.Namespace) (*acr.Namespace, error) {
	if len(namespaces) == 0 {
		return nil, fmt.Errorf("没有可用的命名空间")
	}

	printNamespaceTable(namespaces)

	fmt.Println()
	fmt.Print("请选择命名空间序号: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取输入失败: %w", err)
	}

	idx, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("无效的序号: %s", input)
	}

	if idx < 1 || idx > len(namespaces) {
		return nil, fmt.Errorf("序号超出范围: %d", idx)
	}

	return &namespaces[idx-1], nil
}

func printNamespaceTable(namespaces []acr.Namespace) {
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    阿里云ACR命名空间列表                      │")
	fmt.Println("├──────┬──────────────────────────────────────┬──────┬────────┤")
	fmt.Println("│ 序号 │ 命名空间名称                           │ 状态    │ 地域    │")
	fmt.Println("├──────┼──────────────────────────────────────┼──────┼────────┤")

	for i, ns := range namespaces {
		name := ns.Name
		if len(name) > 36 {
			name = name[:33] + "..."
		}

		fmt.Printf("│ %4d │ %-36s │ %6s │ %8s │\n",
			i+1,
			name,
			ns.Status,
			ns.RegionId)
	}

	fmt.Println("└──────┴──────────────────────────────────────┴──────┴────────┘")
}

func SelectRepository(repositories []acr.Repository, namespace string) (*acr.Repository, error) {
	printRepositoryTable(repositories, namespace)

	fmt.Println()
	fmt.Println("  输入序号选择镜像仓库，或输入 -1 创建新仓库")
	fmt.Print("请选择: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取输入失败: %w", err)
	}

	idx, err := strconv.Atoi(input)
	if err != nil {
		return nil, fmt.Errorf("无效的序号: %s", input)
	}

	if idx == ActionCreateNewRepo {
		return nil, nil
	}

	if idx < 1 || idx > len(repositories) {
		return nil, fmt.Errorf("序号超出范围: %d", idx)
	}

	return &repositories[idx-1], nil
}

func printRepositoryTable(repositories []acr.Repository, namespace string) {
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Printf("│              命名空间 [%s] 下的镜像仓库列表                   │\n", namespace)
	fmt.Println("├──────┬──────────────────────────────────────┬──────┬────────┤")
	fmt.Println("│ 序号 │ 仓库名称                             │ 类型    │ 状态    │")
	fmt.Println("├──────┼──────────────────────────────────────┼──────┼────────┤")

	if len(repositories) == 0 {
		fmt.Println("│      │              当前命名空间下无镜像仓库                │      │")
	} else {
		for i, repo := range repositories {
			name := repo.Name
			if len(name) > 36 {
				name = name[:33] + "..."
			}

			fmt.Printf("│ %4d │ %-36s │ %6s │ %8s │\n",
				i+1,
				name,
				repo.RepoType,
				repo.Status)
		}
	}

	fmt.Println("└──────┴──────────────────────────────────────┴──────┴────────┘")
}

func InputNewRepositoryName() (string, error) {
	fmt.Print("\n请输入新镜像仓库名称: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name := strings.TrimSpace(scanner.Text())

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}

	if name == "" {
		return "", fmt.Errorf("仓库名称不能为空")
	}

	return name, nil
}

func SelectRepositoryType() (string, error) {
	fmt.Println("\n请选择仓库类型:")
	fmt.Println("  1. PUBLIC - 公开仓库")
	fmt.Println("  2. PRIVATE - 私有仓库")
	fmt.Print("请选择 (1/2): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取输入失败: %w", err)
	}

	switch input {
	case "1", "PUBLIC", "public":
		return "PUBLIC", nil
	case "2", "PRIVATE", "private":
		return "PRIVATE", nil
	default:
		return "", fmt.Errorf("无效的选择: %s", input)
	}
}