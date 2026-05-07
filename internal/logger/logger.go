package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func InitLogger(logLevel string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})
	
	if err := setupFileLogging(); err != nil {
		return err
	}
	
	return nil
}

func setupFileLogging() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	logDir := filepath.Join(homeDir, ".acr-uploader", "logs")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}
	
	logPath := filepath.Join(logDir, "acr-uploader.log")
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	
	logrus.SetOutput(io.MultiWriter(os.Stdout, file))
	
	return nil
}

func GetLogLevel() logrus.Level {
	return logrus.GetLevel()
}

func SetLogLevel(level string) error {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)
	return nil
}

type ErrorType string

const (
	NetworkError      ErrorType = "NETWORK_ERROR"
	AuthFailed        ErrorType = "AUTH_FAILED"
	PermissionDenied  ErrorType = "PERMISSION_DENIED"
	ImageNotFound     ErrorType = "IMAGE_NOT_FOUND"
	TagConflict       ErrorType = "TAG_CONFLICT"
	UploadTimeout     ErrorType = "UPLOAD_TIMEOUT"
)

var errorMessages = map[ErrorType]string{
	NetworkError:      "网络连接失败，请检查网络状态后重试",
	AuthFailed:        "认证失败，请检查AccessKey是否正确",
	PermissionDenied:  "权限不足，请联系管理员授权",
	ImageNotFound:     "指定的镜像不存在，请检查镜像名称",
	TagConflict:       "标签已存在，是否覆盖？",
	UploadTimeout:     "上传超时，请重试或检查网络",
}

func GetErrorMessage(errType ErrorType) string {
	return errorMessages[errType]
}

func LogError(errType ErrorType, err error) {
	msg := GetErrorMessage(errType)
	if err != nil {
		logrus.Error(msg + ": " + err.Error())
	} else {
		logrus.Error(msg)
	}
}

func LogWithTime(msg string) {
	logrus.Info(time.Now().Format("2006-01-02 15:04:05") + " - " + msg)
}
