package logger

import (
    "github.com/sirupsen/logrus"
)

var (
    Log      *logrus.Logger
    LogReport *logrus.Logger
)

func InitLogger(logServerURL string) {
    // 初始化普通日志
    Log = logrus.New()
    Log.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })

    // 初始化报告日志
    LogReport = logrus.New()
    LogReport.SetFormatter(&logrus.JSONFormatter{})
    
    // 添加HTTP Hook
    httpHook := NewHttpHook(logServerURL)
    LogReport.AddHook(httpHook)
}