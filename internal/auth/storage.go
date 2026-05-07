package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	configDirName        = ".acr-uploader"
	configFileName       = "config.json"
	secretFileName       = "secret.key"
)

type Config struct {
	Username         string `json:"username"`
	Registry         string `json:"registry"`
	DefaultNamespace string `json:"default_namespace"`
	DefaultVersion   string `json:"default_version"`
	LogLevel         string `json:"log_level"`
	AccessKeyId      string `json:"access_key_id"`
	AccessKeySecret  string `json:"-"`
	RegionId         string `json:"region_id"`
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("无法获取用户目录: %w", err)
	}
	return filepath.Join(homeDir, configDirName), nil
}

func getConfigPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func getSecretPath() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, secretFileName), nil
}

func SaveConfig(config *Config) error {
	dir, err := getConfigDir()
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("无法创建配置目录: %w", err)
	}
	
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	if config.AccessKeySecret != "" {
		secretPath, err := getSecretPath()
		if err != nil {
			return err
		}
		if err := os.WriteFile(secretPath, []byte(config.AccessKeySecret), 0600); err != nil {
			return fmt.Errorf("写入密钥文件失败: %w", err)
		}
	}
	
	logrus.Info("配置已保存")
	return nil
}

func LoadConfig() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件不存在，请先执行 acr-uploader config")
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	secretPath, err := getSecretPath()
	if err != nil {
		return nil, err
	}
	secretData, err := os.ReadFile(secretPath)
	if err == nil {
		config.AccessKeySecret = string(secretData)
	}
	
	return &config, nil
}

func ConfigExists() bool {
	path, err := getConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return !os.IsNotExist(err)
}