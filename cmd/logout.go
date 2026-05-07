package cmd

import (
	"acr-uploader/internal/auth"
	"acr-uploader/internal/logger"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "登出阿里云容器镜像服务",
	Long:  "登出阿里云容器镜像服务",
	Run: func(cmd *cobra.Command, args []string) {
		if err := logger.InitLogger("info"); err != nil {
			logrus.Error(err)
			return
		}
		
		if !auth.ConfigExists() {
			logrus.Error("配置文件不存在，请先执行 acr-uploader config")
			return
		}
		
		manager, err := auth.NewManager()
		if err != nil {
			logrus.Error(err)
			return
		}
		
		if err := manager.Logout(); err != nil {
			logrus.Error(err)
			return
		}
		
		logrus.Info("登出成功")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
