package cmd

import (
	"fmt"
	"os"

	"acr-uploader/internal/auth"
	"acr-uploader/internal/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	configUsername         string
	configRegistry         string
	configNamespace        string
	configDefaultVersion   string
	configLogLevel         string
	configAccessKeyId      string
	configRegionId         string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置阿里云容器镜像服务",
	Long:  "配置阿里云容器镜像服务的用户名、仓库地址等信息",
	Run: func(cmd *cobra.Command, args []string) {
		if err := logger.InitLogger(configLogLevel); err != nil {
			logrus.Error(err)
			return
		}

		fmt.Print("请输入阿里云AccessKey Secret: ")
		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			logrus.Error(fmt.Sprintf("读取AccessKey Secret失败: %v", err))
			return
		}
		fmt.Println()

		config := &auth.Config{
			Username:         configUsername,
			Registry:         configRegistry,
			DefaultNamespace: configNamespace,
			DefaultVersion:   configDefaultVersion,
			LogLevel:         configLogLevel,
			AccessKeyId:      configAccessKeyId,
			AccessKeySecret:  string(password),
			RegionId:         configRegionId,
		}

		if config.DefaultVersion == "" {
			config.DefaultVersion = "latest"
		}

		if config.LogLevel == "" {
			config.LogLevel = "info"
		}

		if err := auth.SaveConfig(config); err != nil {
			logrus.Error(err)
			return
		}

		logrus.Info("配置成功")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVar(&configUsername, "username", "", "阿里云账号用户名")
	configCmd.Flags().StringVar(&configRegistry, "registry", "", "阿里云镜像仓库地址")
	configCmd.Flags().StringVar(&configNamespace, "namespace", "", "默认命名空间")
	configCmd.Flags().StringVar(&configDefaultVersion, "version", "", "默认版本号")
	configCmd.Flags().StringVar(&configLogLevel, "log-level", "info", "日志级别")
	configCmd.Flags().StringVar(&configAccessKeyId, "access-key-id", "", "阿里云AccessKey ID")
	configCmd.Flags().StringVar(&configRegionId, "region-id", "cn-hangzhou", "阿里云区域ID")

	configCmd.MarkFlagRequired("username")
	configCmd.MarkFlagRequired("registry")
	configCmd.MarkFlagRequired("access-key-id")
}