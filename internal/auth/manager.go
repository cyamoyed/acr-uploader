package auth

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

type Manager struct {
	config *Config
}

func NewManager() (*Manager, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return &Manager{config: config}, nil
}

func NewManagerWithConfig(config *Config) *Manager {
	return &Manager{config: config}
}

func (m *Manager) GetCredentials() (*Credentials, error) {
	if m.config.Username == "" || m.config.Registry == "" {
		return nil, fmt.Errorf("请先配置用户名和仓库地址")
	}
	
	logrus.Info("请输入密码")
	fmt.Print("Password: ")
	
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("读取密码失败: %w", err)
	}
	fmt.Println()
	
	return &Credentials{
		Username: m.config.Username,
		Registry: m.config.Registry,
		Password: string(password),
	}, nil
}

func (m *Manager) GetConfig() *Config {
	return m.config
}

func (m *Manager) UpdateConfig(config *Config) error {
	m.config = config
	return SaveConfig(config)
}

func (m *Manager) Login() error {
	credentials, err := m.GetCredentials()
	if err != nil {
		return err
	}
	return credentials.Login()
}

func (m *Manager) Logout() error {
	return (&Credentials{Registry: m.config.Registry}).Logout()
}
