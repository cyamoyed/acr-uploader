package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "acr-uploader",
	Short: "阿里云容器镜像服务上传工具",
	Long: `acr-uploader 是一个用于上传Docker镜像到阿里云容器镜像服务(ACR)的命令行工具。

功能特性：
- 简化Docker镜像上传流程
- 支持交互式镜像选择
- 支持断点续传
- 完善的错误处理和日志记录`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
