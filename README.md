# ACR Uploader - 阿里云容器镜像服务上传工具

[![CI](https://github.com/cyamoyed/acr-uploader/actions/workflows/ci.yml/badge.svg)](https://github.com/cyamoyed/acr-uploader/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org/)

ACR Uploader 是一个用于上传 Docker 镜像到阿里云容器镜像服务（ACR）的命令行工具，提供简化的上传流程和丰富的功能特性。

## 功能特性

- ✅ **简化登录**: 一键登录阿里云容器镜像服务
- ✅ **交互式选择**: 可视化选择要上传的镜像
- ✅ **标签规范化**: 自动将镜像标签转换为 ACR 格式
- ✅ **断点续传**: 支持上传中断后恢复
- ✅ **批量上传**: 支持从文件读取镜像列表批量上传
- ✅ **完善日志**: 详细的操作日志和错误追踪

## 安装指南

### 从源码构建

**前置条件**：确保已安装 Go 1.25+（项目要求 Go 1.25.0）

#### Windows 环境

```powershell
# 克隆仓库
git clone https://github.com/cyamoyed/acr-uploader.git
cd acr-uploader

# 直接使用 go build 构建（推荐）
go build -ldflags="-s -w" -gcflags="all=-l" -o bin/acr-uploader.exe main.go

# 或使用 PowerShell 环境变量进行跨平台构建
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -gcflags="all=-l" -o bin/acr-uploader-linux-amd64 main.go
$env:GOOS="darwin"; $env:GOARCH="arm64"; go build -ldflags="-s -w" -gcflags="all=-l" -o bin/acr-uploader-darwin-arm64 main.go
```

#### Linux/macOS 环境

```bash
# 克隆仓库
git clone https://github.com/cyamoyed/acr-uploader.git
cd acr-uploader

# 使用 make 构建（推荐，包含优化选项）
make build          # 构建当前平台
make build-all      # 构建所有平台
make build-linux    # 构建 Linux 版本
make build-windows  # 构建 Windows 版本
make build-darwin   # 构建 macOS 版本

# 或直接使用 go build（包含优化选项）
go build -ldflags="-s -w" -gcflags="all=-l" -o bin/acr-uploader main.go
```

### 手动安装

```bash
# Linux/macOS
curl -sSL https://github.com/cyamoyed/acr-uploader/releases/download/v1.0.0/acr-uploader-$(uname -s)-$(uname -m) -o /usr/local/bin/acr-uploader
chmod +x /usr/local/bin/acr-uploader

# Windows
# 从发布页面下载对应的 exe 文件
```

## 使用说明

### 配置凭证

首次使用前需要配置阿里云账号信息：

```bash
acr-uploader config --username <阿里云账号> --registry <仓库地址> --namespace <命名空间>

# 示例
acr-uploader config --username your-username --registry registry.cn-hangzhou.aliyuncs.com --namespace your-namespace
```

**参数说明**:
| 参数 | 说明 | 必填 |
|------|------|------|
| --username | 阿里云账号用户名 | 是 |
| --registry | 镜像仓库地址 | 是 |
| --namespace | 默认命名空间 | 否 |
| --version | 默认版本号 | 否（默认 latest） |
| --log-level | 日志级别 | 否（默认 info） |

### 登录

```bash
acr-uploader login
Password: ******
Login Succeeded
```

### 列出本地镜像

```bash
# 列出所有镜像
acr-uploader list

# 按名称筛选
acr-uploader list --filter-name ubuntu

# 按标签筛选
acr-uploader list --filter-tag latest
```

### 上传镜像

#### 交互式上传（推荐）

运行以下命令启动交互式镜像选择：

```bash
acr-uploader upload
```

**交互操作说明**:
- 使用 **上下方向键** 在镜像列表中导航
- 当前选中的镜像会以 **高亮显示**
- 按 **Enter 键** 确认选择
- 按 **Ctrl+C** 取消操作


#### 指定镜像上传

```bash
# 指定镜像名称
acr-uploader upload --image my-app:latest --version 1.0.0

# 指定镜像ID
acr-uploader upload --image abc123def456 --version 1.0.0
```

#### 断点续传

```bash
acr-uploader upload --resume --image my-app:latest
```

**上传参数说明**:
| 参数 | 说明 | 默认值 |
|------|------|--------|
| -i, --image | 镜像ID或名称 | 无（交互式选择） |
| -v, --version | 目标版本号 | latest |
| -f, --force | 强制覆盖已存在标签 | false |
| -R, --resume | 启用断点续传 | false |
| -q, --quiet | 静默模式 | false |

### 登出

```bash
acr-uploader logout
```

### 查看帮助

```bash
# 查看所有命令
acr-uploader help

# 查看特定命令帮助
acr-uploader help upload
acr-uploader help config
```

## 配置文件

配置文件位于 `~/.acr-uploader/config.json`，格式如下：

```json
{
    "username": "{your-username}",
    "registry": "registry.cn-hangzhou.aliyuncs.com",
    "default_namespace": "{your-namespace}",
    "default_version": "latest",
    "log_level": "info"
}
```

**字段说明**:
| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 阿里云账号用户名 |
| registry | string | 镜像仓库地址 |
| default_namespace | string | 默认命名空间 |
| default_version | string | 默认版本号 |
| log_level | string | 日志级别（debug/info/warn/error） |

## 目录结构

```
~/.acr-uploader/
├── config.json          # 配置文件
└── logs/                # 日志目录
    ├── acr-uploader.log # 当前日志
    └── upload-progress/  # 上传进度缓存
        └── <image-id>-<version>.json
```

## 错误处理

| 错误类型 | 错误码 | 用户提示 |
|----------|--------|----------|
| 网络异常 | NETWORK_ERROR | 网络连接失败，请检查网络状态后重试 |
| 认证失败 | AUTH_FAILED | 认证失败，请检查AccessKey是否正确 |
| 权限不足 | PERMISSION_DENIED | 权限不足，请联系管理员授权 |
| 镜像不存在 | IMAGE_NOT_FOUND | 指定的镜像不存在，请检查镜像名称 |
| 标签冲突 | TAG_CONFLICT | 标签已存在，是否覆盖？ |
| 上传超时 | UPLOAD_TIMEOUT | 上传超时，请重试或检查网络 |

## 示例流程

```bash
# 1. 配置凭证
acr-uploader config --username your-username --registry registry.cn-hangzhou.aliyuncs.com --namespace your-namespace

# 2. 登录
acr-uploader login

# 3. 列出镜像
acr-uploader list

# 4. 交互式上传
acr-uploader upload

# 或指定镜像上传
acr-uploader upload --image my-app:latest --version 1.0.0
```

## 技术栈

- **语言**: Go 1.25+
- **CLI框架**: cobra
- **Docker SDK**: docker/docker
- **日志库**: logrus

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
