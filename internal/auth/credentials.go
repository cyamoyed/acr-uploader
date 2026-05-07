package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type Credentials struct {
	Username string `json:"username"`
	Registry string `json:"registry"`
	Password string `json:"-"`
}

func (c *Credentials) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if c.Registry == "" {
		return fmt.Errorf("仓库地址不能为空")
	}
	return nil
}

func (c *Credentials) Login() error {
	logrus.Info(fmt.Sprintf("正在登录 %s", c.Registry))
	
	cmd := exec.Command("docker", "login", "--username", c.Username, "--password-stdin", c.Registry)
	cmd.Stdin = strings.NewReader(c.Password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}
	
	logrus.Info("登录成功")
	return nil
}

func (c *Credentials) Logout() error {
	logrus.Info(fmt.Sprintf("正在登出 %s", c.Registry))
	
	cmd := exec.Command("docker", "logout", c.Registry)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("登出失败: %w", err)
	}
	
	logrus.Info("登出成功")
	return nil
}
