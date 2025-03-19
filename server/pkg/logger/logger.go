package logger

import (
    "github.com/sirupsen/logrus"
)

var (
    Log       *logrus.Logger
    LogReport *logrus.Logger
)

// InitLogger 初始化日志系统
func InitLogger() {
    // 初始化普通日志
    Log = logrus.New()
    Log.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })

    // 初始化报告日志
    LogReport = logrus.New()
    LogReport.SetFormatter(&logrus.JSONFormatter{})
    
    // 添加HTTP Hook，修改端口为8081
    httpHook := NewHttpHook("http://127.0.0.1:8083")
    LogReport.AddHook(httpHook)
}