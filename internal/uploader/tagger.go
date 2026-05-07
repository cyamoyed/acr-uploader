package uploader

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type TagNormalizer struct {
	registry  string
	namespace string
}

func NewTagNormalizer(registry, namespace string) *TagNormalizer {
	return &TagNormalizer{
		registry:  registry,
		namespace: namespace,
	}
}

func (tn *TagNormalizer) NormalizeTag(localTag string, version string) (string, error) {
	repoName := getRepositoryName(localTag)
	if version == "" {
		version = "latest"
	}
	
	if tn.namespace == "" {
		return fmt.Sprintf("%s/%s:%s", tn.registry, repoName, version), nil
	}
	
	return fmt.Sprintf("%s/%s/%s:%s", tn.registry, tn.namespace, repoName, version), nil
}

func (tn *TagNormalizer) NormalizeTagWithRepo(localTag, repoName, version string) (string, error) {
	if version == "" {
		version = "latest"
	}
	
	if repoName == "" {
		repoName = getRepositoryName(localTag)
	}
	
	if tn.namespace == "" {
		return fmt.Sprintf("%s/%s:%s", tn.registry, repoName, version), nil
	}
	
	return fmt.Sprintf("%s/%s/%s:%s", tn.registry, tn.namespace, repoName, version), nil
}

func getRepositoryName(fullTag string) string {
	parts := strings.Split(fullTag, "/")
	if len(parts) == 0 {
		return fullTag
	}
	
	lastPart := parts[len(parts)-1]
	colonIdx := strings.LastIndex(lastPart, ":")
	if colonIdx > 0 {
		return lastPart[:colonIdx]
	}
	return lastPart
}

func (tn *TagNormalizer) TagImage(sourceTag, targetTag string) error {
	logrus.Info(fmt.Sprintf("正在打标签: %s -> %s", sourceTag, targetTag))
	
	cmd := exec.Command("docker", "tag", sourceTag, targetTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("打标签失败: %w", err)
	}
	
	logrus.Info("标签创建成功")
	return nil
}
