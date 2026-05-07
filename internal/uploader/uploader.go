package uploader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type UploadOptions struct {
	Version     string
	Registry    string
	Namespace   string
	Repository  string
	Force       bool
	Resume      bool
	Quiet       bool
}

type Uploader struct {
	client        *client.Client
	tagNormalizer *TagNormalizer
}

func NewUploader(registry, namespace string) (*Uploader, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("创建Docker客户端失败: %w", err)
	}

	return &Uploader{
		client:        cli,
		tagNormalizer: NewTagNormalizer(registry, namespace),
	}, nil
}

func (u *Uploader) Upload(imageID string, options *UploadOptions) error {
	if options.Version == "" {
		options.Version = "latest"
	}

	logrus.Info(fmt.Sprintf("开始上传镜像: %s, 版本: %s", imageID, options.Version))

	img, err := u.getImageByID(imageID)
	if err != nil {
		return fmt.Errorf("获取镜像信息失败: %w", err)
	}

	var targetTag string
	if options.Repository != "" {
		targetTag, err = u.tagNormalizer.NormalizeTagWithRepo(img.RepoTags[0], options.Repository, options.Version)
	} else {
		targetTag, err = u.tagNormalizer.NormalizeTag(img.RepoTags[0], options.Version)
	}
	if err != nil {
		return fmt.Errorf("标签规范化失败: %w", err)
	}

	if err := u.tagNormalizer.TagImage(img.RepoTags[0], targetTag); err != nil {
		return err
	}

	if options.Resume {
		return u.UploadWithResume(imageID, targetTag, options)
	}

	return u.pushImage(targetTag, options)
}

func (u *Uploader) getImageByID(imageID string) (*image.Summary, error) {
	images, err := u.client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, img := range images {
		if img.ID == imageID || len(imageID) <= 12 && img.ID[:len(imageID)] == imageID {
			return &img, nil
		}
	}

	return nil, fmt.Errorf("镜像不存在: %s", imageID)
}

func (u *Uploader) getImageLayers(imageID string) ([]string, error) {
	inspect, _, err := u.client.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		return nil, err
	}

	var layers []string
	for _, layer := range inspect.RootFS.Layers {
		layers = append(layers, layer)
	}

	return layers, nil
}

func (u *Uploader) UploadWithResume(imageID, targetTag string, options *UploadOptions) error {
	logrus.Info("启用断点续传模式")

	progress, err := LoadUploadProgress(imageID, options.Version)
	if err != nil {
		logrus.Warn(fmt.Sprintf("读取上传进度失败，将重新上传: %v", err))
		return u.pushImage(targetTag, options)
	}

	layers, err := u.getImageLayers(imageID)
	if err != nil {
		return fmt.Errorf("获取镜像层失败: %w", err)
	}

	pendingLayers := filterPendingLayers(layers, progress.UploadedLayers)

	if len(pendingLayers) == 0 {
		logrus.Info("镜像已全部上传完成")
		return nil
	}

	logrus.Info(fmt.Sprintf("继续上传，剩余 %d 层", len(pendingLayers)))

	if err := u.pushImage(targetTag, options); err != nil {
		return err
	}

	if err := DeleteUploadProgress(imageID, options.Version); err != nil {
		logrus.Warn(fmt.Sprintf("删除进度文件失败: %v", err))
	}

	return nil
}

var (
	progressRegex = regexp.MustCompile(`(\d+)%`)
	layerRegex    = regexp.MustCompile(`([a-f0-9]{64})`)
)

func (u *Uploader) pushImage(targetTag string, options *UploadOptions) error {
	logrus.Info(fmt.Sprintf("正在推送镜像: %s", targetTag))

	cmd := exec.Command("docker", "push", targetTag)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建输出管道失败: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建错误管道失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动推送命令失败: %w", err)
	}

	if !options.Quiet {
		go u.handlePushOutputWithProgress(stdout)
		go u.handlePushOutputWithProgress(stderr)
	} else {
		go io.Copy(io.Discard, stdout)
		go io.Copy(io.Discard, stderr)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("推送失败: %w", err)
	}

	logrus.Info("镜像推送成功")
	
	if !options.Quiet {
		if err := u.verifyPush(targetTag); err != nil {
			logrus.Warn(fmt.Sprintf("验证推送结果失败: %v", err))
		} else {
			logrus.Info("验证成功：镜像已成功上传到远程仓库")
		}
	}
	
	return nil
}

func (u *Uploader) verifyPush(targetTag string) error {
	logrus.Info(fmt.Sprintf("正在验证镜像: %s", targetTag))
	
	cmd := exec.Command("docker", "manifest", "inspect", targetTag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("验证失败: %w, 输出: %s", err, string(output))
	}
	
	logrus.Debugf("镜像manifest: %s", string(output))
	return nil
}

func (u *Uploader) handlePushOutputWithProgress(src io.Reader) {
	scanner := bufio.NewScanner(src)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		progress := extractProgress(line)
		layer := extractLayer(line)

		if progress > 0 {
			if layer != "" {
				fmt.Printf("\r上传进度: [%s] %d%%", layer[:12], progress)
			} else {
				fmt.Printf("\r上传进度: %d%%", progress)
			}
			os.Stdout.Sync()
		} else {
			if strings.Contains(line, "Pushed") || strings.Contains(line, "Layer already exists") {
				fmt.Printf("\r%s\n", line)
				os.Stdout.Sync()
			} else if strings.Contains(line, "error") || strings.Contains(line, "Error") {
				fmt.Printf("\r\033[31m%s\033[0m\n", line)
				os.Stdout.Sync()
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logrus.Error(fmt.Sprintf("读取输出失败: %v", err))
	}
}

func extractProgress(line string) int {
	matches := progressRegex.FindStringSubmatch(line)
	if len(matches) == 2 {
		var progress int
		fmt.Sscanf(matches[1], "%d", &progress)
		return progress
	}
	return 0
}

func extractLayer(line string) string {
	matches := layerRegex.FindStringSubmatch(line)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func (u *Uploader) handlePushOutput(src io.Reader, dst io.Writer) {
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			dst.Write(buf[:n])
		}
		if err != nil {
			if err != io.EOF {
				logrus.Error(fmt.Sprintf("读取输出失败: %v", err))
			}
			break
		}
	}
}

func (u *Uploader) Close() {
	if u.client != nil {
		u.client.Close()
	}
}

type BatchUploadOptions struct {
	Options *UploadOptions
}

type BatchUploader struct {
	uploader *Uploader
}

func NewBatchUploader(registry, namespace string) (*BatchUploader, error) {
	uploader, err := NewUploader(registry, namespace)
	if err != nil {
		return nil, err
	}
	return &BatchUploader{uploader: uploader}, nil
}

func (bu *BatchUploader) UploadFromFile(filePath string, options *BatchUploadOptions) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		imageID := parts[0]
		version := "latest"
		if len(parts) > 1 {
			version = parts[1]
		}

		logrus.Info(fmt.Sprintf("批量上传: %s -> %s", imageID, version))

		uploadOpts := &UploadOptions{
			Version:   version,
			Registry:  options.Options.Registry,
			Namespace: options.Options.Namespace,
			Force:     options.Options.Force,
			Resume:    options.Options.Resume,
			Quiet:     options.Options.Quiet,
		}

		if err := bu.uploader.Upload(imageID, uploadOpts); err != nil {
			logrus.Error(fmt.Sprintf("上传失败 %s: %v", imageID, err))
		}

		time.Sleep(time.Second)
	}

	return nil
}
