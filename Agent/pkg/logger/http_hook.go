package logger

import (
    "bytes"
    "fmt"
    "net/http"
    "github.com/sirupsen/logrus"
    "honeypot/Agent/pkg/vars"
)

type HttpHook struct {
    client    *http.Client
    baseURL   string
}

func NewHttpHook(baseURL string) *HttpHook {
    return &HttpHook{
        client:  &http.Client{},
        baseURL: baseURL,
    }
}

func (hook *HttpHook) Fire(entry *logrus.Entry) error {
    api, exists := entry.Data["api"]
    if !exists {
        return fmt.Errorf("missing API endpoint")
    }

    url := hook.baseURL + api.(string)
    jsonData := []byte(entry.Message)

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-API-Key", vars.GlobalConfig.Client.Key)

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

func (hook *HttpHook) Levels() []logrus.Level {
    return []logrus.Level{
        logrus.InfoLevel,
        logrus.WarnLevel,
        logrus.ErrorLevel,
    }
}