package cmd

import (
	"fmt"

	"acr-uploader/internal/acr"
	"acr-uploader/internal/auth"
	"acr-uploader/internal/image"
	"acr-uploader/internal/logger"
	"acr-uploader/internal/ui"
	"acr-uploader/internal/uploader"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	uploadImage    string
	uploadVersion  string
	uploadForce    bool
	uploadResume   bool
	uploadQuiet    bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "上传镜像到阿里云容器镜像服务",
	Long:  "上传Docker镜像到阿里云容器镜像服务，支持交互式选择和断点续传",
	Run: func(cmd *cobra.Command, args []string) {
		if err := logger.InitLogger("info"); err != nil {
			logrus.Error(err)
			return
		}

		if !auth.ConfigExists() {
			logrus.Error("配置文件不存在，请先执行 acr-uploader config")
			return
		}

		config, err := auth.LoadConfig()
		if err != nil {
			logrus.Error(err)
			return
		}

		imgManager, err := image.NewManager()
		if err != nil {
			logrus.Error(err)
			return
		}
		defer imgManager.Close()

		var selectedImage *image.ImageInfo

		if uploadImage != "" {
			img, err := imgManager.GetImageByName(uploadImage)
			if err != nil {
				img, err = imgManager.GetImageByID(uploadImage)
				if err != nil {
					logrus.Error(err)
					return
				}
			}
			selectedImage = img
		} else {
			images, err := imgManager.ListImages(nil)
			if err != nil {
				logrus.Error(err)
				return
			}

			if len(images) == 0 {
				logrus.Info("本地没有Docker镜像")
				return
			}

			selectedImage, err = ui.SelectImageInteractive(images)
			if err != nil {
				logrus.Error(err)
				return
			}

			if selectedImage == nil {
				logrus.Info("未选择任何镜像")
				return
			}
		}

		fmt.Println()
		fmt.Printf("您选择了: %s:%s\n", selectedImage.Name, selectedImage.Tag)
		fmt.Printf("镜像ID: %s\n", selectedImage.ID)
		fmt.Printf("大小: %s\n", selectedImage.GetSizeHuman())

		if !confirmUpload() {
			logrus.Info("取消上传")
			return
		}

		if config.AccessKeyId == "" || config.AccessKeySecret == "" || config.RegionId == "" {
			logrus.Error("请先配置阿里云AccessKey和Region，执行 acr-uploader config")
			return
		}

		acrClient, err := acr.NewClient(config.RegionId, config.AccessKeyId, config.AccessKeySecret)
		if err != nil {
			logrus.Error(fmt.Sprintf("创建ACR客户端失败: %v", err))
			return
		}

		fmt.Println("\n正在获取命名空间列表...")
		namespaces, err := acrClient.ListNamespaces()
		if err != nil {
			logrus.Error(fmt.Sprintf("获取命名空间列表失败: %v", err))
			return
		}

		if len(namespaces) == 0 {
			logrus.Error("当前账号下没有可用的命名空间，请先在阿里云控制台创建")
			return
		}

		selectedNamespace, err := ui.SelectNamespace(namespaces)
		if err != nil {
			logrus.Error(fmt.Sprintf("选择命名空间失败: %v", err))
			return
		}

		fmt.Printf("\n已选择命名空间: %s\n", selectedNamespace.Name)

		fmt.Println("\n正在获取镜像仓库列表...")
		repositories, err := acrClient.ListRepositories(selectedNamespace.Name)
		if err != nil {
			logrus.Error(fmt.Sprintf("获取镜像仓库列表失败: %v", err))
			return
		}

		var selectedRepo *acr.Repository
		for {
			selectedRepo, err = ui.SelectRepository(repositories, selectedNamespace.Name)
			if err != nil {
				logrus.Error(fmt.Sprintf("选择镜像仓库失败: %v", err))
				return
			}

			if selectedRepo != nil {
				break
			}

			fmt.Println("\n--- 创建新镜像仓库 ---")
			repoName, err := ui.InputNewRepositoryName()
			if err != nil {
				logrus.Error(fmt.Sprintf("输入仓库名称失败: %v", err))
				continue
			}

			repoType, err := ui.SelectRepositoryType()
			if err != nil {
				logrus.Error(fmt.Sprintf("选择仓库类型失败: %v", err))
				continue
			}

			fmt.Printf("\n正在创建仓库: %s/%s (%s)...\n", selectedNamespace.Name, repoName, repoType)
			selectedRepo, err = acrClient.CreateRepository(selectedNamespace.Name, repoName, repoType)
			if err != nil {
				logrus.Error(fmt.Sprintf("创建仓库失败: %v", err))
				continue
			}

			fmt.Printf("仓库创建成功: %s/%s\n", selectedRepo.Namespace, selectedRepo.Name)
			break
		}

		fmt.Printf("\n已选择镜像仓库: %s/%s\n", selectedRepo.Namespace, selectedRepo.Name)

		version := uploadVersion
		if version == "" {
			version = config.DefaultVersion
			if version == "" {
				version = "latest"
			}
		}

		targetTag := buildTargetTag(selectedImage, selectedRepo.Namespace, selectedRepo.Name, version)
		fmt.Printf("\n准备上传到: %s\n", targetTag)

		uploaderInst, err := uploader.NewUploader(config.Registry, selectedRepo.Namespace)
		if err != nil {
			logrus.Error(err)
			return
		}
		defer uploaderInst.Close()

		options := &uploader.UploadOptions{
			Version:   version,
			Registry:  config.Registry,
			Namespace: selectedRepo.Namespace,
			Repository: selectedRepo.Name,
			Force:     uploadForce,
			Resume:    uploadResume,
			Quiet:     uploadQuiet,
		}

		fmt.Println()
		ui.PrintInfo("开始上传镜像...")

		if err := uploaderInst.Upload(selectedImage.ID, options); err != nil {
			ui.PrintError(fmt.Sprintf("上传失败: %v", err))
			return
		}

		ui.PrintSuccess(fmt.Sprintf("上传成功!"))
		fmt.Printf("\n镜像信息:\n")
		fmt.Printf("  镜像名称: %s\n", selectedImage.Name)
		fmt.Printf("  标签: %s\n", version)
		fmt.Printf("  存储路径: %s\n", targetTag)
		fmt.Println()
		logrus.Info("上传任务完成")
	},
}

func confirmUpload() bool {
	fmt.Print("\n确认上传此镜像? (y/n): ")
	var input string
	fmt.Scanln(&input)
	return input == "y" || input == "Y" || input == "yes" || input == "YES"
}

func buildTargetTag(img *image.ImageInfo, namespace, repository, version string) string {
	return fmt.Sprintf("%s/%s/%s:%s", "registry.cn-hangzhou.aliyuncs.com", namespace, repository, version)
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVarP(&uploadImage, "image", "i", "", "镜像ID或名称")
	uploadCmd.Flags().StringVarP(&uploadVersion, "version", "v", "", "目标版本号")
	uploadCmd.Flags().BoolVarP(&uploadForce, "force", "f", false, "强制覆盖已存在标签")
	uploadCmd.Flags().BoolVarP(&uploadResume, "resume", "R", false, "启用断点续传")
	uploadCmd.Flags().BoolVarP(&uploadQuiet, "quiet", "q", false, "静默模式")
}