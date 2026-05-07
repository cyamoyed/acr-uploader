package cmd

import (
	"acr-uploader/internal/image"
	"acr-uploader/internal/logger"
	"acr-uploader/internal/ui"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	listFilterName string
	listFilterTag  string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出本地Docker镜像",
	Long:  "列出本地Docker镜像，支持按名称和标签筛选",
	Run: func(cmd *cobra.Command, args []string) {
		if err := logger.InitLogger("info"); err != nil {
			logrus.Error(err)
			return
		}
		
		imgManager, err := image.NewManager()
		if err != nil {
			logrus.Error(err)
			return
		}
		defer imgManager.Close()
		
		filterOptions := &image.FilterOptions{
			Name: listFilterName,
			Tag:  listFilterTag,
		}
		
		images, err := imgManager.ListImages(filterOptions)
		if err != nil {
			logrus.Error(err)
			return
		}
		
		if len(images) == 0 {
			logrus.Info("没有找到匹配的镜像")
			return
		}
		
		ui.PrintImageTable(images)
		
		logrus.Info("镜像列表展示完成")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	
	listCmd.Flags().StringVar(&listFilterName, "filter-name", "", "按名称筛选")
	listCmd.Flags().StringVar(&listFilterTag, "filter-tag", "", "按标签筛选")
}
