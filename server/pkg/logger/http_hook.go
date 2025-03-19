package logger

import (
	"bytes"
	"fmt"
	"honeypot/server/pkg/vars"
	"net/http"
	"github.com/sirupsen/logrus"
	"encoding/json"
)

type HttpHook struct {
    client    *http.Client
    baseURL   string
}

// NewHttpHook 创建新的HTTP Hook
func NewHttpHook(baseURL string) *HttpHook {
    return &HttpHook{
        client:  &http.Client{},
        baseURL: baseURL,
    }
}

// Fire 实现Hook接口
func (hook *HttpHook) Fire(entry *logrus.Entry) error {
    // 将日志条目转换为完整JSON对象
    logData := make(map[string]interface{})
    logData["message"] = entry.Message
    for k, v := range entry.Data {
        if k != "api" { // 排除api字段
            logData[k] = v
        }
    }

    jsonData, _ := json.Marshal(logData)
    
    // 获取API路径
    apiPath, exists := entry.Data["api"]
    if !exists {
        return fmt.Errorf("missing API endpoint")
    }

    req, err := http.NewRequest("POST", hook.baseURL + apiPath.(string), bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", vars.GlobalConfig.API.Key)  // 使用相同的 API Key

    resp, err := hook.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
    }

    return nil
}

// Levels 实现Hook接口
func (hook *HttpHook) Levels() []logrus.Level {
    return []logrus.Level{
        logrus.InfoLevel,
        logrus.WarnLevel,
        logrus.ErrorLevel,
        logrus.FatalLevel,
        logrus.PanicLevel,
    }
}